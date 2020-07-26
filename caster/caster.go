package caster

import (
	"bytes"
	"context"
	"fmt"
	cast "github.com/AndreasAbdi/gochromecast"
	"github.com/AndreasAbdi/gochromecast/configs"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/hashicorp/mdns"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/imageloader"
)

const (
	DEVICE_SEARCH_TIMEOUT = time.Second * 30
	CAST_SERVICE          = "_googlecast._tcp"
	CANVAS_WIDTH          = 1920
	CANVAS_HEIGHT         = 1080
	CAST_IMAGE_EVENT      = "caster-internal-cast-image"
)

type Caster struct {
	secret                string
	port                  int
	devices               map[string]*DeviceEntry
	sender                event.Sender
	selectedDevice        string
	path                  string
	currentImage          *common.Handle
	server                *http.Server
	showBackground        bool
	imageCache            imageloader.ImageCache
	alwaysStartHttpServer bool
	imageUpdateMux        sync.Mutex
	imageQueueMux         sync.Mutex
	imageQueue            []*common.Handle
	imageQueueBroker      event.Broker
}

type DeviceEntry struct {
	name         string
	serviceEntry *mdns.ServiceEntry
	device       *cast.Device
	localAddr    net.IP
}

func InitCaster(port int, alwaysStartHttpServer bool, secret string, sender event.Sender, imageCache imageloader.ImageCache) *Caster {
	c := &Caster{
		port:                  port,
		alwaysStartHttpServer: alwaysStartHttpServer,
		secret:                secret,
		sender:                sender,
		imageCache:            imageCache,
		showBackground:        true,
		imageQueueBroker:      *event.InitBus(100),
	}

	c.imageQueueBroker.Subscribe(CAST_IMAGE_EVENT, c.castImageFromQueue)

	if alwaysStartHttpServer {
		c.StartServer(port)
	}

	return c
}

func (s *Caster) StartServer(port int) {
	if s.server == nil {
		log.Printf("Starting HTTP server at port %d", s.port)
		go s.startServerAsync(port)
	} else {
		log.Println("Server already running")
	}
}

func (s *Caster) StopServer() {
	if s.server != nil {
		log.Println("Shutting down HTTP server")
		err := s.server.Shutdown(context.Background())
		if err != nil {
			log.Println(err)
		}
		s.server = nil
	} else {
		log.Println("No server running")
	}
}

func (s *Caster) startServerAsync(port int) {
	log.Printf("Starting HTTP server:\n"+
		" * Port: %d\n"+
		" * Secret: %s", port, s.secret)
	s.port = port

	handler := "/" + s.secret + "/"
	http.HandleFunc(handler, s.imageHandler)
	address := ":" + strconv.Itoa(port)
	s.server = &http.Server{Addr: address}
	if err := s.server.ListenAndServe(); err != nil {
		log.Println("Could not initialize server", err)
	}
}

func (s *Caster) imageHandler(responseWriter http.ResponseWriter, r *http.Request) {
	s.reserveImage()
	defer s.releaseImage()
	imageHandle := s.currentImage
	log.Printf("Sending image '%s' to Chromecast", imageHandle.GetId())
	img := s.imageCache.GetScaled(imageHandle, common.SizeOf(CANVAS_WIDTH, CANVAS_HEIGHT))

	if img != nil {
		writeImageToResponse(responseWriter, img, s.showBackground)
	}
}

func writeImageToResponse(responseWriter http.ResponseWriter, img image.Image, showBackground bool) {
	log.Printf("Start writing image to response")
	img = resizedAndBlurImage(img, showBackground)

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, img, nil); err != nil {
		log.Println("Failed to encode image: ", err)
		return
	}

	responseWriter.Header().Set("Content-Type", "image/jpeg")
	responseWriter.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := responseWriter.Write(buffer.Bytes()); err != nil {
		log.Println("Failed to write image: ", err)
	}
	log.Printf("Image sent to Chromecast")
}

