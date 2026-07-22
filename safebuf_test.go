package msgpunsafe

import (
	"bytes"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

// bufPtrRange returns the [start, end) address range of the SafeBuffer's
// current backing array.
func bufPtrRange(b *SafeBuffer) (uintptr, uintptr) {
	start := uintptr(unsafe.Pointer(unsafe.SliceData(b.buf)))
	return start, start + uintptr(cap(b.buf))
}

// ptrInBuf reports whether p points inside the SafeBuffer's current backing array.
func ptrInBuf(b *SafeBuffer, p unsafe.Pointer) bool {
	start, end := bufPtrRange(b)
	up := uintptr(p)
	return up >= start && up < end
}

// AllocString with a payload that fits into the buffer must store the bytes
// inside the SafeBuffer's backing array (no clone, no independent allocation).
func TestSafeBuffer_AllocString_InBuffer(t *testing.T) {
	sBuf := NewSafeBuffer(64)

	data := []byte("hello world")
	s := sBuf.AllocString(unsafe.Pointer(&data[0]), len(data))

	if s != "hello world" {
		t.Fatalf("got %q, want %q", s, "hello world")
	}
	if !ptrInBuf(sBuf, unsafe.Pointer(unsafe.StringData(s))) {
		t.Fatalf("string data must point into the safe buffer's backing array")
	}
	if len(sBuf.buf) != len(data) {
		t.Fatalf("buf len = %d, want %d", len(sBuf.buf), len(data))
	}
}

// AllocString with a payload larger than the buffer capacity must fall back to an
// immediate strings.Clone — an independent allocation that does not touch the buffer.
func TestSafeBuffer_AllocString_ImmediateClone(t *testing.T) {
	sBuf := NewSafeBuffer(8)

	payload := bytes.Repeat([]byte("abcdefgh"), 4) // 32 bytes > cap 8
	s := sBuf.AllocString(unsafe.Pointer(&payload[0]), len(payload))

	if s != string(payload) {
		t.Fatalf("content mismatch: got %q, want %q", s, string(payload))
	}
	if ptrInBuf(sBuf, unsafe.Pointer(unsafe.StringData(s))) {
		t.Fatalf("expected an independent clone, but data points into the safe buffer")
	}
	if len(sBuf.buf) != 0 {
		t.Fatalf("buffer must stay untouched on the clone path, len = %d", len(sBuf.buf))
	}
}

// AllocBytes with a payload larger than the buffer capacity must fall back to an
// immediate bytes.Clone — an independent allocation that does not touch the buffer.
func TestSafeBuffer_AllocBytes_ImmediateClone(t *testing.T) {
	sBuf := NewSafeBuffer(8)

	payload := bytes.Repeat([]byte("ABCDEFGH"), 4) // 32 bytes > cap 8
	bs := sBuf.AllocBytes(unsafe.Pointer(&payload[0]), len(payload))

	if !bytes.Equal(bs, payload) {
		t.Fatalf("content mismatch: got %v, want %v", bs, payload)
	}
	if ptrInBuf(sBuf, unsafe.Pointer(unsafe.SliceData(bs))) {
		t.Fatalf("expected an independent clone, but data points into the safe buffer")
	}
	if len(sBuf.buf) != 0 {
		t.Fatalf("buffer must stay untouched on the clone path, len = %d", len(sBuf.buf))
	}
}

// When the buffer is exhausted by a payload that still fits the capacity on its
// own (size <= cap, but size+len(buf) > cap), a fresh backing array is allocated
// and the cursor resets to 0. Earlier allocations keep pointing into the old one.
func TestSafeBuffer_BufferExhaustion_NewBuffer(t *testing.T) {
	sBuf := NewSafeBuffer(16)

	first := []byte("0123456789") // 10 bytes, fits, len(buf) becomes 10
	s1 := sBuf.AllocString(unsafe.Pointer(&first[0]), len(first))

	// 7 bytes: 7+10 = 17 > 16, but 7 <= 16 -> new buffer path (not a clone).
	second := []byte("abcdefg")
	s2 := sBuf.AllocString(unsafe.Pointer(&second[0]), len(second))

	if s1 != "0123456789" {
		t.Fatalf("s1 = %q, want %q", s1, "0123456789")
	}
	if s2 != "abcdefg" {
		t.Fatalf("s2 = %q, want %q", s2, "abcdefg")
	}

	// s2 lives in the freshly allocated current buffer; s1 still points at the
	// previous one, which is no longer the active backing array.
	if !ptrInBuf(sBuf, unsafe.Pointer(unsafe.StringData(s2))) {
		t.Fatalf("s2 must point into the current safe buffer")
	}
	if ptrInBuf(sBuf, unsafe.Pointer(unsafe.StringData(s1))) {
		t.Fatalf("s1 must point into the previous (exhausted) buffer, not the current one")
	}
	if len(sBuf.buf) != len(second) {
		t.Fatalf("buf len = %d, want %d (cursor must reset to 0)", len(sBuf.buf), len(second))
	}
}

// Multiple sequential in-buffer allocations must pack back to back without
// overlapping, each returning a view over the shared backing array.
func TestSafeBuffer_SequentialInBufferAllocs(t *testing.T) {
	sBuf := NewSafeBuffer(64)

	a := []byte("AAAA")
	b := []byte("BBBBBB")
	c := []byte("CC")

	sa := sBuf.AllocString(unsafe.Pointer(&a[0]), len(a))
	sb := sBuf.AllocString(unsafe.Pointer(&b[0]), len(b))
	sc := sBuf.AllocString(unsafe.Pointer(&c[0]), len(c))

	if sa != "AAAA" || sb != "BBBBBB" || sc != "CC" {
		t.Fatalf("contents mismatch: %q %q %q", sa, sb, sc)
	}

	pa := uintptr(unsafe.Pointer(unsafe.StringData(sa)))
	pb := uintptr(unsafe.Pointer(unsafe.StringData(sb)))
	pc := uintptr(unsafe.Pointer(unsafe.StringData(sc)))

	if pb-pa != uintptr(len(a)) {
		t.Fatalf("gap between sa and sb = %d, want %d", pb-pa, len(a))
	}
	if pc-pb != uintptr(len(b)) {
		t.Fatalf("gap between sb and sc = %d, want %d", pc-pb, len(b))
	}
	if len(sBuf.buf) != len(a)+len(b)+len(c) {
		t.Fatalf("buf len = %d, want %d", len(sBuf.buf), len(a)+len(b)+len(c))
	}
}

// Strings returned by AllocString must survive garbage collection: the returned
// string keeps the (possibly already-replaced) backing array alive, so its bytes
// remain valid even after the SafeBuffer moves on to a fresh buffer and the GC
// reclaims everything else.
func TestSafeBuffer_GCRetention_String(t *testing.T) {
	sBuf := NewSafeBuffer(32)

	// Fill the buffer so the next allocation is forced onto a new backing array.
	first := bytes.Repeat([]byte("X"), 30) // 30 <= 32, fits
	s1 := sBuf.AllocString(unsafe.Pointer(&first[0]), len(first))
	want1 := string(first) // independent reference copy

	second := bytes.Repeat([]byte("Y"), 30) // 30+30 > 32, but 30 <= 32 -> new buffer
	s2 := sBuf.AllocString(unsafe.Pointer(&second[0]), len(second))
	want2 := string(second)

	// Drop the source slices and hammer the GC with allocation pressure. If the
	// backing arrays were not retained, the memory would be reused and the
	// strings' bytes corrupted.
	first = nil
	second = nil
	for range 10 {
		runtime.GC()
		_ = make([]byte, 1<<20) // 1 MiB of garbage per iteration
	}

	if s1 != want1 {
		t.Fatalf("s1 corrupted after GC: got %q, want %q", s1, want1)
	}
	if s2 != want2 {
		t.Fatalf("s2 corrupted after GC: got %q, want %q", s2, want2)
	}
}

// Bytes returned by AllocBytes must survive garbage collection the same way
// strings do: the slice keeps its (possibly already-replaced) backing array alive.
func TestSafeBuffer_GCRetention_GoroutineCache(t *testing.T) {
	const workers = 8
	const iterations = 1000
	var wg sync.WaitGroup

	for w := 0; w < workers; w++ {
		wg.Go(func() {
			sBuf := NewSafeBuffer(32)
			runtime.Gosched()

			for i := 0; i < iterations; i++ {
				// 1. We'll copy this.
				temp := []byte("test_data_for_msgpack_retention_check")

				// 2. Make a copy.
				bs := sBuf.AllocBytes(unsafe.Pointer(&temp[0]), len(temp))

				// Make independent copy.
				want := append([]byte(nil), bs...)

				// 3. Spam small allocations in the cache of the current goroutine to provoke a GC.
				go func() {
					for k := 0; k < 500; k++ {
						// Create small junk with random size.
						_ = make([]byte, 16+k%64)
					}
				}()

				runtime.GC()

				// 4. Check if the data in the bs was not overridden.
				if !bytes.Equal(bs, want) {
					t.Errorf("a copy  is corrupted: got %s, want %s", bs, want)
					return
				}
			}
		})
	}
	wg.Wait()
}

// A mix of types and new buffers.
func TestSafeBuffer_GCRetention_MixedPaths(t *testing.T) {
	sBuf := NewSafeBuffer(16)

	// In-buffer allocation.
	inBuf := []byte("0123456789") // 10 bytes, fits
	sIn := sBuf.AllocString(unsafe.Pointer(&inBuf[0]), len(inBuf))
	wantIn := string(inBuf)

	// New-buffer allocation (exhausts current buffer, size still <= cap).
	newBuf := []byte("abcdefgh") // 8 bytes; 8+10 = 18 > 16 -> new buffer
	sNew := string(sBuf.AllocBytes(unsafe.Pointer(&newBuf[0]), len(newBuf)))
	wantNew := string(newBuf)

	// Clone allocation (size > cap).
	big := bytes.Repeat([]byte("Z"), 64) // 64 > 16 -> immediate clone
	sClone := sBuf.AllocBytes(unsafe.Pointer(&big[0]), len(big))
	wantClone := append([]byte(nil), big...)

	if sIn != wantIn {
		t.Fatalf("in-buffer value corrupted after GC: got %q, want %q", sIn, wantIn)
	}
	if sNew != wantNew {
		t.Fatalf("new-buffer value corrupted after GC: got %q, want %q", sNew, wantNew)
	}
	if !bytes.Equal(sClone, wantClone) {
		t.Fatalf("cloned value corrupted after GC: got %v, want %v", sClone, wantClone)
	}
}
