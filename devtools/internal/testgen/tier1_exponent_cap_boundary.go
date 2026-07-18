package testgen

import "fmt"

// Tier 1 arithmetic exponent-cap and exponent-floor boundary extension.
//
// The base boundary sets (tier1ArithmeticBoundary32Values and the BID codec
// edge corpora) cross raw exponent fields only near {min, bias, max}, so no
// finite boundary value had a product or FMA product against the +/-Nmax
// probes landing just above the width's exponent cap. That gap left the Intel
// bid128_fma.c Case (1”B) p_sign != z_sign overflow exits (lines 2117-2136
// and 2245-2264, "BID_SWAP128 (res); BID_RETURN (res)") without a Tier 1
// differential row, and the Go-port transcription defect fixed in
// bid754-go/internal/bidgo/bid128_fma_body.go (the dropped BID_RETURN:
// FMA(1E+34, +/-Nmax, -/+Nmax) returned non-canonical bits) passed the full
// Tier 1 arithmetic long gates. The values below close that class for the
// arithmetic runners: coefficients {1, 10^(p-1), 10^p-1} crossed with decimal
// exponents just above the overflow threshold, in both signs, so the
// boundary x probe pairs (add/sub/mul/div/quantize) and the rotated FMA
// triples against the +/-Nmax probes take the overflow clamp paths of every
// width. Values already present in the base sets are absorbed by the
// tier1SharedLongBoundary*Values composition below, which both the
// arithmetic and the compare/conversion long codegens consume (the
// compare/conversion Go runner iterates the arithmetic runner's emitted
// tables at runtime, so their generation paths must agree); the decNumber
// differential corpus and the BID codec vectors keep consuming the
// unchanged base sets.
//
// Layout contract (checked by tier1ArithmeticCheckExponentCapResidues): the
// runners' structured FMA block pairs boundary index i with the probe
// companion pair (probes[j], probes[(i+j)%len(probes)]), so a boundary value
// meets the (+Nmax, -Nmax) or (-Nmax, +Nmax) companions only at two specific
// index residues. The coefficient-1 members are therefore emitted per sign as
// one contiguous run of at least len(probes) consecutive decimal exponents
// around +p, which covers every residue by construction — independent of the
// base-set length — and guarantees FMA(1E+k, +/-Nmax, -/+Nmax) rows whose
// product exponent sits just above the cap. This spreads differential rows
// across the near-cap FMA case space; it does not by itself pin any single
// exponent onto a required residue, so the exact D2 counterexample triples
// (only x = ±1E+34 reaches the fixed exits — measured, see the semantic FMA
// block in tier1ArithmeticSemanticCorpus) are pinned index-independently in
// the semantic corpus instead. A base-set or probe change that breaks this
// residue contract fails generation instead of silently thinning the
// near-cap FMA coverage.

// tier1ArithmeticExponentCapCoefficient1Exponents returns the coefficient-1
// decimal exponents: {+1, +2} (products with Nmax overflow by the smallest
// margins), one contiguous run p-5..p+6 of 12 exponents around +p (the
// Bid128Fma Case (1”B) counterexample shape and full FMA-companion residue
// coverage), and {2p-1, 2p} (product exponent about 2p above the cap).
func tier1ArithmeticExponentCapCoefficient1Exponents(precision int) []int {
	exponents := []int{1, 2}
	for k := precision - 5; k <= precision+6; k++ {
		if k > 2 {
			exponents = append(exponents, k)
		}
	}
	return append(exponents, 2*precision-1, 2*precision)
}

// tier1ArithmeticExponentCapFullCoefficientExponents returns the decimal
// exponents crossed with the full-width coefficients 10^(p-1) and 10^p-1:
// the same threshold classes without the residue-coverage run.
func tier1ArithmeticExponentCapFullCoefficientExponents(precision int) []int {
	return []int{1, 2, precision - 1, precision, precision + 1, 2*precision - 1, 2 * precision}
}

// The exponent-floor mirror: products of these members against the
// width's minimum-subnormal probe (coefficient 1, minimum exponent) land in
// the partial-underflow shift window just below expmin — the region that
// drives the e4 < expmin subnormal rounding paths of bid_add_and_round
// (bid128_fma_body.go), the division remainder underflow handling, and the
// wide-coefficient underflow rounding, none of which the base sets reached
// (the base exponent fields jump from {min, min+1, min+2} straight to the
// bias region, so no product ever lands 1..2p decades below the floor).
// Unlike the cap side there is no FMA-companion residue requirement: the
// rounded pair block (mul among the five rounded operations) already crosses
// every boundary member with every probe in both orders, which is where the
// partial-underflow products arise.

