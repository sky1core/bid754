# Known Issues

These are observed implementation defects, not changes to the contracts in
[SPEC.md](SPEC.md).

## D128-EXP-001: leading-zero exponent breaks exact cancellation

**Open.** Confirmed in repository revision `3e447c2` on 2026-09-11. The
Decimal128 string parser inherits this defect from pinned Intel BID C 2.0U4.
The Go mechanical port and generated Rust retain the affected logic; no local
IEEE-conformance deviation has been registered for this issue.

### Reproduction

The following two strings both represent exactly `1`, with no rounding or
exception flags required:

```go
prefix := "0." + strings.Repeat("0", 999999) + "1e"
control := prefix + "1000000"
affected := prefix + "01000000"
```

The fraction is `10^-1000000`; both exponent spellings mean `10^1000000`.
Only an insignificant leading zero in the exponent differs.

A public Go reproduction:

```go
package main

import (
    "fmt"
    "strings"

    bid754 "github.com/sky1core/bid754/bid754-go"
)

func main() {
    prefix := "0." + strings.Repeat("0", 999999) + "1e"
    for _, exponent := range []string{"1000000", "01000000"} {
        value, flags, err := bid754.NewDecimal128WithFlags(prefix + exponent)
        fmt.Printf("%s value=%s flags=%d error=%v\n", exponent, value.String(), flags, err != nil)
    }
}
```

Observed output:

```text
1000000 value=+1E+0 flags=0 error=false
01000000 value=+0E-6176 flags=3 error=false
```

### Affected behavior

| Public API | Affected input result |
| --- | --- |
| Go `NewDecimal128`, `NewDecimal128BIDDirect` | Rejects an exactly representable value with an error |
| Go `NewDecimal128WithFlags`, `NewDecimal128WithMode` | Returns an incorrect value and Underflow/Inexact, with no error |
| Rust `Decimal128::parse` | Returns `Err` for the exactly representable value |
| Rust `Decimal128::parse_raw`, `parse_with_flags`, `parse_with_mode` | Returns an incorrect value and Underflow/Inexact; the `Result` APIs return `Ok` |

For this positive input, nearest-even, nearest-away, toward-zero, and
toward-negative return `+0E-6176`; toward-positive returns `+1E-6176`.
All five modes should return `+1E+0` with no flags. Underflow/Inexact is
`0x03` in the public Go/Rust flag vocabulary and `0x30` in Intel C raw status.

Go Decimal32/Decimal64 constructors return exactly `1` for both spellings.
`IsValidDecimalString` consequently returns `true`: it accepts an input when
any supported width accepts it. The existing
[any-width equivalence test](../bid754-go/public_parser_convenience_regression_test.go)
checks that predicate's consistency, not the correctness of the Decimal128
result for this input.

### Cause and verification boundary

In the [Go parser](../bid754-go/internal/bidgo/bid128_string.go),
`Bid128FromString` counts the first exponent character with `i = 1`. If that
character is zero, subsequent leading zeros are skipped without resetting
the count. The `i < 7` loop then reads only six significant exponent digits:
`01000000` becomes `100000`. Combining that with the million fractional
places produces exponent `-900000` instead of `0`.
The [generated Rust parser](../bid754-rs/src/generated/bid128_string.rs)
preserves this logic from Intel's `bid128_string.c`.

Public Go and Rust reproductions on macOS arm64 cover both spellings and all
five rounding modes. Earlier direct builds of the pinned C sources reproduced
the same defect with `-O0` and `-O2` on arm64 and x86_64 via Rosetta 2; the C
probe covered eight inputs and five modes per build. Physical Intel/AMD CPU
execution remains unverified. The faulty character/counting logic is shared
parser code, with no ARM-specific branch in the affected region.

This is a string-to-Decimal128 conversion defect. Its reproduction does not
establish an error in `ScaleB` or in arithmetic on already-encoded values.
For the demonstrated input, using `e1000000` instead of `e01000000`, or the
equivalent literal `1`, avoids the failure. This workaround does not establish
correctness for every long-exponent input. The parser defect remains unresolved.
