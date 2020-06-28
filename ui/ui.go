package ui

import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/category"
	"vincit.fi/image-sorter/common"
	"vincit.fi/image-sorter/event"
	"vincit.fi/image-sorter/library"
)

type Ui struct {
	win              *gtk.ApplicationWindow
	application      *gtk.Application
	imageManager     *library.Manager
	pixbufCache      *PixbufCache
	currentImage     *CurrentImage
	nextImages       *ImageList
	prevImages       *ImageList
	nextButton       *gtk.Button
	prevButton       *gtk.Button
	currentImageView *gtk.Viewport
	categoriesView   *gtk.Box
	broker           event.Sender
}

func Init(broker event.Sender) *Ui {

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
		s.win = getObjectOrPanic(builder, "window").(*gtk.ApplicationWindow)
		s.win.SetSizeRequest(800, 600)

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
		s.categoriesView = getObjectOrPanic(builder, "categories").(*gtk.Box)

		s.currentImageView.Connect("size-allocate", func(widget *glib.Object, data uintptr) {
			s.UpdateCurrentImage()
		})

		s.nextButton.Connect("clicked", func() {
			s.broker.SendToTopic(event.NEXT_IMAGE)
		})
		s.prevButton.Connect("clicked", func() {
			s.broker.SendToTopic(event.PREV_IMAGE)
		})

		s.broker.SendToTopic(event.UI_READY)

		// Show the Window and all of its components.
		s.win.Show()
		s.application.AddWindow(s.win)
	})
}

func (s *Ui) UpdateCategories(categories []*category.Entry) {
	children := s.categoriesView.GetChildren()

	for iter := children; iter != nil; iter = children.Next() {
		// TODO: Remove
	}

	for _, entry := range categories {
		send := func() {
			s.broker.SendToTopicWithData(
				event.CATEGORIZE_IMAGE,
				category.CategorizeCommandNew(
					s.currentImage.image, entry, category.COPY,
					))
		}
		button, _ := gtk.ButtonNewWithLabel(entry.GetName())
		button.Connect("clicked", func(button *gtk.Button) {
			send()
		})
		s.categoriesView.Add(button)
	}
	s.win.ShowAll()
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
	for _, img := range images {
		iter := list.model.Append()
		list.model.SetValue(iter, 0, s.pixbufCache.GetThumbnail(img))
	}
}

func (s *Ui) Run(args []string) {
	s.application.Run(args)
}

func (s *Ui) SetImageCategory(commands *category.CategorizeCommand) {
	// TODO: Mark image category
	log.Print("Mark image category")
}

