package msgpunsafe

import (
	"testing"
	"unsafe"
)

// b builds a byte slice from its arguments — a compact way to write msgpack
// payloads in tests.
func b(bts ...byte) []byte { return bts }

// ptrs returns pointers to the start (src) and end (lim) of the input buffer.
// For an empty buffer it returns two equal, valid pointers, which corresponds
// to the "exhausted" state (src >= lim) before reading the first byte.
func ptrs(input []byte) (unsafe.Pointer, unsafe.Pointer) {
	if len(input) == 0 {
		var z [1]byte
		p := unsafe.Pointer(&z[0])
		return p, p
	}

	p := unsafe.Pointer(&input[0])

	return p, unsafe.Add(p, len(input))
}

// runTake invokes fn, recovering from a deserializer panic.
//
// If wantErrCode != 0, a panic with exactly that ErrorCode is expected.
// If wantErrCode == 0, no panic must occur and the returned next pointer
// must point to the end of the buffer (the whole input consumed).
//
// It returns the read value and an ok flag (true means the value is valid
// and can be compared against the expected one).
func runTake[T any](
	t *testing.T,
	name string,
	wantErrCode ErrorCode,
	input []byte,
	src unsafe.Pointer,
	fn func() (T, unsafe.Pointer),
) (got T, ok bool) {
	t.Helper()

	var next unsafe.Pointer
	panicked := false

	func() {
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			panicked = true

			code, isCode := r.(ErrorCode)
			if !isCode {
				t.Fatalf("%s: panic with non-ErrorCode: %v", name, r)
			}

			if code != wantErrCode {
				t.Fatalf("%s: expected error code %d, got %d", name, wantErrCode, code)
			}
		}()

		got, next = fn()
	}()

	if wantErrCode != 0 {
		if !panicked {
			t.Fatalf("%s: expected panic with code %d, but no panic occurred", name, wantErrCode)
		}

		var zero T
		return zero, false
	}

	if panicked {
		// A panic occurred but was not expected — the message was already printed in recover.
		return got, false
	}

	// Happy path: the whole input buffer must be consumed.
	if wantNext := unsafe.Add(src, len(input)); next != wantNext {
		t.Fatalf("%s: incorrect cursor advancement: got %x want %x", name, next, wantNext)
	}

	return got, true
}
