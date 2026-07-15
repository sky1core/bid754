//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import "testing"

func runIntelCBIDBench(b *testing.B, fn func(int)) {
	requireNativeBenchmark(b)
	inputs := loadBenchmarkInputs(b)
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
		b.Fatal("Intel C benchmark inputs are not finite and exact")
	}
	b.ReportAllocs()
	b.ResetTimer()
	fn(b.N)
}

func TestIntelCBenchmarkInputsAreFiniteAndExact(t *testing.T) {
	inputs := loadBenchmarkInputs(t)
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
		t.Fatal("Intel C benchmark inputs are not finite and exact")
	}
}

type intelCBenchmarkResultKind uint8

const (
	intelCBenchmarkBID32 intelCBenchmarkResultKind = iota
	intelCBenchmarkBID64
	intelCBenchmarkBID128
	intelCBenchmarkBool32
	intelCBenchmarkBool64
	intelCBenchmarkInt64
	intelCBenchmarkString
)

type intelCBenchmarkRow struct {
	name     string
	run      func(int)
	result   intelCBenchmarkResultKind
	hasFlags bool
}

var intelCBID32BenchmarkRows = []intelCBenchmarkRow{
	{"add", nativeBenchCBID32Add, intelCBenchmarkBID32, true},
	{"mul", nativeBenchCBID32Mul, intelCBenchmarkBID32, true},
	{"sub", nativeBenchCBID32Sub, intelCBenchmarkBID32, true},
	{"div", nativeBenchCBID32Div, intelCBenchmarkBID32, true},
	{"fma", nativeBenchCBID32Fma, intelCBenchmarkBID32, true},
	{"sqrt", nativeBenchCBID32Sqrt, intelCBenchmarkBID32, true},
	{"remainder", nativeBenchCBID32Remainder, intelCBenchmarkBID32, true},
	{"fmod", nativeBenchCBID32Fmod, intelCBenchmarkBID32, true},
	{"quantize", nativeBenchCBID32Quantize, intelCBenchmarkBID32, true},
	{"scaleb", nativeBenchCBID32ScaleB, intelCBenchmarkBID32, true},
	{"quiet_equal", nativeBenchCBID32QuietEqual, intelCBenchmarkBool32, true},
	{"minnum", nativeBenchCBID32MinNum, intelCBenchmarkBID32, true},
	{"maxnum", nativeBenchCBID32MaxNum, intelCBenchmarkBID32, true},
	{"from_int64", nativeBenchCBID32FromInt64, intelCBenchmarkBID32, true},
	{"to_int64", nativeBenchCBID32ToInt64, intelCBenchmarkInt64, true},
	{"to_decimal64", nativeBenchCBID32ToDecimal64, intelCBenchmarkBID64, true},
	{"to_decimal128", nativeBenchCBID32ToDecimal128, intelCBenchmarkBID128, true},
	{"parse", nativeBenchCBID32Parse, intelCBenchmarkBID32, true},
	{"to_string", nativeBenchCBID32ToString, intelCBenchmarkString, true},
}

var intelCBID64BenchmarkRows = []intelCBenchmarkRow{
	{"add", nativeBenchCBID64Add, intelCBenchmarkBID64, true},
	{"mul", nativeBenchCBID64Mul, intelCBenchmarkBID64, true},
	{"sub", nativeBenchCBID64Sub, intelCBenchmarkBID64, true},
	{"div", nativeBenchCBID64Div, intelCBenchmarkBID64, true},
	{"fma", nativeBenchCBID64Fma, intelCBenchmarkBID64, true},
	{"sqrt", nativeBenchCBID64Sqrt, intelCBenchmarkBID64, true},
	{"remainder", nativeBenchCBID64Remainder, intelCBenchmarkBID64, true},
	{"fmod", nativeBenchCBID64Fmod, intelCBenchmarkBID64, true},
	{"quantize", nativeBenchCBID64Quantize, intelCBenchmarkBID64, true},
	{"scaleb", nativeBenchCBID64ScaleB, intelCBenchmarkBID64, true},
	{"quiet_equal", nativeBenchCBID64QuietEqual, intelCBenchmarkBool64, true},
	{"minnum", nativeBenchCBID64MinNum, intelCBenchmarkBID64, true},
	{"maxnum", nativeBenchCBID64MaxNum, intelCBenchmarkBID64, true},
	{"from_int64", nativeBenchCBID64FromInt64, intelCBenchmarkBID64, true},
	{"to_int64", nativeBenchCBID64ToInt64, intelCBenchmarkInt64, true},
	{"to_decimal32", nativeBenchCBID64ToDecimal32, intelCBenchmarkBID32, true},
	{"to_decimal128", nativeBenchCBID64ToDecimal128, intelCBenchmarkBID128, true},
	{"parse", nativeBenchCBID64Parse, intelCBenchmarkBID64, true},
	{"to_string", nativeBenchCBID64ToString, intelCBenchmarkString, true},
}

