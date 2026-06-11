package testgen

import (
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

func generatedDectestCaseSkipReason(tc parsedCase, testType string) (string, bool) {
	if (testType == "general" || testType == "decimal128") && generatedDectestUsesTaggedLiteral(tc) {
		return "tagged_literal", true
	}
	if generatedDectestTaggedToIntegralCase(tc) {
		return "tagged_to_integral", true
	}
	if generatedDectestNextTowardNaNPayloadPrecedenceCase(tc) {
		return "nexttoward_nan_payload_precedence", true
	}
	if generatedDectestMinMaxZeroTieCase(tc) {
		return "minmax_zero_tie", true
	}
	if generatedDectestMinMaxNaNPayloadPrecedenceCase(tc) {
		return "minmax_nan_payload_precedence", true
	}
	if reason, ok := generatedDectestFMAReason(tc); ok {
		return reason, true
	}
	if reason, ok := generatedDectestScaleBReason(tc); ok {
		return reason, true
	}
	if reason, ok := generatedDectestRemainderFamilyReason(tc, "remainder"); ok {
		return reason, true
	}
	if reason, ok := generatedDectestRemainderFamilyReason(tc, "remaindernear"); ok {
		return reason, true
	}

	switch testType {
	case "decimal32":
		return "precision_over_decimal32", tc.Precision > 7
	case "decimal64":
		return "precision_over_decimal64", tc.Precision > 16
	case "decimal128":
		return "precision_over_decimal128", tc.Precision > 34
	case "general":
		return "precision_over_general", tc.Precision > 34
	default:
		return "unsupported_test_type", true
	}
}

func generatedDectestFMAReason(tc parsedCase) (string, bool) {
	if normalizeDecTestOperation(tc.Operation) != "fma" {
		return "", false
	}
	if !generatedDectestBIDRoundingMode(tc.RoundingMode) {
		return "fma_unsupported_rounding", true
	}
	if generatedDectestFMANaNPayloadPrecedenceCase(tc) {
		return "fma_nan_payload_precedence", true
	}
	if generatedDectestHasFlag(tc.Flags, "clamped") &&
		!generatedDectestHasFlag(tc.Flags, "underflow") &&
		!generatedDectestHasFlag(tc.Flags, "inexact") &&
		!generatedDectestHasFlag(tc.Flags, "overflow") {
		return "fma_clamped_status_gap", true
	}
	if generatedDectestHasFlag(tc.Flags, "rounded") && !generatedDectestHasFlag(tc.Flags, "inexact") {
		return "fma_rounded_only_status_gap", true
	}
	return "", false
}

func generatedDectestScaleBReason(tc parsedCase) (string, bool) {
	if normalizeDecTestOperation(tc.Operation) != "scaleb" {
		return "", false
	}
	if generatedDectestHasFlag(tc.Flags, "clamped") &&
		!generatedDectestHasFlag(tc.Flags, "underflow") &&
		!generatedDectestHasFlag(tc.Flags, "inexact") &&
		!generatedDectestHasFlag(tc.Flags, "overflow") {
		return "scaleb_clamped_status_gap", true
	}
	if generatedDectestHasFlag(tc.Flags, "rounded") && !generatedDectestHasFlag(tc.Flags, "inexact") {
		return "scaleb_rounded_only_status_gap", true
	}
	return "", false
}

func generatedDectestRemainderFamilyReason(tc parsedCase, operation string) (string, bool) {
	if normalizeDecTestOperation(tc.Operation) != operation {
		return "", false
	}
	if generatedDectestHasFlag(tc.Flags, "divisionimpossible") {
		return operation + "_division_impossible_status_gap", true
	}
	if generatedDectestHasFlag(tc.Flags, "clamped") {
		return operation + "_clamped_status_gap", true
	}
	if len(tc.Operands) == 2 && generatedDectestQuietNaN(tc.Operands[0]) && generatedDectestSignalingNaN(tc.Operands[1]) {
		return operation + "_nan_payload_precedence", true
	}
	return "", false
}

func generatedDectestNextTowardNaNPayloadPrecedenceCase(tc parsedCase) bool {
	if normalizeDecTestOperation(tc.Operation) != "nexttoward" || len(tc.Operands) != 2 {
		return false
	}
	left := strings.ToLower(generatedDectestOperandString(tc.Operands[0]))
	right := strings.ToLower(generatedDectestOperandString(tc.Operands[1]))
	return strings.Contains(left, "nan") && generatedDectestHasNaNPayload(left) && strings.Contains(right, "snan")
}

func generatedDectestMinMaxZeroTieCase(tc parsedCase) bool {
	if !generatedDectestMinMaxOperation(tc.Operation) || len(tc.Operands) != 2 {
		return false
	}
	leftSign, leftZero := generatedDectestFiniteZeroSign(tc.Operands[0])
	rightSign, rightZero := generatedDectestFiniteZeroSign(tc.Operands[1])
	return leftZero && rightZero && leftSign != rightSign
}

func generatedDectestMinMaxNaNPayloadPrecedenceCase(tc parsedCase) bool {
	if !generatedDectestMinMaxOperation(tc.Operation) || len(tc.Operands) != 2 {
		return false
	}
	return generatedDectestQuietNaN(tc.Operands[0]) && generatedDectestSignalingNaN(tc.Operands[1])
}

// generatedDectestFMANaNPayloadPrecedenceCase mirrors the generated driver's
// isUnsupportedFMANaNPayloadPrecedenceCase: fma cases where GDA decTest NaN
// propagation (first signaling NaN in operand order x, y, z; otherwise the
// first quiet NaN) selects a different NaN identity (sign plus payload) than
// the Intel BID port, which propagates the first NaN it unpacks in y, z, x
// order (bid-go fma64/bid128_fma NaN unpack order).
func generatedDectestFMANaNPayloadPrecedenceCase(tc parsedCase) bool {
	if len(tc.Operands) != 3 {
		return false
	}
	var infos [3]generatedDectestNaNOperand
	for i := range tc.Operands {
		infos[i] = generatedDectestParseNaNOperand(tc.Operands[i])
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
	case strings.HasPrefix(lower, "qnan"):
		info.isNaN = true
		payload = lower[len("qnan"):]
	case strings.HasPrefix(lower, "nan"):
		info.isNaN = true
		payload = lower[len("nan"):]
	default:
		return generatedDectestNaNOperand{}
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
	trimmed := strings.TrimSpace(generatedDectestOperandString(input))
	trimmed = strings.TrimPrefix(trimmed, "+")
	trimmed = strings.TrimPrefix(trimmed, "-")
	lower := strings.ToLower(trimmed)
	return strings.HasPrefix(lower, "nan") || strings.HasPrefix(lower, "qnan")
}

func generatedDectestSignalingNaN(input string) bool {
	trimmed := strings.TrimSpace(generatedDectestOperandString(input))
	trimmed = strings.TrimPrefix(trimmed, "+")
	trimmed = strings.TrimPrefix(trimmed, "-")
	return strings.HasPrefix(strings.ToLower(trimmed), "snan")
}

func generatedDectestHasNaNPayload(input string) bool {
	for _, r := range input {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
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
	if idx := strings.IndexAny(trimmed, "Ee"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	trimmed = strings.ReplaceAll(trimmed, ".", "")
	if trimmed == "" {
		return 0, false
	}
	for _, r := range trimmed {
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
