//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type alignedBenchmarkObservationMask uint16

const (
	alignedObserve32 alignedBenchmarkObservationMask = 1 << iota
	alignedObserve64
	alignedObserve128
	alignedObserveString
	alignedObserveInt64
	alignedObserveBool
	alignedObserveError
	alignedObserveFlags
	alignedObserveFlagsUnchanged
)

type alignedBenchmarkObservation struct {
	mask     alignedBenchmarkObservationMask
	value32  Decimal32BID
	value64  Decimal64BID
	value128 Decimal128BID
	text     string
	int64    int64
	boolean  bool
	errNil   bool
	errText  string
	flags    ExceptionFlags
}

type alignedBenchmarkWiringFixture struct {
	name   string
	inputs benchmarkInputs
}

type alignedBenchmarkTestGroup struct {
	name      string
	rows      []alignedBenchmarkRow
	wantCount int
	want      func(*testing.T, string) alignedBenchmarkObservation
}

func alignedBenchmarkWiringFixtures(t *testing.T) []alignedBenchmarkWiringFixture {
	t.Helper()
	production := loadBenchmarkInputs(t)
	return []alignedBenchmarkWiringFixture{
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

func alignedBenchmarkTestGroups(t *testing.T, inputs benchmarkInputs) []alignedBenchmarkTestGroup {
	t.Helper()
	x32 := exactBenchmarkDecimal32(t, inputs.Decimal32.X)
	y32 := exactBenchmarkDecimal32(t, inputs.Decimal32.Y)
	z32 := exactBenchmarkDecimal32(t, inputs.Decimal32.Z)
	i32, i32Flags := NewDecimal32FromInt64(inputs.IntegerOperand, RoundNearestEven)
	if i32Flags != 0 {
		t.Fatalf("Decimal32 integer operand flags = %v, want 0", i32Flags)
	}
	x64 := exactBenchmarkDecimal64(t, inputs.Decimal64.X)
	y64 := exactBenchmarkDecimal64(t, inputs.Decimal64.Y)
	z64 := exactBenchmarkDecimal64(t, inputs.Decimal64.Z)
	i64, i64Flags := NewDecimal64FromInt64(inputs.IntegerOperand, RoundNearestEven)
	if i64Flags != 0 {
		t.Fatalf("Decimal64 integer operand flags = %v, want 0", i64Flags)
	}
	x128 := exactBenchmarkDecimal128(t, inputs.Decimal128.X)
	y128 := exactBenchmarkDecimal128(t, inputs.Decimal128.Y)
	z128 := exactBenchmarkDecimal128(t, inputs.Decimal128.Z)
	i128 := NewDecimal128FromInt64(inputs.IntegerOperand)

	return []alignedBenchmarkTestGroup{
		{
			name:      "bid32",
			rows:      alignedBID32BenchmarkRowsForInputs(t, inputs),
			wantCount: 23,
			want: func(t *testing.T, name string) alignedBenchmarkObservation {
				return wantAlignedBID32(t, name, inputs, x32, y32, z32, i32)
			},
		},
		{
			name:      "bid64",
			rows:      alignedBID64BenchmarkRowsForInputs(t, inputs),
			wantCount: 23,
			want: func(t *testing.T, name string) alignedBenchmarkObservation {
				return wantAlignedBID64(t, name, inputs, x64, y64, z64, i64)
			},
		},
		{
			name:      "bid128",
			rows:      alignedBID128BenchmarkRowsForInputs(t, inputs),
			wantCount: 23,
			want: func(t *testing.T, name string) alignedBenchmarkObservation {
				return wantAlignedBID128(t, name, inputs, x128, y128, z128, i128)
			},
		},
		{
			name:      "bid64_mixed",
			rows:      alignedMixedBID64BenchmarkRowsForInputs(t, inputs),
			wantCount: 12,
			want: func(t *testing.T, name string) alignedBenchmarkObservation {
				return wantAlignedMixedBID64(t, name, x64, y64, x128, y128)
			},
		},
		{
			name:      "bid128_mixed",
			rows:      alignedMixedBID128BenchmarkRowsForInputs(t, inputs),
			wantCount: 12,
			want: func(t *testing.T, name string) alignedBenchmarkObservation {
				return wantAlignedMixedBID128(t, name, x64, y64, x128, y128)
			},
		},
	}
}

func TestAlignedBenchmarkRowsExecuteNamedOperations(t *testing.T) {
	for _, fixture := range alignedBenchmarkWiringFixtures(t) {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			for _, group := range alignedBenchmarkTestGroups(t, fixture.inputs) {
				group := group
				t.Run(group.name, func(t *testing.T) {
					if len(group.rows) != group.wantCount {
						t.Fatalf("row count = %d, want %d", len(group.rows), group.wantCount)
					}
					seen := make(map[string]struct{}, len(group.rows))
					for _, row := range group.rows {
						row := row
						t.Run(row.name, func(t *testing.T) {
							if _, duplicate := seen[row.name]; duplicate {
								t.Fatalf("duplicate benchmark row %q", row.name)
							}
							seen[row.name] = struct{}{}
							want := group.want(t, row.name)
							primeAlignedBenchmarkSinks(want)
							row.run(1)
							assertAlignedBenchmarkObservation(t, want)
						})
					}
				})
			}
		})
	}
}

