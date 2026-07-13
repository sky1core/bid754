# Architecture Spec

This document is the detailed architecture document that supplements `SPEC.md`.

This document defines the project's target architecture. It is not a document claiming that the currently checked-out tree has implemented all of this structure.

## Core Principles

- Intel BID C is the canonical source
- BID is the encoding standard
- value types must be fixed-width (`types_layout_check.go` pins the 4/8/16-byte layouts at compile time)
- definitions, tables, and regular verification test specs must be extracted/generated from C or the corresponding official input sources
- table generation extracts/generates directly from Intel BID C to both Go and Rust
- the Go implementation path is a path that directly mechanically ports the Intel BID C implementation
- the Rust implementation path is a path generated from the Go implementation path
- `devtools/tools/go2rs` is the only permitted generation path for full Rust implementation artifacts
- generated outputs must be reproducible
- the Go mechanical port represents every multi-limb `BID_UINT` as scalar `uint64` fields rather than an array field: `BID_UINT128` uses `lo`, `hi`, and `BID_UINT192`/`256`/`320`/`384`/`512` use `w0` through `wN` in least-significant-to-most-significant order; this representation mapping preserves Intel's word order and the respective 16/24/32/40/48/64-byte value layouts, and the generated Rust types mirror the same ordered fields

## Target Structure

Canonical flow:

```text
Intel BID C source
  -> symbols / constants / type metadata extraction
  -> table extraction
  -> test-spec extraction
  -> generated Go table artifacts
  -> generated Rust table artifacts
  -> mechanical Go implementation path
  -> generated Rust implementation path from Go
```

Important:

- neither Go nor Rust is the original source
- do not confuse the table generation path with the implementation generation path
- tables are generated from C to both Go and Rust
- the Go implementation is a direct mechanical port path of the C implementation
- the Rust implementation is generated from the Go implementation
- the full Rust implementation artifacts in `bid754-rs/src/generated` must be `devtools/tools/go2rs` output
- Rust implementation quality, Rust idiom, and performance optimization improvements are made only by fixing `devtools/tools/go2rs` or its support/prelude generation rules and regenerating
- semantic regression checks for `devtools/tools/go2rs` changes are the post-regeneration generated Rust verification (pinned vectors + native readtest) and the go2rs golden tests themselves. Changing go2rs requires passing these gates
- a separate Go->Rust converter, a C->Rust implementation generator, a hand-written Rust replacement implementation, and direct edits to generated Rust are not permitted
- the public Rust API layer — the value-type entrypoints, methods, constructors, associated constants, and the crate-root re-export in `bid754-rs/src/lib.rs` — is generated from the Go mechanical port by the `devtools/tools/go2rs` apiemit subpass into `bid754-rs/src/generated/api/`. Because this implementation path is declared generated, its entrypoints, wrappers, and glue are part of the generated path; they are not hand-written or edited directly
- public Go value-type entrypoints, methods, and constructors are not a separate implementation path but API routing/plumbing, and must be connected through the Go mechanical port path
- a structure where the public Go runtime path calls Intel BID C directly instead of going through the Go port is transitional debt, not the target structure
- fake stubs that return placeholder values in the public Go arithmetic/string/conversion paths are not the target structure
- the `manual` / `mechanical port` / `generated` classification is attached to implementation body paths and generation paths
- API routing/plumbing such as public Go entrypoints/methods/constructors is not described as if it were a separate target of the above three-way classification, and is treated only as public API routing/plumbing
- public Go API routing/plumbing is evaluated by whether it passes through the Go mechanical port path and whether it leaves semantics unchanged
- the public Go API routing doctrine is mechanically enforced by `devtools/internal/publicroute` — public census↔inventory consistency, inventory↔shim call-set equality, reachable-union coverage of every inventoried public function into the mechanical port, and verification that public build files do not use cgo — together with the generated public-API parity runner, whose counts are pinned as `public_api_*` anchors in `devtools/verification_anchors.json`
- generated files are not edited directly
- generation rules and manifests are modified and then regenerated
- a state where, in a generated implementation path, only some of the entrypoints, wrappers, and glue are generated and the rest are hand-written is not the target structure
- a state where, in a generation-target path, only some of the cases/specs, dispatchers/wrappers, and runners/harnesses are generated and the rest are hand-written glue is not the target structure
- Intel BID implementation/verification scope is managed in units of the upstream function groups exposed in `readtest.h` and similar files, and the IEEE mandatory/optional classification is mapped onto those function groups
- the Intel `readtest` operating scope is not lumped together as `all of CMP_FUZZYSTATUS`, but is written as `CMP_FUZZYSTATUS - explicit historical skip function groups + CMP_EQUALSTATUS`, the criterion actually used for past automatic generation