func resizedAndBlurImage(srcImage image.Image, blurBackground bool) image.Image {
	log.Print("Resizing to fit canvas...")
	fullHdCanvas := image.NewRGBA(image.Rect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	draw.Draw(fullHdCanvas, fullHdCanvas.Bounds(), &image.Uniform{C: black}, image.Point{}, draw.Src)

	srcBounds := srcImage.Bounds().Size()
	w, h := common.ScaleToFit(srcBounds.X, srcBounds.Y, CANVAS_WIDTH, CANVAS_HEIGHT)

	if blurBackground {
		log.Print("Blurrign background...")
		// Resize to bigger so that the background surely fills the canvas
		background := imaging.Resize(srcImage, 2*CANVAS_WIDTH, 2*CANVAS_HEIGHT, imaging.Linear)
		// Fill canvas by cropping to the canvas size
		background = imaging.Fill(srcImage, CANVAS_WIDTH, CANVAS_HEIGHT, imaging.Center, imaging.Linear)
		// Blur and grayscale so that the background doesn't distract too much
		background = imaging.Blur(background, 10)
		background = imaging.Grayscale(background)
		draw.Draw(fullHdCanvas, fullHdCanvas.Bounds(), background, image.Point{}, draw.Src)
	}

	srcImage = imaging.Resize(srcImage, w, h, imaging.Linear)
	draw.Draw(fullHdCanvas, fullHdCanvas.Bounds(), srcImage, image.Point{X: (w - CANVAS_WIDTH) / 2}, draw.Src)

	var img image.Image = fullHdCanvas
	return img
}

func (s *Caster) FindDevices() {
	s.devices = map[string]*DeviceEntry{}
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			if !strings.Contains(entry.Name, CAST_SERVICE) {
				return
			}
			deviceName := s.resolveDeviceName(entry)

			fmt.Printf("Found device: %v\n", entry)

			// Resolve the local IP address as which Chromecast sees this host.
			// This needs to be done before connecting to Chromecast otherwise the TCP
			// connection can't be established. Also this can't be figured out later
			// because all the information is private in the connection objects
			localAddr, err := s.resolveLocalAddress(entry)
			if err != nil {
				log.Println("Could not resolve local address", err)
				break
			}

			var deviceEntry = &DeviceEntry{
				serviceEntry: entry,
				localAddr:    localAddr,
			}

			s.devices[deviceName] = deviceEntry
			s.sender.SendToTopicWithData(event.CAST_DEVICE_FOUND, deviceName)
		}
	}()

	c := make(chan os.Signal, 1)
	go func() {
		mdns.Query(&mdns.QueryParam{
			Service: CAST_SERVICE,
			Timeout: DEVICE_SEARCH_TIMEOUT,
			Entries: entriesCh,
		})
		s.sender.SendToTopic(event.CAST_DEVICES_SEARCH_DONE)
		close(c)
	}()

	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	sig := <-c
	fmt.Println("Got signal:", sig)
}

func (s *Caster) resolveDeviceName(entry *mdns.ServiceEntry) string {
	var name string
	for _, field := range entry.InfoFields {
		if strings.HasPrefix(field, "fn=") {
			name = strings.ReplaceAll(field, "fn=", "")
		}
	}
	return name
}

func (s *Caster) resolveLocalAddress(entry *mdns.ServiceEntry) (net.IP, error) {
	log.Printf("Resolving local address when connecting to")
	log.Printf("  - Host:port: %s:%d", entry.Host, entry.Port)
	log.Printf("  - Address: %s", entry.Addr)
	log.Printf("  - Address v4: %s", entry.AddrV4)
	log.Printf("  - Address v6: %s", entry.AddrV6)
	var conn net.Conn
	var err error
	if entry.AddrV4 != nil {
		log.Printf("Connecting (IPv4)...")
		if conn, err = net.Dial("udp", fmt.Sprintf("%s:%d", entry.AddrV4, 32768)); err != nil {
			return nil, err
		}

	} else {
		log.Printf("Connecting (IPv6)...")
		if conn, err = net.Dial("udp", fmt.Sprintf("%s:%d", entry.AddrV6, 32768)); err != nil {
			return nil, err
		}
	}
	addr := conn.LocalAddr().(*net.UDPAddr).IP
	defer conn.Close()
	return addr, nil
}

