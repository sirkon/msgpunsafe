package msgpunsafe

import (
	"unsafe"
)

// TakeBool reads a boolean value from Msgpack.
// Returns the boolean value and an updated pointer.
func TakeBool(src unsafe.Pointer, lim unsafe.Pointer) (bool, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorBoolExhausted)
	}

	lead := *(*byte)(src)

	switch lead {
	case 0xc2: // false
		return false, unsafe.Add(src, 1)
	case 0xc3: // true
		return true, unsafe.Add(src, 1)
	default:
		panicWithError(ErrorBoolCorrupted)
	}

	return false, nil
}
