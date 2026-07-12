# Module Work Guide

Scope map for working on one area without loading the whole repository.
Specs always win over this guide; see `SPEC.md` for document precedence.

## Module matrix

| Area | Responsibility | Contract docs | Gates to run | Do not touch from here |
| --- | --- | --- | --- | --- |
| `bid754-go/internal/bidgo/` | Go mechanical port of Intel BID C (single package `bidgo`, stdlib-only) | `SPEC.md`, `ARCHITECTURE_SPEC.md`, per-file porting headers | `cd bid754-go && go test ./internal/bidgo`, then root `make test-native-readtest`/`test-native-ffi` (C oracle) | no external deps (`verify-zero-deps` enforces); no cgo (`verify-portable-purity` enforces); do not edit `devtools/generated/` tables |
| `bid754-go/` module-root package `bid754` | Public value types and API routing over the port (`github.com/sky1core/bid754/bid754-go`) | `SPEC.md`, `IEEE754_SPEC.md` | `cd bid754-go && go test .`, `make test-native-dectest`, `make verify-all` before claiming done | route through `internal/bidgo`; never call Intel C directly; `generated_*`/`dectest_*` files only via generators |
| `bid754-go/internal/testspec/` | Generated test-spec loader/schema plumbing emitted by testgen | `TEST_GENERATION_SPEC.md` | `make verify-generated`, then affected native gates | hand-editing the emitted files; schema changes go through `devtools/internal/testgen` templates |
| `devtools/internal/testgen` + `devtools/cmd/testgen` | Generates all four regular verification domains (readtest, decTest, FFI bit-compare, BID codec vectors) and their consumers | `TEST_GENERATION_SPEC.md` | `cd devtools && go test ./internal/testgen/`, `make verify-generated`, then affected native gates | hand-editing any generated output; growing verification one hand-written case at a time |
| `devtools/tools/go2rs` (+`go2rs_tables`, +`apiemit`) | Sole generator of `bid754-rs/src/generated` and `src/tables.rs`; the `apiemit` sub-pass additionally owns the generated public API surface under `src/generated/api/` (`Decimal32`/`Decimal64`/`Decimal128` wrappers, `Context`, flag/rounding/class types) and the generated crate-root `lib.rs` | `SPEC.md`, `ARCHITECTURE_SPEC.md` | go2rs unit tests, `make verify-generated`, `make test-rust`, `make test-rust-native`, `make verify-rust-overflow` | adding an alternate Rust generation path; editing generated Rust |
| `devtools/tools/codegen` | Registry-driven `gen_types.rs`/`gen_constants.rs` (and Go check files) | `ARCHITECTURE_SPEC.md` | `make verify-generated` | writing into `bid754-rs/src/generated` (owned by go2rs) |
| `bid754-rs/` | Public Rust implementation; the go2rs `apiemit`-generated public API surface is covered by externally anchored parity and Rust decTest gates, but the crate still carries `publish = false` pending a separate user-approved crates.io publish step | `ARCHITECTURE_SPEC.md`, `bid754-rs/README.md` | `make test-rust` (includes the generated public API parity gate and the Rust decTest portable leg), `make test-rust-native`, `make verify-rust-package` | direct edits anywhere under `src/generated/` (including `src/generated/api/` and the generated `lib.rs`); quality/API-surface changes go through `devtools/tools/go2rs`/`apiemit` |
| `bid754-codec-go/` + `bid754-codec-{rs,java,py,js,swift}/` | Standalone BID codec packages (encode/decode/parse only, not full arithmetic) | `BID_CODEC_SPEC.md`; generation policy in `TEST_GENERATION_SPEC.md` | `make test-bidcodec`, `make verify-bidcodec-packages` | vendoring the generated vector file as package data; changing vector semantics without the generator |
| `bid754-go/internal/bidgo/cexport/` | Inactive C ABI compatibility snapshot | `ARCHITECTURE_SPEC.md` | `make verify-cexport-disabled` (passes only when the module remains outside normal link inputs) | treating the snapshot as an active runtime or verification path |
| Makefile / `devtools/scripts/` | Gates and reproducible setup | `BUILD.md`, `DEPENDENCIES_SPEC.md` | `make check-scripts`, `make verify-all`, `make verify-linux` | removing failure propagation; omitting pinned checksum verification |

## devtools/internal/testgen file map

The package keeps one import path but is split by domain so a change touches
only its area. Shared spec core: `spec.go` (types), `spec_io.go` (index/shard
encode + `LoadGenerated`/`WriteOutput`), `spec_build.go` (`buildSpec`, decTest
file selection, shared parsers). Domain spec/case builders: `readtest_spec.go`,
`ffi_spec.go`. Domain code generators: `readtest_codegen.go`,
`readtest_test_codegen.go`, `dectest_test_codegen.go`, `ffi_test_codegen.go`,
`bid_codec_vectors_*.go`, `bid_codec_reference.go`, `bid_codec_vector_anchors.go`,
`bid_string_vectors_codegen.go`, `dectest_skip_reason.go`,
`testspec_codegen.go` (emits `bid754-go/internal/testspec/` from
`testspec_templates/`). The thin public entry stays in `generate.go`.
