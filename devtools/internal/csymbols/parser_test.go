package csymbols

import "testing"

func TestNormalizeWhitespaceCanonicalizesPointerSpacing(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{in: "_IDEC_flags *pfpsf", want: "_IDEC_flags*pfpsf"},
		{in: "_IDEC_flags* pfpsf", want: "_IDEC_flags*pfpsf"},
		{in: "BID_UINT128 * x", want: "BID_UINT128* x"},
	} {
		if got := normalizeWhitespace(tc.in); got != tc.want {
			t.Fatalf("normalizeWhitespace(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