func alignedBenchmarkObservationFingerprint(want alignedBenchmarkObservation) string {
	return fmt.Sprintf(
		"mask=%04x/value32=%08x/value64=%016x/value128=%x/text=%q/int64=%d/bool=%t/err_nil=%t/err_text=%q/flags=%x",
		want.mask,
		uint32(want.value32),
		uint64(want.value64),
		[16]byte(want.value128),
		want.text,
		want.int64,
		want.boolean,
		want.errNil,
		want.errText,
		want.flags,
	)
}

func TestAlignedBenchmarkWiringOracleDiscriminatesRows(t *testing.T) {
	fixtures := alignedBenchmarkWiringFixtures(t)
	fixtureGroups := make([][]alignedBenchmarkTestGroup, 0, len(fixtures))
	for _, fixture := range fixtures {
		fixtureGroups = append(fixtureGroups, alignedBenchmarkTestGroups(t, fixture.inputs))
	}

	for groupIndex, baseline := range fixtureGroups[0] {
		fingerprints := make(map[string]string, len(baseline.rows))
		for _, row := range baseline.rows {
			var combined strings.Builder
			for fixtureIndex, fixture := range fixtures {
				group := fixtureGroups[fixtureIndex][groupIndex]
				combined.WriteString(fixture.name)
				combined.WriteByte('=')
				combined.WriteString(alignedBenchmarkObservationFingerprint(group.want(t, row.name)))
				combined.WriteByte('\n')
			}
			fingerprint := combined.String()
			if previous, duplicate := fingerprints[fingerprint]; duplicate {
				t.Errorf("%s rows %q and %q have indistinguishable oracle results across all wiring fixtures", baseline.name, previous, row.name)
			} else {
				fingerprints[fingerprint] = row.name
			}
		}
	}
}

