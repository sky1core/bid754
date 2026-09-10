package bid754_test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	bid754 "github.com/sky1core/bid754/bid754-go"
)

func TestPublicParserResourceBounds(t *testing.T) {
	for _, size := range []int{34, 6176, 61760, 1 << 20} {
		zeros := strings.Repeat("0", size)
		for _, tc := range []struct {
			name  string
			input string
		}{
			{"leading-zero", zeros + "1"},
			{"fraction-exponent-cancellation", "0." + zeros + "1e" + fmt.Sprint(size+1)},
			{"rounded-coefficient", "1.234567891234567891234567891234567895" + zeros},
			{"invalid-prefix", "x" + zeros},
			{"invalid-suffix", zeros + "1x"},
			{"nan-payload", "NaN" + zeros + "1"},
			{"invalid-payload", "NaN" + zeros + "!"},
			{"oversized-coefficient", strings.Repeat("9", size)},
			{"oversized-exponent", "1e" + strings.Repeat("9", size)},
		} {
			t.Run(fmt.Sprintf("%s/%d", tc.name, size), func(t *testing.T) {
				checkPublicParserResourceWidth(t, "32", tc.input, bid754.NewDecimal32, bid754.NewDecimal32WithFlags, bid754.NewDecimal32WithMode, bid754.ParseDecimal32BIDRaw)
				checkPublicParserResourceWidth(t, "64", tc.input, bid754.NewDecimal64, bid754.NewDecimal64WithFlags, bid754.NewDecimal64WithMode, bid754.ParseDecimal64BIDRaw)
				checkPublicParserResourceWidth(t, "128", tc.input, bid754.NewDecimal128, bid754.NewDecimal128WithFlags, bid754.NewDecimal128WithMode, bid754.ParseDecimal128BIDRaw)
				checkPublicParserResourceCall(t, "valid", tc.input, func() error {
					valid := bid754.IsValidDecimalString(tc.input)
					runtime.KeepAlive(valid)
					return nil
				})
				checkPublicParserResourceCall(t, "precision", tc.input, func() error {
					precision, err := bid754.GetRequiredPrecision(tc.input)
					runtime.KeepAlive(precision)
					return err
				})
			})
		}
	}
}

func checkPublicParserResourceWidth[T any](t *testing.T, width, input string,
	exact func(string) (T, error),
	flags func(string) (T, bid754.ExceptionFlags, error),
	mode func(string, bid754.RoundingMode) (T, bid754.ExceptionFlags, error),
	raw func(string) (T, bid754.ExceptionFlags)) {
	t.Helper()
	checkPublicParserResourceCall(t, width+"/exact", input, func() error {
		value, err := exact(input)
		runtime.KeepAlive(value)
		return err
	})
	checkPublicParserResourceCall(t, width+"/flags", input, func() error {
		value, status, err := flags(input)
		runtime.KeepAlive(value)
		runtime.KeepAlive(status)
		return err
	})
	for _, rounding := range []bid754.RoundingMode{bid754.RoundNearestEven, bid754.RoundNearestAway, bid754.RoundTowardZero, bid754.RoundTowardPositive, bid754.RoundTowardNegative, bid754.RoundingMode(255)} {
		checkPublicParserResourceCall(t, fmt.Sprintf("%s/mode-%d", width, rounding), input, func() error {
			value, status, err := mode(input, rounding)
			runtime.KeepAlive(value)
			runtime.KeepAlive(status)
			return err
		})
	}
	checkPublicParserResourceCall(t, width+"/raw", input, func() error {
		value, status := raw(input)
		runtime.KeepAlive(value)
		runtime.KeepAlive(status)
		return nil
	})
}

func checkPublicParserResourceCall(t *testing.T, name, input string, call func() error) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		for range 4 {
			_ = call()
		}
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		err := call()
		runtime.ReadMemStats(&after)
		if allocated := after.TotalAlloc - before.TotalAlloc; allocated > 32<<10 {
			t.Fatalf("parser allocation budget exceeded: input=%d allocated=%d budget=%d", len(input), allocated, 32<<10)
		}
		if err != nil && len(err.Error()) > 256 {
			t.Fatalf("parser error budget exceeded: input=%d error=%d budget=256", len(input), len(err.Error()))
		}
	})
}
