package caster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	cast "github.com/AndreasAbdi/gochromecast"
	"github.com/AndreasAbdi/gochromecast/configs"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/hashicorp/mdns"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Caster struct {
	secret                string
	pinCode               string
	port                  int
	devices               map[string]*DeviceEntry
	sender                api.Sender
	selectedDevice        string
	path                  string
	currentImage          apitype.ImageId
	currentImageIndex     int
	totalImages           int
	currentCategories     []*apitype.Category
	server                *http.Server
	showBackground        bool
	imageCache            api.ImageStore
	alwaysStartHttpServer bool
	imageUpdateMux        sync.Mutex
	imageQueueMux         sync.Mutex
	imageQueue            apitype.ImageId
	categories            []*apitype.Category
	websocket             *websocket.Conn
	websocketMux          sync.Mutex
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
		pinCode:               resolvePinCode("9010"),
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

func resolvePinCode(pinCode string) string {
	return pinCode
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

type SettingCategory struct {
	Id   apitype.CategoryId `json:"id"`
	Name string             `json:"name"`
}

type Settings struct {
	Categories []*SettingCategory `json:"categories"`
}

type CurrentImage struct {
	Id                apitype.ImageId      `json:"imageId"`
	CurrentImageIndex int                  `json:"currentImageIndex"`
	TotalImages       int                  `json:"totalImages"`
	Categories        []apitype.CategoryId `json:"categoryIds"`
}

type WebsocketMessage struct {
	Type    string      `json:"type"`
	Message interface{} `json:"data"`
}

func getMessageType(message interface{}) string {
	if t := reflect.TypeOf(message); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

func NewSetting(categories []*apitype.Category) Settings {
	var settingCategories []*SettingCategory

	for _, category := range categories {
		settingCategories = append(settingCategories, &SettingCategory{
			Id:   category.Id(),
			Name: category.Name(),
		})
	}

	return Settings{
		Categories: settingCategories,
	}
}

func (s *Caster) startServer(port int) {
	logger.Debug.Printf("Starting HTTP server:\n"+
		" * Port: %d\n"+
		" * Secret: %s", port, s.secret)
	s.port = port

	router := mux.NewRouter()
	imageHandler := "/" + s.secret + "/{cacheBuster}"
	router.HandleFunc(imageHandler, s.imageHandler)

	imageCommandHandler := "/command/" + s.pinCode + "/image/{command}"
	router.HandleFunc(imageCommandHandler, s.imageCommandHandler)

	categoryCommandHandler := "/command/" + s.pinCode + "/categorize"
	router.HandleFunc(categoryCommandHandler, s.categorizeCommandHandler)

	settingsQueryHandler := "/data/" + s.pinCode + "/settings"
	router.HandleFunc(settingsQueryHandler, s.settingsQueryHandler)

	websocketHandler := "/ws/" + s.pinCode
	router.HandleFunc(websocketHandler, s.websocketHandler)

	address := ":" + strconv.Itoa(port)
	s.server = &http.Server{
		Addr:    address,
		Handler: router,
	}

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

func parseJump(values url.Values) (int, error) {
	nValue := values.Get("n")
	if nValue == "" {
		return 1, nil
	} else if val, err := strconv.Atoi(nValue); err != nil {
		return 0, err
	} else {
		return val, nil
	}
}

func (s *Caster) imageCommandHandler(responseWriter http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		params := mux.Vars(r)
		switch params["command"] {
		case "next":
			if jump, err := parseJump(r.URL.Query()); err != nil {
				responseWriter.WriteHeader(400)
			} else {
				s.sender.SendCommandToTopic(api.ImageRequestNextOffset, &api.ImageAtQuery{Index: jump})
				responseWriter.WriteHeader(200)
			}
		case "previous":
			if jump, err := parseJump(r.URL.Query()); err != nil {
				responseWriter.WriteHeader(400)
			} else {
				s.sender.SendCommandToTopic(api.ImageRequestPreviousOffset, &api.ImageAtQuery{Index: jump})
				responseWriter.WriteHeader(200)
			}
		default:
			responseWriter.WriteHeader(400)
		}
	} else {
		responseWriter.WriteHeader(400)
	}
}

func (s *Caster) categorizeCommandHandler(responseWriter http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		command := &api.CategorizeCommand{}
		if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
			responseWriter.WriteHeader(400)
		} else {
			command.ImageId = s.currentImage
			if command.NextImageDelay <= 0 {
				command.NextImageDelay = 200
			}
			responseWriter.WriteHeader(200)
			s.sender.SendCommandToTopic(api.CategorizeImage, command)
		}
	} else {
		responseWriter.WriteHeader(400)
	}
}

func (s *Caster) websocketHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug.Printf("Client connected")

	if conn, err := upgrader.Upgrade(w, r, nil); err != nil {
		logger.Error.Print("Could not start WebSocket ", err)
	} else {
		s.websocketMux.Lock()
		s.websocket = conn
		s.websocketMux.Unlock()

		defer func() {
			logger.Debug.Print("Closing connection...")
			if err := conn.Close(); err != nil {
				logger.Error.Printf("Error while closing WebSocket: %s", err)
			}
		}()

		s.sendCurrentCategories()

		for {
			if mt, message, err := conn.ReadMessage(); err != nil {
				break
			} else {
				logger.Debug.Printf("Received messageType=%s; message=%s", mt, message)
			}
		}

		s.websocketMux.Lock()
		s.websocket = nil
		s.websocketMux.Unlock()
	}
}

func (s *Caster) sendToClient(message interface{}) error {
	s.websocketMux.Lock()
	defer s.websocketMux.Unlock()
	if s.websocket != nil {
		msg := WebsocketMessage{
			Type:    getMessageType(message),
			Message: message,
		}
		return s.websocket.WriteJSON(msg)
	} else {
		logger.Debug.Print("No websocket connection")
		return nil
	}
}

func (s *Caster) settingsQueryHandler(responseWriter http.ResponseWriter, r *http.Request) {
	if err := json.NewEncoder(responseWriter).Encode(NewSetting(s.categories)); err != nil {
		responseWriter.WriteHeader(400)
	} else {
		responseWriter.WriteHeader(200)
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

func (s *Caster) UpdateCategories(command *api.UpdateCategoriesCommand) {
	s.categories = command.Categories
}

func (s *Caster) SetImageCategory(imageCategorise *api.CategoriesCommand) {
	s.currentCategories = imageCategorise.Categories

	s.sendCurrentCategories()
}

func (s *Caster) SetCurrentImage(command *api.UpdateImageCommand) {
	s.currentImageIndex = command.Index
	s.totalImages = command.Total
}

func (s *Caster) sendCurrentCategories() {
	var c []apitype.CategoryId
	for _, category := range s.currentCategories {
		c = append(c, category.Id())
	}

	s.sendToClient(CurrentImage{
		Id:                s.currentImage,
		CurrentImageIndex: s.currentImageIndex,
		TotalImages:       s.totalImages,
		Categories:        c,
	})
}

func (s *Caster) CastImage(query *api.ImageCategoryQuery) {
	s.imageQueueMux.Lock()
	defer s.imageQueueMux.Unlock()
	if query.ImageId != apitype.NoImage && s.server != nil {
		logger.Debug.Printf("Adding to cast queue: '%d'", query.ImageId)
		s.imageQueue = query.ImageId

		s.imageQueueBroker.SendToTopic(castImageEvent)
	}

	var c []apitype.CategoryId
	for _, category := range s.currentCategories {
		c = append(c, category.Id())
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
