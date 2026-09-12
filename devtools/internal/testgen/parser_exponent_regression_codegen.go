package testgen

import (
	"fmt"
	"math/big"
	"strings"
)

type parserExpRegressionCase struct {
	ID              string
	Width           int
	Neg             bool
	FracZeros       int
	ExpLeadingZeros int
	IntegerZeros    int
	ExpMag          int

	Quantum  int
	ExpectHi uint64
	ExpectLo uint64
}

func parserExpWidthBias(width int) int {
	switch width {
	case 32:
		return 101
	case 64:
		return 398
	default:
		return 6176
	}
}

func parserExpRefEncode(width int, neg bool, q int) (hi, lo uint64) {
	bias := parserExpWidthBias(width)
	switch width {
	case 32:
		lo = (uint64(q+bias) << 23) | 1
		if neg {
			lo |= 1 << 31
		}
	case 64:
		lo = (uint64(q+bias) << 53) | 1
		if neg {
			lo |= 1 << 63
		}
	default:
		hi = uint64(q+bias) << 49
		lo = 1
		if neg {
			hi |= 1 << 63
		}
	}
	return
}

func parserExponentRegressionCases() []parserExpRegressionCase {
	cases := []parserExpRegressionCase{
		{ID: "d32_q0_short", Width: 32, FracZeros: 4, ExpMag: 5},
		{ID: "d64_q0_short", Width: 64, FracZeros: 4, ExpMag: 5},
		{ID: "d128_q0_short", Width: 128, FracZeros: 4, ExpMag: 5},
		{ID: "d128_q0_short_neg", Width: 128, Neg: true, FracZeros: 4, ExpMag: 5},

		{ID: "d128_lz0_7dig", Width: 128, FracZeros: 999999, ExpLeadingZeros: 0, ExpMag: 1000000},
		{ID: "d128_lz1_7dig_confirmed", Width: 128, FracZeros: 999999, ExpLeadingZeros: 1, ExpMag: 1000000},
		{ID: "d128_lz2_7dig", Width: 128, FracZeros: 999999, ExpLeadingZeros: 2, ExpMag: 1000000},

		{ID: "d32_n10485760", Width: 32, FracZeros: 10485759, ExpMag: 10485760},
		{ID: "d64_n10485760", Width: 64, FracZeros: 10485759, ExpMag: 10485760},
		{ID: "d128_n10485760", Width: 128, FracZeros: 10485759, ExpMag: 10485760},
	}
	for _, width := range []int{32, 64, 128} {
		for _, n := range []int{999999, 9999999, 10000000, 10485759, 10485761} {
			cases = append(cases, parserExpRegressionCase{ID: fmt.Sprintf("d%d_neighbor_%d", width, n), Width: width, FracZeros: n - 1, ExpMag: n})
		}
		for _, n := range []int{1000000, 10485760} {
			cases = append(cases, parserExpRegressionCase{ID: fmt.Sprintf("d%d_negative_%d", width, n), Width: width, Neg: true, FracZeros: n - 1, ExpLeadingZeros: 1, ExpMag: n})
			for _, neg := range []bool{false, true} {
				cases = append(cases, parserExpRegressionCase{ID: fmt.Sprintf("d%d_integer_%d_negative_%v", width, n, neg), Width: width, Neg: neg, IntegerZeros: n, ExpLeadingZeros: 1, ExpMag: -n})
			}
		}
		minQ := -parserExpWidthBias(width)
		maxQ := map[int]int{32: 90, 64: 369, 128: 6111}[width]
		for _, q := range []int{minQ, maxQ} {
			cases = append(cases, parserExpRegressionCase{ID: fmt.Sprintf("d%d_quantum_%d", width, q), Width: width, FracZeros: 1000000 - q - 1, ExpLeadingZeros: 1, ExpMag: 1000000})
		}
	}

	for i := range cases {
		c := &cases[i]
		f := c.FracZeros + 1
		c.Quantum = c.ExpMag - f
		c.ExpectHi, c.ExpectLo = parserExpRefEncode(c.Width, c.Neg, c.Quantum)
		if c.IntegerZeros > 0 {
			precision := map[int]int{32: 7, 64: 16, 128: 34}[c.Width]
			c.Quantum = 1 - precision
			coeff := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(precision-1)), nil)
			c.ExpectHi, c.ExpectLo = parserExpRefEncode(c.Width, c.Neg, c.Quantum)
			if c.Width == 128 {
				c.ExpectHi |= new(big.Int).Rsh(new(big.Int).Set(coeff), 64).Uint64()
			}
			c.ExpectLo = (c.ExpectLo &^ 1) | coeff.Uint64()
		}
	}
	return cases
}

