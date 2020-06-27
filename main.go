package main

import (
	"flag"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/image"
)

func CreateImageList(view *gtk.TreeView, title string) *gtk.ListStore {
	view.SetSizeRequest(100, -1)
	store, _ := gtk.ListStoreNew(image.PixbufGetType())
	view.SetModel(store)
	renderer, _ := gtk.CellRendererPixbufNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "pixbuf", 0)
	view.AppendColumn(column)
	return store
}

func AddImagesToStore(model *gtk.ListStore, manager *image.Manager, images []image.Handle) {
	model.Clear()
	for i := range images {
		iter := model.Append()
		img := images[i]
		model.SetValue(iter, 0, manager.GetThumbnail(&img))
	}
}

func main() {
	flag.Parse()
	// Create Gtk Application, change appID to your application domain name reversed.
	const appID = "org.gtk.example"
	application, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	// Check to make sure no errors when creating Gtk Application
	if err != nil {
		log.Fatal("Could not create application.", err)
	}

	// Application signals available
	// startup -> sets up the application when it first starts
	// activate -> shows the default first window of the application (like a new document). This corresponds to the application being launched by the desktop environment.
	// open -> opens files and shows them in a new window. This corresponds to someone trying to open a document (or documents) using the application from the file browser, or similar.
	// shutdown ->  performs shutdown tasks
	// Setup activate signal with a closure function.
	application.Connect("activate", func() {
		log.Println("application activate")

		builder, err := gtk.BuilderNewFromFile("ui/main-view.glade")
		errorCheck(err)

		// Get the object with the id of "main_window".
		win := getObjectOrPanic(builder, "window").(*gtk.ApplicationWindow)

		root := flag.Arg(0)
		manager := image.ManagerForDir(root)
		currentImageView := getObjectOrPanic(builder, "current-image-view").(*gtk.Viewport)

		currentImage := getObjectOrPanic(builder, "current-image").(*gtk.Image)

		_, err = currentImageView.Connect("size-allocate", func(widget *glib.Object, data uintptr) {
			scaled := manager.GetScaled(
				manager.GetCurrentImage(),
				currentImageView.GetAllocatedWidth(),
				currentImageView.GetAllocatedHeight(),
			)
			currentImage.SetFromPixbuf(scaled)
		})
		if err != nil {
			log.Panic(err)
		}

		nextImagesList := getObjectOrPanic(builder, "next-images").(*gtk.TreeView)
		nextImageStore := CreateImageList(nextImagesList, "Next images")
		prevImagesList := getObjectOrPanic(builder, "prev-images").(*gtk.TreeView)
		prevImageStore := CreateImageList(prevImagesList, "Prev images")

		updateImages := func(currentImageHandle *image.Handle) {
			scaled := manager.GetScaled(
				currentImageHandle,
				currentImageView.GetAllocatedWidth(),
				currentImageView.GetAllocatedHeight())
			currentImage.SetFromPixbuf(scaled)
			AddImagesToStore(nextImageStore, &manager, manager.GetNextImages(5))
			AddImagesToStore(prevImageStore, &manager, manager.GetPrevImages(5))
		}

		nextButton := getObjectOrPanic(builder, "next-button").(*gtk.Button)
		nextButton.Connect("clicked", func() {
			updateImages(manager.NextImage())
		})
		prevButton := getObjectOrPanic(builder, "prev-button").(*gtk.Button)
		prevButton.Connect("clicked", func() {
			updateImages(manager.PrevImage())
		})

		updateImages(manager.GetCurrentImage())
		// Show the Window and all of its components.
		win.Show()
		application.AddWindow(win)
	})
	// Run Gtk application
	application.Run([]string{})
}

func getObjectOrPanic(builder *gtk.Builder, name string) glib.IObject {
	obj, err := builder.GetObject(name)
	if err != nil {
		log.Panic("Could not load object " + name)
	}
	return obj
}

func errorCheck(e error) {
	if e != nil {
		// panic for any errors.
		log.Panic(e)
	}
}
