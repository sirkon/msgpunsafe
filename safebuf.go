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

	len int
}

// ptrAlignment is the alignment guarantees for AllocAligned.
const ptrAlignment = 8

// NewSafeBuffer creates safe buffer with preallocated bytes. Beware, you better
// not to use anything larger than 512 unless you are going to reuse it.
// And you better to make it smaller than 512 at that: the smaller the size
// the easier is to alloc.
func NewSafeBuffer(bufCap int) *SafeBuffer {
	return &SafeBuffer{
		buf: make([]byte, bufCap),
	}
}

// AllocString safely stores data referenced with pointer and the size as a string.
func (b *SafeBuffer) AllocString(data unsafe.Pointer, size int) string {
	base := unsafe.Pointer(unsafe.SliceData(b.buf))
	off := b.len

	if size+b.len > cap(b.buf) {
		if size > cap(b.buf) {
			// No trickery makes any sense here. Do it straight.
			return strings.Clone(unsafe.String((*byte)(data), size))
		}

		// Existing strings and slices passed through Alloc*  will still handle previous buffer.
		b.buf = make([]byte, cap(b.buf))
		base = unsafe.Pointer(unsafe.SliceData(b.buf))
		off = 0
		b.len = 0
	}

	copyNonOverlapping(unsafe.Add(base, off), data, size)
	b.len += size
	return unsafe.String((*byte)(unsafe.Add(base, off)), size)
}

// AllocBytes safely stores data referenced with pointer and the size as a []byte.
func (b *SafeBuffer) AllocBytes(data unsafe.Pointer, size int) []byte {
	base := unsafe.Pointer(unsafe.SliceData(b.buf))
	off := b.len

	if size+b.len > cap(b.buf) {
		if size > cap(b.buf) {
			// No trickery makes any sense here. Do it straight.
			return bytes.Clone(unsafe.Slice((*byte)(data), size))
		}

		// Existing strings and slices passed through Alloc*  will still handle previous buffer.
		b.buf = make([]byte, cap(b.buf))
		base = unsafe.Pointer(unsafe.SliceData(b.buf))
		off = 0
		b.len = 0
	}

	copyNonOverlapping(unsafe.Add(base, off), data, size)
	b.len += size
	return unsafe.Slice((*byte)(unsafe.Add(base, off)), size)
}

// AllocAligned reserves size bytes in the buffer and returns a pointer to the
// reserved slot aligned to ptrAlignment bytes. The caller must write the full
// value immediately after allocation; the padding bytes before the slot are
// never read.
//
// Freshly allocated backing arrays from make are already aligned to at least
// ptrAlignment, so no extra alignment is needed on the direct path or after a
// reset.
//
// When the buffer cannot hold the slot even after a reset, a dedicated
// allocation is made instead and b.buf is left untouched; previously returned
// pointers keep pointing into their own backing arrays and remain valid.
func (b *SafeBuffer) AllocAligned(size int) unsafe.Pointer {
	if size > cap(b.buf) {
		// No room in the buffer at all, not even after a reset. Do it straight.
		return unsafe.Pointer(unsafe.SliceData(make([]byte, size)))
	}

	for {
		base := unsafe.Pointer(unsafe.SliceData(b.buf))
		addr := (uintptr(base) + uintptr(b.len) + ptrAlignment - 1) &^ (ptrAlignment - 1)
		start := int(addr - uintptr(base))

		if start+size <= cap(b.buf) {
			b.len = start + size
			return unsafe.Add(base, start)
		}

		// Buffer exhausted: reset and retry. A fresh buffer is aligned and has
		// cap(b.buf) >= size, so the slot is guaranteed to fit and the loop
		// terminates after at most one reset.
		b.buf = make([]byte, cap(b.buf))
		b.len = 0
	}
}

// There can be troubles with int used for size. It is identical with uintptr
// from ABI standpoint for 64-bit architectures though, so let's be it.
//
//go:linkname copyNonOverlapping runtime.memmove
func copyNonOverlapping(dst, src unsafe.Pointer, size int)
