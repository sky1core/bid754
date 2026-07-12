package bid754

// Auxiliary portable fuzz target for the raw bit-pattern arithmetic surface.
// Like FuzzParseNoPanic (parse_fuzz_test.go) this is an exploration tool, not a
// regular generated verification domain: the readtest / decTest / FFI domains
// pin values against the Intel oracle; this target only hunts for panics. It
// injects arbitrary uint32 / uint64 / 16-byte patterns straight into the public
// value types via their defined-constant conversions (Decimal32BID(uint32),
// Decimal64BID(uint64), Decimal128BID([16]byte)) and runs the public methods
// over them. There is deliberately NO oracle comparison here (that is the FFI
// domain's job); the only failure mode asserted is a panic/crash, which a Go
// test surfaces automatically. It needs no native prerequisite and no build
// tag, and its seed corpus replays as ordinary test cases under a plain
// `go test ./...`, keeping the module stdlib-only.

import (
	"encoding/binary"
	"testing"
)

func exerciseDecimal32NoPanic(a, b Decimal32BID) {
	_ = a.Add(b).String()
	_ = a.Sub(b).String()
	_ = a.Mul(b).String()
	_ = a.Div(b).String()
	_ = a.Quantize(b).String()
	_ = a.Abs().String()
	_ = a.Negate().String()
	_ = a.PrettyString()
	_, _ = a.ConvertToInt64(RoundNearestEven)
	_, _ = a.ConvertToInt32(RoundTowardZero)
	_, _ = a.RoundIntegralNearestEven()
	_, _ = a.ScaleB(3)
	_ = a.IsNaN()
	_ = a.IsInf()
	_ = a.IsZero()
	_ = a.IsFinite()
	_ = a.IsSignaling()
}

func exerciseDecimal64NoPanic(a, b Decimal64BID) {
	_ = a.Add(b).String()
	_ = a.Sub(b).String()
	_ = a.Mul(b).String()
	_ = a.Div(b).String()
	_ = a.Quantize(b).String()
	_ = a.Abs().String()
	_ = a.Negate().String()
	_ = a.PrettyString()
	_, _ = a.ConvertToInt64(RoundNearestEven)
	_, _ = a.ConvertToInt32(RoundTowardZero)
	_, _ = a.RoundIntegralNearestEven()
	_, _ = a.ScaleB(3)
	_ = a.IsNaN()
	_ = a.IsInf()
	_ = a.IsZero()
	_ = a.IsFinite()
	_ = a.IsSignaling()
}

func exerciseDecimal128NoPanic(a, b Decimal128BID) {
	_ = a.Add(b).String()
	_ = a.Sub(b).String()
	_ = a.Mul(b).String()
	_ = a.Div(b).String()
	_ = a.Quantize(b).String()
	_ = a.Abs().String()
	_ = a.Negate().String()
	_ = a.PrettyString()
	_, _ = a.ConvertToInt64(RoundNearestEven)
	_, _ = a.ConvertToInt32(RoundTowardZero)
	_, _ = a.RoundIntegralNearestEven()
	_, _ = a.ScaleB(3)
	_ = a.IsNaN()
	_ = a.IsInf()
	_ = a.IsZero()
	_ = a.IsFinite()
	_ = a.IsSignaling()
}

func bytesToDecimal128(raw []byte) Decimal128BID {
	var b [16]byte
	copy(b[:], raw) // short raw zero-pads the tail; long raw keeps the first 16
	return Decimal128BID(b)
}

func uint64PairToDecimal128(hi, lo uint64) Decimal128BID {
	var b [16]byte
	binary.LittleEndian.PutUint64(b[0:8], lo)
	binary.LittleEndian.PutUint64(b[8:16], hi)
	return Decimal128BID(b)
}

// FuzzArithFromBitsNoPanic: constructing a BID value from an arbitrary raw bit
// pattern and then running the public arithmetic, quantize, string, convert,
// predicate, and rounding methods over it must never panic, whatever the bits
// (including non-canonical / reserved encodings that the string parse path
// could never produce). Failures report through ExceptionFlags or sentinel
// results, never through a panic/trap, so a panic is the only asserted failure.
func FuzzArithFromBitsNoPanic(f *testing.F) {
	seeds := []struct {
		a, b uint64
		raw  []byte
	}{
		{0, 0, make([]byte, 16)},
		{^uint64(0), ^uint64(0), []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
		{1, 2, nil},
		{0x8000000000000000, 0x7c00000000000000, []byte{0x01}},
		{uint64(bidNaNBits32), uint64(bidNaNBits64), make([]byte, 8)},
		{0x7800000000000000, 0x7e00000000000000, []byte{0x7c, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
		{0x31c0000000000001, 0x2feb29430a256d21, []byte{0x22, 0x08, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
	}
	for _, s := range seeds {
		f.Add(s.a, s.b, s.raw)
	}
	f.Fuzz(func(t *testing.T, a, b uint64, raw []byte) {
		// 32-bit: both halves of each 64-bit input feed the decimal32 surface.
		exerciseDecimal32NoPanic(Decimal32BID(uint32(a)), Decimal32BID(uint32(b)))
		exerciseDecimal32NoPanic(Decimal32BID(uint32(a>>32)), Decimal32BID(uint32(b>>32)))

		// 64-bit.
		exerciseDecimal64NoPanic(Decimal64BID(a), Decimal64BID(b))

		// 128-bit: one operand from the raw byte slice, one from the two u64s.
		exerciseDecimal128NoPanic(bytesToDecimal128(raw), uint64PairToDecimal128(a, b))
	})
}
