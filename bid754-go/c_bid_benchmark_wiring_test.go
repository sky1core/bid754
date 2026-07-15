//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import (
	"encoding/binary"
	"fmt"
	"strings"
	"testing"
)

const intelCBenchmarkPoisonFlags = ^uint32(0)

type intelCBenchmarkOracleInputs struct {
	x32        Decimal32BID
	y32        Decimal32BID
	z32        Decimal32BID
	integer32  Decimal32BID
	x64        Decimal64BID
	y64        Decimal64BID
	z64        Decimal64BID
	integer64  Decimal64BID
	x128       Decimal128BID
	y128       Decimal128BID
	z128       Decimal128BID
	integer128 Decimal128BID
	inputs     benchmarkInputs
}

type intelCBenchmarkOracleResult struct {
	result      intelCBenchmarkResultKind
	snapshot    nativeCBenchmarkSnapshot
	hasFlags    bool
	stringWidth int
}

type intelCBenchmarkWiringFixture struct {
	name   string
	inputs benchmarkInputs
}

func newIntelCBenchmarkOracleInputs(t *testing.T, fixture intelCBenchmarkWiringFixture) intelCBenchmarkOracleInputs {
	t.Helper()
	inputs := fixture.inputs
	validateBenchmarkInputs(t, inputs)
	integer32, flags32 := NewDecimal32FromInt64(inputs.IntegerOperand, RoundNearestEven)
	if flags32 != 0 {
		t.Fatalf("Decimal32 integer benchmark oracle input raised flags: %v", flags32)
	}
	integer64, flags64 := NewDecimal64FromInt64(inputs.IntegerOperand, RoundNearestEven)
	if flags64 != 0 {
		t.Fatalf("Decimal64 integer benchmark oracle input raised flags: %v", flags64)
	}
	return intelCBenchmarkOracleInputs{
		x32:        exactBenchmarkDecimal32(t, inputs.Decimal32.X),
		y32:        exactBenchmarkDecimal32(t, inputs.Decimal32.Y),
		z32:        exactBenchmarkDecimal32(t, inputs.Decimal32.Z),
		integer32:  integer32,
		x64:        exactBenchmarkDecimal64(t, inputs.Decimal64.X),
		y64:        exactBenchmarkDecimal64(t, inputs.Decimal64.Y),
		z64:        exactBenchmarkDecimal64(t, inputs.Decimal64.Z),
		integer64:  integer64,
		x128:       exactBenchmarkDecimal128(t, inputs.Decimal128.X),
		y128:       exactBenchmarkDecimal128(t, inputs.Decimal128.Y),
		z128:       exactBenchmarkDecimal128(t, inputs.Decimal128.Z),
		integer128: NewDecimal128FromInt64(inputs.IntegerOperand),
		inputs:     inputs,
	}
}

func intelCBenchmarkWiringFixtures(t *testing.T) []intelCBenchmarkWiringFixture {
	t.Helper()
	production := loadBenchmarkInputs(t)
	return []intelCBenchmarkWiringFixture{
		{
			name:   "production",
			inputs: production,
		},
		{
			name: "x_less_than_y",
			inputs: benchmarkInputs{
				FormatVersion:  2,
				IntegerOperand: production.IntegerOperand,
				ScaleExponent:  production.ScaleExponent,
				Decimal32:      benchmarkInputPair{X: "4.25", Y: "33", Z: "2.5"},
				Decimal64:      benchmarkInputPair{X: "7.125", Y: "55", Z: "3.25"},
				Decimal128:     benchmarkInputPair{X: "11.0625", Y: "96", Z: "5.5"},
			},
		},
		{
			name: "x_greater_than_y",
			inputs: benchmarkInputs{
				FormatVersion:  2,
				IntegerOperand: production.IntegerOperand,
				ScaleExponent:  production.ScaleExponent,
				Decimal32:      benchmarkInputPair{X: "33", Y: "4.25", Z: "2.5"},
				Decimal64:      benchmarkInputPair{X: "55", Y: "7.125", Z: "3.25"},
				Decimal128:     benchmarkInputPair{X: "96", Y: "11.0625", Z: "5.5"},
			},
		},
	}
}

