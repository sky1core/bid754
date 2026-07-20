# benchcompare-go

Comparative benchmarks of `bid754-go` against
[shopspring/decimal](https://github.com/shopspring/decimal) (pinned v1.4.0),
the most widely used Go decimal library, over parse, to-string, parts
encode/decode, and the add/mul/div arithmetic core.

This is benchmark infrastructure, not a verification domain. The two
libraries implement different arithmetic models, so every row compares cost
on a shared operand set, never result equality:

- `bid754-go` is fixed-width IEEE 754-2019 Decimal64 (16 significant digits,
  rounding modes, status flags); `shopspring/decimal` is arbitrary-precision
  over `big.Int`.
- Operand literals are exactly representable in both.
- Division rows compare cost only: the division rules differ by design.
- shopspring stores (coefficient, exponent) natively, so its parts accessors
  are near-free by construction, while the BID side decodes the interchange
  encoding through `bid754-codec-go`; the parts rows expose that asymmetry
  rather than hiding it.

Run:

```
make bench-compare-go        # from the repository root
```

or directly:

```
cd benchcompare-go && GOWORK=off go test -run '^TestOperandContract$' -bench . -count 5
```

`TestOperandContract` enforces the shared input contract (every parse input
exact and flag-clean on the BID side and accepted by shopspring, every parts
row accepted by the validating codec encode); the make target runs it before
the benchmarks.

Numbers are environment-specific; treat any published figures as one
machine's snapshot, not a portable claim.

This module is standalone on purpose: the product modules (`bid754-go`,
`bid754-codec-go`) stay zero-dependency, and the comparison library is pinned
only here.
