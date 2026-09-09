package testgen

import (
	"fmt"
	"sort"
	"strings"
)

// The plus/minus quantum strength anchor is a hand-pinned known-answer table
// executed directly by the two portable decTest runners (runDectestGoportCase /
// run_dectest_goport_case) and by the separate value-type unary adapter
// (executeDecTestUnaryOperation), bypassing the skip and flag-exemption
// classifiers.
//
// The official dd/dq Plus and Minus suites do carry q>0 zero witnesses, and the
// portable legs compare those on exact quantum. This table covers what they
// leave open: the consumed Decimal32 suite (dsBase alone) carries no such
// witness, the regular decTest runner compares the unary adapter on value and
// the BID five-flag surface but never on quantum, and no suite closes both
// operations across all three widths, all five rounding modes, and the
// clamp/etiny/special boundaries at once.
//
// Every expected value is an independently decNumber-derived literal, never
// computed from the implementation under test. devtools/verification_sentinels.json
// carries the identical canonical rows outside every generation path, and this
// generator neither reads nor writes it.
const (
	dectestPlusMinusStrengthRowsPerCell = 16
	dectestPlusMinusStrengthCellCount   = 6 // 3 widths x 2 operations
	dectestPlusMinusStrengthTotalRows   = dectestPlusMinusStrengthRowsPerCell * dectestPlusMinusStrengthCellCount
)

// dectestPlusMinusStrengthRow is the language-neutral row shape. BID5 is the raw
// Intel BID five-flag surface (invalid 0x01, zero-divide 0x04, overflow 0x08,
// underflow 0x10, inexact 0x20) the case must raise once operand-parse flags and
// operation flags are accumulated.
type dectestPlusMinusStrengthRow struct {
	ID       string
	Width    string
	Op       string
	Category string
	Mode     string
	Input    string
	Expected string
	BID5     uint32
}

const dectestPlusMinusStrengthBID5Mask uint32 = 0x3d

// dectestPlusMinusStrengthModes maps the decTest rounding token to its Intel BID
// rounding-mode number. Each row carries its number into the generated case
// literal, and the generated Go and Rust tests compare the live
// decTestBIDRoundingMode / dec_test_bid_rounding_mode resolution against it, so
// this map pins the runners' run-time token-to-mode mapping rather than just the
// token vocabulary.
var dectestPlusMinusStrengthModes = map[string]int{
	"half_even": 0, // roundTiesToEven
	"floor":     1, // roundTowardNegative
	"ceiling":   2, // roundTowardPositive
	"down":      3, // roundTowardZero
	"half_up":   4, // roundTiesToAway
}

// dectestPlusMinusStrengthCellCategories is the exact category multiset every
// (width, operation) cell must carry: the five signed-zero rounding rows plus
// the eleven single-mode shape rows.
var dectestPlusMinusStrengthCellCategories = []string{
	"signed_zero", "signed_zero", "signed_zero", "signed_zero", "signed_zero",
	"quantum_positive", "quantum_zero", "quantum_negative", "overprecision",
	"high_clamp_zero", "high_clamp_nonzero", "etiny",
	"quiet_nan", "signaling_nan", "positive_infinity", "negative_infinity",
}

