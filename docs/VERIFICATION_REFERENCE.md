# Verification Implementation Reference

This is a non-normative map of the current verification implementation. It
helps locate the machine-readable source for a question; it does not override
`SPEC.md`, `TEST_GENERATION_SPEC.md`, or `BID_CODEC_SPEC.md`.

Current counts and complete inventories are intentionally absent. Completion
reports measure them from generated inventories and external anchors.

## Source Map

| Question | Current source |
| --- | --- |
| Which profiles, suites, functions, seeds, and baseline parameters are selected? | `devtools/testgen_manifest.json` |
| What does the readtest selection label mean? | `devtools/internal/testgen/readtest_spec.go` |
| How are decTest files classified? | `devtools/internal/testgen/spec_build.go` |
| How are decTest runtime skips and flag exemptions classified? | `devtools/internal/testgen/dectest_skip_reason.go` and its tests |
| How are FFI cases and wrappers generated? | `devtools/internal/testgen/ffi_spec.go` and `devtools/internal/testgen/ffi_test_codegen.go` |
| How are Tier 1 long corpora generated? | `devtools/internal/testgen/tier1_arithmetic_long_codegen.go` and `devtools/internal/testgen/tier1_compare_conversion_long_codegen.go` |
| How are BID codec vectors defined? | `devtools/internal/testgen/bid_codec_reference.go` and the `devtools/internal/testgen/bid_codec_*vectors*.go` family |
| What is selected or excluded in the checked-in tree? | `devtools/generated/testspec/spec_index.json` and sibling dispatch inventories |
| What exact counts and hashes must match? | `devtools/verification_anchors.json` |
| Which generated files are reproducibility-checked? | `Makefile` `verify-generated` and `devtools/scripts/check_generated_marker_coverage.sh` |
| Which commands run each gate? | `BUILD.md`, the root `Makefile`, and `devtools/scripts/` |

## Regular Domain Map

### Intel readtest

Inputs:

- pinned Intel `TESTS/readtest.h`
- pinned Intel `TESTS/readtest.in`
- manifest profile `intel_readtest_current_surface`

Selection:

- the manifest requests `repo_supported_surface`;
- `readtest_spec.go` classifies every upstream function row; and
- `spec_index.json.readtest_profile_inventory` records the resulting selected and
  excluded inventory with reasons.

The current profile follows the normative formula in
`TEST_GENERATION_SPEC.md`. Its exclusion classes cover unsupported binary
formats, DPD interchange, FE APIs, version predicates satisfied on another
surface, and other Intel-only groups outside the declared support surface. The
Tier 1 mixed-width D/Q `add`/`sub`/`mul`/`div` families and the pinned Intel
mixed-width FMA/sqrt extension families are selected. The generated inventory,
not a hand-written function list, answers whether a particular symbol is
currently selected.

Generated consumers:

- native Go test runner against Intel C;
- direct Go mechanical-port runner;
- generated Rust readtest runner; and
- generated Go/Rust dispatch inventories and cross-checks.

Intel string rows are also the canonical C-oracle input for BID string
conversion. String conversion is not duplicated into the numeric FFI profile.

### IBM decTest

Inputs and suite selection are declared by `dectest_suites` in the manifest.
`spec_index.json.dectest_file_inventories` records every official file, its operation
set, whether it was selected, and why it was excluded.

Current executor families:

| Family | Current implementation location |
| --- | --- |
| native decNumber oracle operations | generated native executor |
| Go mechanical-port operation adapters | generated `bid754-go/dectest_*.go` files |
| portable Go fixed-width cross-check | generated Go-port dispatch and runner |
| portable Rust fixed-width cross-check | generated Rust decTest runner |

Adapter semantics belong in templates and generator tests. In particular,
GDA operation names such as `plus`, `minus`, `abs`, `remainder`,
`remaindernear`, `comparetotal`, and `scaleb` are not inferred from similarly
named Intel helpers. The adapter templates define the mapping, while
`dectest_skip_reason.go` defines operation-family skip and flag-exemption
classifications.

The current runtime accounting is emitted in:

