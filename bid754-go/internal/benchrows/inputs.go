package benchrows

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

// InputPair carries one width's exact decimal benchmark operands.
type InputPair struct {
	X string `json:"x"`
	Y string `json:"y"`
	// Z is the third fma operand (result = x*y + z); sqrt reuses the
	// non-negative X operand and needs no dedicated input.
	Z string `json:"z"`
}

// Inputs is the shared format-2 exact-operand benchmark input contract
// (testdata/benchmark_inputs.json) consumed by every measured layer.
type Inputs struct {
	FormatVersion  int       `json:"format_version"`
	IntegerOperand int64     `json:"integer_operand"`
	ScaleExponent  int       `json:"scale_exponent"`
	Decimal32      InputPair `json:"decimal32"`
	Decimal64      InputPair `json:"decimal64"`
	Decimal128     InputPair `json:"decimal128"`
}

// LoadInputs reads the benchmark input contract from path and checks the
// contract shape. Callers pass their own working-directory-relative path
// (module root: "testdata/benchmark_inputs.json"; the bidgo directory:
// "../../testdata/benchmark_inputs.json"). Semantic validation of the operand
// values happens in Prepare.
func LoadInputs(path string) (Inputs, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Inputs{}, fmt.Errorf("read benchmark input contract: %w", err)
	}
	var inputs Inputs
	if err := json.Unmarshal(data, &inputs); err != nil {
		return Inputs{}, fmt.Errorf("parse benchmark input contract: %w", err)
	}
	if inputs.FormatVersion != 2 {
		return Inputs{}, fmt.Errorf("benchmark input format_version = %d, want 2", inputs.FormatVersion)
	}
	for _, item := range []struct {
		name string
		pair InputPair
	}{
		{"decimal32", inputs.Decimal32},
		{"decimal64", inputs.Decimal64},
		{"decimal128", inputs.Decimal128},
	} {
		if item.pair.X == "" || item.pair.Y == "" || item.pair.Z == "" {
			return Inputs{}, fmt.Errorf("benchmark input %s requires non-empty x, y, and z", item.name)
		}
	}
	return inputs, nil
}

// Prepared carries the parsed exact operand bits shared by every Go-port
// benchmark row for one input contract.
type Prepared struct {
	Inputs     Inputs
	X32        uint32
	Y32        uint32
	Z32        uint32
	Integer32  uint32
	X64        uint64
	Y64        uint64
	Z64        uint64
	Integer64  uint64
	X128       bidgo.BID_UINT128
	Y128       bidgo.BID_UINT128
	Z128       bidgo.BID_UINT128
	Integer128 bidgo.BID_UINT128
}

// Prepare semantically validates one benchmark input contract (exact finite
// operands, cohort preservation, non-negative sqrt operands, non-zero
// division operands, exact integer round trips, finite scaleB results) and
// parses the shared operand bits.
func Prepare(inputs Inputs) (Prepared, error) {
	if inputs.IntegerOperand == 0 {
		return Prepared{}, fmt.Errorf("benchmark integer_operand must be non-zero")
	}
	if inputs.ScaleExponent == 0 {
		return Prepared{}, fmt.Errorf("benchmark scale_exponent must be non-zero")
	}
	if !scaleExponentFitsCInt(int64(inputs.ScaleExponent)) {
		return Prepared{}, fmt.Errorf("benchmark scale_exponent %d does not fit the Intel C int32 contract", inputs.ScaleExponent)
	}

	prepared := Prepared{Inputs: inputs}
	var err error
	if prepared.X32, err = parseExactDecimal32(inputs.Decimal32.X); err != nil {
		return Prepared{}, err
	}
	if prepared.Y32, err = parseExactDecimal32(inputs.Decimal32.Y); err != nil {
		return Prepared{}, err
	}
	if prepared.Z32, err = parseExactDecimal32(inputs.Decimal32.Z); err != nil {
		return Prepared{}, err
	}
	if prepared.X64, err = parseExactDecimal64(inputs.Decimal64.X); err != nil {
		return Prepared{}, err
	}
	if prepared.Y64, err = parseExactDecimal64(inputs.Decimal64.Y); err != nil {
		return Prepared{}, err
	}
	if prepared.Z64, err = parseExactDecimal64(inputs.Decimal64.Z); err != nil {
		return Prepared{}, err
	}
	if prepared.X128, err = parseExactDecimal128(inputs.Decimal128.X); err != nil {
		return Prepared{}, err
	}
	if prepared.Y128, err = parseExactDecimal128(inputs.Decimal128.Y); err != nil {
		return Prepared{}, err
	}
	if prepared.Z128, err = parseExactDecimal128(inputs.Decimal128.Z); err != nil {
		return Prepared{}, err
	}

	if err := validateSqrtOperandsNonNegative(prepared); err != nil {
		return Prepared{}, err
	}
	if err := validateDivisionOperandsNonZero(prepared); err != nil {
		return Prepared{}, err
	}
	if err := prepared.parseIntegerOperands(); err != nil {
		return Prepared{}, err
	}
	if err := validateScaleBOperandsStayFinite(prepared); err != nil {
		return Prepared{}, err
	}
	return prepared, nil
}