Regular verification targets:

- Intel `readtest`
- IBM `decTest`
- C FFI exact bit-compare
- `BID codec vectors`

The above four are the regular verification targets of this repository. The expression "regular verification complete" is used only with reference to these categories.

Intel `readtest` operating criteria:

- `CMP_FUZZYSTATUS - explicit historical skip function groups + CMP_EQUALSTATUS`
- `CMP_RELATIVEERR` is excluded as a profile-expansion group, but the Intel duplicate `CMP_RELATIVEERR` comparator rows of `bid32_fmod` / `bid64_fmod` / `bid128_fmod`, which are already included in the `CMP_FUZZYSTATUS` surface, may be applied separately per generated runner
- selection policy follows `TEST_GENERATION_SPEC.md`; exact current exclusions
  and reasons come from the generated readtest profile inventory

Higher-level documents use the criterion formula above. Human-readable routing
to the generated inventory is in `VERIFICATION_REFERENCE.md`; prose does not
duplicate the function list.

## Implementation Boundaries

Required:

- BID-based IEEE 754 behavior
- maintaining fixed-width value types
- reproducible generation paths based on the C source
- verification automation and accurate pass/fail/skip tallies

Auxiliary/optional artifacts:

- additional optimization paths
- per-language auxiliary libraries that are not full Decimal arithmetic implementations

Auxiliary/optional items may exist, but if they are absent from the current tree, they are not described in documents as if implemented.
The Rust implementation path itself belongs to the generated implementation
path of the target structure above, so it is not classified as an optional artifact.

## Relationship to the Current Tree

If hand-written Go code or transitional glue code remains in the current tree, that means the target structure has not yet been achieved. Such a state is not documented as part of the target structure.

In particular, do not confuse the following.

- current native implementation details
- the long-term target structure
- verification smoke paths
- full verification goals
- whether the public Go path is the Go mechanical port, or C direct wrapper/fake stub glue

## Current Repository Layout Rules

File locations in the current tree are interpreted by role. The `manual` / `mechanical port` / `generated` classification is not inferred from directory names alone.

Key layout:

- `bid754-go/`: public Go implementation module (`github.com/sky1core/bid754/bid754-go`). The module root package `bid754` contains the public Go value types, API routing/plumbing, generated module root declarations, and generated module root tests
- `bid754-go/internal/bidgo/`: implementation path that directly mechanically ports Intel BID C to Go (package `bidgo`)
- `bid754-go/internal/bidgo/tables_binarydecimal.go`: Go-port table artifact generated directly from pinned Intel `bid_binarydecimal.c` by c-tablegen; its `BID_UINT128`, `BID_UINT192`, and `BID_UINT256` entries use the mechanical port's ordered scalar-limb representations
- `bid754-go/internal/bidgo/cexport/`: inactive C ABI compatibility snapshot outside the regular `readtest` generated verification path. Its reference C source is excluded from normal link inputs by the `.disabled` extension. `cexport`, `libbidgo.a`, and `libbidgo.h` are local build outputs, not checked-in artifacts
- `bid754-go/internal/testspec/`: generated test-spec loader/schema plumbing produced by `devtools/cmd/testgen`. It is part of the generated verification path and is not edited directly
- `devtools/generated/go/`, `devtools/generated/json/`, `devtools/generated/testspec/`: table/symbol/test-spec artifacts generated from C or official inputs
- `bid754-rs/src/intel_dfp_tables.rs`: Rust table artifact generated from Intel BID C with c-tablegen. It is generated inside the crate so that the published crate is self-contained
- `bid754-rs/src/generated/`: Rust implementation artifacts generated from the Go mechanical port path, including the `api/` subtree (the public value-type API surface) emitted by the `devtools/tools/go2rs` apiemit subpass
- `bid754-rs/src/lib.rs`: generated crate root (apiemit subpass) that re-exports the generated `api/` surface as the public API and gates the internal modules behind `#[doc(hidden)]`
- `bid754-rs/src/tables.rs`: compatibility layer connecting the Rust table artifact generated from Intel BID C to the Rust implementation path
- `bid754-codec-go/`: public standalone Go BID codec module (`github.com/sky1core/bid754/bid754-codec-go`, package `bidcodec`)
- `bid754-codec-rs/`: standalone Rust BID codec helper package
- `bid754-codec-java/`: Java BID codec helper package
- `bid754-codec-py/`: Python BID codec helper package
- `bid754-codec-js/`: JavaScript/TypeScript BID codec helper package
- `bid754-codec-swift/`: Swift BID codec helper package
- `bid754-codec-vectors/`: BID codec cross-language vector artifact generated by `devtools/cmd/testgen`
- `devtools/tests/`: pinned IBM decTest official inputs
- `devtools/third_party/intel_dfp/`: pinned Intel BID C official inputs and native build output location
- `devtools/cmd/*`, `devtools/tools/*`, `devtools/internal/*`: generation/extraction/conversion/test-spec tooling (private tooling module `github.com/sky1core/bid754/devtools`)