func initIntelCBenchmarkOracleNativeInputs(t *testing.T, inputs benchmarkInputs) {
	t.Helper()
	if !nativeBenchCBIDInit(
		inputs.Decimal32.X,
		inputs.Decimal32.Y,
		inputs.Decimal32.Z,
		inputs.Decimal64.X,
		inputs.Decimal64.Y,
		inputs.Decimal64.Z,
		inputs.Decimal128.X,
		inputs.Decimal128.Y,
		inputs.Decimal128.Z,
		inputs.IntegerOperand,
		inputs.ScaleExponent,
	) {
		t.Fatal("Intel C benchmark oracle inputs are not finite and exact")
	}
}

func intelCBenchmarkRawFlags(t *testing.T, flags ExceptionFlags) uint32 {
	t.Helper()
	var raw uint32
	if flags&FlagInvalidOperation != 0 {
		raw |= 0x01
	}
	if flags&FlagDivisionByZero != 0 {
		raw |= 0x04
	}
	if flags&FlagOverflow != 0 {
		raw |= 0x08
	}
	if flags&FlagUnderflow != 0 {
		raw |= 0x10
	}
	if flags&FlagInexact != 0 {
		raw |= 0x20
	}
	known := FlagInvalidOperation | FlagDivisionByZero | FlagOverflow | FlagUnderflow | FlagInexact
	if unknown := flags &^ known; unknown != 0 {
		t.Fatalf("benchmark oracle produced flags outside the Intel raw set: %s", unknown)
	}
	return raw
}

func intelCBenchmarkOracleBID32(t *testing.T, value Decimal32BID, flags ExceptionFlags) intelCBenchmarkOracleResult {
	t.Helper()
	return intelCBenchmarkOracleResult{
		result:   intelCBenchmarkBID32,
		snapshot: nativeCBenchmarkSnapshot{BID32: value.ToUint32(), Flags: intelCBenchmarkRawFlags(t, flags)},
		hasFlags: true,
	}
}

func intelCBenchmarkOracleBID64(t *testing.T, value Decimal64BID, flags ExceptionFlags) intelCBenchmarkOracleResult {
	t.Helper()
	return intelCBenchmarkOracleResult{
		result:   intelCBenchmarkBID64,
		snapshot: nativeCBenchmarkSnapshot{BID64: value.ToUint64(), Flags: intelCBenchmarkRawFlags(t, flags)},
		hasFlags: true,
	}
}

func intelCBenchmarkOracleBID128(t *testing.T, value Decimal128BID, flags ExceptionFlags) intelCBenchmarkOracleResult {
	t.Helper()
	bits := value.ToBytes()
	return intelCBenchmarkOracleResult{
		result: intelCBenchmarkBID128,
		snapshot: nativeCBenchmarkSnapshot{
			BID128Low:  binary.LittleEndian.Uint64(bits[0:8]),
			BID128High: binary.LittleEndian.Uint64(bits[8:16]),
			Flags:      intelCBenchmarkRawFlags(t, flags),
		},
		hasFlags: true,
	}
}

func intelCBenchmarkOracleBID128NoFlags(value Decimal128BID) intelCBenchmarkOracleResult {
	bits := value.ToBytes()
	return intelCBenchmarkOracleResult{
		result: intelCBenchmarkBID128,
		snapshot: nativeCBenchmarkSnapshot{
			BID128Low:  binary.LittleEndian.Uint64(bits[0:8]),
			BID128High: binary.LittleEndian.Uint64(bits[8:16]),
		},
	}
}

func intelCBenchmarkOracleBool(t *testing.T, result intelCBenchmarkResultKind, value bool, flags ExceptionFlags) intelCBenchmarkOracleResult {
	t.Helper()
	var rawValue uint64
	if value {
		rawValue = 1
	}
	want := intelCBenchmarkOracleResult{
		result:   result,
		snapshot: nativeCBenchmarkSnapshot{Flags: intelCBenchmarkRawFlags(t, flags)},
		hasFlags: true,
	}
	if result == intelCBenchmarkBool32 {
		want.snapshot.BID32 = uint32(rawValue)
	} else {
		want.snapshot.BID64 = rawValue
	}
	return want
}

