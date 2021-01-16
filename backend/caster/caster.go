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
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"
	"vincit.fi/image-sorter/api"
	"vincit.fi/image-sorter/api/apitype"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/common/event"
	"vincit.fi/image-sorter/common/logger"
)

const (
	deviceSearchTimeout = time.Second * 30
	imageSendTimeout    = time.Second * 1
	castService         = "_googlecast._tcp"
	canvasWidth         = 1920
	canvasHeight        = 1080
	castImageEvent      = "caster-internal-cast-image"
)

var canvasSize = apitype.SizeOf(canvasWidth, canvasHeight)

type Caster struct {
	secret                string
	port                  int
	devices               map[string]*DeviceEntry
	sender                api.Sender
	selectedDevice        string
	path                  string
	currentImage          apitype.ImageId
	server                *http.Server
	showBackground        bool
	imageCache            api.ImageStore
	alwaysStartHttpServer bool
	imageUpdateMux        sync.Mutex
	imageQueueMux         sync.Mutex
	imageQueue            apitype.ImageId
	imageQueueBroker      event.Broker

	api.Caster
}

type DeviceEntry struct {
	name         string
	serviceEntry *mdns.ServiceEntry
	device       *cast.Device
	localAddr    net.IP
}

func NewCaster(params *common.Params, sender api.Sender, imageCache api.ImageStore) api.Caster {
	c := &Caster{
		port:                  params.HttpPort(),
		alwaysStartHttpServer: params.AlwaysStartHttpServer(),
		secret:                resolveSecret(params.Secret()),
		sender:                sender,
		imageCache:            imageCache,
		showBackground:        true,
		imageQueueBroker:      *event.InitBus(100),
	}

	c.imageQueueBroker.Subscribe(castImageEvent, c.castImageFromQueue)

	if params.AlwaysStartHttpServer() {
		c.StartServer(params.HttpPort())
	}

	return c
}

func resolveSecret(secret string) string {
	if secret == "" {
		if randomSecret, err := uuid.NewRandom(); err != nil {
			logger.Error.Panic("Could not initialize secret for casting", err)
			return ""
		} else {
			return randomSecret.String()
		}
	} else {
		return secret
	}
}

func (s *Caster) StartServer(port int) {
	if s.server == nil {
		logger.Info.Printf("Starting HTTP server at port %d", s.port)
		go s.startServer(port)
	} else {
		logger.Warn.Println("Server already running")
	}
}

func (s *Caster) StopServer() {
	if s.server != nil {
		logger.Info.Println("Shutting down HTTP server")
		err := s.server.Shutdown(context.Background())
		if err != nil {
			s.sender.SendError("Error while shutting down HTTP server", err)
		}
		s.server = nil
	} else {
		logger.Debug.Println("No server running")
	}
}

func (s *Caster) startServer(port int) {
	logger.Debug.Printf("Starting HTTP server:\n"+
		" * Port: %d\n"+
		" * Secret: %s", port, s.secret)
	s.port = port

	handler := "/" + s.secret + "/"
	http.HandleFunc(handler, s.imageHandler)
	address := ":" + strconv.Itoa(port)
	s.server = &http.Server{Addr: address}
	if err := s.server.ListenAndServe(); err != nil {
		s.sender.SendError("Error while initializing HTTP server", err)
		s.server = nil
	}
}

func (s *Caster) imageHandler(responseWriter http.ResponseWriter, r *http.Request) {
	s.reserveImage()
	defer s.releaseImage()
	imageId := s.currentImage
	logger.Debug.Printf("Sending image '%d' to Chromecast", imageId)
	img, err := s.imageCache.GetScaled(imageId, apitype.SizeOf(canvasWidth, canvasHeight))

	if img != nil && err == nil {
		writeImageToResponse(responseWriter, img, s.showBackground)
	}
}

