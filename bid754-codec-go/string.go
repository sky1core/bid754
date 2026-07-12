package bidcodec

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// ToString converts Components to the shared BID codec string format:
// normalized scientific notation with an explicit sign. It validates the
// schema-level Components contract before rendering, so a field that the
// selected Kind cannot carry, a zero/negative/oversized normal coefficient,
// or a negative/oversized NaN payload fails instead of being silently lost or
// producing a string that FromString cannot parse.
// Examples: "+1.234567E+9", "-Inf", "+NaN", "+0E-2"
func ToString(c Components) (string, error) {
	if err := validateStringComponents(c); err != nil {
		return "", err
	}
	prefix := "+"
	if c.Sign {
		prefix = "-"
	}
	switch c.Kind {
	case Infinity:
		return prefix + "Inf", nil
	case QNaN:
		if c.Payload != nil && c.Payload.Sign() != 0 {
			return prefix + "NaN" + c.Payload.String(), nil
		}
		return prefix + "NaN", nil
	case SNaN:
		if c.Payload != nil && c.Payload.Sign() != 0 {
			return prefix + "SNaN" + c.Payload.String(), nil
		}
		return prefix + "SNaN", nil
	case Zero:
		if c.Exponent == 0 {
			return prefix + "0", nil
		}
		return fmt.Sprintf("%s0E%+d", prefix, c.Exponent), nil
	}
	// Normal
	digits := c.Coefficient.String()
	exp := int(c.Exponent) + len(digits) - 1
	if len(digits) == 1 {
		return fmt.Sprintf("%s%sE%+d", prefix, digits, exp), nil
	}
	return fmt.Sprintf("%s%s.%sE%+d", prefix, digits[:1], digits[1:], exp), nil
}

func validateStringComponents(c Components) error {
	if err := validateKindFields("BID codec string", c); err != nil {
		return err
	}
	switch c.Kind {
	case Normal:
		if c.Coefficient == nil {
			return fmt.Errorf("BID codec string: normal value has nil coefficient")
		}
		if c.Coefficient.Sign() < 0 {
			return fmt.Errorf("BID codec string: coefficient %s is negative", c.Coefficient.String())
		}
		if c.Coefficient.Cmp(ten34) >= 0 {
			return fmt.Errorf("BID codec string: coefficient %s exceeds schema max %s", c.Coefficient.String(), bid128MaxCoeffDecimal)
		}
	case Zero, Infinity:
		return nil
	case QNaN, SNaN:
		if c.Payload == nil {
			return nil
		}
		if c.Payload.Sign() < 0 {
			return fmt.Errorf("BID codec string: NaN payload %s is negative", c.Payload.String())
		}
		if c.Payload.Cmp(ten33) >= 0 {
			return fmt.Errorf("BID codec string: NaN payload %s exceeds schema max %s", c.Payload.String(), bid128PayMax.String())
		}
	default:
		return fmt.Errorf("BID codec string: unrecognized kind %d", uint8(c.Kind))
	}
	return nil
}

