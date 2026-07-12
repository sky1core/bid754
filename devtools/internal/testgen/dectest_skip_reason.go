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
	if (op == "compare" || op == "comparesig") && generatedDectestGoportCompareHasNaNOperand(tc) {
		return "compare_nan_operand", true
	}
	return "", false
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

func generatedDectestGoportOracleOperation(op string) bool {
	switch op {
	case "add", "subtract", "multiply", "divide", "quantize",
		"compare", "comparesig", "tosci", "toeng", "tointegral", "tointegralx":
		return true
	default:
		return false
	}
}

func generatedDectestGoportNaNPrecedenceOp(op string) bool {
	switch op {
	case "add", "subtract", "multiply", "divide", "quantize":
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
	if reason, ok := generatedDectestScaleBReason(tc); ok {
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
	if !generatedDectestHasOnlyFiniteOperands(tc, 3) || !generatedDectestFiniteValue(tc.Result) {
		return "", false
	}
	if generatedDectestHasOnlyCondition(tc.Flags, "clamped") {
		return "fma_clamped_status_gap", true
	}
	if generatedDectestHasOnlyRoundedStatusGapConditions(tc.Flags) {
		return "fma_rounded_only_status_gap", true
	}
	return "", false
}

func generatedDectestHasOnlyRoundedStatusGapConditions(flags []string) bool {
	return generatedDectestHasOnlyConditions(flags, "rounded") ||
		generatedDectestHasOnlyConditions(flags, "subnormal", "rounded")
}

func generatedDectestScaleBReason(tc parsedCase) (string, bool) {
	if normalizeDecTestOperation(tc.Operation) != "scaleb" {
		return "", false
	}
	if !generatedDectestHasOnlyFiniteOperands(tc, 2) || !generatedDectestFiniteValue(tc.Result) {
		return "", false
	}
	if generatedDectestHasOnlyCondition(tc.Flags, "clamped") {
		return "scaleb_clamped_status_gap", true
	}
	if generatedDectestHasOnlyRoundedStatusGapConditions(tc.Flags) {
		return "scaleb_rounded_only_status_gap", true
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
		return operation + "_division_impossible_status_gap", true
	}
	if generatedDectestHasOnlyCondition(tc.Flags, "clamped") && generatedDectestFiniteValue(tc.Result) {
		return operation + "_clamped_status_gap", true
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