func wantAlignedBID32(t *testing.T, name string, inputs benchmarkInputs, x, y, z, integerValue Decimal32BID) alignedBenchmarkObservation {
	t.Helper()
	switch name {
	case "add":
		return aligned32(x.Add(y), 0, false)
	case "mul":
		return aligned32(x.Mul(y), 0, false)
	case "sub":
		return aligned32(x.Sub(y), 0, false)
	case "div":
		return aligned32(x.Div(y), 0, false)
	case "add_with_flags":
		value, flags := x.AddWithFlags(y)
		return aligned32(value, flags, true)
	case "mul_with_flags":
		value, flags := x.MulWithFlags(y)
		return aligned32(value, flags, true)
	case "sub_with_flags":
		value, flags := x.SubWithFlags(y)
		return aligned32(value, flags, true)
	case "div_with_flags":
		value, flags := x.DivWithFlags(y)
		return aligned32(value, flags, true)
	case "fma":
		value, flags := x.FMA(y, z)
		return aligned32(value, flags, true)
	case "sqrt":
		value, flags := x.Sqrt()
		return aligned32(value, flags, true)
	case "remainder":
		value, flags := y.Remainder(x)
		return aligned32(value, flags, true)
	case "fmod":
		value, flags := y.Fmod(x)
		return aligned32(value, flags, true)
	case "quantize":
		value, flags := x.QuantizeWithFlags(integerValue)
		return aligned32(value, flags, true)
	case "scaleb":
		value, flags := x.ScaleB(inputs.ScaleExponent)
		return aligned32(value, flags, true)
	case "quiet_equal":
		value, flags := x.QuietEqual(y)
		return alignedBool(value, flags)
	case "minnum":
		value, flags := x.MinNum(y)
		return aligned32(value, flags, true)
	case "maxnum":
		value, flags := x.MaxNum(y)
		return aligned32(value, flags, true)
	case "from_int64":
		value, flags := NewDecimal32FromInt64(inputs.IntegerOperand, RoundNearestEven)
		return aligned32(value, flags, true)
	case "to_int64":
		value, flags := integerValue.ConvertToInt64(RoundNearestEven)
		return alignedInt64(value, flags)
	case "to_decimal64":
		value, flags := x.ToDecimal64()
		return aligned64(value, flags, true)
	case "to_decimal128":
		value, flags := x.ToDecimal128()
		return aligned128(value, flags, true)
	case "parse":
		value, err := NewDecimal32BIDDirect(inputs.Decimal32.X)
		return aligned32Error(value, err)
	case "to_string":
		return alignedText(x.String())
	default:
		t.Fatalf("unknown Decimal32 benchmark row %q", name)
		return alignedBenchmarkObservation{}
	}
}

func wantAlignedBID64(t *testing.T, name string, inputs benchmarkInputs, x, y, z, integerValue Decimal64BID) alignedBenchmarkObservation {
	t.Helper()
	switch name {
	case "add":
		return aligned64(x.Add(y), 0, false)
	case "mul":
		return aligned64(x.Mul(y), 0, false)
	case "sub":
		return aligned64(x.Sub(y), 0, false)
	case "div":
		return aligned64(x.Div(y), 0, false)
	case "add_with_flags":
		value, flags := x.AddWithFlags(y)
		return aligned64(value, flags, true)
	case "mul_with_flags":
		value, flags := x.MulWithFlags(y)
		return aligned64(value, flags, true)
	case "sub_with_flags":
		value, flags := x.SubWithFlags(y)
		return aligned64(value, flags, true)
	case "div_with_flags":
		value, flags := x.DivWithFlags(y)
		return aligned64(value, flags, true)
	case "fma":
		value, flags := x.FMA(y, z)
		return aligned64(value, flags, true)
	case "sqrt":
		value, flags := x.Sqrt()
		return aligned64(value, flags, true)
	case "remainder":
		value, flags := y.Remainder(x)
		return aligned64(value, flags, true)
	case "fmod":
		value, flags := y.Fmod(x)
		return aligned64(value, flags, true)
	case "quantize":
		value, flags := x.QuantizeWithFlags(integerValue)
		return aligned64(value, flags, true)
	case "scaleb":
		value, flags := x.ScaleB(inputs.ScaleExponent)
		return aligned64(value, flags, true)
	case "quiet_equal":
		value, flags := x.QuietEqual(y)
		return alignedBool(value, flags)
	case "minnum":
		value, flags := x.MinNum(y)
		return aligned64(value, flags, true)
	case "maxnum":
		value, flags := x.MaxNum(y)
		return aligned64(value, flags, true)
	case "from_int64":
		value, flags := NewDecimal64FromInt64(inputs.IntegerOperand, RoundNearestEven)
		return aligned64(value, flags, true)
	case "to_int64":
		value, flags := integerValue.ConvertToInt64(RoundNearestEven)
		return alignedInt64(value, flags)
	case "to_decimal32":
		value, flags := x.ToDecimal32(RoundNearestEven)
		return aligned32(value, flags, true)
	case "to_decimal128":
		value, flags := x.ToDecimal128()
		return aligned128(value, flags, true)
	case "parse":
		value, err := NewDecimal64BIDDirect(inputs.Decimal64.X)
		return aligned64Error(value, err)
	case "to_string":
		return alignedText(x.String())
	default:
		t.Fatalf("unknown Decimal64 benchmark row %q", name)
		return alignedBenchmarkObservation{}
	}
}

