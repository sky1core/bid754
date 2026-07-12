# bid754-codec for Java

`io.github.sky1core.bidcodec` is the Java BID codec helper package. It is not a full
decimal arithmetic implementation. Its scope is BID bit layout encode/decode,
little-endian byte encode/decode, and the shared BID codec string format used by
the cross-language vector suite.

## API

- `BidCodec.decode32(int) -> Components`
- `BidCodec.encode32(Components) -> int`
- `BidCodec.decode64(long) -> Components`
- `BidCodec.encode64(Components) -> long`
- `BidCodec.decode128(long lo, long hi) -> Components`
- `BidCodec.encode128(Components) -> long[]`
- `BidCodec.decodeBytes32(byte[]) -> Components`
- `BidCodec.encodeBytes32(Components) -> byte[]`
- `BidCodec.decodeBytes64(byte[]) -> Components`
- `BidCodec.encodeBytes64(Components) -> byte[]`
- `BidCodec.decodeBytes128(byte[]) -> Components`
- `BidCodec.encodeBytes128(Components) -> byte[]`
- `BidCodec.toString(Components) -> String`
- `BidCodec.fromString(String) -> Components`

`decodeBytes32`, `decodeBytes64`, and `decodeBytes128` require exactly 4, 8, and
16 bytes respectively and throw `IllegalArgumentException` for any other length.
BID128 word order is `(lo, hi)`, and byte order is little-endian.

`fromString` accepts one strict ASCII grammar and throws
`IllegalArgumentException` for anything outside it. The whole input must be
ASCII: any non-ASCII character anywhere (including Unicode digit variants and
Unicode whitespace) is malformed. Only ASCII whitespace may surround the token,
and there is no whitespace inside it. Also rejected: a digit-group underscore, a
sign on a NaN payload, a malformed payload or exponent, multiple decimal points,
empty input, an exponent literal at or above the shared exact-integer bound
2^53 in magnitude (the widest bound every consumer's number type can check
exactly), a
fraction-adjusted final exponent outside the signed 32-bit range (the literal
itself may exceed int32 — `toString` renders adjusted exponents up to
`Integer.MAX_VALUE + 33`, far below 2^53, and every `toString` output reparses
through
`fromString`, the round-trip closure), a parsed coefficient value above the schema-wide maximum
`10^34-1`, and a parsed NaN payload value at or above the schema-wide maximum
`10^33` — both are the largest values any supported BID width can hold, schema
constants shared by all six language codecs, not per-width validation.
Per-BID-width range checking is the encode contract below, not the parser's.

## String render reject contract

`toString` validates the width-independent shared `Components` schema and
throws `IllegalArgumentException` before rendering an invalid value. A
`NORMAL` coefficient must be in `1..10^34-1`, a NaN payload in `0..10^33-1`,
and every field the selected `DecimalKind` cannot represent must be zero (or
the documented null-zero form). Thus no field is silently discarded and every
successful rendering reparses without kind or component loss.

## NaN payload

The `Components.payload` field is the full BID128 110-bit NaN payload: an unsigned
`BigInteger` whose value must be below `10^33` (the widest canonical BID128 NaN
payload, mirroring the coefficient's `10^34-1` schema cap). BID32 and BID64
payloads are subsets of this same field. `decode128` preserves the entire 110-bit
payload — including payloads above `2^64` — and a payload at or above `10^33` is
non-canonical decode-only input whose decode normalizes it to `0`. A `null`
payload is the explicitly documented default zero payload (an absent payload
encodes as `+NaN`/`+SNaN`), not a rejected value; it is a documented field
default, not an implicit fallback.

**Breaking change (semver major).** The `Components.payload` component type is now
`BigInteger` (previously `long`). This is a HARD field-type change: code that
constructs, reads, or pattern-matches `payload` as a `long` must migrate to
`BigInteger`. The former behavior where a negative `long` was reinterpreted as the
high-bit pattern of a 64-bit BID128 payload is gone; payloads are now plain
unsigned integers, a negative `BigInteger` payload is rejected on all three
widths, and BID128 payloads above `2^64` (up to `10^33-1`) are expressed directly.

`encode32`, `encode64`, `encode128`, and the byte encode helpers are validating
packing APIs. A `Components` value whose fields are not representable in the
target BID width is rejected with `IllegalArgumentException` — never silently
truncated, masked, or clamped, and never crashed through a trap. The field-level
reject boundaries per width are: a `normal` coefficient above `10^7-1` /
`10^16-1` / `10^34-1`; a `zero` or `normal` unbiased exponent outside
`[-101,+90]` / `[-398,+369]` / `[-6176,+6111]`; a `qnan`/`snan` payload at or
above `10^6` / `10^15` / `10^33`; a zero/negative/null `normal` coefficient;
a nonzero field that the selected kind cannot encode (`payload` on
`normal`/`zero`/`infinity`, `coefficient` on `zero`/`infinity`/NaN, or
`exponent` on `infinity`/NaN); and a negative `payload` on any width. The
message names the field, width, value, and limit.
In-range `Components` still encode to identical bits, so the canonical vector
round-trip is unchanged.

## Verification

From the repository root:

```sh
make test-bidcodec
make verify-bidcodec-packages
```

This package consumes `../bid754-codec-vectors/vectors.json` through a generated
test harness. `make verify-bidcodec-packages` additionally checks the Gradle
package build, Maven publication metadata, and an external jar consumer smoke.
The standalone Maven coordinate is `io.github.sky1core:bid754-codec:0.2.0`.
Gradle dependency resolution is locked by the checked-in `gradle.lockfile`;
refresh it only when the Java package dependencies intentionally change.
