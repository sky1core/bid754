package testgen

// Mixed-format FFI operand-swap routing-sentinel codegen.
//
// The mixed-format FFI differential compares pinned Intel BID C against the Go
// mechanical port for the 24 mixed/equal-width arithmetic functions, reading
// both legs from the SAME operand slots of one case. A runner-glue bug that
// swaps the two operand slots (x<->y) identically for both legs skews both
// legs the same way: on a non-commutative operation the differential still
// finds C == port (an agreed wrong answer) while the real answer is op(x,y),
// not op(y,x). That common-mode gap is reachable only where a swap keeps the
// case shape, i.e. the two EQUAL-width families whose slots share a width —
// bid64qq_* (Decimal128 x Decimal128 -> Decimal64) and bid128dd_*
// (Decimal64 x Decimal64 -> Decimal128) — and only for the non-commutative
// sub/div. That is the closed CORE set of four functions; commutative add/mul
// cannot witness a swap (op(x,y) == op(y,x)) and cross-width dq/qd swaps change
// the width, which the shape table and differential already reject.
//
// Each sentinel row pins, for one hand-declared (function, mode, x, y) with
// x != y, the expected Intel (result bits, raw flags) computed at pin time
// through the PUBLIC mixed bid754-go API (Sub64QQBIDWithMode / Div64QQ... /
// Sub128DD... / Div128DD...; the publicroute gate proves that surface routes
// through the Go mechanical port). The value flows through the shared tier1
// sentinel oracle subprocess — a program and operand path separate from the
// FFI runner glue — so a runner slot swap cannot transfer into the pin, and
// the pin is frozen once. The generated runner replays each row through the
// SAME runGeneratedFFICase dispatch the differential uses, requiring pinned ==
// Intel C == port; a swap in that shared dispatch makes both legs compute
// op(y,x) and diverge from the frozen op(x,y) pin (false-fail direction, never
// false-pass).
//
// Every candidate is mechanically checked at generation time to be
// slot-sensitive (op(x,y) != op(y,x)); div candidates additionally to be
// mode-sensitive. A failed requirement aborts generation with no partial
// output.
//
// GUARDRAILS: this generator never reads or writes
// devtools/verification_sentinels.json; the human pin flows through
// `cmd/testgen -print-mixed-ffi-sentinel-anchors` stdout and a manual paste
// audited by TestVerificationAnchorsMatchGeneratedArtifacts.

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

type mixedFFIRoutingSentinelRow struct {
	text    string
	comment string
	// Audit-only fields (not emitted to the byte-equal Go literal); the anchor
	// proposal uses them to attach decimal interpretations of x, y, the pinned
	// result, and the swapped result so each row can be hand-audited.
	function    string
	operandBits int
	resultBits  int
	xText       string
	yText       string
	mode        int
	payload     string
}

// mixedFFIRoutingSentinelOperand is one hand-declared finite, non-special
// operand as decimal (sign, coeff, exp); the width comes from the candidate's
// function shape. Finite/non-special keeps the pinned decimal interpretation
// stable (NaN payloads on Decimal128 are unstable — see ffi_spec.go).
type mixedFFIRoutingSentinelOperand struct {
	sign  bool
	coeff int64
	exp   int32
}

type mixedFFIRoutingSentinelCandidate struct {
	function string
	x, y     mixedFFIRoutingSentinelOperand
	modes    []int
	// wantModeSensitive requires >= 2 distinct oracle results across the pinned
	// modes (a live rounding-mode discriminant, so a mode-slot miswire on the
	// shared dispatch also breaks the row).
	wantModeSensitive bool
	comment           string
}

func mixedFFIRoutingSentinelFinite(sign bool, coeff int64, exp int32) mixedFFIRoutingSentinelOperand {
	return mixedFFIRoutingSentinelOperand{sign: sign, coeff: coeff, exp: exp}
}

