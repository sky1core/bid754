# Automated Verification Generation Spec

This document defines the normative policy for generated verification. It
supplements `SPEC.md`; it does not describe the current generated inventory or
duplicate generator constants.

The standalone cross-language BID codec API and vector protocol are defined in
`BID_CODEC_SPEC.md`. The current verification implementation map is summarized
in `VERIFICATION_REFERENCE.md`. Commands belong in `BUILD.md` and the root
`Makefile`.

## Document Boundary

The following ownership split is mandatory:

| Information | Source of truth |
| --- | --- |
| verification policy, oracle strength, completion rules | this specification |
| standalone BID codec API and vector protocol | `BID_CODEC_SPEC.md` |
| selected inputs, functions, profiles, seeds, and generation parameters | `devtools/testgen_manifest.json` and generator code |
| exact case counts, class partitions, tuple hashes, and artifact hashes | `devtools/verification_anchors.json` |
| current selected/excluded inventory and skip accounting | generated inventories under `devtools/generated/testspec/` |
| execution graph and local commands | `Makefile`, `devtools/scripts/`, and `BUILD.md` |
| human-readable current implementation map | `VERIFICATION_REFERENCE.md` |

Derived counts, complete function lists, and current skip inventories must not
be copied into this specification. They change with generated state and must be
read from the machine-readable sources above.

## Principles

- Regular verification is generated mechanically from authoritative inputs.
- A regular domain is generated end to end: input selection, case/spec data,
  dispatch or wrappers, and the runner or harness.
- Hand-written glue is not an acceptable final component of a generated regular
  verification path.
- Official source data must not be replaced by hand-maintained expected values.
- Scope grows by changing selection rules and regenerating a batch, not by
  adding functions, files, or cases one at a time.
- Generated artifacts are reproducible and checked in; current inventory is
  measured from those artifacts rather than asserted in prose.
- A smoke or shard result must be reported as such. It is never evidence that a
  full domain passed.
- When Go and Rust verify the same implementation contract, they consume the
  same generated rules and deterministic corpus.

## Performance Measurement Contract

- Benchmarks execute production code through the exact layer being measured.
  Public API, direct Go mechanical port, generated Rust, and native Intel C
  results are identified separately.
- Compared implementations use semantically identical, exactly representable
  operands. Benchmark setup rejects invalid or inexact shared operands, and a
  test fixes that shared input contract.
- Every result reaches an observable sink or the benchmark framework's
  black-box mechanism so the measured operation cannot be removed.
- Before/after claims hold operation, operands, rounding mode, status handling,
  build mode, and benchmark layer constant. Results with multiple major
  variable changes are exploratory, not causal.
- Evidence records the command, platform, CPU, sample or repetition count,
  timing distribution or median, and allocation results when the benchmark
  harness reports them, with distinct baseline and candidate results. Runs
  affected by material host-load changes, thermal
  throttling, or setup errors are discarded and reported separately.
- An optimization is accepted only after the applicable exact correctness
  checks pass. Tier 1 optimization requires the full applicable generated
  differential checks rather than a reduced smoke subset.
- Tier 1 performance work closes only with stable repeated before/after
  measurements for the affected hot operations and an explicit check for
  unexplained regressions in adjacent measured operations.

## Regular Verification Domains

The repository has exactly four regular verification domains:

1. Intel `readtest`
2. IBM `decTest`
3. C FFI exact bit-compare
4. BID codec vectors

Public-API parity runners, long corpus extensions, fuzzing, package verification, and
platform digests are additional gates. They strengthen the architecture or a
regular domain but do not create a fifth regular domain.

A regular domain is closed only when:

- its authoritative inputs and generated outputs are reproducible;
- every selected case is executed or assigned a generated, reasoned
  classification;
- every required consumer runs at the required comparison strength;
- external count and content anchors match; and
- the generated inventory has no `unresolved_required` entry for the claimed scope.

## Common Generation Contract

### Authoritative Inputs

| Domain | Authoritative input |
| --- | --- |
| Intel `readtest` | pinned Intel BID `readtest.h` and `readtest.in` |
| IBM `decTest` | pinned official `*.decTest` files |
| C FFI exact bit-compare | pinned Intel BID signatures plus deterministic generated inputs |
| BID codec vectors | independent BID layout reference code plus the generation manifest |

