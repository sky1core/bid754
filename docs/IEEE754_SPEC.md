# IEEE 754 BID Decimal Floating-Point Specification

This document is the detailed IEEE behavior document that supplements `SPEC.md`.

## Scope

- The encoding baseline of this project is BID
- DPD is not a primary implementation goal
- Even when DPD appears in documents, it is interpreted only in comparison, conversion, or reference contexts

## Formats

| Format | Significant digits | Exponent range | bias | Bytes |
|------|----------|-----------|------|--------|
| Decimal32 | 7 digits | -95 ~ +96 | 101 | 4 |
| Decimal64 | 16 digits | -383 ~ +384 | 398 | 8 |
| Decimal128 | 34 digits | -6143 ~ +6144 | 6176 | 16 |

## Mandatory behavior

For formats/operations claimed as supported, the following are mandatory.

- BID encoding interpretation
- Rounding mode application
- Exception flag accumulation
- Special value handling
- Subnormal/non-canonical input handling
- Match with decTest expected values and flags

## Mandatory operations and optional operations

This project does not decide mandatory status based on "operations currently implemented".

The classification baseline is the IEEE 754-2019 text.

- In 1.6, `shall` is a mandatory requirement
- In 1.6, `should` is a recommendation
- In 5.1, Clause 5 operations shall be provided for supported arithmetic formats
- 9.0 states that Clause 5 completely specifies the required operations, and explicitly fixes Clause 9 as recommended operations

Therefore, the implementation/verification priority baseline of this repository is not the current code state but the IEEE 754-2019 `shall` / `should` / `recommended` classification.

- Clause 5 `shall` operations are mandatory implementation scope
- `should` items within Clause 5 are recommended scope
- Clause 9 `Recommended operations` are optional/recommended scope
- A mandatory operation must not be downgraded to optional just because it is not yet implemented
- The only items that may be classified as optional are those the standard actually places as optional/recommended

### Mandatory implementation operation categories

For decimal arithmetic formats claimed as supported, the mandatory operation categories of Clause 5 are as follows.

| Clause | Classification | Mandatory operations |
|--------|------|-----------|
| 5.3.1 | homogeneous general-computational | `roundToIntegralTiesToEven`, `roundToIntegralTiesToAway`, `roundToIntegralTowardZero`, `roundToIntegralTowardPositive`, `roundToIntegralTowardNegative`, `roundToIntegralExact`, `nextUp`, `nextDown`, `remainder` |
| 5.3.2 | decimal operations | `quantize` |
| 5.3.3 | logBFormat operations | `scaleB`, `logB` |
| 5.4.1 | arithmetic operations | `addition`, `subtraction`, `multiplication`, `division`, `squareRoot`, `fusedMultiplyAdd`, `convertFromInt`, `convertToIntegerTiesToEven`, `convertToIntegerTowardZero`, `convertToIntegerTowardPositive`, `convertToIntegerTowardNegative`, `convertToIntegerTiesToAway`, `convertToIntegerExactTiesToEven`, `convertToIntegerExactTowardZero`, `convertToIntegerExactTowardPositive`, `convertToIntegerExactTowardNegative`, `convertToIntegerExactTiesToAway` |
| 5.4.2 | conversion operations | conversions between supported floating-point formats, decimal character sequence <-> supported floating-point format conversions |
| 5.5.1 | quiet-computational sign operations | `copy`, `negate`, `abs`, `copySign` |
| 5.5.2 | decimal re-encoding | `encodeDecimal`, `decodeDecimal`, `encodeBinary`, `decodeBinary` — within the supported encoding scope only; see the interpretation rules below and the DPD/binary interchange rows in the Intel group table |
| 5.6.1 | comparisons | `compareQuietEqual`, `compareQuietNotEqual`, `compareSignalingEqual`, `compareSignalingGreater`, `compareSignalingGreaterEqual`, `compareSignalingLess`, `compareSignalingLessEqual`, `compareSignalingNotEqual`, `compareSignalingNotGreater`, `compareSignalingLessUnordered`, `compareSignalingNotLess`, `compareSignalingGreaterUnordered`, `compareQuietGreater`, `compareQuietGreaterEqual`, `compareQuietLess`, `compareQuietLessEqual`, `compareQuietUnordered`, `compareQuietNotGreater`, `compareQuietLessUnordered`, `compareQuietNotLess`, `compareQuietGreaterUnordered`, `compareQuietOrdered` |
| 5.7.1 | conformance predicates | `is754version1985`, `is754version2008`, `is754version2019` |
| 5.7.2 | non-computational general operations | `class`, `isSignMinus`, `isNormal`, `isFinite`, `isZero`, `isSubnormal`, `isInfinite`, `isNaN`, `isSignaling`, `isCanonical`, `radix`, `totalOrder`, `totalOrderMag` |
| 5.7.3 | decimal non-computational operation | `sameQuantum` |
| 5.7.4 | operations on subsets of flags | `lowerFlags`, `raiseFlags`, `testFlags`, `testSavedFlags`, `restoreFlags`, `saveAllFlags` |

