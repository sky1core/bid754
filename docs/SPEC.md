# Project Spec

`SPEC.md` is the first-class project definition for this repository.

If documents conflict, use this precedence:

1. `SPEC.md`
2. `ARCHITECTURE_SPEC.md`
3. `IEEE754_SPEC.md`
4. `PLATFORM_SPEC.md`
5. `BID_CODEC_SPEC.md`
6. `TEST_GENERATION_SPEC.md`
7. `DEPENDENCIES_SPEC.md`

Detailed documents must refine this file, not redefine it.

## Goal

This repository exists to implement the required IEEE 754 decimal behavior for the BID-based scope it claims to support.

Project-level commitments:

- Intel BID C is the canonical upstream source of truth
- the supported arithmetic operation families are bounded by the pinned Intel
  BID C BID-decimal surface; the project does not invent additional public
  arithmetic families beyond that upstream surface
- BID is the project encoding model
- `Decimal32`, `Decimal64`, and `Decimal128` remain fixed-width value types (the public Go types are `Decimal32BID`, `Decimal64BID`, and `Decimal128BID`; `types_layout_check.go` pins the 4/8/16-byte layout at compile time)
- extracted/generated artifacts must come from C or other authoritative official inputs
- table artifacts for Go and Rust are extracted/generated from the Intel BID C sources
- the Go implementation path is a direct mechanical port of the Intel BID C implementation rather than a generated translation product
- the Go mechanical port may carry explicitly documented IEEE-conformance deviations from pinned Intel BID C only where the pinned C implementation conflicts with IEEE 754-2019 `shall` behavior; every such deviation must be listed in `IEEE754_SPEC.md`, must keep its native C comparison rows skipped with an accurate reason in the generation manifest, and must be pinned by checked-in regression vectors
- the Rust implementation path is generated from the Go implementation path rather than directly from the Intel BID C implementation
- `devtools/tools/go2rs` is the only permitted generator for the full Rust implementation artifacts under `bid754-rs/src/generated`; alternate Rust implementation generators, direct C-to-Rust implementation generators, and hand-written replacement implementation paths are not acceptable
- Rust implementation quality, idiom, or performance improvements must be made in `devtools/tools/go2rs` or its generated support/prelude rules and then regenerated, not by editing generated Rust artifacts or adding a parallel hand-maintained Rust implementation
- optimization in both implementation paths must preserve a mechanically auditable correspondence to the canonical predecessor: pinned Intel BID C for the Go mechanical port, subject only to the registered IEEE-conformance deviations above, and the Go mechanical port for generated Rust; mechanically auditable means an independent reviewer can identify the corresponding predecessor function or source region and a chain consisting only of an explicitly registered IEEE-deviation step when applicable followed by a finite sequence of local semantics-preserving transformations, not merely observe equal outputs
- optimization may improve representation or lowering and remove language/runtime overhead only through locally semantics-preserving transformations; it must not replace the Intel BID decimal algorithm, special-case partitioning, rounding/status model, or implementation lineage with an independently designed algorithm, approximation, alternate backend, or fast path that cannot be mechanically traced to the canonical predecessor
- a Decimal operation body with no canonical predecessor is outside the target implementation architecture; an existing non-Intel-origin body is transitional debt rather than an exception to the mechanical-port rule, may receive correctness maintenance while its compatibility boundary is resolved, and must not be expanded, performance-specialized as an independent path, or used as precedent for another implementation
- the public Go runtime path must resolve through the Go mechanical port rather than calling Intel BID C directly outside that path
- fake placeholder arithmetic/string/conversion stubs are not an acceptable end state for the public Go implementation path
- no public API in any language package may handle an unsupported or unrepresentable input by silently truncating, clamping, masking, or otherwise coercing it, and may not crash the process through panic/trap/abort on such input; unsupported or unrepresentable input must fail through the language-specific error mechanism (Go `error` return, Rust `Result`, Swift `throws`, and idiomatic exceptions in Java/Python/JavaScript); for operations whose declared failure channel is IEEE exception flags, raising the invalid-operation flag together with the operation's idiomatic invalid result (a NaN or the conversion's invalid sentinel) is that mechanism
- the `manual` / `mechanical port` / `generated` classification applies to implementation paths and generated verification paths, not to every public API routing or plumbing layer around them
- public Go entrypoints, methods, constructors, and other API plumbing must be described as API routing/plumbing; they are judged by whether they preserve scope and semantics and whether they route through the Go mechanical port, not by inventing a fourth hybrid classification
- regular verification must be generated and reported honestly
- when a path is declared generated, cases/specs, dispatch/wrappers, and test runners/harnesses for that path are all generated rather than hand-maintained glue
- when an implementation path is declared generated, implementation entrypoints, wrappers, and glue for that path are also generated rather than hand-maintained

## Correctness And Performance Priority

