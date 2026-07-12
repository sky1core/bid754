# Third-Party Dependency Spec

This document is the detailed dependency/installation policy document that supplements `SPEC.md`.

## Principles

- versions are pinned
- installation procedures must be reproducible via scripts
- version information in documents and scripts must match
- current implementation details and long-term goals are not mixed together

## Primary Dependencies

### Intel Decimal Floating-Point Math Library

This is the canonical upstream C source for this project.

| Item | Value |
|------|-----|
| Version | v20U4 |
| Source | Intel Decimal Floating-Point Math Library |
| Location | `devtools/third_party/intel_dfp/` |
| License | BSD 3-Clause |
| SHA-256 | `1df86132e7a31fd74d784fee1c679b21a088f73a8ec979cfaf784c200392e125` |

Uses:

- the canonical C source for BID arithmetic
- the extraction source for symbols/tables/test specs
- the native link target

The v20U4 upgrade source diff verification results are recorded in `docs/INTEL_BID_V20U4_VERIFICATION.md`.

Pinned build options:

```text
CALL_BY_REF=0
GLOBAL_RND=0
GLOBAL_FLAGS=0
UNCHANGED_BINARY_FLAGS=0
CFLAGS_OPT=-O3 -ffp-contract=off
```

Setup scripts record these semantic build flags in the ignored
`devtools/third_party/intel_dfp/lib/.libbid.build-flags` stamp. A missing or mismatched
stamp means the local `libbid.a` is stale for native validation/benchmarking and
must be rebuilt from the pinned source tree.

ARM64:

- the `BID_SIZE_LONG=8` override may be needed due to a `bid_conf.h` detection issue
- this correction exists to preserve the 64-bit BID path
- native setup does not replace the pinned source build with a checked-in or prebuilt `libbid.a`
- `devtools/third_party/intel_dfp/lib/libbid.a` must be an artifact built from the verified v20U4 source tree with the current machine's toolchain

## Secondary Dependencies

### IBM decTest

This is the official verification data.

| Item | Value |
|------|-----|
| Version | 2.62 |
| Source | `dectest.zip` |
| Location | `devtools/tests/*.decTest` |
| License | ICU License |
| SHA-256 | `b70a224cd52e82b7a8150aedac5efa2d0cb3941696fd829bdbe674f9f65c3926` |

Uses:

- Decimal32/64/128 case verification
- the source data for the smoke subset and full verification

## Dependencies That May Remain as Current Implementation Details

### IBM decNumber

IBM decNumber may be needed by some native paths or auxiliary verification paths in the current working tree. However, this document does not define IBM decNumber as the project's primary implementation goal.

| Item | Value |
|------|-----|
| Version | 3.68.0 |
| Source | `decNumber-icu-368.zip` |
| Location | external installation or current-tree helper flow |
| License | ICU License |
| SHA-256 | `14ec2cf30b58758493a7661b78b80abfb281652b61a425b85cda83173518fe25` |

Permitted uses:

- current-tree native compatibility glue
- auxiliary verification
- conversion/reference implementation comparison

Disallowed interpretations:

- the interpretation that "this project's primary backend goal is IBM decNumber"

## Installation Principles

Current-tree installation/preparation commands follow the current-tree workflow in `README.md` and `BUILD.md`.

Preparing generation input sources:

```bash
make setup-generation-inputs
```

This command downloads the pinned Intel BID C archive and the IBM decTest archive, verifies their checksums, and then extracts them into the generator input locations. Even if an existing Intel BID input tree is already present, it is not validated based merely on a `README` marker or the presence of a few sentinel files; it is reused only when all regular files and their contents exact-compare as identical to the pinned v20U4 archive. Existing IBM decTest input files are likewise not validated based merely on the presence of a sentinel file such as `add.decTest`; they are reused only when the `.decTest` file list and contents exact-compare as identical to the pinned 2.62 archive.

Rust dependencies are written as exact versions in `Cargo.toml`, and `Cargo.lock` is checked in to pin the resolution for CI/local verification. Repository-owned Rust crate verification (`make test-rust`, `make test-bidcodec`, `make test-bid-string`, `make verify-bidcodec-packages`, `make verify-rust-overflow`) runs with `cargo ... --locked` so lockfile drift is treated as a failure. Temporarily generated package-consumer verification, such as the external consumer smoke crate, exists to check that consumer's resolver behavior and is therefore not subject to repository lockfile verification.

The `cexport`, `libbidgo.a`, and `libbidgo.h` outputs of the inactive `bid754-go/internal/bidgo/cexport` compatibility snapshot are local build output and are not checked into the source tree. Any active verification or distribution role requires a defined reproducibility boundary through a generator or scripted build and a documented artifact policy.

GitHub-hosted runners in CI workflows use image labels that expose the version rather than `ubuntu-latest`/`macos-latest`. `actions/*` workflow dependencies are pinned to commit SHAs instead of mutable major-version tags, with the original tag left as a comment so humans can track it. Where per-language toolchain setup resolves provider-internal patch releases, that part is managed within the pinned commit of the corresponding setup action and the manifest/lockfile verification boundary.

BID codec cross-language verification uses the following language tools in the current tree.

- Go: `go`; standalone package verification creates an isolated local git release repository tagged `bid754-codec-go/v0.1.0` (the Go multi-module subdirectory tag convention) and consumes `github.com/sky1core/bid754/bid754-codec-go` through normal module version resolution without a local `replace`
- Rust: `cargo`
- Java: `javac`/`java` standard toolchain for the no-external-dependency vector runner; package builds use pinned Gradle plus checked-in `bid754-codec-java/gradle.lockfile`, and standalone package verification publishes `io.github.sky1core:bid754-codec` to an isolated temporary Maven repository
- Python: temporary virtualenv with pinned `pytest` from `bid754-codec-py/pyproject.toml`; package verification reads the `bid754-codec` version from that same `pyproject.toml` instead of hard-coding the wheel/install version
- JavaScript/TypeScript: `npm ci` from checked-in `package-lock.json`
- Swift: `swift run BidCodecVectorRunner` for the no-XCTest generated-vector runner

The 6 standalone BID codec package checks cover the package boundary in addition to the tools above.

- Go: external module smoke plus generated vector verification through a tagged `github.com/sky1core/bid754/bid754-codec-go v0.1.0` module resolved from the isolated temporary git repository; local `replace` is not the package-consumer boundary
- Rust: `cargo package --locked` without `--allow-dirty`, `cargo doc`, `cargo clippy`, external path-consumer smoke, and external generated vector verification; the package gate must fail on dirty tracked crate source rather than masking unreproducible local edits
- Java: pinned Gradle via `devtools/scripts/run_pinned_gradle.sh`, checked-in dependency lockfile verification, clean build of the exact expected library jar plus sources/javadoc jar output, external jar consumer smoke, and generated vector verification against the built jar
- Python: wheel build with exact-pinned build backend, `py.typed` inclusion, venv install, import smoke, and generated vector verification against the installed wheel
- JavaScript/TypeScript: `npm run build`, `npm pack`, install, import smoke, and generated vector verification against the installed tarball from checked-in `package-lock.json`
- Swift: release build plus external Swift package generated vector verification

Per-language package manifests pin test dependency resolution with exact
versions or lockfiles where possible. The vector JSON is not vendored by copying
it into per-language resources; instead, the `bid754-codec-vectors/vectors.json` generated by `devtools/cmd/testgen` is read directly.

Things that must change together when the dependency spec changes:

- installation scripts
- build documentation
- CI documentation/configuration
- version pins and checksums
