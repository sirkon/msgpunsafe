//go:build aware_of_alignment

package msgpunsafe

import (
	"encoding/binary"
	"math"
	"unsafe"
)

// takeUint64 reads any valid unsigned integer from Msgpack and returns it as uint64.
// It supports positive fixint, uint8, uint16, uint32, and uint64.
func takeUint64(src unsafe.Pointer, lim unsafe.Pointer) (uint64, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorNumberExhausted)
	}

	lead := *(*byte)(src)

	// 1. Positive Fixint: 0x00 - 0x7f (0 to 127)
	if lead <= 0x7f {
		return uint64(lead), unsafe.Add(src, 1)
	}

	// 2. Negative Fixint: 0xe0 - 0xff (-32 to -1)
	// Unsigned integers cannot be negative, so this is a guaranteed overflow.
	if lead >= 0xe0 {
		panicWithError(ErrorUintOverflow)
	}

	switch lead {
	// --- Standard unsigned markers ---
	case 0xcc: // uint 8
		if uintptr(lim)-uintptr(src) < 2 {
			panicWithError(ErrorNumberTruncated)
		}
		return uint64(*(*uint8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	case 0xcd: // uint 16
		if uintptr(lim)-uintptr(src) < 3 {
			panicWithError(ErrorNumberTruncated)
		}
		val := binary.BigEndian.Uint16(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 2))
		return uint64(val), unsafe.Add(src, 3)

	case 0xce: // uint 32
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorNumberTruncated)
		}
		val := binary.BigEndian.Uint32(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 4))
		return uint64(val), unsafe.Add(src, 5)

	case 0xcf: // uint 64
		if uintptr(lim)-uintptr(src) < 9 {
			panicWithError(ErrorNumberTruncated)
		}
		val := binary.BigEndian.Uint64(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 8))
		return val, unsafe.Add(src, 9)

	// --- Compatibility: parse signed ints that are actually positive ---
	case 0xd0: // int 8
		if uintptr(lim)-uintptr(src) < 2 {
			panicWithError(ErrorNumberTruncated)
		}
		val := *(*int8)(unsafe.Add(src, 1))
		if val < 0 {
			panicWithError(ErrorUintOverflow)
		}
		return uint64(val), unsafe.Add(src, 2)

	case 0xd1: // int 16
		if uintptr(lim)-uintptr(src) < 3 {
			panicWithError(ErrorNumberTruncated)
		}
		val := int16(binary.BigEndian.Uint16(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 2)))
		if val < 0 {
			panicWithError(ErrorUintOverflow)
		}
		return uint64(val), unsafe.Add(src, 3)

	case 0xd2: // int 32
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorNumberTruncated)
		}
		val := int32(binary.BigEndian.Uint32(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 4)))
		if val < 0 {
			panicWithError(ErrorUintOverflow)
		}
		return uint64(val), unsafe.Add(src, 5)

	case 0xd3: // int 64
		if uintptr(lim)-uintptr(src) < 9 {
			panicWithError(ErrorNumberTruncated)
		}
		val := int64(binary.BigEndian.Uint64(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 8)))
		if val < 0 {
			panicWithError(ErrorUintOverflow)
		}
		return uint64(val), unsafe.Add(src, 9)

	default:
		panicWithError(ErrorUintCorrupted)
	}

	return 0, nil
}