- Correctness is mandatory for every supported implementation path, width,
  public API, and work tier. A lower tier never permits a wrong value, wrong
  status flag, silent input coercion, panic/trap, or cross-implementation
  divergence.
- Tier classification controls closure order and performance-investment
  priority only. Known Tier 1 correctness defects are resolved first, while
  correctness defects found outside Tier 1 are still reported and fixed.
- Performance work is proportional to observed or expected call frequency.
  The hottest Tier 1 operations receive the greatest profiling, optimization,
  and regression-measurement effort.
- No tier accepts an avoidable performance regression. A measured regression
  is acceptable only when an explicit correctness, public-contract, or
  architecture requirement requires the trade-off and the measurement and
  trade-off are reported; otherwise it is fixed before closure.
- Speed never takes precedence over value, status-flag, rounding-mode, error,
  or panic behavior. Separate value-only, status-aware, public-wrapper,
  Go-port, and generated Rust implementations are checked against the same
  canonical result and directly against each other where their paths differ.

## Non-Goals

Unless this file changes, do not redefine the project as:

- a DPD-first project
- a mandatory multi-backend product
- a hand-maintained test project
- a project whose correctness is defined by third-party library agreement

IBM decNumber may appear as a helper dependency or reference, but it is not the primary implementation target.

## Repository And Package Identity

This repository is the language-neutral `bid754` monorepo. The source
repository URL and the Go module namespace prefix are the same identity:
`github.com/sky1core/bid754`. There is no root Go module.

The first-class deliverables are the per-language bid754 libraries:

- `bid754-go/`: public Go implementation module
  `github.com/sky1core/bid754/bid754-go`. The module-root package is named
  `bid754`, so consumers use a named import:
  `import bid754 "github.com/sky1core/bid754/bid754-go"`. The Go mechanical
  port lives inside this module as the `bid754-go/internal/bidgo/` package
- `bid754-codec-go/`: public standalone Go BID codec module
  `github.com/sky1core/bid754/bid754-codec-go`
- `bid754-rs/`: public Rust implementation crate `bid754`. Its full
  three-width public API surface (the `Decimal32`/`Decimal64`/`Decimal128`
  value types with their context, flag, and rounding-mode types) is generated
  from the Go mechanical port by the `devtools/tools/go2rs` apiemit subpass,
  and is verified through the generated public-API parity gate, the portable
  decTest leg, and the readtest domain. It stays `publish = false` until the
  first crates.io release, which is an explicit release decision
- `bid754-codec-rs/`, `bid754-codec-java/`, `bid754-codec-py/`,
  `bid754-codec-js/`, `bid754-codec-swift/`: standalone BID codec packages

Non-deliverable components in the same repository:

- `devtools/`: non-published tooling module
  `github.com/sky1core/bid754/devtools` (generators, scripts, pinned
  authoritative inputs, generated intermediate artifacts); it is never tagged
  or consumed as a dependency
- `bid754-go/internal/bidgo/cexport` is an inactive repo-internal compatibility snapshot:
  its reference C source is excluded from normal link inputs, and it is never
  published or consumed as a package

The `Intel BID C -> Go mechanical port -> generated Rust` chain is the
manufacturing methodology behind these deliverables, not a ranking among
them; it does not demote any published per-language library to a second-class
artifact.

Release tag convention:

- root `v0.1.0`-style tags version the repository snapshot for Swift Package
  Manager; Go tooling ignores them because there is no root `go.mod`
- `bid754-go/v0.1.0`-style tags version the public Go implementation module
- `bid754-codec-go/v0.1.0`-style tags version the standalone Go codec module
- `bid754-rs/v0.1.0`-style tags version the public Rust implementation crate

Changing this identity again requires an explicit repository-wide migration
plan covering Go modules, package metadata, README examples, and release
publishing.

## Inter-Component Dependency Rules

- `bid754-go` and `bid754-codec-go` must not require each other; both public
  Go modules keep an empty `require` set (`make verify-zero-deps` enforces the
  zero-dependency contract structurally)
- BID codec logic intentionally exists twice per language: once inside the
  full implementation path and once as the standalone codec package. This
  duplication is a design decision, not drift; the shared
  `bid754-codec-vectors` gate (every required language consumes the same
  generated `vectors.json`) is what enforces equivalence between the copies
- `devtools` requires no public module, and no public module requires
  `devtools`. The only relationship is a filesystem one: `devtools`
  generators write generated files into the public components, and public
  module tests read pinned data under `devtools/` by relative path

## Mandatory Scope

Mandatory scope is determined by IEEE 754-2019 `shall` requirements for the formats and encodings this repository actually claims to support.

Interpretation rules:

- unsupported types are not mandatory
- unsupported encodings are not mandatory
- unsupported external interchange forms are not mandatory
- Intel upstream helper presence does not by itself make a function mandatory

For this repository's claimed BID decimal scope, mandatory work includes:

