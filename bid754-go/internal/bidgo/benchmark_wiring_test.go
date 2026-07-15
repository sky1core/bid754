package bidgo

import (
	"fmt"
	"strings"
	"testing"
)

type fairBenchmarkObservationMask uint16

const (
	fairObserve32 fairBenchmarkObservationMask = 1 << iota
	fairObserve64
	fairObserve128
	fairObserveString
	fairObserveInt64
	fairObserveInt
	fairObserveFlags
	fairObserveFlagsUnchanged
)

type fairBenchmarkObservation struct {
	mask     fairBenchmarkObservationMask
	value32  uint32
	value64  uint64
	value128 BID_UINT128
	text     string
	int64    int64
	integer  int
	flags    uint32
}

type fairBenchmarkWiringFixture struct {
	name   string
	inputs benchmarkInputs
}

type fairBenchmarkTestGroup struct {
	name      string
	rows      []fairBenchmarkRow
	wantCount int
	want      func(*testing.T, string) fairBenchmarkObservation
}

func fairBenchmarkWiringFixtures(t *testing.T) []fairBenchmarkWiringFixture {
	t.Helper()
	production := loadBenchmarkInputs(t)
	return []fairBenchmarkWiringFixture{
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

func fairBenchmarkTestGroups(t *testing.T, inputs benchmarkInputs) []fairBenchmarkTestGroup {
	t.Helper()
	x32 := exactBenchmarkDecimal32(t, inputs.Decimal32.X)
	y32 := exactBenchmarkDecimal32(t, inputs.Decimal32.Y)
	z32 := exactBenchmarkDecimal32(t, inputs.Decimal32.Z)
	i32, flags32 := Bid32FromInt64(inputs.IntegerOperand, 0)
	if flags32 != 0 {
		t.Fatalf("Decimal32 integer operand flags = %#x, want 0", flags32)
	}
	x64 := exactBenchmarkDecimal64(t, inputs.Decimal64.X)
	y64 := exactBenchmarkDecimal64(t, inputs.Decimal64.Y)
	z64 := exactBenchmarkDecimal64(t, inputs.Decimal64.Z)
	i64, flags64 := Bid64FromInt64(inputs.IntegerOperand, 0)
	if flags64 != 0 {
		t.Fatalf("Decimal64 integer operand flags = %#x, want 0", flags64)
	}
	x128 := exactBenchmarkDecimal128(t, inputs.Decimal128.X)
	y128 := exactBenchmarkDecimal128(t, inputs.Decimal128.Y)
	z128 := exactBenchmarkDecimal128(t, inputs.Decimal128.Z)
	i128 := Bid128FromInt64(inputs.IntegerOperand)

	return []fairBenchmarkTestGroup{
		{
			name:      "bid32",
			rows:      fairBID32BenchmarkRowsForInputs(t, inputs),
			wantCount: 23,
			want: func(t *testing.T, name string) fairBenchmarkObservation {
				return wantFairBID32(t, name, inputs, x32, y32, z32, i32)
			},
		},
		{
			name:      "bid64",
			rows:      fairBID64BenchmarkRowsForInputs(t, inputs),
			wantCount: 19,
			want: func(t *testing.T, name string) fairBenchmarkObservation {
				return wantFairBID64(t, name, inputs, x64, y64, z64, i64)
			},
		},
		{
			name:      "bid128",
			rows:      fairBID128BenchmarkRowsForInputs(t, inputs),
			wantCount: 19,
			want: func(t *testing.T, name string) fairBenchmarkObservation {
				return wantFairBID128(t, name, inputs, x128, y128, z128, i128)
			},
		},
		{
			name:      "bid64_mixed",
			rows:      fairMixedBID64BenchmarkRowsForInputs(t, inputs),
			wantCount: 12,
			want: func(t *testing.T, name string) fairBenchmarkObservation {
				return wantFairMixedBID64(t, name, x64, y64, x128, y128)
			},
		},
		{
			name:      "bid128_mixed",
			rows:      fairMixedBID128BenchmarkRowsForInputs(t, inputs),
			wantCount: 12,
			want: func(t *testing.T, name string) fairBenchmarkObservation {
				return wantFairMixedBID128(t, name, x64, y64, x128, y128)
			},
		},
	}
}

func TestFairBenchmarkRowsExecuteNamedOperations(t *testing.T) {
	for _, fixture := range fairBenchmarkWiringFixtures(t) {
		fixture := fixture
		t.Run(fixture.name, func(t *testing.T) {
			for _, group := range fairBenchmarkTestGroups(t, fixture.inputs) {
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
							primeFairBenchmarkSinks(want)
							row.run(1)
							assertFairBenchmarkObservation(t, want)
						})
					}
				})
			}
		})
	}
}

