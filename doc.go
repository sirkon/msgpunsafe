// Package msgpunsafe provides zero-dependency unsafe, panic-based
// deserialization primitives for msgpack (MessagePack).
//
// Every consumer-facing function is named Take* and reads exactly one value
// from a raw byte buffer using a src/lim pointer-pair cursor model: src is the
// current read position and lim is one past the end of the buffer. Every Take*
// returns (value, nextPtr) where nextPtr is the advanced cursor; the caller is
// responsible for threading nextPtr into the next call. Failures are reported
// by panicking with an ErrorCode, never with a returned error.
//
// The expected usage pattern is a decoder protected by a deferred
// HandleError call that converts a valid ErrorCode panic into an error and
// re-panics anything else:
//
//	func decode(input []byte) (err error) {
//		defer msgpunsafe.HandleError(recover(), &err)
//
//		src := unsafe.Pointer(&input[0])
//		lim := unsafe.Add(src, len(input))
//
//		buf := msgpunsafe.NewSafeBuffer(512)
//
//		n, src := msgpunsafe.TakeInt64(src, lim)
//		s, src := msgpunsafe.TakeString(src, lim, buf)
//		_ = n
//		_ = s
//		return nil
//	}
//
// Strings and bytes decoded into temporary buffers must be copied to survive
// buffer reuse; SafeBuffer (NewSafeBuffer) is a reusable arena that stores
// values copied from the source. TakeStringZC is the zero-copy variant that
// points into the caller's buffer instead, leaving lifetime management to the
// caller. The Take*Ptr variants parse a scalar value and store it in an
// 8-byte-aligned slot of the arena, returning a stable pointer to it.
//
// Numeric decoding has two backends selected at build time: the default build
// reads numbers with native-endian casts (fast, unsafe on unaligned buffers),
// while the aware_of_alignment build tag reads them byte by byte through
// encoding/binary (alignment-safe, slower).
package msgpunsafe
