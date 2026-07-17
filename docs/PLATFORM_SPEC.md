# Platform Determinism Spec

> Status: **Finalized (ACTIVE)** — decided per the "accuracy comes first" principle (2026-06). This
> sits at position 4 in the `SPEC.md` document precedence, and on conflict `SPEC.md` takes priority.
>
> Decision principle: bit reproducibility is established not by "currently matching in measurements"
> (statistical reassurance) but by "removing the divergeable paths themselves" (structural
> guarantee). Therefore, every floating-point seed path is sealed as unfused, and platforms that
> cannot be guaranteed are nailed down as explicitly unsupported rather than left as best-effort.

This document defines **over which range of platforms bid754's arithmetic results are guaranteed to
be bit-identical**, and **with which build, runtime, and verification pinning that guarantee is enforced**.

## 1. Supported Platform Matrix

Targets for which bit reproducibility (same inputs → same result bits) is **guaranteed**:

| Tier | Platform | Basis |
|---|---|---|
| **Guaranteed** | macOS arm64 (Apple Silicon) | LP64·little-endian, SSE2/NEON-equivalent IEEE 754 |
| **Guaranteed** | Linux amd64 (x86-64, SSE2) | LP64·little-endian |
| **Guaranteed** | Linux arm64 (AArch64) | LP64·little-endian (see below for the verification-path difference) |

Common premise: **64-bit·little-endian·IEEE 754 scalar floating point (SSE2/AArch64, not x87)**.

Production-code bit reproducibility is guaranteed on all three of the above. **The verification path
for Linux arm64 differs**: the Intel BID C library (libbid.a) **does not build** on Linux arm64
because the upstream makefile is x86-only (no `aarch64` in `ARCH_ALIAS`; an x86-64 cross-build
yields incompatible objects). Since Intel BID C is not a production library but the **C oracle** for
native FFI bit-compare, the production correctness of Linux arm64 is verified via the Go/Rust
sealing + platform-independent fixed vectors (`readtest`/`decTest`), and native FFI bit-compare
(direct comparison against Intel C) is performed on macOS arm64 + Linux amd64 where Intel BID builds
(§4). Production bit identity is cross-covered by the two axes of macOS arm64 (same ISA) and Linux amd64 (same OS).

Targets for which bit reproducibility is **not guaranteed** (they may build/run, but identical
result bits are not promised):

- 32-bit x86 (x87 80-bit extended-precision path) — bit-mismatched with float64 due to the extended exponent
- big-endian (s390x etc.) — arithmetic and byte-serialization identity are exercised by the local qemu `digest-s390x` leg (platform digest, standalone codec tests, and the generated Go-port readtest and public parity runners agree with the little-endian platforms); per section 4 this QEMU evidence is an auxiliary signal and native big-endian hardware confirmation remains outstanding
- Windows amd64 — since C `long`=4 (LLP64), the int-width branch in `bid64_lrint`/`bid64_lround` can
  produce results different from POSIX (these functions are optional C-compat, not IEEE mandatory
  fixed-width conversions)

## 2. Floating-Point Determinism Policy (Enforced)

The BID core arithmetic is integer-based, so it is inherently bit-identical across the matrix above.
The remaining platform uncertainty is limited to the hardware-float **seed paths** of div/sqrt/conversion and to denormal/NaN.
This is pinned down as follows.

1. **The Intel BID C library (libbid.a) build must explicitly enforce `-ffp-contract=off`.** Since
   GNU mode defaults to `fast` and Clang 14+ defaults to `on`, FMA fusion is enabled by default, so
   even a standard C build is not validated. **It is not put into native cgo `#cgo CFLAGS`** — Go cgo
   rejects `-ffp-contract` as an invalid flag, and the C that cgo compiles is merely a libbid-call
   wrapper with no FP seed arithmetic, so it is unnecessary (all FP seeds are inside libbid.a).
2. **The hardware-float seed fusion (FMA) state of the Go, Rust, and C paths must match.** Per its
   spec, Go contracts `a*b+c` into FMA (this actually happens on arm64), Rust does not auto-fuse,
   and C follows the flags. Pin the three paths to the same fused/unfused state in the generator
   rules (`devtools/tools/go2rs`, the Go port emit). Generated artifacts are not edited directly.
3. **The rounding mode is pinned to roundTiesToEven (RNE)**, and denormals are assumed preserved
   (FTZ/DAZ off). Since ARM64 FPCR is per-thread and non-inherited, either guarantee that native
   FFI/benchmarks do not change the FPU mode, or use only paths that produce no denormals.
4. **NaN payloads are pinned to the BID rules in bit comparison, and NaNs from auxiliary
   binary-float paths are normalized or excluded from comparison.** Rust NaN bits are
   non-deterministic, and QEMU NaN encodings differ per ISA.

