package testgen

// Tier 1 compare/conversion routing-sentinel codegen.
//
// Same defense as tier1_sentinel_codegen.go, applied to the compare/
// conversion long runners. Family structure follows the runner dispatch
// tables exactly:
//
//	quiet        6 predicates per width (two decimal operands, modeless)
//	minmax       4 operations per width (two decimal operands, modeless)
//	to-int       80 operations per width (one operand; rounding direction and
//	             exactness are part of the operation name, so there is no m=)
//	width        2/6/10 operations per source width (narrowing rows carry m=,
//	             widening rows are exact and modeless)
//	binary       15 operations per source width (dest x mode; m= always)
//	constructor  36 operations over one flat table (dest x kind x mode;
//	             rounded rows carry m=, exact rows do not)
//
// Sensitivity requirements: S1 (slot swap) applies only to the two-operand
// quiet/minmax families; there is no S2 anywhere (rounding modes are part of
// the dispatch-row identity in every moded compare/conversion family); S3
// requires every dispatch-table row to carry at least one sentinel row whose
// pinned answer distinguishes it from every same-signature sibling row.
// Structurally unsatisfiable requirements are waived through
// tier1SentinelCCExceptions (closed world: a waiver the candidate pool can
// actually satisfy, or one that never fires, fails generation).
//
// GUARDRAILS: no read or write of verification_anchors.json or
// verification_sentinels.json; the human pin flows through
// `cmd/testgen -print-sentinel-anchors` stdout only.

import (
	"fmt"
	"strings"
)

// Per-family adopted-tuple caps: the row budget that keeps the pinned set
// auditable. The to-int family needs up to three rows per operation (a
// same-kind tie, a same-kind boundary overflow, and one generic fraction) to
// separate all 79 same-signature siblings; a rounded constructor row needs up
// to four (a tie, a directed fraction, the signed/unsigned register splitter,
// and the 32/64-bit register splitter).
const (
	tier1SentinelCCQuietCap       = 4
	tier1SentinelCCMinMaxCap      = 3
	tier1SentinelCCToIntCap       = 3
	tier1SentinelCCWidthCap       = 2
	tier1SentinelCCBinaryCap      = 2
	tier1SentinelCCConstructorCap = 4
)

// ---------------------------------------------------------------------------
// Operation tables (mirror the runner dispatch declaration order exactly)
// ---------------------------------------------------------------------------

var tier1SentinelCCQuietOps = [...]string{
	"quiet_equal", "quiet_not_equal", "quiet_less", "quiet_less_equal",
	"quiet_greater", "quiet_greater_equal",
}

var tier1SentinelCCMinMaxOps = [...]string{"minnum", "maxnum", "minnum_mag", "maxnum_mag"}

type tier1SentinelCCToIntOp struct {
	kind   string
	exact  bool
	mode   tier1SentinelMode
	suffix string
}

func (o tier1SentinelCCToIntOp) label() string {
	return "to_" + o.kind + "_" + o.suffix
}

func tier1SentinelCCToIntOps() []tier1SentinelCCToIntOp {
	kinds := []string{"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64"}
	suffixes := map[bool]map[int]string{
		false: {0: "rnint", 4: "rninta", 3: "int", 2: "ceil", 1: "floor"},
		true:  {0: "xrnint", 4: "xrninta", 3: "xint", 2: "xceil", 1: "xfloor"},
	}
	ops := make([]tier1SentinelCCToIntOp, 0, 80)
	for _, kind := range kinds {
		for _, exact := range []bool{false, true} {
			for _, mode := range tier1SentinelModes {
				ops = append(ops, tier1SentinelCCToIntOp{
					kind: kind, exact: exact, mode: mode, suffix: suffixes[exact][mode.native],
				})
			}
		}
	}
	return ops
}

type tier1SentinelCCWidthOp struct {
	dest    int
	rounded bool
	mode    tier1SentinelMode
}

func (o tier1SentinelCCWidthOp) label() string { return fmt.Sprintf("to_bid%d", o.dest) }

// key identifies the dispatch-table row (mode is part of the identity for
// narrowing rows).
func (o tier1SentinelCCWidthOp) key() string {
	if o.rounded {
		return fmt.Sprintf("%s_m%d", o.label(), o.mode.native)
	}
	return o.label()
}

func tier1SentinelCCWidthOps(source int) []tier1SentinelCCWidthOp {
	var ops []tier1SentinelCCWidthOp
	addRounded := func(dest int) {
		for _, mode := range tier1SentinelModes {
			ops = append(ops, tier1SentinelCCWidthOp{dest: dest, rounded: true, mode: mode})
		}
	}
	switch source {
	case 32:
		ops = append(ops, tier1SentinelCCWidthOp{dest: 64}, tier1SentinelCCWidthOp{dest: 128})
	case 64:
		addRounded(32)
		ops = append(ops, tier1SentinelCCWidthOp{dest: 128})
	case 128:
		addRounded(32)
		addRounded(64)
	default:
		panic(fmt.Sprintf("unsupported tier1 sentinel width-conversion source %d", source))
	}
	return ops
}

type tier1SentinelCCBinaryOp struct {
	dest int
	mode tier1SentinelMode
}

func (o tier1SentinelCCBinaryOp) label() string { return fmt.Sprintf("to_binary%d", o.dest) }
func (o tier1SentinelCCBinaryOp) key() string {
	return fmt.Sprintf("%s_m%d", o.label(), o.mode.native)
}

