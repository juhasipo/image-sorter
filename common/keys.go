package common
import "C"
import "github.com/go-gl/glfw/v3.3/glfw"

/*
These have to stay under common since
* UI uses them to convert between GTK <-> Human
* Backend uses to convert between Serialized <-> GTK
*/

var keymap = map[string]glfw.Key{}

func InitKeyMap() {
	/*
	for i := glfw.Key(0); i < glfw.KeyLast; i++ {
		if i != glfw.KeyUnknown {
			keymap[glfw.GetKeyName(glfw.Key(i), 0)] = glfw.Key(i)
		}
	}
	 */
}

func KeyvalName(keyval uint) string {
	return glfw.GetKeyName(glfw.Key(keyval), 0)
}

func KeyToUint(key string) uint {
	return uint(keymap[key])
}