var intelCBID128BenchmarkRows = []intelCBenchmarkRow{
	{"add", nativeBenchCBID128Add, intelCBenchmarkBID128, true},
	{"mul", nativeBenchCBID128Mul, intelCBenchmarkBID128, true},
	{"sub", nativeBenchCBID128Sub, intelCBenchmarkBID128, true},
	{"div", nativeBenchCBID128Div, intelCBenchmarkBID128, true},
	{"fma", nativeBenchCBID128Fma, intelCBenchmarkBID128, true},
	{"sqrt", nativeBenchCBID128Sqrt, intelCBenchmarkBID128, true},
	{"remainder", nativeBenchCBID128Remainder, intelCBenchmarkBID128, true},
	{"fmod", nativeBenchCBID128Fmod, intelCBenchmarkBID128, true},
	{"quantize", nativeBenchCBID128Quantize, intelCBenchmarkBID128, true},
	{"scaleb", nativeBenchCBID128ScaleB, intelCBenchmarkBID128, true},
	{"quiet_equal", nativeBenchCBID128QuietEqual, intelCBenchmarkBool64, true},
	{"minnum", nativeBenchCBID128MinNum, intelCBenchmarkBID128, true},
	{"maxnum", nativeBenchCBID128MaxNum, intelCBenchmarkBID128, true},
	{"from_int64", nativeBenchCBID128FromInt64, intelCBenchmarkBID128, false},
	{"to_int64", nativeBenchCBID128ToInt64, intelCBenchmarkInt64, true},
	{"to_decimal32", nativeBenchCBID128ToDecimal32, intelCBenchmarkBID32, true},
	{"to_decimal64", nativeBenchCBID128ToDecimal64, intelCBenchmarkBID64, true},
	{"parse", nativeBenchCBID128Parse, intelCBenchmarkBID128, true},
	{"to_string", nativeBenchCBID128ToString, intelCBenchmarkString, true},
}

var intelCMixedBID64BenchmarkRows = []intelCBenchmarkRow{
	{"dq_add", nativeBenchCBID64DQAdd, intelCBenchmarkBID64, true},
	{"qd_add", nativeBenchCBID64QDAdd, intelCBenchmarkBID64, true},
	{"qq_add", nativeBenchCBID64QQAdd, intelCBenchmarkBID64, true},
	{"dq_sub", nativeBenchCBID64DQSub, intelCBenchmarkBID64, true},
	{"qd_sub", nativeBenchCBID64QDSub, intelCBenchmarkBID64, true},
	{"qq_sub", nativeBenchCBID64QQSub, intelCBenchmarkBID64, true},
	{"dq_mul", nativeBenchCBID64DQMul, intelCBenchmarkBID64, true},
	{"qd_mul", nativeBenchCBID64QDMul, intelCBenchmarkBID64, true},
	{"qq_mul", nativeBenchCBID64QQMul, intelCBenchmarkBID64, true},
	{"dq_div", nativeBenchCBID64DQDiv, intelCBenchmarkBID64, true},
	{"qd_div", nativeBenchCBID64QDDiv, intelCBenchmarkBID64, true},
	{"qq_div", nativeBenchCBID64QQDiv, intelCBenchmarkBID64, true},
}

var intelCMixedBID128BenchmarkRows = []intelCBenchmarkRow{
	{"dd_add", nativeBenchCBID128DDAdd, intelCBenchmarkBID128, true},
	{"dq_add", nativeBenchCBID128DQAdd, intelCBenchmarkBID128, true},
	{"qd_add", nativeBenchCBID128QDAdd, intelCBenchmarkBID128, true},
	{"dd_sub", nativeBenchCBID128DDSub, intelCBenchmarkBID128, true},
	{"dq_sub", nativeBenchCBID128DQSub, intelCBenchmarkBID128, true},
	{"qd_sub", nativeBenchCBID128QDSub, intelCBenchmarkBID128, true},
	{"dd_mul", nativeBenchCBID128DDMul, intelCBenchmarkBID128, true},
	{"dq_mul", nativeBenchCBID128DQMul, intelCBenchmarkBID128, true},
	{"qd_mul", nativeBenchCBID128QDMul, intelCBenchmarkBID128, true},
	{"dd_div", nativeBenchCBID128DDDiv, intelCBenchmarkBID128, true},
	{"dq_div", nativeBenchCBID128DQDiv, intelCBenchmarkBID128, true},
	{"qd_div", nativeBenchCBID128QDDiv, intelCBenchmarkBID128, true},
}

func runIntelCBIDBenchmarkRows(b *testing.B, rows []intelCBenchmarkRow) {
	b.Helper()
	for _, row := range rows {
		row := row
		b.Run(row.name, func(b *testing.B) {
			runIntelCBIDBench(b, row.run)
		})
	}
}

func BenchmarkIntelCBID32(b *testing.B) {
	runIntelCBIDBenchmarkRows(b, intelCBID32BenchmarkRows)
}

func BenchmarkIntelCBID64(b *testing.B) {
	runIntelCBIDBenchmarkRows(b, intelCBID64BenchmarkRows)
}

func BenchmarkIntelCBID128(b *testing.B) {
	runIntelCBIDBenchmarkRows(b, intelCBID128BenchmarkRows)
}

func BenchmarkIntelCMixedBID64(b *testing.B) {
	runIntelCBIDBenchmarkRows(b, intelCMixedBID64BenchmarkRows)
}

func BenchmarkIntelCMixedBID128(b *testing.B) {
	runIntelCBIDBenchmarkRows(b, intelCMixedBID128BenchmarkRows)
}