func tier1SentinelCCBinaryOps() []tier1SentinelCCBinaryOp {
	ops := make([]tier1SentinelCCBinaryOp, 0, 15)
	for _, dest := range []int{32, 64, 128} {
		for _, mode := range tier1SentinelModes {
			ops = append(ops, tier1SentinelCCBinaryOp{dest: dest, mode: mode})
		}
	}
	return ops
}

type tier1SentinelCCConstructorOp struct {
	dest    int
	kind    string
	rounded bool
	mode    tier1SentinelMode
}

func (o tier1SentinelCCConstructorOp) label() string { return "from_" + o.kind }
func (o tier1SentinelCCConstructorOp) key() string {
	if o.rounded {
		return fmt.Sprintf("%s_d%d_m%d", o.label(), o.dest, o.mode.native)
	}
	return fmt.Sprintf("%s_d%d", o.label(), o.dest)
}

func tier1SentinelCCConstructorOps() []tier1SentinelCCConstructorOp {
	ops := make([]tier1SentinelCCConstructorOp, 0, 36)
	for _, kind := range []string{"int32", "uint32", "int64", "uint64"} {
		for _, mode := range tier1SentinelModes {
			ops = append(ops, tier1SentinelCCConstructorOp{dest: 32, kind: kind, rounded: true, mode: mode})
		}
	}
	ops = append(ops,
		tier1SentinelCCConstructorOp{dest: 64, kind: "int32"},
		tier1SentinelCCConstructorOp{dest: 64, kind: "uint32"},
	)
	for _, kind := range []string{"int64", "uint64"} {
		for _, mode := range tier1SentinelModes {
			ops = append(ops, tier1SentinelCCConstructorOp{dest: 64, kind: kind, rounded: true, mode: mode})
		}
	}
	for _, kind := range []string{"int32", "uint32", "int64", "uint64"} {
		ops = append(ops, tier1SentinelCCConstructorOp{dest: 128, kind: kind})
	}
	return ops
}

// ---------------------------------------------------------------------------
// Candidate pools (hand-declared, fixed order)
// ---------------------------------------------------------------------------

// tier1SentinelCCComparePool serves quiet and minmax: asymmetric finite pairs
// for slot/dispatch separation, the equal pair and both orders for the quiet
// predicate vectors, an sNaN pair for the invalid-flag axis, distinct qNaN
// payloads for the value-commutative min/max slot swap, and the signed-zero
// pair as a further min/max slot probe.
func tier1SentinelCCComparePool(width int) []tier1SentinelBinaryTuple {
	qA, qB := tier1SentinelQNaNPayloadPair(width)
	var sNaN, negZero, posZero tier1SentinelOperand
	switch width {
	case 32:
		sNaN, negZero, posZero = sentinelRaw32(0x7e000001), sentinelRaw32(0x80000000), sentinelRaw32(0x00000000)
	case 64:
		sNaN, negZero, posZero = sentinelRaw64(0x7e00000000000001), sentinelRaw64(0x8000000000000000), sentinelRaw64(0)
	case 128:
		sNaN = sentinelRaw128(0x0000000000000001, 0x7e00000000000000)
		negZero = sentinelRaw128(0, 0x8000000000000000)
		posZero = sentinelRaw128(0, 0)
	default:
		panic(fmt.Sprintf("unsupported tier1 sentinel width %d", width))
	}
	return []tier1SentinelBinaryTuple{
		{x: sentinelDec(mdoNeg(5, 0)), y: sentinelDec(mdo(3, 0))},
		{x: sentinelDec(mdoNeg(3, 0)), y: sentinelDec(mdo(5, 0))},
		{x: sentinelDec(mdo(5, 0)), y: sentinelDec(mdo(3, 0))},
		{x: sentinelDec(mdo(3, 0)), y: sentinelDec(mdo(3, 0))},
		{x: sentinelDec(mdo(3, 0)), y: sentinelDec(mdo(5, 0))},
		{x: sNaN, y: sentinelDec(mdo(3, 0))},
		{x: qA, y: qB},
		{x: negZero, y: posZero},
	}
}

// tier1SentinelCCToIntPool serves the 80-operation to-int family. Ordering
// puts the per-kind boundary values first (an own-kind overflow separates
// every other kind at once through the type-specific invalid returns), then
// the per-kind ties, then generic fractions for the direction axis and the
// signed/unsigned axis.
func tier1SentinelCCToIntPool() []tier1SentinelOperand {
	return []tier1SentinelOperand{
		sentinelDec(mdo(1305, -1)),    // 130.5: int8 overflow, uint8 fine
		sentinelDec(mdo(2555, -1)),    // 255.5: uint8 tie at its max
		sentinelDec(mdo(300, 0)),      // uint8 overflow, int16 fine
		sentinelDec(mdo(327675, -1)),  // 32767.5: int16 tie at its max
		sentinelDec(mdo(33000, 0)),    // int16 overflow, uint16 fine
		sentinelDec(mdo(655355, -1)),  // 65535.5: uint16 tie at its max
		sentinelDec(mdo(70000, 0)),    // uint16 overflow, int32 fine
		sentinelDec(mdo(2147484, 3)),  // int32 overflow, uint32 fine
		sentinelDec(mdo(4294968, 3)),  // uint32 overflow, int64 fine
		sentinelDec(mdo(922338, 13)), // 9.22338e18: int64 overflow, uint64 fine
		sentinelDec(mdo(1844675, 13)), // uint64 overflow
		sentinelDec(mdo(1275, -1)),    // 127.5: int8 tie at its max
		sentinelDec(mdo(25, -1)),      // 2.5: half-even/half-away tie
		sentinelDec(mdo(35, -1)),      // 3.5: tie rounding up under half-even
		sentinelDec(mdo(27, -1)),      // 2.7: above-half fraction
		sentinelDec(mdo(22, -1)),      // 2.2: below-half fraction
		sentinelDec(mdoNeg(25, -1)),   // -2.5
		sentinelDec(mdoNeg(27, -1)),   // -2.7
		sentinelDec(mdoNeg(22, -1)),   // -2.2
		sentinelDec(mdoNeg(1, 0)),     // -1: unsigned targets go invalid
		sentinelDec(mdoNeg(7, -1)),    // -0.7: truncation lands on 0 for unsigned targets, floor goes invalid
		sentinelDec(mdoNeg(5, -1)),    // -0.5: tie at the unsigned validity edge
		sentinelDec(mdoNeg(2, -1)),    // -0.2: below-half at the unsigned validity edge
	}
}