Generated Rust overflow policy:

- `bid754-rs` must not disable Rust overflow checks at the Cargo profile level
- C/Go-style integer wraparound and oversized shift behavior must be emitted explicitly by `devtools/tools/go2rs` with generated `wrapping_*` / checked-shift support, not hidden behind Cargo profile settings
- `make verify-rust-overflow` is the current verification boundary: it runs the generated Rust tests under the default Rust test profile and again with `RUSTFLAGS='-C overflow-checks=yes'`
- this policy is not a license to add hand-written unchecked arithmetic; generated implementation behavior must stay in `devtools/tools/go2rs` or generated support/prelude rules

Generated Rust `std` policy:

- full `bid754-rs` is currently a `std` crate
- standalone `bid754-codec-rs` may keep its own `no_std` support, but that does not imply the generated full Rust implementation supports `no_std`
- adding `no_std` support for `bid754-rs` requires a separate generator/support-module pass that removes or gates current `String`, `Vec`, `format!`, and `std::env` usage

The per-language BID codec helpers are not full Decimal arithmetic
implementations. The responsibility of this path is the encode/decode/parse
layer between BID bit patterns and the
`{sign, coefficient, exponent, kind, payload}` components, and verifying the
little-endian bytes API in each language against the same vectors. The
required language set is Go, Rust, Java, Python, JavaScript/TypeScript, and
Swift; if any one of these is missing from the current tree, BID codec
cross-language verification is reported as incomplete. `make test-bidcodec`
is the repo-level generated vector consumer verification, and
`make verify-bidcodec-packages` is a separate quality gate that checks all the
way to the build/package/install/import boundaries of the six standalone
packages.

Generated files may exist in the `bid754-go` module root package. Public declarations or test runners belonging to package `bid754` may have to live in the module root due to Go package constraints; in that case, generated status is judged by file headers and generator/manifest reproducibility. The generated verification plumbing in the module root (e.g. `dectest_spec_test.go`, `generated_*` dispatch/runners) and the `bid754-go/internal/testspec/` loader plumbing are part of the generated verification path, and are kept as symbols/packages that are not exposed outside the module, rather than as exported API.

Unsupported structures:

- editing generated files directly in order to fix a generated artifact
- treating generated module root test/dispatch files as hand-maintained merely because they are in the module root
- documenting public API routing/plumbing as if it were a separate implementation backend
- using DPD canonicalization material as if it were evidence of BID implementation completion

## Non-Goals

- making DPD a primary implementation goal
- using phrasing like "2 mandatory backends" as a goal definition
- mixing the table C extraction and the implementation generation path and describing them as if they were one rule
- incorrectly stating that the Rust implementation is generated directly from Intel BID C
- manual maintenance outside the declared generation paths

When documents conflict, precedence follows `SPEC.md` first.
