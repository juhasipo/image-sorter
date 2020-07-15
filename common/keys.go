package common

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"
import (
	"github.com/gotk3/gotk3/gdk"
	"strings"
)

func KeyvalName(keyval uint) string {
	return C.GoString((*C.char)(C.gdk_keyval_name(C.guint(keyval))))
}

func KeyToUint(key string) []uint {
	return []uint {
		gdk.KeyvalFromName(strings.ToLower(key)),
		gdk.KeyvalFromName(strings.ToUpper(key)),
	}
}
