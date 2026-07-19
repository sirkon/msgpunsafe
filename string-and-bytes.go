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
