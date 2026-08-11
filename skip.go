package msgpunsafe

import (
	"math/bits"
	"unsafe"
)

// skipRecursionLimit is the maximum nesting depth of msgpack containers
// TakeSkip is willing to traverse before panicking with ErrorSkipRecursion.
// The value matches tinylib/msgp for behavioral compatibility.
const skipRecursionLimit = 100000

// TakeSkip skips one complete msgpack object at src, including all elements of
// maps and arrays, and returns the pointer right after the object.
//
// The reader allocates nothing and does not modify the source buffer.
func TakeSkip(src unsafe.Pointer, lim unsafe.Pointer) unsafe.Pointer {
	// Each stack entry is the number of sibling objects left to process at
	// that depth level; a container replaces its own slot with the counts of
	// its children. A map counts twice per pair (key and value).
	stack := []uint64{1}

	for len(stack) > 0 {
		top := len(stack) - 1
		if stack[top] == 0 {
			// This level is complete; continue at the parent.
			stack = stack[:top]
			continue
		}
		stack[top]--

		advance, payload, children, err := takeSkipSize(src, lim)
		if err != 0 {
			panicWithError(err)
		}

		if uintptr(lim)-uintptr(src) < uintptr(advance) {
			panicWithError(ErrorSkipHeaderTruncated)
		}
		src = unsafe.Add(src, int(advance))

		if uintptr(lim)-uintptr(src) < uintptr(payload) {
			panicWithError(ErrorSkipLenConflict)
		}
		src = unsafe.Add(src, int(payload))

		if children > 0 {
			if len(stack) >= skipRecursionLimit {
				panicWithError(ErrorSkipRecursion)
			}
			stack = append(stack, children)
		}
	}

	return src
}

// takeSkipSize inspects the lead byte at src, returning the number of header
// bytes to advance, the payload length, and the number of child objects to
// skip. A non-zero error code means the marker or its header is invalid.
func takeSkipSize(src unsafe.Pointer, lim unsafe.Pointer) (advance, payload, children uint64, err ErrorCode) {
	if uintptr(src) >= uintptr(lim) {
		return 0, 0, 0, ErrorSkipExhausted
	}

	lead := *(*byte)(src)

	// Fixed maps, arrays and strings are entirely determined by the lead byte.
	switch {
	case lead >= 0x80 && lead <= 0x8f: // fixmap 0-15 pairs
		return 1, 0, uint64(lead&0x0f) * 2, 0
	case lead >= 0x90 && lead <= 0x9f: // fixarray 0-15 elements
		return 1, 0, uint64(lead & 0x0f), 0
	case lead >= 0xa0 && lead <= 0xbf: // fixstr 0-31 bytes
		return 1, uint64(lead & 0x1f), 0, 0
	case lead <= 0x7f || lead >= 0xe0: // positive/negative fixint
		return 1, 0, 0, 0
	}

	var header uint64
	switch lead {
	// nil and booleans
	case 0xc0, 0xc2, 0xc3:
		return 1, 0, 0, 0

	// floats
	case 0xca: // float 32
		return 5, 0, 0, 0
	case 0xcb: // float 64
		return 9, 0, 0, 0

	// unsigned integers
	case 0xcc, 0xd0: // uint 8 / int 8
		return 2, 0, 0, 0
	case 0xcd, 0xd1: // uint 16 / int 16
		return 3, 0, 0, 0
	case 0xce, 0xd2: // uint 32 / int 32
		return 5, 0, 0, 0
	case 0xcf, 0xd3: // uint 64 / int 64
		return 9, 0, 0, 0

	// bins
	case 0xc4: // bin 8
		if uintptr(lim)-uintptr(src) < 2 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		return 2, uint64(*(*uint8)(unsafe.Add(src, 1))), 0, 0
	case 0xc5: // bin 16
		if uintptr(lim)-uintptr(src) < 3 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		return 3, uint64(bits.ReverseBytes16(*(*uint16)(unsafe.Add(src, 1)))), 0, 0
	case 0xc6: // bin 32
		if uintptr(lim)-uintptr(src) < 5 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		return 5, uint64(bits.ReverseBytes32(*(*uint32)(unsafe.Add(src, 1)))), 0, 0

	// extensions
	case 0xd4: // fixext 1
		return 2, 1, 0, 0
	case 0xd5: // fixext 2
		return 2, 2, 0, 0
	case 0xd6: // fixext 4
		return 2, 4, 0, 0
	case 0xd7: // fixext 8
		return 2, 8, 0, 0
	case 0xd8: // fixext 16
		return 2, 16, 0, 0
	case 0xc7: // ext 8
		if uintptr(lim)-uintptr(src) < 3 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		return 3, uint64(*(*uint8)(unsafe.Add(src, 1))), 0, 0
	case 0xc8: // ext 16
		if uintptr(lim)-uintptr(src) < 4 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		return 4, uint64(bits.ReverseBytes16(*(*uint16)(unsafe.Add(src, 1)))), 0, 0
	case 0xc9: // ext 32
		if uintptr(lim)-uintptr(src) < 6 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		return 6, uint64(bits.ReverseBytes32(*(*uint32)(unsafe.Add(src, 1)))), 0, 0

	// strings
	case 0xd9: // str 8
		if uintptr(lim)-uintptr(src) < 2 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		return 2, uint64(*(*uint8)(unsafe.Add(src, 1))), 0, 0
	case 0xda: // str 16
		if uintptr(lim)-uintptr(src) < 3 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		return 3, uint64(bits.ReverseBytes16(*(*uint16)(unsafe.Add(src, 1)))), 0, 0
	case 0xdb: // str 32
		if uintptr(lim)-uintptr(src) < 5 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		return 5, uint64(bits.ReverseBytes32(*(*uint32)(unsafe.Add(src, 1)))), 0, 0

	// arrays
	case 0xdc: // array 16
		if uintptr(lim)-uintptr(src) < 3 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		header = uint64(bits.ReverseBytes16(*(*uint16)(unsafe.Add(src, 1))))
		return 3, 0, header, 0
	case 0xdd: // array 32
		if uintptr(lim)-uintptr(src) < 5 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		header = uint64(bits.ReverseBytes32(*(*uint32)(unsafe.Add(src, 1))))
		return 5, 0, header, 0

	// maps
	case 0xde: // map 16
		if uintptr(lim)-uintptr(src) < 3 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		header = uint64(bits.ReverseBytes16(*(*uint16)(unsafe.Add(src, 1))))
		return 3, 0, header * 2, 0
	case 0xdf: // map 32
		if uintptr(lim)-uintptr(src) < 5 {
			return 0, 0, 0, ErrorSkipHeaderTruncated
		}
		header = uint64(bits.ReverseBytes32(*(*uint32)(unsafe.Add(src, 1))))
		return 5, 0, header * 2, 0

	default:
		// 0xc1 is the only unassigned lead byte left.
		return 0, 0, 0, ErrorSkipCorrupted
	}
}
