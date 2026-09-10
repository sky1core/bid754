package bid754_test

import (
	"strings"
	"testing"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

var (
	publicParserConvenience32    bid754.Decimal32BID
	publicParserConvenience64    bid754.Decimal64BID
	publicParserConvenience128   bid754.Decimal128BID
	publicParserConvenienceFlags bid754.ExceptionFlags
	publicParserConvenienceError error
	publicParserConvenienceBool  bool
	publicParserConvenienceInt   int
)

func BenchmarkPublicParserConvenienceParse(b *testing.B) {
	b.Run("Decimal32", func(b *testing.B) {
		benchmarkPublicParserConvenienceParse(b, bid754.NewDecimal32, bid754.NewDecimal32WithFlags, &publicParserConvenience32)
	})
	b.Run("Decimal64", func(b *testing.B) {
		benchmarkPublicParserConvenienceParse(b, bid754.NewDecimal64, bid754.NewDecimal64WithFlags, &publicParserConvenience64)
	})
	b.Run("Decimal128", func(b *testing.B) {
		benchmarkPublicParserConvenienceParse(b, bid754.NewDecimal128, bid754.NewDecimal128WithFlags, &publicParserConvenience128)
	})
}

func benchmarkPublicParserConvenienceParse[T comparable](b *testing.B, parse func(string) (T, error), withFlags func(string) (T, bid754.ExceptionFlags, error), sink *T) {
	for _, tc := range []struct {
		name  string
		input string
		valid bool
	}{
		{"Short", "123.4500", true},
		{"LeadingZeros", strings.Repeat("0", 4096) + "123.4500", true},
		{"NaNLeadingZeros", "-sNaN" + strings.Repeat("0", 4096) + "123", true},
		{"NaNPayloadOverflow", "NaN" + strings.Repeat("9", 4096), false},
		{"MalformedLongCoefficient", strings.Repeat("7", 4096) + "x", false},
	} {
		b.Run(tc.name, func(b *testing.B) {
			if _, err := parse(tc.input); (err == nil) != tc.valid {
				b.Fatalf("fixture acceptance: %v", err)
			}
			b.Run("Exact", func(b *testing.B) {
				var result T
				var err error
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, err = parse(tc.input)
				}
				*sink, publicParserConvenienceError = result, err
			})
			b.Run("WithFlags", func(b *testing.B) {
				if _, flags, err := withFlags(tc.input); (err == nil) != tc.valid || flags != 0 {
					b.Fatalf("fixture flags: %v, %v", flags, err)
				}
				var result T
				var flags bid754.ExceptionFlags
				var err error
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					result, flags, err = withFlags(tc.input)
				}
				*sink, publicParserConvenienceFlags, publicParserConvenienceError = result, flags, err
			})
		})
	}
}

func BenchmarkPublicParserConvenienceIsValid(b *testing.B) {
	for _, tc := range []struct {
		name  string
		input string
		valid bool
	}{
		{"Precision7", "1234567", true},
		{"Precision16", "1234567890123456", true},
		{"Precision34", "1234567890123456789012345678901234", true},
		{"LeadingZeros", strings.Repeat("0", 4096) + "1", true},
		{"Malformed", strings.Repeat("7", 4096) + "x", false},
		{"NaNPayloadOverflow", "NaN" + strings.Repeat("9", 4096), false},
	} {
		b.Run(tc.name, func(b *testing.B) {
			if bid754.IsValidDecimalString(tc.input) != tc.valid {
				b.Fatal("fixture acceptance")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				publicParserConvenienceBool = bid754.IsValidDecimalString(tc.input)
			}
		})
	}
}

func BenchmarkPublicParserConveniencePrecision(b *testing.B) {
	for _, tc := range []struct {
		name  string
		input string
		want  int
	}{
		{"Short", "1.2300000", 3},
		{"LeadingTrailingZeros", "0." + strings.Repeat("0", 4096) + "12003" + strings.Repeat("0", 4096), 5},
		{"HugeExponent", "123.4500e" + strings.Repeat("9", 4096), 5},
		{"NaNPayload", "-sNaN" + strings.Repeat("9", 4096), 1},
	} {
		b.Run(tc.name, func(b *testing.B) {
			if got, err := bid754.GetRequiredPrecision(tc.input); err != nil || got != tc.want {
				b.Fatalf("fixture precision: %d, %v", got, err)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				publicParserConvenienceInt, publicParserConvenienceError = bid754.GetRequiredPrecision(tc.input)
			}
		})
	}
}

func BenchmarkPublicParserConvenienceEmptySum(b *testing.B) {
	b.Run("Decimal32", func(b *testing.B) {
		benchmarkPublicParserConvenienceEmptySum(b, bid754.AddSlice32BID, bid754.AddSlice32BIDWithFlags, bid754.Zero32BID(), &publicParserConvenience32)
	})
	b.Run("Decimal64", func(b *testing.B) {
		benchmarkPublicParserConvenienceEmptySum(b, bid754.AddSlice64BID, bid754.AddSlice64BIDWithFlags, bid754.Zero64BID(), &publicParserConvenience64)
	})
	b.Run("Decimal128", func(b *testing.B) {
		benchmarkPublicParserConvenienceEmptySum(b, bid754.AddSlice128BID, bid754.AddSlice128BIDWithFlags, bid754.Zero128BID(), &publicParserConvenience128)
	})
}

func benchmarkPublicParserConvenienceEmptySum[T comparable](b *testing.B, sum func([]T) T, sumFlags func([]T) (T, bid754.ExceptionFlags), zero T, sink *T) {
	if got, flags := sumFlags(nil); got != zero || flags != 0 || sum(nil) != zero {
		b.Fatal("empty sum representation")
	}
	b.Run("Value", func(b *testing.B) {
		var result T
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result = sum(nil)
		}
		*sink = result
	})
	b.Run("WithFlags", func(b *testing.B) {
		var result T
		var flags bid754.ExceptionFlags
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			result, flags = sumFlags(nil)
		}
		*sink, publicParserConvenienceFlags = result, flags
	})
}
