package testgen

// Decimal32 exhaustive gate routing-sentinel codegen.
//
// The d32 exhaustive runner compares Intel C and the Go port per case, but a
// lane-table glue bug applied identically to both legs (an operation miswire
// that wires both legs of one lane to another operation, or a rounding-mode
// miswire inside the mode-taking lanes) skews the whole differential
// identically: every case still agrees, the counts still close, and only an
// externally pinned expectation can see it. Each sentinel row pins, for one
// hand-declared input, the expected (result bits, raw flags) computed at pin
// time through the public bid754-go API (the shared tier1 sentinel oracle
// subprocess; publicroute proves that surface routes through the Go
// mechanical port). At runtime the runner resolves the row to a lane through
// the same lane table the exhaustive sweep uses and requires
// pinned == Intel C == port, so a pin computed through a broken port
// diverges from live Intel C on the first run (false-fail direction, never
// false-pass). The hand-pinned per-lane result digests in
// devtools/verification_anchors.json are the second, full-space tripwire for
// the same failure class, bound by cmd/verifylog.
//
// Two mode-equivalence identities are mathematical, not harness gaps, and
// are asserted as such below: bid32_sqrt cannot distinguish nearest_even
// from nearest_away (a p+1-digit tie would need its square to be exactly
// representable in p digits, which is impossible), and its nonnegative
// results make toward_zero and toward_negative identical. Mode-pair
// separation across all five modes therefore comes from the
// round_integral_exact rows, whose tie inputs separate every mode pair.
//
// GUARDRAILS: this generator never reads or writes
// devtools/verification_sentinels.json; the human pin flows through
// `cmd/testgen -print-d32-exhaustive-sentinel-anchors` stdout and a manual
// paste audited by TestVerificationAnchorsMatchGeneratedArtifacts.

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

type d32ExhaustiveSentinelRow struct {
	text    string
	comment string
}

// d32ExhaustiveNoncanonicalInput is the steered-encoding finite pattern with
// coefficient 0x800000|0x1FFFFF = 10485759 > 9999999 at biased exponent 101
// (E0): a non-canonical Decimal32 whose canonical reading is zero. It pins
// the treat-as-zero contract through every conversion lane.
const d32ExhaustiveNoncanonicalInput = uint32(0x6cbfffff)

func d32ExhaustiveSentinelEncode(sign bool, coeff int64, exp int32) (uint32, error) {
	if coeff < 0 {
		return 0, fmt.Errorf("d32 exhaustive sentinel operand coefficient %d is negative; use sign", coeff)
	}
	components := bidCodecRefComponents{
		Sign:     sign,
		Exponent: exp,
		Kind:     bidCodecRefNormal,
	}
	if coeff == 0 {
		components.Kind = bidCodecRefZero
	} else {
		components.Coefficient = big.NewInt(coeff)
	}
	return refEncode32(components), nil
}

func d32ExhaustiveSentinelSpecial(kind bidCodecRefKind, sign bool) uint32 {
	return refEncode32(bidCodecRefComponents{Sign: sign, Kind: kind})
}

// d32ExhaustiveSentinelModeVariant maps a native Intel rounding mode onto
// the fixed-attribute round_integral lane exercising the same attribute.
func d32ExhaustiveSentinelModeVariant(native int) (string, error) {
	switch native {
	case 0:
		return "nearest_even", nil
	case 1:
		return "negative", nil
	case 2:
		return "positive", nil
	case 3:
		return "zero", nil
	case 4:
		return "nearest_away", nil
	default:
		return "", fmt.Errorf("d32 exhaustive sentinel: unknown native rounding mode %d", native)
	}
}

// d32ExhaustiveSentinelFixedLaneName is the lane (and row op token) name of
// one fixed-attribute round_integral variant.
func d32ExhaustiveSentinelFixedLaneName(variant string) string {
	return "round_integral_" + variant
}

