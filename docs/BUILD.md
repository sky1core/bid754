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
targets must carry the anchored executed/total comparison counts, the
routing-sentinel full-count lines (row counts pinned in
`devtools/verification_sentinels.json`), and top-level PASS evidence
(including the Rust runners' exact passed-test counts). The native FFI and
readtest targets must additionally carry one exact compact summary whose
run/pass/skip totals match `devtools/verification_anchors.json`; decTest must
carry top-level PASS evidence. A zero-test or reduced-corpus run cannot report
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
`add`, `sub`, `mul`, `div`, `fma`, `sqrt`, `remainder`, `fmod`, `quantize`,
`scaleb`, `quiet_equal`, `minnum`, `maxnum`, `from_int64`, `to_int64`,
`parse`, and `to_string`, plus all six BID-width conversion directions. The
matrix also covers all 24 Tier 1 mixed Decimal64/Decimal128
`add`/`sub`/`mul`/`div` variants. Intel C, the Go mechanical port, and
generated Rust use the shared format-2 exact-operand contract (`x`, `y`, `z`,
`integer_operand`, and `scale_exponent`); `fma` consumes `z` as the addend,
`sqrt` reuses non-negative `x`, remainder/fmod use `y op x`, and mixed rows
use the source-width mapping documented beside their benchmark functions.
The contract rejects a `scale_exponent` outside signed 32-bit range before the
Intel C leg converts it to C `int`, so every layer receives the same exponent.
Public Go API benchmarks are reported as an additional wrapper/API surface.
`bench-native`, `bench-bidgo`, and `bench-rust` run those surfaces
individually.

`make verify-go-benchmark-registry` executes each registered Go subbenchmark
once and discards the timing output, then exact-compares normalized full names
against independent closed-world lists: 85 direct mechanical-port rows and
174 public-Go-plus-Intel-C rows. The portable half is part of
`make test-go-modules`; the native half is part of `make test-native-smoke`,
so the corresponding portable and native CI jobs enforce both actual Go test
binary registries. The one-iteration execution is only a registration and
wiring gate, never performance evidence. The halves can also be run directly
with `make verify-go-benchmark-registry-portable` and
`make verify-go-benchmark-registry-native` (the latter requires `.env.sh` and
the native dependencies).

`make verify-rust-benchmark-registry` asks the actual Criterion bench binary
to list its registry without measuring timings, then exact-compares the
closed-world set of 81 required group/name rows. This catches missing or
misnamed outer registrations that a shared row-macro test alone cannot see;
the gate is part of `make test-rust`, so the portable CI and `make verify-all`
paths both enforce it.

The registry gates check registered names only. The binding of each row name
to its operation is pinned by the hand-maintained cross-layer row descriptor
`bid754-go/testdata/benchmark_rows.json` (340 rows: 93 public Go API, 81
Intel C, 85 Go mechanical port, 81 generated Rust). Like
`devtools/verification_anchors.json` and
`devtools/verification_sentinels.json`, the descriptor stays outside every
generation path: no generator, template, manifest, or emitting script may
read or write it, so a wiring regression cannot re-pin itself.
Three gates consume it. `bid754-go/internal/benchrows` (which also carries
the Go-port row table itself so the module root can execute it)
closed-world-compares the Go-port tables against the descriptor and enforces
per-row sink/status discipline portably as part of `make test-go-modules`.
The native untimed preflight (`TestBenchmarkRow*` in
`bid754-go/benchmark_preflight_test.go`, part of `make test-native-smoke` and
`make test-native`) closed-world-compares the public-API and Intel C tables
(including the Intel C result-kind/flag metadata), then executes every
public-API and Go-port row exactly once per wiring fixture and exact-compares
the observed bits/flags against the executed Intel C benchmark leg as the
anchor — there is deliberately no parallel expected-value switch, so a wiring
mistake cannot skew the observed and expected legs the same way — and
requires the anchor observations to stay pairwise distinct per group across
fixtures. The Rust leg's `benchmark_contracts_match_shared_descriptor`
(`bid754-rs/ffi-verify/tests/benchmark_wiring.rs`, part of
`make test-rust-native`) exact-matches the declared Criterion row contracts
against the descriptor's rust layer, while its existing independent Intel BID
C oracle test keeps closing the Rust rows' observed results. Preflight
executions are untimed wiring evidence, never performance evidence.

Benchmark name to layer mapping (`make summary` groups by these):

- `BenchmarkIntelCBID*` and `BenchmarkIntelCMixedBID*`: Intel C called
  directly (the `b.N` loop runs inside C, so per-call cgo overhead is
  amortized)
- `BenchmarkAlignedBID*` and `BenchmarkAlignedMixedBID*`: public Go API. For
  `bid32` the value-only `add`/`sub`/`mul`/`div` rows measure the separate pure
  port bodies and the
  `*_with_flags` rows measure the status-aware bodies; compare
  `add_with_flags` with `FairBID32/add` (same implementation), and `add` with
  `FairBID32/add_pure` — the two bid32 row families are different
  implementations, not a wrapper-versus-port pair
- `BenchmarkFairBID*` and `BenchmarkFairMixedBID*`: Go mechanical port called
  directly (status-aware bodies; `bid32` also carries `*_pure` rows for the
  value-only bodies)
- Criterion `bid32/…`, `bid64/…`, `bid128/…`, and the two `*_mixed/…`
  groups: generated Rust implementation called directly.
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

The Go benchmark legs have a saved-baseline regression gate mirroring the
Criterion `pinned` baseline on the Rust leg. Workflow:

```bash
make bench-native && make bench-bidgo   # measure the reference state
make bench-go-baseline                  # save it as the explicit baseline
# ...make changes...
make bench-native && make bench-bidgo   # measure the candidate state
make bench-go-check                     # compare candidate vs baseline
```

`bench-go-baseline` copies the latest `bench-native`/`bench-bidgo` result
files to `test_results/bench_baseline_root.txt` and
`test_results/bench_baseline_bidgo.txt`; like `bench-rust-baseline`, it is
the only step that (over)writes the baseline — benchmark runs never update it
implicitly. `bench-go-check` runs `devtools/cmd/benchdiff`, which aggregates
the `BENCH_COUNT` repeated samples of each benchmark into a median ns/op and
compares candidate against baseline: a median regression above the threshold
(default 8%, tunable via `BENCH_REGRESSION_THRESHOLD`; the default clears the
±3–4% run-to-run noise measured on the Apple M1 reference machine) or a
benchmark that vanished from the candidate fails the gate, while
candidate-only benchmarks are reported as `new (no baseline)` without
failing. When benchmarks are added (they start as `new`), re-run
`make bench-go-baseline` — and `make bench-rust-baseline` for the Criterion
leg, whose strict `--baseline pinned` comparison fails outright on
benchmarks missing from an older pinned baseline — to fold them into the
saved baselines.

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
`bid754-go` tests/dispatch files, the c-tablegen-owned
`bid754-go/internal/bidgo/tables_binarydecimal.go`, the generated
`bid754-go/internal/testspec` spec loader package, BID codec vector consumers,
BID string vector consumers, Rust generated readtest runner, Rust readtest
dispatch inventory, and `bid754-rs/src/generated` after rerunning the
generators.

At the end, `make verify-generated` (also available standalone as
`make check-generated-markers`) runs
`devtools/scripts/check_generated_marker_coverage.sh`: every tracked or
untracked non-ignored file carrying a standard
`Code generated ... DO NOT EDIT.` marker must be part of the comparison set
above or listed with a documented reason in
`devtools/scripts/generated_marker_exceptions.txt`, so new generated artifacts
cannot silently stay outside reproducibility verification before staging.

Two further devtools test layers cover verification counts and table values.
`devtools/verification_anchors.json` pins the expected case counts of every
generated verification domain outside the generated path (a hand-edited file
checked against the real artifacts by a devtools test).
`devtools/internal/tablecrosscheck` compares the c-tablegen Go output against
the table literals inside `bid754-go/internal/bidgo` value by value. That
comparison is an independent value anchor for hand-ported tables; for the
c-tablegen-owned `tables_binarydecimal.go`, it is a closed-world value census,
while `make verify-generated` supplies byte reproducibility.

`devtools/verification_sentinels.json` adds two independent hand-maintained
pin families. The Tier 1 routing arrays bind the long runners' glue (operand
slots, rounding-mode wiring, dispatch-row labels) to expected results computed
at generation time through the public `bid754-go` API. `devtools` requires no
public module, so the routing-sentinel codegen reaches that API through the
pin-time oracle subprocess: it runs `go run ./internal/cmd/sentineloracle`
inside the sibling `bid754-go` module directory (a filesystem relationship,
not a module dependency) and receives each expected result over a line
protocol. Generation fails explicitly when the oracle is unavailable.