// mixedFFIRoutingSentinelOperandBits returns the operand width for one CORE
// function. The result width is the complementary equal-width family member.
func mixedFFIRoutingSentinelOperandBits(function string) (int, error) {
	switch function {
	case "bid64qq_sub", "bid64qq_div":
		return 128, nil
	case "bid128dd_sub", "bid128dd_div":
		return 64, nil
	default:
		return 0, fmt.Errorf("mixed FFI routing sentinel: unsupported function %q", function)
	}
}

func mixedFFIRoutingSentinelResultBits(function string) (int, error) {
	switch function {
	case "bid64qq_sub", "bid64qq_div":
		return 64, nil
	case "bid128dd_sub", "bid128dd_div":
		return 128, nil
	default:
		return 0, fmt.Errorf("mixed FFI routing sentinel: unsupported function %q", function)
	}
}

// mixedFFIRoutingSentinelOperandText renders one operand in the width's
// canonical oracle/row text (64-bit "%016x", 128-bit "hi:lo"), matching the
// oracle's parseValue and the human-auditable BID word order used by the
// existing fusedness/tier1 rows.
func mixedFFIRoutingSentinelOperandText(operandBits int, op mixedFFIRoutingSentinelOperand) (string, error) {
	if op.coeff < 0 {
		return "", fmt.Errorf("mixed FFI routing sentinel operand coefficient %d is negative; use sign", op.coeff)
	}
	components := bidCodecRefComponents{Sign: op.sign, Exponent: op.exp, Kind: bidCodecRefNormal}
	if op.coeff == 0 {
		components.Kind = bidCodecRefZero
	} else {
		components.Coefficient = big.NewInt(op.coeff)
	}
	switch operandBits {
	case 64:
		return fmt.Sprintf("%016x", refEncode64(components)), nil
	case 128:
		lo, hi := refEncode128(components)
		return fmt.Sprintf("%016x:%016x", hi, lo), nil
	default:
		return "", fmt.Errorf("mixed FFI routing sentinel: unsupported operand width %d", operandBits)
	}
}

// mixedFFIRoutingSentinelOracle evaluates one (function, mode, x, y) tuple
// through the pin-time oracle and returns "bits/rawflags" in the result
// width's canonical text (the exact leg text the runner reproduces).
func mixedFFIRoutingSentinelOracle(function string, mode int, xText, yText string) (string, error) {
	return tier1SentinelOracleQuery(fmt.Sprintf("mixed %s %d %s %s", function, mode, xText, yText))
}

func mixedFFIRoutingSentinelRowText(function, xText, yText string, mode int, payload string) string {
	return fmt.Sprintf("%s x=%s y=%s m=%d -> %s", function, xText, yText, mode, payload)
}

