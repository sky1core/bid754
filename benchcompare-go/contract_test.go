package benchcompare

import (
	"testing"

	shop "github.com/shopspring/decimal"
	codec "github.com/sky1core/bid754/bid754-codec-go"
	bid "github.com/sky1core/bid754/bid754-go"
)

// TestOperandContract enforces the shared input contract behind every
// benchmark row at every width: each parse input must be exact and
// flag-clean at its BID width and accepted by shopspring, and each parts
// row must pass the width's validating codec encode. The bench-compare-go
// make target runs this test before the benchmarks, so an operand edit that
// breaks the contract fails the run instead of contaminating the numbers.
// (The package init additionally panics on an inexact parse input, so even
// a bench-only invocation cannot skip the parse-side contract.)
func TestOperandContract(t *testing.T) {
	for _, s := range parseInputsD32 {
		if _, flags := bid.ParseDecimal32BIDRaw(s); flags != 0 {
			t.Fatalf("d32 parse input %q is not exact/clean (flags %v)", s, flags)
		}
		if _, err := shop.NewFromString(s); err != nil {
			t.Fatalf("shopspring rejects d32 parse input %q: %v", s, err)
		}
	}
	for _, s := range parseInputsD64 {
		if _, flags := bid.ParseDecimal64BIDRaw(s); flags != 0 {
			t.Fatalf("d64 parse input %q is not exact/clean (flags %v)", s, flags)
		}
		if _, err := shop.NewFromString(s); err != nil {
			t.Fatalf("shopspring rejects d64 parse input %q: %v", s, err)
		}
	}
	for _, s := range parseInputsD128 {
		if _, flags := bid.ParseDecimal128BIDRaw(s); flags != 0 {
			t.Fatalf("d128 parse input %q is not exact/clean (flags %v)", s, flags)
		}
		if _, err := shop.NewFromString(s); err != nil {
			t.Fatalf("shopspring rejects d128 parse input %q: %v", s, err)
		}
	}
	for _, r := range partsRowsD32 {
		if _, err := codec.Encode32(partsComponents(r)); err != nil {
			t.Fatalf("codec.Encode32 rejects parts row (%d, %d): %v", r.mant, r.exp, err)
		}
	}
	for _, r := range partsRowsD64 {
		if _, err := codec.Encode64(partsComponents(r)); err != nil {
			t.Fatalf("codec.Encode64 rejects parts row (%d, %d): %v", r.mant, r.exp, err)
		}
	}
	for _, r := range partsRowsD128 {
		if _, _, err := codec.Encode128(partsComponents(r)); err != nil {
			t.Fatalf("codec.Encode128 rejects parts row (%d, %d): %v", r.mant, r.exp, err)
		}
	}
	// shopspring Div panics on a zero divisor, so pin every divisor pair
	// non-zero here (mirrors the Rust contract's divisor check).
	for _, p := range divIdx {
		for width, vals := range map[string][]shop.Decimal{"d32": shop32Vals, "d64": shop64Vals, "d128": shop128Vals} {
			if vals[p[1]].IsZero() {
				t.Fatalf("%s div pair (%d,%d) has a zero divisor", width, p[0], p[1])
			}
		}
	}
}
