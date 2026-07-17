package testgen

// decNumber differential routing-sentinel codegen.
//
// The decNumber differential runner compares three legs per case, but a glue
// bug applied identically to every leg's dispatch (operand slot swap,
// rounding-mode miswire — the public Go enum order differs from the Intel
// raw 0..4 numbering — or width/context mislabel) can skew the whole
// comparison identically. Each sentinel row pins, for one hand-declared
// operand tuple, the expected Intel (result bits, raw flags) computed at pin
// time through the public bid754-go API (the shared tier1 sentinel oracle
// subprocess; publicroute proves that surface routes through the Go
// mechanical port) plus the expected decNumber comparison triple and
// projected 5-flag set derived from those bits through the independent
// reference layout. At runtime the runner requires pinned == Intel C ==
// port and pinned triple == decNumber leg for every row, so a pin computed
// through a broken port diverges from live Intel C on the first full run
// (false-fail direction, never false-pass).
//
// Every candidate tuple lies inside the gate's exact agreement region:
// canonical operands, NaN payload zero, no L13-class fma tuple, sqrt at
// round-nearest-even only, no tiny/denormal results (the pinned raw flags
// must stay inside the 5-flag surface). The mechanical selection assertions
// below (slot/mode/width sensitivity) fail generation, not runtime.
//
// GUARDRAILS: this generator never reads or writes
// devtools/verification_sentinels.json; the human pin flows through
// `cmd/testgen -print-decnumber-sentinel-anchors` stdout and a manual paste
// audited by TestVerificationAnchorsMatchGeneratedArtifacts.

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

type decnumberDiffSentinelRow struct {
	text    string
	comment string
}

// decnumberDiffSentinelCandidate declares one hand-picked operand tuple.
// Operands are decimal triples (coefficient scaled to the width's precision
// where needed through the coeffScale hook), so one declaration covers all
// three widths with width-appropriate rounding behavior.
type decnumberDiffSentinelCandidate struct {
	op string
	// build returns x, y, z (z used by fma only, sqrt uses x only).
	build func(width decnumberDiffWidth) (x, y, z decnumberDiffTriple)
	// modes lists the native Intel rounding modes pinned for the tuple.
	modes []int
	// wantSlotSensitive requires the oracle result to change when the two
	// binary operand slots (or the fma x/z slots) are swapped at pin time.
	wantSlotSensitive bool
	// wantModeSensitive requires at least two distinct oracle results across
	// the pinned modes.
	wantModeSensitive bool
	// wantWidthDistinct requires the comparison triple of the first pinned
	// mode to differ across all three widths (context digits wiring).
	wantWidthDistinct bool
	comment           string
}

var decnumberDiffSentinelModesAll = []int{0, 1, 2, 3, 4}

func decnumberDiffSentinelFinite(sign bool, coeff int64, exp int32) decnumberDiffTriple {
	return decnumberDiffTriple{kind: decnumberDiffFinite, sign: sign, coeff: big.NewInt(coeff), exp: exp}
}