// mixedFFIRoutingSentinelCandidates is the fixed declaration-order candidate
// table. Every operand pair is asymmetric (x != y and op(x,y) != op(y,x)); at
// least one pair per family is sign-divergent so the swap is witnessed on the
// sign bit, and every div candidate is inexact so the pinned modes separate.
func mixedFFIRoutingSentinelCandidates() []mixedFFIRoutingSentinelCandidate {
	fin := mixedFFIRoutingSentinelFinite
	return []mixedFFIRoutingSentinelCandidate{
		// bid64qq_sub: Decimal128 x Decimal128 -> Decimal64 (non-commutative).
		{
			function: "bid64qq_sub",
			x:        fin(false, 7, 0),
			y:        fin(false, 3, 0),
			modes:    []int{0, 1},
			comment:  "7 - 3 = 4 vs swapped 3 - 7 = -4: sign-divergent slot swap",
		},
		{
			function:          "bid64qq_sub",
			x:                 fin(false, 1, 0),
			y:                 fin(false, 1, -20),
			modes:             []int{0, 1, 3},
			wantModeSensitive: true,
			comment:           "1 - 1E-20 rounds to Decimal64 at 16 digits (mode-sensitive); swap flips sign",
		},
		// bid64qq_div: Decimal128 x Decimal128 -> Decimal64 (non-commutative).
		{
			function:          "bid64qq_div",
			x:                 fin(false, 1, 0),
			y:                 fin(false, 3, 0),
			modes:             []int{0, 2, 3},
			wantModeSensitive: true,
			comment:           "1 / 3 = 0.333... (inexact, mode-sensitive) vs swapped 3 / 1 = 3 (exact)",
		},
		{
			function:          "bid64qq_div",
			x:                 fin(false, 2, 0),
			y:                 fin(false, 7, 0),
			modes:             []int{0, 2, 3},
			wantModeSensitive: true,
			comment:           "2 / 7 = 0.2857... (inexact) vs swapped 7 / 2 = 3.5 (exact): slot-divergent",
		},
		// bid128dd_sub: Decimal64 x Decimal64 -> Decimal128 (non-commutative,
		// exact in-range so a swap is witnessed by the result sign).
		{
			function: "bid128dd_sub",
			x:        fin(false, 7, 0),
			y:        fin(false, 3, 0),
			modes:    []int{0, 1},
			comment:  "7 - 3 = 4 vs swapped 3 - 7 = -4: sign-divergent slot swap",
		},
		{
			function: "bid128dd_sub",
			x:        fin(false, 3, 0),
			y:        fin(false, 8, 0),
			modes:    []int{0, 2},
			comment:  "3 - 8 = -5 vs swapped 8 - 3 = 5: sign-divergent slot swap",
		},
		// bid128dd_div: Decimal64 x Decimal64 -> Decimal128 (non-commutative).
		{
			function:          "bid128dd_div",
			x:                 fin(false, 1, 0),
			y:                 fin(false, 3, 0),
			modes:             []int{0, 2, 3},
			wantModeSensitive: true,
			comment:           "1 / 3 = 0.333... (inexact, mode-sensitive) vs swapped 3 / 1 = 3 (exact)",
		},
		{
			function:          "bid128dd_div",
			x:                 fin(false, 2, 0),
			y:                 fin(false, 7, 0),
			modes:             []int{0, 2, 3},
			wantModeSensitive: true,
			comment:           "2 / 7 = 0.2857... (inexact) vs swapped 7 / 2 = 3.5 (exact): slot-divergent",
		},
	}
}

// GenerateMixedFFIRoutingSentinelRows builds the pinned sentinel rows in
// candidate declaration order then declared mode order, and mechanically
// verifies the slot-sensitivity (all) and mode-sensitivity (declared)
// requirements. Any failed requirement aborts generation with no partial
// output.
func GenerateMixedFFIRoutingSentinelRows() ([]mixedFFIRoutingSentinelRow, error) {
	candidates := mixedFFIRoutingSentinelCandidates()
	var rows []mixedFFIRoutingSentinelRow
	for candidateIndex, candidate := range candidates {
		operandBits, err := mixedFFIRoutingSentinelOperandBits(candidate.function)
		if err != nil {
			return nil, err
		}
		resultBits, err := mixedFFIRoutingSentinelResultBits(candidate.function)
		if err != nil {
			return nil, err
		}
		xText, err := mixedFFIRoutingSentinelOperandText(operandBits, candidate.x)
		if err != nil {
			return nil, err
		}
		yText, err := mixedFFIRoutingSentinelOperandText(operandBits, candidate.y)
		if err != nil {
			return nil, err
		}
		if xText == yText {
			return nil, fmt.Errorf("mixed FFI routing sentinel candidate %d (%s): operands are equal (%q); a swap could not change the result", candidateIndex, candidate.function, xText)
		}
		if len(candidate.modes) == 0 {
			return nil, fmt.Errorf("mixed FFI routing sentinel candidate %d (%s): no modes declared", candidateIndex, candidate.function)
		}

		// Slot sensitivity: op(x,y) must differ from op(y,x) at the first mode,
		// so an operand-slot swap in the shared dispatch is detectable.
		original, err := mixedFFIRoutingSentinelOracle(candidate.function, candidate.modes[0], xText, yText)
		if err != nil {
			return nil, err
		}
		swapped, err := mixedFFIRoutingSentinelOracle(candidate.function, candidate.modes[0], yText, xText)
		if err != nil {
			return nil, err
		}
		if original == swapped {
			return nil, fmt.Errorf("mixed FFI routing sentinel candidate %d (%s) is not slot-sensitive: op(x,y) and op(y,x) both = %q", candidateIndex, candidate.function, original)
		}

		var modeResults []string
		for _, mode := range candidate.modes {
			payload, err := mixedFFIRoutingSentinelOracle(candidate.function, mode, xText, yText)
			if err != nil {
				return nil, err
			}
			rows = append(rows, mixedFFIRoutingSentinelRow{
				text:        mixedFFIRoutingSentinelRowText(candidate.function, xText, yText, mode, payload),
				comment:     fmt.Sprintf("%s m=%d: %s", candidate.function, mode, candidate.comment),
				function:    candidate.function,
				operandBits: operandBits,
				resultBits:  resultBits,
				xText:       xText,
				yText:       yText,
				mode:        mode,
				payload:     payload,
			})
			modeResults = append(modeResults, payload)
		}

		if candidate.wantModeSensitive {
			distinct := map[string]bool{}
			for _, result := range modeResults {
				distinct[result] = true
			}
			if len(distinct) < 2 {
				return nil, fmt.Errorf("mixed FFI routing sentinel candidate %d (%s) is not mode-sensitive: all %d modes agree on %q", candidateIndex, candidate.function, len(candidate.modes), modeResults[0])
			}
		}
	}
	if err := tier1SentinelOracleErr(); err != nil {
		return nil, err
	}
	return rows, nil
}

