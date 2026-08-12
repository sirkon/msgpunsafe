# msgpunsafe

Unsafe routines for msgpack decoding.

`msgpunsafe` (`github.com/sirkon/msgpunsafe`) is a zero-dependency Go library of
**unsafe, panic-based deserialization primitives for msgpack (MessagePack)**,
designed to power generated msgpack decoders.

This package is **LLM-friendly**. Context for AI/LLM tooling ships in two
files:

- [`llms.txt`](llms.txt) — fast-reference usage context: cursor model, API
  reference, build/test commands, architecture, and non-obvious behavior. Read
  this first when working on or with the code.
- [`llms-full.txt`](llms-full.txt) — deep conceptual reference: the *why*
  behind the panic-based error model, the `src`/`lim` cursor, `SafeBuffer`,
  zero-copy, alignment, and the generated-code contract.
