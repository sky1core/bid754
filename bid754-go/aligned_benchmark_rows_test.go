//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

import "testing"

type alignedBenchmarkRow struct {
	name string
	run  func(int)
}

func runAlignedBenchmarkRows(b *testing.B, rows []alignedBenchmarkRow) {
	b.Helper()
	for _, row := range rows {
		row := row
		b.Run(row.name, func(b *testing.B) {
			b.ReportAllocs()
			row.run(b.N)
		})
	}
}

func makeAlignedBenchmarkRows(names []string, run func(string, int)) []alignedBenchmarkRow {
	rows := make([]alignedBenchmarkRow, 0, len(names))
	for _, name := range names {
		name := name
		rows = append(rows, alignedBenchmarkRow{
			name: name,
			run:  func(n int) { run(name, n) },
		})
	}
	return rows
}

func alignedBID32BenchmarkRows(tb testing.TB) []alignedBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return alignedBID32BenchmarkRowsForInputs(tb, inputs)
}

func alignedBID32BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []alignedBenchmarkRow {
	tb.Helper()
	validateBenchmarkInputs(tb, inputs)
	x := exactBenchmarkDecimal32(tb, inputs.Decimal32.X)
	y := exactBenchmarkDecimal32(tb, inputs.Decimal32.Y)
	z := exactBenchmarkDecimal32(tb, inputs.Decimal32.Z)
	integerValue, integerFlags := NewDecimal32FromInt64(inputs.IntegerOperand, RoundNearestEven)
	if integerFlags != 0 {
		tb.Fatalf("Decimal32 integer benchmark operand raised flags: %v", integerFlags)
	}
	names := []string{
		"add", "mul", "sub", "div",
		"add_with_flags", "mul_with_flags", "sub_with_flags", "div_with_flags",
		"fma", "sqrt", "remainder", "fmod", "quantize", "scaleb",
		"quiet_equal", "minnum", "maxnum", "from_int64", "to_int64",
		"to_decimal64", "to_decimal128", "parse", "to_string",
	}
	return makeAlignedBenchmarkRows(names, func(name string, n int) {
		runAlignedBID32BenchmarkRow(name, n, inputs, x, y, z, integerValue)
	})
}

func runAlignedBID32BenchmarkRow(name string, n int, inputs benchmarkInputs, x, y, z, integerValue Decimal32BID) {
	switch name {
	case "add":
		for i := 0; i < n; i++ {
			alignedSink32 = x.Add(y)
		}
	case "mul":
		for i := 0; i < n; i++ {
			alignedSink32 = x.Mul(y)
		}
	case "sub":
		for i := 0; i < n; i++ {
			alignedSink32 = x.Sub(y)
		}
	case "div":
		for i := 0; i < n; i++ {
			alignedSink32 = x.Div(y)
		}
	case "add_with_flags":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.AddWithFlags(y)
		}
	case "mul_with_flags":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.MulWithFlags(y)
		}
	case "sub_with_flags":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.SubWithFlags(y)
		}
	case "div_with_flags":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.DivWithFlags(y)
		}
	case "fma":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.FMA(y, z)
		}
	case "sqrt":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.Sqrt()
		}
	case "remainder":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = y.Remainder(x)
		}
	case "fmod":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = y.Fmod(x)
		}
	case "quantize":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.QuantizeWithFlags(integerValue)
		}
	case "scaleb":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.ScaleB(inputs.ScaleExponent)
		}
	case "quiet_equal":
		for i := 0; i < n; i++ {
			alignedSinkBool, alignedSinkFlags = x.QuietEqual(y)
		}
	case "minnum":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.MinNum(y)
		}
	case "maxnum":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.MaxNum(y)
		}
	case "from_int64":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = NewDecimal32FromInt64(inputs.IntegerOperand, RoundNearestEven)
		}
	case "to_int64":
		for i := 0; i < n; i++ {
			alignedSinkInt64, alignedSinkFlags = integerValue.ConvertToInt64(RoundNearestEven)
		}
	case "to_decimal64":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.ToDecimal64()
		}
	case "to_decimal128":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.ToDecimal128()
		}
	case "parse":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkErr = NewDecimal32BIDDirect(inputs.Decimal32.X)
		}
	case "to_string":
		for i := 0; i < n; i++ {
			alignedSinkString = x.String()
		}
	default:
		panic("unknown Decimal32 benchmark row: " + name)
	}
}

