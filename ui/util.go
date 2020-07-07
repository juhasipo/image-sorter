package ui

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"
import (
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"log"
	"vincit.fi/image-sorter/event"
)

// PixbufGetType is a wrapper around gdk_pixbuf_get_type().
func PixbufGetType() glib.Type {
	return glib.Type(C.gdk_pixbuf_get_type())
}


func getObjectOrPanic(builder *gtk.Builder, name string) glib.IObject {
	obj, err := builder.GetObject(name)
	if err != nil {
		log.Panic("Could not load object ",name, ": ", err)
	}
	return obj
}

type Direction int
const (
	FORWARD Direction = iota
	BACKWARD
)

func CreateImageList(view *gtk.TreeView, title string, direction Direction, sender event.Sender) *gtk.ListStore {
	view.SetSizeRequest(100, -1)
	view.Connect("row-activated", func(view *gtk.TreeView, path *gtk.TreePath, col *gtk.TreeViewColumn) {
		index := path.GetIndices()[0] + 1
		if direction == FORWARD {
			sender.SendToTopicWithData(event.IMAGE_REQUEST_NEXT_OFFSET, index)
		} else {
			sender.SendToTopicWithData(event.IMAGE_REQUEST_PREV_OFFSET, index)
		}
	})
	store, _ := gtk.ListStoreNew(PixbufGetType())
	view.SetModel(store)
	renderer, _ := gtk.CellRendererPixbufNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "pixbuf", 0)
	view.AppendColumn(column)
	return store
}

func CreateDeviceList(modal *gtk.Dialog, view *gtk.TreeView, title string, sender event.Sender) *gtk.ListStore {
	store, _ := gtk.ListStoreNew(glib.TYPE_STRING)
	view.SetSizeRequest(100, -1)
	view.Connect("row-activated", func(view *gtk.TreeView, path *gtk.TreePath, col *gtk.TreeViewColumn) {
		iter, _ := store.GetIter(path)
		value, _ := store.GetValue(iter, 0)
		stringValue, _ := value.GetString()
		sender.SendToTopicWithData(event.CAST_DEVICE_SELECT, stringValue)
		modal.Hide()
	})
	view.SetModel(store)
	renderer, _ := gtk.CellRendererTextNew()
	column, _ := gtk.TreeViewColumnNewWithAttribute(title, renderer, "text", 0)
	view.AppendColumn(column)
	return store
}

func KeyvalName(keyval uint) string {
	return C.GoString((*C.char)(C.gdk_keyval_name(C.guint(keyval))))
}
