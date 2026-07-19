package msgpunsafe_test

import (
	"fmt"
	"math"
	"testing"
	"unsafe"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/sirkon/msgpunsafe"
)

func TestUint(t *testing.T) {
	tests := []struct {
		name  string
		input uint64
		// If checkOverflow is true, we expect the parser to panic with ErrorUintOverflow
		checkOverflow bool
	}{
		// --- Base critical points ---
		{name: "Zero", input: 0},
		{name: "MaxUint64", input: math.MaxUint64},

		// --- Positive Fixint boundaries (0x00 - 0x7f) ---
		{name: "MaxPositiveFixint", input: 127},
		{name: "JustAfterPositiveFixint", input: 128}, // Transitions to uint8 (0xcc)

		// --- 8-bit unsigned integer boundaries (uint8) ---
		{name: "MaxUint8", input: math.MaxUint8},
		{name: "JustAfterMaxUint8", input: math.MaxUint8 + 1}, // Transitions to uint16 (0xcd)

		// --- 16-bit unsigned integer boundaries (uint16) ---
		{name: "MaxUint16", input: math.MaxUint16},
		{name: "JustAfterMaxUint16", input: math.MaxUint16 + 1}, // Transitions to uint32 (0xce)

		// --- 32-bit unsigned integer boundaries (uint32) ---
		{name: "MaxUint32", input: math.MaxUint32},
		{name: "JustAfterMaxUint32", input: math.MaxUint32 + 1}, // Transitions to uint64 (0xcf)

		// --- 64-bit unsigned integer boundaries (uint64) ---
		{name: "BelowMaxUint64", input: math.MaxUint64 - 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal using the vmihailenco package to get real-world Msgpack payloads
			data, err := msgpack.Marshal(tt.input)
			if err != nil {
				t.Fatal(fmt.Errorf("marshal case data: %w", err))
			}

			ptr := unsafe.Pointer(unsafe.SliceData(data))

			// We use TakeUint64 as the maximum available width for unsigned integers
			got, next := msgpunsafe.TakeUint64(ptr, len(data))

			if got != tt.input {
				t.Fatalf("want %d got %d", tt.input, got)
			}

			diff := uintptr(next) - uintptr(ptr)
			if int(diff) != len(data) {
				t.Fatalf(
					"unexpected position of the returned pointer: wanted orig + %d, got orig + %d",
					len(data), diff,
				)
			}
		})
	}

	// --- Compatibility test: Signed int marker containing a positive value ---
	t.Run("SignedInt64AsPositiveUint64", func(t *testing.T) {
		// Forcing vmihailenco to emit a signed int64 marker (0xd3) instead of uint64 (0xcf)
		signedValue := int64(9223372036854775807) // math.MaxInt64
		data, err := msgpack.Marshal(signedValue)
		if err != nil {
			t.Fatal(err)
		}

		ptr := unsafe.Pointer(unsafe.SliceData(data))
		// Our TakeUint64 must transparently handle the 0xd3 marker since the value is positive
		got, next := msgpunsafe.TakeUint64(ptr, len(data))

		if got != uint64(signedValue) {
			t.Fatalf("compatibility check failed: want %d got %d", signedValue, got)
		}

		diff := uintptr(next) - uintptr(ptr)
		if int(diff) != len(data) {
			t.Fatalf("unexpected pointer shift: want %d, got %d", len(data), diff)
		}
	})

	// --- Negative overflow test: Signed int marker containing a negative value ---
	t.Run("NegativeInt64OverflowPanic", func(t *testing.T) {
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic on negative value conversion to uint64, but got none")
			}
			errCode, ok := r.(msgpunsafe.ErrorCode)
			if !ok || errCode != msgpunsafe.ErrorUintOverflow {
				t.Fatalf("expected ErrorUintOverflow panic, but got: %v", r)
			}
		}()

		negativeValue := int64(-42)
		data, err := msgpack.Marshal(negativeValue)
		if err != nil {
			t.Fatal(err)
		}

		ptr := unsafe.Pointer(unsafe.SliceData(data))
		// This must panic because uint64 cannot store negative values
		_, _ = msgpunsafe.TakeUint64(ptr, len(data))
	})
}

func TestInt(t *testing.T) {
	tests := []int{
		// --- Base critical points ---
		0,
		math.MaxInt,
		math.MinInt,

		// --- Positive Fixint boundaries (0x00 - 0x7f) ---
		127, // Maximum value for positive fixint (0x7f)
		128, // Just after fixint -> encodes to uint8 (0xcc) or int8 (0xd0)

		// --- Negative Fixint boundaries (0xe0 - 0xff) ---
		-1,  // Minimum absolute value for negative fixint (0xff)
		-32, // Maximum absolute value for negative fixint (0xe0)
		-33, // Just after negative fixint -> encodes to int8 (0xd0)

		// --- 8-bit integer boundaries (int8 / uint8) ---
		math.MaxInt8,      // 127 (already covered, but included for structure)
		math.MaxInt8 + 1,  // 128 (header change transition point)
		math.MaxUint8,     // 255 (maximum value for uint8 / 0xcc)
		math.MaxUint8 + 1, // 256 -> transitions to uint16 (0xcd) or int16 (0xd1)

		math.MinInt8,     // -128 (minimum value for int8 / 0xd0)
		math.MinInt8 - 1, // -129 -> transitions to int16 (0xd1)

		// --- 16-bit integer boundaries (int16 / uint16) ---
		math.MaxInt16,      // 32767 (maximum value for int16)
		math.MaxInt16 + 1,  // 32768 -> transitions to uint16 or int32
		math.MaxUint16,     // 65535 (maximum value for uint16 / 0xcd)
		math.MaxUint16 + 1, // 65536 -> transitions to uint32 (0xce) or int32 (0xd2)

		math.MinInt16,     // -32768 (minimum value for int16 / 0xd1)
		math.MinInt16 - 1, // -32769 -> transitions to int32 (0xd2)

		// --- 32-bit integer boundaries (int32 / uint32) ---
		math.MaxInt32,      // 2147483647 (maximum value for int32)
		math.MaxInt32 + 1,  // 2147483648 -> transitions to uint32 or int64
		math.MaxUint32,     // 4294967295 (maximum value for uint32 / 0xce)
		math.MaxUint32 + 1, // 4294967296 -> transitions to uint64 (0xcf) or int64 (0xd3)

		math.MinInt32,     // -2147483648 (minimum value for int32 / 0xd2)
		math.MinInt32 - 1, // -2147483649 -> transitions to int64 (0xd3)

		// --- 64-bit integer boundaries (int64) ---
		math.MaxInt64 - 1, // One step below absolute maximum of int64
		// math.MaxInt64 and math.MinInt64 are already covered by math.MaxInt / math.MinInt on 64-bit platforms
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt), func(t *testing.T) {
			data, err := msgpack.Marshal(tt)
			if err != nil {
				t.Fatal(fmt.Errorf("marshal case data: %w", err))
			}

			ptr := unsafe.Pointer(unsafe.SliceData(data))
			got, next := msgpunsafe.TakeInt(ptr, len(data))

			if got != tt {
				t.Fatalf("want %d got %d", tt, got)
			}
			diff := uintptr(next) - uintptr(ptr)
			if int(diff) != len(data) {
				t.Fatalf(
					"unexpected position of the returned pointer: wanted orig + %d, got orig + %d",
					len(data), diff,
				)
			}
		})
	}
}
