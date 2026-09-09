package testgen

import (
	"strconv"
	"strings"
)

func generatedDectestSkipReason(suite GeneratedDectestSuite, tc parsedCase) (string, bool) {
	if generatedDectestIgnoredOperation(suite.IgnoredOperations, tc.Operation) {
		return "ignored_operation_" + normalizeDecTestOperation(tc.Operation), true
	}
	if reason, ok := generatedDectestCaseSkipReason(tc, suite.TestType); ok {
		return reason, true
	}
	return "", false
}

// generatedDectestGoportSkipReason is the generator-side mirror of the runtime
// dectestGoportSkipReason emitted into the generated goport runner. It classifies
// every fixed-width-suite case of the portable Go mechanical-port leg into either
// an executed oracle-op case or one mechanical skip bucket, so the leg's verification is
// a closed accounting (executed = cases - sum(skip_reasons)). The buckets are:
//   - ignored_operation_apply: the suite's ignored decTest operations
//   - adapter_operation_out_of_leg: decTest ops outside this leg's oracle-dispatch
//     set (already covered by the native Go-port decTest leg)
//   - tagged_literal: '#' DPD/encode tagged-literal operands or results
//   - unsupported_rounding: up/half_down/05up (no Intel BID rounding mode) on a
//     result-rounding oracle op
//   - compare_nan_operand: compare/comparesig with a NaN operand, whose GDA
//     NaN-identity result the Intel BID boolean compare predicates cannot produce
//   - binary_op_nan_payload_precedence / fma_nan_payload_precedence: GDA and the
//     pinned Intel BID port select different NaN identities
//   - remainder(near)_gda_division_impossible_context_semantics: GDA's
//     Division_impossible context rule has no Intel BID counterpart
//   - minmax_equal_operand_cohort_unspecified: min/max over numerically equal
//     operands that are different cohort members, where IEEE leaves the returned
//     member unspecified and GDA pins one
//   - abs_nan_operand_gda_propagation: GDA abs propagates the NaN operand's sign
//     and quietizes sNaN; IEEE/Intel abs clears the sign quietly
//   - scaleb_non_integer_exponent_operand / scaleb_exponent_out_of_gda_range: the
//     GDA-only second-operand channel and range rule of scaleb
func generatedDectestGoportSkipReason(suite GeneratedDectestSuite, tc parsedCase) (string, bool) {
	op := normalizeDecTestOperation(tc.Operation)
	if generatedDectestIgnoredOperation(suite.IgnoredOperations, tc.Operation) {
		return "ignored_operation_" + op, true
	}
	if !generatedDectestGoportOracleOperation(op) {
		return "adapter_operation_out_of_leg", true
	}
	if generatedDectestUsesTaggedLiteral(tc) {
		return "tagged_literal", true
	}
	// A Conversion_syntax expected flag marks input the decNumber oracle grammar
	// rejects as malformed (mapping it to a NaN). The Intel BID port's from_string
	// follows the Intel grammar, which is verified separately against Intel in the
	// readtest string domain, and parses some of these inputs to a value or a
	// different NaN. That grammar difference is an Intel-vs-decNumber divergence, not
	// a port defect, so these cases are recorded skips here.
	if generatedDectestHasFlag(tc.Flags, "conversionsyntax") {
		return "conversion_syntax_divergence", true
	}
	// Every oracle op is rounding-sensitive here: decTest rounds the operand to
	// the format precision at the context rounding on parse (see the decTest file
	// header note), and add/subtract/multiply/divide/quantize/tointegral(x) round
	// the result too. The Intel BID port exposes only the 5 IEEE rounding modes, so
	// up/half_down/05up cases are recorded skips for the whole leg.
	if !generatedDectestBIDRoundingMode(tc.RoundingMode) {
		return "unsupported_rounding", true
	}
	if generatedDectestGoportNaNPrecedenceOp(op) && generatedDectestGoportNaNPayloadPrecedence(tc) {
		return "binary_op_nan_payload_precedence", true
	}
	if op == "fma" && generatedDectestGoportFMANaNPayloadPrecedenceCase(tc, suite.TestType) {
		return "fma_nan_payload_precedence", true
	}
	if (op == "compare" || op == "comparesig") && generatedDectestGoportCompareHasNaNOperand(tc) {
		return "compare_nan_operand", true
	}
	if reason, ok := generatedDectestGoportRemainderFamilyReason(tc, op); ok {
		return reason, true
	}
	if generatedDectestGoportMinMaxEqualOperandCohortCase(tc) {
		return "minmax_equal_operand_cohort_unspecified", true
	}
	if generatedDectestGoportAbsNaNOperandCase(tc) {
		return "abs_nan_operand_gda_propagation", true
	}
	if reason, ok := generatedDectestGoportScaleBReason(tc); ok {
		return reason, true
	}
	return "", false
}

