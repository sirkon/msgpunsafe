package msgpunsafe

import (
	"math/bits"
	"unsafe"
)

// TakeSliceHeader read Array header of Msgpack.
// Returns a number of elements in there and updated pointer.
func TakeSliceHeader(src unsafe.Pointer, lim unsafe.Pointer) (int, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorSliceExhausted)
	}

	lead := *(*byte)(src)

	// 1. fixarray: 0x90 - 0x9f (0-15 elements)
	if lead >= 0x90 && lead <= 0x9f {
		return int(lead & 0x0f), unsafe.Add(src, 1)
	}

	switch lead {
	// 3. array 16: 0xdc + 2 bytes of Big-Endian.
	case 0xdc:
		if uintptr(lim)-uintptr(src) < 3 {
			panicWithError(ErrorSliceHeaderTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	// 4. array 32: 0xdd + 4 bytes of Big-Endian
	case 0xdd:
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorSliceHeaderTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	default:
		// TODO probably check what these bytes are actually for?
		//      Like, ErrorSliceHeaderGotMap. Seems unnecessary at this stage though.
		panicWithError(ErrorSliceHeaderCorrupted)
	}

	return 0, nil
}

// TakeMapHeader read Map header of Msgpack.
// Returns a number of elements (key-value pairs) in there and updated pointer.
func TakeMapHeader(src unsafe.Pointer, lim unsafe.Pointer) (int, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorMapExhausted)
	}

	lead := *(*byte)(src)

	// 1. fixmap: 0x80 - 0x8f (0-15 key-value pairs)
	if lead >= 0x80 && lead <= 0x8f {
		return int(lead & 0x0f), unsafe.Add(src, 1)
	}

	switch lead {
	// 2. map 16: 0xde + 2 bytes of Big-Endian.
	case 0xde:
		if uintptr(lim)-uintptr(src) < 3 {
			panicWithError(ErrorMapHeaderTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	// 3. map 32: 0xdf + 4 bytes of Big-Endian.
	case 0xdf:
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorMapHeaderTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	default:
		panicWithError(ErrorMapHeaderCorrupted)
	}

	return 0, nil
}

// TakeStrHeader reads String header of Msgpack.
// Returns the size of the string in bytes and updated pointer.
func TakeStrHeader(src unsafe.Pointer, lim unsafe.Pointer) (int, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorStrExhausted)
	}

	lead := *(*byte)(src)

	// 1. fixstr: 0xa0 - 0xbf (0-31 bytes)
	if lead >= 0xa0 && lead <= 0xbf {
		return int(lead & 0x1f), unsafe.Add(src, 1)
	}

	switch lead {
	// 2. str 8: 0xd9 + 1 byte of size
	case 0xd9:
		if uintptr(lim)-uintptr(src) < 2 {
			panicWithError(ErrorStrHeaderTruncated)
		}
		return int(*(*uint8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	// 3. str 16: 0xda + 2 bytes of Big-Endian size
	case 0xda:
		if uintptr(lim)-uintptr(src) < 3 {
			panicWithError(ErrorStrHeaderTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	// 4. str 32: 0xdb + 4 bytes of Big-Endian size
	case 0xdb:
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorStrHeaderTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	default:
		panicWithError(ErrorStrHeaderCorrupted)
	}

	return 0, nil
}

// TakeBinHeader reads Binary header of Msgpack ([]byte data).
// Returns the size of the slice in bytes and updated pointer.
func TakeBinHeader(src unsafe.Pointer, lim unsafe.Pointer) (int, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorBinExhausted)
	}

	lead := *(*byte)(src)

	switch lead {
	// 1. bin 8: 0xc4 + 1 byte of size
	case 0xc4:
		if uintptr(lim)-uintptr(src) < 2 {
			panicWithError(ErrorBinHeaderTruncated)
		}
		return int(*(*uint8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	// 2. bin 16: 0xc5 + 2 bytes of Big-Endian size
	case 0xc5:
		if uintptr(lim)-uintptr(src) < 3 {
			panicWithError(ErrorBinHeaderTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	// 3. bin 32: 0xc6 + 4 bytes of Big-Endian size
	case 0xc6:
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorBinHeaderTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	default:
		panicWithError(ErrorBinHeaderCorrupted)
	}

	return 0, nil
}
