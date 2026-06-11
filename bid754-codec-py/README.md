# bid754-codec for Python

`bid754-codec` is the Python BID codec helper package on PyPI; its import
module is `bid_codec`. It is not a full decimal
arithmetic implementation. Its scope is BID bit layout encode/decode,
little-endian byte encode/decode, and the shared BID codec string format used by
the cross-language vector suite.

## API

- `decode32(int) -> Components`
- `encode32(Components) -> int`
- `decode64(int) -> Components`
- `encode64(Components) -> int`
- `decode128(lo: int, hi: int) -> Components`
- `encode128(Components) -> tuple[int, int]`
- `decode_bytes(bytes) -> Components`
- `encode_bytes(Components, size: int) -> bytes`
- `decode_bytes32(bytes) -> Components`
- `encode_bytes32(Components) -> bytes`
- `decode_bytes64(bytes) -> Components`
- `encode_bytes64(Components) -> bytes`
- `decode_bytes128(bytes) -> Components`
- `encode_bytes128(Components) -> bytes`
- `to_string(Components) -> str`
- `from_string(str) -> Components`

`decode_bytes32`, `decode_bytes64`, and `decode_bytes128` require exactly 4, 8,
and 16 bytes respectively and raise `ValueError` for any other length.
`from_string` raises `ValueError` for malformed payloads, malformed exponents,
multiple decimal points, empty input, and exponent values outside the signed
32-bit range. BID128 word order is `(lo, hi)`, and byte order is little-endian.

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
test harness. `make audit-bidcodec-packages` additionally checks wheel build,
typed marker inclusion, install, and import smoke.