// generatedDectestGoportRemainderFamilyReason reuses the native leg's
// division-impossible classifier for the goport leg's remainder/remainderNear
// routes. The native leg's sibling NaN-identity bucket is not consulted here:
// remainder and remainderNear are in this leg's
// generatedDectestGoportNaNPrecedenceOp set, so a left-quiet/right-signaling
// NaN case is already counted into binary_op_nan_payload_precedence before this
// runs, keeping one NaN-precedence bucket name per leg.
//
// Boundary note -- this is the ONE goport skip classifier that keys on the
// case's expected Conditions and expected result rather than on operand shapes
// alone, and that is deliberate:
//   - Division_impossible is not an observation of what the port produced. It
//     is the GDA oracle's declaration of WHICH RESULT CHANNEL the case exercises:
//     "the integer quotient needed here exceeds the context precision, so GDA
//     answers on its context-error channel instead of returning a remainder".
//     Intel BID's fmod/rem have no such channel at all, so a case on it cannot
//     be a port failure this skip could swallow -- there is no port behavior for
//     it to be wrong about.
//   - Deriving the same region from the operands alone means reimplementing the
//     GDA integer-quotient digit-count rule in all three mirrors (generator, Go
//     runner, Rust runner). That is a worse trade: three copies of real
//     arithmetic that can drift from each other, replacing a one-line read of
//     the oracle's own channel marker.
//
// Every OTHER goport skip classifier is operand-shape-only on purpose; do not
// use this one as precedent for keying a new class on expected results.
func generatedDectestGoportRemainderFamilyReason(tc parsedCase, op string) (string, bool) {
	if op != "remainder" && op != "remaindernear" {
		return "", false
	}
	if !generatedDectestHasOnlyFiniteOperands(tc, 2) {
		return "", false
	}
	if generatedDectestHasOnlyCondition(tc.Flags, "divisionimpossible") && generatedDectestDefaultQuietNaN(tc.Result) {
		return op + "_gda_division_impossible_context_semantics", true
	}
	return "", false
}

// generatedDectestGoportMinMaxEqualOperandCohortCase reports a min/max/minmag/
// maxmag case whose two finite operands are NUMERICALLY EQUAL (zeros of either
// sign equal) yet are different cohort members -- they differ in sign of zero,
// in exponent, or in both. IEEE 754-2019 5.3.1 leaves the returned operand
// unspecified in exactly that region ("otherwise it is either x or y"), so the
// pinned Intel BID min/max selection and the GDA rule decTest pins are both
// conforming and may disagree. The class is keyed only on the operand shapes,
// never on which operand either library actually returns, so it cannot adapt to
// an implementation change.
//
// The *mag forms use the SAME numeric-equality tie test, not magnitude
// equality. minNumMag/maxNumMag are specified as: compare |x| against |y|, and
// on a magnitude tie fall through to minNum/maxNum on the SIGNED operands. That
// fallthrough is fully determined whenever the signed operands differ -- e.g.
// minmag(-1, 1) falls through to minNum(-1, 1) = -1, which decTest pins as a
// comparable case (minmag.decTest / maxmag.decTest). Only when the signed
// operands are themselves numerically equal does the fallthrough land in
// minNum/maxNum's own unspecified tie region.
func generatedDectestGoportMinMaxEqualOperandCohortCase(tc parsedCase) bool {
	op := normalizeDecTestOperation(tc.Operation)
	if !generatedDectestMinMaxOperation(op) || len(tc.Operands) != 2 {
		return false
	}
	left, leftOK := generatedDectestNumericLiteral(tc.Operands[0])
	right, rightOK := generatedDectestNumericLiteral(tc.Operands[1])
	if !leftOK || !rightOK {
		return false
	}
	if !generatedDectestNumericValueEqual(left, right) {
		return false
	}
	return !generatedDectestSameCohortMember(left, right)
}

// generatedDectestGoportAbsNaNOperandCase reports an abs case whose operand is a
// NaN with a negative sign or a signaling NaN. decTest's abs is the GDA
// arithmetic abs: it propagates the operand NaN under the general rules, keeping
// the NaN's own sign, and quietizes a signaling NaN while signaling
// Invalid_operation. The port routes abs through Intel bid*_abs, the IEEE
// 754-2019 5.5.1 quiet sign operation, which clears the sign bit of every
// operand including NaNs and leaves a signaling NaN signaling without raising a
// flag. Positive quiet NaN operands agree and stay executed.
func generatedDectestGoportAbsNaNOperandCase(tc parsedCase) bool {
	if normalizeDecTestOperation(tc.Operation) != "abs" || len(tc.Operands) != 1 {
		return false
	}
	info := generatedDectestParseNaNOperand(tc.Operands[0])
	if !info.isNaN {
		return false
	}
	return info.signaling || info.sign == "-"
}

