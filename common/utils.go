package common

import (
	"github.com/gotk3/gotk3/gdk"
	"strings"
)
func KeyToUint(key string) []uint {
	return []uint {
		gdk.KeyvalFromName(strings.ToLower(key)),
		gdk.KeyvalFromName(strings.ToUpper(key)),
	}
}