Pinned inputs may be absent from a portable consumer checkout. In that case an
input-dependent sync test may skip the whole input-dependent leg. A present but
unreadable, malformed, or checksum-invalid input must fail. `verify-generated`
must provision the pinned inputs and therefore may not convert absence into a
pass.

### Reproducibility

`make verify-generated` is the reproducibility boundary. It must regenerate
every declared artifact and byte-compare it with the checked-in tree. Every
tracked file carrying the standard generated-code marker must be included in
that comparison or in the documented generated-marker exceptions.

Reproducibility alone does not detect simultaneous reductions in generator
assertions and generated-output checks. `devtools/verification_anchors.json`, maintained outside
the generation path, therefore pins:

- scale through counts and class partitions;
- deterministic corpus identity where required;
- verification-participating artifact content through tree hashes; and
- cross-consumer relationships that must remain equal.

An intentional generated-content change updates the generator or manifest,
regenerates the outputs, reviews the measured change, and only then updates the
external anchor. No generator may update its own external anchor.

### Generated Inventory

`devtools/generated/testspec/spec_index.json` and its referenced shards are the
closed current inventory. Generated dispatch inventories record which selected
entries dispatch, skip, or receive a narrowly classified comparison exemption.
Every generated inventory is closed in both directions: a selected item cannot
disappear silently, and an undeclared item cannot enter a completion count.

## Intel readtest Contract

### Operative Profile

The repository profile is:

`CMP_FUZZYSTATUS - explicit historical skip groups + CMP_EQUALSTATUS`

It must not be described as all of `CMP_FUZZYSTATUS`. Non-`fmod`
`CMP_RELATIVEERR` math/transcendental groups remain profile expansion. Generated
runners may additionally apply the duplicate Intel `CMP_RELATIVEERR` rows for
`bid32_fmod`, `bid64_fmod`, and `bid128_fmod`, because those functions are
already selected by the `CMP_FUZZYSTATUS` surface.

Unsupported binary formats, DPD interchange, FE APIs, version-predicate
helpers, and Intel extensions are classified by the selection code, not by a
prose function list. Mixed-width arithmetic is not one optional bucket: the
Tier 1 D/Q `add`/`sub`/`mul`/`div` families producing Decimal64 or Decimal128
are selected. The pinned Intel mixed-width FMA/sqrt extension families are
also selected through their direct mechanical-port entrypoints; their D/Q
operand order and direct destination-width rounding must not be replaced by a
widen/operate/narrow composition. The exact current decisions and reasons are
owned by `devtools/internal/testgen/readtest_spec.go` and the generated
readtest profile inventory.

Function-group membership is extracted mechanically from pinned Intel headers.
Each discovered group carries an IEEE scope classification, and mandatory
groups in the claimed repository profile must be closed. Documentation and
generated inventories preserve both axes of the upstream model: function group and
result-comparison group.

### Consumers and Comparison Strength

The native C-oracle runner, the Go mechanical-port runner, and the generated
Rust runner consume the same generated readtest specification. Their dispatch
surfaces must be cross-checked, and unresolved or ambiguous mappings fail
generation unless a permitted classification is emitted by the selection
path.

- Decimal results and status are compared with the pinned Intel comparator
  semantics for the row's comparison group.
- `CMP_FUZZYSTATUS` and `CMP_EQUALSTATUS` operation flags compare exactly.
- The duplicate `fmod` `CMP_RELATIVEERR` rows use the pinned Intel relative
  comparator behavior and its flag mask only on that path.
- Scalar and secondary outputs such as `frexp` exponents and `modf` integral
  results compare exactly according to upstream `readtest.c`.
- The Go-port runner may not replace these comparisons with float tolerance or
  a decTest cohort comparator.

Rows added because IEEE 754 requires behavior absent from pinned Intel inputs
enter through manifest-declared readtest blocks:

- `cmatch` rows keep native C comparison;
- `cdiverge` rows are allowed only for deviations registered in
  `IEEE754_SPEC.md`, carry an accurate native skip reason, and still execute in
  the Go and Rust expected-result runners.

## IBM decTest Contract