// tier1SentinelCCWidthPool serves the width-conversion family per source:
// precision ties and directed fractions at both narrowing boundaries plus an
// exact operand for the widening rows.
func tier1SentinelCCWidthPool(source int) []tier1SentinelOperand {
	switch source {
	case 32:
		return []tier1SentinelOperand{sentinelDec(mdo(1, 0)), sentinelDec(mdoNeg(1, 0))}
	case 64:
		return []tier1SentinelOperand{
			sentinelDec(mdo(10000005, -7)),    // 1.0000005: tie at the 7-digit boundary
			sentinelDec(mdo(10000001, -7)),    // below-half drop
			sentinelDec(mdo(10000009, -7)),    // above-half drop
			sentinelDec(mdoNeg(10000001, -7)), // negative mirror
			sentinelDec(mdo(1, 0)),
		}
	case 128:
		return []tier1SentinelOperand{
			sentinelDec(mdo(10000000000000005, -16)),    // tie at the 16-digit boundary
			sentinelDec(mdo(10000005, -7)),              // tie at the 7-digit boundary
			sentinelDec(mdo(10000000000000001, -16)),    // below-half drop
			sentinelDec(mdo(10000000000000009, -16)),    // above-half drop
			sentinelDec(mdoNeg(10000000000000001, -16)), // negative mirror
			sentinelDec(mdo(10000001, -7)),
			sentinelDec(mdo(10000009, -7)),
			sentinelDec(mdoNeg(10000001, -7)),
			sentinelDec(mdo(1, 0)),
		}
	default:
		panic(fmt.Sprintf("unsupported tier1 sentinel width-conversion source %d", source))
	}
}

// tier1SentinelCCBinaryPool serves the binary-conversion family. Every entry
// is representable at every source width. The exact binary ties (values that
// land exactly halfway between two representable binary values) separate
// half-even from half-away per destination:
//
//	1077e6     = 16828125 * 2^6: binary32 tie in [2^30, 2^31) whose lower
//	             neighbor mantissa is even, so half-even stays down while
//	             half-away goes up
//	59049e16   = 59049 * 5^16 * 2^16: binary64 tie in [2^69, 2^70)
//	5708993e39 = odd * 2^39: binary128 tie in [2^152, 2^153)
func tier1SentinelCCBinaryPool() []tier1SentinelOperand {
	return []tier1SentinelOperand{
		sentinelDec(mdo(1077, 6)),
		sentinelDec(mdo(59049, 16)),
		sentinelDec(mdo(5708993, 39)),
		sentinelDec(mdo(1, -1)),    // 0.1: non-terminating in binary
		sentinelDec(mdoNeg(1, -1)), // -0.1
		sentinelDec(mdo(7, -1)),    // 0.7
		sentinelDec(mdoNeg(7, -1)),
		sentinelDec(mdo(1, 0)),
	}
}

// tier1SentinelCCConstructorPool serves the flat 36-operation constructor
// table. Candidates are the raw 64-bit register values the runner dispatch
// consumes; each operation reinterprets the register through its own kind
// cast exactly as the runner does, so a mislabeled row is evaluated on the
// same register image.
func tier1SentinelCCConstructorPool(op tier1SentinelCCConstructorOp) []uint64 {
	tie64 := uint64(12345679012345645) // 17 digits: half-even/half-away tie at 16
	above64 := uint64(12345679012345647)
	small := []uint64{
		12345645,             // 8 digits: tie at the 7-digit boundary
		12345647,             // above-half drop
		12345641,             // below-half drop
		18446744073697205969, // -12345647 as a 64-bit register
		0xffffffffffffffff,   // -1 / uint64 max: signed-unsigned and 32/64 splitter
		10000000000,          // above 2^32: 32-bit kinds truncate
		127,                  // exact everywhere
	}
	switch op.dest {
	case 32:
		return small
	case 64, 128:
		// 18434398394697205969 is -12345679012345647 as a 64-bit register.
		return append([]uint64{tie64, above64, 18434398394697205969}, small...)
	default:
		panic(fmt.Sprintf("unsupported tier1 sentinel constructor destination %d", op.dest))
	}
}

// ---------------------------------------------------------------------------
// Exceptions (closed world)
// ---------------------------------------------------------------------------

// tier1SentinelCCException waives one structurally unsatisfiable requirement
// of one compare/conversion dispatch row. scope is the row-order width/source
// label the waiver applies to ("d32"/"d64"/"d128"); expectedUses pins how
// many selection runs must consume the entry, so a stale or mis-keyed waiver
// fails generation.
type tier1SentinelCCException struct {
	family string
	scope  string
	op     string // op key inside the family
	key    string // requirement key waived for that op
	reason string
}

