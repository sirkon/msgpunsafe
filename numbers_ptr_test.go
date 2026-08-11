package msgpunsafe

import (
	"math"
	"testing"
	"unsafe"
)

// checkPtr asserts the pointer points inside the buffer and is 8-byte aligned.
func checkPtr(t *testing.T, name string, sBuf *SafeBuffer, p unsafe.Pointer) {
	t.Helper()
	if uintptr(p)%8 != 0 {
		t.Fatalf("%s: pointer %x is not 8-byte aligned", name, uintptr(p))
	}
	if !ptrInBuf(sBuf, p) {
		t.Fatalf("%s: pointer %x must point into the safe buffer", name, uintptr(p))
	}
}

func TestTakeUint64Ptr(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        uint64
		wantErrCode ErrorCode
	}
	tests := []testCase{
		{name: "fixint", input: b(0x2a), want: 42},
		{name: "uint8", input: b(0xcc, 0xff), want: 255},
		{name: "uint16", input: b(0xcd, 0x01, 0x00), want: 256},
		{name: "uint32", input: b(0xce, 0x00, 0x01, 0x00, 0x00), want: 0x10000},
		{name: "uint64 max", input: b(0xcf, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff), want: math.MaxUint64},
		{name: "int32 positive", input: b(0xd2, 0x00, 0x01, 0x86, 0xa0), want: 100000},
		// Delegated panics from TakeUint64.
		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
		{name: "negative fixint", input: b(0xff), wantErrCode: ErrorUintOverflow},
		{name: "uint8 truncated", input: b(0xcc), wantErrCode: ErrorNumberTruncated},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			sBuf := NewSafeBuffer(512)
			var got *uint64
			runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (uint64, unsafe.Pointer) {
					var next unsafe.Pointer
					got, next = TakeUint64Ptr(src, lim, sBuf)
					return uint64(0), next
				})
			if tc.wantErrCode != 0 {
				if len(sBuf.buf) != 0 {
					t.Fatalf("%s: buffer must stay untouched on error, len = %d", tc.name, len(sBuf.buf))
				}
				return
			}
			if got == nil || *got != tc.want {
				t.Fatalf("got %v, want %d", got, tc.want)
			}
			checkPtr(t, tc.name, sBuf, unsafe.Pointer(got))
		})
	}
}

func TestTakeInt32Ptr(t *testing.T) {
	type testCase struct {
		name        string
		input       []byte
		want        int32
		wantErrCode ErrorCode
	}
	tests := []testCase{
		{name: "fixint", input: b(0x2a), want: 42},
		{name: "negative fixint", input: b(0xe0), want: -32},
		{name: "int32", input: b(0xd2, 0x00, 0x01, 0x86, 0xa0), want: 100000},
		{name: "uint8 small", input: b(0xcc, 0x64), want: 100},
		// Narrowing overflow of int32.
		{name: "overflow int64", input: b(0xd3, 0x7f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff), wantErrCode: ErrorIntOverflow},
		{name: "exhausted", input: b(), wantErrCode: ErrorNumberExhausted},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src, lim := ptrs(tc.input)
			sBuf := NewSafeBuffer(512)
			var got *int32
			runTake(t, tc.name, tc.wantErrCode, tc.input, src,
				func() (int32, unsafe.Pointer) {
					var next unsafe.Pointer
					got, next = TakeInt32Ptr(src, lim, sBuf)
					return int32(0), next
				})
			if tc.wantErrCode != 0 {
				return
			}
			if got == nil || *got != tc.want {
				t.Fatalf("got %v, want %d", got, tc.want)
			}
			checkPtr(t, tc.name, sBuf, unsafe.Pointer(got))
		})
	}
}

