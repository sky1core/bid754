# Intel BID v20U4 Upgrade Audit

This audit records the required source review for the pinned upstream change from
Intel DFP `v20U3` to `v20U4`.

## Source Archives

| Version | Archive | SHA-256 |
|---|---|---|
| v20U3 | `IntelRDFPMathLib20U3.tar.gz` | `13f6924b2ed71df9b137a7df98706a0dcc3b43c283a0e32f8b6eadca4305136a` |
| v20U4 | `IntelRDFPMathLib20U4.tar.gz` | `1df86132e7a31fd74d784fee1c679b21a088f73a8ec979cfaf784c200392e125` |

The upstream release notes describe v20U4 as fixing:

- `bid128_from_string` for input strings longer than 100 digits
- `bid64_from_string` exponent handling for rounding-up/down with more than 17 coefficient digits
- `wcstod_conversion` buffer accounting
- binary floating-point rounding-mode assumptions in division, sqrt, fmod, and remainder functions
- macOS example/build warnings and an outdated command file

## Diff Inventory

The source diff inventory is reproducible with:

```bash
make audit-intel-bid-v20u4
```

That target verifies the v20U3 and v20U4 archive SHA-256 values below, extracts
both archives into a temporary directory, diffs `LIBRARY/src`, and fails unless
the expected changed-file counts and semantic-review allowlist match this
document.

`LIBRARY/src` contains 240 changed files between v20U3 and v20U4.

- 229 files are copyright/year or otherwise tiny non-semantic deltas.
- 11 files required semantic review:
  - `bid128_string.c`
  - `bid64_string.c`
  - `bid_strtod.h`
  - `bid32_sqrt.c`
  - `bid64_sqrt.c`
  - `bid128_sqrt.c`
  - `bid64_div.c`
  - `bid128_div.c`
  - `bid128_fmod.c`
  - `bid128_rem.c`
  - `bid_functions.h`

Build/example layout also changed:

- `RUNOSXINTEL64` and `macbuild` were removed from upstream.
- `RUNLINUXMACOSINTEL64_CLANG` was added.
- The active makefile now lives under `LIBRARY/`, so local setup scripts must build from that directory when present.

## Porting Decisions

### Applied To Go Mechanical Port

`bid128_string.c`:

- Added `sticky_bit` and `MIN_DIGITS` behavior for input strings longer than `MAX_STRING_DIGITS_128`.
- Prevents reading past the retained digit buffer and fixes rounding for very long decimal128 strings.
- Ported to `bid754-go/internal/bidgo/bid128_string.go` and regenerated to Rust.

`bid64_string.c`:

- Changed a carry-overflow path from resetting `add_expon` to incrementing it.
- Fixes rounding-up/down exponent handling for strings with more than 17 coefficient digits.
- Ported to `bid754-go/internal/bidgo/bid64_from_string.go` and regenerated to Rust.

Regression tests were added for both upstream release-note cases.

### Not Applicable To Go Mechanical Port

`bid32_sqrt.c`, `bid64_sqrt.c`, `bid128_sqrt.c`, `bid64_div.c`, `bid128_div.c`,
`bid128_fmod.c`, and `bid128_rem.c`:

- Upstream now saves/restores the C binary floating-point environment and forces `FE_TONEAREST` around internal binary floating-point helper computations.
- The Go mechanical port does not expose or mutate the C process floating-point rounding mode.
- Go `math` operations are not controlled by C `fenv`, so there is no direct Go source patch for this class.
- Native C FFI verification uses the rebuilt v20U4 `libbid.a`, so this upstream change is active in native C comparisons.

`bid_strtod.h`:

- Upstream fixed wide-character conversion buffer accounting.
- The public Go runtime path does not implement or route through Intel `wcstod` helpers.
- Native C paths use the rebuilt v20U4 source.

`bid_functions.h`:

- `fexcept_t` typedef handling changed for Windows.
- This is a native C header/build concern, not a Go mechanical-port semantic change.

## Generated Inputs And Verification Impact

- C table generation produced no checked-in table changes.
- C symbol generation produced no checked-in symbol inventory changes.
- At the v20U4 upgrade review point, `readtest` generated cases changed from
  79,056 to 79,060. The current checked-in supported-surface profile has since
  expanded to 81,009 generated `readtest` cases (80,964 from the Intel readtest
  profile plus 45 IEEE-deviation regression supplement rows); see `TEST_GENERATION_SPEC.md`
  and `devtools/generated/testspec/spec_index.json` (with its `readtest/` case
  shards) for the current case count.
- The added generated readtest coverage exposed the decimal128 long-string bug in generated Rust before the Go port was updated.
- Native readtest helper parsing was adjusted to match Intel `readtest.c` two-word 128-bit hex parsing for bracketed BID128 literals.

## Verification

The v20U4 upgrade was verified with:

- `make verify-generated`
- `go test ./...`
- `make test-go-modules`
- `make test-rust`
- `zsh -lc 'source ./.env.sh && go test -tags bid754_native -run "TestGenerated(ReadCases|FFIBitCompareSubset)" -timeout 240s'`
- `make test-native-smoke`
- `make test-native`

## Follow-up Current-Tree Verification

After the BID codec package split and generated string-vector gate were added,
the current tree was also checked with:

- `make audit-bidcodec-packages`
- `make test-bid-string`
- `go test ./...`
