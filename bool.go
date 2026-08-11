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

// TakeBoolPtr reads a boolean value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeBoolPtr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*bool, unsafe.Pointer) {
	v, next := TakeBool(src, lim)
	p := (*bool)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}
