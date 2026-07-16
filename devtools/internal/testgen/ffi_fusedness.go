package testgen

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

const ffiProbeFusedness = "fusedness"

// ffiFusednessBits keeps the human-auditable BID word order separate from
// the generated FFI shard's raw little-endian byte string. Decimal128 rows
// are declared below as (hi, lo), while rawString emits the exact [16]byte
// representation consumed by both the Go C bridge and Rust from_le_bytes.
type ffiFusednessBits struct {
	width int
	hi    uint64
	lo    uint64
}

func ffiFusednessD(bits uint64) ffiFusednessBits {
	return ffiFusednessBits{width: 64, lo: bits}
}

func ffiFusednessQ(hi, lo uint64) ffiFusednessBits {
	return ffiFusednessBits{width: 128, hi: hi, lo: lo}
}

func (bits ffiFusednessBits) rawString() string {
	switch bits.width {
	case 64:
		return fmt.Sprintf("%016x", bits.lo)
	case 128:
		var raw [16]byte
		binary.LittleEndian.PutUint64(raw[0:8], bits.lo)
		binary.LittleEndian.PutUint64(raw[8:16], bits.hi)
		return hex.EncodeToString(raw[:])
	default:
		panic(fmt.Sprintf("unsupported fusedness width %d", bits.width))
	}
}

func (bits ffiFusednessBits) wordString() string {
	switch bits.width {
	case 64:
		return fmt.Sprintf("%016x", bits.lo)
	case 128:
		return fmt.Sprintf("%016x:%016x", bits.hi, bits.lo)
	default:
		panic(fmt.Sprintf("unsupported fusedness width %d", bits.width))
	}
}

type ffiFusednessOutcome struct {
	bits  ffiFusednessBits
	flags uint32
}

func (outcome ffiFusednessOutcome) ffiString() string {
	return fmt.Sprintf("%s/%08x", outcome.bits.rawString(), outcome.flags)
}

func (outcome ffiFusednessOutcome) wordString() string {
	return fmt.Sprintf("%s/%08x", outcome.bits.wordString(), outcome.flags)
}

type ffiFusednessProbe struct {
	function  string
	operands  [3]ffiFusednessBits
	rounding  int
	expected  ffiFusednessOutcome
	forbidden ffiFusednessOutcome
}

func (probe ffiFusednessProbe) ffiOperands() []string {
	return []string{
		probe.operands[0].rawString(),
		probe.operands[1].rawString(),
		probe.operands[2].rawString(),
	}
}

func (probe ffiFusednessProbe) row() string {
	return fmt.Sprintf("%s x=%s y=%s z=%s m=%d -> %s forbidden=%s",
		probe.function,
		probe.operands[0].wordString(),
		probe.operands[1].wordString(),
		probe.operands[2].wordString(),
		probe.rounding,
		probe.expected.wordString(),
		probe.forbidden.wordString(),
	)
}

func validateFFIFusednessProbe(probe ffiFusednessProbe, shape ffiMixedDecimalShape) error {
	if shape.operation != "fma" || shape.operandCount != 3 {
		return fmt.Errorf("fusedness probe %s targets non-FMA shape (%s, arity %d)", probe.function, shape.operation, shape.operandCount)
	}
	if probe.rounding < 0 || probe.rounding >= ffiRoundingModeCount {
		return fmt.Errorf("fusedness probe %s rounding mode %d is outside 0..%d", probe.function, probe.rounding, ffiRoundingModeCount-1)
	}
	for i, wantWidth := range shape.operandWidths() {
		if gotWidth := probe.operands[i].width; gotWidth != wantWidth {
			return fmt.Errorf("fusedness probe %s operand %d width = %d, want %d", probe.function, i, gotWidth, wantWidth)
		}
	}
	if probe.expected.bits.width != shape.resultBits {
		return fmt.Errorf("fusedness probe %s expected width = %d, want %d", probe.function, probe.expected.bits.width, shape.resultBits)
	}
	if probe.forbidden.bits.width != shape.resultBits {
		return fmt.Errorf("fusedness probe %s forbidden width = %d, want %d", probe.function, probe.forbidden.bits.width, shape.resultBits)
	}
	if probe.expected.ffiString() == probe.forbidden.ffiString() {
		return fmt.Errorf("fusedness probe %s expected and forbidden outcomes are identical", probe.function)
	}
	return nil
}