// generatedDectestGoportScaleBReason classifies the two GDA-only surfaces of
// scaleb, whose port route is Intel bid*_scalbln with a machine-integer
// exponent parameter:
//   - scaleb_non_integer_exponent_operand: the decTest second operand is not a
//     zero-exponent integer literal (a fractional value, an exponent-notation
//     literal, an infinity, or a NaN). GDA answers these from its own operand
//     grammar -- NaN propagation, or Invalid_operation for a non-integer -- while
//     the port's parameter cannot carry that operand at all, so the whole channel
//     is outside the port operation rather than a value the port gets wrong.
//   - scaleb_exponent_out_of_gda_range: |n| exceeds the GDA context limit
//     2 * (maxExponent + precision), where GDA returns NaN Invalid_operation.
//     Intel bid*_scalbln applies no context range rule and scales to infinity or
//     to zero. A NaN first operand short-circuits in both libraries, so it stays
//     executed.
func generatedDectestGoportScaleBReason(tc parsedCase) (string, bool) {
	if normalizeDecTestOperation(tc.Operation) != "scaleb" || len(tc.Operands) != 2 {
		return "", false
	}
	exponent, ok := generatedDectestScaleBExponentLiteral(tc.Operands[1])
	if !ok {
		return "scaleb_non_integer_exponent_operand", true
	}
	if generatedDectestParseNaNOperand(tc.Operands[0]).isNaN {
		return "", false
	}
	limit := 2 * (tc.MaxExponent + tc.Precision)
	if exponent > limit || exponent < -limit {
		return "scaleb_exponent_out_of_gda_range", true
	}
	return "", false
}

// generatedDectestScaleBExponentLiteral accepts exactly the operand shape the
// port's integer exponent parameter can carry: an optionally signed run of
// decimal digits with no fraction part and no exponent part, fitting an int.
func generatedDectestScaleBExponentLiteral(input string) (int, bool) {
	trimmed := strings.TrimSpace(generatedDectestOperandString(input))
	if trimmed == "" {
		return 0, false
	}
	digits := trimmed
	if digits[0] == '+' || digits[0] == '-' {
		digits = digits[1:]
	}
	if digits == "" {
		return 0, false
	}
	for _, r := range digits {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, false
	}
	return value, true
}

// generatedDectestNumericValue is a finite decTest literal decomposed into its
// sign, its coefficient with trailing zeros stripped, and the matching exponent,
// plus the raw exponent for a zero (whose coefficient carries no information).
type generatedDectestNumericValue struct {
	sign     int
	coeff    string
	exponent int
	isZero   bool
	rawCoeff string
	rawExp   int
}

// generatedDectestNumericLiteral decomposes a finite decTest numeric literal.
// Infinities, NaNs, tagged literals, and malformed tokens return ok=false.
func generatedDectestNumericLiteral(input string) (generatedDectestNumericValue, bool) {
	trimmed := strings.TrimSpace(generatedDectestOperandString(input))
	if trimmed == "" || trimmed == "#" {
		return generatedDectestNumericValue{}, false
	}
	sign := 1
	switch trimmed[0] {
	case '+':
		trimmed = trimmed[1:]
	case '-':
		sign = -1
		trimmed = trimmed[1:]
	}
	if trimmed == "" {
		return generatedDectestNumericValue{}, false
	}
	lower := strings.ToLower(trimmed)
	if lower == "inf" || lower == "infinity" || strings.Contains(lower, "nan") {
		return generatedDectestNumericValue{}, false
	}
	mantissa := trimmed
	exponent := 0
	if idx := strings.IndexAny(trimmed, "Ee"); idx >= 0 {
		mantissa = trimmed[:idx]
		parsed, err := strconv.Atoi(trimmed[idx+1:])
		if err != nil {
			return generatedDectestNumericValue{}, false
		}
		exponent = parsed
	}
	if strings.Count(mantissa, ".") > 1 {
		return generatedDectestNumericValue{}, false
	}
	if dot := strings.IndexByte(mantissa, '.'); dot >= 0 {
		exponent -= len(mantissa) - dot - 1
		mantissa = mantissa[:dot] + mantissa[dot+1:]
	}
	if mantissa == "" {
		return generatedDectestNumericValue{}, false
	}
	for _, r := range mantissa {
		if r < '0' || r > '9' {
			return generatedDectestNumericValue{}, false
		}
	}
	rawCoeff := strings.TrimLeft(mantissa, "0")
	if rawCoeff == "" {
		rawCoeff = "0"
	}
	value := generatedDectestNumericValue{sign: sign, rawCoeff: rawCoeff, rawExp: exponent, exponent: exponent}
	if rawCoeff == "0" {
		value.isZero = true
		value.coeff = "0"
		return value, true
	}
	coeff := rawCoeff
	for strings.HasSuffix(coeff, "0") {
		coeff = coeff[:len(coeff)-1]
		value.exponent++
	}
	value.coeff = coeff
	return value, true
}

func generatedDectestNumericValueEqual(left, right generatedDectestNumericValue) bool {
	if left.isZero || right.isZero {
		return left.isZero && right.isZero
	}
	return left.sign == right.sign && left.coeff == right.coeff && left.exponent == right.exponent
}