// decnumberDiffSentinelCandidates is the fixed declaration-order candidate
// table. Every candidate is adopted for every declared mode; a sensitivity
// requirement that fails aborts generation with no partial output.
func decnumberDiffSentinelCandidates() []decnumberDiffSentinelCandidate {
	return []decnumberDiffSentinelCandidate{
		{
			op: "add",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// 1 + 5*10^-p: a p+1 digit tie — distinct results across RN/RA vs RD/RZ.
				return decnumberDiffSentinelFinite(false, 1, 0),
					decnumberDiffSentinelFinite(false, 5, -int32(w.p)),
					decnumberDiffTriple{}
			},
			modes:             decnumberDiffSentinelModesAll,
			wantModeSensitive: true,
			comment:           "add rounding tie 1 + 5E-p: separates half-even/half-up from directed modes",
		},
		{
			op: "add",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// x + (-x) is +0 in every mode except roundTowardNegative (-0):
				// pins the IEEE zero-sign rule and the dn zero-sign extraction.
				return decnumberDiffSentinelFinite(false, 5, 0),
					decnumberDiffSentinelFinite(true, 5, 0),
					decnumberDiffTriple{}
			},
			modes:             decnumberDiffSentinelModesAll,
			wantModeSensitive: true,
			comment:           "add exact-cancellation zero sign: only toward_negative yields -0",
		},
		{
			op: "add",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// Nmax + Nmax overflows: RN/RA/RU -> +Inf, RZ/RD -> +Nmax (L14 region).
				nmax := decnumberDiffTriple{kind: decnumberDiffFinite, coeff: w.maxCoeff, exp: w.maxExp}
				return nmax, nmax, decnumberDiffTriple{}
			},
			modes:             decnumberDiffSentinelModesAll,
			wantModeSensitive: true,
			comment:           "add overflow at +Nmax: directed modes clamp to Nmax, others to +Inf",
		},
		{
			op: "sub",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// 5E0 - 3E0 = 2E0 vs swapped -2E0: slot-asymmetric.
				return decnumberDiffSentinelFinite(false, 5, 0),
					decnumberDiffSentinelFinite(false, 3, 0),
					decnumberDiffTriple{}
			},
			modes:             []int{0},
			wantSlotSensitive: true,
			comment:           "sub slot asymmetry 5-3 vs 3-5",
		},
		{
			op: "mul",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// (10^(p-1)+1) * 11E-1: p+1 significant digits force one rounded digit.
				coeff := new(big.Int).Add(decnumberDiffPow10(w.p-1), big.NewInt(1))
				return decnumberDiffTriple{kind: decnumberDiffFinite, coeff: coeff, exp: 0},
					decnumberDiffSentinelFinite(false, 11, -1),
					decnumberDiffTriple{}
			},
			modes:             decnumberDiffSentinelModesAll,
			wantModeSensitive: true,
			comment:           "mul inexact p+1-digit product: rounded last digit differs by mode",
		},
		{
			op: "div",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// 1/3: infinite expansion, slot-asymmetric, and the comparison
				// triple coefficient length equals the context precision, so a
				// width/context digits miswire on any leg changes the pinned triple.
				return decnumberDiffSentinelFinite(false, 1, 0),
					decnumberDiffSentinelFinite(false, 3, 0),
					decnumberDiffTriple{}
			},
			modes:             decnumberDiffSentinelModesAll,
			wantSlotSensitive: true,
			wantModeSensitive: true,
			wantWidthDistinct: true,
			comment:           "div 1/3: mode-, slot-, and context-digits-sensitive",
		},
		{
			op: "div",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// -1/3: directed-mode sign asymmetry (floor vs ceiling swap roles).
				return decnumberDiffSentinelFinite(true, 1, 0),
					decnumberDiffSentinelFinite(false, 3, 0),
					decnumberDiffTriple{}
			},
			modes:             decnumberDiffSentinelModesAll,
			wantModeSensitive: true,
			comment:           "div -1/3: floor/ceiling role swap on the negative axis",
		},
		{
			op: "quantize",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// quantize(15E-1, 1E0): tie at integer quantum; slot-asymmetric.
				return decnumberDiffSentinelFinite(false, 15, -1),
					decnumberDiffSentinelFinite(false, 1, 0),
					decnumberDiffTriple{}
			},
			modes:             decnumberDiffSentinelModesAll,
			wantSlotSensitive: true,
			wantModeSensitive: true,
			comment:           "quantize 1.5 to quantum E0: tie behavior by mode, slot-asymmetric",
		},
		{
			op: "fma",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// fma(2,3,7)=13 vs fma(7,3,2)=23: x/z slot swap changes the value.
				return decnumberDiffSentinelFinite(false, 2, 0),
					decnumberDiffSentinelFinite(false, 3, 0),
					decnumberDiffSentinelFinite(false, 7, 0)
			},
			modes:             []int{0},
			wantSlotSensitive: true,
			comment:           "fma x/z slot asymmetry 2*3+7 vs 7*3+2",
		},
		{
			op: "fma",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// fma(1,1,5E-p): the add tie reached through the fused path.
				return decnumberDiffSentinelFinite(false, 1, 0),
					decnumberDiffSentinelFinite(false, 1, 0),
					decnumberDiffSentinelFinite(false, 5, -int32(w.p))
			},
			modes:             decnumberDiffSentinelModesAll,
			wantModeSensitive: true,
			comment:           "fma fused tie 1*1 + 5E-p: mode separation through the fused rounding",
		},
		{
			op: "sqrt",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// sqrt(2): inexact p-digit result; RN only (L11 exclusion keeps
				// non-RN sqrt out of the gate, so the sentinel stays inside it).
				return decnumberDiffSentinelFinite(false, 2, 0), decnumberDiffTriple{}, decnumberDiffTriple{}
			},
			modes:             []int{0},
			wantWidthDistinct: true,
			comment:           "sqrt(2) at RN: context-digits-sensitive inexact result",
		},
		{
			op: "sqrt",
			build: func(w decnumberDiffWidth) (decnumberDiffTriple, decnumberDiffTriple, decnumberDiffTriple) {
				// sqrt(4) = 2 exactly: exact path with preferred exponent 0.
				return decnumberDiffSentinelFinite(false, 4, 0), decnumberDiffTriple{}, decnumberDiffTriple{}
			},
			modes:   []int{0},
			comment: "sqrt(4) at RN: exact result, no flags",
		},
	}
}

