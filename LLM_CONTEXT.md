# AGENTS.md

## What this is

`msgpunsafe` is a zero-dependency Go module (`github.com/sirkon/msgpunsafe`) that provides **unsafe, panic-based deserialization primitives for msgpack** (MessagePack). It is a library: there is no executable, no CLI, no main package. Consumer-facing API functions are named `Take*` (read one value) and operate with a `src`/`lim` pointer-pair cursor model over a raw byte buffer. It is intended to power generated msgpack decoders (the `ErrorStruct*`/`ErrorRequiredFieldMissing` codes exist for that purpose; the generated-code consumer is not in this repo).

## Commands

```sh
go test ./...              # tests, default build (non-aligned readout)
go test -tags=aware_of_alignment ./...   # tests, alignment-safe readout
go vet ./...               # clean
go vet -tags=aware_of_alignment ./...    # clean
```

- Go 1.26 (see `go.mod`). Code relies on Go 1.22+ language features: `for range N` integer loops and Go 1.25's `sync.Go` (used in safebuf_test.go).
- No CI, no Makefile, no linter config. `go test` is the only gate.
- **Gotcha: `go test -race ./...` crashes** with `fatal error: checkptr: pointer arithmetic result points to invalid allocation`, originating in `ptrs()` (testhelp_test.go:24) where `unsafe.Add(p, len(input))` computes the `lim` bound past the slice's real allocation. This is a test-helper limitation, not a library bug; plain and `-tags=aware_of_alignment` runs pass.

## Architecture and control flow

Decoding is cursor-based, represented as a pointer pair:

- `src unsafe.Pointer` — current read position.
- `lim unsafe.Pointer` — one-past-the-end bound of the whole buffer.

Every `Take*` function reads from `src`, returns `(value, nextPtr)` where `nextPtr` is the advanced cursor. Bounds checks are done by comparing `uintptr(src)` against `uintptr(lim)` *before* dereferencing. Multi-byte reads additionally verify `uintptr(lim)-uintptr(src) >= size+1`.

### Files

