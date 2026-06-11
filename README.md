# bid754

BID-oriented IEEE 754 decimal floating-point work rooted in Intel BID C sources.

## Read This First

This repository now separates goal docs from current-tree docs.

Authoritative goal/spec documents:

- `SPEC.md`
- `ARCHITECTURE_SPEC.md`
- `IEEE754_SPEC.md`
- `PLATFORM_SPEC.md`
- `TEST_GENERATION_SPEC.md`
- `DEPENDENCIES_SPEC.md`

This `README.md` describes the current checked-out tree and developer workflow. It must not silently redefine the project goal.

Project goal and scope are defined in `SPEC.md`.

## Repository Identity

The source repository URL and the Go import namespace are the same identity:
`github.com/sky1core/bid754`. The root module carries the public API and the
`bid-go/` mechanical-port package; `bidcodec/` is a standalone module in the
same repository.

## License

Contributor-authored code is MIT licensed (`LICENSE`). The `bid-go/` mechanical
port and several generated artifacts are derivative works of the Intel Decimal
Floating-Point Math Library (BSD 3-Clause) and of IBM decTest data (ICU
License); the full third-party license texts and the exact derived-artifact
list live in `THIRD_PARTY_NOTICES.md`.

## Package Publishing Status

| Path | Status |
| --- | --- |
| `bidcodec/`, `bid-codec-rs/`, `bid-codec-java/`, `bid-codec-py/`, `bid-codec-js/`, `bid-codec-swift/` | standalone BID codec packages intended for publication |
| root Go module (`github.com/sky1core/bid754`) | public Go API surface (includes the `bid-go/` package) |
| `bid-go/` | Go mechanical-port package inside the root module; not a separate module |
| `bid754-rs/` | repo-internal generated verification crate (`publish = false`); not a published Rust API |
| `bid754-rs/libbid-sys/` | repo-internal FFI test bindings (`publish = false`) |
| `bid-go/cexport/` | quarantined legacy stubs guarded against linking |

## Toolchain Prerequisites

| Workflow | Requires |
| --- | --- |
| `make test` (portable Go) | Go (per `go.mod` toolchain) |
| `make test-all` | + Rust stable/cargo, Java 17+, Python 3, Node.js + npm, Swift, ripgrep (`rg`); network on first run (npm/pip fetches) |
| `make full-audit` | + `osv-scanner`, and the native prerequisites below (or `FULL_AUDIT_ALLOW_MISSING_NATIVE=1`) |
| native gates (`make test-native-*`) | C toolchain (clang or gcc), `curl`, `unzip`, `shasum`, network for pinned downloads on first setup |
| `make verify-linux` | Docker (runs the Linux legs locally; no CI needed) |

## Current Tree State

Current verified workflows in this tree:

- portable default: `go test ./...`
- active checked-in language modules with portable test paths: `make test-all`
- active Go module vet checks: `make vet-go-modules`
- active Go module tidy/verify hygiene: `make audit-go-modules`
- full reproducible current-tree audit: `make full-audit`
- shell script syntax gate: `make check-scripts`
- Linux verification legs in local Docker (CI-independent): `make verify-linux`
- focused BID codec verification for the required Go, Rust, Java, Python, JavaScript/TypeScript, and Swift vector consumers: `make test-bidcodec`
- focused BID codec package audit for the six standalone language packages: `make audit-bidcodec-packages`
- focused BID string<->bits verification for the Go mechanical port and generated Rust implementation consumers, using Intel readtest-derived string cases as the canonical C oracle: `make test-bid-string`
- native smoke: `make test-native-smoke` after preparing `.env.sh`
- generated FFI bit-compare native non-short gate: `make test-native-ffi` after preparing `.env.sh`
- generated Intel readtest native non-short gate: `make test-native-readtest` after preparing `.env.sh`
- generated IBM decTest native non-short gate: `make test-native-dectest` after preparing `.env.sh`
- generator input setup: `make setup-generation-inputs`
- generators:
  - `make generate-types`
  - `make generate-tables`
  - `make generate-symbols`
  - `make generate-testspec`

Current tree notes:

