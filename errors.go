package msgpunsafe

import (
	"strconv"
)

func HandleError(r any, err *error) error {
	if e, ok := r.(ErrorCode); ok {
		if e.IsValid() {
			*err = e
		}
	}

	panic(r)
}

type ErrorCode int

const (
	ErrorSliceExhausted ErrorCode = iota
	ErrorSliceHeaderTruncated
	ErrorSliceHeaderCorrupted

	ErrorMapExhausted
	ErrorMapHeaderTruncated
	ErrorMapHeaderCorrupted

	ErrorStrExhausted
	ErrorStrHeaderTruncated
	ErrorStrHeaderCorrupted
	ErrorStrLenConflict

	ErrorBinExhausted
	ErrorBinHeaderTruncated
	ErrorBinHeaderCorrupted
	ErrorBinLenConflict

	ErrorNumberExhausted
	ErrorNumberTruncated
	ErrorIntCorrupted
	ErrorUintCorrupted
	ErrorFloatCorrupted
	ErrorIntOverflow
	ErrorUintOverflow

	ErrorBoolExhausted
	ErrorBoolCorrupted

	// Errors to be raised in generated code.

	ErrorStructExhausted
	ErrorStructCorrupted
	ErrorUnknownField
	ErrorRequiredFieldMissing
)

var _ error = ErrorCode(0)

var errorMessages = [...]string{
	ErrorSliceExhausted:       "msgpack: slice data exhausted",
	ErrorSliceHeaderTruncated: "msgpack: slice header is truncated",
	ErrorSliceHeaderCorrupted: "msgpack: slice header is corrupted",

	ErrorMapExhausted:       "msgpack: map data exhausted",
	ErrorMapHeaderTruncated: "msgpack: map header is truncated",
	ErrorMapHeaderCorrupted: "msgpack: map header is corrupted",

	ErrorStrExhausted:       "msgpack: string data exhausted",
	ErrorStrHeaderTruncated: "msgpack: string header is truncated",
	ErrorStrHeaderCorrupted: "msgpack: string header is corrupted",
	ErrorStrLenConflict:     "msgpack: string length conflicts with remaining buffer size",

	ErrorBinExhausted:       "msgpack: binary data exhausted",
	ErrorBinHeaderTruncated: "msgpack: binary header is truncated",
	ErrorBinHeaderCorrupted: "msgpack: binary header is corrupted",
	ErrorBinLenConflict:     "msgpack: binary length conflicts with remaining buffer size",

	ErrorNumberExhausted: "msgpack: numeric data exhausted",
	ErrorNumberTruncated: "msgpack: numeric value is truncated",
	ErrorIntCorrupted:    "msgpack: int data corrupted",
	ErrorUintCorrupted:   "msgpack: uint data corrupted",
	ErrorFloatCorrupted:  "msgpack: float data corrupted",
	ErrorIntOverflow:     "msgpack: integer value overflows target type range",
	ErrorUintOverflow:    "msgpack: unsigned integer value overflows target type range",

	ErrorBoolExhausted: "msgpack: boolean data exhausted",
	ErrorBoolCorrupted: "msgpack: invalid boolean marker",

	ErrorStructExhausted:      "msgpack: struct data exhausted",
	ErrorStructCorrupted:      "msgpack: struct data corrupted",
	ErrorUnknownField:         "msgpack: unknown field encountered in structured data",
	ErrorRequiredFieldMissing: "msgpack: required structural field is missing in msgpack map",
}

func (e ErrorCode) Error() string {
	if int(e) >= 0 && int(e) < len(errorMessages) {
		return errorMessages[e]
	}

	panic("must not be here with error code " + strconv.Itoa(int(e)))
}

func (e ErrorCode) IsValid() bool {
	return int(e) >= 0 && int(e) < len(errorMessages)
}

func panicWithError(code ErrorCode) {
	panic(code)
}
