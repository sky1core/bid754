# Build Guide

This document describes how to build and verify the current checked-out tree. It does not redefine the project goal; see `SPEC.md`, `ARCHITECTURE_SPEC.md`, `IEEE754_SPEC.md`, `PLATFORM_SPEC.md`, `BID_CODEC_SPEC.md`, `TEST_GENERATION_SPEC.md`, and `DEPENDENCIES_SPEC.md`.

## Portable Default

There is no root Go module. The portable default Go path runs inside the
`bid754-go/` module:

```bash
cd bid754-go && go test ./...
```

Equivalent Make target (from the repository root):

```bash
make test
```

Direct `go` commands must run inside one of the checked-in Go modules
(`bid754-go/`, `bid754-codec-go/`, `devtools/`); the Make targets handle the
module directories for you.

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
make verify-all
```

`make verify-all` is the top-level reproducible verification target. It runs the
shell script syntax gate, the active portable Go module tests, vet checks,
Go module tidy/verify hygiene, the
bid754-go/bid754-codec-go/devtools zero-dependency contract,
the portable cgo-purity contract, generated Rust tests, generated-artifact
reproducibility,
verification that the inactive `bid754-go/internal/bidgo/cexport` module remains outside normal link inputs, the
package manifest version agreement verification, the six-language standalone BID
codec package verification and vector consumers, the Decimal64/128 BID codec
long differential verification, the `bid754-rs` publish-package
shape verification, BID string vector verification, the
generated Rust overflow policy verification, and the native gates: native
smoke, generated FFI bit-compare, the Go and Rust Tier 1 arithmetic and
compare/conversion long differentials, generated readtest, generated decTest,
Rust native readtest, and the Rust ffi-fuzz auxiliary. Gate logs are
additionally re-checked by `devtools/cmd/verifylog`: the canonical-full Tier 1
targets must carry the anchored executed/total comparison counts plus
top-level PASS evidence, and the FFI, readtest, and decTest targets must carry
top-level PASS evidence — a zero-test or reduced-corpus run cannot report
green. The native gates
are required by default: when `.env.sh`, Intel BID `libbid.a`, or IBM
decNumber are missing, the target fails; set
`VERIFY_ALL_ALLOW_MISSING_NATIVE=1` to skip the native gates explicitly.
Compatibility entrypoints `devtools/run_tests.sh`, `devtools/run_tests_and_benchmarks.sh`, and
`devtools/scripts/build_all.sh` are thin wrappers around this target for test
verification.

The Linux verification legs run locally in Docker without a CI service:
`make verify-linux` (or the per-leg
`verify-linux-portable-arm64`/`verify-linux-portable-amd64`/`verify-linux-native-amd64`
targets). See `devtools/scripts/verify_linux.sh` for what each leg covers.

To run the current benchmark boundary:

```bash
make bench
```

This runs Intel BID C direct benchmarks, the `bid754-go` public Go API with
the native tag, direct Go mechanical-port (`bid754-go/internal/bidgo`) calls,
and generated Rust Criterion
benches. The fair cross-implementation matrix is `bid32`/`bid64`/`bid128` by
`add`, `mul`, `div`, `parse`, and `to_string` for Intel C, the Go mechanical
port, and
generated Rust. Public Go API benchmarks are reported as an additional
wrapper/API surface. `bench-native`, `bench-bidgo`, and `bench-rust` run those
surfaces individually.

Benchmark name to layer mapping (`make summary` groups by these):

- `BenchmarkIntelCBID*`: Intel C called directly (the `b.N` loop runs inside
  C, so per-call cgo overhead is amortized)
- `BenchmarkAlignedBID*`: public Go API. For `bid32` the value-only
  `add`/`mul`/`div` rows measure the separate pure port bodies and the
  `*_with_flags` rows measure the status-aware bodies; compare
  `add_with_flags` with `FairBID32/add` (same implementation), and `add` with
  `FairBID32/add_pure` — the two bid32 row families are different
  implementations, not a wrapper-versus-port pair
- `BenchmarkFairBID*`: Go mechanical port called directly (status-aware
  bodies; `bid32` also carries `*_pure` rows for the value-only bodies)
- Criterion `bid32/…`, `bid64/…`, `bid128/…`: generated Rust public API.
  `make bench-rust` reports change percentages against the saved `pinned`
  Criterion baseline only (first run creates it); refresh the baseline
  deliberately with `make bench-rust-baseline` — never read change% from
  back-to-back unnamed runs

Every benchmark target stamps a `BENCH-META` line (tree id, date, and — on
the Go targets — the repetition count; the Rust targets record the Criterion
baseline name instead, since Criterion does its own sampling) into its
`test_results/` output so a result file is attributable to the exact source
state that produced it. The Go matrix targets repeat each benchmark
`BENCH_COUNT` times (default 5) for stable before/after samples;
`bench-quick` is a single-sample smoke and is not regression evidence.

To verify the generated BID codec vector consumers for the required Go, Rust,
Java, Python, JavaScript/TypeScript, and Swift targets:

```bash
make test-bidcodec
```

To verify the six standalone BID codec packages beyond repo-level vector
consumption, including generated vector replay from external package
consumers:

```bash
make verify-bidcodec-packages
```

To verify the `bid754-rs` publish-package shape: a closed-set `cargo package
--list` file check (only `src/**`, `LICENSE`, `NOTICE`, `README.md`,
`Cargo.toml`, plus Cargo's own bookkeeping files may ship) and a
`[dependencies]` pin/FFI-absence check. `cargo publish --dry-run` itself is
not part of this gate: `publish = false` blocks it today, and lifting that is
a separate user-approved step:

```bash
make verify-rust-package
```

To verify Intel readtest-derived string conversion vectors for the current
mandatory implementation consumers. This is the canonical C-oracle boundary for
BID string conversion, separate from the numeric native FFI bit-compare profile:

```bash
make test-bid-string
```

To verify the generated Rust overflow policy:

```bash
make verify-rust-overflow
```

## Native Smoke Path

Prepare the environment:

```bash
make doctor
bash ./devtools/scripts/install_ibm_decnumber.sh
./devtools/scripts/setup_c_libs.sh
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

- current native smoke links Intel BID from `devtools/third_party/intel_dfp/lib`
- the `bid754_native` build tag works only inside a full repository checkout: its cgo paths reference `devtools/third_party/` by relative path, so a `bid754-go` module downloaded with `go get` cannot build native-tagged code
- `make test-native-ffi` is the non-short generated C FFI exact bit-compare gate
- `make test-native-readtest` is the non-short generated Intel readtest native gate
- `make test-native-dectest` is the non-short generated IBM decTest native gate
- some current-tree native paths may also require IBM decNumber
- that requirement is a current implementation detail, not the source-of-truth architecture

## Generators

Generators and extraction tools live in the `devtools/` module and run with
`devtools/` as their working directory; the Make targets below handle the
`cd devtools` step.

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

`make verify-generated` snapshots and compares the checked-in generated
`bid754-go` tests/dispatch files, the generated `bid754-go/internal/testspec`
spec loader package, BID codec vector consumers, BID string vector
consumers, Rust generated readtest runner, Rust readtest dispatch inventory, and
`bid754-rs/src/generated` after rerunning the generators.

At the end, `make verify-generated` (also available standalone as
`make check-generated-markers`) runs
`devtools/scripts/check_generated_marker_coverage.sh`: every tracked file
carrying a standard `Code generated ... DO NOT EDIT.` marker must be part of
the comparison set above or listed with a documented reason in
`devtools/scripts/generated_marker_exceptions.txt`, so new generated artifacts
cannot silently stay outside reproducibility verification.

Two further devtools test layers anchor the verification counts and table
values outside the generated path: `devtools/verification_anchors.json` pins
the expected case counts of every generated verification domain (a hand-edited
file checked against the real artifacts by a devtools test), and
`devtools/internal/tablecrosscheck` compares the c-tablegen Go output against
the hand-ported table literals inside `bid754-go/internal/bidgo` value by
value.

Generated files are reproducible artifacts. Do not edit them directly.
`make generate-testspec` also regenerates the checked-in BID codec vector data at `bid754-codec-vectors/vectors.json`.
It also regenerates the repo-level BID codec vector consumer harnesses for Go,
standalone Rust, Rust full-library, Java, Python, JavaScript/TypeScript, and
Swift, plus the BID string vector consumers for the Go mechanical port and the
generated Rust implementation.

## Verification Scope

Use the target name as the execution boundary and report whether the run was
portable, smoke, full, or sharded:

| Target | Current execution boundary |
| --- | --- |
| `make verify-all` | top-level reproducible project verification |
| `make verify-generated` | regenerate and byte-compare all declared generated artifacts |
| `make test-native-readtest` | native Intel readtest runner, non-short |
| `make test-portable-readtest` | direct Go mechanical-port readtest runner |
| `make test-native-dectest` | native generated decTest runner, non-short |
| `make test-portable-dectest` | portable Go fixed-width decTest runner |
| `make test-native-ffi` | native generated C FFI exact bit-compare runner, non-short |
| `make test-bidcodec` | six-language generated BID codec vector consumers |
| `make verify-bidcodec-packages` | standalone codec package-boundary verification |
| `make test-bid-string` | readtest-derived Go/Rust BID string verification |
| `make verify-rust-overflow` | generated Rust overflow-policy verification |

`TEST_GENERATION_SPEC.md` defines verification policy and comparison strength.
`BID_CODEC_SPEC.md` defines the codec contract. Current selected/excluded
inventory, counts, and hashes are read from the generated inventories and
`devtools/verification_anchors.json`; `VERIFICATION_REFERENCE.md` maps each
question to its machine-readable source.

Portable tests and native smoke are narrower safety paths. They must not be
reported as a full regular-domain run.

## Manual Fuzzing

Auxiliary fuzz targets complement the generated verification domains; they are
exploration tools, never a substitute for the generated gates, and there is
deliberately no Make target for them (a time-boxed fuzz run is not a smoke
check and must not be wired into a pass/fail pipeline as one). Their seed
corpora — including the committed regression corpus under each module's
`testdata/fuzz/<FuzzName>/` — replay automatically as ordinary test cases in
every plain `go test ./...` run (so `make test-portable` / `make
test-go-modules` already re-execute the seeds); the `-fuzz` mutation mode is
manual only.

Portable targets (no native prerequisite):

```bash
cd bid754-codec-go
go test -run xxx -fuzz '^FuzzFromStringRoundTrip$'   -fuzztime 60s .
go test -run xxx -fuzz '^FuzzDecodeToStringReparse$' -fuzztime 60s .

cd bid754-go
go test -run xxx -fuzz '^FuzzParseNoPanic$' -fuzztime 60s .
```

Native differential target (requires `source .env.sh` and the native build
prerequisites):

```bash
cd bid754-go
CGO_ENABLED=1 go test -tags bid754_native -run xxx \
  -fuzz '^FuzzArithmeticPortVsNativeResultOnlyNative$' -fuzztime 300s .
```

When a `-fuzz` run finds a failing input, Go writes it to
`testdata/fuzz/<FuzzName>/` in the module — from then on every `go test` run
replays it. Triage it before moving on: a genuine divergence becomes a
committed regression corpus entry (that is how
`bid754-codec-go/testdata/fuzz/FuzzFromStringRoundTrip/` got its
exponent-closure entry); a harness artifact gets a narrow documented gate in
the fuzz body and the crash file is removed (use `trash`, not `rm`). Never
leave an untriaged crash file behind, and never widen a gate just to silence
a finding.

## ARM64 Intel BID

Keep the ARM64 `BID_SIZE_LONG=8` override explicit when required by the pinned upstream. This preserves the intended 64-bit BID build behavior; it is not an alternate arithmetic implementation.