The manifest selects official suites mechanically from their operation sets.
The generated inventory accounts for every official input file and every case in a
selected fixed-width suite. General arbitrary-precision GDA suites do not become
fixed-width BID requirements merely because an operation name overlaps.

### Native and Portable Meanings

The native decNumber-oracle runner and the portable implementation runners have
different evidentiary meanings:

- native decNumber execution validates the IBM input/expected-value wiring for
  operations dispatched to decNumber, including exact comparison of the full
  parsed GDA condition set;
- operation-adapter cases execute the Go mechanical port against IBM expected
  values and compare the Intel BID five-flag surface; and
- portable Go and generated Rust legs independently cross-check the supported
  fixed-width operation set against IBM expected values on that same five-flag
  surface.

Reports must identify which executor ran. A decNumber-oracle result must not be
reported as direct Go-port verification.

### Operation Adapters

A decTest operation adapter is verification-only code. It may compose current
Go mechanical-port operations to express official decTest semantics, but it
must not:

- redefine a public API;
- call Intel C as the public implementation path;
- correct individual expected results; or
- grow into an independent reimplementation of the official operation.

Operation-family mappings and skip/exemption decisions belong in generator
code and generated tests. The current locations are indexed in
`VERIFICATION_REFERENCE.md`.

### Go-Port Comparison Strength

Executed operation-adapter cases in the generated native runner compare the
normalized decimal result value and the BID five-flag surface: invalid,
division-by-zero, overflow, underflow, and inexact. Their value comparator
collapses numerically equal finite cohorts; it does not independently establish
exact quantum/cohort identity.

The portable fixed-width Go and generated Rust legs compare:

- normalized result value;
- exact quantum/cohort identity for operations whose generated runner marks as
  cohort-preserving; and
- the same BID five-flag surface.

Expected decTest `Conditions` are fully parsed and then projected onto that
same five-flag surface. `Division_undefined`, `Division_impossible`,
`Insufficient_storage`, and `Conversion_syntax` map to invalid;
`Rounded`, `Clamped`, and `Subnormal` remain outside the compared BID mask
rather than being adjusted case by case. Actual case flags accumulate operand
parse flags and operation flags. An unrecognized condition token fails the
harness before the exemption classifier for every executed case. Operation-
scope skip classification occurs before execution and is accounted separately.

The generated native runner therefore has two explicit flag-check modes. A row
dispatched to decNumber uses the exact GDA-condition comparator. A row executed
by a Go mechanical-port operation adapter uses the BID five-flag projection;
the projection is not applied globally to native decNumber results. FMA,
`scaleB`, and remainder-family rows are not skipped merely because their
official conditions contain only `Rounded`, `Subnormal`, or `Clamped`.

The remaining `remainder_gda_division_impossible_context_semantics` and
`remaindernear_gda_division_impossible_context_semantics` classes are value
semantic divergences, not status gaps: the affected finite IBM GDA rows expect
a NaN under `Division_impossible`, while the pinned Intel remainder operations
produce a finite signed zero. These cases stay explicitly classified before
execution and counted in the generated inventory.

`toeng` is value-compared because engineering rendering can deliberately choose
a different cohort representation. A flag exemption is allowed only as a named
operation-family classification that is implemented and tested in generator
code, emitted in the generated inventory, counted outside generation, and applied
after all condition tokens have been validated. It never turns an unrecognized
condition into a skip.

The only permitted portable flag-exemption class is
`from_string_zero_low_clamp_divergence`. It applies only to `tosci`/`toeng`
rows with `Clamped`-only conditions and a zero expected result whose negative
exponent identifies a low-side operand-parse clamp. The exemption reflects the
measured decNumber-versus-Intel behavior: decNumber projects to no BID flag,
while the mechanically ported Intel parse raises inexact and underflow. It
waives only flag comparison; every otherwise-applicable value and quantum
assertion remains. High-side zero clamps and arithmetic-result zero clamps stay
flag-compared. Adding or widening an exemption class requires a specification
change, not only a generator or anchor update.

## C FFI Exact Bit-Compare Contract

The FFI profile is selected by `ffi_tests` in
`devtools/testgen_manifest.json`. The complete function list, patterns,
baseline case strength, and seed belong there rather than in this document.

For every selected case:

- the native Intel C symbol and the Go mechanical port receive identical raw
  operands and rounding mode;
