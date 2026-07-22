package msgpunsafe

import (
	"math"
	"testing"
	"unsafe"
)

func TestTakeUint64(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        uint64
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path: all supported encodings ---
		{name: "fixint 0", input: b(0x00), want: 0},
		{name: "fixint 127", input: b(0x7f), want: 127},
		{name: "uint8 0", input: b(0xcc, 0x00), want: 0},
		{name: "uint8 255", input: b(0xcc, 0xff), want: 255},
		{name: "uint16 256", input: b(0xcd, 0x01, 0x00), want: 256},
		{name: "uint16 max", input: b(0xcd, 0xff, 0xff), want: 0xffff},
		{name: "uint32 65536", input: b(0xce, 0x00, 0x01, 0x00, 0x00), want: 0x10000},
		{name: "uint32 max", input: b(0xce, 0xff, 0xff, 0xff, 0xff), want: 0xffffffff},
		{name: "uint64 max32+1", input: b(0xcf, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00), want: 0x100000000},
		{name: "uint64 max", input: b(0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff), want: math.MaxUint64},
		{name: "int8 positive", input: b(0xd0, 0x64), want: 100},
		{name: "int16 positive", input: b(0xd1, 0x03, 0xe8), want: 1000},
		{name: "int32 positive", input: b(0xd2, 0x00, 0x01, 0x86, 0xa0), want: 100000},
		{name: "int64 positive", input: b(0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86, 0xa0), want: 100000},

		// --- ErrorNumberExhausted ---
		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},

		// --- ErrorUintOverflow (negative values) ---
		{name: "negative fixint", input: b(0xff), wantErrCode: ErrorUintOverflow},
		{name: "int8 negative", input: b(0xd0, 0xff), wantErrCode: ErrorUintOverflow},
		{name: "int16 negative", input: b(0xd1, 0xff, 0xff), wantErrCode: ErrorUintOverflow},
		{name: "int32 negative", input: b(0xd2, 0xff, 0xff, 0xff, 0xff), wantErrCode: ErrorUintOverflow},
		{name: "int64 negative", input: b(0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff), wantErrCode: ErrorUintOverflow},

		// --- ErrorNumberTruncated (each multi-byte marker) ---
		{name: "uint8 truncated", input: b(0xcc), wantErrCode: ErrorNumberTruncated},
		{name: "uint16 truncated", input: b(0xcd, 0x01), wantErrCode: ErrorNumberTruncated},
		{name: "uint32 truncated", input: b(0xce, 0x00, 0x01, 0x00), wantErrCode: ErrorNumberTruncated},
		{name: "uint64 truncated", input: b(0xcf, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00), wantErrCode: ErrorNumberTruncated},
		{name: "int8 truncated", input: b(0xd0), wantErrCode: ErrorNumberTruncated},
		{name: "int16 truncated", input: b(0xd1, 0x03), wantErrCode: ErrorNumberTruncated},
		{name: "int32 truncated", input: b(0xd2, 0x00, 0x01, 0x86), wantErrCode: ErrorNumberTruncated},
		{name: "int64 truncated", input: b(0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86), wantErrCode: ErrorNumberTruncated},

		// --- ErrorUintCorrupted ---
		{name: "corrupted nil marker", input: b(0xc0), wantErrCode: ErrorUintCorrupted},
		{name: "corrupted str marker", input: b(0xa0), wantErrCode: ErrorUintCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (uint64, unsafe.Pointer) { return TakeUint64(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeUint32(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        uint32
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "fixint 5", input: b(0x05), want: 5},
		{name: "uint8 200", input: b(0xcc, 0xc8), want: 200},
		{name: "uint32 max", input: b(0xce, 0xff, 0xff, 0xff, 0xff), want: 0xffffffff},
		{name: "uint64 small", input: b(0xcf, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01), want: 1},

		// Narrowing overflow of uint32.
		{name: "overflow uint64", input: b(0xcf, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00), wantErrCode: ErrorUintOverflow},

		// Inherited panics from takeUint64 (one per code).
		{name: "negative fixint", input: b(0xff), wantErrCode: ErrorUintOverflow},
		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "uint8 truncated", input: b(0xcc), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted", input: b(0xc0), wantErrCode: ErrorUintCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (uint32, unsafe.Pointer) { return TakeUint32(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeUint16(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        uint16
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "fixint 5", input: b(0x05), want: 5},
		{name: "uint16 max", input: b(0xcd, 0xff, 0xff), want: 0xffff},
		{name: "uint32 small", input: b(0xce, 0x00, 0x00, 0x03, 0xe8), want: 1000},
		{name: "int16 positive", input: b(0xd1, 0x03, 0xe8), want: 1000},

		// Narrowing overflow of uint16.
		{name: "overflow uint32", input: b(0xce, 0x00, 0x01, 0x00, 0x00), wantErrCode: ErrorUintOverflow},
		{name: "overflow uint64", input: b(0xcf, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00), wantErrCode: ErrorUintOverflow},

		{name: "negative fixint", input: b(0xff), wantErrCode: ErrorUintOverflow},
		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "uint16 truncated", input: b(0xcd, 0x01), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted", input: b(0xc0), wantErrCode: ErrorUintCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (uint16, unsafe.Pointer) { return TakeUint16(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeUint8(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        uint8
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "fixint 0", input: b(0x00), want: 0},
		{name: "fixint 127", input: b(0x7f), want: 127},
		{name: "uint8 255", input: b(0xcc, 0xff), want: 255},
		{name: "int8 positive", input: b(0xd0, 0x64), want: 100},
		{name: "uint16 small", input: b(0xcd, 0x00, 0x05), want: 5},

		// Narrowing overflow of uint8.
		{name: "overflow uint16", input: b(0xcd, 0x01, 0x00), wantErrCode: ErrorUintOverflow},
		{name: "overflow uint32", input: b(0xce, 0x00, 0x00, 0x01, 0x00), wantErrCode: ErrorUintOverflow},

		{name: "negative fixint", input: b(0xff), wantErrCode: ErrorUintOverflow},
		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "uint8 truncated", input: b(0xcc), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted", input: b(0xc0), wantErrCode: ErrorUintCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (uint8, unsafe.Pointer) { return TakeUint8(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeUint(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        uint
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "fixint 127", input: b(0x7f), want: 127},
		{name: "uint64 max", input: b(0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff), want: math.MaxUint64},
		{name: "int64 positive", input: b(0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86, 0xa0), want: 100000},

		// TakeUint does not narrow on a 64-bit platform, so its own overflow
		// is unreachable here — we test overflow via negative values instead.
		{name: "negative fixint", input: b(0xff), wantErrCode: ErrorUintOverflow},
		{name: "int64 negative", input: b(0xd3, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff), wantErrCode: ErrorUintOverflow},
		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "uint64 truncated", input: b(0xcf, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted", input: b(0xc0), wantErrCode: ErrorUintCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (uint, unsafe.Pointer) { return TakeUint(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeInt64(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int64
		wantErrCode ErrorCode
	}

	tests := []testCase{
		// --- Happy path: all supported encodings ---
		{name: "fixint 0", input: b(0x00), want: 0},
		{name: "fixint 127", input: b(0x7f), want: 127},
		{name: "negative fixint -1", input: b(0xff), want: -1},
		{name: "negative fixint -32", input: b(0xe0), want: -32},
		{name: "int8 100", input: b(0xd0, 0x64), want: 100},
		{name: "int8 -100", input: b(0xd0, 0x9c), want: -100},
		{name: "int16 1000", input: b(0xd1, 0x03, 0xe8), want: 1000},
		{name: "int16 -1000", input: b(0xd1, 0xfc, 0x18), want: -1000},
		{name: "int32 100000", input: b(0xd2, 0x00, 0x01, 0x86, 0xa0), want: 100000},
		{name: "int32 -1", input: b(0xd2, 0xff, 0xff, 0xff, 0xff), want: -1},
		{name: "int64 max", input: b(0xd3, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff), want: math.MaxInt64},
		{name: "int64 min", input: b(0xd3, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00), want: math.MinInt64},
		{name: "uint8 200", input: b(0xcc, 0xc8), want: 200},
		{name: "uint16 300", input: b(0xcd, 0x01, 0x2c), want: 300},
		{name: "uint32 70000", input: b(0xce, 0x00, 0x01, 0x11, 0x70), want: 70000},
		{name: "uint64 max int", input: b(0xcf, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff), want: math.MaxInt64},

		// --- ErrorNumberExhausted ---
		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},

		// --- ErrorNumberTruncated (each multi-byte marker) ---
		{name: "int8 truncated", input: b(0xd0), wantErrCode: ErrorNumberTruncated},
		{name: "int16 truncated", input: b(0xd1, 0x03), wantErrCode: ErrorNumberTruncated},
		{name: "int32 truncated", input: b(0xd2, 0x00, 0x01, 0x86), wantErrCode: ErrorNumberTruncated},
		{name: "int64 truncated", input: b(0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86), wantErrCode: ErrorNumberTruncated},
		{name: "uint8 truncated", input: b(0xcc), wantErrCode: ErrorNumberTruncated},
		{name: "uint16 truncated", input: b(0xcd, 0x01), wantErrCode: ErrorNumberTruncated},
		{name: "uint32 truncated", input: b(0xce, 0x00, 0x01, 0x00), wantErrCode: ErrorNumberTruncated},
		{name: "uint64 truncated", input: b(0xcf, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00), wantErrCode: ErrorNumberTruncated},

		// --- ErrorIntOverflow (uint64 > MaxInt64) ---
		{name: "uint64 overflow", input: b(0xcf, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00), wantErrCode: ErrorIntOverflow},

		// --- ErrorIntCorrupted ---
		{name: "corrupted nil marker", input: b(0xc0), wantErrCode: ErrorIntCorrupted},
		{name: "corrupted str marker", input: b(0xa0), wantErrCode: ErrorIntCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int64, unsafe.Pointer) { return TakeInt64(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeInt32(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int32
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "fixint 5", input: b(0x05), want: 5},
		{name: "int8 -100", input: b(0xd0, 0x9c), want: -100},
		{name: "int32 100000", input: b(0xd2, 0x00, 0x01, 0x86, 0xa0), want: 100000},
		{name: "int32 -1", input: b(0xd2, 0xff, 0xff, 0xff, 0xff), want: -1},
		{name: "uint32 70000", input: b(0xce, 0x00, 0x01, 0x11, 0x70), want: 70000},

		// Narrowing overflow of int32.
		{name: "overflow uint32", input: b(0xce, 0x80, 0x00, 0x00, 0x00), wantErrCode: ErrorIntOverflow},
		{name: "overflow int64", input: b(0xd3, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00), wantErrCode: ErrorIntOverflow},

		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "int32 truncated", input: b(0xd2, 0x00, 0x01, 0x86), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted", input: b(0xc0), wantErrCode: ErrorIntCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int32, unsafe.Pointer) { return TakeInt32(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeInt16(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int16
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "fixint 5", input: b(0x05), want: 5},
		{name: "negative fixint -1", input: b(0xff), want: -1},
		{name: "int16 1000", input: b(0xd1, 0x03, 0xe8), want: 1000},
		{name: "int16 -1000", input: b(0xd1, 0xfc, 0x18), want: -1000},
		{name: "uint16 300", input: b(0xcd, 0x01, 0x2c), want: 300},

		// Narrowing overflow of int16.
		{name: "overflow uint16", input: b(0xcd, 0x80, 0x00), wantErrCode: ErrorIntOverflow},
		{name: "overflow int32", input: b(0xd2, 0x00, 0x00, 0x9c, 0x40), wantErrCode: ErrorIntOverflow},

		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "int16 truncated", input: b(0xd1, 0x03), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted", input: b(0xc0), wantErrCode: ErrorIntCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int16, unsafe.Pointer) { return TakeInt16(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeInt8(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int8
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "fixint 5", input: b(0x05), want: 5},
		{name: "negative fixint -1", input: b(0xff), want: -1},
		{name: "int8 100", input: b(0xd0, 0x64), want: 100},
		{name: "int8 -100", input: b(0xd0, 0x9c), want: -100},
		{name: "uint8 127", input: b(0xcc, 0x7f), want: 127},

		// Narrowing overflow of int8.
		{name: "overflow uint8", input: b(0xcc, 0x80), wantErrCode: ErrorIntOverflow},
		{name: "overflow int16", input: b(0xd1, 0x00, 0xc8), wantErrCode: ErrorIntOverflow},

		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "int8 truncated", input: b(0xd0), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted", input: b(0xc0), wantErrCode: ErrorIntCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int8, unsafe.Pointer) { return TakeInt8(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeInt(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "fixint 127", input: b(0x7f), want: 127},
		{name: "negative fixint -1", input: b(0xff), want: -1},
		{name: "int64 max", input: b(0xd3, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff), want: math.MaxInt64},
		{name: "int64 min", input: b(0xd3, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00), want: math.MinInt64},

		// TakeInt does not narrow on a 64-bit platform; overflow is tested via uint64 > MaxInt64.
		{name: "uint64 overflow", input: b(0xcf, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00), wantErrCode: ErrorIntOverflow},
		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "int64 truncated", input: b(0xd3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x86), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted", input: b(0xc0), wantErrCode: ErrorIntCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int, unsafe.Pointer) { return TakeInt(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %d, want %d", got, tc.want)
			}
		})
	}
}