var tier1SentinelCCExceptions = []tier1SentinelCCException{
	{family: "quiet", scope: "d32", op: "quiet_equal", key: "slot:x,y",
		reason: "quiet equality is a symmetric relation for every operand class (finite, zero, infinite, NaN), so no operand pair can make a slot swap visible"},
	{family: "quiet", scope: "d64", op: "quiet_equal", key: "slot:x,y", reason: "same symmetry as d32"},
	{family: "quiet", scope: "d128", op: "quiet_equal", key: "slot:x,y", reason: "same symmetry as d32"},
	{family: "quiet", scope: "d32", op: "quiet_not_equal", key: "slot:x,y",
		reason: "quiet inequality is the negation of a symmetric relation and inherits its symmetry"},
	{family: "quiet", scope: "d64", op: "quiet_not_equal", key: "slot:x,y", reason: "same symmetry as d32"},
	{family: "quiet", scope: "d128", op: "quiet_not_equal", key: "slot:x,y", reason: "same symmetry as d32"},
	{family: "constructor", scope: "d32", op: "from_uint32_d32_m3", key: "dispatch:from_uint32_d32_m1",
		reason: "unsigned constructor inputs are never negative, so toward_zero and toward_negative always round down identically"},
	{family: "constructor", scope: "d32", op: "from_uint32_d32_m1", key: "dispatch:from_uint32_d32_m3", reason: "same non-negative rounding identity"},
	{family: "constructor", scope: "d32", op: "from_uint64_d32_m3", key: "dispatch:from_uint64_d32_m1", reason: "same non-negative rounding identity"},
	{family: "constructor", scope: "d32", op: "from_uint64_d32_m1", key: "dispatch:from_uint64_d32_m3", reason: "same non-negative rounding identity"},
	{family: "constructor", scope: "d64", op: "from_uint64_d64_m3", key: "dispatch:from_uint64_d64_m1", reason: "same non-negative rounding identity"},
	{family: "constructor", scope: "d64", op: "from_uint64_d64_m1", key: "dispatch:from_uint64_d64_m3", reason: "same non-negative rounding identity"},
}

type tier1SentinelCCExceptionUse map[int]bool

// tier1SentinelCCValidateExceptions structurally validates the cc waiver
// table at generation time, before any selection runs: every entry needs a
// non-empty written reason, a family and width scope from the closed
// compare/conversion universe, a named operation, and a requirement key of a
// waivable kind (operand-slot or dispatch-sibling — the only requirement
// kinds the cc domain defines). Whether the waived key exists for the named
// dispatch row is enforced when the entry is applied, and whether it is
// genuinely unsatisfiable is verified against the full candidate pool per
// selection run.
func tier1SentinelCCValidateExceptions(entries []tier1SentinelCCException) error {
	families := map[string]bool{
		"quiet": true, "minmax": true, "to-int": true,
		"width": true, "binary": true, "constructor": true,
	}
	scopes := map[string]bool{}
	for _, width := range tier1SentinelWidths {
		scopes[tier1SentinelWidthLabel(width)] = true
	}
	for i, exception := range entries {
		if strings.TrimSpace(exception.reason) == "" {
			return fmt.Errorf("tier1 sentinel cc exception %d (%s %s %s waives %s) has no written reason",
				i, exception.family, exception.scope, exception.op, exception.key)
		}
		if !families[exception.family] {
			return fmt.Errorf("tier1 sentinel cc exception %d waives %s for unknown family %q",
				i, exception.key, exception.family)
		}
		if !scopes[exception.scope] {
			return fmt.Errorf("tier1 sentinel cc exception %d (%s %s) has unknown width scope %q",
				i, exception.family, exception.op, exception.scope)
		}
		if exception.op == "" {
			return fmt.Errorf("tier1 sentinel cc exception %d (%s %s waives %s) names no operation",
				i, exception.family, exception.scope, exception.key)
		}
		if !strings.HasPrefix(exception.key, "slot:") && !strings.HasPrefix(exception.key, "dispatch:") {
			return fmt.Errorf("tier1 sentinel cc exception %d (%s %s %s): key %q is not a waivable requirement kind (slot:/dispatch:)",
				i, exception.family, exception.scope, exception.op, exception.key)
		}
	}
	return nil
}

// tier1SentinelCCApplyExceptions removes waived requirement keys for one
// dispatch row before selection. A matched entry whose key is not a live
// requirement of that row fails generation immediately (mis-keyed or
// duplicate waiver).
func tier1SentinelCCApplyExceptions(family, scope, op string, reqs *tier1SentinelRequirements) ([]int, error) {
	return tier1SentinelCCApplyExceptionEntries(tier1SentinelCCExceptions, family, scope, op, reqs)
}

func tier1SentinelCCApplyExceptionEntries(entries []tier1SentinelCCException, family, scope, op string, reqs *tier1SentinelRequirements) ([]int, error) {
	applied := []int{}
	for i, exception := range entries {
		if exception.family != family || exception.scope != scope || exception.op != op {
			continue
		}
		if !reqs.unmet[exception.key] {
			return nil, fmt.Errorf("tier1 sentinel cc exception (%s %s %s waives %s): the waived key is not a live requirement of that dispatch row",
				exception.family, exception.scope, exception.op, exception.key)
		}
		reqs.drop(exception.key)
		applied = append(applied, i)
	}
	return applied, nil
}

// tier1SentinelCCVerifyWaivers re-checks each applied waiver against the full
// candidate pool through the caller-supplied separability probe: a waived
// requirement the pool can actually satisfy is stale and fails generation.
func tier1SentinelCCVerifyWaivers(applied []int, use tier1SentinelCCExceptionUse, separable func(key string) (bool, error)) error {
	for _, index := range applied {
		exception := tier1SentinelCCExceptions[index]
		ok, err := separable(exception.key)
		if err != nil {
			return err
		}
		if ok {
			return fmt.Errorf("tier1 sentinel cc exception (%s %s %s waives %s) is stale: the candidate pool can satisfy the waived requirement",
				exception.family, exception.scope, exception.op, exception.key)
		}
		use[index] = true
	}
	return nil
}

