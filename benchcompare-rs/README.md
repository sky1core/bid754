# benchcompare-rs

Comparative benchmarks of `bid754-rs` against
[rust_decimal](https://crates.io/crates/rust_decimal) (pinned 1.42.1), the
most widely used Rust decimal crate, over parse, to-string, parts
encode/decode, and the add/mul/div arithmetic core.

This is benchmark infrastructure, not a verification domain. The two crates
implement different arithmetic models, so every row compares cost on a
shared operand set, never result equality:

- `bid754` is fixed-width IEEE 754-2019 Decimal64 (16 significant digits,
  rounding modes, status flags); `rust_decimal` is a 96-bit binary mantissa
  with a decimal scale of 0..=28 and no IEEE status/rounding surface.
- Operand literals are exactly representable in both.
- Division rows compare cost only: the division rules differ by design.
- rust_decimal stores (mantissa, scale) natively, so its parts accessors are
  near-free by construction, while the BID side decodes the interchange
  encoding through `bid754::bid_codec`; the parts rows expose that asymmetry
  rather than hiding it.

Run:

```
make bench-compare-rs        # from the repository root
```

or directly:

```
cd benchcompare-rs && cargo test --locked && cargo bench --locked --bench compare
```

The `operand_contract` unit test enforces the shared input contract (every
parse input exact and flag-clean on the BID side, every parts row encodable
and within rust_decimal's scale range); the make target runs it before the
benches.

Numbers are environment-specific; treat any published figures as one
machine's snapshot, not a portable claim.

This package is standalone on purpose (`publish = false`, own workspace):
the comparison crate `rust_decimal` is pinned only here and never enters the
product crate's dependency set.