func intelCBenchmarkOracleInt64(t *testing.T, value int64, flags ExceptionFlags) intelCBenchmarkOracleResult {
	t.Helper()
	return intelCBenchmarkOracleResult{
		result:   intelCBenchmarkInt64,
		snapshot: nativeCBenchmarkSnapshot{Int64: value, Flags: intelCBenchmarkRawFlags(t, flags)},
		hasFlags: true,
	}
}

func intelCBenchmarkOracleString32(value Decimal32BID) intelCBenchmarkOracleResult {
	return intelCBenchmarkOracleResult{
		result:      intelCBenchmarkString,
		snapshot:    nativeCBenchmarkSnapshot{BID32: value.ToUint32()},
		hasFlags:    true,
		stringWidth: 32,
	}
}

func intelCBenchmarkOracleString64(value Decimal64BID) intelCBenchmarkOracleResult {
	return intelCBenchmarkOracleResult{
		result:      intelCBenchmarkString,
		snapshot:    nativeCBenchmarkSnapshot{BID64: value.ToUint64()},
		hasFlags:    true,
		stringWidth: 64,
	}
}

func intelCBenchmarkOracleString128(value Decimal128BID) intelCBenchmarkOracleResult {
	bits := value.ToBytes()
	return intelCBenchmarkOracleResult{
		result: intelCBenchmarkString,
		snapshot: nativeCBenchmarkSnapshot{
			BID128Low:  binary.LittleEndian.Uint64(bits[0:8]),
			BID128High: binary.LittleEndian.Uint64(bits[8:16]),
		},
		hasFlags:    true,
		stringWidth: 128,
	}
}

