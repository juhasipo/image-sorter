package caster

import (
	"bytes"
	"context"
	"fmt"
	cast "github.com/AndreasAbdi/gochromecast"
	"github.com/AndreasAbdi/gochromecast/configs"
	"github.com/AndreasAbdi/gochromecast/controllers"
	"github.com/AndreasAbdi/gochromecast/primitives"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/hashicorp/mdns"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
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
	imageCache            *imageloader.ImageCache
	alwaysStartHttpServer bool
}

type DeviceEntry struct {
	name            string
	serviceEntry    *mdns.ServiceEntry
	heartbeat       *controllers.HeartbeatController
	connection      *controllers.ConnectionController
	receiver        *controllers.ReceiverController
	mediaController *controllers.MediaController
	device          *cast.Device
	localAddr       *net.TCPAddr
}

func InitCaster(port int, alwaysStartHttpServer bool, secret string, sender event.Sender, imageCache *imageloader.ImageCache) (*Caster, error) {
	c := &Caster{
		port:                  port,
		alwaysStartHttpServer: alwaysStartHttpServer,
		secret:                secret,
		sender:                sender,
		imageCache:            imageCache,
		showBackground:        true,
	}

	if alwaysStartHttpServer {
		c.StartServer(port)
	}

	return c, nil
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
	img := s.imageCache.GetScaled(s.currentImage, common.SizeOf(CANVAS_WIDTH, CANVAS_HEIGHT))

	if img != nil {
		writeImageToResponse(responseWriter, img, s.showBackground)
	}
}

func writeImageToResponse(responseWriter http.ResponseWriter, img image.Image, showBackground bool) {
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
}

func resizedAndBlurImage(srcImage image.Image, blurBackground bool) image.Image {
	fullHdCanvas := image.NewRGBA(image.Rect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	draw.Draw(fullHdCanvas, fullHdCanvas.Bounds(), &image.Uniform{C: black}, image.Point{}, draw.Src)

	srcBounds := srcImage.Bounds().Size()
	w, h := common.ScaleToFit(srcBounds.X, srcBounds.Y, CANVAS_WIDTH, CANVAS_HEIGHT)

	if blurBackground {
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

			client, err := primitives.NewClient(entry.Addr, entry.Port)
			if err != nil {
				log.Println("Could not create new client", err)
				break
			}

			receiver := controllers.NewReceiverController(client, "sender-0", "receiver-0")
			var deviceEntry = &DeviceEntry{
				serviceEntry:    entry,
				localAddr:       localAddr,
				heartbeat:       controllers.NewHeartbeatController(client, "sender-0", "receiver-0"),
				connection:      controllers.NewConnectionController(client, "sender-0", "receiver-0"),
				receiver:        receiver,
				mediaController: controllers.NewMediaController(client, "sender-0", receiver),
			}
			response, err := deviceEntry.receiver.GetStatus(time.Second * 5)

			log.Println("Status response: ", response, err)

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

func (s *Caster) resolveLocalAddress(entry *mdns.ServiceEntry) (*net.TCPAddr, error) {
	log.Printf("Resolving local address when connecting to")
	log.Printf("  - Host:port: %s:%d", entry.Host, entry.Port)
	log.Printf("  - Address: %s", entry.Addr)
	log.Printf("  - Address v4: %s", entry.AddrV4)
	log.Printf("  - Address v6: %s", entry.AddrV6)
	var conn net.Conn
	var err error
	if entry.AddrV4 != nil {
		log.Printf("Connecting (IPv4)...")
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", entry.AddrV4, entry.Port))
		if err != nil {
			return nil, err
		}
	} else {
		log.Printf("Connecting (IPv6)...")
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", entry.AddrV6, entry.Port))
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		return nil, err
	}

	defer conn.Close()
	return conn.LocalAddr().(*net.TCPAddr), nil
}

func (s *Caster) SelectDevice(name string, showBackground bool) {
	log.Printf("Selected device '%s'", name)
	s.selectedDevice = name
	s.showBackground = showBackground
	device := s.devices[s.selectedDevice]
	device.heartbeat.Start()
	device.connection.Connect()
	d, err := cast.NewDevice(device.serviceEntry.Addr, device.serviceEntry.Port)
	if err != nil {
		log.Println("Could no select device: '"+name+"'", err)
	}
	device.device = &d
	appId := configs.MediaReceiverAppID
	device.receiver.LaunchApplication(&appId, time.Second*5, false)

	s.StartServer(s.port)

	s.sender.SendToTopic(event.CAST_READY)
}

func (s *Caster) getLocalHost() string {
	hostname, _ := os.Hostname()
	return hostname
}

func (s *Caster) CastImage(handle *common.Handle) {
	s.currentImage = handle
	if device, ok := s.devices[s.selectedDevice]; ok {
		log.Println("Cast image")

		// Send a random string as part of path so that Chromecast
		// triggers image change. The server will decide which image to show
		// This way the outside world can't decide what is served which makes
		// this slightly more secure (no need to validate/sanitize file paths)
		cacheBuster, _ := uuid.NewRandom()
		ip := device.localAddr.IP.String()
		imageUrl := fmt.Sprintf("http://%s:%d/%s/%s", ip, s.port, s.secret, cacheBuster.String())
		log.Printf("Casting image '%s'", imageUrl)
		device.mediaController.Load(imageUrl, "image/jpeg", time.Second*5)
		log.Printf("Casted image")
	}
}

func (s *Caster) StopCasting() {
	if s.selectedDevice != "" {
		log.Printf("Stop casting to '%s'", s.selectedDevice)
		if device, ok := s.devices[s.selectedDevice]; ok {
			device.mediaController.Stop(time.Second * 5)
			device.connection.Close()
			device.heartbeat.Stop()
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
