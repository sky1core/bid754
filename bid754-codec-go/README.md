# bidcodec for Go

`github.com/sky1core/bid754/bid754-codec-go` is the Go BID codec helper package. It is not the
full decimal arithmetic implementation. Its scope is BID bit layout
encode/decode, little-endian byte encode/decode, and the shared BID codec string
format used by the cross-language vector suite.

## API

- `Decode32(uint32) Components`
- `Encode32(Components) (uint32, error)`
- `Decode64(uint64) Components`
- `Encode64(Components) (uint64, error)`
- `Decode128(lo, hi uint64) Components`
- `Encode128(Components) (lo, hi uint64, err error)`
- `Decode32Bytes([]byte) (Components, error)`
- `Encode32Bytes(Components) ([4]byte, error)`
- `Decode64Bytes([]byte) (Components, error)`
- `Encode64Bytes(Components) ([8]byte, error)`
- `Decode128Bytes([]byte) (Components, error)`
- `Encode128Bytes(Components) ([16]byte, error)`
- `ToString(Components) (string, error)`
- `FromString(string) (Components, error)`

`Decode32Bytes`, `Decode64Bytes`, and `Decode128Bytes` require exactly 4, 8,
and 16 bytes respectively and return an error for any other length. BID128
word order is `(lo, hi)`, and byte order is little-endian.

`FromString` accepts one strict ASCII grammar. The whole input must be ASCII:
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
fit a signed 32-bit integer. The literal itself may exceed int32: `ToString`
renders adjusted exponents up to int32 max + 33 (far below 2^53), and every
`ToString` output reparses through `FromString` (round-trip closure).
Underscores and payload-internal signs are malformed. `FromString` validates
grammar and schema limits only, identical in all six language packages: the
parsed coefficient value must not exceed the schema-wide maximum coefficient
`10^34-1` (the largest value any supported BID width can hold — a schema
constant, not per-width validation), and the parsed payload value must be below
the schema-wide NaN payload limit `10^33` (the widest canonical BID128 NaN
payload, the same kind of schema constant). BID width-range validation stays in
the `Encode*` contract below.

## String render reject contract

**Breaking change:** `ToString` now returns an `error`. It accepts only the
width-independent shared `Components` schema: a `Normal` coefficient is in
`1..10^34-1`; a NaN payload is in `0..10^33-1`; and every field the selected
`Kind` cannot represent is zero (or the documented nil-zero form). Invalid
values return an error before rendering, so no kind, coefficient, exponent, or
payload is silently discarded and every successful rendering reparses without
component loss.

## Encode reject contract

**Breaking change:** `Encode32`/`Encode64`/`Encode128` and the byte encode
helpers now return an `error` (previously they returned only the packed value).
They are validating packing APIs: no input reachable through the public API
fails silently. Instead of truncating, masking, clamping, or panicking, they
return an `error` for any `Components` field not representable in the target BID
width exactly as supplied. The reject boundaries are field-level, per width:

- `Normal` coefficient above the width maximum
  (`9999999` / `9999999999999999` / `10^34-1`)
- `Zero`/`Normal` unbiased exponent outside the width range
  (`[-101, 90]` / `[-398, 369]` / `[-6176, 6111]`)
- `QNaN`/`SNaN` payload at or above the width canonical payload limit
  (`10^6` / `10^15` / `10^33`)
- a zero, negative, or `nil` `Coefficient` on a `Normal` value
- a nonzero field that the selected `Kind` cannot encode: `Payload` on
  `Normal`/`Zero`/`Infinity`, `Coefficient` on `Zero`/`Infinity`/NaN, or
  `Exponent` on `Infinity`/NaN
- a negative `Payload` on a NaN or an unrecognized `Kind`

In-range `Components` encode exactly as before (bit-for-bit), so the canonical
vector round-trip is unchanged. Encoding never re-normalizes IEEE cohort members
(it does not move digits between coefficient and exponent to make a field fit).

## Payload

**Breaking change:** `Components.Payload` is now a `*big.Int` (previously
`uint64`). It represents the full BID128 110-bit NaN payload, subject to the
schema-wide value limit of `10^33` (the widest canonical BID128 NaN payload,
mirroring the `10^34-1` coefficient cap). BID32/BID64 NaN payloads are
unaffected subsets. A BID128 NaN payload above `2^64` (up to `10^33-1`) is now
preserved through decode/encode instead of being truncated to its low 64 bits.

A nil `Payload` on a `QNaN`/`SNaN` is the documented default zero payload (it
encodes as payload 0); this is an explicit field default, not an implicit
fallback. `Decode*` returns a non-nil `Payload` (zero for a bare or
non-canonical NaN). A negative `Payload` — now constructible through `*big.Int`
— is rejected by `Encode*`. A payload at or above `10^33` is rejected on encode
and normalized to 0 on decode (non-canonical), the same boundary in both
directions.

## Verification

From the repository root:

```sh
make test-bidcodec
make verify-bidcodec-packages
```

This package consumes `../bid754-codec-vectors/vectors.json` through a generated
test harness. `make verify-bidcodec-packages` additionally checks the standalone
package consumption boundary by creating an isolated local git release
repository tagged `bid754-codec-go/v<version>` (Go multi-module subdirectory tag
convention), then consuming
`github.com/sky1core/bid754/bid754-codec-go v<version>` without a local
`replace`. That `<version>` is not hard-coded in the gate: it is read from the
`Version` constant in `bid754-go/bid754.go`, the single scripted project version
source that `make verify-package-versions` pins every package manifest to, so
the synthetic release always matches the current tree. The generated vector
harness is enabled by the
`bid754_bidcodec_vectors` build tag so ordinary package consumers do not depend
on repository-relative vector files.
