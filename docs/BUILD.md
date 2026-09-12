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
(`bid754-go/`, `bid754-codec-go/`, `devtools/`, and the comparative benchmark
module `benchcompare-go/`); the Make targets handle the module directories for
you.

This path is intentionally portable and does not require local C libraries.
It also does not require untracked authoritative generator input trees. Tests
that need those inputs skip explicitly when they are absent; use
`make verify-generated` when the goal is to require generator inputs and compare
freshly regenerated artifacts against the checked-in tree.

To verify every checked-in product language module with a portable test path:

```bash
make test-all
```

`make test-all` covers the product modules only. The two comparative
benchmark modules under `benchcompare-go/` and `benchcompare-rs/` are separate
benchmark tooling, not product modules, and are run from their own targets.

To run the current project-level verification boundary:

```bash
make verify-all
```

`make verify-all` prepares pinned generation inputs and executes the `full`
profile in `devtools/verification_plan.json`. This plan owns the required
portable/native gates and their execution evidence; `Makefile` owns individual
gate commands. CI and Docker select profiles from the same plan. To inspect a
profile, run `cd devtools && go run ./cmd/verifyplan list --profile full`.

Each invocation writes a fresh `test_results/verification/<profile>-<run>/`
directory containing `result.json` and a log per gate. The final adjudicator
requires every selected gate, anchored corpus counts, required consumers and
comparison strength, matching source/plan/anchor identities, and logs from that
invocation. `GOFLAGS` must be empty; each gate declares its own test selection
and native tags. Tool versions and relevant build configuration are recorded.
Local Git author/committer identity must be configured for temporary release
fixtures; CI and Docker supply an explicit verification identity.

Native prerequisites are required by default. `VERIFY_ALL_ALLOW_MISSING_NATIVE=1`
selects `full-portable`, explicitly reported as partial. It cannot produce a
`full` pass. Independent exhaustive gates remain separate profiles of work.
Compatibility entrypoints `devtools/run_tests.sh`, `devtools/run_tests_and_benchmarks.sh`, and
`devtools/scripts/build_all.sh` are thin wrappers around this target for test
verification.

The Linux verification legs run locally in Docker without a CI service:
`make verify-linux` (or the per-leg
`verify-linux-portable-arm64`/`verify-linux-portable-amd64`/`verify-linux-native-amd64`
targets). An additional local-only big-endian regression leg,
`make verify-linux-digest-s390x`, runs `make digest`, the
`bid754-codec-go` module tests, and the two generated Go-port runners
(readtest goport dispatch and public-API parity, full corpus) under qemu
linux/s390x, capturing `test_results/digest_linux_s390x.txt` for the
`make verify-digest` cross-platform comparison; it has no CI counterpart
and is not part of `make verify-linux`. See
`devtools/scripts/verify_linux.sh` for what each leg covers.

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

## Comparative Benchmarks Against External Libraries

`make bench-compare-go` and `make bench-compare-rs` measure `bid754-go`
against `shopspring/decimal` and `bid754-rs` against `rust_decimal` in the
standalone modules `benchcompare-go/` and `benchcompare-rs/`. These are cost
comparisons on a shared operand set, never result-equality checks, and the
comparison libraries are not product dependencies — each module's README
records the semantic differences and the isolation. They are separate
benchmark tooling: not a verification domain, outside `make verify-all`, and
outside the benchmark row descriptor and the registry gates. Each target runs
its module's operand-contract test before measuring.

`make bench-compare-rs` compares against the named Criterion baseline
`pinned`, saving it on first run, like `make bench-rust` and
`make bench-codec-rs`; `make bench-compare-rs-baseline` is the only target
that moves it. Criterion's anonymous default baseline is not used because it
accumulates unnamed results from other trees and older row names.

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
state that produced it. The Go and Rust legs also record the toolchain they
were built with: the Go legs stamp `go=$(go env GOVERSION)`, the Rust legs
stamp `rustc=<version>`, and the JS and Python codec legs stamp
`node=<version>` and `python=<version>` respectively.
A toolchain change redoes inlining and code-layout
decisions, so it moves medians on source that did not change at all —
`IntelCBID128/minnum` has been observed shifting 6.93 → 8.17 ns from layout
alone — which makes the toolchain part of a result's identity, not incidental
environment detail. The Go matrix targets repeat each benchmark `BENCH_COUNT`
times (default 5) for stable before/after samples; `bench-quick` is a
single-sample smoke and is not regression evidence.

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
failing.

