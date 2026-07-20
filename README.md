# bid754

[![CI](https://github.com/sky1core/bid754/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/sky1core/bid754/actions/workflows/build.yml)

BID-oriented IEEE 754 decimal floating-point work rooted in Intel BID C sources.

## Status at a Glance

- Usable today: the Go implementation module
  `github.com/sky1core/bid754/bid754-go` (fixed-width `Decimal32/64/128`)
  and the standalone Go codec module
  `github.com/sky1core/bid754/bid754-codec-go`; the Swift codec is
  consumable through Swift Package Manager. Release tags are not pushed
  yet: the package manifests carry version 0.2.0, and Go/SwiftPM consumers
  resolve the modules at the `main` branch (or a commit) until the first
  `bid754-go/v0.2.0`-style tags are published.
- Publish-pending: the Rust implementation `bid754-rs` now generates and
  verifies its full public API surface (fixed-width `Decimal32`/`Decimal64`/
  `Decimal128` value types, parity-gated wrapper methods/constructors,
  associated constants, exception-flag/rounding-mode/class/context types)
  through a dedicated go2rs `apiemit` pass, bit-compared against the
  generated port (public API parity gate) and cross-checked against IBM
  decTest (Rust decTest portable leg); the crate still carries
  `publish = false` and crates.io publication remains a separate
  user-approved step (`make verify-rust-package` is the package-shape gate
  ahead of that step).
- The six BID codec packages (Go, Rust, Java, Python, JavaScript/TypeScript,
  Swift) are verified in-tree against shared generated vectors; registry
  publication is currently configured only for the Go module and SwiftPM paths.
- Guaranteed bit-reproducible platforms: macOS arm64, Linux amd64, and
  Linux arm64 only. Windows amd64, 32-bit x86, and big-endian targets are
  explicitly not guaranteed (`docs/PLATFORM_SPEC.md`).
- Consumer quick check: `cd bid754-go && go test ./...` (portable, Go
  toolchain only). The full reproducible verification (`make verify-all`) is a
  maintainer gate with heavier multi-language and native prerequisites.

## Background

Decimal floating-point background, for readers new to the domain:

- [IEEE 754](https://en.wikipedia.org/wiki/IEEE_754) — the floating-point standard this library targets (2019 revision).
- [Decimal floating point](https://en.wikipedia.org/wiki/Decimal_floating_point) — the decimal formats and their two encodings, Binary Integer Decimal (BID, used here) and Densely Packed Decimal (DPD).
- Per-format references: [decimal32](https://en.wikipedia.org/wiki/Decimal32_floating-point_format), [decimal64](https://en.wikipedia.org/wiki/Decimal64_floating-point_format), [decimal128](https://en.wikipedia.org/wiki/Decimal128_floating-point_format).

## Read This First

This repository now separates goal docs from current-tree docs.

Authoritative goal/spec documents (under `docs/`):

- `docs/SPEC.md`
- `docs/ARCHITECTURE_SPEC.md`
- `docs/IEEE754_SPEC.md`
- `docs/PLATFORM_SPEC.md`
- `docs/BID_CODEC_SPEC.md`
- `docs/TEST_GENERATION_SPEC.md`
- `docs/DEPENDENCIES_SPEC.md`

Current verification implementation locations are indexed by the
non-normative `docs/VERIFICATION_REFERENCE.md`.

This `README.md` describes the current checked-out tree and developer workflow. It must not silently redefine the project goal.

Project goal and scope are defined in `docs/SPEC.md`.

## Repository Identity

This repository is the language-neutral `bid754` monorepo. The first-class
deliverables are the per-language bid754 libraries; the
`Intel BID C -> Go mechanical port -> generated Rust` chain is the
manufacturing methodology behind them, not a ranking among them.

The source repository URL and the Go module namespace prefix are the same
identity: `github.com/sky1core/bid754`. There is no root Go module; the two
public Go modules are `github.com/sky1core/bid754/bid754-go` (full
implementation) and `github.com/sky1core/bid754/bid754-codec-go` (standalone
codec). The module-root package of `bid754-go` is named `bid754`, so use a
named import:

```go
import bid754 "github.com/sky1core/bid754/bid754-go"
```

Release tags follow the Go multi-module convention: `bid754-go/v0.1.0` and
`bid754-codec-go/v0.1.0` version the Go modules, while root `v0.1.0`-style
tags version the repository snapshot for Swift Package Manager.

## License

Contributor-authored code is MIT licensed (`LICENSE`). The
`bid754-go/internal/bidgo/` mechanical
port and several generated artifacts are derivative works of the Intel Decimal
Floating-Point Math Library (BSD 3-Clause) and of IBM decTest data (ICU
License); the full third-party license texts and the exact derived-artifact
list live in `THIRD_PARTY_NOTICES.md`.

## Package Publishing Status

| Path | Status |
| --- | --- |
| `bid754-go/` | public Go implementation module (`github.com/sky1core/bid754/bid754-go`); the Go mechanical port lives inside it as `internal/bidgo/` |
| `bid754-codec-go/`, `bid754-codec-rs/`, `bid754-codec-java/`, `bid754-codec-py/`, `bid754-codec-js/`, `bid754-codec-swift/` | standalone BID codec packages intended for publication |
| `bid754-rs/` | public Rust implementation; the full generated public API surface is covered by externally anchored parity and Rust decTest gates, but the crate stays `publish = false` pending a separate user-approved crates.io publication step |
| `bid754-rs/libbid-sys/` | repo-internal FFI test bindings (`publish = false`) |
| `bid754-rs/ffi-verify/` | repo-internal FFI verification harness against the Intel BID C oracle (`publish = false`) |
| `devtools/` | non-published tooling module (generators, scripts, pinned inputs); never tagged or consumed as a dependency |
| `benchcompare-go/`, `benchcompare-rs/` | standalone comparative benchmark modules against shopspring/decimal and rust_decimal (cost comparison only, never a verification domain; the comparison dependencies are pinned here and never enter the product modules) |
| `bid754-go/internal/bidgo/cexport/` | inactive C ABI compatibility snapshot outside normal link inputs |

## Toolchain Prerequisites

| Workflow | Requires |
| --- | --- |
| `make test` (portable Go) | Go (per `bid754-go/go.mod` toolchain) |
| `make test-all` | + Rust stable/cargo, Java 17+, Python 3, Node.js + npm, Swift, ripgrep (`rg`); network on first run (npm/pip fetches) |
| `make verify-all` | + the native prerequisites below (or `VERIFY_ALL_ALLOW_MISSING_NATIVE=1`) |
| native gates (`make test-native-*`) | C toolchain (clang or gcc), `curl`, `unzip`, `shasum`, network for pinned downloads on first setup |
| `make verify-linux` | Docker (runs the Linux legs locally; no CI needed) |

## Current Tree State

Current verified workflows in this tree:

- portable default: `cd bid754-go && go test ./...`
- active checked-in language modules with portable test paths: `make test-all`
- active Go module vet checks: `make vet-go-modules`
- active Go module tidy/verify hygiene: `make verify-go-modules`
- full reproducible current-tree verification: `make verify-all`
- shell script syntax gate: `make check-scripts`
- Linux verification legs in local Docker (CI-independent): `make verify-linux`
- focused BID codec verification for the required Go, Rust, Java, Python, JavaScript/TypeScript, and Swift vector consumers: `make test-bidcodec`
- focused BID codec package verification for the six standalone language packages: `make verify-bidcodec-packages`
- focused `bid754-rs` publish-package shape verification (closed-set `cargo package --list` file check plus a `[dependencies]` pin/FFI-absence check; `cargo publish --dry-run` remains unavailable while `publish = false`): `make verify-rust-package`
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

- the repository contains optional native compatibility glue that depends on Intel BID plus local native prerequisites
- some native paths may still rely on IBM decNumber as a current implementation detail
- those native implementation details do not change the canonical source: Intel BID C
- table generation in this tree already reads Intel BID C inputs and emits both Go and Rust table artifacts
- the implementation split is different from the table split: Go uses the direct mechanical implementation path, and Rust is generated from that Go implementation path
- the public Go value-type runtime path routes through the Go mechanical port; the routing inventory and generated parity tests verify that path
- the generated Rust implementation includes every mapped symbol in the current declared surface; excluded surfaces remain outside the declared scope

## Portable Workflow

There is no root Go module; the portable default Go path runs inside the
`bid754-go/` module and does not require local C libraries:

```bash
cd bid754-go && go test ./...
```

Equivalent Make target (from the repository root):

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
make verify-all
```

`make verify-all` is the top-level reproducible verification gate; the authoritative
step list is the `_verify-all` target in the Makefile, documented in
`docs/BUILD.md`. The native gates are required by default — if `.env.sh`, Intel BID
`libbid.a`, or IBM decNumber are missing, `make verify-all` fails instead of
silently passing a reduced gate (`VERIFY_ALL_ALLOW_MISSING_NATIVE=1` skips
them explicitly). Compatibility entrypoints `devtools/run_tests.sh`,
`devtools/run_tests_and_benchmarks.sh`, and
`devtools/scripts/build_all.sh` delegate to this target.

To run the current benchmark boundary:

```bash
make bench
```

`make bench` runs Intel BID C direct benchmarks, `bid754-go` public Go API
native-tag benchmarks, Go mechanical-port (`internal/bidgo`) direct
benchmarks, and generated Rust
Criterion benchmarks. The fair cross-implementation matrix is
`bid32`/`bid64`/`bid128` across same-width arithmetic, remainder/fmod,
quantize/scaleB, a quiet comparison, MinNum/MaxNum, representative signed
integer conversions, all six BID-width conversions, parsing, and formatting.
It also includes all 24 Tier 1 mixed Decimal64/Decimal128
`add`/`sub`/`mul`/`div` variants. Intel C, the Go mechanical port, and
generated Rust use the same exact operand contract; public Go API benchmarks
are reported as an additional wrapper/API surface over the Go mechanical
port. The shared contract also requires `scale_exponent` to fit signed 32-bit
before the Intel C leg converts it to C `int`.
Intel C native benchmark runs require the pinned source-built `libbid.a` with
the dependency-spec build flags, including `CFLAGS_OPT=-O3 -ffp-contract=off`; setup scripts
record an ignored build-flag stamp and rebuild stale local libraries.

## Native Workflow

Prepare the native environment:

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

`devtools/scripts/verify_linux.sh` injects the working tree (tracked plus
untracked-but-not-ignored files) into a pinned
`ubuntu:24.04`-based image (Go pinned to the `bid754-go/go.mod` toolchain,
rustup
stable), reuses the pinned third-party archives cached under
`devtools/third_party/`
and `devtools/tests/` when present, and writes per-leg logs to
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

- `bid754-go/generated_types.go`
- `devtools/generated/go/intel_dfp_tables.go`
- `bid754-go/internal/bidgo/tables_binarydecimal.go`
- `bid754-rs/src/intel_dfp_tables.rs`
- `devtools/generated/json/intel_dfp_symbols.json`
- `devtools/generated/testspec/` (`spec_index.json` + `readtest/`, `ffi/` case shards)
- `bid754-codec-vectors/vectors.json`

Generated files are not edited directly. Change the manifest/generator and regenerate.
Some generated Go files intentionally remain at the `bid754-go/` module root because they are package `bid754` tests or public declarations; they carry `Code generated` headers rather than living under `devtools/generated/`. The generated spec loader package `bid754-go/internal/testspec/` is emitted by testgen as well.

Current artifact roles:

- `devtools/generated/go/intel_dfp_tables.go` and `bid754-rs/src/intel_dfp_tables.rs` are table artifacts generated from Intel BID C inputs; the same c-tablegen run owns the `bid_binarydecimal.c` subset at `bid754-go/internal/bidgo/tables_binarydecimal.go`
- `bid754-codec-vectors/vectors.json` is generated by `devtools/cmd/testgen` from `devtools/testgen_manifest.json` using an independent BID bit-layout reference codec as the cross-language vector source
- the required BID codec language consumers are `bid754-codec-go/`, `bid754-codec-rs/`, `bid754-codec-java/`, `bid754-codec-py/`, `bid754-codec-js/`, and `bid754-codec-swift/`
- `make test-bidcodec` verifies the generated vector artifact against all six required language consumers; `make verify-bidcodec-packages` additionally checks standalone package build/package/install/import boundaries and replays generated vectors from external consumers where the package artifact is installed or linked
- these table artifacts do not mean the whole Go implementation is generated from C
- the public Go value-type surface routes through the Go mechanical port; the routing inventory and generated parity tests verify that path
- the generated Rust implementation path is produced from the Go mechanical-port path; hand-maintained Rust support modules remain API/support plumbing rather than an alternate arithmetic source of truth
- `devtools/tools/go2rs` is the only permitted generator for the full Rust implementation artifacts under `bid754-rs/src/generated`; Rust idiom or performance improvements for that path must be implemented in `devtools/tools/go2rs` or its generated support/prelude rules and regenerated

## Testing and Verification

The authoritative testing direction lives in `docs/TEST_GENERATION_SPEC.md`.
The standalone codec API and shared vector protocol live in
`docs/BID_CODEC_SPEC.md`.

Important current-tree distinction:

- selection and generation parameters live in `devtools/testgen_manifest.json`
- exact counts and hashes live in `devtools/verification_anchors.json`
- current selected/excluded inventory and reasons live under `devtools/generated/testspec/`
- `docs/VERIFICATION_REFERENCE.md` maps each domain to its manifest, generator, generated inventory, runner, and gate
- `docs/BUILD.md` lists the current commands

The operative readtest profile is not all of Intel readtest. It is the
documented repository-supported surface, with `CMP_RELATIVEERR` functions
outside the explicit Tier 3 adoption list kept as profile expansion and
unsupported encodings/types kept outside the completion claim. Use generated inventory output for the exact current inventory.

If a workflow only covers a subset, documentation must call it a subset.

## ARM64 Note

For Intel DFP on ARM64, `BID_SIZE_LONG=8` is a compatibility fix to keep ARM64 on the intended 64-bit BID code path. It is not an alternate ARM-specific arithmetic design.
