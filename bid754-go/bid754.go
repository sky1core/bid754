// Package bid754 provides IEEE 754-2019 decimal floating-point arithmetic
// over fixed-width BID (Binary Integer Decimal) value types, as a direct
// mechanical port of the Intel Decimal Floating-Point Math Library.
//
// The three value types are:
//   - Decimal32BID: 32-bit BID decimal (7 significant digits)
//   - Decimal64BID: 64-bit BID decimal (16 significant digits)
//   - Decimal128BID: 128-bit BID decimal (34 significant digits)
//
// Each type is a fixed-width value with 1:1 byte correspondence to its BID
// bit pattern and no hidden state. Arithmetic is type-safe (no implicit
// mixing of widths), and every operation routes through the Go mechanical
// port of the pinned Intel BID C sources. Results are verified against the
// Intel C oracle (exact bit-compare and Intel readtest), the IBM decTest
// suites, cross-language BID codec vectors, and a cross-platform output
// digest; intentional deviations from the pinned C library are documented
// in IEEE754_SPEC.md.
//
// String formatting follows the Intel BID coefficient-exponent form, e.g.
// "+12400E-2". See the package examples for typical usage.
package bid754

// Version and Name identify this library build.
const (
	Version = "0.1.0"
	Name    = "bid754"
)