- the repository still contains some legacy/native glue that depends on Intel BID plus local native prerequisites
- some native paths may still rely on IBM decNumber as a current implementation detail
- that current-tree detail does not change the long-term goal: Intel BID C is the canonical source of truth
- table generation in this tree already reads Intel BID C inputs and emits both Go and Rust table artifacts
- the intended implementation split is different from the table split: Go is the direct mechanical implementation path, while Rust is intended to be generated from the Go implementation path
- the public Go value-type runtime path is expected to converge on that Go mechanical port rather than on direct Intel BID C runtime calls or fake non-native stubs
- this tree now contains a generated Rust implementation path for the current in-scope surface and its checked-in verification workflows, but excluded or future-phase surfaces are still outside that path

## Portable Workflow

The default root Go path is portable and does not require local C libraries:

```bash
go test ./...
```

Equivalent Make target:

```bash
make test
```

If authoritative generator input trees have not been prepared locally,
generator-input-dependent reproducibility tests skip with an explicit
`make setup-generation-inputs` / `make verify-generated` message. The portable
path still tests the checked-in generated artifacts; it is not the full
generator reproducibility gate.

To verify every checked-in language module that has a portable test path:

```bash
make test-all
```

To run the current project-level verification boundary before claiming the
tree is clean:

```bash
make full-audit
```

`make full-audit` is the top-level reproducible audit gate; the authoritative
step list is the `_full-audit` target in the Makefile, documented in
`BUILD.md`. The native gates are required by default — if `.env.sh`, Intel BID
`libbid.a`, or IBM decNumber are missing, `make full-audit` fails instead of
silently passing a reduced gate (`FULL_AUDIT_ALLOW_MISSING_NATIVE=1` skips
them explicitly). Legacy `run_tests.sh`, `run_tests_and_benchmarks.sh`, and
`scripts/build_all.sh` delegate to this target.

To run the current benchmark boundary:

```bash
make bench
```

`make bench` runs Intel BID C direct benchmarks, root public Go API native-tag
benchmarks, `bid-go` mechanical-port direct benchmarks, and generated Rust
Criterion benchmarks. The fair cross-implementation matrix is
`bid32`/`bid64`/`bid128` across `add`, `mul`, `div`, `parse`, and `to_string`
for Intel C, `bid-go`, and generated Rust. Root public Go API benchmarks are
reported as an additional wrapper/API surface over the Go mechanical port.
Intel C native benchmark runs require the pinned source-built `libbid.a` with
the dependency-spec build flags, including `CFLAGS_OPT=-O3 -ffp-contract=off`; setup scripts
record an ignored build-flag stamp and rebuild stale local libraries.

## Native Workflow

Prepare the native environment:

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

The native path is a current-tree verification workflow. It should not be described as the architectural source of truth.

## Linux Verification Without CI

The Linux verification legs run locally in Docker, so they do not depend on a
CI service:

```bash
make verify-linux                  # all three legs
make verify-linux-portable-arm64   # linux/arm64: Go modules + Rust portable
make verify-linux-portable-amd64   # linux/amd64: Go modules + Rust portable
make verify-linux-native-amd64     # linux/amd64: Intel BID C oracle native gates
```

`scripts/verify_linux.sh` injects the working tree (tracked plus
untracked-but-not-ignored files) into a pinned
`ubuntu:24.04`-based image (Go pinned to the `go.mod` toolchain, rustup
stable), reuses the pinned third-party archives cached under `third_party/`
and `tests/` when present, and writes per-leg logs to
`test_results/latest_linux_<leg>_results.txt`. The native leg builds IBM
decNumber and Intel BID inside the container and runs the same
smoke/FFI/readtest/decTest/Rust-native gates as the macOS native workflow.

## Generated Artifacts

Prepare authoritative generator inputs before regenerating artifacts:

```bash
make setup-generation-inputs
```

To enforce that checked-in generated artifacts still reproduce from those
inputs, run:

```bash
make verify-generated
```

Representative checked-in generated artifacts (the authoritative full set is
the `verify-generated` recipe in the Makefile):

- `generated_types.go`
- `generated/go/intel_dfp_tables.go`
- `generated/rust/intel_dfp_tables.rs`
- `generated/json/intel_dfp_symbols.json`
- `generated/testspec/` (`spec_index.json` + `readtest/`, `ffi/` case shards)
- `bid-codec-vectors/vectors.json`