func intelCBenchmarkOracle(t *testing.T, group, name string, o intelCBenchmarkOracleInputs) intelCBenchmarkOracleResult {
	t.Helper()
	switch group {
	case "bid32":
		switch name {
		case "add":
			value, flags := o.x32.AddWithFlags(o.y32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "mul":
			value, flags := o.x32.MulWithFlags(o.y32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "sub":
			value, flags := o.x32.SubWithFlags(o.y32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "div":
			value, flags := o.x32.DivWithFlags(o.y32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "fma":
			value, flags := o.x32.FMA(o.y32, o.z32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "sqrt":
			value, flags := o.x32.Sqrt()
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "remainder":
			value, flags := o.y32.Remainder(o.x32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "fmod":
			value, flags := o.y32.Fmod(o.x32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "quantize":
			value, flags := o.x32.QuantizeWithFlags(o.integer32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "scaleb":
			value, flags := o.x32.ScaleB(o.inputs.ScaleExponent)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "quiet_equal":
			value, flags := o.x32.QuietEqual(o.y32)
			return intelCBenchmarkOracleBool(t, intelCBenchmarkBool32, value, flags)
		case "minnum":
			value, flags := o.x32.MinNum(o.y32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "maxnum":
			value, flags := o.x32.MaxNum(o.y32)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "from_int64":
			value, flags := NewDecimal32FromInt64(o.inputs.IntegerOperand, RoundNearestEven)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "to_int64":
			value, flags := o.integer32.ConvertToInt64(RoundNearestEven)
			return intelCBenchmarkOracleInt64(t, value, flags)
		case "to_decimal64":
			value, flags := o.x32.ToDecimal64()
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "to_decimal128":
			value, flags := o.x32.ToDecimal128()
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "parse":
			value, err := NewDecimal32BIDDirect(o.inputs.Decimal32.X)
			if err != nil {
				t.Fatalf("Decimal32 benchmark parse oracle: %v", err)
			}
			return intelCBenchmarkOracleBID32(t, value, 0)
		case "to_string":
			return intelCBenchmarkOracleString32(o.x32)
		}
	case "bid64":
		switch name {
		case "add":
			value, flags := o.x64.AddWithFlags(o.y64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "mul":
			value, flags := o.x64.MulWithFlags(o.y64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "sub":
			value, flags := o.x64.SubWithFlags(o.y64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "div":
			value, flags := o.x64.DivWithFlags(o.y64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "fma":
			value, flags := o.x64.FMA(o.y64, o.z64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "sqrt":
			value, flags := o.x64.Sqrt()
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "remainder":
			value, flags := o.y64.Remainder(o.x64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "fmod":
			value, flags := o.y64.Fmod(o.x64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "quantize":
			value, flags := o.x64.QuantizeWithFlags(o.integer64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "scaleb":
			value, flags := o.x64.ScaleB(o.inputs.ScaleExponent)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "quiet_equal":
			value, flags := o.x64.QuietEqual(o.y64)
			return intelCBenchmarkOracleBool(t, intelCBenchmarkBool64, value, flags)
		case "minnum":
			value, flags := o.x64.MinNum(o.y64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "maxnum":
			value, flags := o.x64.MaxNum(o.y64)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "from_int64":
			value, flags := NewDecimal64FromInt64(o.inputs.IntegerOperand, RoundNearestEven)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "to_int64":
			value, flags := o.integer64.ConvertToInt64(RoundNearestEven)
			return intelCBenchmarkOracleInt64(t, value, flags)
		case "to_decimal32":
			value, flags := o.x64.ToDecimal32(RoundNearestEven)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "to_decimal128":
			value, flags := o.x64.ToDecimal128()
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "parse":
			value, err := NewDecimal64BIDDirect(o.inputs.Decimal64.X)
			if err != nil {
				t.Fatalf("Decimal64 benchmark parse oracle: %v", err)
			}
			return intelCBenchmarkOracleBID64(t, value, 0)
		case "to_string":
			return intelCBenchmarkOracleString64(o.x64)
		}
	case "bid128":
		switch name {
		case "add":
			value, flags := o.x128.AddWithFlags(o.y128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "mul":
			value, flags := o.x128.MulWithFlags(o.y128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "sub":
			value, flags := o.x128.SubWithFlags(o.y128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "div":
			value, flags := o.x128.DivWithFlags(o.y128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "fma":
			value, flags := o.x128.FMA(o.y128, o.z128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "sqrt":
			value, flags := o.x128.Sqrt()
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "remainder":
			value, flags := o.y128.Remainder(o.x128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "fmod":
			value, flags := o.y128.Fmod(o.x128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "quantize":
			value, flags := o.x128.QuantizeWithFlags(o.integer128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "scaleb":
			value, flags := o.x128.ScaleB(o.inputs.ScaleExponent)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "quiet_equal":
			value, flags := o.x128.QuietEqual(o.y128)
			return intelCBenchmarkOracleBool(t, intelCBenchmarkBool64, value, flags)
		case "minnum":
			value, flags := o.x128.MinNum(o.y128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "maxnum":
			value, flags := o.x128.MaxNum(o.y128)
			return intelCBenchmarkOracleBID128(t, value, flags)
		case "from_int64":
			return intelCBenchmarkOracleBID128NoFlags(NewDecimal128FromInt64(o.inputs.IntegerOperand))
		case "to_int64":
			value, flags := o.integer128.ConvertToInt64(RoundNearestEven)
			return intelCBenchmarkOracleInt64(t, value, flags)
		case "to_decimal32":
			value, flags := o.x128.ToDecimal32(RoundNearestEven)
			return intelCBenchmarkOracleBID32(t, value, flags)
		case "to_decimal64":
			value, flags := o.x128.ToDecimal64(RoundNearestEven)
			return intelCBenchmarkOracleBID64(t, value, flags)
		case "parse":
			value, err := NewDecimal128BIDDirect(o.inputs.Decimal128.X)
			if err != nil {
				t.Fatalf("Decimal128 benchmark parse oracle: %v", err)
			}
			return intelCBenchmarkOracleBID128(t, value, 0)
		case "to_string":
			return intelCBenchmarkOracleString128(o.x128)
		}
	case "mixed_bid64":
		return intelCBenchmarkMixed64Oracle(t, name, o)
	case "mixed_bid128":
		return intelCBenchmarkMixed128Oracle(t, name, o)
	}
	t.Fatalf("unknown Intel C benchmark oracle row %s/%s", group, name)
	return intelCBenchmarkOracleResult{}
}

func intelCBenchmarkMixed64Oracle(t *testing.T, name string, o intelCBenchmarkOracleInputs) intelCBenchmarkOracleResult {
	t.Helper()
	var value Decimal64BID
	var flags ExceptionFlags
	switch name {
	case "dq_add":
		value, flags = Add64DQBIDWithMode(o.x64, o.y128, RoundNearestEven)
	case "qd_add":
		value, flags = Add64QDBIDWithMode(o.x128, o.y64, RoundNearestEven)
	case "qq_add":
		value, flags = Add64QQBIDWithMode(o.x128, o.y128, RoundNearestEven)
	case "dq_sub":
		value, flags = Sub64DQBIDWithMode(o.x64, o.y128, RoundNearestEven)
	case "qd_sub":
		value, flags = Sub64QDBIDWithMode(o.x128, o.y64, RoundNearestEven)
	case "qq_sub":
		value, flags = Sub64QQBIDWithMode(o.x128, o.y128, RoundNearestEven)
	case "dq_mul":
		value, flags = Mul64DQBIDWithMode(o.x64, o.y128, RoundNearestEven)
	case "qd_mul":
		value, flags = Mul64QDBIDWithMode(o.x128, o.y64, RoundNearestEven)
	case "qq_mul":
		value, flags = Mul64QQBIDWithMode(o.x128, o.y128, RoundNearestEven)
	case "dq_div":
		value, flags = Div64DQBIDWithMode(o.x64, o.y128, RoundNearestEven)
	case "qd_div":
		value, flags = Div64QDBIDWithMode(o.x128, o.y64, RoundNearestEven)
	case "qq_div":
		value, flags = Div64QQBIDWithMode(o.x128, o.y128, RoundNearestEven)
	default:
		t.Fatalf("unknown Intel C mixed Decimal64 benchmark oracle row %s", name)
	}
	return intelCBenchmarkOracleBID64(t, value, flags)
}

func intelCBenchmarkMixed128Oracle(t *testing.T, name string, o intelCBenchmarkOracleInputs) intelCBenchmarkOracleResult {
	t.Helper()
	var value Decimal128BID
	var flags ExceptionFlags
	switch name {
	case "dd_add":
		value, flags = Add128DDBIDWithMode(o.x64, o.y64, RoundNearestEven)
	case "dq_add":
		value, flags = Add128DQBIDWithMode(o.x64, o.y128, RoundNearestEven)
	case "qd_add":
		value, flags = Add128QDBIDWithMode(o.x128, o.y64, RoundNearestEven)
	case "dd_sub":
		value, flags = Sub128DDBIDWithMode(o.x64, o.y64, RoundNearestEven)
	case "dq_sub":
		value, flags = Sub128DQBIDWithMode(o.x64, o.y128, RoundNearestEven)
	case "qd_sub":
		value, flags = Sub128QDBIDWithMode(o.x128, o.y64, RoundNearestEven)
	case "dd_mul":
		value, flags = Mul128DDBIDWithMode(o.x64, o.y64, RoundNearestEven)
	case "dq_mul":
		value, flags = Mul128DQBIDWithMode(o.x64, o.y128, RoundNearestEven)
	case "qd_mul":
		value, flags = Mul128QDBIDWithMode(o.x128, o.y64, RoundNearestEven)
	case "dd_div":
		value, flags = Div128DDBIDWithMode(o.x64, o.y64, RoundNearestEven)
	case "dq_div":
		value, flags = Div128DQBIDWithMode(o.x64, o.y128, RoundNearestEven)
	case "qd_div":
		value, flags = Div128QDBIDWithMode(o.x128, o.y64, RoundNearestEven)
	default:
		t.Fatalf("unknown Intel C mixed Decimal128 benchmark oracle row %s", name)
	}
	return intelCBenchmarkOracleBID128(t, value, flags)
}

func checkIntelCBenchmarkResult(t *testing.T, row intelCBenchmarkRow, got nativeCBenchmarkSnapshot, want intelCBenchmarkOracleResult) {
	t.Helper()
	if row.result != want.result {
		t.Fatalf("result kind = %d, oracle kind = %d", row.result, want.result)
	}
	if row.hasFlags != want.hasFlags {
		t.Fatalf("hasFlags = %v, oracle hasFlags = %v", row.hasFlags, want.hasFlags)
	}
	switch row.result {
	case intelCBenchmarkBID32, intelCBenchmarkBool32:
		if got.BID32 != want.snapshot.BID32 {
			t.Errorf("result = %#08x, want %#08x", got.BID32, want.snapshot.BID32)
		}
	case intelCBenchmarkBID64, intelCBenchmarkBool64:
		if got.BID64 != want.snapshot.BID64 {
			t.Errorf("result = %#016x, want %#016x", got.BID64, want.snapshot.BID64)
		}
	case intelCBenchmarkBID128:
		if got.BID128Low != want.snapshot.BID128Low || got.BID128High != want.snapshot.BID128High {
			t.Errorf("result = %016x:%016x, want %016x:%016x", got.BID128High, got.BID128Low, want.snapshot.BID128High, want.snapshot.BID128Low)
		}
	case intelCBenchmarkInt64:
		if got.Int64 != want.snapshot.Int64 {
			t.Errorf("result = %d, want %d", got.Int64, want.snapshot.Int64)
		}
	case intelCBenchmarkString:
		checkIntelCBenchmarkStringResult(t, got.String, want)
	default:
		t.Fatalf("unknown result kind %d", row.result)
	}
	if row.hasFlags {
		if got.Flags != want.snapshot.Flags {
			t.Errorf("flags = %#08x, want %#08x", got.Flags, want.snapshot.Flags)
		}
	} else if got.Flags != intelCBenchmarkPoisonFlags {
		t.Errorf("flags = %#08x, want unchanged poison %#08x", got.Flags, intelCBenchmarkPoisonFlags)
	}
}

func checkIntelCBenchmarkStringResult(t *testing.T, got string, want intelCBenchmarkOracleResult) {
	t.Helper()
	if got == "" || got == "<unset>" {
		t.Fatalf("to_string result was not captured: %q", got)
	}
	var oracle string
	switch want.stringWidth {
	case 32:
		oracle = nativeBenchCBID32ToStringOnce()
		value, err := NewDecimal32BIDDirect(got)
		if err != nil {
			t.Fatalf("parse Intel C Decimal32 to_string result %q: %v", got, err)
		}
		if value.ToUint32() != want.snapshot.BID32 {
			t.Errorf("to_string result %q round-tripped to %#08x, want %#08x", got, value.ToUint32(), want.snapshot.BID32)
		}
	case 64:
		oracle = nativeBenchCBID64ToStringOnce()
		value, err := NewDecimal64BIDDirect(got)
		if err != nil {
			t.Fatalf("parse Intel C Decimal64 to_string result %q: %v", got, err)
		}
		if value.ToUint64() != want.snapshot.BID64 {
			t.Errorf("to_string result %q round-tripped to %#016x, want %#016x", got, value.ToUint64(), want.snapshot.BID64)
		}
	case 128:
		oracle = nativeBenchCBID128ToStringOnce()
		value, err := NewDecimal128BIDDirect(got)
		if err != nil {
			t.Fatalf("parse Intel C Decimal128 to_string result %q: %v", got, err)
		}
		bits := value.ToBytes()
		low := binary.LittleEndian.Uint64(bits[0:8])
		high := binary.LittleEndian.Uint64(bits[8:16])
		if low != want.snapshot.BID128Low || high != want.snapshot.BID128High {
			t.Errorf("to_string result %q round-tripped to %016x:%016x, want %016x:%016x", got, high, low, want.snapshot.BID128High, want.snapshot.BID128Low)
		}
	default:
		t.Fatalf("unknown to_string width %d", want.stringWidth)
	}
	if got != oracle {
		t.Errorf("to_string result = %q, one-shot C oracle = %q", got, oracle)
	}
}

type intelCBenchmarkGroup struct {
	name string
	rows []intelCBenchmarkRow
	want int
}

func intelCBenchmarkGroups() []intelCBenchmarkGroup {
	return []intelCBenchmarkGroup{
		{"bid32", intelCBID32BenchmarkRows, 19},
		{"bid64", intelCBID64BenchmarkRows, 19},
		{"bid128", intelCBID128BenchmarkRows, 19},
		{"mixed_bid64", intelCMixedBID64BenchmarkRows, 12},
		{"mixed_bid128", intelCMixedBID128BenchmarkRows, 12},
	}
}

func intelCBenchmarkOracleFingerprint(want intelCBenchmarkOracleResult) string {
	var result string
	switch want.result {
	case intelCBenchmarkBID32, intelCBenchmarkBool32:
		result = fmt.Sprintf("%08x", want.snapshot.BID32)
	case intelCBenchmarkBID64, intelCBenchmarkBool64:
		result = fmt.Sprintf("%016x", want.snapshot.BID64)
	case intelCBenchmarkBID128:
		result = fmt.Sprintf("%016x:%016x", want.snapshot.BID128High, want.snapshot.BID128Low)
	case intelCBenchmarkInt64:
		result = fmt.Sprintf("%d", want.snapshot.Int64)
	case intelCBenchmarkString:
		switch want.stringWidth {
		case 32:
			result = fmt.Sprintf("d32:%08x", want.snapshot.BID32)
		case 64:
			result = fmt.Sprintf("d64:%016x", want.snapshot.BID64)
		case 128:
			result = fmt.Sprintf("d128:%016x:%016x", want.snapshot.BID128High, want.snapshot.BID128Low)
		}
	}
	return fmt.Sprintf("kind=%d/result=%s/has_flags=%v/flags=%08x", want.result, result, want.hasFlags, want.snapshot.Flags)
}

func TestIntelCBenchmarkWiringOracleDiscriminatesRows(t *testing.T) {
	fixtures := intelCBenchmarkWiringFixtures(t)
	for _, group := range intelCBenchmarkGroups() {
		fingerprints := make(map[string]string, len(group.rows))
		for _, row := range group.rows {
			var combined strings.Builder
			for _, fixture := range fixtures {
				oracleInputs := newIntelCBenchmarkOracleInputs(t, fixture)
				combined.WriteString(fixture.name)
				combined.WriteByte('=')
				combined.WriteString(intelCBenchmarkOracleFingerprint(intelCBenchmarkOracle(t, group.name, row.name, oracleInputs)))
				combined.WriteByte('\n')
			}
			fingerprint := combined.String()
			if previous, duplicate := fingerprints[fingerprint]; duplicate {
				t.Errorf("%s rows %q and %q have indistinguishable oracle results across all wiring fixtures", group.name, previous, row.name)
			} else {
				fingerprints[fingerprint] = row.name
			}
		}
	}
}

func TestIntelCBenchmarkWiring(t *testing.T) {
	groups := intelCBenchmarkGroups()
	for _, group := range groups {
		if len(group.rows) != group.want {
			t.Errorf("%s row count = %d, want %d", group.name, len(group.rows), group.want)
		}
		seen := make(map[string]struct{}, len(group.rows))
		for _, row := range group.rows {
			if _, duplicate := seen[row.name]; duplicate {
				t.Errorf("%s duplicate row name %q", group.name, row.name)
			}
			seen[row.name] = struct{}{}
		}
	}

	for _, fixture := range intelCBenchmarkWiringFixtures(t) {
		oracleInputs := newIntelCBenchmarkOracleInputs(t, fixture)
		initIntelCBenchmarkOracleNativeInputs(t, oracleInputs.inputs)
		t.Run(fixture.name, func(t *testing.T) {
			for _, group := range groups {
				for _, row := range group.rows {
					t.Run(group.name+"/"+row.name, func(t *testing.T) {
						nativeBenchCBIDResetSinks()
						row.run(1)
						got := nativeBenchCBIDSnapshot()
						want := intelCBenchmarkOracle(t, group.name, row.name, oracleInputs)
						checkIntelCBenchmarkResult(t, row, got, want)
					})
				}
			}
		})
	}
}