func alignedBID64BenchmarkRows(tb testing.TB) []alignedBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return alignedBID64BenchmarkRowsForInputs(tb, inputs)
}

func alignedBID64BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []alignedBenchmarkRow {
	tb.Helper()
	validateBenchmarkInputs(tb, inputs)
	x := exactBenchmarkDecimal64(tb, inputs.Decimal64.X)
	y := exactBenchmarkDecimal64(tb, inputs.Decimal64.Y)
	z := exactBenchmarkDecimal64(tb, inputs.Decimal64.Z)
	integerValue, integerFlags := NewDecimal64FromInt64(inputs.IntegerOperand, RoundNearestEven)
	if integerFlags != 0 {
		tb.Fatalf("Decimal64 integer benchmark operand raised flags: %v", integerFlags)
	}
	names := []string{
		"add", "mul", "sub", "div",
		"add_with_flags", "mul_with_flags", "sub_with_flags", "div_with_flags",
		"fma", "sqrt", "remainder", "fmod", "quantize", "scaleb",
		"quiet_equal", "minnum", "maxnum", "from_int64", "to_int64",
		"to_decimal32", "to_decimal128", "parse", "to_string",
	}
	return makeAlignedBenchmarkRows(names, func(name string, n int) {
		runAlignedBID64BenchmarkRow(name, n, inputs, x, y, z, integerValue)
	})
}

func runAlignedBID64BenchmarkRow(name string, n int, inputs benchmarkInputs, x, y, z, integerValue Decimal64BID) {
	switch name {
	case "add":
		for i := 0; i < n; i++ {
			alignedSink64 = x.Add(y)
		}
	case "mul":
		for i := 0; i < n; i++ {
			alignedSink64 = x.Mul(y)
		}
	case "sub":
		for i := 0; i < n; i++ {
			alignedSink64 = x.Sub(y)
		}
	case "div":
		for i := 0; i < n; i++ {
			alignedSink64 = x.Div(y)
		}
	case "add_with_flags":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.AddWithFlags(y)
		}
	case "mul_with_flags":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.MulWithFlags(y)
		}
	case "sub_with_flags":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.SubWithFlags(y)
		}
	case "div_with_flags":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.DivWithFlags(y)
		}
	case "fma":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.FMA(y, z)
		}
	case "sqrt":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.Sqrt()
		}
	case "remainder":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = y.Remainder(x)
		}
	case "fmod":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = y.Fmod(x)
		}
	case "quantize":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.QuantizeWithFlags(integerValue)
		}
	case "scaleb":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.ScaleB(inputs.ScaleExponent)
		}
	case "quiet_equal":
		for i := 0; i < n; i++ {
			alignedSinkBool, alignedSinkFlags = x.QuietEqual(y)
		}
	case "minnum":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.MinNum(y)
		}
	case "maxnum":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.MaxNum(y)
		}
	case "from_int64":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = NewDecimal64FromInt64(inputs.IntegerOperand, RoundNearestEven)
		}
	case "to_int64":
		for i := 0; i < n; i++ {
			alignedSinkInt64, alignedSinkFlags = integerValue.ConvertToInt64(RoundNearestEven)
		}
	case "to_decimal32":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.ToDecimal32(RoundNearestEven)
		}
	case "to_decimal128":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.ToDecimal128()
		}
	case "parse":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkErr = NewDecimal64BIDDirect(inputs.Decimal64.X)
		}
	case "to_string":
		for i := 0; i < n; i++ {
			alignedSinkString = x.String()
		}
	default:
		panic("unknown Decimal64 benchmark row: " + name)
	}
}

func alignedBID128BenchmarkRows(tb testing.TB) []alignedBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return alignedBID128BenchmarkRowsForInputs(tb, inputs)
}