func fairBenchmarkObservationFingerprint(want fairBenchmarkObservation) string {
	return fmt.Sprintf(
		"mask=%04x/value32=%08x/value64=%016x/value128=%016x:%016x/text=%q/int64=%d/int=%d/flags=%08x",
		want.mask,
		want.value32,
		want.value64,
		want.value128.hi,
		want.value128.lo,
		want.text,
		want.int64,
		want.integer,
		want.flags,
	)
}

func TestFairBenchmarkWiringOracleDiscriminatesRows(t *testing.T) {
	fixtures := fairBenchmarkWiringFixtures(t)
	fixtureGroups := make([][]fairBenchmarkTestGroup, 0, len(fixtures))
	for _, fixture := range fixtures {
		fixtureGroups = append(fixtureGroups, fairBenchmarkTestGroups(t, fixture.inputs))
	}

	for groupIndex, baseline := range fixtureGroups[0] {
		fingerprints := make(map[string]string, len(baseline.rows))
		for _, row := range baseline.rows {
			var combined strings.Builder
			for fixtureIndex, fixture := range fixtures {
				group := fixtureGroups[fixtureIndex][groupIndex]
				combined.WriteString(fixture.name)
				combined.WriteByte('=')
				combined.WriteString(fairBenchmarkObservationFingerprint(group.want(t, row.name)))
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

func wantFairBID32(t *testing.T, name string, inputs benchmarkInputs, x, y, z, integerValue uint32) fairBenchmarkObservation {
	t.Helper()
	var value uint32
	var flags uint32
	switch name {
	case "add":
		value, flags = Bid32AddWithFlags(x, y, 0)
	case "mul":
		value, flags = Bid32MulWithFlags(x, y, 0)
	case "sub":
		value, flags = Bid32SubWithFlags(x, y, 0)
	case "div":
		value, flags = Bid32DivWithFlags(x, y, 0)
	case "add_pure":
		return fair32(Bid32Add(x, y, 0), 0, false)
	case "mul_pure":
		return fair32(Bid32Mul(x, y, 0), 0, false)
	case "sub_pure":
		return fair32(Bid32Sub(x, y, 0), 0, false)
	case "div_pure":
		return fair32(Bid32Div(x, y, 0), 0, false)
	case "fma":
		value, flags = Bid32Fma(x, y, z, 0)
	case "sqrt":
		value, flags = Bid32Sqrt(x, 0)
	case "remainder":
		value, flags = Bid32Rem(y, x)
	case "fmod":
		value, flags = Bid32Fmod(y, x)
	case "quantize":
		value, flags = Bid32Quantize(x, integerValue, 0)
	case "scaleb":
		value, flags = Bid32ScalbnWithFlags(x, inputs.ScaleExponent, 0)
	case "quiet_equal":
		integer, compareFlags := Bid32QuietEqual(x, y)
		return fairInt(integer, compareFlags)
	case "minnum":
		value, flags = Bid32MinNumWithFlags(x, y)
	case "maxnum":
		value, flags = Bid32MaxNumWithFlags(x, y)
	case "from_int64":
		value, flags = Bid32FromInt64(inputs.IntegerOperand, 0)
	case "to_int64":
		integer, convertFlags := Bid32ToInt64Rnint(integerValue)
		return fairInt64(integer, convertFlags)
	case "to_decimal64":
		converted, convertFlags := Bid32ToBid64(x)
		return fair64(converted, convertFlags, true)
	case "to_decimal128":
		converted, convertFlags := Bid32ToBid128(x)
		return fair128(converted, convertFlags, true)
	case "parse":
		value, flags = Bid32FromStringRaw(inputs.Decimal32.X, 0)
	case "to_string":
		return fairText(Bid32ToStringRaw(x))
	default:
		t.Fatalf("unknown Decimal32 benchmark row %q", name)
	}
	return fair32(value, flags, true)
}

func wantFairBID64(t *testing.T, name string, inputs benchmarkInputs, x, y, z, integerValue uint64) fairBenchmarkObservation {
	t.Helper()
	var value uint64
	var flags uint32
	switch name {
	case "add":
		value, flags = Bid64AddWithFlags(x, y, 0)
	case "mul":
		value, flags = Bid64MulWithFlags(x, y, 0)
	case "sub":
		value, flags = Bid64SubWithFlags(x, y, 0)
	case "div":
		value, flags = Bid64DivWithFlags(x, y, 0)
	case "fma":
		value, flags = Bid64Fma(x, y, z, 0)
	case "sqrt":
		value, flags = Bid64Sqrt(x, 0)
	case "remainder":
		value, flags = Bid64Rem(y, x)
	case "fmod":
		value, flags = Bid64Fmod(y, x)
	case "quantize":
		value, flags = Bid64Quantize(x, integerValue, 0)
	case "scaleb":
		value, flags = Bid64Scalbn(x, inputs.ScaleExponent, 0)
	case "quiet_equal":
		integer, compareFlags := Bid64QuietEqual(x, y)
		return fairInt(integer, compareFlags)
	case "minnum":
		value, flags = Bid64MinNum(x, y)
	case "maxnum":
		value, flags = Bid64MaxNum(x, y)
	case "from_int64":
		value, flags = Bid64FromInt64(inputs.IntegerOperand, 0)
	case "to_int64":
		integer, convertFlags := Bid64ToInt64Rnint(integerValue)
		return fairInt64(integer, convertFlags)
	case "to_decimal32":
		converted, convertFlags := Bid64ToBid32(x, 0)
		return fair32(converted, convertFlags, true)
	case "to_decimal128":
		converted, convertFlags := Bid64ToBid128(x)
		return fair128(converted, convertFlags, true)
	case "parse":
		value, flags = Bid64FromString(inputs.Decimal64.X, 0)
	case "to_string":
		return fairText(Bid64ToString(x))
	default:
		t.Fatalf("unknown Decimal64 benchmark row %q", name)
	}
	return fair64(value, flags, true)
}

func wantFairBID128(t *testing.T, name string, inputs benchmarkInputs, x, y, z, integerValue BID_UINT128) fairBenchmarkObservation {
	t.Helper()
	var value BID_UINT128
	var flags uint32
	switch name {
	case "add":
		value = Bid128Add(x, y, 0, &flags)
	case "mul":
		value, flags = Bid128Mul(x, y, 0)
	case "sub":
		value = Bid128Sub(x, y, 0, &flags)
	case "div":
		value, flags = Bid128Div(x, y, 0)
	case "fma":
		value, flags = Bid128Fma(x, y, z, 0)
	case "sqrt":
		value, flags = Bid128Sqrt(x, 0)
	case "remainder":
		value, flags = Bid128Rem(y, x)
	case "fmod":
		value, flags = Bid128Fmod(y, x)
	case "quantize":
		value, flags = Bid128Quantize(x, integerValue, 0)
	case "scaleb":
		value = Bid128Scalbn(x, inputs.ScaleExponent, 0, &flags)
	case "quiet_equal":
		integer, compareFlags := Bid128QuietEqual(x, y)
		return fairInt(integer, compareFlags)
	case "minnum":
		value = Bid128Minnum(x, y, &flags)
	case "maxnum":
		value = Bid128Maxnum(x, y, &flags)
	case "from_int64":
		return fair128(Bid128FromInt64(inputs.IntegerOperand), 0, false)
	case "to_int64":
		integer, convertFlags := Bid128ToInt64Rnint(integerValue)
		return fairInt64(integer, convertFlags)
	case "to_decimal32":
		converted, convertFlags := Bid128ToBid32(x, 0)
		return fair32(converted, convertFlags, true)
	case "to_decimal64":
		converted, convertFlags := Bid128ToBid64(x, 0)
		return fair64(converted, convertFlags, true)
	case "parse":
		value, flags = Bid128FromString(inputs.Decimal128.X, 0)
	case "to_string":
		return fairText(Bid128ToString(x))
	default:
		t.Fatalf("unknown Decimal128 benchmark row %q", name)
	}
	return fair128(value, flags, true)
}

func wantFairMixedBID64(t *testing.T, name string, dx, dy uint64, qx, qy BID_UINT128) fairBenchmarkObservation {
	t.Helper()
	var value uint64
	var flags uint32
	switch name {
	case "dq_add":
		value, flags = Bid64dqAdd(dx, qy, 0)
	case "qd_add":
		value, flags = Bid64qdAdd(qx, dy, 0)
	case "qq_add":
		value, flags = Bid64qqAdd(qx, qy, 0)
	case "dq_sub":
		value, flags = Bid64dqSub(dx, qy, 0)
	case "qd_sub":
		value, flags = Bid64qdSub(qx, dy, 0)
	case "qq_sub":
		value, flags = Bid64qqSub(qx, qy, 0)
	case "dq_mul":
		value, flags = Bid64dqMul(dx, qy, 0)
	case "qd_mul":
		value, flags = Bid64qdMul(qx, dy, 0)
	case "qq_mul":
		value, flags = Bid64qqMul(qx, qy, 0)
	case "dq_div":
		value, flags = Bid64dqDiv(dx, qy, 0)
	case "qd_div":
		value, flags = Bid64qdDiv(qx, dy, 0)
	case "qq_div":
		value, flags = Bid64qqDiv(qx, qy, 0)
	default:
		t.Fatalf("unknown mixed Decimal64 benchmark row %q", name)
	}
	return fair64(value, flags, true)
}

func wantFairMixedBID128(t *testing.T, name string, dx, dy uint64, qx, qy BID_UINT128) fairBenchmarkObservation {
	t.Helper()
	var value BID_UINT128
	var flags uint32
	switch name {
	case "dd_add":
		value, flags = Bid128ddAdd(dx, dy, 0)
	case "dq_add":
		value, flags = Bid128dqAdd(dx, qy, 0)
	case "qd_add":
		value, flags = Bid128qdAdd(qx, dy, 0)
	case "dd_sub":
		value, flags = Bid128ddSub(dx, dy, 0)
	case "dq_sub":
		value, flags = Bid128dqSub(dx, qy, 0)
	case "qd_sub":
		value, flags = Bid128qdSub(qx, dy, 0)
	case "dd_mul":
		value, flags = Bid128ddMul(dx, dy, 0)
	case "dq_mul":
		value, flags = Bid128dqMul(dx, qy, 0)
	case "qd_mul":
		value, flags = Bid128qdMul(qx, dy, 0)
	case "dd_div":
		value, flags = Bid128ddDiv(dx, dy, 0)
	case "dq_div":
		value, flags = Bid128dqDiv(dx, qy, 0)
	case "qd_div":
		value, flags = Bid128qdDiv(qx, dy, 0)
	default:
		t.Fatalf("unknown mixed Decimal128 benchmark row %q", name)
	}
	return fair128(value, flags, true)
}

func fair32(value uint32, flags uint32, withFlags bool) fairBenchmarkObservation {
	mask := fairObserve32
	if withFlags {
		mask |= fairObserveFlags
	} else {
		mask |= fairObserveFlagsUnchanged
		flags = 0x55
	}
	return fairBenchmarkObservation{mask: mask, value32: value, flags: flags}
}

func fair64(value uint64, flags uint32, withFlags bool) fairBenchmarkObservation {
	mask := fairObserve64
	if withFlags {
		mask |= fairObserveFlags
	} else {
		mask |= fairObserveFlagsUnchanged
		flags = 0x55
	}
	return fairBenchmarkObservation{mask: mask, value64: value, flags: flags}
}

func fair128(value BID_UINT128, flags uint32, withFlags bool) fairBenchmarkObservation {
	mask := fairObserve128
	if withFlags {
		mask |= fairObserveFlags
	} else {
		mask |= fairObserveFlagsUnchanged
		flags = 0x55
	}
	return fairBenchmarkObservation{mask: mask, value128: value, flags: flags}
}

func fairText(value string) fairBenchmarkObservation {
	return fairBenchmarkObservation{mask: fairObserveString | fairObserveFlagsUnchanged, text: value, flags: 0x55}
}

func fairInt64(value int64, flags uint32) fairBenchmarkObservation {
	return fairBenchmarkObservation{mask: fairObserveInt64 | fairObserveFlags, int64: value, flags: flags}
}

func fairInt(value int, flags uint32) fairBenchmarkObservation {
	return fairBenchmarkObservation{mask: fairObserveInt | fairObserveFlags, integer: value, flags: flags}
}

func primeFairBenchmarkSinks(want fairBenchmarkObservation) {
	if want.mask&fairObserve32 != 0 {
		sink32 = ^want.value32
	}
	if want.mask&fairObserve64 != 0 {
		sink64 = ^want.value64
	}
	if want.mask&fairObserve128 != 0 {
		sink128 = BID_UINT128{hi: ^want.value128.hi, lo: ^want.value128.lo}
	}
	if want.mask&fairObserveString != 0 {
		sinkString = want.text + "#unwritten"
	}
	if want.mask&fairObserveInt64 != 0 {
		sinkInt64 = ^want.int64
	}
	if want.mask&fairObserveInt != 0 {
		sinkInt = ^want.integer
	}
	if want.mask&fairObserveFlags != 0 {
		sinkFlags = want.flags ^ 0x55
	}
	if want.mask&fairObserveFlagsUnchanged != 0 {
		sinkFlags = want.flags
	}
}

func assertFairBenchmarkObservation(t *testing.T, want fairBenchmarkObservation) {
	t.Helper()
	if want.mask&fairObserve32 != 0 && sink32 != want.value32 {
		t.Errorf("Decimal32 bits = %#x, want %#x", sink32, want.value32)
	}
	if want.mask&fairObserve64 != 0 && sink64 != want.value64 {
		t.Errorf("Decimal64 bits = %#x, want %#x", sink64, want.value64)
	}
	if want.mask&fairObserve128 != 0 && sink128 != want.value128 {
		t.Errorf("Decimal128 bits = %#x/%#x, want %#x/%#x", sink128.hi, sink128.lo, want.value128.hi, want.value128.lo)
	}
	if want.mask&fairObserveString != 0 && sinkString != want.text {
		t.Errorf("string = %q, want %q", sinkString, want.text)
	}
	if want.mask&fairObserveInt64 != 0 && sinkInt64 != want.int64 {
		t.Errorf("int64 = %d, want %d", sinkInt64, want.int64)
	}
	if want.mask&fairObserveInt != 0 && sinkInt != want.integer {
		t.Errorf("int = %d, want %d", sinkInt, want.integer)
	}
	if want.mask&fairObserveFlags != 0 && sinkFlags != want.flags {
		t.Errorf("flags = %#x, want %#x", sinkFlags, want.flags)
	}
	if want.mask&fairObserveFlagsUnchanged != 0 && sinkFlags != want.flags {
		t.Errorf("flags changed to %#x, want untouched sentinel %#x", sinkFlags, want.flags)
	}
}
