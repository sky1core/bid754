package bid754

import "testing"

func TestDecimalBIDWithFlagsPeersMatchValueMethods(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		a := mustDecimal32BID(t, "9.999999")
		b := mustDecimal32BID(t, "9.999999")

		sum, flags := a.AddWithFlags(b)
		if sum != a.Add(b) {
			t.Fatalf("AddWithFlags result = %s, want Add result %s", sum, a.Add(b))
		}
		if flags == 0 {
			t.Fatal("AddWithFlags did not report expected inexact flags")
		}

		divisor := mustDecimal32BID(t, "3")
		quotient, divFlags := mustDecimal32BID(t, "1").DivWithFlags(divisor)
		if quotient != mustDecimal32BID(t, "1").Div(divisor) {
			t.Fatalf("DivWithFlags result = %s, want Div result %s", quotient, mustDecimal32BID(t, "1").Div(divisor))
		}
		if divFlags == 0 {
			t.Fatal("DivWithFlags did not report expected inexact flags")
		}

		rounded, roundFlags := mustDecimal32BID(t, "1.5").RoundIntegralExactWithFlags()
		if rounded != mustDecimal32BID(t, "1.5").RoundIntegralExact() {
			t.Fatalf("RoundIntegralExactWithFlags result = %s, want RoundIntegralExact result %s", rounded, mustDecimal32BID(t, "1.5").RoundIntegralExact())
		}
		if roundFlags == 0 {
			t.Fatal("RoundIntegralExactWithFlags did not report expected flags")
		}
	})

	t.Run("decimal64", func(t *testing.T) {
		a := mustDecimal64BID(t, "9.999999999999999")
		b := mustDecimal64BID(t, "9.999999999999999")

		sum, flags := a.AddWithFlags(b)
		if sum != a.Add(b) {
			t.Fatalf("AddWithFlags result = %s, want Add result %s", sum, a.Add(b))
		}
		if flags == 0 {
			t.Fatal("AddWithFlags did not report expected inexact flags")
		}

		divisor := mustDecimal64BID(t, "3")
		quotient, divFlags := mustDecimal64BID(t, "1").DivWithFlags(divisor)
		if quotient != mustDecimal64BID(t, "1").Div(divisor) {
			t.Fatalf("DivWithFlags result = %s, want Div result %s", quotient, mustDecimal64BID(t, "1").Div(divisor))
		}
		if divFlags == 0 {
			t.Fatal("DivWithFlags did not report expected inexact flags")
		}

		rounded, roundFlags := mustDecimal64BID(t, "1.5").RoundIntegralExactWithFlags()
		if rounded != mustDecimal64BID(t, "1.5").RoundIntegralExact() {
			t.Fatalf("RoundIntegralExactWithFlags result = %s, want RoundIntegralExact result %s", rounded, mustDecimal64BID(t, "1.5").RoundIntegralExact())
		}
		if roundFlags == 0 {
			t.Fatal("RoundIntegralExactWithFlags did not report expected flags")
		}
	})

	t.Run("decimal128", func(t *testing.T) {
		a := mustDecimal128BID(t, "9.999999999999999999999999999999999")
		b := mustDecimal128BID(t, "9.999999999999999999999999999999999")

		sum, flags := a.AddWithFlags(b)
		if sum != a.Add(b) {
			t.Fatalf("AddWithFlags result = %s, want Add result %s", sum, a.Add(b))
		}
		if flags == 0 {
			t.Fatal("AddWithFlags did not report expected inexact flags")
		}

		divisor := mustDecimal128BID(t, "3")
		quotient, divFlags := mustDecimal128BID(t, "1").DivWithFlags(divisor)
		if quotient != mustDecimal128BID(t, "1").Div(divisor) {
			t.Fatalf("DivWithFlags result = %s, want Div result %s", quotient, mustDecimal128BID(t, "1").Div(divisor))
		}
		if divFlags == 0 {
			t.Fatal("DivWithFlags did not report expected inexact flags")
		}

		rounded, roundFlags := mustDecimal128BID(t, "1.5").RoundIntegralExactWithFlags()
		if rounded != mustDecimal128BID(t, "1.5").RoundIntegralExact() {
			t.Fatalf("RoundIntegralExactWithFlags result = %s, want RoundIntegralExact result %s", rounded, mustDecimal128BID(t, "1.5").RoundIntegralExact())
		}
		if roundFlags == 0 {
			t.Fatal("RoundIntegralExactWithFlags did not report expected flags")
		}
	})
}

func TestAddSliceBIDWithFlagsAccumulatesFlags(t *testing.T) {
	values32 := []Decimal32BID{
		mustDecimal32BID(t, "9.999999"),
		mustDecimal32BID(t, "9.999999"),
		mustDecimal32BID(t, "1"),
	}
	sum32, flags32 := AddSlice32BIDWithFlags(values32)
	if sum32 != AddSlice32BID(values32) {
		t.Fatalf("AddSlice32BIDWithFlags result = %s, want AddSlice32BID result %s", sum32, AddSlice32BID(values32))
	}
	if flags32 == 0 {
		t.Fatal("AddSlice32BIDWithFlags did not accumulate flags")
	}

	values64 := []Decimal64BID{
		mustDecimal64BID(t, "9.999999999999999"),
		mustDecimal64BID(t, "9.999999999999999"),
		mustDecimal64BID(t, "1"),
	}
	sum64, flags64 := AddSlice64BIDWithFlags(values64)
	if sum64 != AddSlice64BID(values64) {
		t.Fatalf("AddSlice64BIDWithFlags result = %s, want AddSlice64BID result %s", sum64, AddSlice64BID(values64))
	}
	if flags64 == 0 {
		t.Fatal("AddSlice64BIDWithFlags did not accumulate flags")
	}

	values128 := []Decimal128BID{
		mustDecimal128BID(t, "9.999999999999999999999999999999999"),
		mustDecimal128BID(t, "9.999999999999999999999999999999999"),
		mustDecimal128BID(t, "1"),
	}
	sum128, flags128 := AddSlice128BIDWithFlags(values128)
	if sum128 != AddSlice128BID(values128) {
		t.Fatalf("AddSlice128BIDWithFlags result = %s, want AddSlice128BID result %s", sum128, AddSlice128BID(values128))
	}
	if flags128 == 0 {
		t.Fatal("AddSlice128BIDWithFlags did not accumulate flags")
	}
}
