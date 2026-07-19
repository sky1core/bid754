# Standalone BID Codec Specification

This document defines the normative cross-language contract for the standalone
BID codec packages and their shared generated vector protocol. It applies to:

- `bid754-codec-go`
- `bid754-codec-rs`
- `bid754-codec-java`
- `bid754-codec-py`
- `bid754-codec-js`
- `bid754-codec-swift`

The codec packages decode, encode, parse, and render BID interchange values.
They are not full decimal-arithmetic implementations. Language-specific API
spellings are documented in each package README; the mathematical and error
semantics below are shared.

## Components Model

A decoded or parsed value is represented by these logical fields:

| Field | Contract |
| --- | --- |
| `sign` | Boolean sign bit |
| `kind` | `normal`, `zero`, `inf`, `qnan`, or `snan` |
| `coefficient` | Unsigned decimal integer |
| `exponent` | Signed 32-bit integer |
| `payload` | Unsigned decimal NaN payload |

The selected kind determines which fields carry information:

| Kind | Coefficient | Exponent | Payload |
| --- | --- | --- | --- |
| `normal` | `1..10^34-1` | signed 32-bit | zero |
| `zero` | zero | signed 32-bit | zero |
| `inf` | zero | zero | zero |
| `qnan`, `snan` | zero | zero | `0..10^33-1` |

Numeric zero is the declared value for an unused field. A language with a
nullable coefficient or payload may also use an absent value where the selected
kind permits numeric zero. This is an explicit field default; it does not permit
an implementation to discard nonzero data.

## BID Width Limits

`Encode*` applies the selected target width in addition to the shared
`Components` schema:

| Width | Bytes | Normal coefficient | Unbiased exponent | Canonical NaN payload |
| --- | ---: | --- | --- | --- |
| BID32 | 4 | `1..10^7-1` | `[-101,+90]` | `0..10^6-1` |
| BID64 | 8 | `1..10^16-1` | `[-398,+369]` | `0..10^15-1` |
| BID128 | 16 | `1..10^34-1` | `[-6176,+6111]` | `0..10^33-1` |

The shared coefficient and payload fields are deliberately wide enough for
BID128. A canonical BID128 NaN payload above `2^64` must survive decode, encode,
and string conversion without narrowing.

## Raw-Word Decode

Raw-word decode accepts an exact fixed-width bit container. It does not accept
an arbitrary integer and narrow it by masking.

- BID32 consumes one 32-bit word.
- BID64 consumes one 64-bit word.
- BID128 consumes low and high 64-bit words in that order.

Statically typed word parameters enforce this at the type boundary. Dynamic
language entrypoints must reject wrong runtime types, negative values,
non-integral or non-finite values where applicable, and values above the word
width before decoding. Python integers exclude `bool`; JavaScript uses `number`
for BID32 and `bigint` for every 64-bit word.

Decode is defined for every raw bit pattern and returns normalized
`Components`; canonicality is not an extra public component field. Generated
raw records classify the original encoding and require encode-after-decode to
reproduce it only when `canonical` is true. Non-canonical encodings follow the
pinned BID layout reference behavior. In particular, an out-of-range NaN
payload decodes with payload zero, and an out-of-range finite coefficient
decodes as zero while preserving the decoded sign and exponent.

## Exact Byte APIs

Byte encoding is little-endian and has exact length:

- BID32: 4 bytes
- BID64: 8 bytes
- BID128: 16 bytes

A dynamic byte buffer of any other length fails through the language-specific
error mechanism. It must not be truncated, padded, indexed out of bounds, or
allowed to panic/trap. Fixed-array APIs may enforce length through their type,
but a dynamic-slice companion used by generated verification must expose the
same failure semantics.

## Validating Encode

`Encode32`, `Encode64`, and `Encode128` are validating packing APIs. They encode
an in-range `Components` value exactly as supplied and reject any value that the
target BID width or selected kind cannot represent.

The following inputs are invalid:

- a `normal` coefficient outside the target width's range, including zero;
- a `zero` or `normal` exponent outside the target width's range;
- a NaN payload outside the target width's canonical payload range;
- a nonzero field unused by the selected kind;
- a negative coefficient or payload where the language can construct one;
- a non-integral, non-finite, or out-of-int32 exponent where the language can
  construct one;
- an unrecognized kind or non-Boolean sign where the language can construct
  one; and
- any runtime field type outside the language's declared public schema.

Encode never moves digits between coefficient and exponent, changes the kind,
clamps a field, masks high bits, or otherwise normalizes an invalid value into a
representable one.

## String Parsing

`fromString` accepts one shared strict ASCII grammar.

### Common Rules

- The whole input must be ASCII.
- One surrounding trim of ASCII TAB, LF, VT, FF, CR, and SPACE
  (`0x09..0x0d`, `0x20`) is allowed; internal whitespace is rejected.
- One leading `+` or `-` is allowed.
- Unicode digits, Unicode whitespace, digit separators, and underscores are
  rejected.
- Standard-library parsers may be used only after the shared lexical rules have
  been enforced; their language-specific extensions do not widen this grammar.

### Special Values

After the optional sign, special values are matched with ASCII
case-insensitivity:

- `Inf` or `Infinity`;
- `NaN` followed by an optional unsigned ASCII payload; or
- `SNaN` followed by an optional unsigned ASCII payload.

The payload must be below `10^33`. An internal sign, non-digit, or oversized
payload is malformed.

### Finite Values

A finite value contains ASCII digits with at most one decimal point and at
least one digit. It may be followed by `E` or `e`, one optional exponent sign,
and at least one ASCII exponent digit.

