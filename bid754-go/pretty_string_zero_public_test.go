package bid754

import "testing"

// Zero values must pretty-print as "0"/"-0" regardless of the stored
// exponent; a zero coefficient with a positive exponent previously expanded
// into a run of zero digits like "000000".
func TestPrettyStringZeroCoefficient(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"0E+5", "0"},
		{"-0E+5", "-0"},
		{"0E-5", "0"},
		{"-0E-5", "-0"},
		{"0E+20", "0"},
		// Exponents beyond the formatDecimalString plain-notation range
		// (|exponent| > 20) previously appeared as "0E+21" etc.
		{"0E+21", "0"},
		{"-0E+21", "-0"},
		{"0E-21", "0"},
		{"-0E-21", "-0"},
		{"0E+30", "0"},
		{"-0E+30", "-0"},
		{"0", "0"},
		{"-0", "-0"},
	}
	for _, tc := range cases {
		d32, err := NewDecimal32(tc.input)
		if err != nil {
			t.Fatalf("NewDecimal32(%q): %v", tc.input, err)
		}
		if got := d32.PrettyString(); got != tc.want {
			t.Errorf("Decimal32(%q).PrettyString() = %q, want %q", tc.input, got, tc.want)
		}

		d64, err := NewDecimal64(tc.input)
		if err != nil {
			t.Fatalf("NewDecimal64(%q): %v", tc.input, err)
		}
		if got := d64.PrettyString(); got != tc.want {
			t.Errorf("Decimal64(%q).PrettyString() = %q, want %q", tc.input, got, tc.want)
		}

		d128, err := NewDecimal128(tc.input)
		if err != nil {
			t.Fatalf("NewDecimal128(%q): %v", tc.input, err)
		}
		if got := d128.PrettyString(); got != tc.want {
			t.Errorf("Decimal128(%q).PrettyString() = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// Zeros stored at each format's extreme representable exponents must still
// pretty-print as "0"/"-0".
func TestPrettyStringZeroCoefficientExtremeExponents(t *testing.T) {
	t.Run("Decimal32", func(t *testing.T) {
		for _, tc := range []struct {
			input string
			want  string
		}{
			{"0E+90", "0"},
			{"-0E+90", "-0"},
			{"0E-101", "0"},
			{"-0E-101", "-0"},
		} {
			d, err := NewDecimal32(tc.input)
			if err != nil {
				t.Fatalf("NewDecimal32(%q): %v", tc.input, err)
			}
			if got := d.PrettyString(); got != tc.want {
				t.Errorf("Decimal32(%q).PrettyString() = %q, want %q", tc.input, got, tc.want)
			}
		}
	})

	t.Run("Decimal64", func(t *testing.T) {
		for _, tc := range []struct {
			input string
			want  string
		}{
			{"0E+369", "0"},
			{"-0E+369", "-0"},
			{"0E-398", "0"},
			{"-0E-398", "-0"},
		} {
			d, err := NewDecimal64(tc.input)
			if err != nil {
				t.Fatalf("NewDecimal64(%q): %v", tc.input, err)
			}
			if got := d.PrettyString(); got != tc.want {
				t.Errorf("Decimal64(%q).PrettyString() = %q, want %q", tc.input, got, tc.want)
			}
		}
	})

	t.Run("Decimal128", func(t *testing.T) {
		for _, tc := range []struct {
			input string
			want  string
		}{
			{"0E+6111", "0"},
			{"-0E+6111", "-0"},
			{"0E-6176", "0"},
			{"-0E-6176", "-0"},
		} {
			d, err := NewDecimal128(tc.input)
			if err != nil {
				t.Fatalf("NewDecimal128(%q): %v", tc.input, err)
			}
			if got := d.PrettyString(); got != tc.want {
				t.Errorf("Decimal128(%q).PrettyString() = %q, want %q", tc.input, got, tc.want)
			}
		}
	})
}