The percentage threshold is paired with an absolute median-delta floor: a row
that clears the threshold but whose absolute median delta is at or below the
floor (default 0.25 ns/op, tunable via `BENCH_REGRESSION_MIN_DELTA_NS`; 0
disables the floor and restores the pure-percentage gate) does not fail and is
reported as `ok (below min delta)` — with its change% — rather than as a plain
`ok`. Both comparisons are strict, so a delta of exactly 0.25 ns is *not*
above the floor. The effective threshold of a row is therefore
`max(8%, 0.25 ns / baseline median)`. 8% of 3.125 ns is exactly 0.25 ns, so at
or above a 3.125 ns baseline median the percentage rule always binds first and
the gate's sensitivity is unchanged; only rows below that boundary are held to
the absolute floor. The motivating case is the inline-budget canary
`BenchmarkAlignedBID128/from_int64`, where a 0.1834 ns timer/inlining wobble
(0.6905 → 0.8739 ns/op) reads as +26.56% and used to break the gate on noise.
Both knobs are passed through by `make bench-go-check`.

The floor is deliberately global rather than scoped to sub-nanosecond rows. A
scoped variant (apply the floor only when the baseline median is under 1 ns)
was considered and rejected: it introduces a discontinuous step at an
arbitrary 1 ns line, where two rows measured 0.99 ns and 1.01 ns would be held
to incomparable standards. The global floor instead encodes one monotone
principle — a median delta at or under 0.25 ns is within a single row's
build-to-build reproducibility, regardless of which row produced it — and the
`max(8%, 0.25 ns / baseline)` formula it yields is continuous in the baseline
median.

The 0.25 ns figure is a reproducibility bound, not a sampling-noise bound: it
is not the harness's measurement resolution, which is an order of magnitude
finer (the 5-sample spread within one binary is 0.0013–0.086 ns across the
twelve floor-governed rows). What 0.25 ns covers is the wobble a *single* row
shows between builds — inlining decisions and code alignment shifting under
unrelated edits — which is the same mechanism as the `timer/inlining wobble`
on the canary row above, and which no amount of resampling one binary
removes.

Scope on the saved baselines (`test_results/` is git-ignored, so these are
local measurement artifacts, not committed ones): 12 of the 259 rows have a
median
under 3.125 ns and are therefore floor-governed rather than
percentage-governed. Two are sub-nanosecond (`FairBID128/from_int64` at
0.548 ns, `AlignedBID128/from_int64` at 0.877 ns) and ten sit in a 2.20–2.55 ns
band (the `scaleb`, `from_int64` and format-widening conversion rows). The
sub-nanosecond rows are the extreme case of the floor's reach, not the whole
of it.

Residual risk, and how to read a report:

- `ok (below min delta)` is **not** evidence that a row did not regress. It
  states only that the gate declined to render a verdict because the absolute
  move is within a single row's build-to-build reproducibility. Every such row
  is a human judgement call; the change% is printed precisely so it can be
  judged.
- Worst-case masking on the current baselines follows directly from the
  formula: the 0.548 ns row passes up to +45.6% and the 0.877 ns row up to
  +28.5%. A uniform +0.2 ns regression across the ten-row 2.20–2.55 ns band
  would also pass on every row, though for two different reasons: the eight
  rows from 2.201 to 2.328 ns clear 8% (+8.6% to +9.1%) and are held by the
  floor, while the two `scaleb` rows at 2.536 and 2.552 ns reach only +7.9%
  and +7.8% and so never engage the floor at all. A change expected to touch
  those rows should be judged on the printed change% or re-run with
  `BENCH_REGRESSION_MIN_DELTA_NS=0`.
- The floor does not blind the canary rows to what they were added to catch.
  The failure mode `AlignedBID128/from_int64` exists to detect — the
  inline-budget overflow fixed in `be6d0d1`, where losing inlining took the row
  from 0.668 ns to 4.16 ns — is a 3.49 ns delta, about 14× the floor, and
  fails the gate on the absolute rule alone.

