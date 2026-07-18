package bidcodec

// Standalone BID codec benchmark leg (Go). The operand set is the hand-pinned
// shared contract testdata/codec_benchmark_operands.json, consumed by all four
// codec benchmark legs (Go testing.B here, Rust criterion in
// bid754-codec-rs/benches/codec.rs, and the dependency-free JS/Python scripts
// bid754-codec-js/bench_runner.mjs and
// bid754-codec-py/benchmarks/bench_runner.py). Every operand must be a
// canonical, exactly representable value: loadCodecBenchmarkOperands rejects
// any entry whose decode/encode/toString/fromString round trips are not exact,
// so benchmark setup fails rather than timing invalid operands, and
// TestCodecBenchmarkOperandContract keeps that rejection as a checked-in test.
// This is benchmark infrastructure, not a regular verification domain.

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

const codecBenchmarkOperandPath = "testdata/codec_benchmark_operands.json"

type codecBenchmarkOperandEntry struct {
	Hex string `json:"hex"`
	// HexHi is a pointer so the shared cross-language guard semantics hold: a
	// decimal32/decimal64 entry must not carry any hex_hi string value (empty
	// included); an absent or null key is the only accepted shape.
	HexHi         *string `json:"hex_hi"`
	DecimalString string  `json:"decimal_string"`
}

type codecBenchmarkOperandFile struct {
	FormatVersion int                          `json:"format_version"`
	Decimal32     []codecBenchmarkOperandEntry `json:"decimal32"`
	Decimal64     []codecBenchmarkOperandEntry `json:"decimal64"`
	Decimal128    []codecBenchmarkOperandEntry `json:"decimal128"`
}

type codecBench32Operand struct {
	bits uint32
	comp Components
	str  string
}

type codecBench64Operand struct {
	bits uint64
	comp Components
	str  string
}

type codecBench128Operand struct {
	lo   uint64
	hi   uint64
	comp Components
	str  string
}

type codecBenchmarkOperands struct {
	decimal32  []codecBench32Operand
	decimal64  []codecBench64Operand
	decimal128 []codecBench128Operand
}

func parseCodecBenchmarkHexWord(tb testing.TB, width string, index int, field, hex string, bits int) uint64 {
	tb.Helper()
	if len(hex) != bits/4 {
		tb.Fatalf("%s[%d].%s: hex %q must be exactly %d lowercase hex digits", width, index, field, hex, bits/4)
	}
	// Lowercase-only, matching the JS/Python legs exactly, so all four legs
	// accept and reject the same contract file (strconv alone would also take
	// uppercase digits).
	for _, b := range []byte(hex) {
		if !(b >= '0' && b <= '9' || b >= 'a' && b <= 'f') {
			tb.Fatalf("%s[%d].%s: hex %q must be exactly %d lowercase hex digits", width, index, field, hex, bits/4)
		}
	}
	v, err := strconv.ParseUint(hex, 16, bits)
	if err != nil {
		tb.Fatalf("%s[%d].%s: invalid hex %q: %v", width, index, field, hex, err)
	}
	return v
}

