package bid754_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

func TestPublicParserConvenienceWidths(t *testing.T) {
	t.Run("Decimal32", func(t *testing.T) {
		testPublicParserConvenienceWidth(t, 7, -101, 90, bid754.NewDecimal32, bid754.NewDecimal32BIDDirect, bid754.NewDecimal32WithFlags, bid754.NewDecimal32WithMode, bid754.Zero32BID(), bid754.AddSlice32BID, bid754.AddSlice32BIDWithFlags)
	})
	t.Run("Decimal64", func(t *testing.T) {
		testPublicParserConvenienceWidth(t, 16, -398, 369, bid754.NewDecimal64, bid754.NewDecimal64BIDDirect, bid754.NewDecimal64WithFlags, bid754.NewDecimal64WithMode, bid754.Zero64BID(), bid754.AddSlice64BID, bid754.AddSlice64BIDWithFlags)
	})
	t.Run("Decimal128", func(t *testing.T) {
		testPublicParserConvenienceWidth(t, 34, -6176, 6111, bid754.NewDecimal128, bid754.NewDecimal128BIDDirect, bid754.NewDecimal128WithFlags, bid754.NewDecimal128WithMode, bid754.Zero128BID(), bid754.AddSlice128BID, bid754.AddSlice128BIDWithFlags)
	})
}