func TestTakeFloat32(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        float32
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "float32 1.0", input: b(0xca, 0x3f, 0x80, 0x00, 0x00), want: 1.0},
		{name: "float32 -1.5", input: b(0xca, 0xbf, 0xc0, 0x00, 0x00), want: -1.5},
		{name: "float64 promoted", input: b(0xcb, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00), want: 2.0},

		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "float32 truncated", input: b(0xca, 0x3f, 0x80, 0x00), wantErrCode: ErrorNumberTruncated},
		{name: "float64 truncated", input: b(0xcb, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted nil marker", input: b(0xc0), wantErrCode: ErrorFloatCorrupted},
		{name: "corrupted int marker", input: b(0x05), wantErrCode: ErrorFloatCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (float32, unsafe.Pointer) { return TakeFloat32(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTakeFloat64(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        float64
		wantErrCode ErrorCode
	}

	tests := []testCase{
		{name: "float64 1.0", input: b(0xcb, 0x3f, 0xf0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00), want: 1.0},
		{name: "float64 2.0", input: b(0xcb, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00), want: 2.0},
		{name: "float32 promoted", input: b(0xca, 0x3f, 0x80, 0x00, 0x00), want: 1.0},

		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "float32 truncated", input: b(0xca, 0x3f, 0x80, 0x00), wantErrCode: ErrorNumberTruncated},
		{name: "float64 truncated", input: b(0xcb, 0x40, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00), wantErrCode: ErrorNumberTruncated},
		{name: "corrupted nil marker", input: b(0xc0), wantErrCode: ErrorFloatCorrupted},
		{name: "corrupted int marker", input: b(0x05), wantErrCode: ErrorFloatCorrupted},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			got, ok := runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (float64, unsafe.Pointer) { return TakeFloat64(src, lim) })
			if ok && got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}