func mixedFFIRoutingSentinelGoRowLiterals(rows []mixedFFIRoutingSentinelRow) string {
	var out strings.Builder
	for _, row := range rows {
		fmt.Fprintf(&out, "\t%q,\n", row.text)
	}
	return out.String()
}

// mixedFFIRoutingSentinelDecimal decorates the anchor proposal with the
// oracle's decimal interpretation of one width-tagged value; a transport
// failure degrades to "?" and is re-raised by tier1SentinelOracleErr before
// any row is returned.
func mixedFFIRoutingSentinelDecimal(width int, value string) string {
	payload, err := tier1SentinelOracleQuery(fmt.Sprintf("str %d %s", width, value))
	if err != nil {
		return "?"
	}
	return payload
}

// MixedFFIRoutingSentinelAnchorProposal prints the proposed
// verification_sentinels.json rows for the human pin (stdout only; no file is
// read or written). Each row carries the decimal interpretation of x, y, the
// pinned result, and the swapped-operand result so the auditor can confirm
// op(x,y) is correct and that swapping the slots changes it.
func MixedFFIRoutingSentinelAnchorProposal() (string, error) {
	rows, err := GenerateMixedFFIRoutingSentinelRows()
	if err != nil {
		return "", err
	}
	texts := make([]string, 0, len(rows))
	for _, row := range rows {
		texts = append(texts, row.text)
	}
	proposal := struct {
		Rows []string `json:"mixed_format_ffi_routing_sentinel_rows"`
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
		resultBits, _, _ := strings.Cut(row.payload, "/")
		swapped, swapErr := mixedFFIRoutingSentinelOracle(row.function, row.mode, row.yText, row.xText)
		swappedResult := swapped
		if swapErr == nil {
			swappedBits, _, _ := strings.Cut(swapped, "/")
			swappedResult = fmt.Sprintf("%s (=%s)", swapped, mixedFFIRoutingSentinelDecimal(row.resultBits, swappedBits))
		} else {
			swappedResult = "?"
		}
		fmt.Fprintf(&out, "# %s\n#   x=%s y=%s -> %s | swapped op(y,x) -> %s\n#   %s\n",
			row.text,
			mixedFFIRoutingSentinelDecimal(row.operandBits, row.xText),
			mixedFFIRoutingSentinelDecimal(row.operandBits, row.yText),
			mixedFFIRoutingSentinelDecimal(row.resultBits, resultBits),
			swappedResult,
			row.comment)
	}
	if err := tier1SentinelOracleErr(); err != nil {
		return "", err
	}
	return out.String(), nil
}