Before comparing any row, `benchdiff` requires the two logs to describe
comparable runs. The `BENCH-META` `count=` and `go=` tokens and the
`goos`/`goarch` lines must be present in both logs and must match, and any
`cpu` line must match too (it may be absent, but then on both sides). Any
disagreement is an input error (exit 2), not a performance verdict — comparing
medians across sample counts, toolchains, or machines would launder an
environment change as a result. A `BENCH-META` line that omits a token records
`(none)` for it, so two logs that both predate a token still compare with each
other, while a log without the token never pairs silently with one that has it.

**Any baseline saved before the `go=` token existed will fail the first
`make bench-go-check` after this change with a `BENCH-META go` mismatch
(`baseline (none) vs candidate go1.26.5`).** That is the intended outcome, not
a bug: the baselines currently in `test_results/` were captured on 2026-07-20,
before the host Go toolchain was upgraded, so they were already stale in a way
the gate could not see. Clearing it requires re-measuring on an idle host and
re-saving:

```bash
make bench-native && make bench-bidgo   # on the current toolchain, idle host
make bench-go-baseline
```

The Rust legs stamp `rustc=` for provenance only. Criterion owns its own
baseline comparison, so nothing enforces a `rustc=` match; the token exists so
a Criterion `change%` can be attributed to a toolchain after the fact. The JS
and Python codec legs' `node=`/`python=` tokens are provenance only for the
same reason and one more: those legs have no saved-baseline comparison at all
and `benchdiff` never reads their logs, so nothing enforces a match — the
tokens exist so a later reader can tell whether a runtime upgrade sits between
two result files.

When benchmarks are added (they start as `new`), re-run
`make bench-go-baseline` — and `make bench-rust-baseline` for the Criterion
leg, whose strict `--baseline pinned` comparison fails outright on
benchmarks missing from an older pinned baseline — to fold them into the
saved baselines.

To benchmark the standalone BID codec packages (benchmark infrastructure
only, not a regular verification domain):

```bash
make bench-codec        # all four legs: Go, Rust, JS, Python
```

- `make bench-codec-go` — `bid754-codec-go` via stdlib `testing.B`
  (`BenchmarkCodecBID{32,64,128}/{decode,encode,to_string,from_string}`,
  repeated `BENCH_COUNT` times, results consumed through sink variables)
- `make bench-codec-rs` — `bid754-codec-rs` via Criterion
  (`benches/codec.rs`, dev-dependency only; change% reads against the named
  `pinned` baseline exactly like `bench-rust`, refreshed only by
  `make bench-codec-rs-baseline`)
- `make bench-codec-js` / `make bench-codec-py` — dependency-free scripts
  (`bid754-codec-js/bench_runner.mjs`, node built-in timing;
  `bid754-codec-py/benchmarks/bench_runner.py`, stdlib timing)
- all four legs load the same hand-pinned exact-operand contract
  `bid754-codec-go/testdata/codec_benchmark_operands.json`; each leg's setup
  rejects non-canonical or inexact operands, and
  `TestCodecBenchmarkOperandContract` (`bid754-codec-go/benchmark_test.go`)
  keeps that contract as a checked-in failing test

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

The Java leg additionally asks Gradle which publishing repositories it actually
resolved (`tasks --all`, one `publish…To<Repo>Repository` task per declared
repository) and fails unless `verification` is the only one, so a Maven Central
repository switched on by a resident deploy property is caught before anything
is built, signed, or uploaded. That leg also drops the
`ORG_GRADLE_PROJECT_bid754{MavenCentralDeploy,CentralPortal*,Signing*}`
variables from its own environment, so a release operator's shell cannot change
what the gate builds.