- result bits compare exactly for decimal and binary results;
- scalar results compare exactly; and
- exception flags compare exactly whenever the C signature exposes flags.

The deterministic corpus must include format-correct special values,
canonical/non-canonical boundaries, directional operand combinations, and
deterministic pseudo-random coverage. A rounding-sensitive Tier 1 operation is
crossed with all five supported rounding modes at its semantic boundaries.
Every selected mixed-width FMA/sqrt extension is likewise crossed with all five
rounding modes on one unchanged operand tuple, and the native runner must prove
that the pinned Intel C result bits split into at least two distinct values;
merely enumerating five mode numbers is not rounding-mode coverage.

Every selected mixed-width FMA also carries a generated fusedness sentinel whose
expected direct result and forbidden sequential result are hand-pinned outside
the generation path. The native runner compares both Intel C and the Go port to
the direct expected bits and raw flags, then proves that a widened
`multiply -> add` sequence (plus destination narrowing for Decimal64 results)
produces the pinned forbidden result instead. Generated Rust public and direct
implementation paths execute the same sentinel contract. Where finite
single-rounding inputs cannot distinguish FMA from the sequential operations,
the sentinel must use an Intel-defined special-value ordering case that does;
no mixed FMA entrypoint is waived from this closed census.

The generated Tier 1 long runners strengthen this domain for both production
implementations. Go public APIs and the mechanical port, and generated Rust
public APIs and implementation, must run the same deterministic corpus against
pinned Intel C with exact bits and flags. The corpus must include all finite
`ScaleB` transition deltas in complete five-mode groups, signed 64-bit exponent
boundaries on guaranteed LP64 platforms, semantic remainder/fmod distinctions,
comparison/MinMax semantics, and all supported conversion shapes. Exact corpus
sizes and tuple identities are generator/anchor data, not prose.

Quiet comparison also requires a generator-owned semantic matrix whose expected
relations do not come from Intel C. For every supported BID width it covers
finite equality and ordering, signed-zero equality, infinity ordering, qNaN
unordered results, and sNaN unordered results with invalid raised. Both the Go
and Rust long runners execute this matrix. Its exact rows and count remain
generator and external-anchor data.

Optional shard variables must be paired and validated. Canonical full-verification
targets unset inherited shard variables before execution, and a shard result is
reported only as a shard.

String conversion is not duplicated into the numeric FFI profile. Its C-oracle
boundary is Intel readtest string data, where input text, output text, bits, and
status can be compared together.

## BID Codec Vector Contract

The public codec semantics and vector protocol are defined in
`BID_CODEC_SPEC.md`.

The regular vector domain must:

- generate vectors with an independent BID layout reference that does not
  import a production codec as its oracle;
- generate every required language harness from the same vector artifact;
- require standalone Go, Rust, Java, Python, JavaScript/TypeScript, and Swift
  consumers;
- treat the full `bid754-rs` consumer as additional evidence, not a replacement
  for standalone Rust;
- exercise decode, canonical encode, exact little-endian bytes, string
  rendering/parsing, reject channels, and successful string closure; and
- fail if any required consumer is absent or accepts a reject row through
  coercion, truncation, panic, or trap.

Decimal32 exhaustive raw and structured verification, and Decimal64/128
deterministic long verification, are generated strengthening gates for this
domain. Decimal32 may claim exhaustive raw coverage only after all `2^32`
patterns execute. Decimal64/128 must be described as deterministic long
coverage, never exhaustive. Full targets unset inherited shard configuration.
Exact totals and class partitions are owned by
`devtools/verification_anchors.json`.

## Auxiliary Gates

Generated public-API parity runners verify that public Go and Rust wrappers
route through their required implementation paths and preserve semantics. They
are architecture-contract gates, not a fifth regular verification domain.

Fuzz targets are auxiliary unless they have a generated case schema, generated
executor, authoritative oracle, exact result/status comparison, and external
anchors. Seed replay can be part of ordinary tests, but a mutation run or a
result-only differential fuzz path is not evidence that a regular domain is
closed.

## decNumber Differential Gate

