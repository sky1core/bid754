package bid754

import "testing"

func TestParseDecimalBIDRawRejectsMalformedInputThroughFlags(t *testing.T) {
	inputs := []string{
		"",
		"not-a-decimal",
		".",
		"1e2junk",
		"1 ",
		"--1",
	}
	for _, input := range inputs {
		t.Run(input+"/decimal32", func(t *testing.T) {
			got, flags := ParseDecimal32BIDRaw(input)
			if got != canonicalQNaN32BID() || flags != FlagInvalidOperation {
				t.Fatalf("ParseDecimal32BIDRaw(%q) = (%08x, %v), want canonical qNaN and InvalidOperation", input, got.ToUint32(), flags)
			}
		})
		t.Run(input+"/decimal64", func(t *testing.T) {
			got, flags := ParseDecimal64BIDRaw(input)
			if got != canonicalQNaN64BID() || flags != FlagInvalidOperation {
				t.Fatalf("ParseDecimal64BIDRaw(%q) = (%016x, %v), want canonical qNaN and InvalidOperation", input, got.ToUint64(), flags)
			}
		})
		t.Run(input+"/decimal128", func(t *testing.T) {
			got, flags := ParseDecimal128BIDRaw(input)
			if got != canonicalQNaN128BID() || flags != FlagInvalidOperation {
				t.Fatalf("ParseDecimal128BIDRaw(%q) = (%x, %v), want canonical qNaN and InvalidOperation", input, got.ToBytes(), flags)
			}
		})
	}
}

func TestParseDecimalBIDRawKeepsValidFlagSemantics(t *testing.T) {
	if got, flags := ParseDecimal32BIDRaw("1.2345678"); got.IsNaN() || flags != FlagInexact {
		t.Fatalf("Decimal32 rounded finite parse = (%08x, %v), want finite and Inexact", got.ToUint32(), flags)
	}
	if got, flags := ParseDecimal64BIDRaw("1.2345678901234567"); got.IsNaN() || flags != FlagInexact {
		t.Fatalf("Decimal64 rounded finite parse = (%016x, %v), want finite and Inexact", got.ToUint64(), flags)
	}
	if got, flags := ParseDecimal128BIDRaw("1.2345678901234567890123456789012345"); got.IsNaN() || flags != FlagInexact {
		t.Fatalf("Decimal128 rounded finite parse = (%x, %v), want finite and Inexact", got.ToBytes(), flags)
	}

	if got, flags := ParseDecimal32BIDRaw("NaN123"); got != Decimal32BID(0x7c00007b) || flags != 0 {
		t.Fatalf("Decimal32 NaN payload parse = (%08x, %v), want payload 123 and no flags", got.ToUint32(), flags)
	}
	if got, flags := ParseDecimal64BIDRaw("NaN123"); got != Decimal64BID(0x7c0000000000007b) || flags != 0 {
		t.Fatalf("Decimal64 NaN payload parse = (%016x, %v), want payload 123 and no flags", got.ToUint64(), flags)
	}
	if got, flags := ParseDecimal128BIDRaw("NaN123"); got.String() != "+NaN123" || flags != 0 {
		t.Fatalf("Decimal128 NaN payload parse = (%x/%q, %v), want payload 123 and no flags", got.ToBytes(), got.String(), flags)
	}

	if got, flags := ParseDecimal32BIDRaw("1e1000"); !got.IsInf() || flags != FlagOverflow|FlagInexact {
		t.Fatalf("Decimal32 valid overflow parse = (%08x, %v), want infinity and Overflow|Inexact", got.ToUint32(), flags)
	}
	if got, flags := ParseDecimal64BIDRaw("1e1000"); !got.IsInf() || flags != FlagOverflow|FlagInexact {
		t.Fatalf("Decimal64 valid overflow parse = (%016x, %v), want infinity and Overflow|Inexact", got.ToUint64(), flags)
	}
	if got, flags := ParseDecimal128BIDRaw("1e1000"); got.IsNaN() || flags != 0 {
		t.Fatalf("Decimal128 in-range exponent parse = (%x, %v), want finite and no flags", got.ToBytes(), flags)
	}
}
