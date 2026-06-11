# bid754-codec for Rust

`bid754-codec` is the standalone Rust BID codec helper crate. It is not the full
`bid754-rs` decimal arithmetic implementation. Its scope is BID bit layout
encode/decode, little-endian byte encode/decode, and the shared BID codec string
format used by the cross-language vector suite.

## API

- `decode32(u32) -> Components`
- `encode32(&Components) -> u32`
- `decode64(u64) -> Components`
- `encode64(&Components) -> u64`
- `decode128(lo: u64, hi: u64) -> Components`
- `encode128(&Components) -> (u64, u64)`
- `decode32_bytes(&[u8; 4]) -> Components`
- `try_decode32_bytes(&[u8]) -> Result<Components, String>`
- `encode32_bytes(&Components) -> [u8; 4]`
- `decode64_bytes(&[u8; 8]) -> Components`
- `try_decode64_bytes(&[u8]) -> Result<Components, String>`
- `encode64_bytes(&Components) -> [u8; 8]`
- `decode128_bytes(&[u8; 16]) -> Components`
- `try_decode128_bytes(&[u8]) -> Result<Components, String>`
- `encode128_bytes(&Components) -> [u8; 16]`
- `to_string(&Components) -> String`
- `from_string(&str) -> Result<Components, String>`

The fixed-array byte decode APIs enforce byte length at the type level; the
`try_decode*_bytes` APIs return `Err` for dynamic slices with invalid lengths.
`from_string` returns `Err` for malformed payloads, malformed exponents,
multiple decimal points, empty input, and exponent values outside the signed
32-bit range. BID128 word order is `(lo, hi)`, and byte order is little-endian.

`encode32`, `encode64`, `encode128`, and byte encode helpers are
trusted-component packing APIs, not validation APIs. They canonicalize into the
target BID bit layout and may clamp exponent fields or mask/truncate
coefficient and payload fields. Invalid-`Components` rejection is not part of
the current API contract.

## Verification

From the repository root:

```sh
make test-bidcodec
make audit-bidcodec-packages
```

This crate consumes `../bid754-codec-vectors/vectors.json` through a generated
test harness during repository verification. `make audit-bidcodec-packages`
additionally checks `cargo package --locked`, docs, lints, and an external
path-consumer smoke. The package gate intentionally runs without
`--allow-dirty` so dirty tracked crate source fails instead of being packaged
silently. If the crate is tested outside the
repository without that generated vector artifact, the repo-level vector tests
skip themselves rather than depending on repository-relative files.
