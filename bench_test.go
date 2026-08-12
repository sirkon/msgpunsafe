package msgpunsafe

import (
	"fmt"
	"testing"
	"unsafe"
)

// msgpackString encodes s as a msgpack string: fixstr for <32 bytes, str8 for
// 32-255 bytes.
func msgpackString(s string) []byte {
	out := make([]byte, 0, len(s)+2)
	if len(s) < 32 {
		out = append(out, 0xa0|byte(len(s)))
	} else {
		out = append(out, 0xd9, byte(len(s)))
	}
	return append(out, s...)
}

// buildStringArrayPayload builds a msgpack fixarray/array16 of n strings, each
// size bytes long.
func buildStringArrayPayload(n, size int) []byte {
	str := make([]byte, size)
	for i := range str {
		str[i] = byte('a' + i%26)
	}
	enc := msgpackString(string(str))

	out := make([]byte, 0, n*len(enc)+1)
	if n < 16 {
		out = append(out, 0x90|byte(n))
	} else {
		out = append(out, 0xdc, byte(n>>8), byte(n))
	}
	for range n {
		out = append(out, enc...)
	}
	return out
}

var benchSizes = []int{8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096}

// BenchmarkNewSafeBuffer measures the raw cost of allocating (and zeroing) a
// SafeBuffer backing array of a given capacity.
func BenchmarkNewSafeBuffer(b *testing.B) {
	for _, cap := range benchSizes {
		b.Run(fmt.Sprintf("cap=%d", cap), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = NewSafeBuffer(cap)
			}
		})
	}
}

// BenchmarkTakeString_FreshBuffer decodes a fixed msgpack string payload with a
// freshly allocated SafeBuffer per iteration, matching the per-message decode
// pattern where the arena is created, used, and discarded. This accounts for
// the upfront allocation cost, unlike a long-lived reused buffer.
func BenchmarkTakeString_FreshBuffer(b *testing.B) {
	// 8 strings of 48 bytes = 384 bytes of string data per message.
	payload := buildStringArrayPayload(8, 48)

	for _, cap := range benchSizes {
		b.Run(fmt.Sprintf("cap=%d", cap), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				sBuf := NewSafeBuffer(cap)
				src := unsafe.Pointer(&payload[0])
				lim := unsafe.Add(src, len(payload))

				n, src := TakeSliceHeader(src, lim)
				for range n {
					_, src = TakeString(src, lim, sBuf)
				}
			}
		})
	}
}

// BenchmarkTakeString_ReusedBuffer decodes the same payload with one
// long-lived SafeBuffer reused across iterations, matching a hot decoder that
// recycles a single arena. In steady state the arena stays full, so capacity
// determines how often a fresh backing array must be allocated.
func BenchmarkTakeString_ReusedBuffer(b *testing.B) {
	payload := buildStringArrayPayload(8, 48)

	for _, cap := range benchSizes {
		b.Run(fmt.Sprintf("cap=%d", cap), func(b *testing.B) {
			sBuf := NewSafeBuffer(cap)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				src := unsafe.Pointer(&payload[0])
				lim := unsafe.Add(src, len(payload))

				n, src := TakeSliceHeader(src, lim)
				for range n {
					_, src = TakeString(src, lim, sBuf)
				}
			}
		})
	}
}
