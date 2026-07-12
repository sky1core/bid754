package bid754

import "testing"

func TestDecimalBIDIsSignMinusObservesSignedZero(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		negZero := mustDecimal32BID(t, "-0")
		posZero := mustDecimal32BID(t, "+0")
		negOne := mustDecimal32BID(t, "-1")
		posOne := mustDecimal32BID(t, "+1")

		assertSignMinus(t, "decimal32 -0", negZero.Sign(), negZero.IsSignMinus(), 0, true)
		assertSignMinus(t, "decimal32 +0", posZero.Sign(), posZero.IsSignMinus(), 0, false)
		assertSignMinus(t, "decimal32 -1", negOne.Sign(), negOne.IsSignMinus(), -1, true)
		assertSignMinus(t, "decimal32 +1", posOne.Sign(), posOne.IsSignMinus(), 1, false)
	})

	t.Run("decimal64", func(t *testing.T) {
		negZero := mustDecimal64BID(t, "-0")
		posZero := mustDecimal64BID(t, "+0")
		negOne := mustDecimal64BID(t, "-1")
		posOne := mustDecimal64BID(t, "+1")

		assertSignMinus(t, "decimal64 -0", negZero.Sign(), negZero.IsSignMinus(), 0, true)
		assertSignMinus(t, "decimal64 +0", posZero.Sign(), posZero.IsSignMinus(), 0, false)
		assertSignMinus(t, "decimal64 -1", negOne.Sign(), negOne.IsSignMinus(), -1, true)
		assertSignMinus(t, "decimal64 +1", posOne.Sign(), posOne.IsSignMinus(), 1, false)
	})

	t.Run("decimal128", func(t *testing.T) {
		negZero := mustDecimal128BID(t, "-0")
		posZero := mustDecimal128BID(t, "+0")
		negOne := mustDecimal128BID(t, "-1")
		posOne := mustDecimal128BID(t, "+1")

		assertSignMinus(t, "decimal128 -0", negZero.Sign(), negZero.IsSignMinus(), 0, true)
		assertSignMinus(t, "decimal128 +0", posZero.Sign(), posZero.IsSignMinus(), 0, false)
		assertSignMinus(t, "decimal128 -1", negOne.Sign(), negOne.IsSignMinus(), -1, true)
		assertSignMinus(t, "decimal128 +1", posOne.Sign(), posOne.IsSignMinus(), 1, false)
	})
}

func assertSignMinus(t *testing.T, label string, gotSign int, gotSignMinus bool, wantSign int, wantSignMinus bool) {
	t.Helper()
	if gotSign != wantSign {
		t.Fatalf("%s Sign() = %d, want %d", label, gotSign, wantSign)
	}
	if gotSignMinus != wantSignMinus {
		t.Fatalf("%s IsSignMinus() = %v, want %v", label, gotSignMinus, wantSignMinus)
	}
}
