package benchcompare

import (
	"math/big"
	"testing"

	shop "github.com/shopspring/decimal"
	codec "github.com/sky1core/bid754/bid754-codec-go"
	bid "github.com/sky1core/bid754/bid754-go"
)

// TestOperandContract enforces the shared input contract behind every
// benchmark row: each parse input must be exact and flag-clean for
// Decimal64 and accepted by shopspring, and each parts row must pass the
// validating codec encode. The bench-compare-go make target runs this test
// before the benchmarks, so an operand edit that breaks the contract fails
// the run instead of contaminating the numbers.
func TestOperandContract(t *testing.T) {
	for _, s := range parseInputs {
		if _, flags := bid.ParseDecimal64BIDRaw(s); flags != 0 {
			t.Fatalf("parse input %q is not exact/clean for Decimal64 (flags %v)", s, flags)
		}
		if _, err := shop.NewFromString(s); err != nil {
			t.Fatalf("shopspring rejects parse input %q: %v", s, err)
		}
	}
	for _, r := range partsRows {
		m := r.mant
		neg := m < 0
		if m < 0 {
			m = -m
		}
		kind := codec.Normal
		if m == 0 {
			kind = codec.Zero
		}
		if _, err := codec.Encode64(codec.Components{
			Sign: neg, Coefficient: big.NewInt(m), Exponent: r.exp, Kind: kind,
		}); err != nil {
			t.Fatalf("codec.Encode64 rejects parts row (%d, %d): %v", r.mant, r.exp, err)
		}
	}
}
