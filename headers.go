package msgpunsafe

import (
	"math/bits"
	"unsafe"
)

// TakeSliceHeader read Array header of Msgpack.
// Returns a number of elements in there and updated pointer.
func TakeSliceHeader(src unsafe.Pointer, lim int) (int, unsafe.Pointer) {
	if lim == 0 {
		panic(ErrorSliceExhausted)
	}

	lead := *(*byte)(src)

	// 1. fixarray: 0x90 - 0x9f (0-15 elements)
	if lead >= 0x90 && lead <= 0x9f {
		return int(lead & 0x0f), unsafe.Add(src, 1)
	}

	switch lead {
	// 3. array 16: 0xdc + 2 bytes of Big-Endian.
	case 0xdc:
		if lim < 3 {
			panic(ErrorSliceHeaderTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	// 4. array 32: 0xdd + 4 bytes of Big-Endian
	case 0xdd:
		if lim < 5 {
			panic(ErrorSliceHeaderTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	default:
		// TODO probably check what these bytes are actually for?
		//      Like, ErrorSliceHeaderGotMap. Seems unnecessary at this stage though.
		panic(ErrorSliceHeaderCorrupted)
	}
}

// TakeMapHeader read Map header of Msgpack.
// Returns a number of elements (key-value pairs) in there and updated pointer.
func TakeMapHeader(src unsafe.Pointer, lim int) (int, unsafe.Pointer) {
	if lim == 0 {
		panic(ErrorMapExhausted)
	}

	lead := *(*byte)(src)

	// 1. fixmap: 0x80 - 0x8f (0-15 key-value pairs)
	if lead >= 0x80 && lead <= 0x8f {
		return int(lead & 0x0f), unsafe.Add(src, 1)
	}

	switch lead {
	// 2. map 16: 0xde + 2 bytes of Big-Endian.
	case 0xde:
		if lim < 3 {
			panic(ErrorMapHeaderTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	// 3. map 32: 0xdf + 4 bytes of Big-Endian.
	case 0xdf:
		if lim < 5 {
			panic(ErrorMapHeaderTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	default:
		panic(ErrorMapHeaderCorrupted)
	}
}

// TakeStrHeader reads String header of Msgpack.
// Returns the size of the string in bytes and updated pointer.
func TakeStrHeader(src unsafe.Pointer, lim int) (int, unsafe.Pointer) {
	if lim == 0 {
		panic(ErrorStrExhausted)
	}

	lead := *(*byte)(src)

	// 1. fixstr: 0xa0 - 0xbf (0-31 bytes)
	if lead >= 0xa0 && lead <= 0xbf {
		return int(lead & 0x1f), unsafe.Add(src, 1)
	}

	switch lead {
	// 2. str 8: 0xd9 + 1 byte of size
	case 0xd9:
		if lim < 2 {
			panic(ErrorStrHeaderTruncated)
		}
		return int(*(*uint8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	// 3. str 16: 0xda + 2 bytes of Big-Endian size
	case 0xda:
		if lim < 3 {
			panic(ErrorStrHeaderTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	// 4. str 32: 0xdb + 4 bytes of Big-Endian size
	case 0xdb:
		if lim < 5 {
			panic(ErrorStrHeaderTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	default:
		panic(ErrorStrHeaderCorrupted)
	}
}

// TakeBinHeader reads Binary header of Msgpack ([]byte data).
// Returns the size of the slice in bytes and updated pointer.
func TakeBinHeader(src unsafe.Pointer, lim int) (int, unsafe.Pointer) {
	if lim == 0 {
		panic(ErrorBinExhausted)
	}

	lead := *(*byte)(src)

	switch lead {
	// 1. bin 8: 0xc4 + 1 byte of size
	case 0xc4:
		if lim < 2 {
			panic(ErrorBinHeaderTruncated)
		}
		return int(*(*uint8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	// 2. bin 16: 0xc5 + 2 bytes of Big-Endian size
	case 0xc5:
		if lim < 3 {
			panic(ErrorBinHeaderTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	// 3. bin 32: 0xc6 + 4 bytes of Big-Endian size
	case 0xc6:
		if lim < 5 {
			panic(ErrorBinHeaderTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	default:
		panic(ErrorBinHeaderCorrupted)
	}
}