- BID decimal arithmetic and conversions the repository claims to support
- rounding modes
- exception flags
- special values and canonical BID behavior
- Clause 5 required operations within the supported BID decimal scope

Optional/recommended scope is limited to:

- Clause 5 `should` items
- Clause 9 recommended operations
- unsupported-format or unsupported-encoding helpers

The detailed mandatory/optional mapping lives in `IEEE754_SPEC.md`.

## Current Supported Boundary Notes

Current phase notes:

- one-way BID decimal -> `binary32` / `binary64` / `binary128` conversion helpers are part of the current supported surface
- the six BID width conversions (`bid32<->bid64<->bid128`, widening and narrowing) are part of the current supported surface
- `bid32_nexttoward`, `bid64_nexttoward`, and `bid128_nexttoward` are part of the current supported surface; `bid*_nextafter` is covered by the generated readtest verification surface without a public Go wrapper
- IBM GDA `reduce` is outside the public arithmetic surface because pinned Intel BID C has no canonical `bid*_reduce` predecessor
- this does not by itself redefine the repository as a full binary arithmetic implementation
- reverse binary -> BID conversion, `binary80` support, and any still-undocumented binary interchange surface do not become supported merely because one-way BID -> binary helpers exist

## Regular Verification

The regular verification domains of this repository are:

1. Intel `readtest`
2. IBM `decTest`
3. C FFI exact bit-compare
4. `BID codec vectors`

These are generated verification domains, not hand-maintained primary test domains.

Regular verification must be generated from authoritative inputs when those inputs exist.

`BID codec vectors` are a cross-language verification domain. The required
BID codec language target set is:

- Go
- Rust
- Java
- Python
- JavaScript/TypeScript
- Swift

The Rust full `bid754-rs` implementation may also consume the same vectors, but
that does not replace the standalone Rust BID codec target. Reporting BID codec
verification as complete requires the generated vector artifact to be consumed
by all six required language targets.

Regular verification is batch-generated from authoritative source inputs. Do not grow regular verification one function, one file, or one hand-curated subset at a time.

For these regular verification domains, "generated" means the whole verification path:

- source extraction / case selection
- generated case/spec artifacts
- generated dispatch / wrapper code when needed
- generated test runner / harness code

Hand-written glue in the regular verification path is not an acceptable end state.

IBM `decTest` operation names are verification-source operations, not automatic
public API names. A decTest operation adapter may map an official decTest
operation to one or more Go mechanical-port entrypoints when the official
operation semantics differ from an Intel non-computational helper with a similar
name. Such an adapter is part of the verification path only; it must not
redefine public Go API semantics or route outside the Go mechanical port.

## Intel readtest Scope

The operative Intel `readtest` scope is:

- `CMP_FUZZYSTATUS - explicit historical skip function groups + CMP_EQUALSTATUS`
- `CMP_RELATIVEERR` functions enter the profile only through an explicit closed adoption list owned by the selection code and reflected in the generated readtest profile inventory; functions outside that list remain excluded as a profile-expansion group
- the adoption list starts with the duplicate Intel `bid32_fmod` / `bid64_fmod` / `bid128_fmod` `CMP_RELATIVEERR` comparator rows (those functions are already selected by the `CMP_FUZZYSTATUS` surface) and grows only with mechanically ported Tier 3 transcendental functions; adopted rows use the pinned Intel relative comparator semantics for that path and never replace or weaken the comparator of any `CMP_FUZZYSTATUS`/`CMP_EQUALSTATUS` row

The selection policy lives in `TEST_GENERATION_SPEC.md`. Exact current
selected/excluded functions and reasons live in the generated readtest profile
inventory; `VERIFICATION_REFERENCE.md` indexes that machine-readable source.

Do not describe Intel `readtest` scope as `all of CMP_FUZZYSTATUS`.

## Document Roles

- `ARCHITECTURE_SPEC.md`: source-of-truth architecture and generation structure
- `IEEE754_SPEC.md`: required vs optional IEEE behavior for the supported BID scope
- `PLATFORM_SPEC.md`: cross-platform bit-reproducibility (OS/CPU) policy, supported platform matrix, and floating-point determinism rules
- `BID_CODEC_SPEC.md`: standalone cross-language BID codec API and vector protocol
- `TEST_GENERATION_SPEC.md`: generated-verification policy, oracle strength, and completion rules
- `DEPENDENCIES_SPEC.md`: pinned dependency and install policy
- `VERIFICATION_REFERENCE.md`: non-normative map from verification questions to manifests, generator code, generated inventories, and anchors
- `MODULE_GUIDE.md`: per-module scope map for working on one area without loading the whole repository
- `INTEL_BID_V20U4_VERIFICATION.md`: pinned upstream v20U3→v20U4 diff verification record backing `make verify-intel-bid-v20u4`
- `README.md` / `README.ko.md`: current checked-out tree and developer workflow only
- `BUILD.md`: current build and verification commands only
