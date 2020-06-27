package main

import (
	"flag"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/image"
)

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
		var manager = image.ManagerForDir(root)
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

		nextButton := getObjectOrPanic(builder, "next-button").(*gtk.Button)
		nextButton.Connect("clicked", func() {
			scaled := manager.GetScaled(manager.NextImage(), currentImageView.GetAllocatedWidth(), currentImageView.GetAllocatedHeight())
			currentImage.SetFromPixbuf(scaled)
		})
		prevButton := getObjectOrPanic(builder, "prev-button").(*gtk.Button)
		prevButton.Connect("clicked", func() {
			scaled := manager.GetScaled(manager.PrevImage(), currentImageView.GetAllocatedWidth(), currentImageView.GetAllocatedHeight())
			currentImage.SetFromPixbuf(scaled)
		})

		// Show the Window and all of its components.
		win.Show()
		application.AddWindow(win)
	})
	// Run Gtk application
	application.Run()
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
