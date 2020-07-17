package ui

// #cgo pkg-config: gdk-3.0 glib-2.0 gobject-2.0
// #include <gdk/gdk.h>
import "C"
import (
	"errors"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"runtime"
	"unsafe"
)

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}
var nilPtrErr = errors.New("cgo returned unexpected nil pointer")

func PixbufNewFromData(pixbufData []byte, cs gdk.Colorspace, hasAlpha bool, bitsPerSample, width, height, rowStride int) (*gdk.Pixbuf, error) {
	arrayPtr := (*C.guchar)(unsafe.Pointer(&pixbufData[0]))

	c := C.gdk_pixbuf_new_from_data(
		arrayPtr,
		C.GdkColorspace(cs),
		gbool(hasAlpha),
		C.int(bitsPerSample),
		C.int(width),
		C.int(height),
		C.int(rowStride),
		nil, // TODO: missing support for GdkPixbufDestroyNotify
		nil,
	)

	if c == nil {
		return nil, nilPtrErr
	}

	obj := &glib.Object{GObject: glib.ToGObject(unsafe.Pointer(c))}
	p := &gdk.Pixbuf{Object: obj}
	//obj.Ref()
	runtime.SetFinalizer(p, func(_ interface{}) { obj.Unref() })

	return p, nil
}
