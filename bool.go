package msgpunsafe

import (
	"unsafe"
)

// TakeBool reads a boolean value from Msgpack.
// Returns the boolean value and an updated pointer.
func TakeBool(src unsafe.Pointer, lim int) (bool, unsafe.Pointer) {
	if lim == 0 {
		panic(ErrorBoolExhausted)
	}

	lead := *(*byte)(src)

	switch lead {
	case 0xc2: // false
		return false, unsafe.Add(src, 1)
	case 0xc3: // true
		return true, unsafe.Add(src, 1)
	default:
		panic(ErrorBoolCorrupted)
	}
}
