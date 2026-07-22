package msgpunsafe

import (
	"testing"
	"unsafe"
)

func TestTakeSliceHeader(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path ---
		{name: "fixarray 0", input: b(0x90), want: 0},
		{name: "fixarray 15", input: b(0x9f), want: 15},
		{name: "array16 16", input: b(0xdc, 0x00, 0x10), want: 16},
		{name: "array16 max", input: b(0xdc, 0xff, 0xff), want: 0xffff},
		{name: "array32 65536", input: b(0xdd, 0x00, 0x01, 0x00, 0x00), want: 65536},

		// --- ErrorSliceExhausted ---
		{name: "exhausted", input: b(), wantErrCode: ErrorSliceExhausted},

		// --- ErrorSliceHeaderTruncated ---
		{name: "array16 truncated", input: b(0xdc, 0x00), wantErrCode: ErrorSliceHeaderTruncated},
		{name: "array32 truncated", input: b(0xdd, 0x00, 0x01, 0x00), wantErrCode: ErrorSliceHeaderTruncated},

		// --- ErrorSliceHeaderCorrupted ---
		{name: "corrupted nil marker", input: b(0xc0), wantErrCode: ErrorSliceHeaderCorrupted},
		{name: "corrupted map marker", input: b(0x80), wantErrCode: ErrorSliceHeaderCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int, unsafe.Pointer) { return TakeSliceHeader(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeMapHeader(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path ---
		{name: "fixmap 0", input: b(0x80), want: 0},
		{name: "fixmap 15", input: b(0x8f), want: 15},
		{name: "map16 16", input: b(0xde, 0x00, 0x10), want: 16},
		{name: "map16 max", input: b(0xde, 0xff, 0xff), want: 0xffff},
		{name: "map32 65536", input: b(0xdf, 0x00, 0x01, 0x00, 0x00), want: 65536},

		// --- ErrorMapExhausted ---
		{name: "exhausted", input: b(), wantErrCode: ErrorMapExhausted},

		// --- ErrorMapHeaderTruncated ---
		{name: "map16 truncated", input: b(0xde, 0x00), wantErrCode: ErrorMapHeaderTruncated},
		{name: "map32 truncated", input: b(0xdf, 0x00, 0x01, 0x00), wantErrCode: ErrorMapHeaderTruncated},

		// --- ErrorMapHeaderCorrupted ---
		{name: "corrupted nil marker", input: b(0xc0), wantErrCode: ErrorMapHeaderCorrupted},
		{name: "corrupted array marker", input: b(0x90), wantErrCode: ErrorMapHeaderCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int, unsafe.Pointer) { return TakeMapHeader(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeStrHeader(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path ---
		{name: "fixstr 0", input: b(0xa0), want: 0},
		{name: "fixstr 31", input: b(0xbf), want: 31},
		{name: "str8 32", input: b(0xd9, 0x20), want: 32},
		{name: "str8 255", input: b(0xd9, 0xff), want: 255},
		{name: "str16 256", input: b(0xda, 0x01, 0x00), want: 256},
		{name: "str16 max", input: b(0xda, 0xff, 0xff), want: 0xffff},
		{name: "str32 65536", input: b(0xdb, 0x00, 0x01, 0x00, 0x00), want: 65536},

		// --- ErrorStrExhausted ---
		{name: "exhausted", input: b(), wantErrCode: ErrorStrExhausted},

		// --- ErrorStrHeaderTruncated ---
		{name: "str8 truncated", input: b(0xd9), wantErrCode: ErrorStrHeaderTruncated},
		{name: "str16 truncated", input: b(0xda, 0x01), wantErrCode: ErrorStrHeaderTruncated},
		{name: "str32 truncated", input: b(0xdb, 0x00, 0x01, 0x00), wantErrCode: ErrorStrHeaderTruncated},

		// --- ErrorStrHeaderCorrupted ---
		{name: "corrupted nil marker", input: b(0xc0), wantErrCode: ErrorStrHeaderCorrupted},
		{name: "corrupted bin marker", input: b(0xc4), wantErrCode: ErrorStrHeaderCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int, unsafe.Pointer) { return TakeStrHeader(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeBinHeader(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path ---
		{name: "bin8 0", input: b(0xc4, 0x00), want: 0},
		{name: "bin8 255", input: b(0xc4, 0xff), want: 255},
		{name: "bin16 256", input: b(0xc5, 0x01, 0x00), want: 256},
		{name: "bin16 max", input: b(0xc5, 0xff, 0xff), want: 0xffff},
		{name: "bin32 65536", input: b(0xc6, 0x00, 0x01, 0x00, 0x00), want: 65536},

		// --- ErrorBinExhausted ---
		{name: "exhausted", input: b(), wantErrCode: ErrorBinExhausted},

		// --- ErrorBinHeaderTruncated ---
		{name: "bin8 truncated", input: b(0xc4), wantErrCode: ErrorBinHeaderTruncated},
		{name: "bin16 truncated", input: b(0xc5, 0x01), wantErrCode: ErrorBinHeaderTruncated},
		{name: "bin32 truncated", input: b(0xc6, 0x00, 0x01, 0x00), wantErrCode: ErrorBinHeaderTruncated},

		// --- ErrorBinHeaderCorrupted ---
		{name: "corrupted nil marker", input: b(0xc0), wantErrCode: ErrorBinHeaderCorrupted},
		{name: "corrupted str marker", input: b(0xa0), wantErrCode: ErrorBinHeaderCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int, unsafe.Pointer) { return TakeBinHeader(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}
