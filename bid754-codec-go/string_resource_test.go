package bidcodec_test

import (
	"runtime"
	"strings"
	"testing"

	codec "github.com/sky1core/bid754/bid754-codec-go"
)

func TestFromStringResourceBoundaryRejects(t *testing.T) {
	digits := strings.Repeat("9", 200000)
	zeros := strings.Repeat("0", 200000)
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
		t.Run(tc.name, func(t *testing.T) {
			runtime.GC()
			var before, after runtime.MemStats
			runtime.ReadMemStats(&before)
			_, err := codec.FromString(tc.input)
			runtime.ReadMemStats(&after)
			if err == nil {
				t.Fatal("oversized or malformed input accepted")
			}
			allocated := after.TotalAlloc - before.TotalAlloc
			if allocated > 1<<20 {
				t.Errorf("FromString allocated %d bytes, limit %d", allocated, 1<<20)
			}
			if n := len(err.Error()); n > 256 {
				t.Errorf("error has %d bytes, limit 256", n)
			}
			t.Logf("allocated=%d bytes error=%d bytes", allocated, len(err.Error()))
		})
	}
}

func TestFromStringResourceBoundaryLeadingZeros(t *testing.T) {
	zeros := strings.Repeat("0", 200000)
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
		{"fractional_leading_zeros", "0." + zeros + "1200E+200004", codec.Normal, false, "1200", 0, ""},
		{"all_zero", "-" + zeros, codec.Zero, true, "", 0, ""},
		{"fractional_all_zero", "-0." + zeros + "E+7", codec.Zero, true, "", -199993, ""},
		{"positive_exponent", "1E+" + zeros + "1", codec.Normal, false, "1", 1, ""},
		{"negative_exponent", "1e-" + zeros + "1", codec.Normal, false, "1", -1, ""},
		{"zero_exponent", "1E-" + zeros, codec.Normal, false, "1", 0, ""},
		{"fractional_exponent", "1.2300E+" + zeros + "7", codec.Normal, false, "12300", 3, ""},
		{"nan_payload_limit", "NaN" + zeros + strings.Repeat("9", 33), codec.QNaN, false, "", 0, strings.Repeat("9", 33)},
		{"snan_payload", "-sNaN" + zeros + "1200", codec.SNaN, true, "", 0, "1200"},
		{"zero_nan_payload", "NaN" + zeros, codec.QNaN, false, "", 0, "0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, err := codec.FromString(tc.input)
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
