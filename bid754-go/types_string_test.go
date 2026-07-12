package bid754

import "testing"

func TestExceptionFlagsStringPreservesUnknownBits(t *testing.T) {
	const (
		unknown  ExceptionFlags = 1 << 20
		allKnown                = FlagInexact | FlagUnderflow | FlagOverflow |
			FlagDivisionByZero | FlagInvalidOperation | FlagSubnormal |
			FlagRounded | FlagClamped
		allKnownString = "Inexact|Underflow|Overflow|DivisionByZero|" +
			"InvalidOperation|Subnormal|Rounded|Clamped"
	)

	tests := []struct {
		name  string
		flags ExceptionFlags
		want  string
	}{
		{name: "none", flags: 0, want: "None"},
		{name: "all known", flags: allKnown, want: allKnownString},
		{name: "unknown only", flags: unknown, want: "Unknown(0x100000)"},
		{name: "known and unknown", flags: FlagInexact | unknown, want: "Inexact|Unknown(0x100000)"},
		{name: "multiple unknown", flags: unknown | 1<<22, want: "Unknown(0x500000)"},
		{name: "negative unknown only", flags: ExceptionFlags(-256), want: "Unknown(-0x100)"},
		{name: "negative all bits", flags: ExceptionFlags(-1), want: allKnownString + "|Unknown(-0x100)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.String(); got != tt.want {
				t.Fatalf("ExceptionFlags(%#x).String() = %q, want %q", int(tt.flags), got, tt.want)
			}
		})
	}
}