func wantAlignedBID128(t *testing.T, name string, inputs benchmarkInputs, x, y, z, integerValue Decimal128BID) alignedBenchmarkObservation {
	t.Helper()
	switch name {
	case "add":
		return aligned128(x.Add(y), 0, false)
	case "mul":
		return aligned128(x.Mul(y), 0, false)
	case "sub":
		return aligned128(x.Sub(y), 0, false)
	case "div":
		return aligned128(x.Div(y), 0, false)
	case "add_with_flags":
		value, flags := x.AddWithFlags(y)
		return aligned128(value, flags, true)
	case "mul_with_flags":
		value, flags := x.MulWithFlags(y)
		return aligned128(value, flags, true)
	case "sub_with_flags":
		value, flags := x.SubWithFlags(y)
		return aligned128(value, flags, true)
	case "div_with_flags":
		value, flags := x.DivWithFlags(y)
		return aligned128(value, flags, true)
	case "fma":
		value, flags := x.FMA(y, z)
		return aligned128(value, flags, true)
	case "sqrt":
		value, flags := x.Sqrt()
		return aligned128(value, flags, true)
	case "remainder":
		value, flags := y.Remainder(x)
		return aligned128(value, flags, true)
	case "fmod":
		value, flags := y.Fmod(x)
		return aligned128(value, flags, true)
	case "quantize":
		value, flags := x.QuantizeWithFlags(integerValue)
		return aligned128(value, flags, true)
	case "scaleb":
		value, flags := x.ScaleB(inputs.ScaleExponent)
		return aligned128(value, flags, true)
	case "quiet_equal":
		value, flags := x.QuietEqual(y)
		return alignedBool(value, flags)
	case "minnum":
		value, flags := x.MinNum(y)
		return aligned128(value, flags, true)
	case "maxnum":
		value, flags := x.MaxNum(y)
		return aligned128(value, flags, true)
	case "from_int64":
		return aligned128(NewDecimal128FromInt64(inputs.IntegerOperand), 0, false)
	case "to_int64":
		value, flags := integerValue.ConvertToInt64(RoundNearestEven)
		return alignedInt64(value, flags)
	case "to_decimal32":
		value, flags := x.ToDecimal32(RoundNearestEven)
		return aligned32(value, flags, true)
	case "to_decimal64":
		value, flags := x.ToDecimal64(RoundNearestEven)
		return aligned64(value, flags, true)
	case "parse":
		value, err := NewDecimal128BIDDirect(inputs.Decimal128.X)
		return aligned128Error(value, err)
	case "to_string":
		return alignedText(x.String())
	default:
		t.Fatalf("unknown Decimal128 benchmark row %q", name)
		return alignedBenchmarkObservation{}
	}
}

func wantAlignedMixedBID64(t *testing.T, name string, dx, dy Decimal64BID, qx, qy Decimal128BID) alignedBenchmarkObservation {
	t.Helper()
	var value Decimal64BID
	var flags ExceptionFlags
	switch name {
	case "dq_add":
		value, flags = Add64DQBIDWithMode(dx, qy, RoundNearestEven)
	case "qd_add":
		value, flags = Add64QDBIDWithMode(qx, dy, RoundNearestEven)
	case "qq_add":
		value, flags = Add64QQBIDWithMode(qx, qy, RoundNearestEven)
	case "dq_sub":
		value, flags = Sub64DQBIDWithMode(dx, qy, RoundNearestEven)
	case "qd_sub":
		value, flags = Sub64QDBIDWithMode(qx, dy, RoundNearestEven)
	case "qq_sub":
		value, flags = Sub64QQBIDWithMode(qx, qy, RoundNearestEven)
	case "dq_mul":
		value, flags = Mul64DQBIDWithMode(dx, qy, RoundNearestEven)
	case "qd_mul":
		value, flags = Mul64QDBIDWithMode(qx, dy, RoundNearestEven)
	case "qq_mul":
		value, flags = Mul64QQBIDWithMode(qx, qy, RoundNearestEven)
	case "dq_div":
		value, flags = Div64DQBIDWithMode(dx, qy, RoundNearestEven)
	case "qd_div":
		value, flags = Div64QDBIDWithMode(qx, dy, RoundNearestEven)
	case "qq_div":
		value, flags = Div64QQBIDWithMode(qx, qy, RoundNearestEven)
	default:
		t.Fatalf("unknown mixed Decimal64 benchmark row %q", name)
	}
	return aligned64(value, flags, true)
}