// tier1ArithmeticExponentFloorCoefficient1Exponents returns the negated
// coefficient-1 exponent classes: 10^-k with the same magnitudes as the cap
// run, so member x minimum-subnormal products sit 1..2p decades below the
// width's minimum exponent.
func tier1ArithmeticExponentFloorCoefficient1Exponents(precision int) []int {
	exponents := tier1ArithmeticExponentCapCoefficient1Exponents(precision)
	negated := make([]int, len(exponents))
	for i, exponent := range exponents {
		negated[i] = -exponent
	}
	return negated
}

// tier1ArithmeticExponentFloorFullCoefficientExponents returns the negated
// full-coefficient exponent classes of the floor mirror.
func tier1ArithmeticExponentFloorFullCoefficientExponents(precision int) []int {
	exponents := tier1ArithmeticExponentCapFullCoefficientExponents(precision)
	negated := make([]int, len(exponents))
	for i, exponent := range exponents {
		negated[i] = -exponent
	}
	return negated
}

// tier1ArithmeticExponentCapFloorBoundary32Values builds the BID32 extension
// (p = 7, bias 101): coefficient 1 and 10^6 in the small form, and
// 10^7-1 = 9999999 = 0x800000|0x18967f in the steered large-coefficient form.
// Per sign the coefficient-1 run comes first so its appended indices stay
// contiguous (the residue-coverage contract above).
func tier1ArithmeticExponentCapFloorBoundary32Values() []uint32 {
	var values []uint32
	for _, sign := range [...]uint32{0, 0x80000000} {
		for _, exponent := range tier1ArithmeticExponentCapCoefficient1Exponents(7) {
			values = append(values, sign|uint32(101+exponent)<<23|1)
		}
		for _, exponent := range tier1ArithmeticExponentCapFullCoefficientExponents(7) {
			field := uint32(101 + exponent)
			values = append(values,
				sign|field<<23|1000000,
				sign|0x60000000|field<<21|0x0018967f,
			)
		}
		for _, exponent := range tier1ArithmeticExponentFloorCoefficient1Exponents(7) {
			values = append(values, sign|uint32(101+exponent)<<23|1)
		}
		for _, exponent := range tier1ArithmeticExponentFloorFullCoefficientExponents(7) {
			field := uint32(101 + exponent)
			values = append(values,
				sign|field<<23|1000000,
				sign|0x60000000|field<<21|0x0018967f,
			)
		}
	}
	return values
}

// tier1ArithmeticExponentCapFloorBoundary64Values builds the BID64 extension
// (p = 16, bias 398): coefficient 1 and 10^15 in the small form, and
// 10^16-1 = 0x20000000000000|0x386f26fc0ffff in the steered form.
func tier1ArithmeticExponentCapFloorBoundary64Values() []uint64 {
	var values []uint64
	for _, sign := range [...]uint64{0, 0x8000000000000000} {
		for _, exponent := range tier1ArithmeticExponentCapCoefficient1Exponents(16) {
			values = append(values, sign|uint64(398+exponent)<<53|1)
		}
		for _, exponent := range tier1ArithmeticExponentCapFullCoefficientExponents(16) {
			field := uint64(398 + exponent)
			values = append(values,
				sign|field<<53|1000000000000000,
				sign|0x6000000000000000|field<<51|0x000386f26fc0ffff,
			)
		}
		for _, exponent := range tier1ArithmeticExponentFloorCoefficient1Exponents(16) {
			values = append(values, sign|uint64(398+exponent)<<53|1)
		}
		for _, exponent := range tier1ArithmeticExponentFloorFullCoefficientExponents(16) {
			field := uint64(398 + exponent)
			values = append(values,
				sign|field<<53|1000000000000000,
				sign|0x6000000000000000|field<<51|0x000386f26fc0ffff,
			)
		}
	}
	return values
}

