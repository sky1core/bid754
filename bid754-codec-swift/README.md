# BidCodec for Swift

`BidCodec` is the Swift BID codec helper package. It is not a full decimal
arithmetic implementation. Its scope is BID bit layout encode/decode,
little-endian byte encode/decode, and the shared BID codec string format used by
the cross-language vector suite.

## API

- `BidCodec.decode32(_:) -> Components`
- `try BidCodec.encode32(_:) -> UInt32`
- `BidCodec.decode64(_:) -> Components`
- `try BidCodec.encode64(_:) -> UInt64`
- `BidCodec.decode128(lo:hi:) -> Components`
- `try BidCodec.encode128(_:) -> (lo: UInt64, hi: UInt64)`
- `try BidCodec.decodeBytes32(_:) -> Components`
- `try BidCodec.encodeBytes32(_:) -> Data`
- `try BidCodec.decodeBytes64(_:) -> Components`
- `try BidCodec.encodeBytes64(_:) -> Data`
- `try BidCodec.decodeBytes128(_:) -> Components`
- `try BidCodec.encodeBytes128(_:) -> Data`
- `try BidCodec.toString(_:) -> String`
- `try BidCodec.fromString(_:) -> Components`

`decodeBytes32`, `decodeBytes64`, and `decodeBytes128` require exactly 4, 8, and
16 bytes respectively and throw for any other length. BID128 word order is
`(lo, hi)`, and byte order is little-endian.

`fromString` accepts one strict ASCII grammar. The whole input must be ASCII:
Unicode digit variants (e.g. `１`, `١`), Unicode whitespace (e.g. `U+00A0`), and
fractions (e.g. `½`) are rejected anywhere in the input. Only ASCII whitespace
may surround the token; a digit-group underscore, a sign inside a NaN payload,
and whitespace inside the token are all malformed. It throws for empty input,
malformed payloads, malformed exponents, multiple decimal points, an exponent
literal at or above the shared exact-integer bound 2^53 in magnitude (the
widest bound every consumer's number type can check exactly; rejected through
the same error channel), and a fraction-adjusted final
exponent outside the signed 32-bit range. The literal itself may exceed
int32: `toString` renders adjusted exponents up to `Int32.max + 33` (far
below 2^53), and
every `toString` output reparses through `fromString` (round-trip closure). The parsed coefficient value must not exceed the schema-wide
maximum coefficient `10^34-1` — the largest value any supported BID width can
hold, a schema constant rather than per-width validation; BID width-range
checks belong to `encode*`. A NaN payload must be below the schema-wide NaN
payload limit `10^33` (the widest canonical BID128 NaN payload), the same kind
of schema constant.

## String render reject contract

**Breaking change:** `toString` is now `throws`. It accepts only the
width-independent shared `Components` schema: a `normal` coefficient is in
`1..10^34-1`; a NaN payload is in `0..10^33-1`; and every field the selected
kind cannot represent is zero. Invalid values throw before rendering, so no
field is silently discarded and every successful rendering reparses without
kind or component loss.

`encode32`, `encode64`, `encode128`, and the byte encode helpers are validating
packing APIs and are `throws` (a breaking change from the previous
non-throwing, validated-component packing signatures; call sites must now use
`try`). `Components.payload` (`UInt64`) has also been split into the
`payloadHi`/`payloadLo` `UInt64` pair, mirroring `coefficientHi`/`coefficientLo`,
so it can hold the full BID128 110-bit NaN payload — a HARD breaking change
(semver major): code that constructed or read `.payload` must migrate to the two
words (BID32/BID64 use `payloadLo` only).

A `Components` value whose fields are not representable in the target
BID width is rejected with `BidCodecError.invalidComponents`, whose message
names the field, width, value, and limit. No field is silently truncated,
masked, clamped, or coerced, and no input traps the process. The reject
boundaries are field-level, per width:

- a `normal` coefficient above the width maximum (`10^7-1` / `10^16-1` /
  `10^34-1` for BID32 / BID64 / BID128)
- a `zero`/`normal` unbiased exponent outside the width range
  (`[-101,+90]` / `[-398,+369]` / `[-6176,+6111]`)
- a `qnan`/`snan` payload at or above the width canonical limit (`10^6` / `10^15`
  / `10^33` for BID32 / BID64 / BID128). The payload is the `(payloadHi,
  payloadLo)` `UInt64` pair holding the full BID128 110-bit NaN payload; a payload
  at or above `10^33` (the widest canonical BID128 NaN payload) is non-canonical
  and rejected
- a zero `normal` coefficient, or a nonzero field that the selected kind cannot
  encode: payload on `normal`/`zero`/`infinity`, coefficient on
  `zero`/`infinity`/NaN, or exponent on `infinity`/NaN

`Components` fields typed as `UInt64`/`Int32`/enum are already domain-constrained
by the type system, so no extra domain check is needed; a BID32/BID64 coefficient
or payload with a non-zero `coefficientHi`/`payloadHi` is out of range and is
rejected. In-range `Components` encode to exactly the same bits as before.

`Foundation.Decimal` conversion helpers are internal adapters, not part of the
standalone BID codec public API. They must stay internal unless their semantics
are specified here and covered by generated vectors.

## Verification

From the repository root:

```sh
make test-bidcodec
make verify-bidcodec-packages
```

This package consumes `../bid754-codec-vectors/vectors.json` through a generated
test harness. `make verify-bidcodec-packages` additionally checks release build
and an external Swift package consumer smoke.