An additional generated differential gate (like public-API parity, not a
fifth regular verification domain) compares three legs per case — pinned
Intel BID C, the Go mechanical port, and pinned IBM decNumber 3.68 — over a
batch-generated corpus. decNumber is a divergence tripwire, never a
correctness definition (SPEC.md Non-Goals): the adjudication authority for
any mismatch is IEEE 754-2019, and every divergence resolves through
exactly one of four paths — harness/mapping fix; a hand-audited
known-divergence regression row pinned in the manifest with a reason id
(decNumber-side defect or GDA-specific choice); the documented
IEEE-deviation procedure when pinned Intel C violates an IEEE `shall`
requirement; or a generation-time exclusion-class change when IEEE leaves
both behaviors legal.

Requirements:

- comparison is exact over the (class, sign, coefficient, exponent) triple
  — cohort-exact for finite results — plus the projected IEEE 5-flag word;
  the runner carries no tolerance, no runtime heuristic, and no runtime
  skip beyond replaying the generation-time class predicates it counts
  against externally anchored totals
- the decNumber leg is evaluated structurally (decNumberGetBCD + exponent +
  class predicates) in explicit width-fixed GDA contexts (digits/emax/emin
  per width, clamp=1, traps=0); operands transfer as exact
  integer-coefficient strings through a wide exact parse context, and a
  nonzero parse status is a hard failure, never a skip
- the decNumber status projection is the fixed table (Invalid_operation and
  Division_undefined to invalid; Division_by_zero, Overflow, Underflow,
  Inexact one-to-one; Rounded/Clamped/Subnormal dropped; every other bit is
  a hard failure); per-case adjustment functions are forbidden, and the
  decTest-domain file-convention flag heuristics must not be inherited
- corpus, independent BID triple codec, runner, stub, and the closed-world
  exclusion inventory are all testgen batch outputs; corpus sources are the
  filtered Tier 1 boundary reuse (canonical-only, NaN payload zero),
  normalized Tier 1 probes, a manifest-owned exact-product overflow class,
  and seeded deterministic random triples with stream hashes pinned in the
  runner constants and re-pinned by hand in devtools/verification_anchors.json
- exclusion classes are generation-time class predicates identified by
  reason id and counted in the generated inventory; adding or widening an
  exclusion class is a specification change requiring approval, exactly as
  for decTest flag exemptions
- known-divergence rows are hand-audited manifest entries; the runner keeps
  executing them on every leg and requires both sides to reproduce the
  pinned divergent results exactly, counts them against externally anchored
  totals, and a row matching no executed case fails generation as stale
- routing sentinels (hand-pinned in devtools/verification_sentinels.json,
  byte-equal to the generated runner literal) cover operand-slot,
  rounding-mode, and width/context-wiring skew that a common-mode glue bug
  would hide from the differential; the hand-written comparator strength
  anchor (bid754-go/decnumber_comparator_strength_test.go) pins the exact
  comparison semantics outside the generation path
- a PASS means "no divergence from an independent implementation inside the
  declared exact region", not IEEE conformance; excluded regions stay
  covered by the Intel-oracle domains only and must be reported that way

## Scope Classification

Generated inventories use explicit classifications with these meanings:

- `selected`: required by the claimed profile and dispatched;
- `out_of_scope_not_required`: outside the repository's supported BID scope;
- `optional_not_required` or `optional_scope_gap`: optional/recommended work,
  reported separately;
- `unresolved_required*`: required work not closed; any such entry prevents a
  completion claim.

Skip and exemption records carry both a stable reason identifier and a
classification. Unsupported operation names alone are not sufficient evidence.

## Reporting and Pass Criteria

A full regular-verification report states:

- the domain and exact execution boundary;
- the generated profile or suite identity;
- executed, skipped, exempted, and failed counts from generated output;
- whether the run was full, smoke, or sharded; and
- the oracle and comparison strength used.

Pass requires failed/mismatch count zero, required result and flag agreement,
matching external anchors, and no unresolved required inventory for the claimed
scope.

## Unsupported Patterns

- Do not implement missing production behavior inside tests.
- Do not add hand-written case lists, dispatch switches, or runners to a
  regular generated domain.
- Do not use manual regressions as evidence of regular-domain completion.
- Do not copy current function inventories or derived counts into normative
  prose.
- Do not describe a smoke, subset, or shard as full verification.
- Do not hide skipped, exempted, optional, or unresolved work behind a pass total.