func alignedBID128BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []alignedBenchmarkRow {
	tb.Helper()
	validateBenchmarkInputs(tb, inputs)
	x := exactBenchmarkDecimal128(tb, inputs.Decimal128.X)
	y := exactBenchmarkDecimal128(tb, inputs.Decimal128.Y)
	z := exactBenchmarkDecimal128(tb, inputs.Decimal128.Z)
	integerValue := NewDecimal128FromInt64(inputs.IntegerOperand)
	names := []string{
		"add", "mul", "sub", "div",
		"add_with_flags", "mul_with_flags", "sub_with_flags", "div_with_flags",
		"fma", "sqrt", "remainder", "fmod", "quantize", "scaleb",
		"quiet_equal", "minnum", "maxnum", "from_int64", "to_int64",
		"to_decimal32", "to_decimal64", "parse", "to_string",
	}
	return makeAlignedBenchmarkRows(names, func(name string, n int) {
		runAlignedBID128BenchmarkRow(name, n, inputs, x, y, z, integerValue)
	})
}

func runAlignedBID128BenchmarkRow(name string, n int, inputs benchmarkInputs, x, y, z, integerValue Decimal128BID) {
	switch name {
	case "add":
		for i := 0; i < n; i++ {
			alignedSink128 = x.Add(y)
		}
	case "mul":
		for i := 0; i < n; i++ {
			alignedSink128 = x.Mul(y)
		}
	case "sub":
		for i := 0; i < n; i++ {
			alignedSink128 = x.Sub(y)
		}
	case "div":
		for i := 0; i < n; i++ {
			alignedSink128 = x.Div(y)
		}
	case "add_with_flags":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.AddWithFlags(y)
		}
	case "mul_with_flags":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.MulWithFlags(y)
		}
	case "sub_with_flags":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.SubWithFlags(y)
		}
	case "div_with_flags":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.DivWithFlags(y)
		}
	case "fma":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.FMA(y, z)
		}
	case "sqrt":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.Sqrt()
		}
	case "remainder":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = y.Remainder(x)
		}
	case "fmod":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = y.Fmod(x)
		}
	case "quantize":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.QuantizeWithFlags(integerValue)
		}
	case "scaleb":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.ScaleB(inputs.ScaleExponent)
		}
	case "quiet_equal":
		for i := 0; i < n; i++ {
			alignedSinkBool, alignedSinkFlags = x.QuietEqual(y)
		}
	case "minnum":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.MinNum(y)
		}
	case "maxnum":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = x.MaxNum(y)
		}
	case "from_int64":
		for i := 0; i < n; i++ {
			alignedSink128 = NewDecimal128FromInt64(inputs.IntegerOperand)
		}
	case "to_int64":
		for i := 0; i < n; i++ {
			alignedSinkInt64, alignedSinkFlags = integerValue.ConvertToInt64(RoundNearestEven)
		}
	case "to_decimal32":
		for i := 0; i < n; i++ {
			alignedSink32, alignedSinkFlags = x.ToDecimal32(RoundNearestEven)
		}
	case "to_decimal64":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = x.ToDecimal64(RoundNearestEven)
		}
	case "parse":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkErr = NewDecimal128BIDDirect(inputs.Decimal128.X)
		}
	case "to_string":
		for i := 0; i < n; i++ {
			alignedSinkString = x.String()
		}
	default:
		panic("unknown Decimal128 benchmark row: " + name)
	}
}

func alignedMixedBID64BenchmarkRows(tb testing.TB) []alignedBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return alignedMixedBID64BenchmarkRowsForInputs(tb, inputs)
}

func alignedMixedBID64BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []alignedBenchmarkRow {
	tb.Helper()
	validateBenchmarkInputs(tb, inputs)
	dx := exactBenchmarkDecimal64(tb, inputs.Decimal64.X)
	dy := exactBenchmarkDecimal64(tb, inputs.Decimal64.Y)
	qx := exactBenchmarkDecimal128(tb, inputs.Decimal128.X)
	qy := exactBenchmarkDecimal128(tb, inputs.Decimal128.Y)
	names := []string{
		"dq_add", "qd_add", "qq_add",
		"dq_sub", "qd_sub", "qq_sub",
		"dq_mul", "qd_mul", "qq_mul",
		"dq_div", "qd_div", "qq_div",
	}
	return makeAlignedBenchmarkRows(names, func(name string, n int) {
		runAlignedMixedBID64BenchmarkRow(name, n, dx, dy, qx, qy)
	})
}

