package bid754

import (
	"math/big"
	"strconv"
	"strings"
)

type bidNaNLiteral struct {
	negative  bool
	signaling bool
	payload   string
}

// parseBIDNaNLiteral recognizes the public NaN literal grammar:
// [+|-] (nan | qnan | snan) [decimal digits], case-insensitive, after the
// same leading-whitespace skip the Intel bid<w>_from_string port applies to
// numeric/infinity input: leading spaces and tabs (exactly " \t", not the
// full Unicode space set) are trimmed, and whitespace anywhere else --
// trailing, after the sign, or inside the payload -- is outside the grammar
// and rejected. So " nan" parses exactly as " 1" and " inf" do, while
// "nan " is rejected exactly as "1 " and "inf " are. decTest-style quoting
// ('NaN', "NaN") is likewise not part of the public grammar; the decTest
// harness strips quotes before reaching this layer.
func parseBIDNaNLiteral(input string) (bidNaNLiteral, bool) {
	rest := strings.TrimLeft(input, " \t")
	if rest == "" {
		return bidNaNLiteral{}, false
	}

	lit := bidNaNLiteral{}
	switch rest[0] {
	case '+':
		rest = rest[1:]
	case '-':
		lit.negative = true
		rest = rest[1:]
	}
	if rest == "" {
		return bidNaNLiteral{}, false
	}

	lower := strings.ToLower(rest)
	switch {
	case strings.HasPrefix(lower, "snan"):
		lit.signaling = true
		lit.payload = rest[4:]
	case strings.HasPrefix(lower, "qnan"):
		lit.payload = rest[4:]
	case strings.HasPrefix(lower, "nan"):
		lit.payload = rest[3:]
	default:
		return bidNaNLiteral{}, false
	}

	if lit.payload != "" {
		for _, r := range lit.payload {
			if r < '0' || r > '9' {
				return bidNaNLiteral{}, false
			}
		}
		lit.payload = strings.TrimLeft(lit.payload, "0")
	}
	return lit, true
}

func parseDecimal32BIDNaN(input string) (Decimal32BID, bool) {
	lit, ok := parseBIDNaNLiteral(input)
	if !ok {
		return 0, false
	}
	payload, ok := parseUintPayload(lit.payload, 999999)
	if !ok {
		return 0, false
	}

	bits := uint32(0x7c000000)
	if lit.signaling {
		bits = 0x7e000000
	}
	if lit.negative {
		bits |= 0x80000000
	}
	bits |= uint32(payload)
	return Decimal32BID(bits), true
}

func parseDecimal64BIDNaN(input string) (Decimal64BID, bool) {
	lit, ok := parseBIDNaNLiteral(input)
	if !ok {
		return 0, false
	}
	payload, ok := parseUintPayload(lit.payload, 999999999999999)
	if !ok {
		return 0, false
	}

	bits := uint64(0x7c00000000000000)
	if lit.signaling {
		bits = 0x7e00000000000000
	}
	if lit.negative {
		bits |= 0x8000000000000000
	}
	bits |= payload
	return Decimal64BID(bits), true
}

func parseDecimal128BIDNaN(input string) (Decimal128BID, bool) {
	lit, ok := parseBIDNaNLiteral(input)
	if !ok {
		return Decimal128BID{}, false
	}
	payload, ok := parseBigPayload(lit.payload, decimal128NaNPayloadLimit())
	if !ok {
		return Decimal128BID{}, false
	}

	lo := new(big.Int).And(payload, new(big.Int).SetUint64(^uint64(0))).Uint64()
	hi := new(big.Int).Rsh(payload, 64).Uint64() | 0x7c00000000000000
	if lit.signaling {
		hi = (hi &^ uint64(0x7c00000000000000)) | 0x7e00000000000000
	}
	if lit.negative {
		hi |= 0x8000000000000000
	}
	return decimal128BIDFromWords(hi, lo), true
}

func parseUintPayload(payload string, max uint64) (uint64, bool) {
	if payload == "" {
		return 0, true
	}
	value, err := strconv.ParseUint(payload, 10, 64)
	if err != nil || value > max {
		return 0, false
	}
	return value, true
}

func parseBigPayload(payload string, limit *big.Int) (*big.Int, bool) {
	if payload == "" {
		return new(big.Int), true
	}
	value, ok := new(big.Int).SetString(payload, 10)
	if !ok || value.Sign() < 0 || value.Cmp(limit) >= 0 {
		return nil, false
	}
	return value, true
}

func decimal128NaNPayloadLimit() *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(33), nil)
}

func formatDecimal32BIDNaN(bits uint32) (string, bool) {
	if bits&0x7c000000 != 0x7c000000 {
		return "", false
	}
	payload := uint64(bits & 0x000fffff)
	return formatBIDNaN(bits&0x80000000 != 0, bits&0x7e000000 == 0x7e000000, payloadString(payload, 1000000)), true
}

func formatDecimal64BIDNaN(bits uint64) (string, bool) {
	if bits&0x7c00000000000000 != 0x7c00000000000000 {
		return "", false
	}
	payload := bits & 0x0003ffffffffffff
	return formatBIDNaN(bits&0x8000000000000000 != 0, bits&0x7e00000000000000 == 0x7e00000000000000, payloadString(payload, 1000000000000000)), true
}

func formatDecimal128BIDNaN(d Decimal128BID) (string, bool) {
	hi, lo := decimal128BIDWords(d)
	if hi&0x7c00000000000000 != 0x7c00000000000000 {
		return "", false
	}
	payload := new(big.Int).SetUint64(hi & 0x00003fffffffffff)
	payload.Lsh(payload, 64)
	payload.Or(payload, new(big.Int).SetUint64(lo))
	payloadText := ""
	if payload.Sign() != 0 && payload.Cmp(decimal128NaNPayloadLimit()) < 0 {
		payloadText = payload.String()
	}
	return formatBIDNaN(hi&0x8000000000000000 != 0, hi&0x7e00000000000000 == 0x7e00000000000000, payloadText), true
}

// payloadString renders a NaN payload only when it is canonical: nonzero and
// strictly below canonicalLimit (10^6 for Decimal32, 10^15 for Decimal64). A
// noncanonical payload is suppressed to an empty string so String() never emits
// a payload that the NaN literal parser (parseUintPayload, same bounds) would
// reject; otherwise the String()->parse round-trip would drop the payload. This
// keeps the 32/64-bit paths symmetric with formatDecimal128BIDNaN, which
// already clamps at 10^33.
func payloadString(payload, canonicalLimit uint64) string {
	if payload == 0 || payload >= canonicalLimit {
		return ""
	}
	return strconv.FormatUint(payload, 10)
}

func formatBIDNaN(negative, signaling bool, payload string) string {
	var b strings.Builder
	if negative {
		b.WriteByte('-')
	} else {
		b.WriteByte('+')
	}
	if signaling {
		b.WriteString("SNaN")
	} else {
		b.WriteString("NaN")
	}
	b.WriteString(payload)
	return b.String()
}
