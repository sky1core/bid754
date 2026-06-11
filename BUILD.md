# Build Guide

This document describes how to build and verify the current checked-out tree. It does not redefine the project goal; see `SPEC.md`, `ARCHITECTURE_SPEC.md`, `IEEE754_SPEC.md`, `PLATFORM_SPEC.md`, `TEST_GENERATION_SPEC.md`, and `DEPENDENCIES_SPEC.md`.

## Portable Default

The default root Go path is:

```bash
go test ./...
```

Equivalent Make target:

```bash
make test
```

This path is intentionally portable and does not require local C libraries.
It also does not require untracked authoritative generator input trees. Tests
that need those inputs skip explicitly when they are absent; use
`make verify-generated` when the goal is to require generator inputs and compare
freshly regenerated artifacts against the checked-in tree.

To verify every checked-in language module with a portable test path:

```bash
make test-all
```

To run the current project-level verification boundary:

```bash
make full-audit
```

`make full-audit` is the top-level reproducible audit target. It runs the
shell script syntax gate, the active portable Go module tests, vet checks,
Go module tidy/verify hygiene, the bid-go/bidcodec zero-dependency contract,
the portable cgo-purity contract, generated Rust tests, generated-artifact
reproducibility, the dependency vulnerability scan (requires `osv-scanner`),
the `bid-go/cexport` quarantine guard audit, the six-language standalone BID
codec package audit and vector consumers, BID string vector verification, the
generated Rust overflow policy audit, and the native smoke/generated
FFI/generated readtest/generated decTest/Rust native gates. The native gates
are required by default: when `.env.sh`, Intel BID `libbid.a`, or IBM
decNumber are missing, the target fails; set
`FULL_AUDIT_ALLOW_MISSING_NATIVE=1` to skip the native gates explicitly.
Legacy `run_tests.sh`, `run_tests_and_benchmarks.sh`, and
`scripts/build_all.sh` are thin wrappers around this target for test
verification.

The Linux verification legs run locally in Docker without a CI service:
`make verify-linux` (or the per-leg
`verify-linux-portable-arm64`/`verify-linux-portable-amd64`/`verify-linux-native-amd64`
targets). See `scripts/verify_linux.sh` for what each leg covers.

To run the current benchmark boundary:

```bash
make bench
```

This runs Intel BID C direct benchmarks, root public Go API with the native
tag, `bid-go` direct mechanical-port calls, and generated Rust Criterion
benches. The fair cross-implementation matrix is `bid32`/`bid64`/`bid128` by
`add`, `mul`, `div`, `parse`, and `to_string` for Intel C, `bid-go`, and
generated Rust. Root public Go API benchmarks are reported as an additional
wrapper/API surface. `bench-native`, `bench-bid-go`, and `bench-rust` run those
surfaces individually.

To verify the generated BID codec vector consumers for the required Go, Rust,
Java, Python, JavaScript/TypeScript, and Swift targets:

```bash
make test-bidcodec
```

To audit the six standalone BID codec packages beyond repo-level vector
consumption, including generated vector replay from external package
consumers:

```bash
make audit-bidcodec-packages
```

To verify Intel readtest-derived string conversion vectors for the current
mandatory implementation consumers. This is the canonical C-oracle boundary for
BID string conversion, separate from the numeric native FFI bit-compare profile:

```bash
make test-bid-string
```

To audit the generated Rust overflow policy:

```bash
make audit-rust-overflow
```

## Native Smoke Path

Prepare the environment:

```bash
make doctor
bash ./scripts/install_ibm_decnumber.sh
./scripts/setup_c_libs.sh
```

Then run:

```bash
source .env.sh
make test-native-smoke
make test-native-ffi
make test-native-readtest
make test-native-dectest
```

Notes:

- current native smoke links Intel BID from `third_party/intel_dfp/lib`
- `make test-native-ffi` is the non-short generated C FFI exact bit-compare gate
- `make test-native-readtest` is the non-short generated Intel readtest native gate
- `make test-native-dectest` is the non-short generated IBM decTest native gate
- some current-tree native paths may also require IBM decNumber
- that requirement is a current implementation detail, not the source-of-truth architecture

## Generators

Prepare authoritative generator inputs first:

```bash
make setup-generation-inputs
```

Regenerate checked-in artifacts with:

```bash
make generate-types
make generate-tables
make generate-symbols
make generate-testspec
```

To verify reproducibility instead of merely running the portable checked-in
artifact tests:

```bash
make verify-generated
```

