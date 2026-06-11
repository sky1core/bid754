# BidCodec for Swift

`BidCodec` is the Swift BID codec helper package. It is not a full decimal
arithmetic implementation. Its scope is BID bit layout encode/decode,
little-endian byte encode/decode, and the shared BID codec string format used by
the cross-language vector suite.

## API

- `BidCodec.decode32(_:) -> Components`
- `BidCodec.encode32(_:) -> UInt32`
- `BidCodec.decode64(_:) -> Components`
- `BidCodec.encode64(_:) -> UInt64`
- `BidCodec.decode128(lo:hi:) -> Components`
- `BidCodec.encode128(_:) -> (lo: UInt64, hi: UInt64)`
- `try BidCodec.decodeBytes32(_:) -> Components`
- `BidCodec.encodeBytes32(_:) -> Data`
- `try BidCodec.decodeBytes64(_:) -> Components`
- `BidCodec.encodeBytes64(_:) -> Data`
- `try BidCodec.decodeBytes128(_:) -> Components`
- `BidCodec.encodeBytes128(_:) -> Data`
- `BidCodec.toString(_:) -> String`
- `try BidCodec.fromString(_:) -> Components`

`decodeBytes32`, `decodeBytes64`, and `decodeBytes128` require exactly 4, 8, and
16 bytes respectively and throw for any other length. `fromString` throws for
malformed payloads, malformed exponents, multiple decimal points, empty input,
and exponent values outside the signed 32-bit range. BID128 word order is
`(lo, hi)`, and byte order is little-endian.

`encode32`, `encode64`, `encode128`, and byte encode helpers are
trusted-component packing APIs, not validation APIs. They canonicalize into the
target BID bit layout and may clamp exponent fields or mask/truncate
coefficient and payload fields. Invalid `Components` rejection is not part of
the current API contract.

`Foundation.Decimal` conversion helpers are internal adapters, not part of the
standalone BID codec public API. They must stay internal unless their semantics
are specified here and covered by generated vectors.

## Verification

From the repository root:

```sh
make test-bidcodec
make audit-bidcodec-packages
```

This package consumes `../bid754-codec-vectors/vectors.json` through a generated
test harness. `make audit-bidcodec-packages` additionally checks release build
and an external Swift package consumer smoke.
