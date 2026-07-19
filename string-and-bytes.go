package msgpunsafe

import (
	"unsafe"
)

// TakeString reads a string from Msgpack, copies its content into the SafeBuffer,
// and returns a long-lived string that safely survives the temporary Tarantool buffer.
func TakeString(src unsafe.Pointer, lim int, sBuf *SafeBuffer) (string, unsafe.Pointer) {
	strLen, dataPtr := TakeStrHeader(src, lim)

	headerLen := int(uintptr(dataPtr) - uintptr(src))
	if lim < headerLen+strLen {
		panic(ErrorStrLenConflict)
	}

	safeStr := sBuf.AllocString(dataPtr, strLen)

	return safeStr, unsafe.Add(dataPtr, strLen)
}

// TakeStringZC reads a string from Msgpack using Zero-Copy.
// The returned string points directly into the incoming 'src' memory buffer.
// WARNING: The string is only valid as long as the underlying source buffer is not cleared or reused.
func TakeStringZC(src unsafe.Pointer, lim int) (string, unsafe.Pointer) {
	// 1. Parse the string header to get the length and data pointer
	strLen, dataPtr := TakeStrHeader(src, lim)

	// Calculate how many bytes the header consumed
	headerLen := int(uintptr(dataPtr) - uintptr(src))

	// Ensure the actual string data fits within the remaining buffer
	if lim < headerLen+strLen {
		panic(ErrorStrLenConflict)
	}

	// 2. Create an immutable string looking directly into the incoming data buffer
	res := unsafe.String((*byte)(dataPtr), strLen)

	// 3. Return the zero-copy string and the shifted pointer past the string content
	return res, unsafe.Add(dataPtr, strLen)
}

// TakeBytes reads a byte slice from Msgpack, copies its content into the SafeBuffer,
// and returns a long-lived []byte that safely survives the temporary Tarantool buffer.
func TakeBytes(src unsafe.Pointer, lim int, sBuf *SafeBuffer) ([]byte, unsafe.Pointer) {
	binLen, dataPtr := TakeBinHeader(src, lim)

	headerLen := int(uintptr(dataPtr) - uintptr(src))
	if lim < headerLen+binLen {
		panic(ErrorBinLenConflict)
	}

	safeBytes := sBuf.AllocBytes(dataPtr, binLen)

	// 3. Возвращаем слайс и сдвинутый указатель
	return safeBytes, unsafe.Add(dataPtr, binLen)
}
