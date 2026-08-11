package msgpunsafe

import (
	"testing"
)

// runSkip executes TakeSkip on input, expecting it to land exactly at
// wantOff bytes into the buffer. When an error code is given, it asserts a
// panic with exactly that code instead.
func runSkip(t *testing.T, name string, wantErrCode ErrorCode, input []byte, wantOff int) {
	t.Helper()

	src, lim := ptrs(input)

	defer func() {
		r := recover()
		if wantErrCode == 0 {
			if r != nil {
				t.Fatalf("%s: unexpected panic: %v", name, r)
			}
			return
		}
		code, ok := r.(ErrorCode)
		if !ok {
			t.Fatalf("%s: panic with non-ErrorCode: %v", name, r)
		}
		if code != wantErrCode {
			t.Fatalf("%s: expected error code %d, got %d", name, wantErrCode, code)
		}
	}()

	next := TakeSkip(src, lim)

	if wantErrCode != 0 {
		t.Fatalf("%s: expected panic with code %d, but no panic occurred", name, wantErrCode)
	}
	if got := int(uintptr(next) - uintptr(src)); got != wantOff {
		t.Fatalf("%s: consumed %d bytes, want %d", name, got, wantOff)
	}
}

func TestTakeSkip_Scalars(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		// fixints
		{"positive fixint low", b(0x00)},
		{"positive fixint high", b(0x7f)},
		{"negative fixint low", b(0xe0)},
		{"negative fixint high", b(0xff)},
		// ints
		{"int8", b(0xd0, 0x01)},
		{"int16", b(0xd1, 0x00, 0x01)},
		{"int32", b(0xd2, 0x00, 0x00, 0x00, 0x01)},
		{"int64", b(0xd3, 0, 0, 0, 0, 0, 0, 0, 1)},
		// uints
		{"uint8", b(0xcc, 0x01)},
		{"uint16", b(0xcd, 0x00, 0x01)},
		{"uint32", b(0xce, 0x00, 0x00, 0x00, 0x01)},
		{"uint64", b(0xcf, 0, 0, 0, 0, 0, 0, 0, 1)},
		// floats
		{"float32", b(0xca, 0x3f, 0x80, 0x00, 0x00)},
		{"float64", b(0xcb, 0x3f, 0xf0, 0, 0, 0, 0, 0, 0)},
		// nil and bools
		{"nil", b(0xc0)},
		{"false", b(0xc2)},
		{"true", b(0xc3)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runSkip(t, tc.name, 0, tc.input, len(tc.input))
		})
	}
}

func TestTakeSkip_StringsAndBins(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"fixstr empty", b(0xa0)},
		{"fixstr", append(b(0xa2), 'h', 'i')},
		{"str8", b(0xd9, 0x03, 'a', 'b', 'c')},
		{"str16", b(0xda, 0x00, 0x01, 'x')},
		{"str32", b(0xdb, 0x00, 0x00, 0x00, 0x01, 'y')},
		{"bin8", b(0xc4, 0x02, 'h', 'i')},
		{"bin16", b(0xc5, 0x00, 0x03, 'a', 'b', 'c')},
		{"bin32", b(0xc6, 0x00, 0x00, 0x00, 0x01, 'x')},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runSkip(t, tc.name, 0, tc.input, len(tc.input))
		})
	}
}

func TestTakeSkip_Extensions(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"fixext1", b(0xd4, 0x05, 0x01)},
		{"fixext2", b(0xd5, 0x05, 0x01, 0x02)},
		{"fixext4", b(0xd6, 0x05, 0x01, 0x02, 0x03, 0x04)},
		{"fixext8", b(0xd7, 0x05, 0, 0, 0, 0, 0, 0, 0, 0)},
		{"fixext16", append(b(0xd8, 0x05), make([]byte, 16)...)},
		{"ext8", b(0xc7, 0x02, 0x05, 0xaa, 0xbb)},
		{"ext16", b(0xc8, 0x00, 0x02, 0x05, 0xaa, 0xbb)},
		{"ext32", b(0xc9, 0x00, 0x00, 0x00, 0x02, 0x05, 0xaa, 0xbb)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runSkip(t, tc.name, 0, tc.input, len(tc.input))
		})
	}
}

func TestTakeSkip_Containers(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		// Flat arrays and maps of scalars.
		{"fixarray empty", b(0x90)},
		{"fixarray", b(0x93, 0x01, 0x02, 0x03)},
		{"array16", b(0xdc, 0x00, 0x02, 0x01, 0x02)},
		{"array32", b(0xdd, 0x00, 0x00, 0x00, 0x02, 0x01, 0x02)},
		{"fixmap empty", b(0x80)},
		{"fixmap", b(0x81, 0xa1, 'a', 0x01)},
		{"map16", b(0xde, 0x00, 0x01, 0xa1, 'b', 0x02)},
		{"map32", b(0xdf, 0x00, 0x00, 0x00, 0x01, 0xa1, 'c', 0x03)},
		// Nested containers.
		{"nested array in map", b(0x81, 0xa1, 'k', 0x92, 0x01, 0x02)},
		{"deeply nested", b(0x91, 0x91, 0x91, 0x91, 0x01)},
		{"map of arrays", b(0x82, 0xa1, 'a', 0x92, 0x01, 0x02, 0xa1, 'b', 0x92, 0x03, 0x04)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runSkip(t, tc.name, 0, tc.input, len(tc.input))
		})
	}
}

func TestTakeSkip_TrailingData(t *testing.T) {
	// The reader must stop exactly at the boundary of the first object,
	// ignoring whatever follows.
	input := b(0x92, 0x01, 0x02, 0xca, 0x3f, 0x80, 0x00, 0x00, 0x05)
	runSkip(t, "array then more", 0, input, 3)
}

func TestTakeSkip_Errors(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		wantErrCode ErrorCode
	}{
		{"exhausted", b(), ErrorSkipExhausted},
		{"corrupted 0xc1", b(0xc1), ErrorSkipCorrupted},
		{"uint16 truncated", b(0xcd), ErrorSkipHeaderTruncated},
		{"bin8 truncated header", b(0xc4), ErrorSkipHeaderTruncated},
		{"bin16 truncated header", b(0xc5, 0x00), ErrorSkipHeaderTruncated},
		{"str8 truncated header", b(0xd9), ErrorSkipHeaderTruncated},
		{"str32 truncated header", b(0xdb, 0x00, 0x00), ErrorSkipHeaderTruncated},
		{"array16 truncated header", b(0xdc, 0x00), ErrorSkipHeaderTruncated},
		{"map16 truncated header", b(0xde, 0x00), ErrorSkipHeaderTruncated},
		{"ext32 truncated header", b(0xc9, 0x00, 0x00, 0x00), ErrorSkipHeaderTruncated},
		{"str8 len conflict", b(0xd9, 0x05, 'a', 'b'), ErrorSkipLenConflict},
		{"bin16 len conflict", b(0xc5, 0x00, 0x05, 'a'), ErrorSkipLenConflict},
		{"array element truncated", b(0x92, 0x01), ErrorSkipExhausted},
		{"nested len conflict", b(0x91, 0xda, 0x00, 0x05), ErrorSkipLenConflict},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runSkip(t, tc.name, tc.wantErrCode, tc.input, 0)
		})
	}
}