func writeImageToResponse(responseWriter http.ResponseWriter, img image.Image, showBackground bool) {
	logger.Debug.Printf("Start writing image to response")
	img = resizedAndBlurImage(img, showBackground)

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, img, nil); err != nil {
		logger.Error.Println("Failed to encode image: ", err)
		return
	}

	responseWriter.Header().Set("Content-Type", "image/jpeg")
	responseWriter.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := responseWriter.Write(buffer.Bytes()); err != nil {
		logger.Error.Println("Failed to write image: ", err)
	}
	logger.Debug.Printf("Image sent to Chromecast")
}

func resizedAndBlurImage(srcImage image.Image, blurBackground bool) image.Image {
	logger.Debug.Print("Resizing to fit canvas...")
	fullHdCanvas := image.NewRGBA(image.Rect(0, 0, canvasWidth, canvasHeight))
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	draw.Draw(fullHdCanvas, fullHdCanvas.Bounds(), &image.Uniform{C: black}, image.Point{}, draw.Src)

	srcBounds := srcImage.Bounds().Size()
	size := apitype.PointOfScaledToFit(srcBounds, canvasSize)

	if blurBackground {
		logger.Debug.Print("Blurring background...")
		// Resize to bigger so that the background surely fills the canvas
		background := imaging.Resize(srcImage, 2*canvasWidth, 2*canvasHeight, imaging.Linear)
		// Fill canvas by cropping to the canvas size
		background = imaging.Fill(srcImage, canvasWidth, canvasHeight, imaging.Center, imaging.Linear)
		// Blur and grayscale so that the background doesn't distract too much
		background = imaging.Blur(background, 10)
		background = imaging.Grayscale(background)
		draw.Draw(fullHdCanvas, fullHdCanvas.Bounds(), background, image.Point{}, draw.Src)
	}

	srcImage = imaging.Resize(srcImage, size.Width(), size.Height(), imaging.Linear)
	draw.Draw(fullHdCanvas, fullHdCanvas.Bounds(), srcImage, image.Point{X: (size.Width() - canvasWidth) / 2}, draw.Src)

	var img image.Image = fullHdCanvas
	return img
}

func (s *Caster) FindDevices() {
	s.devices = map[string]*DeviceEntry{}
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			if !strings.Contains(entry.Name, castService) {
				return
			}
			deviceName := s.resolveDeviceName(entry)

			logger.Debug.Printf("Found device: %v\n", entry)

			// Resolve the local IP address as which Chromecast sees this host.
			// This needs to be done before connecting to Chromecast otherwise the TCP
			// connection can't be established. Also this can't be figured out later
			// because all the information is private in the connection objects
			localAddr, err := s.resolveLocalAddress(entry)
			if err != nil {
				s.sender.SendError("Could not resolve local address", err)
				break
			}

			var deviceEntry = &DeviceEntry{
				serviceEntry: entry,
				localAddr:    localAddr,
			}

			s.devices[deviceName] = deviceEntry
			s.sender.SendCommandToTopic(api.CastDeviceFound, &api.DeviceFoundCommand{
				DeviceName: deviceName,
			})
		}
	}()

	c := make(chan os.Signal, 1)
	go func() {
		mdns.Query(&mdns.QueryParam{
			Service: castService,
			Timeout: deviceSearchTimeout,
			Entries: entriesCh,
		})
		s.sender.SendToTopic(api.CastDevicesSearchDone)
		close(c)
	}()

	signal.Notify(c, os.Interrupt, os.Kill)

	// Block until a signal is received.
	sig := <-c
	logger.Trace.Println("Got signal:", sig)
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
	logger.Debug.Printf("Resolving local address when connecting to")
	logger.Debug.Printf("  - Host:port: %s:%d", entry.Host, entry.Port)
	logger.Debug.Printf("  - Address: %s", entry.Addr)
	logger.Debug.Printf("  - Address v4: %s", entry.AddrV4)
	logger.Debug.Printf("  - Address v6: %s", entry.AddrV6)
	var conn net.Conn
	var err error
	const chromecastTestPort = 32768 // Just some valid UDP port on Chromecast to connect
	if entry.AddrV4 != nil {
		logger.Trace.Printf("Connecting (IPv4)...")
		if conn, err = net.Dial("udp", fmt.Sprintf("%s:%d", entry.AddrV4, chromecastTestPort)); err != nil {
			return nil, err
		}

	} else {
		logger.Trace.Printf("Connecting (IPv6)...")
		if conn, err = net.Dial("udp", fmt.Sprintf("%s:%d", entry.AddrV6, chromecastTestPort)); err != nil {
			return nil, err
		}
	}
	addr := conn.LocalAddr().(*net.UDPAddr).IP
	logger.Debug.Printf("Resolved local address to '%s'", addr.String())
	defer conn.Close()
	return addr, nil
}