// d32ExhaustiveSentinelOracle answers one lane evaluation through the
// pin-time public-API oracle; the payload is the exact leg text the runner
// compares at runtime ("bits/rawflags" in lane result width).
func d32ExhaustiveSentinelOracle(opToken string, nativeMode int, x uint32) (string, error) {
	switch opToken {
	case "sqrt":
		return tier1SentinelOracleQuery(fmt.Sprintf("sqrt 32 %d %08x", nativeMode, x))
	case "round_integral_exact":
		return tier1SentinelOracleQuery(fmt.Sprintf("roundintexact 32 %d %08x", nativeMode, x))
	case "to_bid64":
		return tier1SentinelOracleQuery(fmt.Sprintf("widthconv 32 64 0 %08x", x))
	case "to_bid128":
		return tier1SentinelOracleQuery(fmt.Sprintf("widthconv 32 128 0 %08x", x))
	default:
		variant, found := strings.CutPrefix(opToken, "round_integral_")
		if !found {
			return "", fmt.Errorf("d32 exhaustive sentinel: unknown op token %q", opToken)
		}
		return tier1SentinelOracleQuery(fmt.Sprintf("roundint 32 %s %08x", variant, x))
	}
}

func d32ExhaustiveSentinelRowText(opToken string, nativeMode int, x uint32, payload string) string {
	if nativeMode >= 0 {
		return fmt.Sprintf("d32 %s x=%08x m=%d -> %s", opToken, x, nativeMode, payload)
	}
	return fmt.Sprintf("d32 %s x=%08x -> %s", opToken, x, payload)
}

func d32ExhaustiveSentinelPayloadFlags(payload string) (uint32, error) {
	_, flagsText, found := strings.Cut(payload, "/")
	if !found {
		return 0, fmt.Errorf("d32 exhaustive sentinel payload %q is not <bits>/<flags>", payload)
	}
	var flags uint32
	if _, err := fmt.Sscanf(flagsText, "%x", &flags); err != nil {
		return 0, fmt.Errorf("d32 exhaustive sentinel payload flags %q: %w", flagsText, err)
	}
	return flags, nil
}

