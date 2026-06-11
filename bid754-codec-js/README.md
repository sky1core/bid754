# @bid754/bid-codec

`@bid754/bid-codec` is the JavaScript/TypeScript BID codec helper package. It is
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
16 bytes respectively and throw for any other length. `fromString` throws for
malformed payloads, malformed exponents, multiple decimal points, empty input,
and exponent values outside the signed 32-bit range. BID128 word order is
`(lo, hi)`, and byte order is little-endian.

`encode32`, `encode64`, `encode128`, and byte encode helpers are
trusted-component packing APIs, not validation APIs. They canonicalize into the
target BID bit layout and may clamp exponent fields or mask/truncate
coefficient and payload fields. Invalid `Components` rejection is not part of
the current API contract.

## Verification

From the repository root:

```sh
make test-bidcodec
make audit-bidcodec-packages
```

This package consumes `../bid754-codec-vectors/vectors.json` through a generated
test harness. `make audit-bidcodec-packages` additionally checks package build,
npm pack, install, and import smoke.