## 3. Build Flag Pinning

| Target | Pinned |
|---|---|
| Intel BID C (`devtools/scripts/setup_c_libs.sh` etc.) | `-O3 -ffp-contract=off`, `BID_SIZE_LONG=8` (64-bit). Because the upstream makefile is x86-only, Intel BID builds only on Linux amd64 and macOS arm64 and is not built on Linux arm64 (§1) — Intel BID is C-oracle-only, so this is irrelevant to production |
| native cgo (`#cgo CFLAGS`) | `-ffp-contract=off` is not added (cgo rejects it; the wrapper has no FP seeds). The `-lm` link + the BID values `BID_UINT*`/`_IDEC_flags*` are applied |
| Rust (`bid754-rs`) | No automatic FMA → no extra flags needed. The generator emits no explicit `mul_add` use. |
| Go | FMA blocking on the hardware-float seed paths via generator/port rules (§2.2) |

## 4. Verification Gate Requirements (Enforced)

1. Run native verification (FFI bit-compare, readtest, decTest) **in CI on both platforms where
   Intel BID builds: macOS arm64 + Linux amd64**. (Current status: locally, the macOS native gate
   (the native stage of `make verify-all`) and the Docker Linux leg
   (`make verify-linux-native-amd64`) perform native verification for the two platforms. The native
   job matrix of the checked-in `.github/workflows/build.yml` is the gate definition for this CI
   requirement, and the workflow is applied to remote main. Confirming live CI (green) remains a
   task for after the public transition.)
2. **Direct cross-platform diff gate**: compare output bit digests for the same input set across
   platforms against each other. This complements the indirect guarantee of the fixed-vector gates
   (platform-independent expected values). (Current tree implementation: `make digest` produces,
   from generated testspec inputs, a SHA-256 digest of the input/result bits and flags of the
   seed-sensitive core operations (`bid32/64/128` add/sub/mul/div/fma/sqrt); the `make verify-linux`
   portable leg retrieves the Linux digest; and `make verify-digest` enforces cross-platform agreement.)
3. **Linux arm64 is verified with production (Go portable + Rust) + fixed vectors
   (readtest/decTest Go-vs-expected)**. Since Intel BID does not build on Linux arm64 (§1·§3),
   native FFI bit-compare (direct comparison against Intel C) is not run on this platform. The
   arm64 production bits are cross-covered by macOS arm64 (native FFI on the same ISA). For the
   readtest half, the concrete gate is the generated `TestGeneratedReadCasesGoPort` runner. It
   executes every generated row except the `status_control` rows, which model Intel global-state
   helpers rather than the explicit-flags Go mechanical-port surface. The full, executed, and
   excluded counts are generated and independently pinned under the count-ownership rules in
   `TEST_GENERATION_SPEC.md`; they are not duplicated as a platform contract here. The runner
   needs no cgo and therefore runs inside `make test-go-modules` on the portable arm64 Linux leg.
4. QEMU execution results are used only as auxiliary signals; bit-reproducibility confirmation is
   done with native hardware CI (GitHub ubuntu=amd64, macos=arm64).

## 5. Finalized Decisions (Accuracy First)

1. **Supported matrix**: pinned as in §1. Bit reproducibility guaranteed for {macOS arm64, Linux
   amd64, Linux arm64}; {x86 32-bit, big-endian, Windows amd64 (long=4)} are **explicitly
   unsupported**. They are not left as best-effort.
2. **`-ffp-contract=off` enforced**: applied to the Intel BID C library (libbid.a) build. It is not
   put into native cgo — Go cgo rejects this flag, and the C that cgo compiles is a wrapper with no
   seeds (§2.1·§3). After applying, rerun native FFI bit-compare/readtest/decTest on macOS arm64 +
   Linux amd64 and confirm no regression. Linux arm64 is verified with production + fixed vectors
   since Intel BID is not built there (§4).
3. **precedence**: this document sits after `IEEE754_SPEC.md` (position 4).
4. **FP seed sealing = structural sealing adopted**: not "it currently absorbs, so watch it with a
   regression gate", but **seal the seed paths as unfused regardless of language and platform**.
   Stages: ① C `-ffp-contract=off` ② explicit unfused Go and Rust seeds (the Go mechanical port
   `bid754-go/internal/bidgo` via mechanical-port porting rules, with Rust following via go2rs
   regeneration; generated artifacts are changed only through their generators) ③ native bit-compare
   re-verification at each stage. Since the sealing matches the IEEE 754 default semantics (rounding
   after each operation), it does not damage the mechanical-port semantics.