`make verify-generated` snapshots and compares the checked-in generated Go
root tests/dispatch files, BID codec vector consumers, BID string vector
consumers, Rust generated readtest runner, Rust readtest dispatch audit, and
`bid754-rs/src/generated` after rerunning the generators.

Generated files are reproducible artifacts. Do not edit them directly.
`make generate-testspec` also regenerates the checked-in BID codec vector data at `bid-codec-vectors/vectors.json`.
It also regenerates the repo-level BID codec vector consumer harnesses for Go,
standalone Rust, Rust full-library, Java, Python, JavaScript/TypeScript, and
Swift, plus the BID string vector consumers for the Go mechanical port and the
generated Rust implementation.

## Verification Scope

Be explicit about scope:

- portable test is the default repo-safety path
- `make full-audit` is the current top-level project verification boundary
- `make test-bidcodec` is the repo-level generated verification gate for the required six BID codec language consumers
- `make audit-bidcodec-packages` is the stronger standalone package-quality gate for those six BID codec helper packages, including generated vector replay through package-consumer boundaries
- `make test-bid-string` currently targets the Go mechanical port and generated Rust implementation string conversion consumers, not the standalone BID codec helper packages
- BID string conversion is intentionally verified by readtest-derived string vectors rather than by the numeric native FFI bit-compare subset
- `make audit-rust-overflow` documents and enforces the current generated Rust policy: the checked-in Rust implementation must not disable Cargo profile overflow checks, and it must pass both the default Rust test profile and `RUSTFLAGS='-C overflow-checks=yes'`
- `make test-native-ffi` runs `TestGeneratedFFIBitCompareSubset` without `-short` and is the native generated C FFI exact bit-compare gate
- `make test-native-readtest` runs `TestGeneratedReadCases` without `-short` and is the native generated Intel readtest gate
- `make test-native-dectest` runs `TestGeneratedDectestSuites` without `-short` and is the native generated decTest CI gate
- native smoke is a narrower native verification path
- `generated/testspec/` (`spec_index.json` plus the `readtest/` and `ffi/` case shards) is generated from the verification manifests; for Intel `readtest.in` the generator derives the active checked-in BID readtest subset mechanically from `readtest.h`, `readtest.in`, the repository's discoverable BID constructors/methods, the documented historical scope rule (`CMP_FUZZYSTATUS - explicit historical skip 함수군 + CMP_EQUALSTATUS`), and the current spec-phase exclusion list
- the checked-in Intel readtest subset is source-driven and includes a generated `readtest.h` function audit; the current counts live in `TEST_GENERATION_SPEC.md` and the generated audit artifacts
- this closes the current supported-surface readtest required gap; do not describe it as Intel readtest 전체 because `CMP_RELATIVEERR` and out-of-scope binary/DPD/reverse conversion functions remain outside the operative profile
- the exact readtest case count and profile mix can change when upstream Intel inputs or the repository's currently wired BID surface changes
- decTest suites are selected mechanically from official `tests/*.decTest` inputs by scanning file operations and keeping only files whose non-ignored operations fit the current checked-in supported operation sets; the file counts live in `TEST_GENERATION_SPEC.md`
- the generated native FFI exact bit-compare subset compares return value and `_IDEC_flags` where exposed; the covered function groups and counts live in `TEST_GENERATION_SPEC.md`
- `bid754-rs` `ffi-fuzz` is a randomized Rust-vs-Intel-C complement for selected arithmetic functions; it compares both result bits and `_IDEC_flags`, but it is not the generated regular FFI profile
- Go `FuzzGeneratedArithmeticResultOnlyNative` is a native-only result-string fuzz complement from generated decTest seeds; it does not compare decTest status or IEEE flags and is not a regular generated verification domain
- do not describe that subset as full readtest/decTest/FFI coverage; FFI bit-compare still excludes out-of-scope reverse binary-to-BID, binary80, DPD, FE, mixed-width Intel extension, and string-conversion groups
- the portable path and native smoke remain narrower safety paths; non-short generated FFI/readtest/decTest coverage belongs to `make test-native-ffi`, `make test-native-readtest`, and `make test-native-dectest`
- general precision >34 remains out of scope for the current checked-in subset

Full verification policy is defined in `TEST_GENERATION_SPEC.md`.

## ARM64 Intel BID

Keep the ARM64 `BID_SIZE_LONG=8` override explicit when required by the pinned upstream. This preserves the intended 64-bit BID build behavior; it is not an alternate arithmetic implementation.