Generated files are not edited directly. Change the manifest/generator and regenerate.
Some generated Go files intentionally remain in the repository root because they are package `bid754` tests or public root-package declarations; they carry `Code generated` headers rather than living under `generated/`.

Current artifact roles:

- `generated/go/intel_dfp_tables.go` and `generated/rust/intel_dfp_tables.rs` are table artifacts generated from Intel BID C inputs
- `bid-codec-vectors/vectors.json` is generated by `cmd/testgen` from `testgen_manifest.json` using an independent BID bit-layout reference codec as the cross-language vector source
- the required BID codec language consumers are `bidcodec/`, `bid-codec-rs/`, `bid-codec-java/`, `bid-codec-py/`, `bid-codec-js/`, and `bid-codec-swift/`
- `make test-bidcodec` verifies the generated vector artifact against all six required language consumers; `make audit-bidcodec-packages` additionally checks standalone package build/package/install/import boundaries and replays generated vectors from external consumers where the package artifact is installed or linked
- these table artifacts do not mean the whole Go implementation is generated from C
- the intended Go runtime path still means the public Go value-type surface should ride the Go mechanical port rather than direct C runtime glue or fake stubs
- the generated Rust implementation path is produced from the Go mechanical-port path; hand-maintained Rust support modules remain API/support plumbing rather than an alternate arithmetic source of truth
- `tools/go2rs` is the only permitted generator for the full Rust implementation artifacts under `bid754-rs/src/generated`; Rust idiom or performance improvements for that path must be implemented in `tools/go2rs` or its generated support/prelude rules and regenerated

## Testing and Verification

The authoritative testing direction lives in `TEST_GENERATION_SPEC.md`.

Important current-tree distinction:

- `generated/testspec/` (`spec_index.json` plus the `readtest/` and `ffi/` case shards) is generated from the verification manifests; for Intel `readtest.in` the generator derives the active checked-in BID readtest subset mechanically from `readtest.h`, `readtest.in`, the repository's discoverable BID methods/constructors, the documented historical scope rule (`CMP_FUZZYSTATUS - explicit historical skip 함수군 + CMP_EQUALSTATUS`), and the current spec-phase exclusion list
- the checked-in Intel readtest subset is source-driven and includes a generated `readtest.h` function audit; the current selected/excluded function counts live in `TEST_GENERATION_SPEC.md` and the generated audit artifacts
- this closes the current supported-surface readtest required gap; it is still not "Intel readtest 전체" because non-`fmod` `CMP_RELATIVEERR` math/transcendental groups and out-of-scope binary/DPD/reverse conversion functions remain outside the operative profile
- the Rust generated readtest dispatch audit dispatches every selected function with 0 skips, plus the duplicate Intel `CMP_RELATIVEERR` comparator rows for `bid32/64/128_fmod`; counts live in `TEST_GENERATION_SPEC.md`
- the exact readtest case count can change when upstream Intel inputs or the repository's currently wired BID surface changes
- decTest suites are selected mechanically from official `tests/*.decTest` inputs by scanning each file's operations and keeping only files whose non-ignored operations stay within the current checked-in supported operation sets; the selected/remaining file counts live in `TEST_GENERATION_SPEC.md`
- the generated native FFI exact bit-compare subset compares result and `_IDEC_flags` where the Intel symbol exposes flags and cycles rounding modes `0..4` for `_IDEC_round` symbols; the covered function groups and counts live in `TEST_GENERATION_SPEC.md`
- that is useful subset verification
- it is not the same thing as full readtest/decTest/FFI verification; FFI bit-compare still excludes out-of-scope reverse binary-to-BID, binary80, DPD, FE, mixed-width Intel extension, and string-conversion groups
- BID string conversion is not counted as an FFI bit-compare gap; it is covered by the generated readtest-derived `make test-bid-string` boundary because that domain compares parsed bits, status, and normalized text rather than a simple decimal-return C symbol
- current decTest coverage remains a subset; native and portable paths still skip unsupported subset edges such as general precision >34, tagged-literal `tointegralx` clamp cases, and some arithmetic `Clamped`/`Division_undefined` flag cases

If a workflow only covers a subset, documentation must call it a subset.

## ARM64 Note

For Intel DFP on ARM64, `BID_SIZE_LONG=8` is a compatibility fix to keep ARM64 on the intended 64-bit BID code path. It is not an alternate ARM-specific arithmetic design.