- `dectest_runtime_skip_inventory`;
- `dectest_goport_runtime_skip_inventory`; and
- `dectest_rust_dispatch_inventory.json`.

Deferred General/GDA, tagged-literal/DPD, logical-digit, optional math, and
unsupported fixed-width operation buckets are current inventory state. A
`unresolved_required` classification is the only deferred class that prevents a
completion claim for the selected mandatory scope.

### C FFI exact bit-compare

The current profile name is `bid_native_bitcompare_subset`. Its explicit
function list, function patterns, baseline strength, and seed are under
`ffi_tests` in `devtools/testgen_manifest.json`.

`devtools/generated/testspec/ffi/` contains the current generated case shards.
The generated native wrapper calls pinned Intel C and compares the same inputs
against the Go mechanical port. The profile includes decimal results, scalar
predicates/classes, integer conversions, BID width conversions, and supported
one-way BID-to-binary conversions according to the manifest.

The selected mixed-width FMA/sqrt functions additionally carry one unchanged
operand tuple across all five rounding modes, with a runtime assertion that
Intel C produces more than one result bit pattern. Mixed-width FMA functions
also carry fusedness sentinels: direct expected and sequential forbidden
bits/flags are externally pinned in `devtools/verification_sentinels.json` and
executed by the generated Go-native and Rust public/direct paths.

The Tier 1 long Go and Rust runners are generated from shared corpus rules:

- arithmetic, fused multiply-add, square root, quantize, remainder/fmod, and
  ScaleB;
- quiet comparisons and MinNum/MaxNum/MinNumMag/MaxNumMag; and
- integer/BID, BID-width, and one-way BID-to-binary conversions.

Their exact operation census, structured/random counts, finite-transition
limits, complete rounding-mode groups, and tuple hashes are pinned in
`devtools/verification_anchors.json`. The generated source and anchors are the
reference for those values.

### BID codec vectors

The API and protocol contract is `BID_CODEC_SPEC.md`.

Current implementation sources:

- generation seed and random sample parameter: `bid_codec_vectors` in
  `devtools/testgen_manifest.json`;
- independent layout oracle: `bid_codec_reference.go`;
- successful raw vectors: `bid_codec_vectors_data.go`;
- reject channels: `bid_codec_reject_vectors.go`;
- successful string closure: `bid_codec_string_vectors.go`;
- fixed anchor records: `bid_codec_vector_anchors.go`; and
- output artifact: `bid754-codec-vectors/vectors.json`.

Generated harness templates under `devtools/internal/testgen/bidcodec_templates/` own
the per-language vector consumers. Exact record totals, class partitions,
long-corpus strengths, and consumer accounting are external-anchor data.

## Additional Gates

### Generated public-API parity

Go and Rust public parity runners verify API routing and constant/value parity.
They are generated architecture gates rather than a regular verification
domain. Their wrapper census and case totals are externally anchored.

### Long codec coverage

Decimal32 exhaustive verification traverses all raw 32-bit patterns and a
generated structured components corpus. Decimal64/128 long verification uses
deterministic raw samples and structured boundaries; it is not exhaustive.
Current scale and class totals are maintained only in the external anchors and
generated runner constants.

### Auxiliary fuzzing

The repository contains native differential and portable no-panic/closure fuzz
targets. Their implementation and seed locations are indexed by the manifest
and `BUILD.md`. They remain auxiliary because mutation execution is not a
generated regular domain with externally anchored result-and-status coverage.

### Platform digest

The platform digest checks a deterministic portable corpus across guaranteed
OS/architecture pairs. Its platform policy is `PLATFORM_SPEC.md`; commands and
result-file handling are in `BUILD.md`, `Makefile`, and
`devtools/scripts/verify_digest.sh`.

## Maintenance Routing

When current state changes, update the owning machine source:

- selection or seed: manifest;
- selection semantics or skip reason: generator code plus tests;
- generated inventory: regenerate;
- expected scale or identity: external anchors after reviewing the generated
  change;
- execution graph: Makefile or script; and
- normative behavior: the applicable specification.

Current function lists, numeric snapshots, and skip tables remain in generated
inventories rather than this map.