// GenerateD32ExhaustiveSentinelRows builds the pinned sentinel rows in lane
// order and mechanically verifies lane coverage plus the mode/op
// discrimination requirements. Any failed requirement aborts generation.
func GenerateD32ExhaustiveSentinelRows() ([]d32ExhaustiveSentinelRow, error) {
	lanes := d32ExhaustiveLaneSpecs()
	if err := d32ExhaustiveValidateLanes(lanes); err != nil {
		return nil, err
	}

	tie1, err := d32ExhaustiveSentinelEncode(false, 15, -1) // 1.5: RN/RA tie at integer quantum
	if err != nil {
		return nil, err
	}
	tie2, err := d32ExhaustiveSentinelEncode(false, 25, -1) // 2.5: separates RN from RA
	if err != nil {
		return nil, err
	}
	tie3, err := d32ExhaustiveSentinelEncode(true, 15, -1) // -1.5: separates the directed modes by sign
	if err != nil {
		return nil, err
	}
	tieInputs := []uint32{tie1, tie2, tie3}

	sqrt1, err := d32ExhaustiveSentinelEncode(false, 2, 0) // sqrt(2): inexact, rounds up at RN
	if err != nil {
		return nil, err
	}
	sqrt2, err := d32ExhaustiveSentinelEncode(false, 102, -2) // sqrt(1.02): inexact, truncates at RN
	if err != nil {
		return nil, err
	}
	sqrtInputs := []uint32{sqrt1, sqrt2}
	sqrtNegOne, err := d32ExhaustiveSentinelEncode(true, 1, 0)
	if err != nil {
		return nil, err
	}
	posInf := d32ExhaustiveSentinelSpecial(bidCodecRefInfinity, false)
	sNaN := d32ExhaustiveSentinelSpecial(bidCodecRefSNaN, false)

	conv1, err := d32ExhaustiveSentinelEncode(false, 7654321, -3) // full-precision finite promotion
	if err != nil {
		return nil, err
	}
	if decnumberDiffCanonical(decnumberDiffWidths[0], bid128BidCodecValue{lo: uint64(d32ExhaustiveNoncanonicalInput)}) {
		return nil, fmt.Errorf("d32 exhaustive sentinel: declared non-canonical input %08x round-trips canonically", d32ExhaustiveNoncanonicalInput)
	}
	convInputs := []uint32{conv1, sNaN, d32ExhaustiveNoncanonicalInput}

	var rows []d32ExhaustiveSentinelRow
	coveredLanes := map[string]bool{}
	// Result vectors for the discrimination assertions, keyed by native mode
	// (mode-taking ops) or variant (fixed lanes).
	sqrtVectors := map[int]string{}
	rieVectors := map[int]string{}
	fixedVectors := map[string]string{}
	riePayloadByMode := map[int]string{}
	fixedPayloadByVariant := map[string]string{}

	appendRow := func(lane d32ExhaustiveLaneSpec, x uint32, comment string) (string, error) {
		payload, err := d32ExhaustiveSentinelOracle(lane.opToken, lane.nativeMode, x)
		if err != nil {
			return "", err
		}
		rows = append(rows, d32ExhaustiveSentinelRow{
			text:    d32ExhaustiveSentinelRowText(lane.opToken, lane.nativeMode, x, payload),
			comment: fmt.Sprintf("%s x=%08x: %s", lane.name, x, comment),
		})
		coveredLanes[lane.name] = true
		return payload, nil
	}

	for _, lane := range lanes {
		switch {
		case lane.opToken == "sqrt":
			vector := ""
			for _, x := range sqrtInputs {
				payload, err := appendRow(lane, x, "inexact square root; separates the RN-class/RZ-class/toward_positive attribute classes")
				if err != nil {
					return nil, err
				}
				vector += payload + ";"
			}
			sqrtVectors[lane.nativeMode] = vector
			if lane.nativeMode == 0 {
				if _, err := appendRow(lane, sqrtNegOne, "sqrt(-1) raises invalid and returns qNaN"); err != nil {
					return nil, err
				}
				if _, err := appendRow(lane, posInf, "sqrt(+Inf) is +Inf exactly"); err != nil {
					return nil, err
				}
				payload, err := appendRow(lane, sNaN, "sqrt(sNaN) raises invalid and quiets the NaN")
				if err != nil {
					return nil, err
				}
				flags, err := d32ExhaustiveSentinelPayloadFlags(payload)
				if err != nil {
					return nil, err
				}
				if flags&0x01 == 0 {
					return nil, fmt.Errorf("d32 exhaustive sentinel: sqrt(sNaN) pin %q carries no invalid flag", payload)
				}
			}
		case lane.opToken == "round_integral_exact":
			vector := ""
			for _, x := range tieInputs {
				payload, err := appendRow(lane, x, "half-way tie; the five mode lanes pin pairwise-distinct result vectors")
				if err != nil {
					return nil, err
				}
				vector += payload + ";"
				if x == tie1 {
					riePayloadByMode[lane.nativeMode] = payload
				}
			}
			rieVectors[lane.nativeMode] = vector
		case strings.HasPrefix(lane.opToken, "round_integral_"):
			variant := strings.TrimPrefix(lane.opToken, "round_integral_")
			vector := ""
			for _, x := range tieInputs {
				payload, err := appendRow(lane, x, "half-way tie; the five fixed-attribute lanes pin pairwise-distinct result vectors")
				if err != nil {
					return nil, err
				}
				vector += payload + ";"
				if x == tie1 {
					fixedPayloadByVariant[variant] = payload
				}
			}
			fixedVectors[variant] = vector
		case lane.opToken == "to_bid64" || lane.opToken == "to_bid128":
			for _, x := range convInputs {
				payload, err := appendRow(lane, x, "exact width promotion: finite full-precision, sNaN quieting with invalid, non-canonical treat-as-zero")
				if err != nil {
					return nil, err
				}
				if x == sNaN {
					flags, err := d32ExhaustiveSentinelPayloadFlags(payload)
					if err != nil {
						return nil, err
					}
					if flags&0x01 == 0 {
						return nil, fmt.Errorf("d32 exhaustive sentinel: %s(sNaN) pin %q carries no invalid flag", lane.opToken, payload)
					}
				}
			}
		default:
			return nil, fmt.Errorf("d32 exhaustive sentinel: lane %q has no candidate rows", lane.name)
		}
	}

	for _, lane := range lanes {
		if !coveredLanes[lane.name] {
			return nil, fmt.Errorf("d32 exhaustive sentinel: lane %q is not covered by any sentinel row", lane.name)
		}
	}

	// round_integral_exact: every mode pair must be separated.
	for a := 0; a <= 4; a++ {
		for b := a + 1; b <= 4; b++ {
			if rieVectors[a] == rieVectors[b] {
				return nil, fmt.Errorf("d32 exhaustive sentinel: round_integral_exact modes %d and %d pin identical result vectors %q", a, b, rieVectors[a])
			}
		}
	}
	// fixed variants: every lane pair must be separated.
	variants := []string{"nearest_even", "nearest_away", "zero", "positive", "negative"}
	for i, a := range variants {
		for _, b := range variants[i+1:] {
			if fixedVectors[a] == fixedVectors[b] {
				return nil, fmt.Errorf("d32 exhaustive sentinel: fixed round_integral variants %s and %s pin identical result vectors %q", a, b, fixedVectors[a])
			}
		}
	}
	// sqrt: assert the three reachable attribute classes and the two
	// mathematical identities (see the header comment).
	if sqrtVectors[0] != sqrtVectors[4] {
		return nil, fmt.Errorf("d32 exhaustive sentinel: sqrt RN/RA identity broken (%q vs %q); the tie-impossibility argument no longer holds", sqrtVectors[0], sqrtVectors[4])
	}
	if sqrtVectors[3] != sqrtVectors[1] {
		return nil, fmt.Errorf("d32 exhaustive sentinel: sqrt RZ/toward_negative identity broken (%q vs %q)", sqrtVectors[3], sqrtVectors[1])
	}
	if sqrtVectors[0] == sqrtVectors[3] || sqrtVectors[0] == sqrtVectors[2] || sqrtVectors[2] == sqrtVectors[3] {
		return nil, fmt.Errorf("d32 exhaustive sentinel: sqrt attribute classes are not separated (RN-class %q, RZ-class %q, toward_positive %q)", sqrtVectors[0], sqrtVectors[3], sqrtVectors[2])
	}
	// exact-vs-fixed: at every mode the exact variant's inexact flag must
	// separate the round_integral_exact lane from its fixed twin on the
	// same input.
	for native := 0; native <= 4; native++ {
		variant, err := d32ExhaustiveSentinelModeVariant(native)
		if err != nil {
			return nil, err
		}
		if riePayloadByMode[native] == fixedPayloadByVariant[variant] {
			return nil, fmt.Errorf("d32 exhaustive sentinel: round_integral_exact mode %d and fixed variant %s pin the identical payload %q on the tie input", native, variant, riePayloadByMode[native])
		}
	}

	if err := tier1SentinelOracleErr(); err != nil {
		return nil, err
	}
	return rows, nil
}