The `mixed_fma_fusedness_rows` array has a different source and update path.
Its direct expected and sequential forbidden bits/flags are audited against
the pinned Intel BID C implementation and recorded in
`devtools/internal/testgen/ffi_fusedness.go`; generated Go-native and Rust
runners consume that table. The external JSON array pins the resulting row
strings byte-for-byte. It is not produced by the public-Go sentinel oracle,
and `-print-sentinel-anchors` does not print it. No generator reads or writes
the external pin file.

Updating Tier 1 routing pins is a deliberate manual step:

1. `make generate-testspec` — the sentinel codegen re-selects the rows and
   self-asserts its sensitivity requirements (a selection that cannot
   distinguish an operand-slot swap, a rounding-mode pair, or a dispatch-row
   sibling fails the whole generation run).
2. `cd devtools && go run ./cmd/testgen -print-sentinel-anchors` — prints the
   two proposed Tier 1 routing arrays plus a per-row decimal interpretation.
   It writes no file and does not print mixed-FMA fusedness rows.
3. Audit the printed rows and paste them into
   `devtools/verification_sentinels.json` by hand.
4. `cd devtools && go test ./internal/testgen` — the anchor test requires the
   pinned rows to be byte-equal, in order, with the generated Go and Rust
   runner literals.
5. Because the runner bytes changed, hand-update the
   `goport_verification_runners` and `rust_tier1_long_runners` entries under
   `verification_artifact_sha256` in `devtools/verification_anchors.json` to
   the hashes the failing content-hash test prints, then re-run step 4.

Updating a mixed-FMA fusedness row instead requires a fresh pinned-Intel-C
direct-versus-sequential audit, a reviewed edit to both
`ffi_fusedness.go` and `verification_sentinels.json`, regeneration, the native
exact FFI gate, the generated Rust fusedness gate, and the same anchor/hash
checks. There is deliberately no auto-repin command for this family.

This friction is intended: a value-behavior change that moves any sentinel
answer must pass through a human re-audit, and the generator cannot re-pin its
own regression.

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
| `make verify-rust-benchmark-registry` | exact-set check of all 81 registered Rust Criterion rows (no timing measurement) |
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
