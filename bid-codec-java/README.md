# bid-codec for Java

`dev.bid754.bidcodec` is the Java BID codec helper package. It is not a full
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
`fromString` throws `IllegalArgumentException` for malformed payloads, malformed
exponents, multiple decimal points, empty input, and exponent values outside the
signed 32-bit range. BID128 word order is `(lo, hi)`, and byte order is
little-endian.

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

This package consumes `../bid-codec-vectors/vectors.json` through a generated
test harness. `make audit-bidcodec-packages` additionally checks the Gradle
package build, Maven publication metadata, and an external jar consumer smoke.
The standalone Maven coordinate is `dev.bid754.bidcodec:bid-codec-java:0.1.0`.
Gradle dependency resolution is locked by the checked-in `gradle.lockfile`;
refresh it only when the Java package dependencies intentionally change.