func (u tier1SentinelCCExceptionUse) requireAllUsed() error {
	for i, exception := range tier1SentinelCCExceptions {
		if !u[i] {
			return fmt.Errorf("tier1 sentinel cc exception %d (%s %s %s waives %s) never fired; remove or rescope the stale waiver",
				i, exception.family, exception.scope, exception.op, exception.key)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Oracle (pin-time oracle subprocess -> canonical result strings)
// ---------------------------------------------------------------------------

func tier1SentinelCCEvalQuiet(width int, op string, x, y bid128BidCodecValue) (string, error) {
	xText, err := tier1SentinelWidthValueText(width, x)
	if err != nil {
		return "", err
	}
	yText, err := tier1SentinelWidthValueText(width, y)
	if err != nil {
		return "", err
	}
	return tier1SentinelOracleQuery(fmt.Sprintf("quiet %d %s %s %s", width, op, xText, yText))
}

func tier1SentinelCCEvalMinMax(width int, op string, x, y bid128BidCodecValue) (string, error) {
	xText, err := tier1SentinelWidthValueText(width, x)
	if err != nil {
		return "", err
	}
	yText, err := tier1SentinelWidthValueText(width, y)
	if err != nil {
		return "", err
	}
	return tier1SentinelOracleQuery(fmt.Sprintf("minmax %d %s %s %s", width, op, xText, yText))
}

// tier1SentinelCCEvalToInt normalizes every integer conversion to the
// cross-language canonical form: the result as a 64-bit two's-complement
// register (Rust's `as i64 as u64` convention) plus the raw flag word.
func tier1SentinelCCEvalToInt(width int, op tier1SentinelCCToIntOp, x bid128BidCodecValue) (string, error) {
	xText, err := tier1SentinelWidthValueText(width, x)
	if err != nil {
		return "", err
	}
	exact := 0
	if op.exact {
		exact = 1
	}
	return tier1SentinelOracleQuery(fmt.Sprintf("toint %d %s %d %d %s", width, op.kind, exact, op.mode.native, xText))
}

func tier1SentinelCCEvalWidth(source int, op tier1SentinelCCWidthOp, x bid128BidCodecValue) (string, error) {
	xText, err := tier1SentinelWidthValueText(source, x)
	if err != nil {
		return "", err
	}
	return tier1SentinelOracleQuery(fmt.Sprintf("widthconv %d %d %d %s", source, op.dest, op.mode.native, xText))
}

func tier1SentinelCCEvalBinary(source int, op tier1SentinelCCBinaryOp, x bid128BidCodecValue) (string, error) {
	xText, err := tier1SentinelWidthValueText(source, x)
	if err != nil {
		return "", err
	}
	return tier1SentinelOracleQuery(fmt.Sprintf("binaryconv %d %d %d %s", source, op.dest, op.mode.native, xText))
}

// tier1SentinelCCConstructorInput renders the register through the
// operation's own kind cast, exactly as the runner's structured differential
// builds it (32-bit kinds consume the low register word).
func tier1SentinelCCConstructorInput(op tier1SentinelCCConstructorOp, raw uint64) string {
	switch op.kind {
	case "int32":
		return fmt.Sprint(int32(uint32(raw)))
	case "uint32":
		return fmt.Sprint(uint32(raw))
	case "int64":
		return fmt.Sprint(int64(raw))
	case "uint64":
		return fmt.Sprint(raw)
	default:
		panic("unknown tier1 sentinel constructor kind: " + op.kind)
	}
}

func tier1SentinelCCEvalConstructor(op tier1SentinelCCConstructorOp, raw uint64) (string, error) {
	return tier1SentinelOracleQuery(fmt.Sprintf("constructor %d %s %d %d", op.dest, op.kind, op.mode.native, raw))
}

// ---------------------------------------------------------------------------
// Selection
// ---------------------------------------------------------------------------
//
// The compare/conversion selectors use a deterministic maximum-gain set-cover
// greedy (pick the candidate that covers the most unmet requirements, ties
// broken by the earliest pool index) instead of the arithmetic domain's
// first-gain walk: to-int operations face 79 sibling-separation requirements
// each, and a first-gain walk adopts weak early candidates that inflate the
// pinned row count. Both flavors are pure functions of the fixed tables, so
// regeneration stays byte-identical.

// tier1SentinelCCCoverSelect runs the max-gain cover over precomputed
// per-candidate coverage sets. Returns the adopted candidate indices in
// adoption order.
func tier1SentinelCCCoverSelect(reqs *tier1SentinelRequirements, coverage []map[string]bool) []int {
	adopted := []int{}
	for len(reqs.unmet) > 0 {
		best, bestGain := -1, 0
		for i, cover := range coverage {
			gain := 0
			for key := range reqs.unmet {
				if cover[key] {
					gain++
				}
			}
			if gain > bestGain {
				best, bestGain = i, gain
			}
		}
		if best < 0 {
			return adopted // caller reports the remaining requirements
		}
		adopted = append(adopted, best)
		for key := range coverage[best] {
			reqs.satisfy(key)
		}
	}
	return adopted
}

// tier1SentinelCCSelectBinaryFamily runs the cover greedy for a two-operand
// modeless family (quiet, minmax): requirements are the slot swap plus
// dispatch separation from every sibling operation.
func tier1SentinelCCSelectBinaryFamily(family, scope string, width int, ops []string, op string,
	pool []tier1SentinelBinaryTuple, cap int, use tier1SentinelCCExceptionUse,
	eval func(op string, x, y bid128BidCodecValue) (string, error)) ([]tier1SentinelBinaryTuple, error) {
	keys := []string{"slot:x,y"}
	for _, other := range ops {
		if other != op {
			keys = append(keys, "dispatch:"+other)
		}
	}
	reqs := newTier1SentinelRequirements(keys)
	applied, err := tier1SentinelCCApplyExceptions(family, scope, op, reqs)
	if err != nil {
		return nil, err
	}

	coverage := make([]map[string]bool, 0, len(pool))
	for _, cand := range pool {
		x, err := cand.x.encode(width)
		if err != nil {
			return nil, err
		}
		y, err := cand.y.encode(width)
		if err != nil {
			return nil, err
		}
		result, err := eval(op, x, y)
		if err != nil {
			return nil, err
		}
		cover := map[string]bool{}
		swapped, err := eval(op, y, x)
		if err != nil {
			return nil, err
		}
		if swapped != result {
			cover["slot:x,y"] = true
		}
		for _, other := range ops {
			if other == op {
				continue
			}
			otherResult, err := eval(other, x, y)
			if err != nil {
				return nil, err
			}
			if otherResult != result {
				cover["dispatch:"+other] = true
			}
		}
		coverage = append(coverage, cover)
	}
	adoptedIndexes := tier1SentinelCCCoverSelect(reqs, coverage)
	if remaining := reqs.remaining(); len(remaining) > 0 {
		return nil, tier1SentinelSelectionFailure(width, family+" "+op, remaining)
	}
	if len(adoptedIndexes) > cap {
		return nil, tier1SentinelCapFailure(width, family+" "+op, len(adoptedIndexes), cap)
	}
	if err := tier1SentinelCCVerifyWaivers(applied, use, func(key string) (bool, error) {
		for _, cover := range coverage {
			if cover[key] {
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		return nil, err
	}
	adopted := make([]tier1SentinelBinaryTuple, len(adoptedIndexes))
	for i, index := range adoptedIndexes {
		adopted[i] = pool[index]
	}
	return adopted, nil
}

// tier1SentinelCCSelectUnaryFamily runs the cover greedy for a one-operand
// family (to-int, width, binary): requirements are dispatch separation from
// every same-signature sibling row, keyed by the sibling row identity.
func tier1SentinelCCSelectUnaryFamily(family, scope string, width int, opKeys []string, opIndex int,
	pool []tier1SentinelOperand, cap int, use tier1SentinelCCExceptionUse,
	eval func(index int, x bid128BidCodecValue) (string, error)) ([]tier1SentinelOperand, error) {
	keys := []string{}
	for i, key := range opKeys {
		if i != opIndex {
			keys = append(keys, "dispatch:"+key)
		}
	}
	reqs := newTier1SentinelRequirements(keys)
	applied, err := tier1SentinelCCApplyExceptions(family, scope, opKeys[opIndex], reqs)
	if err != nil {
		return nil, err
	}

	coverage := make([]map[string]bool, 0, len(pool))
	for _, cand := range pool {
		x, err := cand.encode(width)
		if err != nil {
			return nil, err
		}
		result, err := eval(opIndex, x)
		if err != nil {
			return nil, err
		}
		cover := map[string]bool{}
		for i, key := range opKeys {
			if i == opIndex {
				continue
			}
			otherResult, err := eval(i, x)
			if err != nil {
				return nil, err
			}
			if otherResult != result {
				cover["dispatch:"+key] = true
			}
		}
		coverage = append(coverage, cover)
	}
	adoptedIndexes := tier1SentinelCCCoverSelect(reqs, coverage)
	if remaining := reqs.remaining(); len(remaining) > 0 {
		return nil, tier1SentinelSelectionFailure(width, family+" "+opKeys[opIndex], remaining)
	}
	if len(adoptedIndexes) > cap {
		return nil, tier1SentinelCapFailure(width, family+" "+opKeys[opIndex], len(adoptedIndexes), cap)
	}
	if err := tier1SentinelCCVerifyWaivers(applied, use, func(key string) (bool, error) {
		for _, cover := range coverage {
			if cover[key] {
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		return nil, err
	}
	adopted := make([]tier1SentinelOperand, len(adoptedIndexes))
	for i, index := range adoptedIndexes {
		adopted[i] = pool[index]
	}
	return adopted, nil
}

// tier1SentinelCCSelectConstructor runs the cover greedy for one constructor
// dispatch row over the flat 36-row table (the register image is shared
// across kinds, exactly as the runner dispatch consumes it).
func tier1SentinelCCSelectConstructor(scope string, ops []tier1SentinelCCConstructorOp, opIndex int,
	use tier1SentinelCCExceptionUse) ([]uint64, error) {
	op := ops[opIndex]
	keys := []string{}
	for i, other := range ops {
		if i != opIndex {
			keys = append(keys, "dispatch:"+other.key())
		}
	}
	reqs := newTier1SentinelRequirements(keys)
	applied, err := tier1SentinelCCApplyExceptions("constructor", scope, op.key(), reqs)
	if err != nil {
		return nil, err
	}

	pool := tier1SentinelCCConstructorPool(op)
	coverage := make([]map[string]bool, 0, len(pool))
	for _, raw := range pool {
		result, err := tier1SentinelCCEvalConstructor(op, raw)
		if err != nil {
			return nil, err
		}
		cover := map[string]bool{}
		for i, other := range ops {
			if i == opIndex {
				continue
			}
			otherResult, err := tier1SentinelCCEvalConstructor(other, raw)
			if err != nil {
				return nil, err
			}
			if otherResult != result {
				cover["dispatch:"+other.key()] = true
			}
		}
		coverage = append(coverage, cover)
	}
	adoptedIndexes := tier1SentinelCCCoverSelect(reqs, coverage)
	if remaining := reqs.remaining(); len(remaining) > 0 {
		return nil, tier1SentinelSelectionFailure(op.dest, "constructor "+op.key(), remaining)
	}
	if len(adoptedIndexes) > tier1SentinelCCConstructorCap {
		return nil, tier1SentinelCapFailure(op.dest, "constructor "+op.key(), len(adoptedIndexes), tier1SentinelCCConstructorCap)
	}
	if err := tier1SentinelCCVerifyWaivers(applied, use, func(key string) (bool, error) {
		for _, cover := range coverage {
			if cover[key] {
				return true, nil
			}
		}
		return false, nil
	}); err != nil {
		return nil, err
	}
	adopted := make([]uint64, len(adoptedIndexes))
	for i, index := range adoptedIndexes {
		adopted[i] = pool[index]
	}
	return adopted, nil
}

// ---------------------------------------------------------------------------
// Row emission
// ---------------------------------------------------------------------------

// GenerateTier1CompareConversionSentinelRows selects and serializes the
// compare/conversion routing-sentinel rows in canonical order: width
// ascending, then family declaration order (quiet, minmax, to-int, width
// conversion, binary conversion, constructor by destination), then operation
// declaration order, then tuple adoption order.
func GenerateTier1CompareConversionSentinelRows() ([]tier1SentinelRow, error) {
	if err := tier1SentinelCCValidateExceptions(tier1SentinelCCExceptions); err != nil {
		return nil, err
	}
	exceptionUse := tier1SentinelCCExceptionUse{}
	rows := []tier1SentinelRow{}
	var err error
	appendRow := func(text, comment string) error {
		rows, err = tier1SentinelAppendRow(rows, text, comment)
		return err
	}
	toIntOps := tier1SentinelCCToIntOps()
	toIntKeys := make([]string, len(toIntOps))
	for i, op := range toIntOps {
		toIntKeys[i] = op.label()
	}
	binaryOps := tier1SentinelCCBinaryOps()
	binaryKeys := make([]string, len(binaryOps))
	for i, op := range binaryOps {
		binaryKeys[i] = op.key()
	}
	constructorOps := tier1SentinelCCConstructorOps()

	quietOps := tier1SentinelCCQuietOps[:]
	minMaxOps := tier1SentinelCCMinMaxOps[:]
	for _, width := range tier1SentinelWidths {
		label := tier1SentinelWidthLabel(width)
		pool := tier1SentinelCCComparePool(width)
		for _, op := range quietOps {
			adopted, selErr := tier1SentinelCCSelectBinaryFamily("quiet", label, width, quietOps, op,
				pool, tier1SentinelCCQuietCap, exceptionUse, func(evalOp string, x, y bid128BidCodecValue) (string, error) {
					return tier1SentinelCCEvalQuiet(width, evalOp, x, y)
				})
			if selErr != nil {
				return nil, selErr
			}
			for _, tuple := range adopted {
				x, encErr := tuple.x.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				y, encErr := tuple.y.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				result, evalErr := tier1SentinelCCEvalQuiet(width, op, x, y)
				if evalErr != nil {
					return nil, evalErr
				}
				text := fmt.Sprintf("%s %s x=%s y=%s -> %s",
					label, op, tier1SentinelValueText(width, x), tier1SentinelValueText(width, y), result)
				comment := fmt.Sprintf("%s %s %s , %s = %s , flags %s",
					label, op, tier1SentinelDecimalComment(width, x), tier1SentinelDecimalComment(width, y),
					strings.TrimSuffix(strings.SplitN(result, "/", 2)[0], ""), tier1SentinelFlagsComment(result))
				if err := appendRow(text, comment); err != nil {
					return nil, err
				}
			}
		}
		for _, op := range minMaxOps {
			adopted, selErr := tier1SentinelCCSelectBinaryFamily("minmax", label, width, minMaxOps, op,
				pool, tier1SentinelCCMinMaxCap, exceptionUse, func(evalOp string, x, y bid128BidCodecValue) (string, error) {
					return tier1SentinelCCEvalMinMax(width, evalOp, x, y)
				})
			if selErr != nil {
				return nil, selErr
			}
			for _, tuple := range adopted {
				x, encErr := tuple.x.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				y, encErr := tuple.y.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				result, evalErr := tier1SentinelCCEvalMinMax(width, op, x, y)
				if evalErr != nil {
					return nil, evalErr
				}
				text := fmt.Sprintf("%s %s x=%s y=%s -> %s",
					label, op, tier1SentinelValueText(width, x), tier1SentinelValueText(width, y), result)
				comment := fmt.Sprintf("%s %s %s , %s = %s , flags %s",
					label, op, tier1SentinelDecimalComment(width, x), tier1SentinelDecimalComment(width, y),
					tier1SentinelResultComment(width, result), tier1SentinelFlagsComment(result))
				if err := appendRow(text, comment); err != nil {
					return nil, err
				}
			}
		}
		toIntPool := tier1SentinelCCToIntPool()
		for opIndex, op := range toIntOps {
			adopted, selErr := tier1SentinelCCSelectUnaryFamily("to-int", label, width, toIntKeys, opIndex,
				toIntPool, tier1SentinelCCToIntCap, exceptionUse, func(index int, x bid128BidCodecValue) (string, error) {
					return tier1SentinelCCEvalToInt(width, toIntOps[index], x)
				})
			if selErr != nil {
				return nil, selErr
			}
			for _, operand := range adopted {
				x, encErr := operand.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				result, evalErr := tier1SentinelCCEvalToInt(width, op, x)
				if evalErr != nil {
					return nil, evalErr
				}
				text := fmt.Sprintf("%s %s x=%s -> %s",
					label, op.label(), tier1SentinelValueText(width, x), result)
				comment := fmt.Sprintf("%s %s %s = %s , flags %s",
					label, op.label(), tier1SentinelDecimalComment(width, x),
					strings.SplitN(result, "/", 2)[0], tier1SentinelFlagsComment(result))
				if err := appendRow(text, comment); err != nil {
					return nil, err
				}
			}
		}
		widthOps := tier1SentinelCCWidthOps(width)
		widthKeys := make([]string, len(widthOps))
		for i, op := range widthOps {
			widthKeys[i] = op.key()
		}
		widthPool := tier1SentinelCCWidthPool(width)
		for opIndex, op := range widthOps {
			adopted, selErr := tier1SentinelCCSelectUnaryFamily("width", label, width, widthKeys, opIndex,
				widthPool, tier1SentinelCCWidthCap, exceptionUse, func(index int, x bid128BidCodecValue) (string, error) {
					return tier1SentinelCCEvalWidth(width, widthOps[index], x)
				})
			if selErr != nil {
				return nil, selErr
			}
			for _, operand := range adopted {
				x, encErr := operand.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				result, evalErr := tier1SentinelCCEvalWidth(width, op, x)
				if evalErr != nil {
					return nil, evalErr
				}
				var text string
				if op.rounded {
					text = fmt.Sprintf("%s %s x=%s m=%d -> %s",
						label, op.label(), tier1SentinelValueText(width, x), op.mode.native, result)
				} else {
					text = fmt.Sprintf("%s %s x=%s -> %s",
						label, op.label(), tier1SentinelValueText(width, x), result)
				}
				modeName := "(none)"
				if op.rounded {
					modeName = op.mode.name
				}
				comment := fmt.Sprintf("%s %s %s [%s] = %s , flags %s",
					label, op.label(), tier1SentinelDecimalComment(width, x), modeName,
					tier1SentinelResultComment(op.dest, result), tier1SentinelFlagsComment(result))
				if err := appendRow(text, comment); err != nil {
					return nil, err
				}
			}
		}
		binaryPool := tier1SentinelCCBinaryPool()
		for opIndex, op := range binaryOps {
			adopted, selErr := tier1SentinelCCSelectUnaryFamily("binary", label, width, binaryKeys, opIndex,
				binaryPool, tier1SentinelCCBinaryCap, exceptionUse, func(index int, x bid128BidCodecValue) (string, error) {
					return tier1SentinelCCEvalBinary(width, binaryOps[index], x)
				})
			if selErr != nil {
				return nil, selErr
			}
			for _, operand := range adopted {
				x, encErr := operand.encode(width)
				if encErr != nil {
					return nil, encErr
				}
				result, evalErr := tier1SentinelCCEvalBinary(width, op, x)
				if evalErr != nil {
					return nil, evalErr
				}
				text := fmt.Sprintf("%s %s x=%s m=%d -> %s",
					label, op.label(), tier1SentinelValueText(width, x), op.mode.native, result)
				comment := fmt.Sprintf("%s %s %s [%s] = binary bits %s , flags %s",
					label, op.label(), tier1SentinelDecimalComment(width, x), op.mode.name,
					strings.SplitN(result, "/", 2)[0], tier1SentinelFlagsComment(result))
				if err := appendRow(text, comment); err != nil {
					return nil, err
				}
			}
		}
		scopeLabel := label
		for opIndex, op := range constructorOps {
			if op.dest != width {
				continue
			}
			adopted, selErr := tier1SentinelCCSelectConstructor(scopeLabel, constructorOps, opIndex, exceptionUse)
			if selErr != nil {
				return nil, selErr
			}
			for _, raw := range adopted {
				result, evalErr := tier1SentinelCCEvalConstructor(op, raw)
				if evalErr != nil {
					return nil, evalErr
				}
				var text string
				if op.rounded {
					text = fmt.Sprintf("%s %s i=%s m=%d -> %s",
						label, op.label(), tier1SentinelCCConstructorInput(op, raw), op.mode.native, result)
				} else {
					text = fmt.Sprintf("%s %s i=%s -> %s",
						label, op.label(), tier1SentinelCCConstructorInput(op, raw), result)
				}
				modeName := "(none)"
				if op.rounded {
					modeName = op.mode.name
				}
				comment := fmt.Sprintf("%s %s %s [%s] = %s , flags %s",
					label, op.label(), tier1SentinelCCConstructorInput(op, raw), modeName,
					tier1SentinelResultComment(op.dest, result+ccExactFlagsSuffix(result)), tier1SentinelFlagsComment(result+ccExactFlagsSuffix(result)))
				if err := appendRow(text, comment); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := exceptionUse.requireAllUsed(); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("tier1 sentinel cc selection produced zero rows")
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if seen[row.text] {
			return nil, fmt.Errorf("tier1 sentinel cc selection produced duplicate row %q", row.text)
		}
		seen[row.text] = true
	}
	if err := tier1SentinelOracleErr(); err != nil {
		return nil, err
	}
	return rows, nil
}

// ccExactFlagsSuffix lets the audit-comment helpers reuse the "<bits>/<flags>"
// parsers on exact constructor results that carry no flags channel.
func ccExactFlagsSuffix(result string) string {
	if strings.Contains(result, "/") {
		return ""
	}
	return "/00000000"
}
