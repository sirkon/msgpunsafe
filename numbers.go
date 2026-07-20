package msgpunsafe

import (
	"unsafe"
)

// TakeUint reads numeric value uinto uint.
func TakeUint(src unsafe.Pointer, lim unsafe.Pointer) (uint, unsafe.Pointer) {
	res, next := takeUint64(src, lim)

	v := uint(res)
	if uint64(v) != res {
		panicWithError(ErrorUintOverflow)
	}

	return v, next
}

// TakeUint8 reads numeric value uinto uint8.
func TakeUint8(src unsafe.Pointer, lim unsafe.Pointer) (uint8, unsafe.Pointer) {
	res, next := takeUint64(src, lim)

	v := uint8(res)
	if uint64(v) != res {
		panicWithError(ErrorUintOverflow)
	}

	return v, next
}

// TakeUint16 reads numeric value uinto uint16.
func TakeUint16(src unsafe.Pointer, lim unsafe.Pointer) (uint16, unsafe.Pointer) {
	res, next := takeUint64(src, lim)

	v := uint16(res)
	if uint64(v) != res {
		panicWithError(ErrorUintOverflow)
	}

	return v, next
}

// TakeUint32 reads numeric value uinto uint32.
func TakeUint32(src unsafe.Pointer, lim unsafe.Pointer) (uint32, unsafe.Pointer) {
	res, next := takeUint64(src, lim)

	v := uint32(res)
	if uint64(v) != res {
		panicWithError(ErrorUintOverflow)
	}

	return v, next
}

// TakeUint64 reads numeric value uinto uint64.
func TakeUint64(src unsafe.Pointer, lim unsafe.Pointer) (uint64, unsafe.Pointer) {
	return takeUint64(src, lim)
}

// TakeInt reads numeric value into int.
func TakeInt(src unsafe.Pointer, lim unsafe.Pointer) (int, unsafe.Pointer) {
	res, next := takeInt64(src, lim)

	v := int(res)
	if int64(v) != res {
		panicWithError(ErrorIntOverflow)
	}

	return v, next
}

// TakeInt8 reads numeric value into int8.
func TakeInt8(src unsafe.Pointer, lim unsafe.Pointer) (int8, unsafe.Pointer) {
	res, next := takeInt64(src, lim)

	v := int8(res)
	if int64(v) != res {
		panicWithError(ErrorIntOverflow)
	}

	return v, next
}

// TakeInt16 reads numeric value into int16.
func TakeInt16(src unsafe.Pointer, lim unsafe.Pointer) (int16, unsafe.Pointer) {
	res, next := takeInt64(src, lim)

	v := int16(res)
	if int64(v) != res {
		panicWithError(ErrorIntOverflow)
	}

	return v, next
}

// TakeInt32 reads numeric value into int32.
func TakeInt32(src unsafe.Pointer, lim unsafe.Pointer) (int32, unsafe.Pointer) {
	res, next := takeInt64(src, lim)

	v := int32(res)
	if int64(v) != res {
		panicWithError(ErrorIntOverflow)
	}

	return v, next
}

// TakeInt64 reads numeric value into int64.
func TakeInt64(src unsafe.Pointer, lim unsafe.Pointer) (int64, unsafe.Pointer) {
	return takeInt64(src, lim)
}

// TakeFloat32 reads numeric value into float32.
func TakeFloat32(src unsafe.Pointer, lim unsafe.Pointer) (float32, unsafe.Pointer) {
	return takeFloat32(src, lim)
}

// TakeFloat64 reads numeric value into float64.
func TakeFloat64(src unsafe.Pointer, lim unsafe.Pointer) (float64, unsafe.Pointer) {
	return takeFloat64(src, lim)
}
