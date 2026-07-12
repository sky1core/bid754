package bid754

import "testing"

func TestInvalidRoundingModeDoesNotMaskRejectedParseInput(t *testing.T) {
	badMode := RoundingMode(99)
	tests := []struct {
		name     string
		input32  string
		input64  string
		input128 string
	}{
		{name: "malformed", input32: "not-a-decimal", input64: "not-a-decimal", input128: "not-a-decimal"},
		{name: "cohort", input32: "10000000", input64: "10000000000000000", input128: "10000000000000000000000000000000000"},
		{name: "nan_payload", input32: "NaN1000000", input64: "NaN1000000000000000", input128: "NaN1000000000000000000000000000000000"},
	}
	for _, tc := range tests {
		t.Run(tc.name+"/decimal32", func(t *testing.T) {
			got, flags, err := NewDecimal32WithMode(tc.input32, badMode)
			if err == nil || got != 0 || flags != 0 {
				t.Fatalf("NewDecimal32WithMode(%q, %d) = (%#x, %v, %v), want zero, zero flags, error", tc.input32, badMode, uint32(got), flags, err)
			}
		})
		t.Run(tc.name+"/decimal64", func(t *testing.T) {
			got, flags, err := NewDecimal64WithMode(tc.input64, badMode)
			if err == nil || got != 0 || flags != 0 {
				t.Fatalf("NewDecimal64WithMode(%q, %d) = (%#x, %v, %v), want zero, zero flags, error", tc.input64, badMode, uint64(got), flags, err)
			}
		})
		t.Run(tc.name+"/decimal128", func(t *testing.T) {
			got, flags, err := NewDecimal128WithMode(tc.input128, badMode)
			if err == nil || got != (Decimal128BID{}) || flags != 0 {
				t.Fatalf("NewDecimal128WithMode(%q, %d) = (%x, %v, %v), want zero, zero flags, error", tc.input128, badMode, got.ToBytes(), flags, err)
			}
		})
	}
}

func TestInvalidRoundingModeStillUsesFlagChannelForAcceptedParseInput(t *testing.T) {
	badMode := RoundingMode(99)
	for _, input := range []string{"1", "1.2345678"} {
		if got, flags, err := NewDecimal32WithMode(input, badMode); err != nil || got != canonicalQNaN32BID() || flags != FlagInvalidOperation {
			t.Fatalf("NewDecimal32WithMode(%q, %d) = (%#x, %v, %v), want canonical qNaN, InvalidOperation, nil", input, badMode, uint32(got), flags, err)
		}
	}
	for _, input := range []string{"1", "1.2345678901234567"} {
		if got, flags, err := NewDecimal64WithMode(input, badMode); err != nil || got != canonicalQNaN64BID() || flags != FlagInvalidOperation {
			t.Fatalf("NewDecimal64WithMode(%q, %d) = (%#x, %v, %v), want canonical qNaN, InvalidOperation, nil", input, badMode, uint64(got), flags, err)
		}
	}
	for _, input := range []string{"1", "1.2345678901234567890123456789012345"} {
		if got, flags, err := NewDecimal128WithMode(input, badMode); err != nil || got != canonicalQNaN128BID() || flags != FlagInvalidOperation {
			t.Fatalf("NewDecimal128WithMode(%q, %d) = (%x, %v, %v), want canonical qNaN, InvalidOperation, nil", input, badMode, got.ToBytes(), flags, err)
		}
	}
}
