package bid754

import "testing"

// ILogB is IEEE 754-2019 Clause 5.3.3 logB with an integer logBFormat. The
// finite expectations below are derived from the IEEE definition
// (floor(log10(|x|)) for a radix-10 format), not from the port, and are
// deliberately cohort-independent: "1.00" and "1" must both answer 0.
//
// The special-operand expectations are the two Intel sentinel values pinned as
// literals here rather than read back from the port or from a production
// helper, so a sentinel regression in bid<w>_ilogb cannot make the wrapper and
// this gate wrong together:
//
//   - +-Infinity        -> 2147483647 (Intel 0x7fffffff, INT_MAX)
//   - NaN / sNaN / +-0  -> -2147483648 (Intel 0x80000000, INT_MIN)
//
// all with FlagInvalidOperation and nothing else. IEEE 754-2019 Clause 5.3.3
// requires exactly this shape for an integer logBFormat: zero, infinite, and
// NaN operands signal the invalid operation exception and return a
// language-defined value outside the range +-2*(emax + p - 1), which for the
// widest supported format (Decimal128: emax 6144, p 34) is +-12354 -- both
// sentinels are far outside it at every width.
const (
	ilogbInfSentinel = 2147483647
	ilogbNaNSentinel = -2147483648
)

// ilogbFiniteCase is one finite operand and its IEEE logB value. Every string
// here is exactly representable at all three widths.
type ilogbFiniteCase struct {
	input string
	want  int
}

var ilogbSharedFiniteCases = []ilogbFiniteCase{
	{input: "1", want: 0},
	{input: "1.00", want: 0},
	{input: "9.99", want: 0},
	{input: "10", want: 1},
	{input: "-100", want: 2},
	{input: "0.001", want: -3},
}

// ilogbSpecialCase is one non-finite (or zero) operand and its pinned sentinel.
type ilogbSpecialCase struct {
	input string
	want  int
}

var ilogbSharedSpecialCases = []ilogbSpecialCase{
	{input: "Infinity", want: ilogbInfSentinel},
	{input: "-Infinity", want: ilogbInfSentinel},
	{input: "NaN", want: ilogbNaNSentinel},
	{input: "sNaN", want: ilogbNaNSentinel},
	{input: "0", want: ilogbNaNSentinel},
	{input: "-0", want: ilogbNaNSentinel},
	{input: "0.000", want: ilogbNaNSentinel},
}

func TestDecimal32ILogB(t *testing.T) {
	cases := append([]ilogbFiniteCase{}, ilogbSharedFiniteCases...)
	cases = append(cases,
		// Decimal32: emax 96, p 7 -> largest finite 9.999999E+96, smallest
		// subnormal 1E-101.
		ilogbFiniteCase{input: "9.999999E+96", want: 96},
		ilogbFiniteCase{input: "1E-101", want: -101},
	)
	for _, tc := range cases {
		got, flags := mustDecimal32BID(t, tc.input).ILogB()
		if got != tc.want || flags != 0 {
			t.Errorf("Decimal32BID(%s).ILogB() = %d/%s, want %d/None", tc.input, got, flags, tc.want)
		}
	}
	for _, tc := range ilogbSharedSpecialCases {
		got, flags := mustDecimal32BID(t, tc.input).ILogB()
		if got != tc.want || flags != FlagInvalidOperation {
			t.Errorf("Decimal32BID(%s).ILogB() = %d/%s, want %d/%s", tc.input, got, flags, tc.want, FlagInvalidOperation)
		}
	}
}

func TestDecimal64ILogB(t *testing.T) {
	cases := append([]ilogbFiniteCase{}, ilogbSharedFiniteCases...)
	cases = append(cases,
		// Decimal64: emax 384, p 16 -> largest finite 9.999999999999999E+384,
		// smallest subnormal 1E-398.
		ilogbFiniteCase{input: "9.999999999999999E+384", want: 384},
		ilogbFiniteCase{input: "1E-398", want: -398},
	)
	for _, tc := range cases {
		got, flags := mustDecimal64BID(t, tc.input).ILogB()
		if got != tc.want || flags != 0 {
			t.Errorf("Decimal64BID(%s).ILogB() = %d/%s, want %d/None", tc.input, got, flags, tc.want)
		}
	}
	for _, tc := range ilogbSharedSpecialCases {
		got, flags := mustDecimal64BID(t, tc.input).ILogB()
		if got != tc.want || flags != FlagInvalidOperation {
			t.Errorf("Decimal64BID(%s).ILogB() = %d/%s, want %d/%s", tc.input, got, flags, tc.want, FlagInvalidOperation)
		}
	}
}

func TestDecimal128ILogB(t *testing.T) {
	cases := append([]ilogbFiniteCase{}, ilogbSharedFiniteCases...)
	cases = append(cases,
		// Decimal128: emax 6144, p 34 -> largest finite
		// 9.999999999999999999999999999999999E+6144, smallest subnormal
		// 1E-6176.
		ilogbFiniteCase{input: "9.999999999999999999999999999999999E+6144", want: 6144},
		ilogbFiniteCase{input: "1E-6176", want: -6176},
	)
	for _, tc := range cases {
		got, flags := mustDecimal128BID(t, tc.input).ILogB()
		if got != tc.want || flags != 0 {
			t.Errorf("Decimal128BID(%s).ILogB() = %d/%s, want %d/None", tc.input, got, flags, tc.want)
		}
	}
	for _, tc := range ilogbSharedSpecialCases {
		got, flags := mustDecimal128BID(t, tc.input).ILogB()
		if got != tc.want || flags != FlagInvalidOperation {
			t.Errorf("Decimal128BID(%s).ILogB() = %d/%s, want %d/%s", tc.input, got, flags, tc.want, FlagInvalidOperation)
		}
	}
}