| File | Contents |
|---|---|
| `headers.go` | `TakeSliceHeader`, `TakeMapHeader`, `TakeStrHeader`, `TakeBinHeader` — read container/str/bin headers, return `(countOrSize, nextPtr)`. |
| `numbers.go` | Typed wrappers `TakeUint`, `TakeUint8…64`, `TakeInt`, `TakeInt8…64`, `TakeFloat32/64` — each calls the internal `takeUint64`/`takeInt64`/`takeFloat32`/`takeFloat64` and narrows with an overflow `panic` if the value doesn't fit the target width. |
| `numbers_base.go` | The actual numeric decoding. **Build tag `//go:build !aware_of_alignment` (default).** Reads multi-byte big-endian values by casting `*(*uint16/32/64)(unsafe.Add(src,1))` and byte-swapping with `math/bits.ReverseBytes*`. Faster, but unsafe if the underlying buffer may be unaligned. |
| `numbers_base_align.go` | **Build tag `//go:build aware_of_alignment`.** Same `take*` functions, but reads via `binary.BigEndian` over `unsafe.Slice` — alignment-safe, slightly slower. Must be kept in lockstep with `numbers_base.go` (same function names/signatures/error codes). |
| `bool.go` | `TakeBool` — `0xc2`/`0xc3` markers. `TakeBoolPtr` — bool stored in an 8-byte-aligned SafeBuffer slot. |
| `nil.go` | `IsNil` — checks `0xc0`, returns `(bool, ptr)` where the pointer is unchanged on non-nil (unusual: all other `Take*` return the *next* pointer). |
| `string_and_bytes.go` | `TakeString`, `TakeStringZC` (zero-copy, points **into** the caller's buffer — caller must keep it alive), `TakeBytes`. Copy variants take a `*SafeBuffer`. All three check `dataPtr+len > lim` and panic `Error*LenConflict`. |
| `safebuf.go` | `SafeBuffer` — a reusable arena that holds copies of strings/bytes so they outlive the temporary source buffer (e.g., a Tarantool response buffer). |
| `numbers_ptr.go` | `Take*Ptr` for every int/uint/float width — parse via `Take*`, store the value in an 8-byte-aligned SafeBuffer slot (`AllocAligned`), return `(*T, nextPtr)`. One generic pattern; `TakeBoolPtr` is in bool.go. |
| `skip.go` | `TakeSkip` — skips one complete msgpack object (containers recursively) using an explicit stack, zero allocation. Own error family `ErrorSkip*`, includes a recursion-depth limit (`skipRecursionLimit = 100000`). |
| `errors.go` | `ErrorCode` (`int`) constants, error messages, `HandleError` (recover helper: converts a valid `ErrorCode` panic into `*err`, repanics anything else — designed for use in `defer`), `panicWithError`. |

### Error model (important)

- All failures are **panics with `ErrorCode` values**, never returned errors. Callers are expected to use `HandleError` in a deferred recover, or generated code handles them.
- Error categories per type: `*Exhausted` (nothing left to read at all), `*HeaderTruncated` (marker present, its payload bytes missing), `*HeaderCorrupted` (wrong marker byte), `*LenConflict` (declared length exceeds remaining buffer), `*Overflow` (value doesn't fit target type), `*Corrupted` (invalid marker after disambiguation).
- `ErrorCode.Error()` and `IsValid()` panic on out-of-range codes — `HandleError` relies on `IsValid()` to distinguish real msgs errors from unrelated panics and re-panics the latter.
- `ErrorStructExhausted`, `ErrorStructCorrupted`, `ErrorUnknownField`, `ErrorRequiredFieldMissing` are placeholders for the (external) generated-code layer; nothing in this repo raises them yet. `ErrorSkip*` codes are raised by `TakeSkip` only.
- Adding a new `ErrorCode`: append at the **end** of the const block and extend `errorMessages` in the same index order. `ErrorSliceExhausted` is `iota == 0` and tests use `wantErrCode == 0` as "no panic expected", so a new zero code would break that convention.

### SafeBuffer semantics

`NewSafeBuffer(cap)` preallocates a backing array. `AllocString`/`AllocBytes` payoffs:

- If `size > cap`: immediate `strings.Clone`/`bytes.Clone` — independent allocation, buffer untouched.
- Else if `size+len(buf) > cap`: a **fresh backing array** is allocated and the cursor resets to 0 (records already returned keep pointing into the previous array, which stays alive via the returned string/slice headers).
- Else: appends into the current array; returns `unsafe.String`/`unsafe.Slice` views into it.
- `AllocAligned(size)` (used by `Take*Ptr`) reserves a slot whose address is a multiple of 8, inserting padding before it as needed. Its direct fallback triggers at `size+7 > cap` (not `size > cap`, unlike the string/bytes clone path) so a reset buffer can never loop forever. Callers must write the slot immediately after allocation.

Doc comment on `NewSafeBuffer`: prefer capacities ≤ 512 and reuse the buffer.

### Int/uint cross-parsing (non-obvious)

The `take*` functions deliberately accept the "wrong" signedness for compatibility: `takeUint64` parses `0xd0–0xd3` (int markers) as unsigned, panicking `ErrorUintOverflow` on negative values, and `takeInt64` parses `0xcc–0xcf` (uint markers), panicking `ErrorIntOverflow` if the value exceeds `math.MaxInt64`. Header parsing (`TakeSliceHeader` etc.) has a similar open TODO about distinguishing wrong-type markers — currently a wrong marker type is just `*HeaderCorrupted`.

## Conventions

- Public API: `Take*` + full doc comments (godoc is the only documentation; README is a stub). Internal helpers are lowercase: `take*`, `panicWithError`, `HandleError` is public for consumers.
- Every functions returns the value plus the advanced `unsafe.Pointer` (`IsNil` is the exception: unchanged pointer on non-nil).
- Multi-byte big-endian decoding in the default build uses `bits.ReverseBytes*` on native-endian casts; the alignment-safe twin uses `encoding/binary.BigEndian`. **When changing numeric decoding, edit both `numbers_base.go` and `numbers_base_align.go` and test with both tags.**
- Bounds-checked reads compare `uintptr(lim)-uintptr(src)` against the total bytes needed (lead byte + payload).
- gopls shows a warning for `numbers_base_align.go` ("No packages found… may be excluded due to its build tags") — expected; configure `-tags=aware_of_alignment` in gopls build flags if you edit it frequently.

## Testing

- One `_test.go` per source file plus `testhelp_test.go`.
- `testhelp_test.go` provides the shared helpers; reuse them, do not duplicate:
  - `b(...byte) []byte` — compact msgpack payload construction.
  - `ptrs(input) (src, lim)` — pointer pair; empty input yields two equal pointers (the "exhausted" precondition). Note: incompatible with `-race` (see above).
  - `runTake[T](t, name, wantErrCode, input, src, fn)` — generically runs any `Take*` under `recover`, asserting: exact `ErrorCode` on panic (`wantErrCode != 0`), or full buffer consumption (`next == src+len(input)`) without panic.
- Tests are table-driven per typed wrapper (`TestTakeUint64`, `TestTakeUint32`, …), each covering happy path (every encoding form), truncation (each multi-byte marker), corruption, exhaustion, and narrowing overflow. `numbers_test.go` is the model to copy.
- Unsafe-pointer behavior is verified structurally: e.g. `safebuf_test.go` asserts returned pointers physically live inside (or outside) the `SafeBuffer` backing array (`ptrInBuf`), and GC retention tests hammer `runtime.GC()` after dropping sources.
- Run the suite with both build tags; the two numeric backends must pass identically.