package testgen

import (
	"strings"
	"testing"
)

func TestEmitMixedRustPublicParityUsesBothOperandWidths(t *testing.T) {
	row := rustParityInventoryRow{
		GoSymbol:      "Add64DQBIDWithMode",
		RustOwner:     "Decimal64",
		RustSurface:   "add_dq_with_mode",
		Shape:         "mixed_binary_mode_flags_dq",
		BidgoFunction: "Bid64dqAdd",
	}
	var b strings.Builder
	_, cases, err := emitRustParityUnit(&b, row, publicParityCorpus{})
	if err != nil {
		t.Fatal(err)
	}
	if want := (len(parityLabelPairs) + 4) * len(publicParityModeOrderNames); cases != want {
		t.Fatalf("mixed Rust parity cases = %d, want %d", cases, want)
	}
	got := b.String()
	for _, want := range []string{
		"Decimal64::add_dq_with_mode(Decimal64::from_bits(left_bits), Decimal128::from_le_bytes(right_bits), mode)",
		"bid754::generated::bid128_add::bid64dq_add(left_bits, to_port128(right_bits), port_mode)",
		"let left_pair = PAIRS_64[pair_index]",
		"let right_pair = PAIRS_128[pair_index]",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("mixed Rust parity output missing %q:\n%s", want, got)
		}
	}
}

func TestEmitMixedRustPublicParityDoesNotInventModeDiscriminationForExactDDMul(t *testing.T) {
	row := rustParityInventoryRow{
		GoSymbol:      "Mul128DDBIDWithMode",
		RustOwner:     "Decimal128",
		RustSurface:   "mul_dd_with_mode",
		Shape:         "mixed_binary_mode_flags_dd",
		BidgoFunction: "Bid128ddMul",
	}
	var b strings.Builder
	_, cases, err := emitRustParityUnit(&b, row, publicParityCorpus{})
	if err != nil {
		t.Fatal(err)
	}
	if want := len(parityLabelPairs) * len(publicParityModeOrderNames); cases != want {
		t.Fatalf("exact DD multiplication cases = %d, want %d", cases, want)
	}
	if strings.Contains(b.String(), "mode_seen") {
		t.Fatalf("exact Decimal64-by-Decimal64 to Decimal128 multiplication must not claim a mode-discriminating corpus:\n%s", b.String())
	}
}