// tier1ArithmeticExponentCapFloorBoundary128Values builds the BID128 extension
// (p = 34, bias 6176). All three coefficients fit the small form:
// 1, 10^33 = 0x314dc6448d93:38c15b0a00000000, and
// 10^34-1 = 0x1ed09bead87c0:378d8e63ffffffff (the Nmax coefficient).
// The (coefficient 1, exponent +34) member is the pinned D2 counterexample
// operand 0x3084000000000000:0000000000000001 from
// bid754-go/internal/bidgo/bid128_fma_overflow_test.go.
func tier1ArithmeticExponentCapFloorBoundary128Values() []bid128BidCodecValue {
	fullCoefficients := [...]bid128BidCodecValue{
		{lo: 0x38c15b0a00000000, hi: 0x0000314dc6448d93},
		{lo: 0x378d8e63ffffffff, hi: 0x0001ed09bead87c0},
	}
	var values []bid128BidCodecValue
	for _, sign := range [...]uint64{0, 0x8000000000000000} {
		for _, exponent := range tier1ArithmeticExponentCapCoefficient1Exponents(34) {
			values = append(values, bid128BidCodecValue{lo: 1, hi: sign | uint64(6176+exponent)<<49})
		}
		for _, exponent := range tier1ArithmeticExponentCapFullCoefficientExponents(34) {
			field := uint64(6176 + exponent)
			for _, coefficient := range fullCoefficients {
				values = append(values, bid128BidCodecValue{
					lo: coefficient.lo,
					hi: sign | field<<49 | coefficient.hi,
				})
			}
		}
		for _, exponent := range tier1ArithmeticExponentFloorCoefficient1Exponents(34) {
			values = append(values, bid128BidCodecValue{lo: 1, hi: sign | uint64(6176+exponent)<<49})
		}
		for _, exponent := range tier1ArithmeticExponentFloorFullCoefficientExponents(34) {
			field := uint64(6176 + exponent)
			for _, coefficient := range fullCoefficients {
				values = append(values, bid128BidCodecValue{
					lo: coefficient.lo,
					hi: sign | field<<49 | coefficient.hi,
				})
			}
		}
	}
	return values
}

// tier1SharedLongBoundary32Values is the single composition point of the
// shared Tier 1 long boundary set (base + exponent-cap/floor extension) for
// BID32. Both long runners consume it: the arithmetic codegen emits the
// value tables (shared at runtime by the compare/conversion Go runner, which
// cross-checks the table length against its own generated count), and the
// compare/conversion codegen derives its counts and the Rust-embedded value
// tables from it. Composing in one place keeps those two generation paths
// from silently diverging again.
func tier1SharedLongBoundary32Values() []uint32 {
	return appendUnique32(tier1ArithmeticBoundary32Values(), tier1ArithmeticExponentCapFloorBoundary32Values())
}

// tier1SharedLongBoundary64Values is the BID64 shared composition point.
func tier1SharedLongBoundary64Values() []uint64 {
	return appendUnique64(bid64BidCodecEdgeValues(), tier1ArithmeticExponentCapFloorBoundary64Values())
}

// tier1SharedLongBoundary128Values is the BID128 shared composition point.
func tier1SharedLongBoundary128Values() []bid128BidCodecValue {
	return appendUnique128(bid128BidCodecEdgeValues(), tier1ArithmeticExponentCapFloorBoundary128Values())
}