// FromString parses an IEEE 754 string into Components using one strict ASCII
// grammar (identical across the cross-language codec packages).
//
// The whole input must be ASCII: any non-ASCII byte anywhere is malformed,
// including Unicode digit variants and Unicode whitespace. Only ASCII
// whitespace may surround the token (removed by one surrounding trim); there is
// no whitespace inside the token. After the optional single leading +/-, the
// input is either a special token (Inf/Infinity, or NaN/SNaN followed by an
// optional unsigned ASCII-digit payload whose value must be below the
// schema-wide NaN payload limit 10^33, matched case-insensitively) or a number:
// ASCII digits with at most one '.'
// and at least one digit, optionally followed by E/e, one optional sign, and
// ASCII exponent digits where the exponent literal must be below the shared
// exact-integer literal bound 2^53 in magnitude (the widest bound every
// consumer's number type can check exactly; a literal at or beyond it is
// rejected through the same error channel) and the fraction-adjusted FINAL
// exponent must fit a signed 32-bit integer.
// The literal itself is deliberately allowed past int32: ToString renders the
// adjusted exponent (exponent + digits - 1, up to int32 max + 33, far below
// 2^53), and every such rendering must reparse (round-trip closure:
// FromString(ToString(c)) always succeeds for valid Components). Underscores
// and payload-internal signs are malformed everywhere.
//
// FromString validates grammar and schema limits only, identical in all six
// language packages: the parsed coefficient value must not exceed the
// schema-wide maximum coefficient 10^34-1 (the largest value any supported BID
// width can hold — a schema constant, not per-width validation), and the parsed
// payload value must be below the schema-wide NaN payload limit 10^33 (the
// widest canonical BID128 NaN payload, the same kind of schema constant). BID
// width-range validation is the Encode* contract.
func FromString(s string) (Components, error) {
	// (1) Whole-input ASCII scan, before any trimming, so that non-ASCII
	// whitespace and Unicode digit variants are rejected rather than silently
	// trimmed away or parsed.
	for i := 0; i < len(s); i++ {
		if s[i] > 0x7f {
			return Components{}, fmt.Errorf("non-ASCII byte 0x%02x at position %d", s[i], i)
		}
	}

	// (2) Trim only ASCII whitespace. strings.TrimSpace is deliberately avoided:
	// its Unicode whitespace set differs across languages and would let the
	// grammar diverge.
	s = trimASCIISpace(s)
	if len(s) == 0 {
		return Components{}, fmt.Errorf("empty string")
	}

	// (3) Parse. Optional single leading sign.
	sign := false
	if s[0] == '+' {
		s = s[1:]
	} else if s[0] == '-' {
		sign = true
		s = s[1:]
	}
	if len(s) == 0 {
		return Components{}, fmt.Errorf("no digits")
	}

	upper := strings.ToUpper(s)
	if upper == "INF" || upper == "INFINITY" {
		return Components{Sign: sign, Kind: Infinity}, nil
	}
	if strings.HasPrefix(upper, "SNAN") {
		payload, err := parseNaNPayload(s[4:])
		if err != nil {
			return Components{}, err
		}
		return Components{Sign: sign, Kind: SNaN, Payload: payload}, nil
	}
	if strings.HasPrefix(upper, "NAN") {
		payload, err := parseNaNPayload(s[3:])
		if err != nil {
			return Components{}, err
		}
		return Components{Sign: sign, Kind: QNaN, Payload: payload}, nil
	}

	// Number: ASCII digits, at most one '.', at least one digit.
	var digits []byte
	expAdjust := 0
	foundDot := false
	i := 0
	for i < len(s) && s[i] != 'E' && s[i] != 'e' {
		switch {
		case s[i] == '.':
			if foundDot {
				return Components{}, fmt.Errorf("multiple decimal points")
			}
			foundDot = true
		case s[i] >= '0' && s[i] <= '9':
			digits = append(digits, s[i])
			if foundDot {
				expAdjust--
			}
		default:
			return Components{}, fmt.Errorf("unexpected character %q in coefficient", s[i])
		}
		i++
	}

	if len(digits) == 0 {
		return Components{}, fmt.Errorf("no digits")
	}

	// Optional exponent: E/e, one optional sign, at least one ASCII digit.
	expPart := int64(0)
	if i < len(s) {
		// The loop only stops early on 'E'/'e', so s[i] is the exponent marker.
		n, err := parseExponentLiteral(s[i+1:])
		if err != nil {
			return Components{}, err
		}
		expPart = n
	}

	// Remove leading zeros (keep at least one digit).
	start := 0
	for start < len(digits)-1 && digits[start] == '0' {
		start++
	}
	digits = digits[start:]

	coeff, ok := new(big.Int).SetString(string(digits), 10)
	if !ok {
		return Components{}, fmt.Errorf("invalid coefficient: %s", string(digits))
	}

	// Schema-wide coefficient cap: the parsed value (not the digit count) must
	// not exceed 10^34-1, the largest coefficient any supported BID width can
	// hold. This is a shared schema constant, identical in all six language
	// packages, so big-integer and fixed-width-integer languages fail the same
	// inputs the same way. Per-width range validation stays in Encode*.
	if coeff.Cmp(ten34) >= 0 {
		return Components{}, fmt.Errorf("coefficient %s exceeds schema max %s", coeff.String(), bid128MaxCoeffDecimal)
	}

	// Check the fold against int64 wrap before adding. Proof of exactness:
	// past this check expAdjust > math.MinInt64 + 2^53, and the literal bound
	// gives |expPart| < 2^53, so expPart + expAdjust > math.MinInt64 — no
	// negative wrap — while expAdjust <= 0 and expPart < 2^53 bound the sum
	// above by 2^53 — no positive wrap; the int64 sum is exact. When this
	// condition holds (a fraction of at least 2^63 - 2^53 digits), the mathematical
	// final exponent is at most 2^53 - (2^63 - 2^53), far below the int32
	// minimum, so the true answer is also a rejection: the accepted-input set
	// is unchanged.
	if int64(expAdjust) <= math.MinInt64+(1<<53) {
		return Components{}, fmt.Errorf("exponent out of int32 range: literal %d with %d fractional digits", expPart, -expAdjust)
	}
	exponent64 := expPart + int64(expAdjust)
	if exponent64 < -2147483648 || exponent64 > 2147483647 {
		return Components{}, fmt.Errorf("exponent out of int32 range: %d", exponent64)
	}
	exponent := int32(exponent64)

	if coeff.Sign() == 0 {
		return Components{Sign: sign, Exponent: exponent, Kind: Zero}, nil
	}
	return Components{
		Sign:        sign,
		Coefficient: coeff,
		Exponent:    exponent,
		Kind:        Normal,
	}, nil
}