// ffiMixedFMAFusednessProbes is the closed mixed-FMA fusedness census. The
// 12 finite rows make the one-round result differ from sequential mul->add;
// bid128ddd/ddq use 0*Inf+qNaN because two Decimal64 multiplicands have an
// exact Decimal128 product, while fused special-value/flag semantics still
// distinguish FMA from sequential composition.
//
// The expected and forbidden values were audited against the pinned Intel
// BID C library. This generator-owned table is duplicated by hand in
// verification_sentinels.json through canonical row strings; no generator
// reads that pin file.
var ffiMixedFMAFusednessProbes = []ffiFusednessProbe{
	{
		function: "bid64ddq_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessD(0x31c0000000000001),
			ffiFusednessD(0x31c0000000000001),
			ffiFusednessQ(0x2ffc000000000000, 0x4563918244f40001),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68001), flags: 0x20},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68000), flags: 0x20},
	},
	{
		function: "bid64dqd_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessD(0x31c0000000000003),
			ffiFusednessQ(0x2ffca45894e48295, 0x7efb0aa216fc0001),
			ffiFusednessD(0x31c0000000000000),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68001), flags: 0x20},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68000), flags: 0x20},
	},
	{
		function: "bid64dqq_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessD(0x31c0000000000003),
			ffiFusednessQ(0x2ffca45894e48295, 0x7efb0aa216fc0001),
			ffiFusednessQ(0x3040000000000000, 0x0000000000000000),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68001), flags: 0x20},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68000), flags: 0x20},
	},
	{
		function: "bid64qdd_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessQ(0x2ffca45894e48295, 0x7efb0aa216fc0001),
			ffiFusednessD(0x31c0000000000003),
			ffiFusednessD(0x31c0000000000000),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68001), flags: 0x20},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68000), flags: 0x20},
	},
	{
		function: "bid64qdq_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessQ(0x2ffca45894e48295, 0x7efb0aa216fc0001),
			ffiFusednessD(0x31c0000000000003),
			ffiFusednessQ(0x3040000000000000, 0x0000000000000000),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68001), flags: 0x20},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68000), flags: 0x20},
	},
	{
		function: "bid64qqd_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessQ(0x2ffca45894e48295, 0x7efb0aa216fc0001),
			ffiFusednessQ(0x3040000000000000, 0x0000000000000003),
			ffiFusednessD(0x31c0000000000000),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68001), flags: 0x20},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68000), flags: 0x20},
	},
	{
		function: "bid64qqq_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessQ(0x2ffca45894e48295, 0x7efb0aa216fc0001),
			ffiFusednessQ(0x3040000000000000, 0x0000000000000003),
			ffiFusednessQ(0x3040000000000000, 0x0000000000000000),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68001), flags: 0x20},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessD(0x2fe38d7ea4c68000), flags: 0x20},
	},
	{
		function: "bid128ddd_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessD(0x0000000000000000),
			ffiFusednessD(0x7800000000000000),
			ffiFusednessD(0x7c00000000000001),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessQ(0x7c00000000000000, 0x0de0b6b3a7640000), flags: 0x00},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessQ(0x7c00000000000000, 0x0000000000000000), flags: 0x01},
	},
	{
		function: "bid128ddq_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessD(0x0000000000000000),
			ffiFusednessD(0x7800000000000000),
			ffiFusednessQ(0x7c00000000000000, 0x0000000000000001),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessQ(0x7c00000000000000, 0x0000000000000001), flags: 0x00},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessQ(0x7c00000000000000, 0x0000000000000000), flags: 0x01},
	},
	{
		function: "bid128dqd_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessD(0x2fe38d7ea4c68001),
			ffiFusednessQ(0x2ffded09bead87c0, 0x378d8e63ffffffff),
			ffiFusednessD(0xafe38d7ea4c68001),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessQ(0xafde000000000000, 0x00038d7ea4c68001), flags: 0x00},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessQ(0x2ffe000000000000, 0x0000000000000000), flags: 0x20},
	},
	{
		function: "bid128dqq_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessD(0x2fe38d7ea4c68001),
			ffiFusednessQ(0x2ffded09bead87c0, 0x378d8e63ffffffff),
			ffiFusednessQ(0xb022000000000000, 0x00038d7ea4c68001),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessQ(0xafde000000000000, 0x00038d7ea4c68001), flags: 0x00},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessQ(0x2ffe000000000000, 0x0000000000000000), flags: 0x20},
	},
	{
		function: "bid128qdd_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessQ(0x2ffded09bead87c0, 0x378d8e63ffffffff),
			ffiFusednessD(0x2fe38d7ea4c68001),
			ffiFusednessD(0xafe38d7ea4c68001),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessQ(0xafde000000000000, 0x00038d7ea4c68001), flags: 0x00},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessQ(0x2ffe000000000000, 0x0000000000000000), flags: 0x20},
	},
	{
		function: "bid128qdq_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessQ(0x2ffded09bead87c0, 0x378d8e63ffffffff),
			ffiFusednessD(0x2fe38d7ea4c68001),
			ffiFusednessQ(0xb022000000000000, 0x00038d7ea4c68001),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessQ(0xafde000000000000, 0x00038d7ea4c68001), flags: 0x00},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessQ(0x2ffe000000000000, 0x0000000000000000), flags: 0x20},
	},
	{
		function: "bid128qqd_fma",
		operands: [3]ffiFusednessBits{
			ffiFusednessQ(0x3022000000000000, 0x00038d7ea4c68001),
			ffiFusednessQ(0x2ffded09bead87c0, 0x378d8e63ffffffff),
			ffiFusednessD(0xafe38d7ea4c68001),
		},
		rounding:  0,
		expected:  ffiFusednessOutcome{bits: ffiFusednessQ(0xafde000000000000, 0x00038d7ea4c68001), flags: 0x00},
		forbidden: ffiFusednessOutcome{bits: ffiFusednessQ(0x2ffe000000000000, 0x0000000000000000), flags: 0x20},
	},
}

func ffiMixedFMAFusednessProbeFor(function string) (ffiFusednessProbe, bool) {
	for _, probe := range ffiMixedFMAFusednessProbes {
		if probe.function == function {
			return probe, true
		}
	}
	return ffiFusednessProbe{}, false
}

func ffiMixedFMAFusednessRows() []string {
	rows := make([]string, len(ffiMixedFMAFusednessProbes))
	for i, probe := range ffiMixedFMAFusednessProbes {
		rows[i] = probe.row()
	}
	return rows
}