// loadCodecBenchmarkOperands loads the shared benchmark operand contract and
// rejects (via tb.Fatalf) any operand that is not canonical and exactly
// representable in both the binary and the string channel.
func loadCodecBenchmarkOperands(tb testing.TB) codecBenchmarkOperands {
	tb.Helper()
	raw, err := os.ReadFile(codecBenchmarkOperandPath)
	if err != nil {
		tb.Fatalf("read benchmark operand contract: %v", err)
	}
	var file codecBenchmarkOperandFile
	if err := json.Unmarshal(raw, &file); err != nil {
		tb.Fatalf("parse benchmark operand contract: %v", err)
	}
	if file.FormatVersion != 1 {
		tb.Fatalf("benchmark operand contract format_version = %d, want 1", file.FormatVersion)
	}
	if len(file.Decimal32) == 0 || len(file.Decimal64) == 0 || len(file.Decimal128) == 0 {
		tb.Fatalf("benchmark operand contract requires non-empty decimal32/decimal64/decimal128 entry lists")
	}

	var ops codecBenchmarkOperands
	for i, entry := range file.Decimal32 {
		bits := uint32(parseCodecBenchmarkHexWord(tb, "decimal32", i, "hex", entry.Hex, 32))
		if entry.HexHi != nil {
			tb.Fatalf("decimal32[%d]: hex_hi is only valid for decimal128 entries", i)
		}
		comp := Decode32(bits)
		if comp.Kind != Normal {
			tb.Fatalf("decimal32[%d] %q: benchmark operands must decode to a Normal value, got kind %d", i, entry.Hex, comp.Kind)
		}
		encoded, err := Encode32(comp)
		if err != nil {
			tb.Fatalf("decimal32[%d] %q: re-encode failed: %v", i, entry.Hex, err)
		}
		if encoded != bits {
			tb.Fatalf("decimal32[%d] %q: not canonical, re-encodes to %08x", i, entry.Hex, encoded)
		}
		rendered, err := ToString(comp)
		if err != nil {
			tb.Fatalf("decimal32[%d] %q: ToString failed: %v", i, entry.Hex, err)
		}
		if rendered != entry.DecimalString {
			tb.Fatalf("decimal32[%d] %q: canonical string %q, contract pins %q", i, entry.Hex, rendered, entry.DecimalString)
		}
		parsed, err := FromString(entry.DecimalString)
		if err != nil {
			tb.Fatalf("decimal32[%d] %q: FromString failed: %v", i, entry.DecimalString, err)
		}
		reencoded, err := Encode32(parsed)
		if err != nil {
			tb.Fatalf("decimal32[%d] %q: string operand is not exactly representable in decimal32: %v", i, entry.DecimalString, err)
		}
		if reencoded != bits {
			tb.Fatalf("decimal32[%d] %q: string round trip drifts to %08x, want %08x", i, entry.DecimalString, reencoded, bits)
		}
		ops.decimal32 = append(ops.decimal32, codecBench32Operand{bits: bits, comp: comp, str: entry.DecimalString})
	}
	for i, entry := range file.Decimal64 {
		bits := parseCodecBenchmarkHexWord(tb, "decimal64", i, "hex", entry.Hex, 64)
		if entry.HexHi != nil {
			tb.Fatalf("decimal64[%d]: hex_hi is only valid for decimal128 entries", i)
		}
		comp := Decode64(bits)
		if comp.Kind != Normal {
			tb.Fatalf("decimal64[%d] %q: benchmark operands must decode to a Normal value, got kind %d", i, entry.Hex, comp.Kind)
		}
		encoded, err := Encode64(comp)
		if err != nil {
			tb.Fatalf("decimal64[%d] %q: re-encode failed: %v", i, entry.Hex, err)
		}
		if encoded != bits {
			tb.Fatalf("decimal64[%d] %q: not canonical, re-encodes to %016x", i, entry.Hex, encoded)
		}
		rendered, err := ToString(comp)
		if err != nil {
			tb.Fatalf("decimal64[%d] %q: ToString failed: %v", i, entry.Hex, err)
		}
		if rendered != entry.DecimalString {
			tb.Fatalf("decimal64[%d] %q: canonical string %q, contract pins %q", i, entry.Hex, rendered, entry.DecimalString)
		}
		parsed, err := FromString(entry.DecimalString)
		if err != nil {
			tb.Fatalf("decimal64[%d] %q: FromString failed: %v", i, entry.DecimalString, err)
		}
		reencoded, err := Encode64(parsed)
		if err != nil {
			tb.Fatalf("decimal64[%d] %q: string operand is not exactly representable in decimal64: %v", i, entry.DecimalString, err)
		}
		if reencoded != bits {
			tb.Fatalf("decimal64[%d] %q: string round trip drifts to %016x, want %016x", i, entry.DecimalString, reencoded, bits)
		}
		ops.decimal64 = append(ops.decimal64, codecBench64Operand{bits: bits, comp: comp, str: entry.DecimalString})
	}
	for i, entry := range file.Decimal128 {
		lo := parseCodecBenchmarkHexWord(tb, "decimal128", i, "hex", entry.Hex, 64)
		if entry.HexHi == nil {
			tb.Fatalf("decimal128[%d]: missing hex_hi", i)
		}
		hiHex := *entry.HexHi
		hi := parseCodecBenchmarkHexWord(tb, "decimal128", i, "hex_hi", hiHex, 64)
		comp := Decode128(lo, hi)
		if comp.Kind != Normal {
			tb.Fatalf("decimal128[%d] %q/%q: benchmark operands must decode to a Normal value, got kind %d", i, entry.Hex, hiHex, comp.Kind)
		}
		encodedLo, encodedHi, err := Encode128(comp)
		if err != nil {
			tb.Fatalf("decimal128[%d] %q/%q: re-encode failed: %v", i, entry.Hex, hiHex, err)
		}
		if encodedLo != lo || encodedHi != hi {
			tb.Fatalf("decimal128[%d] %q/%q: not canonical, re-encodes to %016x/%016x", i, entry.Hex, hiHex, encodedLo, encodedHi)
		}
		rendered, err := ToString(comp)
		if err != nil {
			tb.Fatalf("decimal128[%d] %q/%q: ToString failed: %v", i, entry.Hex, hiHex, err)
		}
		if rendered != entry.DecimalString {
			tb.Fatalf("decimal128[%d] %q/%q: canonical string %q, contract pins %q", i, entry.Hex, hiHex, rendered, entry.DecimalString)
		}
		parsed, err := FromString(entry.DecimalString)
		if err != nil {
			tb.Fatalf("decimal128[%d] %q: FromString failed: %v", i, entry.DecimalString, err)
		}
		reencodedLo, reencodedHi, err := Encode128(parsed)
		if err != nil {
			tb.Fatalf("decimal128[%d] %q: string operand is not exactly representable in decimal128: %v", i, entry.DecimalString, err)
		}
		if reencodedLo != lo || reencodedHi != hi {
			tb.Fatalf("decimal128[%d] %q: string round trip drifts to %016x/%016x, want %016x/%016x", i, entry.DecimalString, reencodedLo, reencodedHi, lo, hi)
		}
		ops.decimal128 = append(ops.decimal128, codecBench128Operand{lo: lo, hi: hi, comp: comp, str: entry.DecimalString})
	}
	return ops
}