The table above is treated as required scope for the BID decimal formats this repository claims to support.

Clause 5.7.1 conformance predicates are satisfied by the hand-written
`bid754-go/version.go` constants (`Is754Version1985`, `Is754Version2008`,
`Is754Version2019`). The generated readtest selection excludes the Intel
`bid_is754`/`bid_is754R` helper rows because 5.7.1 is answered by these
constants, not by routing the helpers through the port. The exact current rows
and reasons are emitted in the generated readtest profile inventory.

Additional enforced principles:

- If `Decimal32`, `Decimal64`, and `Decimal128` are documented as supported, all Clause 5 `shall` operations that apply to those formats are implementation targets
- Even when IEEE 754 defines an operation, operations related only to types, encodings, or external interchange forms this repository does not support are not raised to mandatory scope
- Support status is judged by the scope the project actually claims. In this repository BID decimal is the primary scope, and binary formats not raised to the support surface in the current phase or the unsupported DPD codec are not interpreted as mandatory
- Even when familiar names such as `nextToward`, `minimum`, and `maximum` appear in the codebase, mandatory status is judged by the IEEE 754-2019 classification, not by name similarity

### Public Go exception flag policy

Public Go BID methods use the following flag policy:

- value-only compatibility methods such as `Add`, `Sub`, `Mul`, `Div`, `Quantize`,
  and `RoundIntegralExact` are compatibility convenience APIs and may discard
  exception flags
- every public BID operation that can raise IEEE exception flags must expose a
  flag-returning public path, either as the primary method signature or as a
  `WithFlags` peer for an existing value-only method
- context-based operations accumulate flags into `ArithmeticContext.Flags` using
  sticky OR semantics
- `DefaultArithmeticContext()` returns a snapshot of the atomic default
  rounding mode; callers must not rely on mutating the returned pointer to
  change package-global state
- slice helper APIs that perform flag-raising operations must have a
  flag-returning peer that accumulates per-step flags
- flag-returning public methods must route through the Go mechanical port flag
  path and must not manufacture flags independently of that path
- frequently omitted items such as integer conversion, string conversion, flag subset operations, `sameQuantum`, `totalOrder`, and `totalOrderMag` are also included in mandatory scope

### Rust rounding-mode compatibility policy

The full `bid754-rs` crate is the public Rust implementation, and its public
API is generated from the Go mechanical port by the `devtools/tools/go2rs`
apiemit subpass. The public `RoundingMode` enum exposes exactly the five IEEE
rounding modes (`NearestEven`, `NearestAway`, `TowardZero`, `TowardPositive`,
`TowardNegative`) and nothing else. The non-IEEE decTest-compatibility rounding
value (`BID_ROUNDING_NEAREST_DOWN`) is deliberately not representable in that
public enum; it exists only on the `#[doc(hidden)]` non-stable
verification-plumbing surface — the generated internal modules plus the
crate-root raw-parse shim that the generated verification harnesses call.
Non-IEEE rounding modes must stay behind that verification/compatibility
surface and must not be exposed as normal IEEE rounding modes in the public
rounding enum.

### Public Go constant policy

`Zero*BID`, `One*BID`, `Pi*BID`, and `E*BID` are convenience constants for the
public Go BID value types. They are not an independent mathematical constant
engine. Their values are defined by the checked-in decimal literals in
`api_v2.go` and must be initialized through the same Go mechanical-port string
constructor path as ordinary public values.

`Pi*BID` and `E*BID` use one literal per BID width: 7 significant digits for
`Decimal32`, 16 significant digits for `Decimal64`, and 34 significant digits
for `Decimal128`. Changing these literals is an API change and requires tests
that compare the exported constants with constructor results for the documented
literals.

### Recommended/optional operation categories

The recommended/optional categories are as follows.

