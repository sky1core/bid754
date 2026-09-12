# Known Issues

Observed defects and their resolution status. The contracts remain defined in
[SPEC.md](SPEC.md).

## D128-EXP-001: leading-zero exponent breaks exact cancellation

**Fixed.** First confirmed in revision `3e447c2` on 2026-09-11. The Go
mechanical port and generated Rust now accumulate the complete exponent up to
a bound derived from input length and the destination exponent range. The
[registered IEEE deviation](IEEE754_SPEC.md) also covers the related fixed
exponent limits inherited from pinned Intel BID C 2.0U4 in Decimal32/64/128.

Both spellings below now produce exact `1` with no flags in all five rounding
modes. Regular generated readtest vectors and Go/Rust parser boundary tests
cover the regression; production fault injection checks that restoring a
fixed exponent cap fails the numeric assertions.

### Original reproduction

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

Output before the fix:

```text
1000000 value=+1E+0 flags=0 error=false
01000000 value=+0E-6176 flags=3 error=false
```

### Behavior before the fix

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

### Cause and coverage

The old Decimal128 parser counted the first exponent character toward its
seven-digit limit even when that character was zero. It interpreted
`01000000` as `100000`, so the million fractional places yielded exponent
`-900000` instead of `0`. Its seven-digit cap also broke exact cancellation
at `10000000` without any leading zero. Decimal32/64 stopped accumulation at
`1 << 20`; an exponent of `10485760` reproduced the same class of failure.

The regression family checks exponent cancellation, leading zeros, signs,
rounding modes, written cohorts, rejection boundaries, and exception flags
through the Go mechanical port, public Go constructors, generated Rust, and
public Rust constructors. Native comparison skips are limited to registered
vectors on which pinned Intel C disagrees with the required IEEE result.

Earlier direct builds of pinned C reproduced the original leading-zero
case with `-O0` and `-O2` on arm64 and x86_64 via Rosetta 2. Physical Intel/AMD
CPU execution remains unverified. This defect concerns string conversion;
it does not establish an error in `ScaleB` or arithmetic on encoded values.
