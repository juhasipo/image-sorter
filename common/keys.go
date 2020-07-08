package common

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"

func KeyvalName(keyval uint) string {
	return C.GoString((*C.char)(C.gdk_keyval_name(C.guint(keyval))))
}