// decnumberDiffSentinelOracle evaluates one tuple through the pin-time
// oracle and returns "bits/rawflags" in the width's canonical text.
func decnumberDiffSentinelOracle(width decnumberDiffWidth, op string, mode int, x, y, z bid128BidCodecValue) (string, error) {
	xText := tier1SentinelValueText(width.bits, x)
	switch op {
	case "add", "sub", "mul", "div", "quantize":
		return tier1SentinelOracleQuery(fmt.Sprintf("rounded %d %s %d %s %s", width.bits, op, mode, xText, tier1SentinelValueText(width.bits, y)))
	case "fma":
		return tier1SentinelOracleQuery(fmt.Sprintf("fma %d %d %s %s %s", width.bits, mode, xText, tier1SentinelValueText(width.bits, y), tier1SentinelValueText(width.bits, z)))
	case "sqrt":
		return tier1SentinelOracleQuery(fmt.Sprintf("sqrt %d %d %s", width.bits, mode, xText))
	default:
		return "", fmt.Errorf("decnumber differential sentinel: unknown operation %q", op)
	}
}

// decnumberDiffSentinelParseResult splits an oracle "bits/rawflags" payload
// into the raw bit words and the raw flag word.
func decnumberDiffSentinelParseResult(width decnumberDiffWidth, payload string) (bid128BidCodecValue, uint32, error) {
	bitsText, flagsText, ok := strings.Cut(payload, "/")
	if !ok {
		return bid128BidCodecValue{}, 0, fmt.Errorf("oracle payload %q is not <bits>/<flags>", payload)
	}
	var value bid128BidCodecValue
	switch width.bits {
	case 32, 64:
		var lo uint64
		if _, err := fmt.Sscanf(bitsText, "%x", &lo); err != nil {
			return value, 0, fmt.Errorf("oracle bits %q: %w", bitsText, err)
		}
		value.lo = lo
	default:
		hiText, loText, ok := strings.Cut(bitsText, ":")
		if !ok {
			return value, 0, fmt.Errorf("oracle 128-bit payload %q is not <hi>:<lo>", bitsText)
		}
		if _, err := fmt.Sscanf(hiText, "%x", &value.hi); err != nil {
			return value, 0, fmt.Errorf("oracle hi bits %q: %w", hiText, err)
		}
		if _, err := fmt.Sscanf(loText, "%x", &value.lo); err != nil {
			return value, 0, fmt.Errorf("oracle lo bits %q: %w", loText, err)
		}
	}
	var flags uint32
	if _, err := fmt.Sscanf(flagsText, "%x", &flags); err != nil {
		return value, 0, fmt.Errorf("oracle flags %q: %w", flagsText, err)
	}
	return value, flags, nil
}

