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

// ptrAlignment is the alignment guarantees for AllocAligned.
const ptrAlignment = 8

// NewSafeBuffer creates safe buffer with preallocated bytes. Beware, you better
// not to use anything larger than 512 unless you are going to reuse it.
// And you better to make it smaller than 512 at that: the smaller the size
// the easier is to alloc.
func NewSafeBuffer(bufCap int) *SafeBuffer {
	return &SafeBuffer{
		buf: make([]byte, 0, bufCap),
	}
}

// AllocString safely stores data referenced with pointer and the size as a string.
func (b *SafeBuffer) AllocString(data unsafe.Pointer, size int) string {
	off := len(b.buf)

	if size+len(b.buf) > cap(b.buf) {
		if size > cap(b.buf) {
			// No trickery makes any sense here. Do it straight.
			return strings.Clone(unsafe.String((*byte)(data), size))
		}

		b.buf = make([]byte, 0, cap(b.buf))
		off = 0
	}

	// Existing strings and slices passed through Alloc*  will still handle previous buffer.
	b.buf = append(b.buf, unsafe.Slice((*byte)(data), size)...)
	return unsafe.String(
		(*byte)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b.buf)), off)),
		size,
	)
}

// AllocBytes safely stores data referenced with pointer and the size as a []byte.
func (b *SafeBuffer) AllocBytes(data unsafe.Pointer, size int) []byte {
	off := len(b.buf)

	if size+len(b.buf) > cap(b.buf) {
		if size > cap(b.buf) {
			// No trickery makes any sense here. Do it straight.
			return bytes.Clone(unsafe.Slice((*byte)(data), size))
		}

		// Existing strings and slices passed through Alloc*  will still handle previous buffer.
		b.buf = make([]byte, 0, cap(b.buf))
		off = 0
	}

	b.buf = append(b.buf, unsafe.Slice((*byte)(data), size)...)
	return unsafe.Slice(
		(*byte)(unsafe.Add(unsafe.Pointer(unsafe.SliceData(b.buf)), off)),
		size,
	)
}

// AllocAligned reserves size bytes in the buffer and returns a pointer to the
// reserved slot aligned to ptrAlignment bytes. The caller must write the full
// value immediately after allocation; the padding bytes before the slot are
// never read.
//
// When the buffer cannot hold the slot, a dedicated allocation is made instead
// and b.buf is left untouched; previously returned pointers keep pointing into
// their own backing arrays and remain valid.
func (b *SafeBuffer) AllocAligned(size int) unsafe.Pointer {
	if size+ptrAlignment-1 > cap(b.buf) {
		// No room in the buffer at all, not even after a reset. Do it straight.
		raw := make([]byte, size+ptrAlignment)
		base := unsafe.Pointer(unsafe.SliceData(raw))
		pad := (ptrAlignment - uintptr(base)%ptrAlignment) % ptrAlignment
		return unsafe.Add(base, int(pad))
	}

	var zeros [ptrAlignment]byte

	for {
		off := len(b.buf)
		base := unsafe.Pointer(unsafe.SliceData(b.buf))
		addr := uintptr(base) + uintptr(off)
		pad := (ptrAlignment - addr%ptrAlignment) % ptrAlignment

		if uintptr(off)+pad+uintptr(size) <= uintptr(cap(b.buf)) {
			b.buf = append(b.buf, zeros[:pad]...)
			start := len(b.buf)
			b.buf = b.buf[:start+size]
			return unsafe.Add(base, int(uintptr(off)+pad))
		}

		// Buffer exhausted: reset and retry. A fresh buffer has cap(b.buf) >=
		// size+ptrAlignment-1, so the aligned slot is guaranteed to fit and the
		// loop terminates after at most one reset.
		b.buf = make([]byte, 0, cap(b.buf))
	}
}