// trimASCIISpace removes only ASCII whitespace (TAB, LF, VT, FF, CR, SPACE)
// from both ends. It intentionally does not use strings.TrimSpace, whose Unicode
// whitespace set differs across languages.
func trimASCIISpace(s string) string {
	start := 0
	for start < len(s) && isASCIISpace(s[start]) {
		start++
	}
	end := len(s)
	for end > start && isASCIISpace(s[end-1]) {
		end--
	}
	return s[start:end]
}

func isASCIISpace(b byte) bool {
	switch b {
	case 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x20:
		return true
	default:
		return false
	}
}

// isASCIIDigits reports whether s is non-empty and consists solely of ASCII
// digits 0-9 (no sign, underscore, whitespace, or Unicode digit variants).
func isASCIIDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

// parseNaNPayload parses an optional NaN payload. Empty means a bare NaN/SNaN
// (payload 0); otherwise it must be unsigned ASCII digits whose value is below
// the schema-wide NaN payload limit 10^33 (the widest canonical BID128 NaN
// payload, replacing the former 64-bit fit rule so fixed-width and big-integer
// languages fail the same inputs the same way). The explicit ASCII pre-check
// prevents delegating a signed substring to big.Int parsing. It returns a
// non-nil, non-negative *big.Int (zero for a bare NaN).
func parseNaNPayload(s string) (*big.Int, error) {
	if s == "" {
		return big.NewInt(0), nil
	}
	if !isASCIIDigits(s) {
		return nil, fmt.Errorf("invalid NaN payload %q: must be unsigned ASCII digits", s)
	}
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid NaN payload %q", s)
	}
	if v.Cmp(ten33) >= 0 {
		return nil, fmt.Errorf("NaN payload %s exceeds schema max %s", v.String(), bid128PayMax.String())
	}
	return v, nil
}

// sharedExponentLiteralBound is the shared exact-integer exponent-literal
// bound 2^53: the widest bound every language consumer's number type can
// check exactly (JavaScript's safe-integer range pins it). A literal at or
// beyond this magnitude is rejected in every consumer through the same error
// channel, so every consumer decides each input its runtime can represent
// by the same mathematical rule (literal below 2^53, fraction-adjusted final
// exponent in int32) — a fixed-width fraction counter can force a rejection
// only in regions (over ~2^63 fraction digits) where that rule itself
// rejects, so no representable input is decided differently anywhere.
const sharedExponentLiteralBound = int64(1) << 53

// parseExponentLiteral parses an exponent literal: an optional single leading
// sign followed by at least one ASCII digit. Underscores, embedded whitespace,
// and Unicode digits are malformed, and the literal's magnitude must be below
// the shared exact-integer bound 2^53 (a literal at or beyond it — including
// anything past int64 — is rejected through the same error channel). The
// caller checks the fraction-adjusted FINAL exponent against the signed 32-bit
// range — the literal itself is allowed past int32 so every ToString rendering
// (adjusted-exponent literal at most int32 max + 33, far below 2^53) reparses
// successfully. The explicit ASCII pre-check keeps strconv from widening the
// grammar.
func parseExponentLiteral(s string) (int64, error) {
	body := s
	if len(body) > 0 && (body[0] == '+' || body[0] == '-') {
		body = body[1:]
	}
	if !isASCIIDigits(body) {
		return 0, fmt.Errorf("invalid exponent %q", s)
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil || n >= sharedExponentLiteralBound || n <= -sharedExponentLiteralBound {
		return 0, fmt.Errorf("exponent literal %q at or above the shared exact-integer bound 2^53", s)
	}
	return n, nil
}
