package msgpunsafe

import (
	"testing"
	"unsafe"
)

func TestTakeBool(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        bool
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path ---
		{name: "false", input: b(0xc2), want: false},
		{name: "true", input: b(0xc3), want: true},

		// --- ErrorBoolExhausted ---
		{name: "exhausted", input: b(), wantErrCode: ErrorBoolExhausted},

		// --- ErrorBoolCorrupted ---
		{name: "corrupted nil marker", input: b(0xc0), wantErrCode: ErrorBoolCorrupted},
		{name: "corrupted fixint", input: b(0x05), wantErrCode: ErrorBoolCorrupted},
		{name: "corrupted true-ish", input: b(0xc4), wantErrCode: ErrorBoolCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (bool, unsafe.Pointer) { return TakeBool(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTakeBoolPtr(t *testing.T) {
	src, lim := ptrs(b(0xc3))
	sBuf := NewSafeBuffer(512)
	var got *bool
	runTake(t, "boolptr", 0, b(0xc3), src,
		func() (bool, unsafe.Pointer) {
			var next unsafe.Pointer
			got, next = TakeBoolPtr(src, lim, sBuf)
			return false, next
		})
	if got == nil || !*got {
		t.Fatalf("got %v, want true", got)
	}
	if uintptr(unsafe.Pointer(got))%8 != 0 {
		t.Fatalf("pointer %x is not 8-byte aligned", uintptr(unsafe.Pointer(got)))
	}
	if !ptrInBuf(sBuf, unsafe.Pointer(got)) {
		t.Fatalf("pointer must point into the safe buffer")
	}
}

func TestTakeBoolPtr_ErrorNoAlloc(t *testing.T) {
	src, lim := ptrs(b(0x05)) // corrupted marker
	sBuf := NewSafeBuffer(512)
	var got *bool
	runTake(t, "boolptr corrupted", ErrorBoolCorrupted, b(0x05), src,
		func() (bool, unsafe.Pointer) {
			var next unsafe.Pointer
			got, next = TakeBoolPtr(src, lim, sBuf)
			return false, next
		})
	_ = got
	if len(sBuf.buf) != 0 {
		t.Fatalf("buffer must stay untouched on error, len = %d", len(sBuf.buf))
	}
}
