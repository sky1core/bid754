# benchcompare-rs

Comparative benchmarks of `bid754-rs` against
[rust_decimal](https://crates.io/crates/rust_decimal) (pinned 1.42.1), the
most widely used Rust decimal crate, over parse, to-string, parts
encode/decode, and the add/mul/div arithmetic core, at every product width:
Decimal32, Decimal64, and Decimal128.

This is benchmark infrastructure, not a verification domain. The two crates
implement different arithmetic models, so every row compares cost on a
shared operand set, never result equality:

- `bid754` is fixed-width IEEE 754-2019 Decimal32/64/128 (7/16/34
  significant digits, rounding modes, status flags); `rust_decimal` is a
  96-bit binary mantissa with a decimal scale of 0..=28 and no IEEE
  status/rounding surface.
- Each width's operand literals are exactly representable at that BID
  width and in rust_decimal: the Decimal128 list is capped at 28
  significant digits with value magnitudes in [1e-8, 1e12], because
  rust_decimal cannot represent the wider Decimal128 range and its
  operators panic on overflow (the operand-contract test checks every
  benchmark pair with checked add/mul/div).
- Division rows compare cost only: the division rules differ by design.
- rust_decimal stores (mantissa, scale) natively, so its parts accessors are
  near-free by construction, while the BID side decodes the interchange
  encoding through `bid754::bid_codec`; the parts rows expose that asymmetry
  rather than hiding it.

Run:

```
make bench-compare-rs        # from the repository root
```

or directly — `--baseline` requires the baseline to already exist, so save it
once before comparing:

```
cd benchcompare-rs && cargo test --locked
cargo bench --locked --bench compare -- --save-baseline pinned   # first run only
cargo bench --locked --bench compare -- --baseline pinned        # every run after
```

The make target handles that split itself: it compares against the named
Criterion baseline `pinned` and saves it when it is missing, the same policy
as `make bench-rust` and `make bench-codec-rs`; `make bench-compare-rs-baseline`
is the only target that moves an existing baseline. Criterion's anonymous
default baseline is deliberately not used: it accumulates unnamed results from
other trees and older row names, so its change percentages point at an
unidentifiable predecessor.

The `operand_contract` unit test enforces the shared input contract (every
parse input exact and flag-clean on the BID side, every parts row encodable
and within rust_decimal's scale range); the make target runs it before the
benches. The per-width operand lists are shared with `benchcompare-go` by
hand; edit them here and in the Go module in the same change.

This package is separate benchmark tooling and is outside `make verify-all`.

Numbers are environment-specific; treat any published figures as one
machine's snapshot, not a portable claim.

This package is standalone on purpose (`publish = false`, own workspace):
the comparison crate `rust_decimal` is pinned only here and never enters the
product crate's dependency set.
