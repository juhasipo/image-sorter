package caster

import (
	"fmt"
	cast "github.com/AndreasAbdi/gochromecast"
	"github.com/AndreasAbdi/gochromecast/configs"
	"github.com/AndreasAbdi/gochromecast/controllers"
	"github.com/AndreasAbdi/gochromecast/primitives"
	"github.com/hashicorp/mdns"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"time"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
)

type Caster struct {
	secret  string
	port    int
	devices map[string]*DeviceEntry
	sender  event.Sender
	selectedDevice string
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

func InitCaster(secret string, sender event.Sender) (*Caster, error) {
	c := &Caster{
		secret: secret,
		sender: sender,
	}

	return c, nil
}

func (s* Caster) StartServer(port int, path string) {
	go s.startServerAsync(port, path)
}

func (s *Caster) startServerAsync(port int, path string) {
	s.port = port
	server := http.FileServer(http.Dir(path))
	prefix := http.StripPrefix("/" + s.secret, server)
	http.Handle("/", prefix)
	address := ":" + strconv.Itoa(port)
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Panic(err)
	}
}

func (s* Caster) FindDevices() {
	s.devices = map[string]*DeviceEntry{}
	castService := "_googlecast._tcp"
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			if !strings.Contains(entry.Name, castService) {
				return
			}
			deviceName := s.resolveDeviceName(entry)

			fmt.Printf("Found device: %v\n", entry)

			// Resolve the local IP address as which Chromecast sees this host.
			// This needs to be done before connecting to Chromecast otherwise the TCP
			// connection can't be established. Also this can't be figured out later
			// because all the information is private in the connection objects
			localAddr, err := s.resolveLocalAddress(entry.Host, entry.Port)
			if err != nil {
				log.Panic(err)
			}

			client, err := primitives.NewClient(entry.Addr, entry.Port)
			if err != nil {
				log.Panic(err)
			}

			receiver := controllers.NewReceiverController(client, "sender-0", "receiver-0")
			var deviceEntry = &DeviceEntry {
				serviceEntry:    entry,
				localAddr:       localAddr,
				heartbeat:       controllers.NewHeartbeatController(client, "sender-0", "receiver-0"),
				connection:      controllers.NewConnectionController(client, "sender-0", "receiver-0"),
				receiver:        receiver,
				mediaController: controllers.NewMediaController(client, "sender-0", receiver),
			}
			response, err := deviceEntry.receiver.GetStatus(time.Second * 5)

			log.Print("Status response: ", response, err)

			s.devices[deviceName] = deviceEntry
			s.sender.SendToTopicWithData(event.CAST_DEVICE_FOUND, deviceName)
		}
	}()

	go func() {
		mdns.Query(&mdns.QueryParam{
			Service: castService,
			Timeout: time.Second * 30,
			Entries: entriesCh,
		})
	}()

	c := make(chan os.Signal, 1)
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

func (s *Caster) resolveLocalAddress(host string, port int) (*net.TCPAddr, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	defer conn.Close()
	return conn.LocalAddr().(*net.TCPAddr), err
}

func (s* Caster) SelectDevice(name string) {
	log.Printf("Selected device '%s'", name)
	s.selectedDevice = name
	device := s.devices[s.selectedDevice]
	device.heartbeat.Start()
	device.connection.Connect()
	d, err := cast.NewDevice(device.serviceEntry.Addr, device.serviceEntry.Port)
	if err != nil {
		log.Panic(err)
	}
	device.device = &d
	appId := configs.MediaReceiverAppID
	device.receiver.LaunchApplication(&appId, time.Second * 5, false)
	s.sender.SendToTopic(event.CAST_READY)
}

func (s *Caster) getLocalHost() string {
	hostname, _ := os.Hostname()
	return hostname
}

func (s* Caster) CastImage(handle *common.Handle) {
	log.Print("Cast image")
	if device, ok := s.devices[s.selectedDevice]; ok {
	 	ip := device.localAddr.IP.String()
		imageUrl := fmt.Sprintf("http://%s:%d/%s/%s", ip, s.port, s.secret, path.Base(handle.GetPath()))
		log.Printf("Casting image '%s'", imageUrl)
		device.mediaController.Load(imageUrl, "image/jpeg", time.Second * 5)
		//device.device.PlayMedia(imageUrl, "image/jpeg")
		log.Printf("Casted image")
	} else {
		log.Print("No device selected")
	}
}

func (s* Caster) StopCasting() {
	device := s.devices[s.selectedDevice]
	device.mediaController.Stop(time.Second * 5)
	device.connection.Close()
	device.heartbeat.Stop()
	s.selectedDevice = ""
}
