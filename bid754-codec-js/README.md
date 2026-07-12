# bid754-codec

`bid754-codec` is the JavaScript/TypeScript BID codec helper package. It is
not a full decimal arithmetic implementation. Its scope is BID bit layout
encode/decode, little-endian byte encode/decode, and the shared BID codec string
format used by the cross-language vector suite.

## API

- `decode32(number): Components`
- `encode32(Components): number`
- `decode64(bigint): Components`
- `encode64(Components): bigint`
- `decode128(lo: bigint, hi: bigint): Components`
- `encode128(Components): [bigint, bigint]`
- `decodeBytes32(Uint8Array): Components`
- `encodeBytes32(Components): Uint8Array`
- `decodeBytes64(Uint8Array): Components`
- `encodeBytes64(Components): Uint8Array`
- `decodeBytes128(Uint8Array): Components`
- `encodeBytes128(Components): Uint8Array`
- `toString(Components): string`
- `fromString(string): Components`

`decodeBytes32`, `decodeBytes64`, and `decodeBytes128` require exactly 4, 8, and
16 bytes respectively and throw for any other length. BID128 word order is
`(lo, hi)`, and byte order is little-endian.

The raw-word decode APIs validate runtime inputs because JavaScript callers do
not use TypeScript declarations at runtime. `decode32` requires an integer `number` in
`0..2^32-1`; `decode64` and each `lo`/`hi` argument of `decode128` require a
`bigint` in `0..2^64-1`. Wrong runtime types, negative values, non-finite or
fractional BID32 numbers, and values above the word maximum throw before any bit
operation, so no input is truncated, masked, or coerced into another encoding.

`fromString` accepts one strict ASCII grammar. The whole input must be ASCII:
Unicode digit variants (e.g. `１`, `١`), Unicode whitespace (e.g. `U+00A0`), and
fractions (e.g. `½`) are rejected anywhere in the input. Only ASCII whitespace
may surround the token; a digit-group underscore, a sign inside a NaN payload,
and whitespace inside the token are all malformed. It throws for empty input,
malformed payloads, malformed exponents, multiple decimal points, an exponent
literal at or above the shared exact-integer bound 2^53 in magnitude (the
`fromString` grammar constant every language codec enforces — the widest
bound every consumer's number type can check exactly; `Number.isSafeInteger`
pins it here, and the other codecs enforce the same constant explicitly, so
every codec decides each input its runtime can represent by the same
mathematical rule), and a
fraction-adjusted final exponent outside
the signed 32-bit range. The literal itself may exceed int32: `toString`
renders adjusted exponents up to int32 max + 33 (far below 2^53), and every
`toString` output
reparses through `fromString` (round-trip closure). The parsed coefficient value must not exceed the schema-wide
maximum coefficient `10^34-1` — the largest value any supported BID width can
hold, a schema constant rather than per-width validation; BID width-range
checks belong to `encode*`. A NaN payload must be below the schema-wide NaN
payload limit `10^33` (the widest canonical BID128 NaN payload), the same kind
of schema constant.

## String render reject contract

`toString` validates the width-independent shared `Components` schema and
throws before rendering an invalid value. A `Normal` coefficient must be in
`1..10^34-1`, a NaN payload in `0..10^33-1`, `sign` must be a boolean,
`exponent` must be an int32 integer, and every field the selected `Kind` cannot
represent must be zero. Thus no field is silently discarded and every
successful rendering reparses without kind or component loss. The same boolean
sign requirement applies to `encode*`.

`encode32`, `encode64`, `encode128`, and the byte encode helpers are validating
packing APIs. A `Components` value whose fields are not representable in the
target BID width is rejected by throwing an `Error` whose message names the
field, width, value, and limit. No field is silently truncated, masked,
clamped, or coerced, and no input crashes the process. The reject boundaries are
field-level, per width:

- a `Normal` coefficient above the width maximum (`10^7-1` / `10^16-1` /
  `10^34-1` for BID32 / BID64 / BID128)
- a `Zero`/`Normal` unbiased exponent outside the width range
  (`[-101,+90]` / `[-398,+369]` / `[-6176,+6111]`)
- a `QNaN`/`SNaN` payload at or above the width canonical limit (`10^6` / `10^15`
  / `10^33` for BID32 / BID64 / BID128). The `bigint` `payload` holds the full
  BID128 110-bit NaN payload; a payload at or above `10^33` (the widest canonical
  BID128 NaN payload) is non-canonical and rejected
- a zero `Normal` coefficient, or a nonzero field that the selected `Kind`
  cannot encode: `payload` on `Normal`/`Zero`/`Infinity`, `coefficient` on
  `Zero`/`Infinity`/NaN, or `exponent` on `Infinity`/NaN
- a field outside its declared domain: a non-`bigint` or negative `coefficient`
  or `payload`, a non-integer `exponent`, or an unrecognized `kind`

In-range `Components` encode to exactly the same bits as before. The `payload`
field carries the full BID128 110-bit NaN payload: `decode128` preserves payload
bits above `2^64`, and `encode128`/`fromString` accept any value below `10^33`.
BID32/BID64 payloads are the low-value subset of this field.

## Verification

From the repository root:

```sh
make test-bidcodec
make verify-bidcodec-packages
```

This package consumes `../bid754-codec-vectors/vectors.json` through a generated
test harness. `make verify-bidcodec-packages` additionally checks package build,
npm pack, install, and import smoke.
