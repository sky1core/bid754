package bid754

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/sky1core/bid754/bid754-go/internal/decimalprobe"
	"github.com/sky1core/bid754/bid754-go/internal/decimalref"
)

var finiteModes = []string{"nearest_even", "nearest_away", "toward_zero", "toward_positive", "toward_negative"}

type finiteArithmetic[T any] interface {
	AddWithMode(T, RoundingMode) (T, ExceptionFlags)
	SubWithMode(T, RoundingMode) (T, ExceptionFlags)
	MulWithMode(T, RoundingMode) (T, ExceptionFlags)
	DivWithMode(T, RoundingMode) (T, ExceptionFlags)
	FMAWithMode(T, T, RoundingMode) (T, ExceptionFlags)
	QuantizeWithMode(T, RoundingMode) (T, ExceptionFlags)
}

func finiteApply[T finiteArithmetic[T]](op string, operands []T, mode RoundingMode) (T, ExceptionFlags, error) {
	var zero T
	if len(operands) != 2 && !(op == "fma" && len(operands) == 3) {
		return zero, 0, fmt.Errorf("invalid operand count")
	}
	x, y := operands[0], operands[1]
	var r T
	var flags ExceptionFlags
	switch op {
	case "add":
		r, flags = x.AddWithMode(y, mode)
	case "sub":
		r, flags = x.SubWithMode(y, mode)
	case "mul":
		r, flags = x.MulWithMode(y, mode)
	case "div":
		r, flags = x.DivWithMode(y, mode)
	case "fma":
		if len(operands) != 3 {
			return zero, 0, fmt.Errorf("fma requires three operands")
		}
		r, flags = x.FMAWithMode(y, operands[2], mode)
	case "quantize":
		r, flags = x.QuantizeWithMode(y, mode)
	default:
		return zero, 0, fmt.Errorf("unsupported operation %q", op)
	}
	return r, flags, nil
}

func finitePublic(c decimalref.Case) (string, uint32, error) {
	modes := map[string]RoundingMode{
		"nearest_even": RoundNearestEven, "nearest_away": RoundNearestAway, "toward_zero": RoundTowardZero,
		"toward_positive": RoundTowardPositive, "toward_negative": RoundTowardNegative,
	}
	mode, ok := modes[c.Mode]
	if !ok {
		return "", 0, fmt.Errorf("unknown rounding mode")
	}
	switch c.Width {
	case 32:
		var operands []Decimal32BID
		for _, raw := range c.Operands {
			x, err := strconv.ParseUint(raw, 16, 32)
			if err != nil {
				return "", 0, err
			}
			operands = append(operands, Decimal32BIDFromBits(uint32(x)))
		}
		r, flags, err := finiteApply(c.Op, operands, mode)
		return finiteOutput(fmt.Sprintf("%08x", r.ToUint32()), flags, err)
	case 64:
		var operands []Decimal64BID
		for _, raw := range c.Operands {
			x, err := strconv.ParseUint(raw, 16, 64)
			if err != nil {
				return "", 0, err
			}
			operands = append(operands, Decimal64BIDFromBits(x))
		}
		r, flags, err := finiteApply(c.Op, operands, mode)
		return finiteOutput(fmt.Sprintf("%016x", r.ToUint64()), flags, err)
	case 128:
		var operands []Decimal128BID
		for _, raw := range c.Operands {
			parts := strings.Split(raw, ":")
			if len(parts) != 2 {
				return "", 0, fmt.Errorf("invalid decimal128 raw image")
			}
			hi, err := strconv.ParseUint(parts[0], 16, 64)
			if err != nil {
				return "", 0, err
			}
			lo, err := strconv.ParseUint(parts[1], 16, 64)
			if err != nil {
				return "", 0, err
			}
			var bytes [16]byte
			binary.LittleEndian.PutUint64(bytes[:8], lo)
			binary.LittleEndian.PutUint64(bytes[8:], hi)
			operands = append(operands, Decimal128BIDFromBytes(bytes))
		}
		r, flags, err := finiteApply(c.Op, operands, mode)
		bytes := r.ToBytes()
		return finiteOutput(fmt.Sprintf("%016x:%016x", binary.LittleEndian.Uint64(bytes[8:]), binary.LittleEndian.Uint64(bytes[:8])), flags, err)
	default:
		return "", 0, fmt.Errorf("unknown width")
	}
}

func finiteOutput(raw string, flags ExceptionFlags, err error) (string, uint32, error) {
	if err != nil {
		return "", 0, err
	}
	if flags & ^(FlagInvalidOperation|FlagDivisionByZero|FlagOverflow|FlagUnderflow|FlagInexact) != 0 {
		return "", 0, fmt.Errorf("unexpected public arithmetic flags %#x", flags)
	}
	var native uint32
	for flag, bit := range map[ExceptionFlags]uint32{FlagInvalidOperation: 0x01, FlagDivisionByZero: 0x04, FlagOverflow: 0x08, FlagUnderflow: 0x10, FlagInexact: 0x20} {
		if flags&flag != 0 {
			native |= bit
		}
	}
	return raw, native, nil
}

func FuzzFiniteArithmeticExact(f *testing.F) {
	for family := range decimalprobe.Families() {
		for width := uint8(0); width < 3; width++ {
			for mode := uint8(0); mode < 5; mode++ {
				f.Add(uint8(family), width, mode, uint64(0), uint64(754), int32(0), false)
				f.Add(uint8(family), width, mode, ^uint64(0), ^uint64(0), int32(-31), true)
			}
		}
	}
	f.Fuzz(func(t *testing.T, family, width, mode uint8, hi, lo uint64, exponent int32, negative bool) {
		families := decimalprobe.Families()
		sample, err := decimalprobe.Generate(families[int(family)%len(families)], []int{32, 64, 128}[width%3], finiteModes[mode%5], hi, lo, exponent, negative)
		if err != nil {
			t.Fatal(err)
		}
		want, err := decimalprobe.Validate(sample)
		if err != nil {
			t.Fatal(err)
		}
		observations, err := finitePublicPaths(sample.Case)
		if err != nil {
			t.Fatal(err)
		}
		if err := finiteCheckPaths(sample.Case, want, observations, []string{"go"}); err != nil {
			t.Fatalf("sample=%+v: %v", sample, err)
		}
	})
}