func testPublicParserConvenienceWidth[T comparable](t *testing.T, precision, minQ, maxQ int,
	parse, direct func(string) (T, error),
	withFlags func(string) (T, bid754.ExceptionFlags, error),
	withMode func(string, bid754.RoundingMode) (T, bid754.ExceptionFlags, error),
	zero T, sum func([]T) T, sumFlags func([]T) (T, bid754.ExceptionFlags)) {
	t.Helper()
	modes := []bid754.RoundingMode{bid754.RoundNearestEven, bid754.RoundNearestAway, bid754.RoundTowardZero, bid754.RoundTowardPositive, bid754.RoundTowardNegative}
	mustParse := func(s string) T {
		t.Helper()
		v, err := parse(s)
		if err != nil {
			t.Fatalf("parse(%q): %v", s, err)
		}
		return v
	}
	for _, values := range [][]T{nil, {}} {
		if got := sum(values); got != zero || got != mustParse("0") {
			t.Fatalf("empty sum changed the zero representation: %v", got)
		}
		if got, flags := sumFlags(values); got != zero || flags != 0 {
			t.Fatalf("empty sum with flags = %v, %v", got, flags)
		}
	}
	negativeZero := mustParse("-0.00")
	if got, flags := sumFlags([]T{negativeZero}); got != negativeZero || flags != 0 || sum([]T{negativeZero}) != negativeZero {
		t.Fatal("singleton sum changed signed-zero cohort")
	}
	for _, q := range []int{minQ, minQ + 1, 0, maxQ - 1, maxQ} {
		for _, coefficient := range []string{"0", "1", strings.Repeat("9", precision)} {
			s := coefficient + "e" + strconv.Itoa(q)
			want := mustParse(s)
			for _, input := range []string{s, " \t+" + strings.Repeat("0", 4096) + s} {
				if got, err := direct(input); err != nil || got != want {
					t.Fatalf("direct leading-zero boundary q=%d: %v, %v", q, got, err)
				}
				if got, flags, err := withFlags(input); err != nil || flags != 0 || got != want {
					t.Fatalf("flag boundary q=%d: %v, %v, %v", q, got, flags, err)
				}
				for _, mode := range modes {
					if got, flags, err := withMode(input, mode); err != nil || flags != 0 || got != want {
						t.Fatalf("mode boundary q=%d mode=%v: %v, %v, %v", q, mode, got, flags, err)
					}
				}
			}
		}
	}
	for _, kind := range []string{"NaN", "-sNaN", "+qNaN"} {
		for _, payload := range []string{"", "0", strings.Repeat("9", precision-1)} {
			want := mustParse(kind + payload)
			input := " \t" + kind + strings.Repeat("0", 4096) + payload
			if got, flags, err := withFlags(input); err != nil || flags != 0 || got != want {
				t.Fatalf("NaN leading-zero payload: %v, %v, %v", got, flags, err)
			}
			if got, err := direct(input); err != nil || got != want {
				t.Fatalf("direct NaN payload: %v, %v", got, err)
			}
		}
	}
	rejected := []string{"", ".", ".e1", "1e5x", "1 ", "+ 1", "\n1", "1\x00", "Inf ", "nan(1)", "ſNaN", "NaN" + strings.Repeat("9", precision), "1" + strings.Repeat("0", precision), "0e" + strconv.Itoa(minQ-1), "0e" + strconv.Itoa(maxQ+1)}
	for _, suffix := range []string{"x", ":", "\x00", "９", " ", "e1x"} {
		rejected = append(rejected, strings.Repeat("7", precision+40)+suffix)
	}
	for _, input := range rejected {
		var empty T
		if got, err := parse(input); err == nil || got != empty {
			t.Fatalf("exact rejection(%q): %v, %v", input, got, err)
		}
		if got, err := direct(input); err == nil || got != empty {
			t.Fatalf("direct rejection(%q): %v, %v", input, got, err)
		}
		if got, flags, err := withFlags(input); err == nil || flags != 0 || got != empty {
			t.Fatalf("flag rejection(%q): %v, %v, %v", input, got, flags, err)
		}
		for _, mode := range append(modes[:len(modes):len(modes)], bid754.RoundingMode(255)) {
			if got, flags, err := withMode(input, mode); err == nil || flags != 0 || got != empty {
				t.Fatalf("mode rejection(%q, %v): %v, %v, %v", input, mode, got, flags, err)
			}
		}
	}
	for _, sign := range []string{"", "-"} {
		for _, extraDigits := range []int{40, 4096} {
			input := sign + "1." + strings.Repeat("0", precision-1) + "51" + strings.Repeat("0", extraDigits)
			down := mustParse(sign + "1." + strings.Repeat("0", precision-1))
			up := mustParse(sign + "1." + strings.Repeat("0", precision-2) + "1")
			if _, err := parse(input); err == nil {
				t.Fatal("exact constructor accepted inexact long coefficient")
			}
			if got, flags, err := withFlags(input); err != nil || flags != bid754.FlagInexact || got != up {
				t.Fatalf("long coefficient nearest: %v, %v, %v", got, flags, err)
			}
			for _, mode := range modes {
				want := down
				if mode == bid754.RoundNearestEven || mode == bid754.RoundNearestAway || mode == bid754.RoundTowardPositive && sign == "" || mode == bid754.RoundTowardNegative && sign == "-" {
					want = up
				}
				if got, flags, err := withMode(input, mode); err != nil || flags != bid754.FlagInexact || got != want {
					t.Fatalf("long coefficient mode=%v sign=%q: %v, %v, %v; want %v", mode, sign, got, flags, err, want)
				}
			}
			if got, flags, err := withMode(input, bid754.RoundingMode(255)); err != nil || flags != bid754.FlagInvalidOperation || got != mustParse("NaN") {
				t.Fatalf("bad mode on accepted rounded input: %v, %v, %v", got, flags, err)
			}
		}
	}
	for _, input := range []string{strings.Repeat("x", 1<<16), strings.Repeat("\x00\xff", 1<<15), "NaN" + strings.Repeat("9", 1<<16), strings.Repeat("7", 1<<16) + "x"} {
		checkError := func(err error) {
			t.Helper()
			if err == nil {
				t.Fatal("large invalid input accepted")
			}
			message := err.Error()
			if len(message) > 256 || !strings.Contains(message, fmt.Sprintf("%d bytes", len(input))) || !strings.Contains(message, strconv.Quote(input[:32])) {
				t.Fatalf("error must have bounded prefix and original byte length: length=%d", len(message))
			}
		}
		_, err := parse(input)
		checkError(err)
		_, err = direct(input)
		checkError(err)
		_, _, err = withFlags(input)
		checkError(err)
		for _, mode := range []bid754.RoundingMode{bid754.RoundNearestEven, bid754.RoundingMode(255)} {
			_, _, err = withMode(input, mode)
			checkError(err)
		}
	}
}