func d32ExhaustiveSentinelGoRowLiterals(rows []d32ExhaustiveSentinelRow) string {
	var out strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&out, "\t%q,\n", row.text)
	}
	return out.String()
}

// D32ExhaustiveSentinelAnchorProposal prints the proposed
// verification_sentinels.json rows for the human pin (stdout only; no file
// is read or written).
func D32ExhaustiveSentinelAnchorProposal() (string, error) {
	rows, err := GenerateD32ExhaustiveSentinelRows()
	if err != nil {
		return "", err
	}
	texts := make([]string, 0, len(rows))
	for _, row := range rows {
		texts = append(texts, row.text)
	}
	proposal := struct {
		Rows []string `json:"d32_exhaustive_sentinel_rows"`
	}{Rows: texts}
	var encoded strings.Builder
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false) // keep the "->" row arrow literal for the human audit
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(proposal); err != nil {
		return "", err
	}
	var out strings.Builder
	out.WriteString("Proposed devtools/verification_sentinels.json rows (audit each row, then paste manually — no generator writes that file):\n")
	out.WriteString(encoded.String())
	out.WriteString("\nRow interpretations (audit aid, not pinned data):\n")
	for _, row := range rows {
		fmt.Fprintf(&out, "# %s\n", row.comment)
	}
	return out.String(), nil
}