func scaleExponentFitsCInt(exponent int64) bool {
	return exponent >= -1<<31 && exponent <= 1<<31-1
}

func parseExactDecimal32(input string) (uint32, error) {
	value, flags := bidgo.Bid32FromStringRaw(input, 0)
	if flags != 0 || bidgo.Bid32IsFinite(value) == 0 {
		return 0, fmt.Errorf("Decimal32 benchmark input %q is not finite and exact: flags=%#x bits=%#x", input, flags, value)
	}
	if err := requireCohortPreserved("Decimal32", input, bidgo.Bid32ToStringRaw(value)); err != nil {
		return 0, err
	}
	return value, nil
}

func parseExactDecimal64(input string) (uint64, error) {
	value, flags := bidgo.Bid64FromString(input, 0)
	if flags != 0 || bidgo.Bid64IsFinite(value) == 0 {
		return 0, fmt.Errorf("Decimal64 benchmark input %q is not finite and exact: flags=%#x bits=%#x", input, flags, value)
	}
	if err := requireCohortPreserved("Decimal64", input, bidgo.Bid64ToString(value)); err != nil {
		return 0, err
	}
	return value, nil
}

func parseExactDecimal128(input string) (bidgo.BID_UINT128, error) {
	value, flags := bidgo.Bid128FromString(input, 0)
	if flags != 0 || bidgo.Bid128IsFinite(value) == 0 {
		lo, hi := bid128Bits(value)
		return bidgo.BID_UINT128{}, fmt.Errorf("Decimal128 benchmark input %q is not finite and exact: flags=%#x bits=%#x/%#x", input, flags, hi, lo)
	}
	if err := requireCohortPreserved("Decimal128", input, bidgo.Bid128ToString(value)); err != nil {
		return bidgo.BID_UINT128{}, err
	}
	return value, nil
}

func requireCohortPreserved(width, input, actual string) error {
	want, err := inputCohort(input)
	if err != nil {
		return fmt.Errorf("%s benchmark input %q has no exact decimal cohort: %w", width, input, err)
	}
	if actual != want {
		return fmt.Errorf("%s benchmark input %q does not preserve the requested cohort: got %s, want %s", width, input, actual, want)
	}
	return nil
}

// inputCohort renders a finite benchmark literal as the exact
// coefficient/exponent pair requested by its spelling. Trailing coefficient
// zeros are intentionally retained because they select a different cohort.
func inputCohort(input string) (string, error) {
	s := strings.TrimLeft(input, " \t")
	if s == "" {
		return "", fmt.Errorf("empty decimal literal")
	}

	sign := byte('+')
	position := 0
	if s[position] == '+' || s[position] == '-' {
		sign = s[position]
		position++
		if position == len(s) {
			return "", fmt.Errorf("missing coefficient")
		}
	}

	coefficient := make([]byte, 0, len(s)-position)
	fractionalDigits := int64(0)
	radixSeen := false
	digitSeen := false
	exponentPosition := len(s)
	for ; position < len(s); position++ {
		character := s[position]
		switch {
		case character >= '0' && character <= '9':
			coefficient = append(coefficient, character)
			digitSeen = true
			if radixSeen {
				fractionalDigits++
			}
		case character == '.' && !radixSeen:
			radixSeen = true
		case (character == 'e' || character == 'E') && digitSeen:
			exponentPosition = position
			position = len(s)
		default:
			return "", fmt.Errorf("invalid character %q", character)
		}
	}
	if !digitSeen {
		return "", fmt.Errorf("missing coefficient digits")
	}

	explicitExponent := int64(0)
	if exponentPosition < len(s) {
		exponentText := s[exponentPosition+1:]
		if exponentText == "" {
			return "", fmt.Errorf("missing exponent digits")
		}
		digitStart := 0
		if exponentText[0] == '+' || exponentText[0] == '-' {
			digitStart = 1
		}
		if digitStart == len(exponentText) {
			return "", fmt.Errorf("missing exponent digits")
		}
		for _, character := range exponentText[digitStart:] {
			if character < '0' || character > '9' {
				return "", fmt.Errorf("invalid exponent character %q", character)
			}
		}
		parsed, err := strconv.ParseInt(exponentText, 10, 64)
		if err != nil {
			return "", fmt.Errorf("exponent out of range: %w", err)
		}
		explicitExponent = parsed
	}

	const minInt64 = -1 << 63
	if explicitExponent < minInt64+fractionalDigits {
		return "", fmt.Errorf("cohort exponent out of range")
	}
	exponent := explicitExponent - fractionalDigits
	canonicalCoefficient := strings.TrimLeft(string(coefficient), "0")
	if canonicalCoefficient == "" {
		canonicalCoefficient = "0"
	}
	return fmt.Sprintf("%c%sE%+d", sign, canonicalCoefficient, exponent), nil
}

