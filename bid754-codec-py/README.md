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
and 16 bytes respectively and raise `ValueError` for any other length. BID128
word order is `(lo, hi)`, and byte order is little-endian.

The raw-word decode APIs accept exact unsigned bit containers only. `decode32`
requires an `int` in `0..2^32-1`; `decode64` and each `lo`/`hi` argument of
`decode128` require an `int` in `0..2^64-1`. A `bool`, non-`int`, negative value,
or value above the word maximum raises `ValueError` before any bit operation, so
no input is truncated, masked, or reinterpreted as another encoding.

`from_string` accepts one strict ASCII grammar and raises `ValueError` for
anything outside it. The whole input must be ASCII: any non-ASCII character
anywhere (including Unicode digit variants and Unicode whitespace) is malformed.
Only ASCII whitespace may surround the token, and there is no whitespace inside
it. Also rejected: a digit-group underscore, a sign on a NaN payload, a
malformed payload or exponent, multiple decimal points, empty input, an
exponent literal at or above the shared exact-integer bound 2^53 in magnitude
(the widest bound every consumer's number type can check exactly), a
fraction-adjusted final exponent outside the signed 32-bit range (the literal
itself may exceed int32 — `to_string` renders adjusted exponents up to
int32 max + 33, far below 2^53, and every `to_string` output reparses through
`from_string`,
the round-trip closure), a parsed coefficient value above the schema-wide maximum `10^34-1`, and a
parsed NaN payload value at or above the schema-wide maximum `10^33` — both are
the largest values any supported BID width can hold, schema constants shared by
all six language codecs, not per-width validation. Per-BID-width range checking
is the encode contract below, not the parser's.

## String render reject contract

`to_string` validates the width-independent shared `Components` schema and
raises `ValueError` before rendering an invalid value. A `normal` coefficient
must be in `1..10^34-1`, a NaN payload in `0..10^33-1`, the sign must be a
`bool`, the exponent must be an int32 value, and every field the selected kind
cannot represent must be zero. Thus no field is silently discarded and every
successful rendering reparses without kind or component loss.

## NaN payload

The `Components.payload` field is the full BID128 110-bit NaN payload: a
non-negative `int` whose value must be below `10^33` (the widest canonical BID128
NaN payload, mirroring the coefficient's `10^34-1` schema cap). BID32 and BID64
payloads are subsets of this same field. `decode128` preserves the entire 110-bit
payload — including payloads above `2^64` — and a payload at or above `10^33` is
non-canonical decode-only input whose decode normalizes it to `0`.

The `payload` field type is unchanged (`int`), but its behavior changed (a SOFT
breaking change): `decode128` now returns the full 110-bit payload instead of only
the low 64 bits, `encode128` now accepts payloads above `2^64` (up to `10^33-1`)
and rejects payloads at or above `10^33` (previously the field was bounded at
`2^64-1`), and `from_string` accepts NaN payloads up to `10^33-1`. Although
`bool` is an `int` subtype in Python, coefficient, exponent, and payload reject
boolean values; only `sign` accepts `bool`.

`encode32`, `encode64`, `encode128`, and the byte encode helpers are validating
packing APIs. A `Components` value whose fields are not representable in the
target BID width is rejected with `ValueError` — never silently truncated,
masked, or clamped, and never crashed through an overflow. The field-level
reject boundaries per width are: a `normal` coefficient above `10^7-1` /
`10^16-1` / `10^34-1`; a `zero` or `normal` unbiased exponent outside
`[-101,+90]` / `[-398,+369]` / `[-6176,+6111]`; a `qnan`/`snan` payload at or
above `10^6` / `10^15` / `10^33`; and a `coefficient`, `payload`, or `exponent`
that is not an `int`, or a negative `coefficient` or `payload`, or an
unrecognized `kind`. A zero `normal` coefficient and any nonzero field that the
selected kind cannot encode are also rejected: `payload` on
`normal`/`zero`/`infinity`, `coefficient` on `zero`/`infinity`/NaN, and
`exponent` on `infinity`/NaN. The message names the field, width, value, and limit.
In-range `Components` still encode to identical bits, so the canonical vector
round-trip is unchanged.

## Verification

From the repository root:

```sh
make test-bidcodec
make verify-bidcodec-packages
```

This package consumes `../bid754-codec-vectors/vectors.json` through a generated
test harness. `make verify-bidcodec-packages` additionally checks wheel build,
typed marker inclusion, install, and import smoke.