Remove the decimal point to form the coefficient digits. If `f` digits followed
the point and the explicit exponent is `q` (zero when absent), the stored
exponent is `q - f`. The coefficient is the unsigned base-10 value of all
coefficient digits; if it is zero the result kind is `zero`, otherwise it is
`normal`. Leading zeros do not change the integer coefficient, while trailing
zeros remain part of the coefficient and therefore preserve the parsed cohort.

- The exponent literal magnitude must be strictly below `2^53`.
- After subtracting fractional digits, the final stored exponent must fit a
  signed 32-bit integer.
- The parsed coefficient must not exceed `10^34-1`.
- Parsing applies the shared schema only; target-width validation belongs to
  `Encode*`.

The `2^53` literal bound is shared even in languages with wider integer types so
that every consumer accepts and rejects the same mathematical inputs. Checked
digit accumulation is required; overflow, saturation, wrapping, and panic/trap
are invalid parser behavior.

## String Rendering

`toString` validates the shared `Components` schema before rendering. It must
not ignore an invalid field or emit a string that reparses to a different kind
or component value.

The output is ASCII and always carries an explicit leading sign:

- infinity is `+Inf` or `-Inf`;
- quiet and signaling NaNs are `+NaN`, `-NaN`, `+SNaN`, or `-SNaN`, with a
  nonzero payload appended as unsigned decimal digits and a zero payload
  omitted;
- zero with exponent zero is `+0` or `-0`; zero with a nonzero exponent is
  rendered as signed zero followed by `E` and the explicitly signed decimal
  exponent; and
- a normal coefficient with decimal digits `d` and stored exponent `e` is
  normalized to one digit before the decimal point and an explicitly signed
  adjusted exponent `e + len(d) - 1`. A one-digit coefficient has no decimal
  point; every remaining coefficient digit, including trailing zeros, is
  preserved after the point.

For example, coefficient `1100` with exponent `-3` renders as `+1.100E+0`,
not as a numerically equivalent cohort member.

For every valid value:

1. `toString(value)` succeeds;
2. `fromString(toString(value))` succeeds with the same logical components; and
3. rendering that reparsed value produces the same bytes again.

Adjusted-exponent calculations must not overflow signed 32-bit arithmetic.
Valid components can render an exponent above int32 because the displayed
scientific exponent includes the coefficient digit count; the parser's wider
literal bound exists so every valid rendering remains parseable.

## Error Semantics

Unsupported, malformed, or unrepresentable input fails through the
language-specific error mechanism:

- Go returns `error`;
- Rust returns `Result`;
- Swift throws; and
- Java, Python, and JavaScript/TypeScript raise their idiomatic exceptions.

Silent truncation, masking, clamping, type coercion, and panic/trap/abort are
all contract violations. When a language's static type system makes an invalid
shape unconstructible, generated verification records the capability difference
instead of pretending to execute that case.

## Shared Vector Protocol

`bid754-codec-vectors/vectors.json` is a generated cross-language verification
artifact. Its top-level object has exactly these public fields:

| Field | Meaning |
| --- | --- |
| `format_version` | protocol version; currently `5` |
| `vectors` | successful raw decode/encode records |
| `reject_vectors` | inputs that must fail |
| `string_vectors` | successful parse/render closure records |

A consumer must reject an unsupported `format_version` explicitly.
Any incompatible field, encoding, or interpretation change increments
`format_version` before new vectors and consumers are generated. A compatible
record addition does not change the version.

### Successful Raw Records

Raw records contain the input words, expected logical fields, shared decimal
string, canonical flag, and expected canonical encoded words. Coefficient and
payload magnitudes are decimal strings so values above a language's native
64-bit integer range remain exact.

Every required consumer verifies:

- raw decode fields;
- decimal rendering;
- decimal parsing followed by encode;
- canonical raw round-trip when `canonical` is true; and
- exact little-endian byte decode and encode.

The exact anchor records and current record totals are generator data. They are
defined in `devtools/internal/testgen/bid_codec_vector_anchors.go`, emitted into
the vector artifact, checked independently by every generated consumer, and
pinned externally by `devtools/verification_anchors.json`; they are not
duplicated here.

### Reject Records

Each reject record has a `channel`:

- `from_string`: malformed input text;
- `encode`: width-specific invalid `Components`; or
- `to_string`: width-independent invalid or lossy `Components`.

An encode row names `bid32`, `bid64`, or `bid128`. Component magnitudes use
decimal strings, and `reason` names the violated boundary. An optional
`requires` capability tag marks a row that some language type systems cannot
construct. Such a consumer reports the capability skip; every constructible
row must fail through the public error channel.

### Successful String Records

A string record contains:

- `input`: text that every consumer must parse successfully; and
- `expected`: the exact rendering of the parsed components.

The consumer compares `toString(fromString(input))` byte-for-byte with
`expected`, then reparses and rerenders `expected` to prove closure. These rows
cover grammar normalization and valid component regions that no BID width can
encode directly.

## Required Consumers

Full BID codec vector verification requires all six standalone consumers: Go,
Rust, Java, Python, JavaScript/TypeScript, and Swift. The full Rust decimal
library and the full Go decimal library are additional consumers and cannot
replace the standalone Rust and Go consumers. A full-library consumer reaches
the reject and string channels through its own public parse surface, so a
vector whose channel that surface cannot express is reported as a counted
channel skip rather than silently dropped.

All consumers read the same generated vector artifact. They may not maintain
language-specific copies of expected records or substitute native decimal
formatting for the shared string contract. Language-native decimal adapters are
outside this shared public contract until separately specified and covered by
generated vectors.

Current record counts, long-corpus sizes, class partitions, and per-language
consumed/skip counts are verification state. Their source of truth is
`devtools/verification_anchors.json` and generated test output, not this
specification.
