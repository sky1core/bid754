# bid754-codec for Rust

`bid754-codec` is the standalone Rust BID codec helper crate. It is not the full
`bid754-rs` decimal arithmetic implementation. Its scope is BID bit layout
encode/decode, little-endian byte encode/decode, and the shared BID codec string
format used by the cross-language vector suite.

## API

- `decode32(u32) -> Components`
- `encode32(&Components) -> Result<u32, String>`
- `decode64(u64) -> Components`
- `encode64(&Components) -> Result<u64, String>`
- `decode128(lo: u64, hi: u64) -> Components`
- `encode128(&Components) -> Result<(u64, u64), String>`
- `decode32_bytes(&[u8; 4]) -> Components`
- `try_decode32_bytes(&[u8]) -> Result<Components, String>`
- `encode32_bytes(&Components) -> Result<[u8; 4], String>`
- `decode64_bytes(&[u8; 8]) -> Components`
- `try_decode64_bytes(&[u8]) -> Result<Components, String>`
- `encode64_bytes(&Components) -> Result<[u8; 8], String>`
- `decode128_bytes(&[u8; 16]) -> Components`
- `try_decode128_bytes(&[u8]) -> Result<Components, String>`
- `encode128_bytes(&Components) -> Result<[u8; 16], String>`
- `to_string(&Components) -> Result<String, String>`
- `from_string(&str) -> Result<Components, String>`

The fixed-array byte decode APIs enforce byte length at the type level; the
`try_decode*_bytes` APIs return `Err` for dynamic slices with invalid lengths.
BID128 word order is `(lo, hi)`, and byte order is little-endian.

`from_string` accepts one strict ASCII grammar. The whole input must be ASCII:
any non-ASCII byte anywhere (Unicode digit variants, Unicode whitespace) is
rejected, and only ASCII whitespace may surround the token. After an optional
single leading `+`/`-` the input is either `Inf`/`Infinity`, or `NaN`/`SNaN`
with an optional unsigned ASCII-digit payload (case-insensitive, whose value
must be below the schema-wide NaN payload limit `10^33`), or a number: ASCII
digits with at most one `.` and at least one digit, optionally followed by
`E`/`e`, one optional sign, and ASCII exponent digits where the exponent
literal must be below the shared exact-integer bound 2^53 in magnitude (the
widest bound every consumer's number type can check exactly; a literal at or
beyond it is rejected through the same error channel) and the
fraction-adjusted final exponent must
fit a signed 32-bit integer. The literal itself may exceed int32: `to_string`
renders adjusted exponents up to i32::MAX + 33 (far below 2^53), and every
`to_string` output reparses through `from_string` (round-trip closure).
Underscores and payload-internal signs are malformed. `from_string` validates
grammar and schema limits only, identical in all six language packages: the
parsed coefficient value must not exceed the schema-wide maximum coefficient
`10^34-1` (the largest value any supported BID width can hold — a schema
constant, not per-width validation), and the parsed payload value must be below
the schema-wide NaN payload limit `10^33` (the widest canonical BID128 NaN
payload, the same kind of schema constant). BID width-range validation stays in
the encode contract below.

## String render reject contract

**Breaking change:** `to_string` now returns `Result<String, String>`, and
`Components` no longer implements `Display`: formatting cannot carry an error
without causing a panic. The renderer accepts only the width-independent shared
schema: a `Normal` coefficient is in `1..10^34-1`; a NaN payload is in
`0..10^33-1`; and every field the selected `Kind` cannot represent is zero.
Invalid values return `Err` before rendering, so every successful rendering
reparses without kind or component loss.

## Encode reject contract

**Breaking change:** `encode32`/`encode64`/`encode128` and the byte encode
helpers now return `Result<_, String>` (previously they returned the packed
value directly). They are validating packing APIs: no input reachable through
the public API fails silently. Instead of truncating, masking, clamping, or
panicking, they return `Err` for any `Components` field not representable in the
target BID width exactly as supplied. The reject boundaries are field-level, per
width:

- `Normal` coefficient above the width maximum
  (`9_999_999` / `9_999_999_999_999_999` / `10^34-1`)
- `Zero`/`Normal` unbiased exponent outside the width range
  (`[-101, 90]` / `[-398, 369]` / `[-6176, 6111]`)
- `QNaN`/`SNaN` payload at or above the width canonical payload limit
  (`10^6` / `10^15` / `10^33`)
- a zero `Normal` coefficient, or a nonzero field that the selected `Kind`
  cannot encode: `payload` on `Normal`/`Zero`/`Infinity`, `coefficient` on
  `Zero`/`Infinity`/NaN, or `exponent` on `Infinity`/NaN

The `Components` field types make the remaining invalid inputs unconstructible,
so the type system represents the constraint there: `coefficient` is unsigned `u128`
(no negatives, nothing at or above `2^128`), `payload` is unsigned `u128` (no
negatives), and `kind` is a closed enum. In-range `Components` encode exactly as
before (bit-for-bit), so the canonical vector round-trip is unchanged. Encoding
never re-normalizes IEEE cohort members (it does not move digits between
coefficient and exponent to make a field fit).

## Payload

**Breaking change:** `Components.payload` is now a `u128` (previously `u64`). It
represents the full BID128 110-bit NaN payload, subject to the schema-wide value
limit of `10^33` (the widest canonical BID128 NaN payload, mirroring the
`10^34-1` coefficient cap). BID32/BID64 NaN payloads are unaffected subsets. A
BID128 NaN payload above `2^64` (up to `10^33-1`) is now preserved through
decode/encode instead of being truncated to its low 64 bits.

The unsigned `u128` type makes a negative payload unconstructible, so there is no
negative-payload check. A payload at or above `10^33` is rejected on encode and
normalized to 0 on decode (non-canonical), the same boundary in both directions.

## Verification

From the repository root:

```sh
make test-bidcodec
make verify-bidcodec-packages
```

This crate consumes `../bid754-codec-vectors/vectors.json` through a generated
test harness during repository verification. `make verify-bidcodec-packages`
additionally checks `cargo package --locked`, docs, lints, and an external
path-consumer smoke. The package gate intentionally runs without
`--allow-dirty` so dirty tracked crate source fails instead of being packaged
silently. If the crate is tested outside the
repository without that generated vector artifact, the repo-level vector tests
skip themselves rather than depending on repository-relative files.