func parserExpRegressionGoSource(cases []parserExpRegressionCase) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("func TestGeneratedParserExponentCancellationRegression(t *testing.T) {\n")
	b.WriteString("\ttype expCancelCase struct {\n")
	b.WriteString("\t\tid       string\n\t\twidth    int\n\t\tneg      bool\n\t\tfracZeros int\n\t\texpLZ    int\n\t\texpMag   int\n\t\tintegerZeros int\n\t\thi, lo   uint64\n\t}\n")
	b.WriteString("\tcases := []expCancelCase{\n")
	for _, c := range cases {
		fmt.Fprintf(&b, "\t\t{%q, %d, %t, %d, %d, %d, %d, %#x, %#x},\n",
			c.ID, c.Width, c.Neg, c.FracZeros, c.ExpLeadingZeros, c.ExpMag, c.IntegerZeros, c.ExpectHi, c.ExpectLo)
	}
	b.WriteString("\t}\n")
	b.WriteString("\tbuild := func(c expCancelCase) string {\n")
	b.WriteString("\t\tvar sb strings.Builder\n")
	b.WriteString("\t\tif c.neg {\n\t\t\tsb.WriteByte('-')\n\t\t}\n")
	b.WriteString(`
        if c.integerZeros > 0 {
            sb.WriteByte('1')
            sb.WriteString(strings.Repeat("0", c.integerZeros))
            sb.WriteByte('e')
        } else {
            sb.WriteString("0.")
            sb.WriteString(strings.Repeat("0", c.fracZeros))
            sb.WriteString("1e+")
        }
        exp := c.expMag
        if exp < 0 { sb.WriteByte('-'); exp = -exp }
        sb.WriteString(strings.Repeat("0", c.expLZ))
        sb.WriteString(strconv.Itoa(exp))
`)
	b.WriteString("\t\treturn sb.String()\n\t}\n")
	b.WriteString("\tmodes := []int{BID_ROUNDING_TO_NEAREST, BID_ROUNDING_DOWN, BID_ROUNDING_UP, BID_ROUNDING_TO_ZERO, BID_ROUNDING_TIES_AWAY}\n")
	b.WriteString("\tfor _, c := range cases {\n")
	b.WriteString("\t\ts := build(c)\n")
	b.WriteString("\t\tfor _, m := range modes {\n")
	b.WriteString("\t\t\tvar hi, lo uint64\n\t\t\tvar fl uint32\n")
	b.WriteString("\t\t\tswitch c.width {\n")
	b.WriteString("\t\t\tcase 32:\n\t\t\t\tb, f := Bid32FromStringRaw(s, m)\n\t\t\t\thi, lo, fl = 0, uint64(b), f\n")
	b.WriteString("\t\t\tcase 64:\n\t\t\t\tb, f := Bid64FromString(s, m)\n\t\t\t\thi, lo, fl = 0, b, f\n")
	b.WriteString("\t\t\tdefault:\n\t\t\t\tr, f := Bid128FromString(s, m)\n\t\t\t\thi, lo = Bid128Words(r)\n\t\t\t\tfl = f\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t\tif hi != c.hi || lo != c.lo || fl != 0 {\n")
	b.WriteString("\t\t\t\tt.Errorf(\"%s mode=%d: got %#x,%#x fl=%#x; want exact cohort %#x,%#x fl=0\", c.id, m, hi, lo, fl, c.hi, c.lo)\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t}\n")
	b.WriteString("}\n")
	return b.String()
}

