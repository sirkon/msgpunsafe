package msgpunsafe

import (
	"unsafe"
)

// TakeIntPtr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeIntPtr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*int, unsafe.Pointer) {
	v, next := TakeInt(src, lim)
	p := (*int)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeInt8Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeInt8Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*int8, unsafe.Pointer) {
	v, next := TakeInt8(src, lim)
	p := (*int8)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeInt16Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeInt16Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*int16, unsafe.Pointer) {
	v, next := TakeInt16(src, lim)
	p := (*int16)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeInt32Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeInt32Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*int32, unsafe.Pointer) {
	v, next := TakeInt32(src, lim)
	p := (*int32)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeInt64Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeInt64Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*int64, unsafe.Pointer) {
	v, next := TakeInt64(src, lim)
	p := (*int64)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeUintPtr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeUintPtr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*uint, unsafe.Pointer) {
	v, next := TakeUint(src, lim)
	p := (*uint)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeUint8Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeUint8Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*uint8, unsafe.Pointer) {
	v, next := TakeUint8(src, lim)
	p := (*uint8)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeUint16Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeUint16Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*uint16, unsafe.Pointer) {
	v, next := TakeUint16(src, lim)
	p := (*uint16)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeUint32Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeUint32Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*uint32, unsafe.Pointer) {
	v, next := TakeUint32(src, lim)
	p := (*uint32)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeUint64Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeUint64Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*uint64, unsafe.Pointer) {
	v, next := TakeUint64(src, lim)
	p := (*uint64)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeFloat32Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeFloat32Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*float32, unsafe.Pointer) {
	v, next := TakeFloat32(src, lim)
	p := (*float32)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}

// TakeFloat64Ptr reads a numeric value, stores it in the SafeBuffer and returns an
// 8-byte-aligned pointer to it.
func TakeFloat64Ptr(src unsafe.Pointer, lim unsafe.Pointer, sBuf *SafeBuffer) (*float64, unsafe.Pointer) {
	v, next := TakeFloat64(src, lim)
	p := (*float64)(sBuf.AllocAligned(int(unsafe.Sizeof(v))))
	*p = v
	return p, next
}