// takeInt64 reads any valid signed integer from Msgpack and returns it as int64.
// It supports positive/negative fixint, int8, int16, int32, and int64.
func takeInt64(src unsafe.Pointer, lim unsafe.Pointer) (int64, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorNumberExhausted)
	}

	lead := *(*byte)(src)

	// 1. Positive Fixint: 0x00 - 0x7f (0 to 127)
	if lead <= 0x7f {
		return int64(lead), unsafe.Add(src, 1)
	}

	// 2. Negative Fixint: 0xe0 - 0xff (-32 to -1)
	if lead >= 0xe0 {
		return int64(int8(lead)), unsafe.Add(src, 1)
	}

	switch lead {
	// --- Standard signed markers ---
	case 0xd0: // int 8
		if uintptr(lim)-uintptr(src) < 2 {
			panicWithError(ErrorNumberTruncated)
		}
		return int64(*(*int8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	case 0xd1: // int 16
		if uintptr(lim)-uintptr(src) < 3 {
			panicWithError(ErrorNumberTruncated)
		}
		val := binary.BigEndian.Uint16(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 2))
		return int64(int16(val)), unsafe.Add(src, 3)

	case 0xd2: // int 32
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorNumberTruncated)
		}
		val := binary.BigEndian.Uint32(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 4))
		return int64(int32(val)), unsafe.Add(src, 5)

	case 0xd3: // int 64
		if uintptr(lim)-uintptr(src) < 9 {
			panicWithError(ErrorNumberTruncated)
		}
		val := binary.BigEndian.Uint64(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 8))
		return int64(val), unsafe.Add(src, 9)

	// --- Compatibility: parse uints that arrive in place of ints ---
	case 0xcc: // uint 8
		if uintptr(lim)-uintptr(src) < 2 {
			panicWithError(ErrorNumberTruncated)
		}
		return int64(*(*uint8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	case 0xcd: // uint 16
		if uintptr(lim)-uintptr(src) < 3 {
			panicWithError(ErrorNumberTruncated)
		}
		val := binary.BigEndian.Uint16(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 2))
		return int64(val), unsafe.Add(src, 3)

	case 0xce: // uint 32
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorNumberTruncated)
		}
		val := binary.BigEndian.Uint32(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 4))
		return int64(val), unsafe.Add(src, 5)

	case 0xcf: // uint 64
		if uintptr(lim)-uintptr(src) < 9 {
			panicWithError(ErrorNumberTruncated)
		}
		val := binary.BigEndian.Uint64(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 8))

		if val > math.MaxInt64 {
			panicWithError(ErrorIntOverflow)
		}
		return int64(val), unsafe.Add(src, 9)

	default:
		panicWithError(ErrorIntCorrupted)
	}

	return 0, nil
}

func takeFloat32(src unsafe.Pointer, lim unsafe.Pointer) (float32, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorNumberExhausted)
	}
	lead := *(*byte)(src)

	switch lead {
	case 0xca: // float 32
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorNumberTruncated)
		}
		bits32 := binary.BigEndian.Uint32(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 4))
		return math.Float32frombits(bits32), unsafe.Add(src, 5)

	case 0xcb: // float 64
		if uintptr(lim)-uintptr(src) < 9 {
			panicWithError(ErrorNumberTruncated)
		}
		bits64 := binary.BigEndian.Uint64(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 8))
		return float32(math.Float64frombits(bits64)), unsafe.Add(src, 9)

	default:
		panicWithError(ErrorFloatCorrupted)
	}

	return 0, nil
}

// takeFloat64 works strictly within its own precision
func takeFloat64(src unsafe.Pointer, lim unsafe.Pointer) (float64, unsafe.Pointer) {
	if uintptr(src) >= uintptr(lim) {
		panicWithError(ErrorNumberExhausted)
	}
	lead := *(*byte)(src)

	switch lead {
	case 0xca: // float 32
		if uintptr(lim)-uintptr(src) < 5 {
			panicWithError(ErrorNumberTruncated)
		}
		bits32 := binary.BigEndian.Uint32(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 4))
		return float64(math.Float32frombits(bits32)), unsafe.Add(src, 5)

	case 0xcb: // float 64
		if uintptr(lim)-uintptr(src) < 9 {
			panicWithError(ErrorNumberTruncated)
		}
		bits64 := binary.BigEndian.Uint64(unsafe.Slice((*byte)(unsafe.Add(src, 1)), 8))
		return math.Float64frombits(bits64), unsafe.Add(src, 9)

	default:
		panicWithError(ErrorFloatCorrupted)
	}

	return 0, nil
}
