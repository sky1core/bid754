package bidcodec_test

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	codec "github.com/sky1core/bid754/bid754-codec-go"
)

func TestFromStringResourceBoundaryRejects(t *testing.T) {
	checkResourceRejects(t)
}

func checkResourceRejects(t *testing.T) int {
	cases := 0
	for _, size := range []int{35, 256, 4096, 6176, 6177, 61760, 65536, 1048576} {
		digits := strings.Repeat("9", size)
		zeros := strings.Repeat("0", size)
		for _, tc := range []struct {
			name  string
			input string
		}{
			{"coefficient", digits},
			{"fractional_coefficient", "0." + digits},
			{"trailing_zero_coefficient", "1" + zeros},
			{"nan_payload", "NaN" + digits},
			{"snan_payload", "sNaN" + digits},
			{"invalid_nan_payload", "NaN" + zeros + "x"},
			{"positive_exponent", "1E+" + digits},
			{"negative_exponent", "1e-" + digits},
			{"malformed_exponent", "1E" + zeros + "x"},
			{"whitespace_exponent", "1E" + zeros + "\t1"},
		} {
			cases++
			t.Run(fmt.Sprintf("%s/%d", tc.name, size), func(t *testing.T) {
				_, err, allocated := measureFromString(tc.input)

				if err == nil {
					t.Fatal("oversized or malformed input accepted")
				}
				if allocated > 32<<10 {
					t.Errorf("FromString allocated %d bytes, limit %d", allocated, 32<<10)
				}
				if n := len(err.Error()); n > 256 {
					t.Errorf("error has %d bytes, limit 256", n)
				}
				t.Logf("allocated=%d bytes error=%d bytes", allocated, len(err.Error()))
			})
		}
	}
	return cases
}

func TestFromStringResourceBoundaryLeadingZeros(t *testing.T) {
	checkResourceLeadingZeros(t)
}

func checkResourceLeadingZeros(t *testing.T) int {
	cases := 0
	for _, size := range []int{0, 34, 256, 4096, 6176, 6177, 61760, 65536, 1048576} {
		zeros := strings.Repeat("0", size)
		for _, tc := range []struct {
			name        string
			input       string
			kind        codec.Kind
			sign        bool
			coefficient string
			exponent    int32
			payload     string
		}{
			{"coefficient_limit", zeros + strings.Repeat("9", 34), codec.Normal, false, strings.Repeat("9", 34), 0, ""},
			{"fractional_cohort", "-" + zeros + "1.2300E+7", codec.Normal, true, "12300", 3, ""},
			{"fractional_leading_zeros", "0." + zeros + fmt.Sprintf("1200E+%d", size+4), codec.Normal, false, "1200", 0, ""},
			{"all_zero", "-0" + zeros, codec.Zero, true, "", 0, ""},
			{"fractional_all_zero", "-0.0" + zeros + "E+7", codec.Zero, true, "", int32(6 - size), ""},
			{"positive_exponent", "1E+" + zeros + "1", codec.Normal, false, "1", 1, ""},
			{"negative_exponent", "1e-" + zeros + "1", codec.Normal, false, "1", -1, ""},
			{"zero_exponent", "1E-0" + zeros, codec.Normal, false, "1", 0, ""},
			{"fractional_exponent", "1.2300E+" + zeros + "7", codec.Normal, false, "12300", 3, ""},
			{"nan_payload_limit", "NaN" + zeros + strings.Repeat("9", 33), codec.QNaN, false, "", 0, strings.Repeat("9", 33)},
			{"snan_payload", "-sNaN" + zeros + "1200", codec.SNaN, true, "", 0, "1200"},
			{"zero_nan_payload", "NaN" + zeros, codec.QNaN, false, "", 0, "0"},
		} {
			cases++
			t.Run(fmt.Sprintf("%s/%d", tc.name, size), func(t *testing.T) {
				c, err, allocated := measureFromString(tc.input)
				if allocated > 32<<10 {
					t.Errorf("FromString allocated %d bytes, limit %d", allocated, 32<<10)
				}
				t.Logf("input=%d allocated=%d coefficient_limit=34 payload_limit=33 exponent_literal_limit=16", len(tc.input), allocated)
				if err != nil {
					t.Fatalf("FromString: %v", err)
				}
				if c.Kind != tc.kind || c.Sign != tc.sign || c.Exponent != tc.exponent {
					t.Errorf("kind/sign/exponent = %v/%v/%d, want %v/%v/%d", c.Kind, c.Sign, c.Exponent, tc.kind, tc.sign, tc.exponent)
				}
				if tc.coefficient == "" {
					if c.Coefficient != nil {
						t.Errorf("unexpected coefficient %v", c.Coefficient)
					}
				} else if c.Coefficient == nil || c.Coefficient.String() != tc.coefficient {
					t.Errorf("coefficient = %v, want %s", c.Coefficient, tc.coefficient)
				}
				if tc.payload == "" {
					if c.Payload != nil {
						t.Errorf("unexpected payload %v", c.Payload)
					}
				} else if c.Payload == nil || c.Payload.String() != tc.payload {
					t.Errorf("payload = %v, want %s", c.Payload, tc.payload)
				}
			})
		}
	}
	return cases
}

func measureFromString(input string) (codec.Components, error, uint64) {
	runtime.GC()
	var before, after runtime.MemStats
	var result codec.Components
	var err error
	runtime.ReadMemStats(&before)
	for i := 0; i < 3; i++ {
		result, err = codec.FromString(input)
	}
	runtime.ReadMemStats(&after)
	return result, err, (after.TotalAlloc - before.TotalAlloc) / 3
}