// generatedDectestGoportFMANaNPayloadPrecedenceCase is the goport leg's
// OPERAND-ONLY fma NaN-identity divergence test. GDA fma propagation selects the
// first signaling NaN in operand order x, y, z and otherwise the first quiet
// NaN; the pinned Intel BID port propagates the first NaN it unpacks in y, z, x
// order (bid-go fma64 / bid128_fma unpack order). Both libraries quietize the
// selected NaN, so the two results differ exactly when the two selected
// operands carry different quietized identities (sign plus payload) -- decidable
// from the operands alone.
//
// It deliberately does NOT consult tc.Result or tc.Flags. The native leg's
// generatedDectestFMANaNPayloadPrecedenceCase does (it additionally requires the
// expected result to equal the GDA selection and the Conditions to match), and
// that stays untouched here: this leg's rule is that a skip class must be
// decidable from the case INPUT so it can never absorb a wrong port answer.
//
// testType is still consulted, for the operand payload width only: a payload
// wider than the format's NaN payload field is not a NaN identity either
// library can carry, so such a case is not classified.
func generatedDectestGoportFMANaNPayloadPrecedenceCase(tc parsedCase, testType string) bool {
	if len(tc.Operands) != 3 {
		return false
	}
	var infos [3]generatedDectestNaNOperand
	for i := range tc.Operands {
		infos[i] = generatedDectestParseNaNOperand(tc.Operands[i])
		if infos[i].isNaN {
			if !generatedDectestNaNPayloadFitsType(infos[i], testType) {
				return false
			}
			continue
		}
		if !generatedDectestFiniteValue(tc.Operands[i]) && !generatedDectestInfinity(tc.Operands[i]) {
			return false
		}
	}
	gda := -1
	for i := range infos {
		if infos[i].isNaN && infos[i].signaling {
			gda = i
			break
		}
	}
	if gda < 0 {
		for i := range infos {
			if infos[i].isNaN {
				gda = i
				break
			}
		}
	}
	if gda < 0 {
		return false
	}
	intel := -1
	for _, i := range [3]int{1, 2, 0} {
		if infos[i].isNaN {
			intel = i
			break
		}
	}
	if intel < 0 {
		return false
	}
	return infos[gda].sign != infos[intel].sign || infos[gda].payload != infos[intel].payload
}

// generatedDectestSameCohortMember reports whether two finite literals name the
// identical cohort member: same sign (a negative zero stays distinct from a
// positive zero), same integer coefficient, same exponent.
func generatedDectestSameCohortMember(left, right generatedDectestNumericValue) bool {
	return left.sign == right.sign && left.rawCoeff == right.rawCoeff && left.rawExp == right.rawExp
}

// generatedDectestGoportFlagExemptReason is the generator-side mirror of the
// runtime dectestGoportFlagExemptReason emitted into the generated goport
// runner: executed cases whose expected flags are not compared (value and
// quantum still are), each a documented decNumber-vs-Intel-BID semantic
// divergence. It must stay in lockstep with the runtime function so the
// pinned FlagExempt buckets match the live recount.
//
//   - from_string_zero_low_clamp_divergence: a string-conversion (tosci/toeng)
//     case whose zero operand carries an exponent below the format minimum,
//     clamped upward on parse. decNumber raises only Clamped (projecting to
//     None on the five-flag surface) while the Intel BID from_string path the
//     port mechanically reproduces raises Inexact|Underflow for the same low
//     clamp (measured). The measured complement — high-side zero clamps and
//     every arithmetic-result zero clamp (divide/multiply) — matched
//     decNumber exactly, so the class is keyed off a tosci/toeng op,
//     clamped-only Conditions, and a zero expected result with a negative
//     exponent.
//
// Strict: a case carrying any condition token outside the recognized
// decTest set is never classified as exempt — the generated runner validates
// the expected Conditions before consulting its mirror of this classifier and
// fails the harness on an unrecognized token, and this side refuses the
// classification so the pinned exemption buckets cannot absorb such a case.
func generatedDectestGoportFlagExemptReason(tc parsedCase) (string, bool) {
	if !generatedDectestGoportRecognizedConditions(tc.Flags) {
		return "", false
	}
	op := normalizeDecTestOperation(tc.Operation)
	if op != "tosci" && op != "toeng" {
		return "", false
	}
	if !generatedDectestGoportHasOnlyClampedCondition(tc.Flags) {
		return "", false
	}
	if generatedDectestZeroResultLowExponent(tc.Result) {
		return "from_string_zero_low_clamp_divergence", true
	}
	return "", false
}

// generatedDectestGoportRecognizedConditions reports whether every condition
// token is in the decTest condition set the generated runner's
// parseDecTestFlags mapping recognizes. The token list must stay in lockstep
// with parseDecTestFlags in bid754-go/dectest_driver.go; an unrecognized
// token fails the generated runner as a harness failure, so this side must
// never count such a case into an exemption bucket.
func generatedDectestGoportRecognizedConditions(flags []string) bool {
	for _, flag := range flags {
		switch generatedDectestNormalizeFlag(flag) {
		case "", "none", "noflags", "inexact", "underflow", "overflow", "divisionbyzero",
			"invalidoperation", "divisionundefined", "divisionimpossible", "insufficientstorage",
			"conversionsyntax", "subnormal", "rounded", "clamped":
		default:
			return false
		}
	}
	return true
}