| Source | Classification | Examples |
|------|------|------|
| Clause 5 `should` | recommended | decimal `quantum`, translation-time predicate availability, some non-interchange/sign-bit canonicalization recommendations |
| 9.2 | additional mathematical operations | `exp`, `expm1`, `exp2`, `exp10`, `log`, `log2`, `log10`, `logp1`, `hypot`, `rSqrt`, `compound`, `pow`, `pown`, `powr`, `sin`, `cos`, `tan`, `asin`, `atan`, `sinh`, `cosh`, `tanh`, etc. |
| 9.3 | dynamic mode operations | `setDecimalRoundingDirection`, `setBinaryRoundingDirection`, `saveModes`, `restoreModes`, `defaultModes` |
| 9.4 | reduction operations | `reduceSum`, `dot`, `sum`, `sumAbs`, `sumSquare` family |
| 9.5 | augmented arithmetic operations | `augmentedAddition`, `augmentedSubtraction`, `augmentedMultiplication`, `augmentedDivision` family |
| 9.6 | minimum/maximum operations | `minimum`, `minimumNumber`, `maximum`, `maximumNumber`, `minimumMagnitude`, `minimumMagnitudeNumber`, `maximumMagnitude`, `maximumMagnitudeNumber` |
| 9.7 | NaN payload operations | `getPayload`, `setPayload`, `setPayloadSignaling` |

This repository manages only the Clause 9 and Clause 5 `should` items in the table above as optional/recommended.

## Implementation targets organized by Intel BID function group

The implementation target inventory is not written scattered arbitrarily.

Since the Intel BID upstream already exposes function groups as group units in its test files, this repository follows that structure as-is.

Primary reference files:

- `devtools/third_party/intel_dfp/TESTS/readtest.h`
- `devtools/third_party/intel_dfp/TESTS/readtest.in`
- `devtools/third_party/intel_dfp/TESTS/test_bid_functions.h`

Documentation principles:

- Implementation targets are written first at the Intel BID function-group level rather than as enumerations of individual function names
- Intel `readtest` regular verification documentation also records the actual result comparison groups (`CMP_FUZZYSTATUS`, `CMP_EQUALSTATUS`, `CMP_RELATIVEERR`), separately from the function-group table
- When describing the Intel `readtest` implementation scope, do not write `all of CMP_FUZZYSTATUS`; write the historical operative scope as-is
- For each function group, mark `IEEE status` as `mandatory` or `optional/recommended`
- The fact that a function exists in the Intel BID upstream and its IEEE mandatory status are written separately
- Implementation/verification plans are managed based on this function-group table

historical operative scope:

- `CMP_FUZZYSTATUS - explicit historical skip function groups + CMP_EQUALSTATUS`
- `CMP_RELATIVEERR` is excluded as a profile-expansion group, but the Intel duplicate `CMP_RELATIVEERR` comparator rows for `bid32_fmod` / `bid64_fmod` / `bid128_fmod`, already included in the `CMP_FUZZYSTATUS` surface, may be applied separately per generated runner
- `TEST_GENERATION_SPEC.md` defines the selection policy; the generated
  readtest profile inventory contains the exact current exclusions and reasons

