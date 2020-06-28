package ui

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/library"
)

type ImageList struct {
	component *gtk.TreeView
	model *gtk.ListStore
}

type CurrentImage struct {
	view *gtk.Image
	image *common.Handle
}

type Ui struct {
	application      *gtk.Application
	imageManager     *library.Manager
	pixbufCache      *PixbufCache
	currentImage     *CurrentImage
	nextImages       *ImageList
	prevImages       *ImageList
	nextButton       *gtk.Button
	prevButton       *gtk.Button
	currentImageView *gtk.Viewport
	broker           *event.Broker
}

func Init(broker *event.Broker) *Ui {

	// Create Gtk Application, change appID to your application domain name reversed.
	const appID = "org.gtk.example"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)

	// Check to make sure no errors when creating Gtk Application
	if err != nil {
		log.Fatal("Could not create application.", err)
	}

	ui := Ui{
		application: application,
		pixbufCache: &PixbufCache {
			imageCache: map[common.Handle]*Instance{},
		},
		broker: broker,
	}

	ui.Init()
	return &ui
}

func (s *Ui) Init() {
	// Application signals available
	// startup -> sets up the application when it first starts
	// activate -> shows the default first window of the application (like a new document). This corresponds to the application being launched by the desktop environment.
	// open -> opens files and shows them in a new window. This corresponds to someone trying to open a document (or documents) using the application from the file browser, or similar.
	// shutdown ->  performs shutdown tasks
	// Setup activate signal with a closure function.
	s.application.Connect("activate", func() {
		log.Println("application activate")

		builder, err := gtk.BuilderNewFromFile("ui/main-view.glade")
		if err != nil {
			log.Fatal("Could not load Glade file.", err)
		}

		// Get the object with the id of "main_window".
		win := getObjectOrPanic(builder, "window").(*gtk.ApplicationWindow)

		nextImagesList := getObjectOrPanic(builder, "next-images").(*gtk.TreeView)
		nextImageStore := CreateImageList(nextImagesList, "Next images")
		prevImagesList := getObjectOrPanic(builder, "prev-images").(*gtk.TreeView)
		prevImageStore := CreateImageList(prevImagesList, "Prev images")

		s.nextImages = &ImageList{
			component: nextImagesList,
			model:    nextImageStore,
		}
		s.prevImages = &ImageList{
			component: prevImagesList,
			model:    prevImageStore,
		}
		s.currentImage = &CurrentImage {
			view: getObjectOrPanic(builder, "current-image").(*gtk.Image),
		}
		s.currentImageView = getObjectOrPanic(builder, "current-image-view").(*gtk.Viewport)
		s.nextButton = getObjectOrPanic(builder, "next-button").(*gtk.Button)
		s.prevButton = getObjectOrPanic(builder, "prev-button").(*gtk.Button)

		s.currentImageView.Connect("size-allocate", func(widget *glib.Object, data uintptr) {
			s.UpdateCurrentImage()
		})

		s.nextButton.Connect("clicked", func() {
			s.broker.Send(event.New(event.NEXT_IMAGE))
		})
		s.prevButton.Connect("clicked", func() {
			s.broker.Send(event.New(event.PREV_IMAGE))
		})

		s.broker.Send(event.New(event.CURRENT_IMAGE))

		// Show the Window and all of its components.
		win.Show()
		s.application.AddWindow(win)
	})
}

func (s *Ui) UpdateCurrentImage() {
	scaled := s.pixbufCache.GetScaled(
		s.currentImage.image,
		SizeFromViewport(s.currentImageView),
	)
	s.currentImage.view.SetFromPixbuf(scaled)
}

func (s* Ui) SetImages(handles []*common.Handle, imageTarget event.Topic) {
	if imageTarget == event.NEXT_IMAGE {
		s.AddImagesToStore(s.nextImages, handles)
	} else if imageTarget == event.PREV_IMAGE {
		s.AddImagesToStore(s.prevImages, handles)
	} else {
		s.SetCurrentImage(handles[0])
	}
}

func (s *Ui) SetCurrentImage(handle *common.Handle) {
	s.currentImage.image = handle
	s.UpdateCurrentImage()
}

func (s *Ui) AddImagesToStore(list *ImageList, images []*common.Handle) {
	list.model.Clear()
	for i := range images {
		iter := list.model.Append()
		img := images[i]
		list.model.SetValue(iter, 0, s.pixbufCache.GetThumbnail(img))
	}
}

func (s *Ui) Run(args []string) {
	s.application.Run(args)
}


func getObjectOrPanic(builder *gtk.Builder, name string) glib.IObject {
	obj, err := builder.GetObject(name)
	if err != nil {
		log.Panic("Could not load object " + name)
	}
	return obj
}


func CreateImageList(view *gtk.TreeView, title string) *gtk.ListStore {
	view.SetSizeRequest(100, -1)
	store, _ := gtk.ListStoreNew(PixbufGetType())
	view.SetModel(store)
	renderer, _ := gtk.CellRendererPixbufNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "pixbuf", 0)
	view.AppendColumn(column)
	return store
}