// generatedDectestZeroResultLowExponent reports whether an expected result
// literal is a zero with a negative exponent (the low-side clamp shape, e.g.
// 0E-398 or -0E-6176), mirroring the generated runner's quantum decomposition
// of the same literal.
func generatedDectestZeroResultLowExponent(result string) bool {
	trimmed := strings.TrimSpace(strings.Trim(result, "'\""))
	if trimmed == "" {
		return false
	}
	switch trimmed[0] {
	case '+', '-':
		trimmed = trimmed[1:]
	}
	if trimmed == "" {
		return false
	}
	mantissa := trimmed
	exponent := 0
	if idx := strings.IndexAny(trimmed, "Ee"); idx >= 0 {
		mantissa = trimmed[:idx]
		parsed, err := strconv.Atoi(trimmed[idx+1:])
		if err != nil {
			return false
		}
		exponent = parsed
	}
	if dot := strings.IndexByte(mantissa, '.'); dot >= 0 {
		exponent -= len(mantissa) - dot - 1
		mantissa = mantissa[:dot] + mantissa[dot+1:]
	}
	if mantissa == "" {
		return false
	}
	for _, r := range mantissa {
		if r != '0' {
			return false
		}
	}
	return exponent < 0
}

// generatedDectestGoportOracleOperation is the goport leg's oracle-dispatch set:
// every decTest operation the Go mechanical port has a routing target for. An op
// outside it is counted as adapter_operation_out_of_leg. Keep it in lockstep with
// the runtime dectestGoportOracleOperation and the Rust
// dectest_goport_oracle_operation.
func generatedDectestGoportOracleOperation(op string) bool {
	switch op {
	case "add", "subtract", "multiply", "divide", "quantize",
		"compare", "comparesig", "tosci", "toeng", "tointegral", "tointegralx",
		"abs", "plus", "minus", "copy", "copyabs", "copynegate", "copysign",
		"class", "samequantum", "comparetotal", "comparetotmag",
		"min", "max", "minmag", "maxmag", "logb", "scaleb",
		"nextplus", "nextminus", "nexttoward", "fma", "remainder", "remaindernear":
		return true
	default:
		return false
	}
}

// generatedDectestGoportNaNPrecedenceOp lists the binary oracle ops whose NaN
// identity selection can diverge between GDA (signaling NaN first) and the
// positional propagation of the pinned Intel BID port. scaleb is excluded: its
// second operand is not a decimal operand on the port route at all, so its NaN
// shapes are classified by generatedDectestGoportScaleBReason instead.
func generatedDectestGoportNaNPrecedenceOp(op string) bool {
	switch op {
	case "add", "subtract", "multiply", "divide", "quantize",
		"min", "max", "minmag", "maxmag",
		"remainder", "remaindernear", "nexttoward":
		return true
	default:
		return false
	}
}

// generatedDectestGoportNaNPayloadPrecedence reports binary-op cases where a
// quiet-NaN left operand meets a signaling-NaN right operand. decNumber (GDA)
// propagates the signaling NaN first, while the Intel BID port propagates operands
// in positional order and returns the left quiet NaN, so the two libraries pick
// different NaN identities. This mirrors the documented remainder/fma NaN-payload
// precedence skip and is an Intel-vs-IBM NaN-identity divergence, not a port defect.
func generatedDectestGoportNaNPayloadPrecedence(tc parsedCase) bool {
	if len(tc.Operands) != 2 {
		return false
	}
	if !generatedDectestQuietNaN(tc.Operands[0]) || !generatedDectestSignalingNaN(tc.Operands[1]) {
		return false
	}
	// Skip only genuine divergences: decNumber returns the quietized right sNaN
	// while the port returns the left qNaN, so same-sign, same-payload pairs still
	// agree and must execute (mirrors the precise fma NaN-precedence check).
	left := generatedDectestParseNaNOperand(tc.Operands[0])
	right := generatedDectestParseNaNOperand(tc.Operands[1])
	return left.sign != right.sign || left.payload != right.payload
}

func generatedDectestGoportCompareHasNaNOperand(tc parsedCase) bool {
	for _, operand := range tc.Operands {
		if generatedDectestQuietNaN(operand) || generatedDectestSignalingNaN(operand) {
			return true
		}
	}
	return false
}

// generatedDectestGoportHasOnlyClampedCondition closes the exemption over the
// exact Conditions shape registered by the specification. Empty/no-flags
// aliases contribute no condition; every non-empty condition must be Clamped,
// and at least one Clamped token must be present.
func generatedDectestGoportHasOnlyClampedCondition(flags []string) bool {
	clamped := false
	for _, flag := range flags {
		switch generatedDectestNormalizeFlag(flag) {
		case "", "none", "noflags":
			continue
		case "clamped":
			clamped = true
		default:
			return false
		}
	}
	return clamped
}