func wantAlignedMixedBID128(t *testing.T, name string, dx, dy Decimal64BID, qx, qy Decimal128BID) alignedBenchmarkObservation {
	t.Helper()
	var value Decimal128BID
	var flags ExceptionFlags
	switch name {
	case "dd_add":
		value, flags = Add128DDBIDWithMode(dx, dy, RoundNearestEven)
	case "dq_add":
		value, flags = Add128DQBIDWithMode(dx, qy, RoundNearestEven)
	case "qd_add":
		value, flags = Add128QDBIDWithMode(qx, dy, RoundNearestEven)
	case "dd_sub":
		value, flags = Sub128DDBIDWithMode(dx, dy, RoundNearestEven)
	case "dq_sub":
		value, flags = Sub128DQBIDWithMode(dx, qy, RoundNearestEven)
	case "qd_sub":
		value, flags = Sub128QDBIDWithMode(qx, dy, RoundNearestEven)
	case "dd_mul":
		value, flags = Mul128DDBIDWithMode(dx, dy, RoundNearestEven)
	case "dq_mul":
		value, flags = Mul128DQBIDWithMode(dx, qy, RoundNearestEven)
	case "qd_mul":
		value, flags = Mul128QDBIDWithMode(qx, dy, RoundNearestEven)
	case "dd_div":
		value, flags = Div128DDBIDWithMode(dx, dy, RoundNearestEven)
	case "dq_div":
		value, flags = Div128DQBIDWithMode(dx, qy, RoundNearestEven)
	case "qd_div":
		value, flags = Div128QDBIDWithMode(qx, dy, RoundNearestEven)
	default:
		t.Fatalf("unknown mixed Decimal128 benchmark row %q", name)
	}
	return aligned128(value, flags, true)
}

func aligned32(value Decimal32BID, flags ExceptionFlags, withFlags bool) alignedBenchmarkObservation {
	mask := alignedObserve32
	if withFlags {
		mask |= alignedObserveFlags
	} else {
		mask |= alignedObserveFlagsUnchanged
		flags = ExceptionFlags(0x55)
	}
	return alignedBenchmarkObservation{mask: mask, value32: value, flags: flags}
}

func aligned64(value Decimal64BID, flags ExceptionFlags, withFlags bool) alignedBenchmarkObservation {
	mask := alignedObserve64
	if withFlags {
		mask |= alignedObserveFlags
	} else {
		mask |= alignedObserveFlagsUnchanged
		flags = ExceptionFlags(0x55)
	}
	return alignedBenchmarkObservation{mask: mask, value64: value, flags: flags}
}

func aligned128(value Decimal128BID, flags ExceptionFlags, withFlags bool) alignedBenchmarkObservation {
	mask := alignedObserve128
	if withFlags {
		mask |= alignedObserveFlags
	} else {
		mask |= alignedObserveFlagsUnchanged
		flags = ExceptionFlags(0x55)
	}
	return alignedBenchmarkObservation{mask: mask, value128: value, flags: flags}
}

func alignedBool(value bool, flags ExceptionFlags) alignedBenchmarkObservation {
	return alignedBenchmarkObservation{mask: alignedObserveBool | alignedObserveFlags, boolean: value, flags: flags}
}

func alignedInt64(value int64, flags ExceptionFlags) alignedBenchmarkObservation {
	return alignedBenchmarkObservation{mask: alignedObserveInt64 | alignedObserveFlags, int64: value, flags: flags}
}

func alignedText(value string) alignedBenchmarkObservation {
	return alignedBenchmarkObservation{
		mask:  alignedObserveString | alignedObserveFlagsUnchanged,
		text:  value,
		flags: ExceptionFlags(0x55),
	}
}

func aligned32Error(value Decimal32BID, err error) alignedBenchmarkObservation {
	want := aligned32(value, 0, false)
	return alignedWithError(want, err)
}

