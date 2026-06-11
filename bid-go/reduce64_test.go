package bidgo

import "testing"

func mustBid64FromString(t *testing.T, s string) uint64 {
	t.Helper()
	v, flags := Bid64FromString(s, BID_ROUNDING_TO_NEAREST)
	if flags != 0 {
		t.Fatalf("Bid64FromString(%q) flags = %02x, want 0", s, flags)
	}
	return v
}

func TestBid64ReduceFiniteCases(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "integer", in: "1.00", want: "1"},
		{name: "power_of_ten", in: "120.00", want: "1.2E+2"},
		{name: "positive_zero", in: "0E+5", want: "0"},
		{name: "negative_zero", in: "-0E+5", want: "-0"},
		{name: "small", in: "1.000000000000000E-383", want: "1E-383"},
		{name: "subnormal", in: "2.000E-395", want: "2E-395"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := mustBid64FromString(t, tt.in)
			want := mustBid64FromString(t, tt.want)
			got, flags := Bid64Reduce(in)
			if got != want || flags != 0 {
				t.Fatalf("Bid64Reduce(%q) = %016x/%02x, want %016x/00", tt.in, got, flags, want)
			}
		})
	}
}

func TestBid64ReduceSpecialCases(t *testing.T) {
	t.Run("quiet NaN payload preserved", func(t *testing.T) {
		in := mustBid64FromString(t, "NaN101")
		got, flags := Bid64Reduce(in)
		if got != in || flags != 0 {
			t.Fatalf("Bid64Reduce(NaN101) = %016x/%02x, want %016x/00", got, flags, in)
		}
	})

	t.Run("signaling NaN quieted", func(t *testing.T) {
		in := mustBid64FromString(t, "sNaN010")
		want := in & QUIET_MASK64
		got, flags := Bid64Reduce(in)
		if got != want || flags != BID_INVALID_EXCEPTION {
			t.Fatalf("Bid64Reduce(sNaN010) = %016x/%02x, want %016x/%02x", got, flags, want, BID_INVALID_EXCEPTION)
		}
	})

	t.Run("infinity unchanged", func(t *testing.T) {
		in := mustBid64FromString(t, "-Inf")
		got, flags := Bid64Reduce(in)
		if got != in || flags != 0 {
			t.Fatalf("Bid64Reduce(-Inf) = %016x/%02x, want %016x/00", got, flags, in)
		}
	})
}
