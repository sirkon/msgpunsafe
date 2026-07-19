package msgpunsafe

import (
	"bytes"
	"strings"
	"unsafe"
)

// SafeBuffer buffer for safe storage of strings and bytes for msgpack data from
// temporary buffer.
type SafeBuffer struct {
	buf []byte
}

// 512 is the point here, as it the largest size for the buffer with runtimize optimization for small
// buffers. 513+ will be slower.
const bufCap = 512

// NewSafeBuffer creates safe buffer with preallocated bytes.
func NewSafeBuffer() *SafeBuffer {
	return &SafeBuffer{
		buf: make([]byte, 0, bufCap),
	}
}

// AllocString safely stores data referenced with pointer and the size as a string.
func (b *SafeBuffer) AllocString(data unsafe.Pointer, size int) string {
	off := len(b.buf)

	if size+len(b.buf) > bufCap {
		if size > bufCap {
			// No trickery makes any sense here. Do it straight.
			return strings.Clone(unsafe.String((*byte)(data), size))
		}

		// Existing strings and slices passed through Alloc*  will still handle previous buffer.
		b.buf = make([]byte, 0, bufCap)
	}

	b.buf = append(b.buf, unsafe.Slice((*byte)(data), size)...)
	return unsafe.String(
		(*byte)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b.buf)), off)),
		size,
	)
}

// AllocBytes safely stores data referenced with pointer and the size as a []byte.
func (b *SafeBuffer) AllocBytes(data unsafe.Pointer, size int) []byte {
	off := len(b.buf)

	if size+len(b.buf) > bufCap {
		if size > bufCap {
			// No trickery makes any sense here. Do it straight.
			return bytes.Clone(unsafe.Slice((*byte)(data), size))
		}

		// Existing strings and slices passed through Alloc*  will still handle previous buffer.
		b.buf = make([]byte, 0, bufCap)
	}

	b.buf = append(b.buf, unsafe.Slice((*byte)(data), size)...)
	return unsafe.Slice(
		(*byte)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b.buf)), off)),
		size,
	)
}