func (s *Caster) SelectDevice(command *api.SelectDeviceCommand) {
	logger.Debug.Printf("Selected device '%s'", command.Name)
	s.selectedDevice = command.Name
	s.showBackground = command.ShowBackground
	device := s.devices[s.selectedDevice]
	if d, err := cast.NewDevice(device.serviceEntry.Addr, device.serviceEntry.Port); err != nil {
		s.sender.SendError("Error while selecting device", err)
	} else {
		device.device = &d
		appId := configs.MediaReceiverAppID
		device.device.ReceiverController.LaunchApplication(&appId, time.Second*5, false)

		s.StartServer(s.port)

		s.sender.SendToTopic(api.CastReady)
	}
}

func (s *Caster) localHost() string {
	if hostname, err := os.Hostname(); err != nil {
		s.sender.SendError("Could not resolve hostname", err)
		return ""
	} else {
		return hostname
	}
}

func (s *Caster) CastImage(query *api.ImageCategoryQuery) {
	s.imageQueueMux.Lock()
	defer s.imageQueueMux.Unlock()
	if query.ImageId != apitype.NoImage && s.server != nil {
		logger.Debug.Printf("Adding to cast queue: '%d'", query.ImageId)
		s.imageQueue = query.ImageId

		s.imageQueueBroker.SendToTopic(castImageEvent)
	}
}

func (s *Caster) castImageFromQueue() {
	img := s.nextImageFromQueue()
	if img != apitype.NoImage {
		s.reserveImage()
		s.currentImage = img
		s.releaseImage()
	} else {
		return
	}
	time.Sleep(1 * time.Second)

	if s.server == nil {
		logger.Error.Print("Can't cast image, server not running")
		s.sender.SendError("Can't cast image because server is not running", nil)
		return
	}

	if device, ok := s.devices[s.selectedDevice]; ok {
		logger.Debug.Println("Cast image")

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
		logger.Debug.Printf("Casting image '%s'", imageUrl)
		if _, err := device.device.MediaController.Load(imageUrl, "image/jpeg", imageSendTimeout); err != nil {
			logger.Warn.Print("Timed out while trying to cast image: ", err.Error())
		} else {
			logger.Debug.Printf("Casted image")
		}
	}
}

func (s *Caster) nextImageFromQueue() apitype.ImageId {
	s.reserveImage()
	defer s.releaseImage()
	s.imageQueueMux.Lock()
	defer s.imageQueueMux.Unlock()

	if s.imageQueue != apitype.NoImage {
		img := s.imageQueue
		logger.Debug.Printf("Getting from cast queue: '%d'", img)
		s.imageQueue = apitype.NoImage
		return img
	} else {
		return apitype.NoImage
	}
}

func (s *Caster) StopCasting() {
	if s.selectedDevice != "" {
		logger.Info.Printf("Stop casting to '%s'", s.selectedDevice)
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
	logger.Info.Println("Shutdown caster")
	s.StopCasting()
}

func (s *Caster) reserveImage() {
	s.imageUpdateMux.Lock()
}
func (s *Caster) releaseImage() {
	s.imageUpdateMux.Unlock()
}