// dectestPlusMinusStrengthTable is the pinned oracle. signed_zero walks all five
// rounding modes over an exact zero sum (only roundTowardNegative keeps the
// negative sign); quantum_positive/zero/negative fix the result quantum at Q(x)
// for the three signs of the operand exponent; overprecision is the row whose
// PARSED quantum is the source exponent plus one, so only bid*_quantexp of the
// parsed operand names it; high_clamp_* and etiny sit at the format's extreme
// encodable exponents; the NaN and infinity rows have no quantum at all.
var dectestPlusMinusStrengthTable = []dectestPlusMinusStrengthRow{
	// decimal32 / plus
	{ID: "d32_plus_signed_zero_half_even", Width: "decimal32", Op: "plus", Category: "signed_zero", Mode: "half_even", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d32_plus_signed_zero_floor", Width: "decimal32", Op: "plus", Category: "signed_zero", Mode: "floor", Input: "-0E+3", Expected: "-0E+3", BID5: 0x00},
	{ID: "d32_plus_signed_zero_ceiling", Width: "decimal32", Op: "plus", Category: "signed_zero", Mode: "ceiling", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d32_plus_signed_zero_down", Width: "decimal32", Op: "plus", Category: "signed_zero", Mode: "down", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d32_plus_signed_zero_half_up", Width: "decimal32", Op: "plus", Category: "signed_zero", Mode: "half_up", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d32_plus_quantum_positive", Width: "decimal32", Op: "plus", Category: "quantum_positive", Mode: "half_even", Input: "56267E+1", Expected: "5.6267E+5", BID5: 0x00},
	{ID: "d32_plus_quantum_zero", Width: "decimal32", Op: "plus", Category: "quantum_zero", Mode: "half_even", Input: "56267", Expected: "56267", BID5: 0x00},
	{ID: "d32_plus_quantum_negative", Width: "decimal32", Op: "plus", Category: "quantum_negative", Mode: "half_even", Input: "56267E-2", Expected: "562.67", BID5: 0x00},
	{ID: "d32_plus_overprecision", Width: "decimal32", Op: "plus", Category: "overprecision", Mode: "half_even", Input: "12345675E+3", Expected: "1.234568E+10", BID5: 0x20},
	{ID: "d32_plus_high_clamp_zero", Width: "decimal32", Op: "plus", Category: "high_clamp_zero", Mode: "half_even", Input: "0E+96", Expected: "0E+90", BID5: 0x00},
	{ID: "d32_plus_high_clamp_nonzero", Width: "decimal32", Op: "plus", Category: "high_clamp_nonzero", Mode: "half_even", Input: "1E+96", Expected: "1.000000E+96", BID5: 0x00},
	{ID: "d32_plus_etiny", Width: "decimal32", Op: "plus", Category: "etiny", Mode: "half_even", Input: "1E-101", Expected: "1E-101", BID5: 0x00},
	{ID: "d32_plus_quiet_nan", Width: "decimal32", Op: "plus", Category: "quiet_nan", Mode: "half_even", Input: "NaN13", Expected: "NaN13", BID5: 0x00},
	{ID: "d32_plus_signaling_nan", Width: "decimal32", Op: "plus", Category: "signaling_nan", Mode: "half_even", Input: "sNaN13", Expected: "NaN13", BID5: 0x01},
	{ID: "d32_plus_positive_infinity", Width: "decimal32", Op: "plus", Category: "positive_infinity", Mode: "half_even", Input: "Infinity", Expected: "Infinity", BID5: 0x00},
	{ID: "d32_plus_negative_infinity", Width: "decimal32", Op: "plus", Category: "negative_infinity", Mode: "half_even", Input: "-Infinity", Expected: "-Infinity", BID5: 0x00},

	// decimal32 / minus
	{ID: "d32_minus_signed_zero_half_even", Width: "decimal32", Op: "minus", Category: "signed_zero", Mode: "half_even", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d32_minus_signed_zero_floor", Width: "decimal32", Op: "minus", Category: "signed_zero", Mode: "floor", Input: "0E+3", Expected: "-0E+3", BID5: 0x00},
	{ID: "d32_minus_signed_zero_ceiling", Width: "decimal32", Op: "minus", Category: "signed_zero", Mode: "ceiling", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d32_minus_signed_zero_down", Width: "decimal32", Op: "minus", Category: "signed_zero", Mode: "down", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d32_minus_signed_zero_half_up", Width: "decimal32", Op: "minus", Category: "signed_zero", Mode: "half_up", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d32_minus_quantum_positive", Width: "decimal32", Op: "minus", Category: "quantum_positive", Mode: "half_even", Input: "56267E+1", Expected: "-5.6267E+5", BID5: 0x00},
	{ID: "d32_minus_quantum_zero", Width: "decimal32", Op: "minus", Category: "quantum_zero", Mode: "half_even", Input: "56267", Expected: "-56267", BID5: 0x00},
	{ID: "d32_minus_quantum_negative", Width: "decimal32", Op: "minus", Category: "quantum_negative", Mode: "half_even", Input: "56267E-2", Expected: "-562.67", BID5: 0x00},
	{ID: "d32_minus_overprecision", Width: "decimal32", Op: "minus", Category: "overprecision", Mode: "half_even", Input: "12345675E+3", Expected: "-1.234568E+10", BID5: 0x20},
	{ID: "d32_minus_high_clamp_zero", Width: "decimal32", Op: "minus", Category: "high_clamp_zero", Mode: "half_even", Input: "0E+96", Expected: "0E+90", BID5: 0x00},
	{ID: "d32_minus_high_clamp_nonzero", Width: "decimal32", Op: "minus", Category: "high_clamp_nonzero", Mode: "half_even", Input: "1E+96", Expected: "-1.000000E+96", BID5: 0x00},
	{ID: "d32_minus_etiny", Width: "decimal32", Op: "minus", Category: "etiny", Mode: "half_even", Input: "1E-101", Expected: "-1E-101", BID5: 0x00},
	{ID: "d32_minus_quiet_nan", Width: "decimal32", Op: "minus", Category: "quiet_nan", Mode: "half_even", Input: "NaN13", Expected: "NaN13", BID5: 0x00},
	{ID: "d32_minus_signaling_nan", Width: "decimal32", Op: "minus", Category: "signaling_nan", Mode: "half_even", Input: "sNaN13", Expected: "NaN13", BID5: 0x01},
	{ID: "d32_minus_positive_infinity", Width: "decimal32", Op: "minus", Category: "positive_infinity", Mode: "half_even", Input: "Infinity", Expected: "-Infinity", BID5: 0x00},
	{ID: "d32_minus_negative_infinity", Width: "decimal32", Op: "minus", Category: "negative_infinity", Mode: "half_even", Input: "-Infinity", Expected: "Infinity", BID5: 0x00},

	// decimal64 / plus
	{ID: "d64_plus_signed_zero_half_even", Width: "decimal64", Op: "plus", Category: "signed_zero", Mode: "half_even", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d64_plus_signed_zero_floor", Width: "decimal64", Op: "plus", Category: "signed_zero", Mode: "floor", Input: "-0E+3", Expected: "-0E+3", BID5: 0x00},
	{ID: "d64_plus_signed_zero_ceiling", Width: "decimal64", Op: "plus", Category: "signed_zero", Mode: "ceiling", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d64_plus_signed_zero_down", Width: "decimal64", Op: "plus", Category: "signed_zero", Mode: "down", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d64_plus_signed_zero_half_up", Width: "decimal64", Op: "plus", Category: "signed_zero", Mode: "half_up", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d64_plus_quantum_positive", Width: "decimal64", Op: "plus", Category: "quantum_positive", Mode: "half_even", Input: "56267E+1", Expected: "5.6267E+5", BID5: 0x00},
	{ID: "d64_plus_quantum_zero", Width: "decimal64", Op: "plus", Category: "quantum_zero", Mode: "half_even", Input: "56267", Expected: "56267", BID5: 0x00},
	{ID: "d64_plus_quantum_negative", Width: "decimal64", Op: "plus", Category: "quantum_negative", Mode: "half_even", Input: "56267E-2", Expected: "562.67", BID5: 0x00},
	{ID: "d64_plus_overprecision", Width: "decimal64", Op: "plus", Category: "overprecision", Mode: "half_even", Input: "12345678901234565E+3", Expected: "1.234567890123456E+19", BID5: 0x20},
	{ID: "d64_plus_high_clamp_zero", Width: "decimal64", Op: "plus", Category: "high_clamp_zero", Mode: "half_even", Input: "0E+384", Expected: "0E+369", BID5: 0x00},
	{ID: "d64_plus_high_clamp_nonzero", Width: "decimal64", Op: "plus", Category: "high_clamp_nonzero", Mode: "half_even", Input: "1E+384", Expected: "1.000000000000000E+384", BID5: 0x00},
	{ID: "d64_plus_etiny", Width: "decimal64", Op: "plus", Category: "etiny", Mode: "half_even", Input: "1E-398", Expected: "1E-398", BID5: 0x00},
	{ID: "d64_plus_quiet_nan", Width: "decimal64", Op: "plus", Category: "quiet_nan", Mode: "half_even", Input: "NaN13", Expected: "NaN13", BID5: 0x00},
	{ID: "d64_plus_signaling_nan", Width: "decimal64", Op: "plus", Category: "signaling_nan", Mode: "half_even", Input: "sNaN13", Expected: "NaN13", BID5: 0x01},
	{ID: "d64_plus_positive_infinity", Width: "decimal64", Op: "plus", Category: "positive_infinity", Mode: "half_even", Input: "Infinity", Expected: "Infinity", BID5: 0x00},
	{ID: "d64_plus_negative_infinity", Width: "decimal64", Op: "plus", Category: "negative_infinity", Mode: "half_even", Input: "-Infinity", Expected: "-Infinity", BID5: 0x00},

	// decimal64 / minus
	{ID: "d64_minus_signed_zero_half_even", Width: "decimal64", Op: "minus", Category: "signed_zero", Mode: "half_even", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d64_minus_signed_zero_floor", Width: "decimal64", Op: "minus", Category: "signed_zero", Mode: "floor", Input: "0E+3", Expected: "-0E+3", BID5: 0x00},
	{ID: "d64_minus_signed_zero_ceiling", Width: "decimal64", Op: "minus", Category: "signed_zero", Mode: "ceiling", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d64_minus_signed_zero_down", Width: "decimal64", Op: "minus", Category: "signed_zero", Mode: "down", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d64_minus_signed_zero_half_up", Width: "decimal64", Op: "minus", Category: "signed_zero", Mode: "half_up", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d64_minus_quantum_positive", Width: "decimal64", Op: "minus", Category: "quantum_positive", Mode: "half_even", Input: "56267E+1", Expected: "-5.6267E+5", BID5: 0x00},
	{ID: "d64_minus_quantum_zero", Width: "decimal64", Op: "minus", Category: "quantum_zero", Mode: "half_even", Input: "56267", Expected: "-56267", BID5: 0x00},
	{ID: "d64_minus_quantum_negative", Width: "decimal64", Op: "minus", Category: "quantum_negative", Mode: "half_even", Input: "56267E-2", Expected: "-562.67", BID5: 0x00},
	{ID: "d64_minus_overprecision", Width: "decimal64", Op: "minus", Category: "overprecision", Mode: "half_even", Input: "12345678901234565E+3", Expected: "-1.234567890123456E+19", BID5: 0x20},
	{ID: "d64_minus_high_clamp_zero", Width: "decimal64", Op: "minus", Category: "high_clamp_zero", Mode: "half_even", Input: "0E+384", Expected: "0E+369", BID5: 0x00},
	{ID: "d64_minus_high_clamp_nonzero", Width: "decimal64", Op: "minus", Category: "high_clamp_nonzero", Mode: "half_even", Input: "1E+384", Expected: "-1.000000000000000E+384", BID5: 0x00},
	{ID: "d64_minus_etiny", Width: "decimal64", Op: "minus", Category: "etiny", Mode: "half_even", Input: "1E-398", Expected: "-1E-398", BID5: 0x00},
	{ID: "d64_minus_quiet_nan", Width: "decimal64", Op: "minus", Category: "quiet_nan", Mode: "half_even", Input: "NaN13", Expected: "NaN13", BID5: 0x00},
	{ID: "d64_minus_signaling_nan", Width: "decimal64", Op: "minus", Category: "signaling_nan", Mode: "half_even", Input: "sNaN13", Expected: "NaN13", BID5: 0x01},
	{ID: "d64_minus_positive_infinity", Width: "decimal64", Op: "minus", Category: "positive_infinity", Mode: "half_even", Input: "Infinity", Expected: "-Infinity", BID5: 0x00},
	{ID: "d64_minus_negative_infinity", Width: "decimal64", Op: "minus", Category: "negative_infinity", Mode: "half_even", Input: "-Infinity", Expected: "Infinity", BID5: 0x00},

	// decimal128 / plus
	{ID: "d128_plus_signed_zero_half_even", Width: "decimal128", Op: "plus", Category: "signed_zero", Mode: "half_even", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d128_plus_signed_zero_floor", Width: "decimal128", Op: "plus", Category: "signed_zero", Mode: "floor", Input: "-0E+3", Expected: "-0E+3", BID5: 0x00},
	{ID: "d128_plus_signed_zero_ceiling", Width: "decimal128", Op: "plus", Category: "signed_zero", Mode: "ceiling", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d128_plus_signed_zero_down", Width: "decimal128", Op: "plus", Category: "signed_zero", Mode: "down", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d128_plus_signed_zero_half_up", Width: "decimal128", Op: "plus", Category: "signed_zero", Mode: "half_up", Input: "-0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d128_plus_quantum_positive", Width: "decimal128", Op: "plus", Category: "quantum_positive", Mode: "half_even", Input: "56267E+1", Expected: "5.6267E+5", BID5: 0x00},
	{ID: "d128_plus_quantum_zero", Width: "decimal128", Op: "plus", Category: "quantum_zero", Mode: "half_even", Input: "56267", Expected: "56267", BID5: 0x00},
	{ID: "d128_plus_quantum_negative", Width: "decimal128", Op: "plus", Category: "quantum_negative", Mode: "half_even", Input: "56267E-2", Expected: "562.67", BID5: 0x00},
	{ID: "d128_plus_overprecision", Width: "decimal128", Op: "plus", Category: "overprecision", Mode: "half_even", Input: "12345678901234567890123456789012345E+3", Expected: "1.234567890123456789012345678901234E+37", BID5: 0x20},
	{ID: "d128_plus_high_clamp_zero", Width: "decimal128", Op: "plus", Category: "high_clamp_zero", Mode: "half_even", Input: "0E+6144", Expected: "0E+6111", BID5: 0x00},
	{ID: "d128_plus_high_clamp_nonzero", Width: "decimal128", Op: "plus", Category: "high_clamp_nonzero", Mode: "half_even", Input: "1E+6144", Expected: "1.000000000000000000000000000000000E+6144", BID5: 0x00},
	{ID: "d128_plus_etiny", Width: "decimal128", Op: "plus", Category: "etiny", Mode: "half_even", Input: "1E-6176", Expected: "1E-6176", BID5: 0x00},
	{ID: "d128_plus_quiet_nan", Width: "decimal128", Op: "plus", Category: "quiet_nan", Mode: "half_even", Input: "NaN13", Expected: "NaN13", BID5: 0x00},
	{ID: "d128_plus_signaling_nan", Width: "decimal128", Op: "plus", Category: "signaling_nan", Mode: "half_even", Input: "sNaN13", Expected: "NaN13", BID5: 0x01},
	{ID: "d128_plus_positive_infinity", Width: "decimal128", Op: "plus", Category: "positive_infinity", Mode: "half_even", Input: "Infinity", Expected: "Infinity", BID5: 0x00},
	{ID: "d128_plus_negative_infinity", Width: "decimal128", Op: "plus", Category: "negative_infinity", Mode: "half_even", Input: "-Infinity", Expected: "-Infinity", BID5: 0x00},

	// decimal128 / minus
	{ID: "d128_minus_signed_zero_half_even", Width: "decimal128", Op: "minus", Category: "signed_zero", Mode: "half_even", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d128_minus_signed_zero_floor", Width: "decimal128", Op: "minus", Category: "signed_zero", Mode: "floor", Input: "0E+3", Expected: "-0E+3", BID5: 0x00},
	{ID: "d128_minus_signed_zero_ceiling", Width: "decimal128", Op: "minus", Category: "signed_zero", Mode: "ceiling", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d128_minus_signed_zero_down", Width: "decimal128", Op: "minus", Category: "signed_zero", Mode: "down", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d128_minus_signed_zero_half_up", Width: "decimal128", Op: "minus", Category: "signed_zero", Mode: "half_up", Input: "0E+3", Expected: "0E+3", BID5: 0x00},
	{ID: "d128_minus_quantum_positive", Width: "decimal128", Op: "minus", Category: "quantum_positive", Mode: "half_even", Input: "56267E+1", Expected: "-5.6267E+5", BID5: 0x00},
	{ID: "d128_minus_quantum_zero", Width: "decimal128", Op: "minus", Category: "quantum_zero", Mode: "half_even", Input: "56267", Expected: "-56267", BID5: 0x00},
	{ID: "d128_minus_quantum_negative", Width: "decimal128", Op: "minus", Category: "quantum_negative", Mode: "half_even", Input: "56267E-2", Expected: "-562.67", BID5: 0x00},
	{ID: "d128_minus_overprecision", Width: "decimal128", Op: "minus", Category: "overprecision", Mode: "half_even", Input: "12345678901234567890123456789012345E+3", Expected: "-1.234567890123456789012345678901234E+37", BID5: 0x20},
	{ID: "d128_minus_high_clamp_zero", Width: "decimal128", Op: "minus", Category: "high_clamp_zero", Mode: "half_even", Input: "0E+6144", Expected: "0E+6111", BID5: 0x00},
	{ID: "d128_minus_high_clamp_nonzero", Width: "decimal128", Op: "minus", Category: "high_clamp_nonzero", Mode: "half_even", Input: "1E+6144", Expected: "-1.000000000000000000000000000000000E+6144", BID5: 0x00},
	{ID: "d128_minus_etiny", Width: "decimal128", Op: "minus", Category: "etiny", Mode: "half_even", Input: "1E-6176", Expected: "-1E-6176", BID5: 0x00},
	{ID: "d128_minus_quiet_nan", Width: "decimal128", Op: "minus", Category: "quiet_nan", Mode: "half_even", Input: "NaN13", Expected: "NaN13", BID5: 0x00},
	{ID: "d128_minus_signaling_nan", Width: "decimal128", Op: "minus", Category: "signaling_nan", Mode: "half_even", Input: "sNaN13", Expected: "NaN13", BID5: 0x01},
	{ID: "d128_minus_positive_infinity", Width: "decimal128", Op: "minus", Category: "positive_infinity", Mode: "half_even", Input: "Infinity", Expected: "-Infinity", BID5: 0x00},
	{ID: "d128_minus_negative_infinity", Width: "decimal128", Op: "minus", Category: "negative_infinity", Mode: "half_even", Input: "-Infinity", Expected: "Infinity", BID5: 0x00},
}

// dectestPlusMinusStrengthCanonicalRow renders the language-neutral row string
// both generated runners re-render at run time and both generated literals pin.
// The Go and Rust renderers must reproduce it byte for byte.
func dectestPlusMinusStrengthCanonicalRow(row dectestPlusMinusStrengthRow) string {
	return fmt.Sprintf("%s %s mode=%s in=%s out=%s bid5=%08x",
		row.Width, row.Op, row.Mode, row.Input, row.Expected, row.BID5)
}

// validateDectestPlusMinusStrengthTable closes the table at generation time: the
// 3x2x16 shape, unique row identity, the per-cell category multiset, the
// rounding-mode vocabulary, and the compared flag surface.
func validateDectestPlusMinusStrengthTable() error {
	if len(dectestPlusMinusStrengthTable) != dectestPlusMinusStrengthTotalRows {
		return fmt.Errorf("dectest plus/minus strength table has %d rows, want %d",
			len(dectestPlusMinusStrengthTable), dectestPlusMinusStrengthTotalRows)
	}
	seenID := map[string]bool{}
	seenRow := map[string]bool{}
	cells := map[string][]string{}
	cellModes := map[string]map[string]bool{}
	for _, row := range dectestPlusMinusStrengthTable {
		if row.Width != "decimal32" && row.Width != "decimal64" && row.Width != "decimal128" {
			return fmt.Errorf("dectest plus/minus strength row %q has unknown width %q", row.ID, row.Width)
		}
		if row.Op != "plus" && row.Op != "minus" {
			return fmt.Errorf("dectest plus/minus strength row %q has unknown operation %q", row.ID, row.Op)
		}
		if _, ok := dectestPlusMinusStrengthModes[row.Mode]; !ok {
			return fmt.Errorf("dectest plus/minus strength row %q has unknown rounding mode %q", row.ID, row.Mode)
		}
		if row.Input == "" || row.Expected == "" {
			return fmt.Errorf("dectest plus/minus strength row %q has an empty operand or expected result", row.ID)
		}
		if row.BID5&^dectestPlusMinusStrengthBID5Mask != 0 {
			return fmt.Errorf("dectest plus/minus strength row %q pins flags %#08x outside the BID five-flag surface", row.ID, row.BID5)
		}
		if seenID[row.ID] {
			return fmt.Errorf("dectest plus/minus strength row id %q is duplicated", row.ID)
		}
		seenID[row.ID] = true
		canonical := dectestPlusMinusStrengthCanonicalRow(row)
		if seenRow[canonical] {
			return fmt.Errorf("dectest plus/minus strength canonical row %q is duplicated", canonical)
		}
		seenRow[canonical] = true
		cell := row.Width + "/" + row.Op
		cells[cell] = append(cells[cell], row.Category)
		if row.Category == "signed_zero" {
			if cellModes[cell] == nil {
				cellModes[cell] = map[string]bool{}
			}
			if cellModes[cell][row.Mode] {
				return fmt.Errorf("dectest plus/minus strength cell %s repeats signed-zero rounding mode %q", cell, row.Mode)
			}
			cellModes[cell][row.Mode] = true
		} else if row.Mode != "half_even" {
			return fmt.Errorf("dectest plus/minus strength row %q is a %s shape row at rounding mode %q, want half_even", row.ID, row.Category, row.Mode)
		}
	}
	if len(cells) != dectestPlusMinusStrengthCellCount {
		return fmt.Errorf("dectest plus/minus strength table covers %d width/operation cells, want %d",
			len(cells), dectestPlusMinusStrengthCellCount)
	}
	wantCategories := append([]string(nil), dectestPlusMinusStrengthCellCategories...)
	sort.Strings(wantCategories)
	for cell, categories := range cells {
		got := append([]string(nil), categories...)
		sort.Strings(got)
		if strings.Join(got, ",") != strings.Join(wantCategories, ",") {
			return fmt.Errorf("dectest plus/minus strength cell %s carries categories [%s], want [%s]",
				cell, strings.Join(got, ","), strings.Join(wantCategories, ","))
		}
		if len(cellModes[cell]) != len(dectestPlusMinusStrengthModes) {
			return fmt.Errorf("dectest plus/minus strength cell %s covers %d signed-zero rounding modes, want %d",
				cell, len(cellModes[cell]), len(dectestPlusMinusStrengthModes))
		}
	}
	return nil
}

func dectestPlusMinusStrengthGoRowLiteral() string {
	var b strings.Builder
	for _, row := range dectestPlusMinusStrengthTable {
		fmt.Fprintf(&b, "\t%q,\n", dectestPlusMinusStrengthCanonicalRow(row))
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func dectestPlusMinusStrengthGoCaseLiteral() string {
	var b strings.Builder
	for _, row := range dectestPlusMinusStrengthTable {
		fmt.Fprintf(&b, "\t{id: %q, width: %q, op: %q, mode: %q, modeNum: %d, input: %q, expected: %q, bid5: %#02x},\n",
			row.ID, row.Width, row.Op, row.Mode, dectestPlusMinusStrengthModes[row.Mode], row.Input, row.Expected, row.BID5)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func dectestPlusMinusStrengthRustRowLiteral() string {
	var b strings.Builder
	for _, row := range dectestPlusMinusStrengthTable {
		fmt.Fprintf(&b, "    %q,\n", dectestPlusMinusStrengthCanonicalRow(row))
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func dectestPlusMinusStrengthRustCaseLiteral() string {
	var b strings.Builder
	for _, row := range dectestPlusMinusStrengthTable {
		fmt.Fprintf(&b, "    PlusMinusStrengthCase { id: %q, width: %q, op: %q, mode: %q, mode_num: %d, input: %q, expected: %q, bid5: %#02x },\n",
			row.ID, row.Width, row.Op, row.Mode, dectestPlusMinusStrengthModes[row.Mode], row.Input, row.Expected, row.BID5)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// dectestPlusMinusStrengthGoSource emits the anchor into the portable Go decTest
// runner, which executes every row twice: through runDectestGoportCase and
// through executeDecTestUnaryOperation, the two production adapter paths the
// decTest domain dispatches plus/minus to.
func dectestPlusMinusStrengthGoSource() (string, error) {
	if err := validateDectestPlusMinusStrengthTable(); err != nil {
		return "", err
	}
	return strings.NewReplacer(
		"@@PLUS_MINUS_STRENGTH_ROWS@@", dectestPlusMinusStrengthGoRowLiteral(),
		"@@PLUS_MINUS_STRENGTH_CASES@@", dectestPlusMinusStrengthGoCaseLiteral(),
		"@@PLUS_MINUS_STRENGTH_COUNT@@", fmt.Sprintf("%d", dectestPlusMinusStrengthTotalRows),
		"@@PLUS_MINUS_STRENGTH_BID5_MASK@@", fmt.Sprintf("%#02x", dectestPlusMinusStrengthBID5Mask),
	).Replace(dectestPlusMinusStrengthGoTemplate), nil
}

// dectestPlusMinusStrengthRustSource emits the same anchor into the generated
// Rust decTest runner, which executes every row through run_dectest_goport_case.
func dectestPlusMinusStrengthRustSource() (string, error) {
	if err := validateDectestPlusMinusStrengthTable(); err != nil {
		return "", err
	}
	return strings.NewReplacer(
		"@@PLUS_MINUS_STRENGTH_ROWS@@", dectestPlusMinusStrengthRustRowLiteral(),
		"@@PLUS_MINUS_STRENGTH_CASES@@", dectestPlusMinusStrengthRustCaseLiteral(),
		"@@PLUS_MINUS_STRENGTH_COUNT@@", fmt.Sprintf("%d", dectestPlusMinusStrengthTotalRows),
		"@@PLUS_MINUS_STRENGTH_BID5_MASK@@", fmt.Sprintf("%#02x", dectestPlusMinusStrengthBID5Mask),
	).Replace(dectestPlusMinusStrengthRustTemplate), nil
}

var dectestPlusMinusStrengthGoTemplate = `

// plus/minus quantum strength anchor: hand-pinned known answers derived
// independently from IBM decNumber 3.68, never computed from this port. They
// close what the consumed suites leave open -- the Decimal32 suite carries no
// q>0 zero plus/minus witness, and the regular decTest runner compares
// executeDecTestUnaryOperation on value and the BID five-flag surface but never
// on quantum. Every row runs through both production adapter paths
// (runDectestGoportCase and executeDecTestUnaryOperation) with no skip or
// flag-exemption classifier in between, and
// devtools/verification_sentinels.json pins the identical canonical rows outside
// every generation path.

const (
	dectestPlusMinusStrengthRowCount        = uint64(@@PLUS_MINUS_STRENGTH_COUNT@@)
	dectestPlusMinusStrengthBID5Mask uint32 = @@PLUS_MINUS_STRENGTH_BID5_MASK@@
)

var dectestPlusMinusStrengthRows = []string{
@@PLUS_MINUS_STRENGTH_ROWS@@
}

type dectestPlusMinusStrengthCase struct {
	id       string
	width    string
	op       string
	mode     string
	modeNum  int
	input    string
	expected string
	bid5     uint32
}

var dectestPlusMinusStrengthCases = []dectestPlusMinusStrengthCase{
@@PLUS_MINUS_STRENGTH_CASES@@
}

func dectestPlusMinusStrengthCanonicalRow(c dectestPlusMinusStrengthCase) string {
	return fmt.Sprintf("%s %s mode=%s in=%s out=%s bid5=%08x", c.width, c.op, c.mode, c.input, c.expected, c.bid5)
}

// dectestPlusMinusStrengthExpectedFlags translates the pinned raw Intel BID
// five-flag value into this leg's ExceptionFlags encoding, which numbers the
// same five bits differently.
func dectestPlusMinusStrengthExpectedFlags(t *testing.T, bid5 uint32) ExceptionFlags {
	t.Helper()
	if bid5&^dectestPlusMinusStrengthBID5Mask != 0 {
		t.Fatalf("pinned flag value %#08x carries a bit outside the BID five-flag surface", bid5)
	}
	var flags ExceptionFlags
	if bid5&0x01 != 0 {
		flags |= FlagInvalidOperation
	}
	if bid5&0x04 != 0 {
		flags |= FlagDivisionByZero
	}
	if bid5&0x08 != 0 {
		flags |= FlagOverflow
	}
	if bid5&0x10 != 0 {
		flags |= FlagUnderflow
	}
	if bid5&0x20 != 0 {
		flags |= FlagInexact
	}
	return flags
}

func dectestPlusMinusStrengthCheckTableShape(t *testing.T) {
	t.Helper()
	if uint64(len(dectestPlusMinusStrengthRows)) != dectestPlusMinusStrengthRowCount {
		t.Fatalf("pinned row literal count %d diverges from the generated constant %d", len(dectestPlusMinusStrengthRows), dectestPlusMinusStrengthRowCount)
	}
	if uint64(len(dectestPlusMinusStrengthCases)) != dectestPlusMinusStrengthRowCount {
		t.Fatalf("executable case count %d diverges from the generated constant %d", len(dectestPlusMinusStrengthCases), dectestPlusMinusStrengthRowCount)
	}
}

// dectestPlusMinusStrengthCheckPin validates one row's own pins before it is
// executed: its canonical rendering against the pinned literal, and the BID
// rounding mode the runners resolve from its token at run time.
func dectestPlusMinusStrengthCheckPin(t *testing.T, index int, c dectestPlusMinusStrengthCase) {
	t.Helper()
	if got := dectestPlusMinusStrengthCanonicalRow(c); got != dectestPlusMinusStrengthRows[index] {
		t.Fatalf("case renders %q, pinned row %q", got, dectestPlusMinusStrengthRows[index])
	}
	mode, ok := decTestBIDRoundingMode(c.mode)
	if !ok {
		t.Fatalf("decTestBIDRoundingMode(%q) rejected the pinned rounding token", c.mode)
	}
	if mode != c.modeNum {
		t.Fatalf("decTestBIDRoundingMode(%q) = %d, pinned %d", c.mode, mode, c.modeNum)
	}
}

func dectestPlusMinusStrengthDecTestCase(c dectestPlusMinusStrengthCase) decTestCase {
	return decTestCase{
		ID:           c.id,
		Operation:    c.op,
		Operands:     []string{c.input},
		RoundingMode: c.mode,
	}
}

// dectestPlusMinusStrengthCheckOutcome is the single assertion semantics both
// adapter legs use: value, exact quantum, and the BID five-flag surface.
func dectestPlusMinusStrengthCheckOutcome(t *testing.T, c dectestPlusMinusStrengthCase, got string, gotFlags ExceptionFlags) {
	t.Helper()
	want := dectestPlusMinusStrengthExpectedFlags(t, c.bid5)
	if !compareDecimalResults(c.expected, got) {
		t.Fatalf("value mismatch: expected %q, adapter produced %q", c.expected, got)
	}
	if !dectestGoportQuantumEqual(c.expected, got) {
		t.Fatalf("quantum mismatch: expected %q, adapter produced %q (same value, different cohort member)", c.expected, got)
	}
	if gotFlags&dectestGoportBIDFlagMask != want {
		t.Fatalf("flag mismatch: expected %s, adapter raised %s", want.String(), (gotFlags & dectestGoportBIDFlagMask).String())
	}
}

func TestGeneratedDectestPlusMinusQuantumStrengthGoPort(t *testing.T) {
	dectestPlusMinusStrengthCheckTableShape(t)
	for i, c := range dectestPlusMinusStrengthCases {
		t.Run(c.id, func(t *testing.T) {
			dectestPlusMinusStrengthCheckPin(t, i, c)
			got, gotFlags, err := runDectestGoportCase(dectestPlusMinusStrengthDecTestCase(c), c.width)
			if err != nil {
				t.Fatalf("runDectestGoportCase(%s %s %s): %v", c.width, c.op, c.input, err)
			}
			dectestPlusMinusStrengthCheckOutcome(t, c, got, gotFlags)
		})
	}
}

// TestGeneratedDectestPlusMinusQuantumStrengthUnaryAdapter re-executes the same
// rows through executeDecTestUnaryOperation, the second adapter the decTest
// domain dispatches plus/minus to. Its regular runner (runDecTestCaseV2) is
// contracted to compare value and the BID five-flag surface only, so a
// quantum regression confined to this adapter is invisible there and is caught
// only here.
func TestGeneratedDectestPlusMinusQuantumStrengthUnaryAdapter(t *testing.T) {
	dectestPlusMinusStrengthCheckTableShape(t)
	for i, c := range dectestPlusMinusStrengthCases {
		t.Run(c.id, func(t *testing.T) {
			dectestPlusMinusStrengthCheckPin(t, i, c)
			exec, err := executeDecTestUnaryOperation(dectestPlusMinusStrengthDecTestCase(c), c.width)
			if err != nil {
				t.Fatalf("executeDecTestUnaryOperation(%s %s %s): %v", c.width, c.op, c.input, err)
			}
			dectestPlusMinusStrengthCheckOutcome(t, c, exec.Result, exec.Flags)
		})
	}
}
`

var dectestPlusMinusStrengthRustTemplate = `

// plus/minus quantum strength anchor: hand-pinned known answers derived
// independently from IBM decNumber 3.68, never computed from this engine. They
// close what the consumed suites leave open -- the Decimal32 suite carries no
// q>0 zero plus/minus witness, and no suite closes both operations across all
// three widths, all five rounding modes, and the clamp/etiny/special boundaries
// at once. Every row runs through run_dectest_goport_case with no skip or
// flag-exemption classifier in between. The canonical rows are byte-identical to
// the Go leg's literal and pinned outside every generation path in
// devtools/verification_sentinels.json.

const DECTEST_PLUS_MINUS_STRENGTH_ROW_COUNT: usize = @@PLUS_MINUS_STRENGTH_COUNT@@;
const DECTEST_PLUS_MINUS_STRENGTH_BID5_MASK: u32 = @@PLUS_MINUS_STRENGTH_BID5_MASK@@;

const DECTEST_PLUS_MINUS_STRENGTH_ROWS: [&str; @@PLUS_MINUS_STRENGTH_COUNT@@] = [
@@PLUS_MINUS_STRENGTH_ROWS@@
];

struct PlusMinusStrengthCase {
    id: &'static str,
    width: &'static str,
    op: &'static str,
    mode: &'static str,
    mode_num: i64,
    input: &'static str,
    expected: &'static str,
    bid5: u32,
}

const DECTEST_PLUS_MINUS_STRENGTH_CASES: [PlusMinusStrengthCase; @@PLUS_MINUS_STRENGTH_COUNT@@] = [
@@PLUS_MINUS_STRENGTH_CASES@@
];

fn dectest_plus_minus_strength_canonical_row(c: &PlusMinusStrengthCase) -> String {
    format!(
        "{} {} mode={} in={} out={} bid5={:08x}",
        c.width, c.op, c.mode, c.input, c.expected, c.bid5
    )
}

#[test]
fn generated_dectest_plus_minus_quantum_strength_go_port() {
    assert_eq!(
        DECTEST_PLUS_MINUS_STRENGTH_ROWS.len(),
        DECTEST_PLUS_MINUS_STRENGTH_ROW_COUNT,
        "pinned row literal count diverges from the generated constant"
    );
    assert_eq!(
        DECTEST_PLUS_MINUS_STRENGTH_CASES.len(),
        DECTEST_PLUS_MINUS_STRENGTH_ROW_COUNT,
        "executable case count diverges from the generated constant"
    );
    for (i, c) in DECTEST_PLUS_MINUS_STRENGTH_CASES.iter().enumerate() {
        assert_eq!(
            dectest_plus_minus_strength_canonical_row(c),
            DECTEST_PLUS_MINUS_STRENGTH_ROWS[i],
            "case {} renders a different row than the pinned literal",
            c.id
        );
        assert_eq!(
            c.bid5 & !DECTEST_PLUS_MINUS_STRENGTH_BID5_MASK,
            0,
            "case {} pins a flag bit outside the BID five-flag surface",
            c.id
        );
        assert_eq!(
            dec_test_bid_rounding_mode(c.mode),
            Some(c.mode_num),
            "case {} rounding token {:?} resolves to a different BID mode than the pinned number",
            c.id,
            c.mode
        );
        let tc = DecTestCase {
            id: c.id.to_string(),
            operation: c.op.to_string(),
            operands: vec![c.input.to_string()],
            result: String::new(),
            flags: Vec::new(),
            rounding_mode: c.mode.to_string(),
            precision: 0,
            max_exponent: 0,
        };
        let (got, got_flags) = run_dectest_goport_case(&tc, c.width)
            .unwrap_or_else(|e| panic!("run_dectest_goport_case({}): {}", c.id, e));
        assert!(
            compare_decimal_results(c.expected, &got),
            "case {} value mismatch: expected {:?}, engine produced {:?}",
            c.id,
            c.expected,
            got
        );
        assert!(
            quantum_equal(c.expected, &got),
            "case {} quantum mismatch: expected {:?}, engine produced {:?} (same value, different cohort member)",
            c.id,
            c.expected,
            got
        );
        assert_eq!(
            got_flags & FLAG_MASK,
            c.bid5,
            "case {} flag mismatch: expected {:#010x}, engine raised {:#010x}",
            c.id,
            c.bid5,
            got_flags & FLAG_MASK
        );
    }
}
`