func parserExpRegressionRustSource(cases []parserExpRegressionCase) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("#[test]\n")
	b.WriteString("fn test_generated_parser_exponent_cancellation_regression() {\n")
	b.WriteString("    struct ExpCancelCase {\n")
	b.WriteString("        id: &'static str,\n        width: u32,\n        neg: bool,\n        frac_zeros: usize,\n        exp_lz: usize,\n        exp_mag: i64,\n        integer_zeros: usize,\n        hi: u64,\n        lo: u64,\n    }\n")
	b.WriteString("    let cases = [\n")
	for _, c := range cases {
		fmt.Fprintf(&b, "        ExpCancelCase { id: %q, width: %d, neg: %t, frac_zeros: %d, exp_lz: %d, exp_mag: %d, integer_zeros: %d, hi: %#x, lo: %#x },\n",
			c.ID, c.Width, c.Neg, c.FracZeros, c.ExpLeadingZeros, c.ExpMag, c.IntegerZeros, c.ExpectHi, c.ExpectLo)
	}
	b.WriteString("    ];\n")
	b.WriteString("    let build = |c: &ExpCancelCase| -> String {\n")
	b.WriteString("        let mut s = String::new();\n")
	b.WriteString("        if c.neg { s.push('-'); }\n")
	b.WriteString(`
        if c.integer_zeros > 0 {
            s.push('1');
            s.push_str(&"0".repeat(c.integer_zeros));
            s.push('e');
        } else {
            s.push_str("0.");
            s.push_str(&"0".repeat(c.frac_zeros));
            s.push_str("1e+");
        }
        if c.exp_mag < 0 { s.push('-'); }
        s.push_str(&"0".repeat(c.exp_lz));
        s.push_str(&c.exp_mag.abs().to_string());
`)
	b.WriteString("        s\n    };\n")
	b.WriteString("    for c in cases.iter() {\n")
	b.WriteString("        let input = build(c);\n")
	b.WriteString("        for &mode in &[0i64, 1, 2, 3, 4] {\n")
	b.WriteString("            let (hi, lo, flags) = match c.width {\n")
	b.WriteString("                32 => { let (bits, f) = bid32_from_string_raw(&input, mode); (0u64, bits as u64, f) }\n")
	b.WriteString("                64 => { let (bits, f) = bid64_from_string_raw(&input, mode as i32); (0u64, bits, f) }\n")
	b.WriteString("                _ => { let (r, f) = bid128_from_string(&input, mode); (r.hi, r.lo, f) }\n")
	b.WriteString("            };\n")
	b.WriteString("            assert_eq!(\n")
	b.WriteString("                (hi, lo, flags),\n")
	b.WriteString("                (c.hi, c.lo, 0u32),\n")
	b.WriteString("                \"{} mode={} exact cohort\",\n                c.id,\n                mode\n")
	b.WriteString("            );\n")
	b.WriteString("        }\n")
	b.WriteString(`
        macro_rules! check_public {
            ($ty:ty, $bits:expr) => {{
                let bits = $bits;
                if c.integer_zeros > 0 {
                    assert!(<$ty>::parse(&input).is_err(), "{} exact rejection", c.id);
                    assert!(<$ty>::parse_with_flags(&input).is_err(), "{} flag rejection", c.id);
                    let (v, f) = <$ty>::parse_raw(&input);
                    assert!(v.is_nan() && f.contains(bid754::ExceptionFlags::INVALID_OPERATION), "{} raw rejection", c.id);
                } else {
                    let v = <$ty>::parse(&input).expect(c.id);
                    assert_eq!(bits(v), (c.hi, c.lo), "{} exact public", c.id);
                    let (v, f) = <$ty>::parse_with_flags(&input).expect(c.id);
                    assert_eq!((bits(v), f.bits()), ((c.hi, c.lo), 0), "{} flags public", c.id);
                    let (v, f) = <$ty>::parse_raw(&input);
                    assert_eq!((bits(v), f.bits()), ((c.hi, c.lo), 0), "{} raw public", c.id);
                }
                for mode in [bid754::RoundingMode::NearestEven, bid754::RoundingMode::NearestAway,
                    bid754::RoundingMode::TowardZero, bid754::RoundingMode::TowardPositive,
                    bid754::RoundingMode::TowardNegative] {
                    let result = <$ty>::parse_with_mode(&input, mode);
                    if c.integer_zeros > 0 {
                        assert!(result.is_err(), "{} mode {:?} rejection", c.id, mode);
                    } else {
                        let (v, f) = result.expect(c.id);
                        assert_eq!((bits(v), f.bits()), ((c.hi, c.lo), 0), "{} mode {:?} public", c.id, mode);
                    }
                }
            }};
        }
        match c.width {
            32 => check_public!(bid754::Decimal32, |v: bid754::Decimal32| (0u64, v.to_bits() as u64)),
            64 => check_public!(bid754::Decimal64, |v: bid754::Decimal64| (0u64, v.to_bits())),
            _ => check_public!(bid754::Decimal128, |v: bid754::Decimal128| {
                let n = u128::from_le_bytes(v.to_le_bytes()); ((n >> 64) as u64, n as u64)
            }),
        }
`)

	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}
