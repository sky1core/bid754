# Intel BID Issues

Defects confirmed in pinned Intel BID C v20U4, with separate bid754 resolution
status. The source archive and build configuration are pinned in
[DEPENDENCIES_SPEC.md](DEPENDENCIES_SPEC.md). Upstream observations and bid754 resolution status are tracked separately.

All entries concern the pinned upstream version. The U5 column records the
measured recheck below; external report status remains **not checked**. This
list does not claim first discovery. A bid754 fix does not mean the upstream
is fixed.
The authoritative deviation policy remains in
[IEEE754_SPEC.md](IEEE754_SPEC.md#intentional-ieee-deviations-from-pinned-intel-bid-c);
listing a defect here does not register a deviation or permit a comparison skip.

## Index

| ID | Affected operation | Defect | bid754 status | U5 reproduction |
|---|---|---|---|---|
| INTEL-BID-001 | D32/D64 from-string | No-exponent overflow ignores directed rounding | Fixed; registered deviation | Reproduced |
| INTEL-BID-002 | D32/D64/D128 from-string | Exponent cap breaks exact cancellation | Fixed; registered deviation | Reproduced |
| INTEL-BID-003 | D128 scaleB | Representable finite value overflows | Fixed; registered deviation | Not reproduced |
| INTEL-BID-004 | D128 FMA | Exact subtraction packs an out-of-range exponent | Fixed; registered deviation | Reproduced |
| INTEL-BID-005 | D128 FMA | Overflow correction loses the negative sign | Fixed; registered deviation | Reproduced |
| INTEL-BID-006 | D32 quantum | Steering encoding decoded with the wrong mask | Fixed; registered optional-helper deviation | Reproduced |
| INTEL-BID-007 | D32/D64 from-string | Exponent underflow discards directed rounding | Fixed; registered deviation | Reproduced |
| INTEL-BID-008 | D32 from-string | No-exponent subnormal input is rounded twice | Fixed; registered deviation | Reproduced |
| INTEL-BID-009 | D64 from-string | Tiny positive input becomes a large negative value | Fixed; registered deviation | Reproduced |
| INTEL-BID-010 | D32/D64/D128 from-string | Exact zero raises range flags or becomes nonzero | Fixed; registered deviation | Reproduced |
| INTEL-BID-011 | D128 FMA | Opposite-sign subtraction triggers premature overflow | Fixed; registered deviation | Reproduced |
| INTEL-BID-012 | D32/D64/D128 scaleB | Prior Inexact causes false Underflow on an exact result | Fixed D128 raw port; registered deviation; public paths isolated | Reproduced |
| INTEL-BID-013 | D32/D64 from-string | Delayed rounding carry omits Underflow at the minimum normal boundary | Fixed; registered deviation | Not checked |

INTEL-BID-001/002 retain the existing registered deviations and generated
regression vectors. INTEL-BID-003 through 010 were confirmed by direct native
C execution on macOS arm64 on 2026-09-26; INTEL-BID-011 was found while
expanding their boundary families. These native observations have not been
verified on other platforms. The registered corrections and their matching
C neighbors are generated from [independent boundary models](../devtools/scripts/generate_numeric_regressions.py),
with [recorded pinned-C observations](../devtools/testdata/readtest_numeric_c_observations.json).
INTEL-BID-006 is a narrowly registered optional-helper correction; quantum
remains optional and has no public Go wrapper.

Native probes used the pinned build options, including `CALL_BY_REF=0`,
`GLOBAL_RND=0`, `GLOBAL_FLAGS=0`, `UNCHANGED_BINARY_FLAGS=0`,
`CFLAGS_OPT=-O3 -ffp-contract=off`, and the arm64 `BID_SIZE_LONG=8` override.
Calls started with status zero unless an entry specifies incoming status.
Hexadecimal flags below use the **Intel C**
layout: `0x20` Inexact, `0x28` Overflow|Inexact, `0x30` Underflow|Inexact.
Decimal128 raw bits are written `hi:lo`.

## U5 recheck

Direct native calls on macOS arm64 on 2026-10-03 compared the pinned U4 build
with [Intel's v20U5 archive](https://www.netlib.org/misc/intel/IntelRDFPMathLib20U5.tar.gz),
whose README is dated 2026-09-28. The U5 archive SHA-256 is
`85dafd70f0fe2a8da218ade4233fca9d3228b0b04cd6d8527f7499926be01037`.
Both builds used Apple Clang 21.0.0 and the options above. The repository's
canonical dependency remains U4.

Each version executed the existing 2,486 boundary rows plus 83 calls covering
the documented reproducers and controls. U4 matched all recorded pinned-C
observations. In U5, 180 boundary rows changed to the independent expected
value and status; no previously matching boundary row became a mismatch.
All 540 scaleB coefficient/exponent threshold rows and all 15 INTEL-BID-003
reproducer calls matched in U5. At least one documented reproducer still
failed for each of the other eleven issue IDs, including INTEL-BID-012 at
all three widths and all five modes with incoming Inexact.

These results cover the measured cases on macOS arm64. They do not establish
general U5 conformance or results on another platform. U5's other changelog
items are outside this twelve-issue recheck.

## INTEL-BID-001: no-exponent overflow rounding

Pinned `bid32_from_string` and `bid64_from_string` return infinity on the
no-exponent overflow path even when directed rounding requires the largest
finite value. For example, parse `"1" + "0" repeated 97 times` as D32 with
roundTowardZero: C returns `+Inf`; the required result is `9999999E90`, with
Overflow|Inexact. The D64 counterpart uses 385 zeros and expects
`9999999999999999E369`.

**bid754: fixed.** The Go mechanical port and generated Rust follow the
registered IEEE deviation. Regression sources:
[D32](../devtools/testdata/readtest_ieee754_regressions_bid32_cdiverge.in),
[D64](../devtools/testdata/readtest_ieee754_regressions_bid64_cdiverge.in).
This overflow defect is distinct from INTEL-BID-007's underflow defect.

## INTEL-BID-002: exponent cancellation

Pinned parsers stop exponent accumulation before subtracting fractional
digits. Decimal128 additionally counts a leading exponent zero against its
digit limit. A fraction equal to `10^-1000000` followed by `e01000000` becomes
zero with Underflow|Inexact instead of exact `1`. A fraction `10^-10485760`
followed by `e10485760` exposes the fixed cap in all three widths.

**bid754: fixed.** Existing issue
[D128-EXP-001](KNOWN_ISSUES.md#d128-exp-001-leading-zero-exponent-breaks-exact-cancellation)
contains the detailed reproducer and platform evidence. The shared
[regression source](../devtools/testdata/readtest_ieee754_exponent_cancellation.in)
covers the registered deviation. This issue concerns character conversion,
not the encoded-value scaleB defect below.

## INTEL-BID-003: D128 scaleB false overflow

- Call: `bid128_scalbn(x, 1, roundTiesToEven)`, with
  `x=(10^33-1)*10^6111`.
- C result: `7800000000000000:0000000000000000` (`+Inf`), flags `0x28`.
- Required: `(10^34-10)*10^6111`, flags zero; the result is exactly representable.
- Cause: the coefficient normalization condition compares only the upper limb
  at the `10^33` threshold. Pinned source: `bid128_scalb.c`.
- bid754: fixed; registered deviation, Tier 1.
  [Go port](../bid754-go/internal/bidgo/bid128_misc.go) and
  [ldexp port](../bid754-go/internal/bidgo/bid128_ldexp.go).

The existing `tier1ref` and Java BigDecimal probe both compute the correct
finite value for this input. The original deterministic corpus missed this
boundary combination. The expanded ScaleB corpus now includes coefficient-limb
thresholds, exponent transitions, both signs, and all five modes.

## INTEL-BID-004: D128 FMA missing overflow check

- Call: `bid128_fma(1E35, 1E6111, -10E6111, roundTiesToEven)`.
- C result: `6001ed09bead87c0:378d8e63ffffffff`, flags zero. The noncanonical
  encoding is interpreted as zero; the public string is `+0E-6173`.
- Required: `+Inf`, flags `0x28`. The exact result is
  `(10^34-1)*10^6112`, beyond the D128 finite range.
- Cause: an exact-subtraction branch packs exponent 6112 without checking
  the upper bound. Pinned source: `bid128_fma.c`.
- bid754: fixed; registered deviation, Tier 2.
  [Go port](../bid754-go/internal/bidgo/bid128_fma_body.go).

## INTEL-BID-005: D128 FMA overflow loses sign

- Call: `bid128_fma(1E35, -1E6111, 11E6111, roundTowardNegative)`.
- C result: `5fffed09bead87c0:378d8e63ffffffff` (positive maximum finite),
  flags `0x28`.
- Required: `-Inf`, flags `0x28`.
- Cause: overflow correction reads the sign before the result's sign bit is
  assembled. Pinned source: `bid128_fma.c`.
- bid754: fixed; registered deviation, Tier 2.
  [Go port](../bid754-go/internal/bidgo/bid128_fma_body.go) and
  [rounding helper](../bid754-go/internal/bidgo/bid128_fma_helpers.go).

INTEL-BID-004/005 differ from the previously fixed Go-only FMA reassembly
defect covered by
[bid128_fma_overflow_test.go](../bid754-go/internal/bidgo/bid128_fma_overflow_test.go).
The new counterexamples fail in pinned C itself.

## INTEL-BID-006: D32 quantum steering mask

- Call: `bid32_quantum(0x6ca00000)`; the input is `8388608E0`.
- C result: `0x6c800001` (`838860.9`), flags zero.
- Required quantum value: `1`, raw `0x32800001`, flags zero.
- Cause: `bid32_quantumd.c` uses a 64-bit steering mask on a 32-bit input,
  so the alternate exponent layout is never selected.
- bid754: fixed by correcting finite exponent extraction in the
  [Go port](../bid754-go/internal/bidgo/bid32_minmax.go) and regenerating Rust.
  The registered optional-helper exception does not add a public Go wrapper.
- Regression boundary: 896 generated cases cover all 192 finite biased
  exponents in both layouts and signs, coefficient boundaries, noncanonical
  finite inputs, and special values. Direct pinned-C observations separate
  432 matching controls from 464 steering-layout deviations. The two affected
  native FFI samples verify the independent correct result and the pinned-C
  erroneous result separately; they are not skipped.

## INTEL-BID-007: exponent underflow ignores the mode

- Calls: `bid32_from_string("1e-102", roundTowardPositive)` and
  `bid64_from_string("1e-399", roundTowardPositive)`.
- C result: raw zero in both widths, flags `0x30`.
- Required: D32 `1e-101`, D64 `1e-398` (raw `1` in each width), flags `0x30`.
- Cause: `bid32_string.c` and `bid64_string.c` overwrite the requested
  rounding mode with nearest-even in the exponent-underflow path.
- bid754: fixed; registered deviation, Tier 1.
  [D32 port](../bid754-go/internal/bidgo/bid32_string.go),
  [D64 port](../bid754-go/internal/bidgo/bid64_from_string.go).

The parser regression expectations now use the requested directed rounding
instead of preserving the erroneous zero result.

## INTEL-BID-008: D32 no-exponent double rounding

- Call: `bid32_from_string("0." + "0" repeated 100 times + "14999999",
  roundTiesToEven)`.
- C result: raw `0x00000002` (`2e-101`), flags `0x30`.
- Required: raw `0x00000001` (`1e-101`), flags `0x30`. The equivalent
  exponent spelling `14999999e-108` returns this correct result in C.
- Cause: `bid32_string.c` first rounds the coefficient to seven digits,
  then rounds again at underflow without preserving the original tail.
- bid754: fixed; registered deviation, Tier 1.
  [Go port](../bid754-go/internal/bidgo/bid32_string.go).

## INTEL-BID-009: D64 tiny positive input becomes negative

- Call: `bid64_from_string("0." + "0" repeated 894 times + "12345678901234567",
  roundTiesToEven)`.
- C result: raw `0xc00462d53c8abac1`, value `-1234567890123457E114`,
  flags `0x20`.
- Required: `+0`, flags `0x30`.
- Cause: `bid64_string.c` passes a negative biased exponent to an
  overflow-only packing path in `bid_internal.h`.
- bid754: fixed; registered deviation, Tier 1.
  [Go port](../bid754-go/internal/bidgo/bid64_from_string.go).

## INTEL-BID-010: exact-zero underflow packing

- Call: `bid128_from_string("0e-6211", roundTowardPositive)`.
- C result: `0000000000000000:0000000000000001` (`1e-6176`), flags `0x30`.
- Numerical requirement: an exact zero must not become nonzero.
- Cause: the extreme-underflow packer in `bid_internal.h` chooses a minimum
  nonzero coefficient for directed rounding without first checking for zero.
- bid754: fixed; registered deviation, Tier 1. [Go helper](../bid754-go/internal/bidgo/bid128_div.go).

D32 `0e-109` and D64 `0e-415` also raise `0x30` in pinned C under all five
modes, despite being exact zero. The mode-reset defect in INTEL-BID-007
keeps their C result numerically zero. Restoring the requested mode exposes
the same missing zero guard, so all three widths require the guard. These
additional native results were verified on macOS arm64.

The public bid754 contract additionally requires rejecting an unrepresentable
written zero cohort. Its status-dependent validation bypass in
[types_bidgo_runtime.go](../bid754-go/types_bidgo_runtime.go) is a separate
project-side defect, not an Intel public-API contract.

## INTEL-BID-011: D128 FMA premature overflow near the finite limit

- Call: `bid128_fma(1E35, 1E6110, -9E6110, roundTiesToEven)`.
- C result: `+Inf`, flags `0x28`.
- Required: maximum finite `(10^34-1)*10^6111`, flags `0x20`.
  The exact result exceeds maximum finite by only `0.1` ulp, so nearest
  rounding remains finite.
- Cause: `bid128_fma.c` classifies overflow before opposite-sign subtraction
  can bring the rounded result back into range.
- Direct pinned-C probes on macOS arm64 also fail for subtrahends
  `6E6110` through `9E6110`, both signs, and both nearest modes.
  The `4E6110`, `5E6110`, `10E6110`, and `11E6110` neighbors distinguish
  overflow, exact maximum finite, and inexact finite results.
- bid754: the correction in
  [the Go FMA port](../bid754-go/internal/bidgo/bid128_fma_body.go)
  is registered and covered by generated regression vectors. This is distinct from the out-of-range packing
  and lost-sign defects in INTEL-BID-004/005.

## INTEL-BID-012: prior Inexact causes false scaleB Underflow

- Call: `bid*_scalbln(x, -1, mode, &status)`, with raw `x=10` and
  incoming status `0x20` (Inexact). The encoded exponent is the width's
  minimum: -101, -398, or -6176.
- Required: raw `1`, retaining status `0x20`. Division of the coefficient
  by ten is exact; an exact subnormal result does not raise Underflow.
- Pinned C result: raw `1`, status `0x30` (Underflow|Inexact), at all three
  widths and all five modes. With incoming status zero, the same calls
  return raw `1` and status zero.
- Cause: underflow packers in `bid_internal.h` use the accumulated Inexact
  bit to decide whether the current operation raises Underflow.
- Direct native execution on macOS arm64 confirmed these results on
  2026-10-03 with the pinned build options above. Other platforms have
  not been checked.
- bid754: the Go D128 pointer-status scaleB port computes the final packer's
  flags in fresh operation-local status and then accumulates them into the
  caller's word. Generated Rust inherits this correction. The registered
  deviation has generated regression vectors with incoming Inexact;
  exact subnormal discrepancies skip native comparison, while matching
  zero, inexact-underflow, overflow, and NaN neighbors still execute it.
  The readtest status-strength check exercises both raw entrypoints, both
  signs, all five modes, and all combinations of the five incoming IEEE flags.
  Public Go and Rust scaleB wrappers start each
  operation with fresh status and accumulate returned flags in the separate
  Context carrier; the tested public scaleB calls do not raise Underflow
  when Context already carries Inexact.

## INTEL-BID-013: delayed carry loses parser Underflow

- Calls: `bid32_from_string("999999901e-104", roundTowardPositive)` and
  `bid64_from_string("999999999999999901e-401", roundTowardPositive)`.
- Pinned C returns the minimum normal value, respectively raw `000f4240`
  and `00038d7ea4c68000`, with only Inexact (`0x20`).
- Required: the same value with Underflow|Inexact (`0x30`). The exact input
  is tiny before rounding. Negative inputs with roundTowardNegative have
  the same defect.
- Cause: a zero first discarded digit followed by a nonzero digit triggers
  a delayed directed-rounding carry. Premature coefficient normalization
  increments the exponent before the parser tests for underflow.
- bid754: normalization is deferred to the packer; registered deviation,
  Tier 1. Generated regressions cover both spellings, signs, all modes,
  carry and non-carry coefficients, and neighboring exponent ranges.
- Confirmed against pinned v20U4 on macOS arm64 on 2026-10-07. U5 has not
  been checked for this entry.