// decnumberDiffTripleKey renders the pinned decNumber comparison triple in
// the same text form the runner renders at runtime (NaN payload and sign
// projected out per the v1 comparison surface).
func decnumberDiffTripleKey(t decnumberDiffTriple) string {
	sign := ""
	if t.sign {
		sign = "-"
	}
	switch t.kind {
	case decnumberDiffInf:
		return sign + "Inf"
	case decnumberDiffQNaN:
		return "qNaN"
	case decnumberDiffSNaN:
		return "sNaN"
	default:
		return sign + t.coeff.String() + "E" + fmt.Sprint(t.exp)
	}
}

const decnumberDiffFlags5Mask = uint32(0x3d) // invalid|divzero|overflow|underflow|inexact

// decnumberDiffSentinelLegText renders the pinned expected leg exactly as
// the generated FFI shim renders its runtime leg strings ("%08x"/"%016x"
// words for 32/64, the little-endian 16-byte hex image for 128), so the
// runner can compare the live Intel C and port legs byte-for-byte against
// the pin.
func decnumberDiffSentinelLegText(width decnumberDiffWidth, bits bid128BidCodecValue, rawFlags uint32) string {
	switch width.bits {
	case 32:
		return fmt.Sprintf("%08x/%08x", uint32(bits.lo), rawFlags)
	case 64:
		return fmt.Sprintf("%016x/%08x", bits.lo, rawFlags)
	default:
		var raw [16]byte
		for i := 0; i < 8; i++ {
			raw[i] = byte(bits.lo >> (8 * i))
			raw[8+i] = byte(bits.hi >> (8 * i))
		}
		return fmt.Sprintf("%x/%08x", raw[:], rawFlags)
	}
}

func decnumberDiffSentinelRowText(width decnumberDiffWidth, op string, mode int, x, y, z bid128BidCodecValue, arity int, bits bid128BidCodecValue, rawFlags uint32) (string, error) {
	if rawFlags&^decnumberDiffFlags5Mask != 0 {
		return "", fmt.Errorf("decnumber differential sentinel %s width %d: pinned raw flags %08x leave the 5-flag surface", op, width.bits, rawFlags)
	}
	expected := decnumberDiffDecode(width, bits)
	if !decnumberDiffCanonical(width, bits) {
		return "", fmt.Errorf("decnumber differential sentinel %s width %d: pinned result bits are non-canonical", op, width.bits)
	}
	operands := "x=" + tier1SentinelValueText(width.bits, x)
	if arity >= 2 {
		operands += " y=" + tier1SentinelValueText(width.bits, y)
	}
	if arity >= 3 {
		operands += " z=" + tier1SentinelValueText(width.bits, z)
	}
	return fmt.Sprintf("d%d %s %s m=%d -> %s dn=%s/%08x",
		width.bits, op, operands, mode,
		decnumberDiffSentinelLegText(width, bits, rawFlags),
		decnumberDiffTripleKey(expected), rawFlags&decnumberDiffFlags5Mask), nil
}

func decnumberDiffSentinelArity(op string) int {
	switch op {
	case "fma":
		return 3
	case "sqrt":
		return 1
	default:
		return 2
	}
}