func TestTakeFloat32Ptr(t *testing.T) {
	src, lim := ptrs(b(0xca, 0x3f, 0x80, 0x00, 0x00))
	sBuf := NewSafeBuffer(512)
	var got *float32
	runTake(t, "float32", 0, b(0xca, 0x3f, 0x80, 0x00, 0x00), src,
		func() (float32, unsafe.Pointer) {
			var next unsafe.Pointer
			got, next = TakeFloat32Ptr(src, lim, sBuf)
			return 0, next
		})
	if got == nil || *got != 1.0 {
		t.Fatalf("got %v, want 1.0", got)
	}
	checkPtr(t, "float32", sBuf, unsafe.Pointer(got))
}

func TestTakeFloat64Ptr(t *testing.T) {
	src, lim := ptrs(b(0xcb, 0x3f, 0xf0, 0, 0, 0, 0, 0, 0))
	sBuf := NewSafeBuffer(512)
	var got *float64
	runTake(t, "float64", 0, b(0xcb, 0x3f, 0xf0, 0, 0, 0, 0, 0, 0), src,
		func() (float64, unsafe.Pointer) {
			var next unsafe.Pointer
			got, next = TakeFloat64Ptr(src, lim, sBuf)
			return 0, next
		})
	if got == nil || *got != 1.0 {
		t.Fatalf("got %v, want 1.0", got)
	}
	checkPtr(t, "float64", sBuf, unsafe.Pointer(got))
}

func TestTakePtr_MixedAlignment(t *testing.T) {
	// Interleave many scalar types from one stream into one SafeBuffer and
	// assert every returned pointer stays 8-byte aligned.
	sBuf := NewSafeBuffer(1024)

	inputs := [][]byte{
		b(0xc3),                         // true
		b(0xcc, 0x01),                   // uint8 1
		b(0xd3, 0, 0, 0, 0, 0, 0, 0, 1), // int64 1
		b(0xca, 0x3f, 0x80, 0x00, 0x00), // float32 1.0
		b(0x2a),                         // fixint 42
	}

	var ptrsList []unsafe.Pointer

	for i, input := range inputs {
		src, lim := ptrs(input)
		switch input[0] {
		case 0xc3:
			var p *bool
			runTake(t, "bool", 0, input, src, func() (bool, unsafe.Pointer) {
				var next unsafe.Pointer
				p, next = TakeBoolPtr(src, lim, sBuf)
				return false, next
			})
			ptrsList = append(ptrsList, unsafe.Pointer(p))
		case 0xcc:
			var p *uint8
			runTake(t, "uint8", 0, input, src, func() (uint8, unsafe.Pointer) {
				var next unsafe.Pointer
				p, next = TakeUint8Ptr(src, lim, sBuf)
				return 0, next
			})
			ptrsList = append(ptrsList, unsafe.Pointer(p))
		case 0xd3:
			var p *int64
			runTake(t, "int64", 0, input, src, func() (int64, unsafe.Pointer) {
				var next unsafe.Pointer
				p, next = TakeInt64Ptr(src, lim, sBuf)
				return 0, next
			})
			ptrsList = append(ptrsList, unsafe.Pointer(p))
		case 0xca:
			var p *float32
			runTake(t, "float32", 0, input, src, func() (float32, unsafe.Pointer) {
				var next unsafe.Pointer
				p, next = TakeFloat32Ptr(src, lim, sBuf)
				return 0, next
			})
			ptrsList = append(ptrsList, unsafe.Pointer(p))
		default:
			var p *int
			runTake(t, "int", 0, input, src, func() (int, unsafe.Pointer) {
				var next unsafe.Pointer
				p, next = TakeIntPtr(src, lim, sBuf)
				return 0, next
			})
			ptrsList = append(ptrsList, unsafe.Pointer(p))
		}
		if i > 0 {
			if len(sBuf.buf) == 0 {
				t.Fatal("buffer len unexpectedly reset")
			}
		}
	}

	for i, p := range ptrsList {
		if uintptr(p)%8 != 0 {
			t.Fatalf("pointer %d (%x) is not 8-byte aligned", i, uintptr(p))
		}
		if !ptrInBuf(sBuf, p) {
			t.Fatalf("pointer %d (%x) must point into the safe buffer", i, uintptr(p))
		}
	}
}
