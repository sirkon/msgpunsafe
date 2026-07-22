package msgpunsafe

import (
	"bytes"
	"testing"
	"unsafe"
)

func TestTakeString(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        string
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path ---
		{name: "fixstr empty", input: b(0xa0), want: ""},
		{name: "fixstr hi", input: b(0xa2, 'h', 'i'), want: "hi"},
		{name: "fixstr max", input: append(b(0xbf), []byte("1234567890123456789012345678901")...), want: "1234567890123456789012345678901"},
		{name: "str8 abc", input: b(0xd9, 0x03, 'a', 'b', 'c'), want: "abc"},
		{name: "str16 x", input: b(0xda, 0x00, 0x01, 'x'), want: "x"},
		{name: "str32 y", input: b(0xdb, 0x00, 0x00, 0x00, 0x01, 'y'), want: "y"},

		// --- Panics from TakeStrHeader ---
		{name: "exhausted", input: b(), wantErrCode: ErrorStrExhausted},
		{name: "str8 truncated", input: b(0xd9), wantErrCode: ErrorStrHeaderTruncated},
		{name: "str header corrupted", input: b(0xc0), wantErrCode: ErrorStrHeaderCorrupted},

		// --- ErrorStrLenConflict: declared length exceeds the remaining buffer ---
		{name: "len conflict fixstr", input: b(0xa5), wantErrCode: ErrorStrLenConflict},
		{name: "len conflict str8", input: b(0xd9, 0x05, 'a', 'b'), wantErrCode: ErrorStrLenConflict},
		{name: "len conflict str16", input: b(0xda, 0x00, 0x05, 'a', 'b'), wantErrCode: ErrorStrLenConflict},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			sBuf := NewSafeBuffer(512)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (string, unsafe.Pointer) { return TakeString(src, lim, sBuf) })
			if ok && got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTakeStringZC(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        string
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path ---
		{name: "fixstr empty", input: b(0xa0), want: ""},
		{name: "fixstr hi", input: b(0xa2, 'h', 'i'), want: "hi"},
		{name: "str8 abc", input: b(0xd9, 0x03, 'a', 'b', 'c'), want: "abc"},
		{name: "str16 x", input: b(0xda, 0x00, 0x01, 'x'), want: "x"},
		{name: "str32 y", input: b(0xdb, 0x00, 0x00, 0x00, 0x01, 'y'), want: "y"},

		// --- Panics from TakeStrHeader ---
		{name: "exhausted", input: b(), wantErrCode: ErrorStrExhausted},
		{name: "str8 truncated", input: b(0xd9), wantErrCode: ErrorStrHeaderTruncated},
		{name: "str header corrupted", input: b(0xc0), wantErrCode: ErrorStrHeaderCorrupted},

		// --- ErrorStrLenConflict ---
		{name: "len conflict fixstr", input: b(0xa5), wantErrCode: ErrorStrLenConflict},
		{name: "len conflict str8", input: b(0xd9, 0x05, 'a', 'b'), wantErrCode: ErrorStrLenConflict},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (string, unsafe.Pointer) { return TakeStringZC(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestTakeBytes(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        []byte
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path ---
		{name: "bin8 empty", input: b(0xc4, 0x00), want: []byte{}},
		{name: "bin8 hi", input: b(0xc4, 0x02, 'h', 'i'), want: []byte("hi")},
		{name: "bin16 abc", input: b(0xc5, 0x00, 0x03, 'a', 'b', 'c'), want: []byte("abc")},
		{name: "bin32 x", input: b(0xc6, 0x00, 0x00, 0x00, 0x01, 'x'), want: []byte("x")},

		// --- Panics from TakeBinHeader ---
		{name: "exhausted", input: b(), wantErrCode: ErrorBinExhausted},
		{name: "bin8 truncated", input: b(0xc4), wantErrCode: ErrorBinHeaderTruncated},
		{name: "bin header corrupted", input: b(0xa0), wantErrCode: ErrorBinHeaderCorrupted},

		// --- ErrorBinLenConflict: declared length exceeds the remaining buffer ---
		{name: "len conflict bin8", input: b(0xc4, 0x05), wantErrCode: ErrorBinLenConflict},
		{name: "len conflict bin16", input: b(0xc5, 0x00, 0x05, 'a', 'b'), wantErrCode: ErrorBinLenConflict},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			sBuf := NewSafeBuffer(512)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() ([]byte, unsafe.Pointer) { return TakeBytes(src, lim, sBuf) })
			if ok && !bytes.Equal(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