func aligned64Error(value Decimal64BID, err error) alignedBenchmarkObservation {
	want := aligned64(value, 0, false)
	return alignedWithError(want, err)
}

func aligned128Error(value Decimal128BID, err error) alignedBenchmarkObservation {
	want := aligned128(value, 0, false)
	return alignedWithError(want, err)
}

func alignedWithError(want alignedBenchmarkObservation, err error) alignedBenchmarkObservation {
	want.mask |= alignedObserveError
	want.errNil = err == nil
	if err != nil {
		want.errText = err.Error()
	}
	return want
}

func primeAlignedBenchmarkSinks(want alignedBenchmarkObservation) {
	if want.mask&alignedObserve32 != 0 {
		alignedSink32 = ^want.value32
	}
	if want.mask&alignedObserve64 != 0 {
		alignedSink64 = ^want.value64
	}
	if want.mask&alignedObserve128 != 0 {
		for i := range alignedSink128 {
			alignedSink128[i] = ^want.value128[i]
		}
	}
	if want.mask&alignedObserveString != 0 {
		alignedSinkString = want.text + "#unwritten"
	}
	if want.mask&alignedObserveInt64 != 0 {
		alignedSinkInt64 = ^want.int64
	}
	if want.mask&alignedObserveBool != 0 {
		alignedSinkBool = !want.boolean
	}
	if want.mask&alignedObserveError != 0 {
		if want.errNil {
			alignedSinkErr = errors.New("benchmark row did not write error sink")
		} else {
			alignedSinkErr = nil
		}
	}
	if want.mask&alignedObserveFlags != 0 {
		alignedSinkFlags = want.flags ^ ExceptionFlags(0x55)
	}
	if want.mask&alignedObserveFlagsUnchanged != 0 {
		alignedSinkFlags = want.flags
	}
}

func assertAlignedBenchmarkObservation(t *testing.T, want alignedBenchmarkObservation) {
	t.Helper()
	if want.mask&alignedObserve32 != 0 && alignedSink32 != want.value32 {
		t.Errorf("Decimal32 bits = %#x, want %#x", uint32(alignedSink32), uint32(want.value32))
	}
	if want.mask&alignedObserve64 != 0 && alignedSink64 != want.value64 {
		t.Errorf("Decimal64 bits = %#x, want %#x", uint64(alignedSink64), uint64(want.value64))
	}
	if want.mask&alignedObserve128 != 0 && alignedSink128 != want.value128 {
		t.Errorf("Decimal128 bits = %x, want %x", [16]byte(alignedSink128), [16]byte(want.value128))
	}
	if want.mask&alignedObserveString != 0 && alignedSinkString != want.text {
		t.Errorf("string = %q, want %q", alignedSinkString, want.text)
	}
	if want.mask&alignedObserveInt64 != 0 && alignedSinkInt64 != want.int64 {
		t.Errorf("int64 = %d, want %d", alignedSinkInt64, want.int64)
	}
	if want.mask&alignedObserveBool != 0 && alignedSinkBool != want.boolean {
		t.Errorf("bool = %t, want %t", alignedSinkBool, want.boolean)
	}
	if want.mask&alignedObserveError != 0 {
		gotNil := alignedSinkErr == nil
		gotText := ""
		if alignedSinkErr != nil {
			gotText = alignedSinkErr.Error()
		}
		if gotNil != want.errNil || gotText != want.errText {
			t.Errorf("error = (%t, %q), want (%t, %q)", gotNil, gotText, want.errNil, want.errText)
		}
	}
	if want.mask&alignedObserveFlags != 0 && alignedSinkFlags != want.flags {
		t.Errorf("flags = %v, want %v", alignedSinkFlags, want.flags)
	}
	if want.mask&alignedObserveFlagsUnchanged != 0 && alignedSinkFlags != want.flags {
		t.Errorf("flags changed to %v, want untouched sentinel %v", alignedSinkFlags, want.flags)
	}
	if t.Failed() {
		t.Logf("observation mask = %s", formatAlignedObservationMask(want.mask))
	}
}

func formatAlignedObservationMask(mask alignedBenchmarkObservationMask) string {
	return fmt.Sprintf("%08b", mask)
}