// GenerateDecnumberDifferentialSentinelRows builds the pinned sentinel rows
// (width ascending, candidate declaration order, then declared mode order)
// and mechanically verifies the slot/mode/width sensitivity requirements.
func GenerateDecnumberDifferentialSentinelRows() ([]decnumberDiffSentinelRow, error) {
	candidates := decnumberDiffSentinelCandidates()
	var rows []decnumberDiffSentinelRow
	// widthFirstTriples[candidateIndex][widthIndex] holds the mode-0-pinned
	// comparison triple for the width-distinct assertion.
	widthFirstTriples := make(map[int]map[int]string)
	for widthIndex, width := range decnumberDiffWidths {
		for candidateIndex, candidate := range candidates {
			xt, yt, zt := candidate.build(width)
			arity := decnumberDiffSentinelArity(candidate.op)
			x, err := decnumberDiffEncode(width, xt)
			if err != nil {
				return nil, err
			}
			y := bid128BidCodecValue{}
			z := bid128BidCodecValue{}
			if arity >= 2 {
				if y, err = decnumberDiffEncode(width, yt); err != nil {
					return nil, err
				}
			}
			if arity >= 3 {
				if z, err = decnumberDiffEncode(width, zt); err != nil {
					return nil, err
				}
				if decnumberDiffFmaExcluded(xt, yt, zt) {
					return nil, fmt.Errorf("decnumber differential sentinel candidate %d pins an L13-class fma tuple", candidateIndex)
				}
			}
			var modeResults []string
			for _, mode := range candidate.modes {
				payload, err := decnumberDiffSentinelOracle(width, candidate.op, mode, x, y, z)
				if err != nil {
					return nil, err
				}
				bits, rawFlags, err := decnumberDiffSentinelParseResult(width, payload)
				if err != nil {
					return nil, err
				}
				text, err := decnumberDiffSentinelRowText(width, candidate.op, mode, x, y, z, arity, bits, rawFlags)
				if err != nil {
					return nil, err
				}
				rows = append(rows, decnumberDiffSentinelRow{
					text:    text,
					comment: fmt.Sprintf("d%d %s m=%d: %s", width.bits, candidate.op, mode, candidate.comment),
				})
				modeResults = append(modeResults, payload)
				if mode == candidate.modes[0] {
					if widthFirstTriples[candidateIndex] == nil {
						widthFirstTriples[candidateIndex] = map[int]string{}
					}
					widthFirstTriples[candidateIndex][widthIndex] = decnumberDiffTripleKey(decnumberDiffDecode(width, bits))
				}
			}
			if candidate.wantModeSensitive {
				distinct := map[string]bool{}
				for _, result := range modeResults {
					distinct[result] = true
				}
				if len(distinct) < 2 {
					return nil, fmt.Errorf("decnumber differential sentinel candidate %d (%s width %d) is not mode-sensitive: all %d modes agree on %q",
						candidateIndex, candidate.op, width.bits, len(candidate.modes), modeResults[0])
				}
			}
			if candidate.wantSlotSensitive {
				swappedX, swappedY, swappedZ := x, y, z
				switch arity {
				case 2:
					swappedX, swappedY = y, x
				case 3:
					swappedX, swappedZ = z, x
				default:
					return nil, fmt.Errorf("decnumber differential sentinel candidate %d: slot sensitivity is undefined for arity %d", candidateIndex, arity)
				}
				original, err := decnumberDiffSentinelOracle(width, candidate.op, candidate.modes[0], x, y, z)
				if err != nil {
					return nil, err
				}
				swapped, err := decnumberDiffSentinelOracle(width, candidate.op, candidate.modes[0], swappedX, swappedY, swappedZ)
				if err != nil {
					return nil, err
				}
				if original == swapped {
					return nil, fmt.Errorf("decnumber differential sentinel candidate %d (%s width %d) is not slot-sensitive: swapped operands agree on %q",
						candidateIndex, candidate.op, width.bits, original)
				}
			}
		}
	}
	for candidateIndex, candidate := range candidates {
		if !candidate.wantWidthDistinct {
			continue
		}
		triples := widthFirstTriples[candidateIndex]
		seen := map[string]int{}
		for widthIndex, key := range triples {
			if otherIndex, dup := seen[key]; dup {
				return nil, fmt.Errorf("decnumber differential sentinel candidate %d (%s) is not width-distinct: widths %d and %d both pin %q",
					candidateIndex, candidate.op, decnumberDiffWidths[otherIndex].bits, decnumberDiffWidths[widthIndex].bits, key)
			}
			seen[key] = widthIndex
		}
	}
	if err := tier1SentinelOracleErr(); err != nil {
		return nil, err
	}
	return rows, nil
}

func decnumberDiffSentinelGoRowLiterals(rows []decnumberDiffSentinelRow) string {
	var out strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&out, "\t%q,\n", row.text)
	}
	return out.String()
}

// DecnumberDifferentialSentinelAnchorProposal prints the proposed
// verification_sentinels.json rows for the human pin (stdout only; no file
// is read or written).
func DecnumberDifferentialSentinelAnchorProposal() (string, error) {
	rows, err := GenerateDecnumberDifferentialSentinelRows()
	if err != nil {
		return "", err
	}
	texts := make([]string, 0, len(rows))
	for _, row := range rows {
		texts = append(texts, row.text)
	}
	proposal := struct {
		Rows []string `json:"decnumber_differential_sentinel_rows"`
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
