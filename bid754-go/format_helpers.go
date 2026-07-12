package bid754

import (
	"strconv"
	"strings"
)

// formatDecimalString normalizes decimal string representations from different backends
// It converts scientific notation to plain decimal format when appropriate
func formatDecimalString(s string) string {
	// Remove leading '+' sign
	s = strings.TrimPrefix(s, "+")

	// Check if it's in scientific notation
	if !strings.Contains(s, "E") && !strings.Contains(s, "e") {
		return s
	}

	// Parse the scientific notation
	parts := strings.Split(strings.ToUpper(s), "E")
	if len(parts) != 2 {
		return s
	}

	mantissa := parts[0]
	exponentStr := parts[1]

	// Remove leading '+' from exponent
	exponentStr = strings.TrimPrefix(exponentStr, "+")

	exponent, err := strconv.Atoi(exponentStr)
	if err != nil {
		return s
	}

	// Parse mantissa
	negative := false
	if strings.HasPrefix(mantissa, "-") {
		negative = true
		mantissa = mantissa[1:]
	}

	// Remove decimal point to get digits
	digits := strings.Replace(mantissa, ".", "", 1)

	// A zero coefficient is just zero at any exponent, so this check sits
	// before the exponent-range check: without it a positive exponent would
	// expand into a run of zero digits, and a |exponent| > 20 zero would
	// stay in scientific notation (e.g. "0E+21").
	if strings.Trim(digits, "0") == "" {
		if negative {
			return "-0"
		}
		return "0"
	}

	// Special case: if exponent is within reasonable range, convert to simple format
	// For decimal64 (16 digits precision), handle exponents from -20 to 20
	if exponent >= -20 && exponent <= 20 {
		decimalPos := strings.Index(mantissa, ".")
		if decimalPos == -1 {
			decimalPos = len(mantissa)
		}

		// Calculate new decimal position
		newDecimalPos := decimalPos + exponent

		// Build result
		var result strings.Builder
		if negative {
			result.WriteByte('-')
		}

		if newDecimalPos <= 0 {
			// Need leading zeros
			result.WriteString("0.")
			for i := 0; i < -newDecimalPos; i++ {
				result.WriteByte('0')
			}
			result.WriteString(digits)
		} else if newDecimalPos >= len(digits) {
			// Need trailing zeros
			result.WriteString(digits)
			for i := len(digits); i < newDecimalPos; i++ {
				result.WriteByte('0')
			}
			// Add .0 if it's a whole number to match expected format
			if strings.Contains(s, ".") || exponent < 0 {
				result.WriteString(".0")
			}
		} else {
			// Insert decimal point in the middle
			result.WriteString(digits[:newDecimalPos])
			result.WriteByte('.')
			result.WriteString(digits[newDecimalPos:])
		}

		// Remove trailing zeros after decimal point, but keep at least one
		resultStr := result.String()
		if strings.Contains(resultStr, ".") {
			resultStr = strings.TrimRight(resultStr, "0")
			if strings.HasSuffix(resultStr, ".") {
				resultStr += "0"
			}
		}

		return resultStr
	}

	// For large exponents, keep scientific notation but normalize format
	return s
}