func generatedDectestCaseSkipReason(tc parsedCase, testType string) (string, bool) {
	if (testType == "general" || testType == "decimal128") && generatedDectestUsesTaggedLiteral(tc) {
		return "tagged_literal", true
	}
	if generatedDectestTaggedToIntegralCase(tc) {
		return "tagged_to_integral", true
	}
	if generatedDectestNextTowardNaNPayloadPrecedenceCase(tc, testType) {
		return "nexttoward_nan_payload_precedence", true
	}
	if generatedDectestMinMaxZeroTieCase(tc, testType) {
		return "minmax_zero_tie", true
	}
	if generatedDectestMinMaxNaNPayloadPrecedenceCase(tc, testType) {
		return "minmax_nan_payload_precedence", true
	}
	if reason, ok := generatedDectestFMAReason(tc, testType); ok {
		return reason, true
	}
	if reason, ok := generatedDectestRemainderFamilyReason(tc, "remainder", testType); ok {
		return reason, true
	}
	if reason, ok := generatedDectestRemainderFamilyReason(tc, "remaindernear", testType); ok {
		return reason, true
	}

	switch testType {
	case "decimal32":
		if tc.Precision > 7 {
			return "precision_over_decimal32", true
		}
	case "decimal64":
		if tc.Precision > 16 {
			return "precision_over_decimal64", true
		}
	case "decimal128":
		if tc.Precision > 34 {
			return "precision_over_decimal128", true
		}
	case "general":
		if tc.Precision > 34 {
			return "precision_over_general", true
		}
	default:
		return "unsupported_test_type", true
	}
	return "", false
}

func generatedDectestFMAReason(tc parsedCase, testType string) (string, bool) {
	if normalizeDecTestOperation(tc.Operation) != "fma" {
		return "", false
	}
	if !generatedDectestBIDRoundingMode(tc.RoundingMode) {
		return "fma_unsupported_rounding", true
	}
	if generatedDectestFMANaNPayloadPrecedenceCase(tc, testType) {
		return "fma_nan_payload_precedence", true
	}
	return "", false
}

func generatedDectestRemainderFamilyReason(tc parsedCase, operation, testType string) (string, bool) {
	if normalizeDecTestOperation(tc.Operation) != operation {
		return "", false
	}
	if generatedDectestIntelLeftNaNGDARightNaNIdentityDivergence(tc, testType) {
		return operation + "_nan_payload_precedence", true
	}
	if !generatedDectestHasOnlyFiniteOperands(tc, 2) {
		return "", false
	}
	if generatedDectestHasOnlyCondition(tc.Flags, "divisionimpossible") && generatedDectestDefaultQuietNaN(tc.Result) {
		return operation + "_gda_division_impossible_context_semantics", true
	}
	return "", false
}

func generatedDectestHasOnlyFiniteOperands(tc parsedCase, count int) bool {
	if len(tc.Operands) != count {
		return false
	}
	for _, operand := range tc.Operands {
		if !generatedDectestFiniteValue(operand) {
			return false
		}
	}
	return true
}

func generatedDectestNextTowardNaNPayloadPrecedenceCase(tc parsedCase, testType string) bool {
	if normalizeDecTestOperation(tc.Operation) != "nexttoward" || len(tc.Operands) != 2 {
		return false
	}
	return generatedDectestIntelLeftNaNGDARightNaNIdentityDivergence(tc, testType)
}

func generatedDectestIntelLeftNaNGDARightNaNIdentityDivergence(tc parsedCase, testType string) bool {
	if len(tc.Operands) != 2 {
		return false
	}
	left := generatedDectestParseNaNOperand(tc.Operands[0])
	right := generatedDectestParseNaNOperand(tc.Operands[1])
	expected := generatedDectestParseNaNOperand(tc.Result)
	if !left.isNaN || left.signaling || !right.isNaN || !right.signaling ||
		!expected.isNaN || expected.signaling ||
		!generatedDectestNaNPayloadFitsType(left, testType) ||
		!generatedDectestNaNPayloadFitsType(right, testType) ||
		!generatedDectestNaNPayloadFitsType(expected, testType) ||
		!generatedDectestHasOnlyInvalidOperationCondition(tc.Flags) {
		return false
	}
	// GDA quietizes the right signaling NaN. These pinned Intel BID operations
	// return the left quiet NaN after signaling invalid. Skip only when the
	// authoritative GDA result and Intel identity genuinely differ.
	if expected.sign != right.sign || expected.payload != right.payload {
		return false
	}
	return left.sign != expected.sign || left.payload != expected.payload
}

func generatedDectestMinMaxZeroTieCase(tc parsedCase, testType string) bool {
	if !generatedDectestMinMaxOperation(tc.Operation) || len(tc.Operands) != 2 ||
		!generatedDectestHasOnlyConditions(tc.Flags) {
		return false
	}
	leftSign, leftZero := generatedDectestFiniteZeroSign(tc.Operands[0])
	rightSign, rightZero := generatedDectestFiniteZeroSign(tc.Operands[1])
	if !leftZero || !rightZero || leftSign == rightSign {
		return false
	}

	selected, ok := generatedDectestIntelMinMaxZeroTieSelectedOperand(tc.Operation, testType)
	if !ok {
		return false
	}
	expectedSign, expectedZero := generatedDectestFiniteZeroSign(tc.Result)
	if !expectedZero {
		return false
	}
	operandSigns := [2]int{leftSign, rightSign}
	return operandSigns[selected] != expectedSign
}

func generatedDectestIntelMinMaxZeroTieSelectedOperand(operation, testType string) (int, bool) {
	// Pinned Intel BID C returns a fixed operand when both coefficients are
	// zero. Keep this source-derived rule independent of generated runtime code
	// so the verification cannot adapt itself to an implementation regression.
	switch testType {
	case "decimal32", "decimal64":
		switch normalizeDecTestOperation(operation) {
		case "min", "max", "maxmag":
			return 1, true
		case "minmag":
			return 0, true
		}
	case "decimal128":
		switch normalizeDecTestOperation(operation) {
		case "min", "max", "minmag":
			return 0, true
		case "maxmag":
			return 1, true
		}
	}
	return 0, false
}

