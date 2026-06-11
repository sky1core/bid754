# Module Work Guide

Scope map for working on one area without loading the whole repository.
Specs always win over this guide; see `SPEC.md` for document precedence.

## Module matrix

| Area | Responsibility | Contract docs | Gates to run | Do not touch from here |
| --- | --- | --- | --- | --- |
| `bid754-go/internal/bidgo/` | Go mechanical port of Intel BID C (single package `bidgo`, stdlib-only) | `SPEC.md`, `ARCHITECTURE_SPEC.md`, per-file porting headers | `cd bid754-go && go test ./internal/bidgo`, then root `make test-native-readtest`/`test-native-ffi` (C oracle) | no external deps (`audit-zero-deps` enforces); no cgo (`audit-portable-purity` enforces); do not edit `devtools/generated/` tables |
| `bid754-go/` module-root package `bid754` | Public value types and API routing over the port (`github.com/sky1core/bid754/bid754-go`) | `SPEC.md`, `IEEE754_SPEC.md` | `cd bid754-go && go test .`, `make test-native-dectest`, `make full-audit` before claiming done | route through `internal/bidgo`; never call Intel C directly; `generated_*`/`dectest_*` files only via generators |
| `bid754-go/internal/testspec/` | Generated test-spec loader/schema plumbing emitted by testgen | `TEST_GENERATION_SPEC.md` | `make verify-generated`, then affected native gates | hand-editing the emitted files; schema changes go through `devtools/internal/testgen` templates |
| `devtools/internal/testgen` + `devtools/cmd/testgen` | Generates all four regular verification domains (readtest, decTest, FFI bit-compare, BID codec vectors) and their consumers | `TEST_GENERATION_SPEC.md` | `cd devtools && go test ./internal/testgen/`, `make verify-generated`, then affected native gates | hand-editing any generated output; growing verification one hand-written case at a time |
| `devtools/tools/go2rs` (+`go2rs_tables`) | Sole generator of `bid754-rs/src/generated` and `src/tables.rs` | `SPEC.md`, `ARCHITECTURE_SPEC.md` | go2rs unit tests, `make verify-generated`, `make test-rust`, `make test-rust-native`, `make audit-rust-overflow` | adding an alternate Rust generation path; editing generated Rust |
| `devtools/tools/codegen` | Registry-driven `gen_types.rs`/`gen_constants.rs` (and Go check files) | `ARCHITECTURE_SPEC.md` | `make verify-generated` | writing into `bid754-rs/src/generated` (owned by go2rs) |
| `bid754-rs/` | Public Rust implementation, pre-release (`publish = false`, no stable external API yet; user-facing API layer is a later phase) | `ARCHITECTURE_SPEC.md`, `bid754-rs/README.md` | `make test-rust`, `make test-rust-native` | direct edits anywhere under `src/generated/`; quality changes go through `devtools/tools/go2rs` |
| `bid754-codec-go/` + `bid754-codec-{rs,java,py,js,swift}/` | Standalone BID codec packages (encode/decode/parse only, not full arithmetic) | `TEST_GENERATION_SPEC.md` BID codec section | `make test-bidcodec`, `make audit-bidcodec-packages` | vendoring the generated vector file as package data; changing vector semantics without the generator |
| `bid754-go/internal/bidgo/cexport/` | Quarantined legacy stubs | `ARCHITECTURE_SPEC.md` | `make audit-cexport-quarantine` (must fail to pass) | reviving the stubs as runtime or verification evidence |
| Makefile / `devtools/scripts/` | Gates and reproducible setup | `BUILD.md`, `DEPENDENCIES_SPEC.md` | `make check-scripts`, `make full-audit`, `make verify-linux` | weakening failure propagation; bypassing pinned checksums |

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

## Delegation notes

- Self-contained units: a single codec language package, a `bidgo` function
  port that is not publicly wired yet, cexport quarantine upkeep.
- Center-coupled units (need wider context): `devtools/internal/testgen` domain
  changes, `devtools/tools/go2rs` changes, public API wiring. Budget extra
  review for these.
- Every task ends with the gates listed for its row; `make full-audit` is the
  repository-level boundary before any "done" claim.