func TestPublicParserConvenienceAnyWidthEquivalence(t *testing.T) {
	inputs := []string{"", ".", "1e", "1e5x", "1 ", "Inf", "-sNaN", "NaN123", "nan(123)", "\n1", "1e18446744073709551616", "-0.00", " \t1"}
	for _, precision := range []int{7, 16, 34} {
		for _, digits := range []int{precision - 1, precision, precision + 1} {
			for _, q := range []int{-6177, -6176, -399, -398, -102, -101, 0, 90, 91, 369, 370, 6111, 6112} {
				for _, coefficient := range []string{"0", strings.Repeat("9", digits), "1" + strings.Repeat("0", digits-1)} {
					inputs = append(inputs, coefficient+"e"+strconv.Itoa(q), " \t-000"+coefficient+"e"+strconv.Itoa(q))
				}
			}
			inputs = append(inputs, "NaN"+strings.Repeat("9", digits), "-sNaN000"+strings.Repeat("9", digits))
		}
	}
	for _, fractionalDigits := range []int{999999, 1000000, 1048576} {
		for _, exponentZeros := range []string{"", "0", "00000000"} {
			s := "0." + strings.Repeat("0", fractionalDigits-1) + "1e" + exponentZeros + strconv.Itoa(fractionalDigits)
			value, err := bid754.NewDecimal64(s)
			if err != nil || value.ToUint64() != bid754.One64BID().ToUint64() {
				t.Fatalf("exact exponent cancellation %d/%d: value=%v error=%v", fractionalDigits, len(exponentZeros), value, err)
			}
			inputs = append(inputs, s)
		}
	}
	for _, s := range inputs {
		_, err32 := bid754.NewDecimal32(s)
		_, err64 := bid754.NewDecimal64(s)
		_, err128 := bid754.NewDecimal128(s)
		want := err32 == nil || err64 == nil || err128 == nil
		if got := bid754.IsValidDecimalString(s); got != want {
			t.Fatalf("IsValidDecimalString(%q)=%v; errors=%v / %v / %v", s, got, err32, err64, err128)
		}
	}
}

func TestPublicParserConvenienceMinimumPrecision(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"0", 1}, {"-0.000e-99999999999999999999999999", 1},
		{" \t+0001200300.000E+18446744073709551616", 5},
		{"1.2300000", 3}, {"10000000", 1}, {"NaN" + strings.Repeat("9", 4096), 1},
		{"-sNaN" + strings.Repeat("0", 4096) + "12", 1}, {"+INFINITY", 1},
		{"0." + strings.Repeat("0", 4096) + "12003" + strings.Repeat("0", 4096), 5},
	}
	for _, precision := range []int{7, 16, 34, 4096} {
		cases = append(cases, struct {
			input string
			want  int
		}{"000" + strings.Repeat("7", precision) + "000e-999999", precision})
	}
	for _, tc := range cases {
		if got, err := bid754.GetRequiredPrecision(tc.input); err != nil || got != tc.want {
			t.Fatalf("precision for %d bytes: %d, %v; want %d", len(tc.input), got, err, tc.want)
		}
		allocs := testing.AllocsPerRun(10, func() {
			_, _ = bid754.GetRequiredPrecision(tc.input)
		})
		if allocs != 0 {
			t.Fatalf("precision for %d bytes allocated %v objects", len(tc.input), allocs)
		}
	}
	for _, input := range []string{"", ".", "1e", "1 ", "\n1", "1e" + strings.Repeat("9", 4096) + "x", strings.Repeat("\xff", 4096)} {
		got, err := bid754.GetRequiredPrecision(input)
		if got != 0 || err == nil {
			t.Fatal("malformed precision query accepted")
		}
		if len(err.Error()) > 256 || len(input) > 32 && !strings.Contains(err.Error(), fmt.Sprintf("%d bytes", len(input))) {
			t.Fatalf("precision error is not bounded: %d bytes", len(err.Error()))
		}
	}
}