func generatedDectestMinMaxNaNPayloadPrecedenceCase(tc parsedCase, testType string) bool {
	if !generatedDectestMinMaxOperation(tc.Operation) || len(tc.Operands) != 2 {
		return false
	}
	return generatedDectestIntelLeftNaNGDARightNaNIdentityDivergence(tc, testType)
}

func generatedDectestNaNPayloadFitsType(info generatedDectestNaNOperand, testType string) bool {
	var maxDigits int
	switch testType {
	case "decimal32":
		maxDigits = 6
	case "decimal64":
		maxDigits = 15
	case "decimal128":
		maxDigits = 33
	default:
		return false
	}
	return len(info.payload) <= maxDigits
}

func generatedDectestHasOnlyInvalidOperationCondition(flags []string) bool {
	return generatedDectestHasOnlyCondition(flags, "invalidoperation")
}

func generatedDectestHasOnlyCondition(flags []string, want string) bool {
	return generatedDectestHasOnlyConditions(flags, want)
}

func generatedDectestHasOnlyConditions(flags []string, wants ...string) bool {
	if len(flags) != len(wants) {
		return false
	}
	matched := make([]bool, len(wants))
	for _, flag := range flags {
		normalized := generatedDectestNormalizeFlag(flag)
		found := false
		for i, want := range wants {
			if !matched[i] && normalized == generatedDectestNormalizeFlag(want) {
				matched[i] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// generatedDectestFMANaNPayloadPrecedenceCase mirrors the generated driver's
// isUnsupportedFMANaNPayloadPrecedenceCase: fma cases where GDA decTest NaN
// propagation (first signaling NaN in operand order x, y, z; otherwise the
// first quiet NaN) selects a different NaN identity (sign plus payload) than
// the Intel BID port, which propagates the first NaN it unpacks in y, z, x
// order (bid-go fma64/bid128_fma NaN unpack order).
func generatedDectestFMANaNPayloadPrecedenceCase(tc parsedCase, testType string) bool {
	if len(tc.Operands) != 3 {
		return false
	}
	var infos [3]generatedDectestNaNOperand
	for i := range tc.Operands {
		infos[i] = generatedDectestParseNaNOperand(tc.Operands[i])
		if infos[i].isNaN {
			if !generatedDectestNaNPayloadFitsType(infos[i], testType) {
				return false
			}
			continue
		}
		if !generatedDectestFiniteValue(tc.Operands[i]) && !generatedDectestInfinity(tc.Operands[i]) {
			return false
		}
	}
	gda := -1
	for i := range infos {
		if infos[i].isNaN && infos[i].signaling {
			gda = i
			break
		}
	}
	if gda < 0 {
		for i := range infos {
			if infos[i].isNaN {
				gda = i
				break
			}
		}
	}
	if gda < 0 {
		return false
	}
	intel := -1
	for _, i := range [3]int{1, 2, 0} {
		if infos[i].isNaN {
			intel = i
			break
		}
	}
	if intel < 0 {
		return false
	}
	expected := generatedDectestParseNaNOperand(tc.Result)
	if !expected.isNaN || expected.signaling || !generatedDectestNaNPayloadFitsType(expected, testType) ||
		expected.sign != infos[gda].sign || expected.payload != infos[gda].payload {
		return false
	}
	if infos[gda].signaling {
		if !generatedDectestHasOnlyInvalidOperationCondition(tc.Flags) {
			return false
		}
	} else if !generatedDectestHasOnlyConditions(tc.Flags) {
		return false
	}
	return infos[gda].sign != infos[intel].sign || infos[gda].payload != infos[intel].payload
}

type generatedDectestNaNOperand struct {
	isNaN     bool
	signaling bool
	sign      string
	payload   string
}

// generatedDectestParseNaNOperand extracts the quietized NaN identity of a
// decTest operand literal: sign and payload digits with leading zeros
// stripped, so "-sNaN00" and "-NaN" compare equal after quietization.
func generatedDectestParseNaNOperand(input string) generatedDectestNaNOperand {
	trimmed := strings.TrimSpace(generatedDectestOperandString(input))
	sign := "+"
	if strings.HasPrefix(trimmed, "-") {
		sign = "-"
		trimmed = trimmed[1:]
	} else {
		trimmed = strings.TrimPrefix(trimmed, "+")
	}
	lower := strings.ToLower(trimmed)
	info := generatedDectestNaNOperand{sign: sign}
	var payload string
	switch {
	case strings.HasPrefix(lower, "snan"):
		info.isNaN = true
		info.signaling = true
		payload = lower[len("snan"):]
	case strings.HasPrefix(lower, "nan"):
		info.isNaN = true
		payload = lower[len("nan"):]
	default:
		return generatedDectestNaNOperand{}
	}
	for _, r := range payload {
		if r < '0' || r > '9' {
			return generatedDectestNaNOperand{}
		}
	}
	info.payload = strings.TrimLeft(payload, "0")
	return info
}

func generatedDectestMinMaxOperation(op string) bool {
	switch normalizeDecTestOperation(op) {
	case "min", "max", "minmag", "maxmag":
		return true
	default:
		return false
	}
}

func generatedDectestQuietNaN(input string) bool {
	info := generatedDectestParseNaNOperand(input)
	return info.isNaN && !info.signaling
}

func generatedDectestSignalingNaN(input string) bool {
	info := generatedDectestParseNaNOperand(input)
	return info.isNaN && info.signaling
}

func generatedDectestDefaultQuietNaN(input string) bool {
	info := generatedDectestParseNaNOperand(input)
	return info.isNaN && !info.signaling && info.sign == "+" && info.payload == ""
}

func generatedDectestInfinity(input string) bool {
	trimmed := strings.TrimSpace(generatedDectestOperandString(input))
	if trimmed == "" {
		return false
	}
	switch trimmed[0] {
	case '+', '-':
		trimmed = trimmed[1:]
	}
	lower := strings.ToLower(trimmed)
	return lower == "inf" || lower == "infinity"
}

func generatedDectestFiniteValue(input string) bool {
	trimmed := strings.TrimSpace(generatedDectestOperandString(input))
	if trimmed == "" || trimmed == "#" {
		return false
	}
	switch trimmed[0] {
	case '+', '-':
		trimmed = trimmed[1:]
	}
	if trimmed == "" {
		return false
	}

	mantissa := trimmed
	if idx := strings.IndexAny(trimmed, "Ee"); idx >= 0 {
		mantissa = trimmed[:idx]
		if _, err := strconv.Atoi(trimmed[idx+1:]); err != nil {
			return false
		}
	}
	if strings.Count(mantissa, ".") > 1 {
		return false
	}
	if dot := strings.IndexByte(mantissa, '.'); dot >= 0 {
		mantissa = mantissa[:dot] + mantissa[dot+1:]
	}
	if mantissa == "" {
		return false
	}
	for _, r := range mantissa {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func generatedDectestTaggedToIntegralCase(tc parsedCase) bool {
	op := normalizeDecTestOperation(tc.Operation)
	return (op == "tointegral" || op == "tointegralx") && generatedDectestUsesTaggedLiteral(tc)
}

func generatedDectestUsesTaggedLiteral(tc parsedCase) bool {
	for _, operand := range tc.Operands {
		if strings.Contains(operand, "#") {
			return true
		}
	}
	return strings.Contains(tc.Result, "#")
}

func generatedDectestIgnoredOperation(ignoredOperations []string, operation string) bool {
	normalized := normalizeDecTestOperation(operation)
	for _, ignored := range ignoredOperations {
		if normalizeDecTestOperation(ignored) == normalized {
			return true
		}
	}
	return false
}

func generatedDectestBIDRoundingMode(rounding string) bool {
	switch strings.ToLower(strings.TrimSpace(rounding)) {
	case "", "half_even", "half_up", "down", "ceiling", "floor":
		return true
	default:
		return false
	}
}

func generatedDectestHasFlag(flags []string, want string) bool {
	normalizedWant := generatedDectestNormalizeFlag(want)
	for _, flag := range flags {
		if generatedDectestNormalizeFlag(flag) == normalizedWant {
			return true
		}
	}
	return false
}

func generatedDectestNormalizeFlag(flag string) string {
	flag = strings.Trim(flag, "'\"")
	flag = strings.ToLower(flag)
	flag = strings.ReplaceAll(flag, "_", "")
	flag = strings.ReplaceAll(flag, "-", "")
	flag = strings.ReplaceAll(flag, " ", "")
	return flag
}

func generatedDectestFiniteZeroSign(input string) (int, bool) {
	trimmed := strings.TrimSpace(generatedDectestOperandString(input))
	if trimmed == "" || trimmed == "#" {
		return 0, false
	}
	sign := 1
	switch trimmed[0] {
	case '+':
		trimmed = trimmed[1:]
	case '-':
		sign = -1
		trimmed = trimmed[1:]
	}
	if trimmed == "" {
		return 0, false
	}
	lower := strings.ToLower(trimmed)
	if lower == "inf" || lower == "infinity" || strings.HasPrefix(lower, "nan") || strings.HasPrefix(lower, "qnan") || strings.HasPrefix(lower, "snan") {
		return 0, false
	}
	mantissa := trimmed
	if idx := strings.IndexAny(trimmed, "Ee"); idx >= 0 {
		mantissa = trimmed[:idx]
		if _, err := strconv.Atoi(trimmed[idx+1:]); err != nil {
			return 0, false
		}
	}
	if strings.Count(mantissa, ".") > 1 {
		return 0, false
	}
	if dot := strings.IndexByte(mantissa, '.'); dot >= 0 {
		mantissa = mantissa[:dot] + mantissa[dot+1:]
	}
	if mantissa == "" {
		return 0, false
	}
	for _, r := range mantissa {
		if r < '0' || r > '9' {
			return 0, false
		}
		if r != '0' {
			return 0, false
		}
	}
	return sign, true
}

func generatedDectestOperandString(input string) string {
	return strings.Trim(input, "'\"")
}