func runAlignedMixedBID64BenchmarkRow(name string, n int, dx, dy Decimal64BID, qx, qy Decimal128BID) {
	switch name {
	case "dq_add":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Add64DQBIDWithMode(dx, qy, RoundNearestEven)
		}
	case "qd_add":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Add64QDBIDWithMode(qx, dy, RoundNearestEven)
		}
	case "qq_add":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Add64QQBIDWithMode(qx, qy, RoundNearestEven)
		}
	case "dq_sub":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Sub64DQBIDWithMode(dx, qy, RoundNearestEven)
		}
	case "qd_sub":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Sub64QDBIDWithMode(qx, dy, RoundNearestEven)
		}
	case "qq_sub":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Sub64QQBIDWithMode(qx, qy, RoundNearestEven)
		}
	case "dq_mul":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Mul64DQBIDWithMode(dx, qy, RoundNearestEven)
		}
	case "qd_mul":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Mul64QDBIDWithMode(qx, dy, RoundNearestEven)
		}
	case "qq_mul":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Mul64QQBIDWithMode(qx, qy, RoundNearestEven)
		}
	case "dq_div":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Div64DQBIDWithMode(dx, qy, RoundNearestEven)
		}
	case "qd_div":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Div64QDBIDWithMode(qx, dy, RoundNearestEven)
		}
	case "qq_div":
		for i := 0; i < n; i++ {
			alignedSink64, alignedSinkFlags = Div64QQBIDWithMode(qx, qy, RoundNearestEven)
		}
	default:
		panic("unknown mixed Decimal64 benchmark row: " + name)
	}
}

func alignedMixedBID128BenchmarkRows(tb testing.TB) []alignedBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return alignedMixedBID128BenchmarkRowsForInputs(tb, inputs)
}

func alignedMixedBID128BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []alignedBenchmarkRow {
	tb.Helper()
	validateBenchmarkInputs(tb, inputs)
	dx := exactBenchmarkDecimal64(tb, inputs.Decimal64.X)
	dy := exactBenchmarkDecimal64(tb, inputs.Decimal64.Y)
	qx := exactBenchmarkDecimal128(tb, inputs.Decimal128.X)
	qy := exactBenchmarkDecimal128(tb, inputs.Decimal128.Y)
	names := []string{
		"dd_add", "dq_add", "qd_add",
		"dd_sub", "dq_sub", "qd_sub",
		"dd_mul", "dq_mul", "qd_mul",
		"dd_div", "dq_div", "qd_div",
	}
	return makeAlignedBenchmarkRows(names, func(name string, n int) {
		runAlignedMixedBID128BenchmarkRow(name, n, dx, dy, qx, qy)
	})
}

func runAlignedMixedBID128BenchmarkRow(name string, n int, dx, dy Decimal64BID, qx, qy Decimal128BID) {
	switch name {
	case "dd_add":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Add128DDBIDWithMode(dx, dy, RoundNearestEven)
		}
	case "dq_add":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Add128DQBIDWithMode(dx, qy, RoundNearestEven)
		}
	case "qd_add":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Add128QDBIDWithMode(qx, dy, RoundNearestEven)
		}
	case "dd_sub":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Sub128DDBIDWithMode(dx, dy, RoundNearestEven)
		}
	case "dq_sub":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Sub128DQBIDWithMode(dx, qy, RoundNearestEven)
		}
	case "qd_sub":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Sub128QDBIDWithMode(qx, dy, RoundNearestEven)
		}
	case "dd_mul":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Mul128DDBIDWithMode(dx, dy, RoundNearestEven)
		}
	case "dq_mul":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Mul128DQBIDWithMode(dx, qy, RoundNearestEven)
		}
	case "qd_mul":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Mul128QDBIDWithMode(qx, dy, RoundNearestEven)
		}
	case "dd_div":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Div128DDBIDWithMode(dx, dy, RoundNearestEven)
		}
	case "dq_div":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Div128DQBIDWithMode(dx, qy, RoundNearestEven)
		}
	case "qd_div":
		for i := 0; i < n; i++ {
			alignedSink128, alignedSinkFlags = Div128QDBIDWithMode(qx, dy, RoundNearestEven)
		}
	default:
		panic("unknown mixed Decimal128 benchmark row: " + name)
	}
}
