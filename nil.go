package msgpunsafe

import (
	"unsafe"
)

// IsNil checks if the current Msgpack value is nil (marker 0xc0).
// If it is nil, it returns true and shifts the pointer by 1 byte.
// If it is not nil, it returns false and the original untouched pointer.
func IsNil(src unsafe.Pointer, lim unsafe.Pointer) (bool, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorNumberExhausted) // or reuse any generic exhaustion error
	}

	if *(*byte)(src) == 0xc0 {
		return true, unsafe.Add(src, 1)
	}

	return false, src
}
