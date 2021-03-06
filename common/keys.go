package common

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"
import (
	"github.com/gotk3/gotk3/gdk"
	"strings"
)

/*
These have to stay under common since
* UI uses them to convert between GTK <-> Human
* Backend uses to convert between Serialized <-> GTK
*/

func KeyvalName(keyval uint) string {
	return C.GoString((*C.char)(C.gdk_keyval_name(C.guint(keyval))))
}

func KeyToUint(key string) uint {
	return gdk.KeyvalFromName(strings.ToUpper(key))
}
