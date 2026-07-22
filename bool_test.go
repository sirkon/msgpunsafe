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