The JavaScript/TypeScript leg audits the tree `npm ci` just installed and fails
on any high/critical advisory for which npm reports a fix, since the build/test
toolchain's transitive dev tree can pick one up without a direct dependency
changing and the lockfile is where it gets resolved. The threshold is applied by
`devtools/scripts/lib/npm_audit_gate.mjs` rather than by `npm audit
--audit-level=…`, because that level is npm config an untracked `.npmrc` could
lower; a report npm could not produce, or one carrying a severity or a fix
status the gate cannot classify, fails closed instead of reading as clean, and a
high/critical advisory npm explicitly marks unfixable is printed as an
unresolved risk rather than wedging the gate on a release that does not exist
yet. A clean tree cannot show that the detector still works, so
`devtools/scripts/lib/npm_audit_gate_selftest.sh` drives it with fixed reports of
known verdict immediately before the live audit.

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
make test-native-decnumber-differential
```

Notes:

- current native smoke links Intel BID from `devtools/third_party/intel_dfp/lib`
- the `bid754_native` build tag works only inside a full repository checkout: its cgo paths reference `devtools/third_party/` by relative path, so a `bid754-go` module downloaded with `go get` cannot build native-tagged code
- `make test-native-ffi` is the non-short generated C FFI exact bit-compare gate
- `make test-native-readtest` is the non-short generated Intel readtest native gate
- `make test-native-dectest` is the non-short generated IBM decTest native gate
- `make test-native-decnumber-differential` is the generated decNumber
  third-oracle 3-leg differential gate (pinned Intel BID C / Go mechanical
  port / pinned IBM decNumber 3.68; requires the `bid754_decnumber_diff`
  build tag which the target sets itself); `verify-all-native-gates` runs it
  as `_test-native-decnumber-differential-full` with verifylog evidence
  binding
- `make test-native-d32-exhaustive` is the generated Decimal32 unary
  exhaustive differential gate (pinned Intel BID C / Go mechanical port,
  bit+flag exact over all 2^32 inputs for a fixed 20-lane unary table); it
  is an independent long gate outside the `verify-all` chain, and
  `_test-native-d32-exhaustive-full` is the canonical unsharded run with
  verifylog evidence binding (domain `d32-exhaustive`)
- `make test-rust-native-d32-exhaustive` is the generated Rust leg of that
  same gate (pinned Intel BID C / go2rs-generated Rust port, same 20-lane
  table); it is likewise an independent long gate outside the `verify-all`
  chain, and `_test-rust-native-d32-exhaustive-full` is its canonical
  unsharded run with verifylog evidence binding (domain
  `d32-exhaustive-rust`, bound to the same pinned per-lane digests as the
  Go leg)
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

`devtools/verification_sentinels.json` adds independent hand-maintained pin
families of two kinds. The routing-sentinel arrays — the Tier 1 long runners,
the decNumber differential runners, the D32 exhaustive runners, and the
mixed-format FFI rows — bind their runners' glue (operand slots, rounding-mode
wiring, dispatch-row labels, operation/mode lane selection) to expected results
computed at generation time through the public `bid754-go` API. `devtools`
requires no public module, so the routing-sentinel codegens reach that API
through the pin-time oracle subprocess: it runs `go run ./internal/cmd/sentineloracle`
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
| `make test-native-decnumber-differential` | generated decNumber third-oracle 3-leg differential gate, always full-scale |
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

`make explore-fresh-seed` runs `devtools/cmd/explorediff`, a fresh-seed
exploration fuzzer with the same discovery/audit standing as
`devtools/cmd/mutgate`: every run draws a new seed (pin one with `-seed`)
and differentially compares the Go mechanical port against pinned Intel BID
C with exact bits+flags over Tier 1 arithmetic
(add/sub/mul/div/fma/sqrt/quantize × three widths × five rounding modes),
with a `-bias` option that steers operands toward boundary pools and
exponent-interaction windows. It is not a verification domain, is never part
of `make verify-all`, and does not replace the pinned-seed gates; it exits 3
when it records counterexamples, writes the seed, config, and every mismatch
(operation, width, mode, operand bits, both results and flag words) as JSONL
under `test_results/`, and prints the exact reproduction command. Findings
enter the tree only through the existing manual procedures (regression
vectors, routing sentinels, corpus promotion). To observe the exit contract
(0 = clean, 3 = counterexamples recorded, 1 = run failure), invoke a built
binary directly — the Make target and the printed reproduction command
already do; `go run` folds every nonzero child exit into 1. It needs the
pinned Intel build (`make setup-native`); in a worktree,
`devtools/third_party/intel_dfp` already exists (it carries tracked files),
so symlink its `lib`, `src`, `include`, and `LIBRARY` subdirectories from
the primary checkout first.

`explorediff -campaign relations -ops add,sub,mul,div,fma,quantize`
adds an independent exact integer/rational model to every generated case.
`-cases` is the number of samples per family, width and rounding mode for
this campaign. `-campaign uniform-finite` uses separate finite
operand fields as a comparison campaign, with samples per operation, width
and mode. Both compare C/Go raw bits and flags separately from the model's
numeric value, sign, class and flags. Neither supports square root or nonfinite
inputs in the model. Counter records distinguish generated inputs, executed
operations and completed model comparisons, with numeric situation histograms.

The driver builds from a captured source snapshot and copies its native build
inputs into that isolated directory. Results record the snapshot, executable,
linked static library and pinned archive identities. Existing output files are
never overwritten. Input, model, execution, comparison or output errors fail the run.

Findings include original raw samples and, when successful, a reduced sample
that preserves the numeric situation and discrepancy. `-shrink-attempts 128`
sets the reduction budget; `0` explicitly disables it. Replay a JSONL file
containing only `sample` or `finding` records with `explorediff -replay FILE`;
operands are consumed as recorded, without seed regeneration. Extract finding
records from a run's mixed output before replay. A recorded mismatch still
requires adjudication against the specification; model agreement does not
define correctness.

`make test-finite-reference` checks relation predicates, the public-path
comparator's fault detection, semantic shrinking and official raw readtest
calibration. `make test-native-finite-reference` additionally calibrates against
pinned decNumber. Their bounded evidence belongs to the shared portable/native
profiles. To search beyond their seed inputs, run Go's coverage-guided target:

```bash
cd bid754-go
go test -run '^$' -fuzz '^FuzzFiniteArithmeticExact$' -fuzztime=60s .
```

`make test-finite-paths` runs the same raw finite operands through Go and Rust
public explicit-mode, default-value/default-flags, and direct port paths. All
five modes use each operand tuple; default APIs run under nearest-even only.
Results compare to the independent model including finite coefficient/exponent
and all observable flags. Missing or duplicate paths and coverage cells fail.
Each width/mode/status-aware path must observe all five IEEE flags across the
campaign, including nonzero division by both signs of zero.
The bounded portable profile runs 2 samples per relation and per uniform
operation, plus 20 signed zero/cohort/range-boundary tuples per operation at
each width. It does not extend the model to nonfinite inputs or
other operation families.

To expand the campaign with a recorded seed:

```bash
BID754_FINITE_PATHS='{"seed":20260911,"samples":32,"uniform":32}' make test-finite-paths
```

A failure records original/reduced raw samples and executable identities under
`test_results/finite-paths.*/finding.json`. Replay both samples with
`BID754_FINITE_PATHS='{"replay":"/path/to/finding.json"}' make test-finite-paths`.
The ordinary `FuzzFiniteArithmeticExact` target uses the strengthened Go path
comparisons. Rust participates in `test-finite-paths` and the BigDecimal fuzz
target below.

`make test-bigdecimal` adds Java `BigDecimal` to the same Go/Rust campaign.
It requires `java` and `javac` on `PATH` (JDK 17 or newer), compiles the generated
runner with `--release 17`, and records the actual JVM version and class hash.
The Java leg uses precision 7/16/34 and the selected rounding mode; FMA computes
the exact product and sum before one final rounding. Inputs transfer as integer
coefficients and exponents, without binary floating-point conversion.

Java independently compares numeric values on every public/port path for normal
or zero results. Its own calculation classifies division by zero, nonzero
results outside the BID normal exponent range, and quantize results exceeding
the width's precision as excluded, with separate reconciled counts. These are
not Java passes. Signed zero, finite cohorts and IEEE flags remain checked by
the existing finite model. JVM/protocol errors fail the run.

Use `BID754_FINITE_PATHS='{"seed":20260911,"samples":32,"uniform":32}' make test-bigdecimal`
for a larger deterministic campaign. A failure includes the Java result and
runtime identity alongside the original/reduced raw inputs. Replay Java findings
with `BID754_FINITE_PATHS='{"replay":"/path/to/finding.json"}' make test-bigdecimal`.

`make verify-numeric` runs the portable regression profile, including Go/Rust
public routing, readtest/decTest consumers and the finite reference checks,
followed by both BigDecimal fuzz targets and the remaining Tier 1 comparison
campaign. The verification runner records the source,
tool versions, per-gate logs and their hashes under `test_results/verification/`.
This profile is bounded numeric verification; native and exhaustive closure
remain separate profiles and targets.

`make fuzz-bigdecimal` runs `FuzzFiniteArithmeticBigDecimal` for 60 seconds with
two workers. Each input is checked through the actual Go and generated Rust
public/port paths against Java and the exact model. Inputs include relational,
uniform-finite and boundary lanes across all three widths and five modes.
Coverage feedback comes from Go; Rust and Java execute as comparison processes.
The numeric profile rejects seed-only replay, incomplete warmup, skipped targets
and runs with no executions beyond the initial corpus.

| Scope | Value reference | Representation and flags |
| --- | --- | --- |
| Finite add/sub/mul/div/quantize/FMA | BigDecimal for normal/zero numeric results; exact integer/rational model for every input | Exact model and Go/Rust public/port comparison |
| Remainder/fmod, scaleB, 12 quiet predicates, minNum/maxNum, integer and BID-width conversions | BigDecimal for eligible finite results; independent integer model for every input, including nonfinite encodings | Independent model checks class/sign, defined quantum and five flags; Go/Rust public and port paths |
| Other operations and nonfinite inputs to the six-operation finite model | Applicable generated readtest/decTest rows | Existing domain comparators and generated public routing checks |

`FuzzFiniteArithmeticBigDecimal` retains the six-operation scope and Java
exclusions of its seeded campaign. A failure reports raw operands,
expected/actual results and executable identities. The semantic shrinker preserves the discrepancy and numeric
situation, recording original/reduced raw samples in
`test_results/bigdecimal-fuzz.*/finding-*.json`. Replay both with
`BID754_FINITE_PATHS='{"replay":"/path/to/finding.json"}' make test-bigdecimal`.
Go separately saves the discovery input under
`bid754-go/testdata/fuzz/FuzzFiniteArithmeticBigDecimal/`; replay requires the
same Java and Rust setup, so a missing comparison process cannot silently turn a
replay into a Go-only pass.

Use `bash devtools/scripts/fuzz_bigdecimal.sh --duration 5m --parallel 2` for a
longer search, or `bash devtools/scripts/fuzz_bigdecimal.sh --replay <corpus-id>`
to rebuild the comparison processes and execute one saved input. Initial corpus
entries can also be selected with `--replay 'seed#N'`. A missing or unexecuted
entry fails replay.

`make test-tier1-bigdecimal` exercises the additional Tier 1 families in the
coverage table. Three PCG seeds (754, 2019, 57025) produce 14,040 cases over
585 width/operation/mode/target cells. Integer inputs cover signed/unsigned
32/64-bit types; integer outputs cover 8/16/32/64-bit types, including variants
that signal inexact. All six BID-width conversions execute. Every cell must
have a completed Java numeric comparison. Nonfinite inputs, invalid integer
results, zero divisors and results outside Java's supported BID region are
counted separately; the independent model still checks them. Equal min/max
operands may select either canonical input; a third cohort or a different
combination of zero sign and exponent is rejected. NaN results are
checked for class, quietness, canonical form and flags, without claiming an
independent payload-selection oracle.

This deterministic gate is required by the portable, CI portable and full
profiles. The two BigDecimal fuzz targets remain optional discovery gates.

The same gate adds 66,270 systematic integer-construction and narrowing cases
across 75 cells. Midpoint neighbors, coefficient parity, decimal carries,
integer limits and exponent boundaries are compared through Go and Rust public
and port paths. Java independently checks 59,550 numeric results; 6,720
exponent-range cases remain checked by the exact model. The native conversion
gates include the same boundary families in their Intel C bit/flag comparisons.

`make test-native-tier1-reference` independently calibrates the Tier 1 model
and Go public/port paths against pinned IBM decNumber 3.68. The native and
full profiles require its 14,040 cases, covering 585 width/operation/mode/target
cells with seeds 10754, 12019 and 67025. Values, signed zeros, defined quantum
and IEEE flags are compared; quiet comparisons and integer conversions use
decNumber comparison and integral-rounding primitives. Width conversion and
integer construction use multiplication by exact one in the destination
context, preserving zero sign and operand quantum before rounding.
The bounded calibration corpus keeps remainder quotients within decNumber's
precision and scaleB shifts within its supported range. It has no runtime
exclusions and does not claim NaN payload/sign-selection validation.

The deterministic Tier 1 gate also requires 1,596 law comparisons across the
Go public/port paths and reference model for remainder signs, scaleB, exact
width round trips, quiet comparisons and special-value flags. Harness checks
inject five source faults into a temporary copy of the reference model and
require the corresponding witness assertions to fail.
Production fault injection also restores intermediate-width rounding in signed
and unsigned integer construction and Decimal128-to-32 narrowing. Parser
witnesses restore fixed exponent caps in all three widths. A crash does not
count as a numeric assertion failure.

`make fuzz-tier1-bigdecimal` searches these families for 60 seconds with two
workers. Every byte input maps to a valid case; mutations include raw BID
patterns, finite values, integer boundaries, midpoint cases, zeros, NaNs and
extreme scaleB exponents. Go supplies coverage guidance and minimizes failing
byte inputs. Raw findings retain tool/source identities and all observations
under `test_results/tier1-bigdecimal.*/tier1-finding-*.json`.
The systematic boundary corpus contributes 1,250 representative fuzz seeds;
integer and raw BID bits mutate beyond those seeds. The full boundary corpus
remains in the deterministic gate so seed replay leaves time for exploration.

```bash
bash devtools/scripts/test_tier1_bigdecimal.sh --fuzz 5m --parallel 2
bash devtools/scripts/test_tier1_bigdecimal.sh --replay /path/to/tier1-finding.json
bash devtools/scripts/test_tier1_bigdecimal.sh --corpus '<corpus-id|seed#N>'
```

The Tier 1 corpus lives under `bid754-go/testdata/fuzz/FuzzTier1BigDecimal/`.
Replay rebuilds Java and generated Rust; missing or unexecuted corpus entries
fail. These auxiliary campaigns do not replace the six-language BID codec
vector domain or establish full IEEE conformance.

The fuzz input exposes family, width, mode, coefficient entropy, exponent and
sign separately. A failed case is saved by Go's fuzz runner and replayed by
ordinary `go test`; inspect its raw case diagnostic before promotion.

`mutgate -stages exactprobe` evaluates real mechanical-port mutations with
the same model and public Go API. Select explicit `-exact-seeds`, then either
`-exact-cases N` or `-exact-cases 0 -exact-cpu-budget 250ms`. CPU accounting is
process user plus system time for generation, model, arithmetic and comparison,
including GC; build/startup time is separate. The budget is checked between
inputs, so the report includes measured CPU and any overshoot. Each campaign
and seed runs independently, even after another campaign detects the mutation.

`-mode exactcheck -stages exactprobe -exact-probes calibration` checks three
specific kernel faults and their precise wrong-answer witnesses. The `heldout`
set uses different mutation sites and requires `-exact-tuning-seeds` to reject
seed overlap with tuning. These small sets evaluate specified fault classes;
they do not establish a general mutation score for the whole library. Findings
outside the pristine run's tested input prefix remain inconclusive.

`mutgate -mode pathcheck -stages finitepaths,rustfinite` tests the selected
`-exact-probes` set against each language separately using the finite path
executor and a declared operation-specific witness. The pristine baseline runs
the complete bounded campaign. For Rust, it mutates the Go predecessor and regenerates through
go2rs before building the real crate; generated implementation files are never
hand-patched. The Go predecessor is restored and Rust regenerated after each
probe. Only recorded discrepancies in the specified width, operation and fault count
as detections; build errors,
missing paths, crashes and timeouts remain distinct. The `calibration` and
`widths` sets together exercise the three widths. `-exact-probes quantum`
selects three zero-result quantize exponent faults that preserve numeric value
and flags; these require the strengthened cohort comparison.

Mutation runs require an isolated detached `-worktree` at `-commit` and a
`-jsonl` result path. For uncommitted changes, `-mode snapshot` captures tracked
edits and explicitly listed `-snapshot-files` into a temporary commit without
moving a branch or changing the main index. The ordinary verification profiles
run the bounded harness checks; CPU campaigns and kernel mutation audits are
opt-in.

## ARM64 Intel BID

Keep the ARM64 `BID_SIZE_LONG=8` override explicit when required by the pinned upstream. This preserves the intended 64-bit BID build behavior; it is not an alternate arithmetic implementation.
