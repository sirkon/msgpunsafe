package msgpunsafe

import (
	"math"
	"math/bits"
	"unsafe"
)

// takeUint64 reads any valid unsigned integer from Msgpack and returns it as uint64.
// It supports positive fixint, uint8, uint16, uint32, and uint64.
func takeUint64(src unsafe.Pointer, lim int) (uint64, unsafe.Pointer) {
	lead := *(*byte)(src)

	// 1. Positive Fixint: 0x00 - 0x7f (0 to 127)
	if lead <= 0x7f {
		return uint64(lead), unsafe.Add(src, 1)
	}

	// 2. Negative Fixint: 0xe0 - 0xff (-32 to -1)
	// Unsigned integers cannot be negative, so this is a guaranteed overflow.
	if lead >= 0xe0 {
		panic(ErrorUintOverflow)
	}

	switch lead {
	// --- Standard unsigned markers ---
	case 0xcc: // uint 8
		if lim < 2 {
			panic(ErrorNumberTruncated)
		}
		return uint64(*(*uint8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	case 0xcd: // uint 16
		if lim < 3 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return uint64(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	case 0xce: // uint 32
		if lim < 5 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return uint64(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	case 0xcf: // uint 64
		if lim < 9 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint64)(unsafe.Add(src, 1))
		return bits.ReverseBytes64(raw), unsafe.Add(src, 9)

	// --- Compatibility: parse signed ints that are actually positive ---
	case 0xd0: // int 8
		if lim < 2 {
			panic(ErrorNumberTruncated)
		}
		val := *(*int8)(unsafe.Add(src, 1))
		if val < 0 {
			panic(ErrorUintOverflow)
		}
		return uint64(val), unsafe.Add(src, 2)

	case 0xd1: // int 16
		if lim < 3 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		val := int16(bits.ReverseBytes16(raw))
		if val < 0 {
			panic(ErrorUintOverflow)
		}
		return uint64(val), unsafe.Add(src, 3)

	case 0xd2: // int 32
		if lim < 5 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		val := int32(bits.ReverseBytes32(raw))
		if val < 0 {
			panic(ErrorUintOverflow)
		}
		return uint64(val), unsafe.Add(src, 5)

	case 0xd3: // int 64
		if lim < 9 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint64)(unsafe.Add(src, 1))
		val := int64(bits.ReverseBytes64(raw))
		if val < 0 {
			panic(ErrorUintOverflow)
		}
		return uint64(val), unsafe.Add(src, 9)

	default:
		panic(ErrorUintCorrupted)
	}
}

// takeInt64 reads any valid signed integer from Msgpack and returns it as int64.
// It supports positive/negative fixint, int8, int16, int32, and int64.
func takeInt64(src unsafe.Pointer, lim int) (int64, unsafe.Pointer) {
	lead := *(*byte)(src)

	// 1. Positive Fixint: 0x00 - 0x7f (0 до 127)
	if lead <= 0x7f {
		return int64(lead), unsafe.Add(src, 1)
	}

	// 2. Negative Fixint: 0xe0 - 0xff (-32 до -1)
	if lead >= 0xe0 {
		return int64(int8(lead)), unsafe.Add(src, 1)
	}

	switch lead {
	// --- Стандартные знаковые маркеры ---
	case 0xd0: // int 8
		if lim < 2 {
			panic(ErrorNumberTruncated)
		}
		return int64(*(*int8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	case 0xd1: // int 16
		if lim < 3 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int64(int16(bits.ReverseBytes16(raw))), unsafe.Add(src, 3)

	case 0xd2: // int 32
		if lim < 5 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int64(int32(bits.ReverseBytes32(raw))), unsafe.Add(src, 5)

	case 0xd3: // int 64
		if lim < 9 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint64)(unsafe.Add(src, 1))
		return int64(bits.ReverseBytes64(raw)), unsafe.Add(src, 9)

	// --- Совместимость: парсим uint-ы, которые прилетают вместо int-ов ---
	case 0xcc: // uint 8
		if lim < 2 {
			panic(ErrorNumberTruncated)
		}
		return int64(*(*uint8)(unsafe.Add(src, 1))), unsafe.Add(src, 2)

	case 0xcd: // uint 16
		if lim < 3 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint16)(unsafe.Add(src, 1))
		return int64(bits.ReverseBytes16(raw)), unsafe.Add(src, 3)

	case 0xce: // uint 32
		if lim < 5 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		return int64(bits.ReverseBytes32(raw)), unsafe.Add(src, 5)

	case 0xcf: // uint 64
		if lim < 9 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint64)(unsafe.Add(src, 1))
		val := bits.ReverseBytes64(raw)

		if val > math.MaxInt64 {
			panic(ErrorIntOverflow)
		}
		return int64(val), unsafe.Add(src, 9)

	default:
		panic(ErrorIntCorrupted)
	}
}

func takeFloat32(src unsafe.Pointer, lim int) (float32, unsafe.Pointer) {
	if lim == 0 {
		panic(ErrorNumberExhausted)
	}
	lead := *(*byte)(src)

	switch lead {
	case 0xca: // float 32
		if lim < 5 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		bits32 := bits.ReverseBytes32(raw)
		return math.Float32frombits(bits32), unsafe.Add(src, 5)

	case 0xcb: // float 64
		if lim < 9 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint64)(unsafe.Add(src, 1))
		bits64 := bits.ReverseBytes64(raw)
		return float32(math.Float64frombits(bits64)), unsafe.Add(src, 9)

	default:
		panic(ErrorFloatCorrupted)
	}
}

// takeFloat64 работает строго со своей разрядностью
func takeFloat64(src unsafe.Pointer, lim int) (float64, unsafe.Pointer) {
	if lim == 0 {
		panic(ErrorNumberExhausted)
	}
	lead := *(*byte)(src)

	switch lead {
	case 0xca: // адщфе32
		if lim < 5 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint32)(unsafe.Add(src, 1))
		bits32 := bits.ReverseBytes32(raw)
		return float64(math.Float32frombits(bits32)), unsafe.Add(src, 5)

	case 0xcb: // float 64
		if lim < 9 {
			panic(ErrorNumberTruncated)
		}
		raw := *(*uint64)(unsafe.Add(src, 1))
		bits64 := bits.ReverseBytes64(raw)
		return math.Float64frombits(bits64), unsafe.Add(src, 9)

	default:
		panic(ErrorFloatCorrupted)
	}
}