func (s *Caster) SelectDevice(name string, showBackground bool) {
	log.Printf("Selected device '%s'", name)
	s.selectedDevice = name
	s.showBackground = showBackground
	device := s.devices[s.selectedDevice]
	if d, err := cast.NewDevice(device.serviceEntry.Addr, device.serviceEntry.Port); err != nil {
		log.Println("Could no select device: '"+name+"'", err)
	} else {
		device.device = &d
		appId := configs.MediaReceiverAppID
		device.device.ReceiverController.LaunchApplication(&appId, time.Second*5, false)

		s.StartServer(s.port)

		s.sender.SendToTopic(event.CAST_READY)
	}
}

func (s *Caster) getLocalHost() string {
	if hostname, err := os.Hostname(); err != nil {
		log.Panic("Could not get hostname", err)
		return ""
	} else {
		return hostname
	}
}

func (s *Caster) CastImage(handle *common.Handle) {
	s.imageQueueMux.Lock()
	defer s.imageQueueMux.Unlock()
	log.Printf("Adding to cast queue: '%s'", handle.GetId())
	s.imageQueue = append(s.imageQueue, handle)

	s.imageQueueBroker.SendToTopic(CAST_IMAGE_EVENT)
}

func (s *Caster) castImageFromQueue() {
	img := s.getImageFromQueue()
	if img != nil {
		s.reserveImage()
		s.currentImage = img
		s.releaseImage()
	} else {
		log.Printf("Nothing new to cast")
		return
	}
	time.Sleep(1 * time.Second)

	if device, ok := s.devices[s.selectedDevice]; ok {
		log.Println("Cast image")

		// Send a random string as part of path so that Chromecast
		// triggers image change. The server will decide which image to show
		// This way the outside world can't decide what is served which makes
		// this slightly more secure (no need to validate/sanitize file paths)
		var cacheBusterStr string
		if cacheBuster, err := uuid.NewRandom(); err != nil {
			cacheBusterStr = strconv.Itoa(rand.Int())
		} else {
			cacheBusterStr = cacheBuster.String()
		}

		ip := device.localAddr.String()
		imageUrl := fmt.Sprintf("http://%s:%d/%s/%s", ip, s.port, s.secret, cacheBusterStr)
		log.Printf("Casting image '%s'", imageUrl)
		if _, err := device.device.MediaController.Load(imageUrl, "image/jpeg", time.Second*5); err != nil {
			log.Print("Could not cast image", err)
		} else {
			log.Printf("Casted image")
		}
	}
}

func (s *Caster) getImageFromQueue() *common.Handle {
	s.reserveImage()
	defer s.releaseImage()
	s.imageQueueMux.Lock()
	defer s.imageQueueMux.Unlock()

	if len(s.imageQueue) > 0 {
		img := s.imageQueue[len(s.imageQueue)-1]
		log.Printf("Getting from cast queue: '%s'", img.GetId())
		s.imageQueue = []*common.Handle{}
		return img
	} else {
		return nil
	}
}

func (s *Caster) StopCasting() {
	if s.selectedDevice != "" {
		log.Printf("Stop casting to '%s'", s.selectedDevice)
		if device, ok := s.devices[s.selectedDevice]; ok {
			device.device.QuitApplication(time.Second * 5)
			s.selectedDevice = ""
		}

		if !s.alwaysStartHttpServer {
			s.StopServer()
		}
	}
}

func (s *Caster) Close() {
	log.Println("Shutdown caster")
	s.StopCasting()
}

func (s *Caster) reserveImage() {
	s.imageUpdateMux.Lock()
}
func (s *Caster) releaseImage() {
	s.imageUpdateMux.Unlock()
}
