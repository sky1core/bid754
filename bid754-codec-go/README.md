# bidcodec for Go

`github.com/sky1core/bid754/bidcodec` is the Go BID codec helper package. It is not the
full decimal arithmetic implementation. Its scope is BID bit layout
encode/decode, little-endian byte encode/decode, and the shared BID codec string
format used by the cross-language vector suite.

## API

- `Decode32(uint32) Components`
- `Encode32(Components) uint32`
- `Decode64(uint64) Components`
- `Encode64(Components) uint64`
- `Decode128(lo, hi uint64) Components`
- `Encode128(Components) (lo, hi uint64)`
- `Decode32Bytes([]byte) (Components, error)`
- `Encode32Bytes(Components) [4]byte`
- `Decode64Bytes([]byte) (Components, error)`
- `Encode64Bytes(Components) [8]byte`
- `Decode128Bytes([]byte) (Components, error)`
- `Encode128Bytes(Components) [16]byte`
- `ToString(Components) string`
- `FromString(string) (Components, error)`

`Decode32Bytes`, `Decode64Bytes`, and `Decode128Bytes` require exactly 4, 8,
and 16 bytes respectively and return an error for any other length. `FromString`
returns an error for malformed payloads, malformed exponents, multiple decimal
points, empty input, and exponent values outside the signed 32-bit range. BID128
word order is `(lo, hi)`, and byte order is little-endian.

`Encode32`, `Encode64`, `Encode128`, and byte encode helpers are
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

This package consumes `../bid-codec-vectors/vectors.json` through a generated
test harness. `make audit-bidcodec-packages` additionally checks the standalone
package consumption boundary by creating an isolated local git release
repository tagged `bidcodec/v0.1.0` (Go multi-module subdirectory tag convention), then consuming `github.com/sky1core/bid754/bidcodec v0.1.0`
without a local `replace`. The generated vector harness is guarded by the
`bid754_bidcodec_vectors` build tag so ordinary package consumers do not depend
on repository-relative vector files.