// validateSqrtOperandsNonNegative pins the sqrt-benchmark precondition: the
// sqrt rows reuse the X operands, so a negative X would silently turn every
// sqrt benchmark into a NaN/invalid path instead of a real square root. The
// check inspects the parsed value's sign and NaN class, not the raw text, so
// a disguised spelling (e.g. " -1") cannot slip past it.
func validateSqrtOperandsNonNegative(p Prepared) error {
	if bidgo.Bid32IsNaN(p.X32) || bidgo.Bid32IsSigned(p.X32) != 0 {
		return fmt.Errorf("benchmark input decimal32 x = %q parses negative or NaN; sqrt benchmarks reuse x and require a non-negative operand", p.Inputs.Decimal32.X)
	}
	if bidgo.Bid64IsNaN(p.X64) != 0 || bidgo.Bid64IsSigned(p.X64) != 0 {
		return fmt.Errorf("benchmark input decimal64 x = %q parses negative or NaN; sqrt benchmarks reuse x and require a non-negative operand", p.Inputs.Decimal64.X)
	}
	if bidgo.Bid128IsNaN(p.X128) != 0 || bidgo.Bid128IsSigned(p.X128) != 0 {
		return fmt.Errorf("benchmark input decimal128 x = %q parses negative or NaN; sqrt benchmarks reuse x and require a non-negative operand", p.Inputs.Decimal128.X)
	}
	return nil
}

func validateDivisionOperandsNonZero(p Prepared) error {
	if bidgo.Bid32IsZero(p.X32) {
		return fmt.Errorf("benchmark input decimal32.x must be non-zero for remainder rows")
	}
	if bidgo.Bid32IsZero(p.Y32) {
		return fmt.Errorf("benchmark input decimal32.y must be non-zero for division rows")
	}
	if bidgo.Bid64IsZero(p.X64) != 0 {
		return fmt.Errorf("benchmark input decimal64.x must be non-zero for remainder and mixed rows")
	}
	if bidgo.Bid64IsZero(p.Y64) != 0 {
		return fmt.Errorf("benchmark input decimal64.y must be non-zero for division and mixed rows")
	}
	if bidgo.Bid128IsZero(p.X128) != 0 {
		return fmt.Errorf("benchmark input decimal128.x must be non-zero for remainder and mixed rows")
	}
	if bidgo.Bid128IsZero(p.Y128) != 0 {
		return fmt.Errorf("benchmark input decimal128.y must be non-zero for division and mixed rows")
	}
	return nil
}

func (p *Prepared) parseIntegerOperands() error {
	operand := p.Inputs.IntegerOperand

	d32, flags32 := bidgo.Bid32FromInt64(operand, 0)
	if flags32 != 0 {
		return fmt.Errorf("benchmark integer_operand %d is not exact as Decimal32: flags=%#x", operand, flags32)
	}
	got32, flags32 := bidgo.Bid32ToInt64Rnint(d32)
	if got32 != operand || flags32 != 0 {
		return fmt.Errorf("Decimal32 benchmark integer round trip = (%d, %#x), want (%d, 0)", got32, flags32, operand)
	}
	p.Integer32 = d32

	d64, flags64 := bidgo.Bid64FromInt64(operand, 0)
	if flags64 != 0 {
		return fmt.Errorf("benchmark integer_operand %d is not exact as Decimal64: flags=%#x", operand, flags64)
	}
	got64, flags64 := bidgo.Bid64ToInt64Rnint(d64)
	if got64 != operand || flags64 != 0 {
		return fmt.Errorf("Decimal64 benchmark integer round trip = (%d, %#x), want (%d, 0)", got64, flags64, operand)
	}
	p.Integer64 = d64

	d128 := bidgo.Bid128FromInt64(operand)
	got128, flags128 := bidgo.Bid128ToInt64Rnint(d128)
	if got128 != operand || flags128 != 0 {
		return fmt.Errorf("Decimal128 benchmark integer round trip = (%d, %#x), want (%d, 0)", got128, flags128, operand)
	}
	p.Integer128 = d128
	return nil
}

func validateScaleBOperandsStayFinite(p Prepared) error {
	exponent := p.Inputs.ScaleExponent
	if got, flags := bidgo.Bid32Scalbn(p.X32, exponent, 0); flags != 0 || bidgo.Bid32IsFinite(got) == 0 {
		return fmt.Errorf("Decimal32 benchmark scaleB input raised flags or became non-finite: bits=%#x flags=%#x", got, flags)
	}
	if got, flags := bidgo.Bid64Scalbn(p.X64, exponent, 0); flags != 0 || bidgo.Bid64IsFinite(got) == 0 {
		return fmt.Errorf("Decimal64 benchmark scaleB input raised flags or became non-finite: bits=%#x flags=%#x", got, flags)
	}
	var flags128 uint32
	got128 := bidgo.Bid128Scalbn(p.X128, exponent, 0, &flags128)
	if flags128 != 0 || bidgo.Bid128IsFinite(got128) == 0 {
		lo, hi := bid128Bits(got128)
		return fmt.Errorf("Decimal128 benchmark scaleB input raised flags or became non-finite: bits=%#x/%#x flags=%#x", hi, lo, flags128)
	}
	return nil
}