// appendUnique32 mirrors appendUnique64 for the BID32 boundary composition:
// base order is preserved and only extension values not already present are
// appended, in their declared order.
func appendUnique32(base, extra []uint32) []uint32 {
	result := append([]uint32(nil), base...)
	seen := make(map[uint32]bool, len(base)+len(extra))
	for _, value := range result {
		seen[value] = true
	}
	for _, value := range extra {
		if !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result
}

// appendUnique128 mirrors appendUnique64 for the BID128 boundary composition.
func appendUnique128(base, extra []bid128BidCodecValue) []bid128BidCodecValue {
	result := append([]bid128BidCodecValue(nil), base...)
	seen := make(map[bid128BidCodecValue]bool, len(base)+len(extra))
	for _, value := range result {
		seen[value] = true
	}
	for _, value := range extra {
		if !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result
}

// tier1ArithmeticVerifyExponentCapContract checks the residue contract for
// all three composed boundary sets against the runner probe tables. It
// derives the +/-Nmax probe values from the width constants and fails if a
// probe-set edit removed them, so the contract cannot degrade into a no-op.
func tier1ArithmeticVerifyExponentCapContract(boundary32 []uint32, boundary64 []uint64, boundary128 []bid128BidCodecValue) error {
	probeIndex32 := func(want uint32) int {
		for i, value := range tier1ArithmeticProbes32Values {
			if value == want {
				return i
			}
		}
		return -1
	}
	probeIndex64 := func(want uint64) int {
		for i, value := range tier1ArithmeticProbes64Values {
			if value == want {
				return i
			}
		}
		return -1
	}
	probeIndex128 := func(want bid128BidCodecValue) int {
		for i, value := range tier1ArithmeticProbes128Values {
			if value == want {
				return i
			}
		}
		return -1
	}
	record := func(byResidue map[int]map[bool]bool, index, probeCount int, negative bool) {
		residue := index % probeCount
		if byResidue[residue] == nil {
			byResidue[residue] = map[bool]bool{}
		}
		byResidue[residue][negative] = true
	}

	// BID32: Nmax = 9999999E+90 in the steered form (exponent field 191).
	nmaxPlus32 := uint32(0x60000000 | uint32(101+90)<<21 | 0x0018967f)
	plus32, minus32 := probeIndex32(nmaxPlus32), probeIndex32(nmaxPlus32|0x80000000)
	if plus32 < 0 || minus32 < 0 {
		return fmt.Errorf("Tier 1 arithmetic width 32 exponent-cap contract: +/-Nmax probes not found in tier1ArithmeticProbes32Values")
	}
	byResidue32 := map[int]map[bool]bool{}
	for index, value := range boundary32 {
		for _, sign := range [...]uint32{0, 0x80000000} {
			for k := 7 - 5; k <= 7+6; k++ {
				if value == sign|uint32(101+k)<<23|1 {
					record(byResidue32, index, len(tier1ArithmeticProbes32Values), sign != 0)
				}
			}
		}
	}
	if err := tier1ArithmeticCheckExponentCapResidues(32, len(tier1ArithmeticProbes32Values), plus32, minus32, byResidue32); err != nil {
		return err
	}

	// BID64: Nmax = 9999999999999999E+369 in the steered form (field 767).
	nmaxPlus64 := uint64(0x6000000000000000 | uint64(398+369)<<51 | 0x000386f26fc0ffff)
	plus64, minus64 := probeIndex64(nmaxPlus64), probeIndex64(nmaxPlus64|0x8000000000000000)
	if plus64 < 0 || minus64 < 0 {
		return fmt.Errorf("Tier 1 arithmetic width 64 exponent-cap contract: +/-Nmax probes not found in tier1ArithmeticProbes64Values")
	}
	byResidue64 := map[int]map[bool]bool{}
	for index, value := range boundary64 {
		for _, sign := range [...]uint64{0, 0x8000000000000000} {
			for k := 16 - 5; k <= 16+6; k++ {
				if value == sign|uint64(398+k)<<53|1 {
					record(byResidue64, index, len(tier1ArithmeticProbes64Values), sign != 0)
				}
			}
		}
	}
	if err := tier1ArithmeticCheckExponentCapResidues(64, len(tier1ArithmeticProbes64Values), plus64, minus64, byResidue64); err != nil {
		return err
	}

	// BID128: Nmax = (10^34-1)E+6111 in the small form (field 12287).
	nmaxPlus128 := bid128BidCodecValue{lo: 0x378d8e63ffffffff, hi: uint64(6176+6111)<<49 | 0x0001ed09bead87c0}
	nmaxMinus128 := bid128BidCodecValue{lo: nmaxPlus128.lo, hi: nmaxPlus128.hi | 0x8000000000000000}
	plus128, minus128 := probeIndex128(nmaxPlus128), probeIndex128(nmaxMinus128)
	if plus128 < 0 || minus128 < 0 {
		return fmt.Errorf("Tier 1 arithmetic width 128 exponent-cap contract: +/-Nmax probes not found in tier1ArithmeticProbes128Values")
	}
	byResidue128 := map[int]map[bool]bool{}
	for index, value := range boundary128 {
		for _, sign := range [...]uint64{0, 0x8000000000000000} {
			for k := 34 - 5; k <= 34+6; k++ {
				if value == (bid128BidCodecValue{lo: 1, hi: sign | uint64(6176+k)<<49}) {
					record(byResidue128, index, len(tier1ArithmeticProbes128Values), sign != 0)
				}
			}
		}
	}
	return tier1ArithmeticCheckExponentCapResidues(128, len(tier1ArithmeticProbes128Values), plus128, minus128, byResidue128)
}

// tier1ArithmeticCheckExponentCapResidues enforces the FMA-companion residue
// contract on one composed boundary set: for both required residues (the two
// orderings of the +/-Nmax companion pair) there must be a coefficient-1
// exponent-cap member 1E+k, k in p-5..p+6, of each sign at that residue.
// probeCount is the runner probe count, nmaxPlusIndex/nmaxMinusIndex the
// probe positions of +Nmax/-Nmax, and capSignsByResidue maps each composed
// boundary index residue of a near-cap coefficient-1 member to the signs
// present there (negative=true).
func tier1ArithmeticCheckExponentCapResidues(width int, probeCount, nmaxPlusIndex, nmaxMinusIndex int, capSignsByResidue map[int]map[bool]bool) error {
	for _, residue := range []int{
		((nmaxMinusIndex-nmaxPlusIndex)%probeCount + probeCount) % probeCount,
		((nmaxPlusIndex-nmaxMinusIndex)%probeCount + probeCount) % probeCount,
	} {
		signs := capSignsByResidue[residue]
		for negative, label := range map[bool]string{false: "positive", true: "negative"} {
			if !signs[negative] {
				return fmt.Errorf(
					"Tier 1 arithmetic width %d exponent-cap contract: no %s coefficient-1 near-cap member at boundary index residue %d; the FMA overflow-clamp companion coverage would silently regress",
					width, label, residue)
			}
		}
	}
	return nil
}