// TestCodecBenchmarkOperandContract pins the shared benchmark operand
// contract: it fails when any checked-in operand stops being canonical or
// exactly representable, so a broken contract file cannot silently feed the
// four language benchmark legs.
func TestCodecBenchmarkOperandContract(t *testing.T) {
	ops := loadCodecBenchmarkOperands(t)
	if len(ops.decimal32) == 0 || len(ops.decimal64) == 0 || len(ops.decimal128) == 0 {
		t.Fatalf("benchmark operand contract loaded empty operand lists")
	}
}

var (
	codecBenchSinkComponents Components
	codecBenchSink32         uint32
	codecBenchSink64         uint64
	codecBenchSink128Lo      uint64
	codecBenchSink128Hi      uint64
	codecBenchSinkString     string
	codecBenchSinkErr        error
)

func BenchmarkCodecBID32(b *testing.B) {
	entries := loadCodecBenchmarkOperands(b).decimal32
	n := len(entries)
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSinkComponents = Decode32(entries[i%n].bits)
		}
	})
	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSink32, codecBenchSinkErr = Encode32(entries[i%n].comp)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSinkString, codecBenchSinkErr = ToString(entries[i%n].comp)
		}
	})
	b.Run("from_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSinkComponents, codecBenchSinkErr = FromString(entries[i%n].str)
		}
	})
}

func BenchmarkCodecBID64(b *testing.B) {
	entries := loadCodecBenchmarkOperands(b).decimal64
	n := len(entries)
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSinkComponents = Decode64(entries[i%n].bits)
		}
	})
	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSink64, codecBenchSinkErr = Encode64(entries[i%n].comp)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSinkString, codecBenchSinkErr = ToString(entries[i%n].comp)
		}
	})
	b.Run("from_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSinkComponents, codecBenchSinkErr = FromString(entries[i%n].str)
		}
	})
}

func BenchmarkCodecBID128(b *testing.B) {
	entries := loadCodecBenchmarkOperands(b).decimal128
	n := len(entries)
	b.Run("decode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSinkComponents = Decode128(entries[i%n].lo, entries[i%n].hi)
		}
	})
	b.Run("encode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSink128Lo, codecBenchSink128Hi, codecBenchSinkErr = Encode128(entries[i%n].comp)
		}
	})
	b.Run("to_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSinkString, codecBenchSinkErr = ToString(entries[i%n].comp)
		}
	})
	b.Run("from_string", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			codecBenchSinkComponents, codecBenchSinkErr = FromString(entries[i%n].str)
		}
	})
}