| Intel BID function group | Representative function examples | IEEE status | Notes |
|------------------|----------------|-------------|------|
| String conversion | `bid32/64/128_from_string`, `bid32/64/128_to_string` | mandatory | Clause 5.4.2 decimal character sequence conversion |
| Basic arithmetic | `bid32/64/128_add`, `sub`, `mul`, `div` | mandatory | Clause 5.4.1 |
| Square root | `bid32/64/128_sqrt` | mandatory | Clause 5.4.1 |
| fused multiply-add | `bid32_fma`, `bid64*_fma`, `bid128*_fma` | mandatory | Clause 5.4.1 |
| round-to-integral family | `bid*_round_integral_nearest_even`, `nearest_away`, `positive`, `negative`, `zero`, `exact` | mandatory | Clause 5.3.1 |
| next family | `bid*_nextup`, `bid*_nextdown` | mandatory | Clause 5.3.1 |
| next family extension | `bid*_nexttoward`, `bid*_nextafter` | Intel inventory | IEEE mandatory status is judged by the standard classification, not by name similarity. In the current phase this repository raises `bid32/64/128_nexttoward` to the support surface, and `bid*_nextafter` is handled only on the generated readtest verification surface without public wiring |
| remainder family | `bid*_rem` | mandatory | Clause 5.3.1 remainder |
| remainder variant | `bid*_fmod` | Intel inventory | not treated as identical to the IEEE mandatory remainder |
| decimal quantize | `bid32/64/128_quantize` | mandatory | Clause 5.3.2 |
| logB / scaleB | `bid*_logb`, `bid*_ilogb`, `bid*_scalbn`, `bid*_scalbln` | mandatory core | Based on Clause 5.3.3. The Intel names appear as the `logb`/`ilogb`/`scalbn` family |
| Integer conversion | `bid*_to_int*`, `bid*_to_uint*` and exact/rounding variants | mandatory | Clause 5.4.1 integer conversion family |
| sign-bit quiet operations | `bid*_copy`, `negate`, `abs`, `copySign` | mandatory | Clause 5.5.1 |
| decimal re-encoding / codec | `bid_dpd_to_bid*`, `bid_to_dpd*`, etc. | optional/out of scope unless explicitly claimed | Clause 5.5.2 is interpreted within the supported encoding scope. This repository does not place DPD as a first-class support goal, so the DPD codec is not treated as mandatory |
| Comparison | `bid*_quiet_*`, `bid*_signaling_*`, compare family | mandatory | Clause 5.6.1 |
| Classification / predicate / ordering | `bid*_class`, `isSigned`, `isNormal`, `isSubnormal`, `isFinite`, `isZero`, `isInf`, `isNaN`, `isSignaling`, `isCanonical`, `totalOrder*` | mandatory | Clause 5.7.2 |
| decimal quantum relation | `bid*_sameQuantum` | mandatory | Clause 5.7.3 |
| flag subset operations | state-control paths corresponding to `lowerFlags`, `raiseFlags`, `testFlags`, `testSavedFlags`, `restoreFlags`, `saveAllFlags` | mandatory | Clause 5.7.4 |
| minimum/maximum family | `bid*_minnum`, `maxnum`, `minnum_mag`, `maxnum_mag` | optional/recommended | in IEEE 754-2019 the Clause 9.6 minimum/maximum family is a recommended category |
| quantum query | `bid*_quantum` | optional/recommended | Clause 5 `should` example |
| Additional mathematical functions | `exp`, `log`, `pow`, `sin`, `cos`, `tan`, `hypot`, `tgamma`, etc. | optional/recommended | Clause 9.2 |
| NaN payload operations | payload getter/setter family | optional/recommended | Clause 9.7 |

Support scope interpretation rules:

- binary format related conversions not raised to the support surface in the current phase are not mandatory
- in the current phase, the one-way `bid32/64/128 -> binary32/64/128` conversion helpers are regarded as support surface
- the 6 BID width conversions (`bid32<->bid64<->bid128` widening/narrowing) and `bid32/64/128_nexttoward` are included in the current phase support surface
- binary80 and reverse binary -> BID conversions are still not part of the current phase support surface
- the BID <-> DPD codec is not mandatory unless DPD support is explicitly declared
- the mere fact that a function appears in the Intel upstream does not automatically place it in this repository's mandatory scope

The table above is the base mapping table in this repository that connects Intel BID implementation targets with the IEEE mandatory/optional classification.

## Rounding modes

IEEE 754 mandatory mapping:

| Mode | Meaning |
|------|------|
| roundTiesToEven | nearest even |
| roundTowardNegative | toward negative infinity |
| roundTowardPositive | toward positive infinity |
| roundTowardZero | toward zero |
| roundTiesToAway | away from zero on ties |

Non-IEEE modes are handled separately as distinct support. If unsupported, they are explicitly handled as skip/unsupported.

## Exception flags

Mandatory IEEE exception categories within the supported BID decimal scope:

- invalid
- division by zero
- overflow
- underflow
- inexact

The public Go `ExceptionFlags` type also carries `Rounded`, `Subnormal`, and
`Clamped` for IBM GDA/decTest verification plumbing. Those names are GDA
conditions, not additional IEEE exception categories, and the public
arithmetic runtime does not synthesize them. The pinned Intel BID C arithmetic
`_IDEC_flags` path reports the five categories above. Intel separately defines
the IA-specific `DEC_FE_UNNORMAL` / `BID_DENORMAL_EXCEPTION` bit and sets it for
subnormal binary operands in binary-to-decimal conversion and through
`bid_feraiseexcept`; that bit is not the IBM GDA `Subnormal` result condition
and is not mapped onto the public decimal exception surface. The Go mechanical
port and generated Rust therefore preserve the Intel five-flag decimal runtime
surface rather than inventing status behavior with no canonical predecessor.

decTest handling is executor-specific: rows dispatched to native decNumber
compare the complete parsed GDA condition set exactly, while cases executed by
Go mechanical-port operation adapters and the portable Go/Rust legs project
conditions onto the Intel BID five-flag surface. `Rounded`, `Subnormal`, and
`Clamped` alone are not implementation-gap skip reasons. The exact comparison
and remaining value-semantic divergence rules are defined in
`TEST_GENERATION_SPEC.md`.

## Special values

Mandatory handling targets:

- `+0`, `-0`
- `+Inf`, `-Inf`
- `qNaN`
- `sNaN`

NaN payload, the quiet/signaling distinction, and sign preservation must be precisely defined and verified within the scope claimed as supported.

## Non-canonical input

Encoding interpretation of non-canonical BID input must follow the same rules
as the original C (this rule is limited to BID bit-encoding interpretation —
IEEE deviations in operation semantics such as `from_string` overflow follow
the "Intentional IEEE deviations from pinned Intel BID C" section below). If
the encoding interpretation differs from C, it is a bug.

## Intentional IEEE deviations from pinned Intel BID C

In principle, the Go mechanical port in this repository must match the
behavior of pinned Intel BID C. However, only where the pinned C implementation
conflicts with an IEEE 754-2019 `shall` requirement may an intentional
deviation that follows IEEE behavior be made. All intentional deviations are
registered in this section. Any C mismatch not registered here is a bug.

Registration requirements:

- the deviation boundary is measured by direct execution (official C probe or native comparison) and recorded
- an accurate `native_compare_skip_reason` is attached to the corresponding
  readtest block in `devtools/testgen_manifest.json` (rows matching C are not skipped along with it)
- the IEEE expected behavior is checked in as regression vectors and placed in the regular verification domain

Rows carrying a `native_compare_skip_reason` are skipped only by the native C
comparison; the generated Go-port readtest gate (`TestGeneratedReadCasesGoPort`)
executes them directly against the Go mechanical port and they must pass there.

Currently registered deviations:

1. `bid32_from_string` / `bid64_from_string` no-exponent overflow path:
   pinned Intel C ignores the rounding mode on this path and always returns
   Inf. Following the IEEE 754-2019 §7.4 overflow + directed rounding
   semantics, this repository returns the largest-magnitude finite value
   including the sign under roundTowardZero/roundTowardNegative (positive) and
   roundTowardZero/roundTowardPositive (negative) (the negative largest finite
   when negative). The exponent-notation path is not a deviation because
   pinned C also conforms to IEEE there. `bid128_from_string` has no deviation
   because pinned C conforms to IEEE on both paths.

## decTest

decTest is verification data. It is a pass/fail criterion, not a documentation reference.

Files by format:

- `ds*.decTest` -> Decimal32
- `dd*.decTest` -> Decimal64
- `dq*.decTest` -> Decimal128

Generic case files are dispatched appropriately according to precision/context.

## BID vs DPD

Summary:

- The implementation goal of this project is BID
- DPD is not a primary implementation goal
- BID and DPD must not be documented as equal product goals

## IEEE sign operations and decTest name collisions

The quiet-computational sign operations of IEEE 754 Clause 5.5.1 are managed in this repository as `copy`, `negate`, `abs`, and `copySign`. The representative paths on the Intel BID side are `bid*_copy`, `bid*_negate`, `bid*_abs`, and `bid*_copySign`, and public Go value-type methods must connect through this Go mechanical port path.

IBM decTest contains General Decimal Arithmetic operations that are similar in name but different in nature.

- `copy`, `copyAbs`, `copyNegate`, `copySign`: a quiet sign-bit copy family that corresponds to the current BID copy-family verification
- `abs`, `minus`, `plus`: GDA computational-style operations that include sNaN quieting, Invalid flag, and zero sign rules, and are not treated the same as the copy-family above
- `canonical`: includes DPD/encoding canonicalization tagged literal (`#...`) verification, so it is not automatically placed into the BID-only current surface
- `and`, `or`, `xor`, `invert`, `rotate`, `shift`: GDA decimal logical/digit operations; since there is currently no public Go BID mechanical-port path, they are not used as evidence that current mandatory BID verification is complete
- `divideInt`: a GDA integer-quotient divide operation; since there is currently no adapter fixed to the Intel BID/Go mechanical-port combination, it is not treated as identical to `divide` verification
- `reduce`: the trailing-zero reduction operation of decTest is a GDA operation with no canonical `bid*_reduce` predecessor in pinned Intel BID C, so it is outside the public BID arithmetic surface at every width

Therefore, when expanding decTest files, IEEE sign-bit quiet operations are not considered verified based on operation names alone. A file is placed into the current supported subset only when the result values, sNaN handling, flags, and tagged literal encoding scope match the support surface of the public BID path.

Items currently deferred from decTest expansion are classified not as mandatory current BID fixed-width scope but as DPD/tagged literal, GDA logical/digit, General arbitrary-precision, optional/recommended operation, or an unselected public surface gap. To reclassify an item as mandatory, the support surface and IEEE rationale must be documented first.
