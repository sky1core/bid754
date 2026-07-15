package bidgo

import "testing"

type fairBenchmarkRow struct {
	name string
	run  func(int)
}

func runFairBenchmarkRows(b *testing.B, rows []fairBenchmarkRow) {
	b.Helper()
	for _, row := range rows {
		row := row
		b.Run(row.name, func(b *testing.B) {
			b.ReportAllocs()
			row.run(b.N)
		})
	}
}

func makeFairBenchmarkRows(names []string, run func(string, int)) []fairBenchmarkRow {
	rows := make([]fairBenchmarkRow, 0, len(names))
	for _, name := range names {
		name := name
		rows = append(rows, fairBenchmarkRow{
			name: name,
			run:  func(n int) { run(name, n) },
		})
	}
	return rows
}

func fairBID32BenchmarkRows(tb testing.TB) []fairBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return fairBID32BenchmarkRowsForInputs(tb, inputs)
}

func fairBID32BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []fairBenchmarkRow {
	tb.Helper()
	validateBenchmarkInputs(tb, inputs)
	x := exactBenchmarkDecimal32(tb, inputs.Decimal32.X)
	y := exactBenchmarkDecimal32(tb, inputs.Decimal32.Y)
	z := exactBenchmarkDecimal32(tb, inputs.Decimal32.Z)
	integerValue, integerFlags := Bid32FromInt64(inputs.IntegerOperand, 0)
	if integerFlags != 0 {
		tb.Fatalf("Decimal32 integer benchmark operand raised flags: %#x", integerFlags)
	}
	names := []string{
		"add", "mul", "sub", "div",
		"add_pure", "mul_pure", "sub_pure", "div_pure",
		"fma", "sqrt", "remainder", "fmod", "quantize", "scaleb",
		"quiet_equal", "minnum", "maxnum", "from_int64", "to_int64",
		"to_decimal64", "to_decimal128", "parse", "to_string",
	}
	return makeFairBenchmarkRows(names, func(name string, n int) {
		runFairBID32BenchmarkRow(name, n, inputs, x, y, z, integerValue)
	})
}

func runFairBID32BenchmarkRow(name string, n int, inputs benchmarkInputs, x, y, z, integerValue uint32) {
	switch name {
	case "add":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32AddWithFlags(x, y, 0)
		}
	case "mul":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32MulWithFlags(x, y, 0)
		}
	case "sub":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32SubWithFlags(x, y, 0)
		}
	case "div":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32DivWithFlags(x, y, 0)
		}
	case "add_pure":
		for i := 0; i < n; i++ {
			sink32 = Bid32Add(x, y, 0)
		}
	case "mul_pure":
		for i := 0; i < n; i++ {
			sink32 = Bid32Mul(x, y, 0)
		}
	case "sub_pure":
		for i := 0; i < n; i++ {
			sink32 = Bid32Sub(x, y, 0)
		}
	case "div_pure":
		for i := 0; i < n; i++ {
			sink32 = Bid32Div(x, y, 0)
		}
	case "fma":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32Fma(x, y, z, 0)
		}
	case "sqrt":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32Sqrt(x, 0)
		}
	case "remainder":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32Rem(y, x)
		}
	case "fmod":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32Fmod(y, x)
		}
	case "quantize":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32Quantize(x, integerValue, 0)
		}
	case "scaleb":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32ScalbnWithFlags(x, inputs.ScaleExponent, 0)
		}
	case "quiet_equal":
		for i := 0; i < n; i++ {
			sinkInt, sinkFlags = Bid32QuietEqual(x, y)
		}
	case "minnum":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32MinNumWithFlags(x, y)
		}
	case "maxnum":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32MaxNumWithFlags(x, y)
		}
	case "from_int64":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32FromInt64(inputs.IntegerOperand, 0)
		}
	case "to_int64":
		for i := 0; i < n; i++ {
			sinkInt64, sinkFlags = Bid32ToInt64Rnint(integerValue)
		}
	case "to_decimal64":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid32ToBid64(x)
		}
	case "to_decimal128":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid32ToBid128(x)
		}
	case "parse":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid32FromStringRaw(inputs.Decimal32.X, 0)
		}
	case "to_string":
		for i := 0; i < n; i++ {
			sinkString = Bid32ToStringRaw(x)
		}
	default:
		panic("unknown Decimal32 benchmark row: " + name)
	}
}

func fairBID64BenchmarkRows(tb testing.TB) []fairBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return fairBID64BenchmarkRowsForInputs(tb, inputs)
}

func fairBID64BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []fairBenchmarkRow {
	tb.Helper()
	validateBenchmarkInputs(tb, inputs)
	x := exactBenchmarkDecimal64(tb, inputs.Decimal64.X)
	y := exactBenchmarkDecimal64(tb, inputs.Decimal64.Y)
	z := exactBenchmarkDecimal64(tb, inputs.Decimal64.Z)
	integerValue, integerFlags := Bid64FromInt64(inputs.IntegerOperand, 0)
	if integerFlags != 0 {
		tb.Fatalf("Decimal64 integer benchmark operand raised flags: %#x", integerFlags)
	}
	names := []string{
		"add", "mul", "sub", "div", "fma", "sqrt", "remainder", "fmod",
		"quantize", "scaleb", "quiet_equal", "minnum", "maxnum",
		"from_int64", "to_int64", "to_decimal32", "to_decimal128", "parse", "to_string",
	}
	return makeFairBenchmarkRows(names, func(name string, n int) {
		runFairBID64BenchmarkRow(name, n, inputs, x, y, z, integerValue)
	})
}

func runFairBID64BenchmarkRow(name string, n int, inputs benchmarkInputs, x, y, z, integerValue uint64) {
	switch name {
	case "add":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64AddWithFlags(x, y, 0)
		}
	case "mul":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64MulWithFlags(x, y, 0)
		}
	case "sub":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64SubWithFlags(x, y, 0)
		}
	case "div":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64DivWithFlags(x, y, 0)
		}
	case "fma":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64Fma(x, y, z, 0)
		}
	case "sqrt":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64Sqrt(x, 0)
		}
	case "remainder":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64Rem(y, x)
		}
	case "fmod":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64Fmod(y, x)
		}
	case "quantize":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64Quantize(x, integerValue, 0)
		}
	case "scaleb":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64Scalbn(x, inputs.ScaleExponent, 0)
		}
	case "quiet_equal":
		for i := 0; i < n; i++ {
			sinkInt, sinkFlags = Bid64QuietEqual(x, y)
		}
	case "minnum":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64MinNum(x, y)
		}
	case "maxnum":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64MaxNum(x, y)
		}
	case "from_int64":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64FromInt64(inputs.IntegerOperand, 0)
		}
	case "to_int64":
		for i := 0; i < n; i++ {
			sinkInt64, sinkFlags = Bid64ToInt64Rnint(integerValue)
		}
	case "to_decimal32":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid64ToBid32(x, 0)
		}
	case "to_decimal128":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid64ToBid128(x)
		}
	case "parse":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64FromString(inputs.Decimal64.X, 0)
		}
	case "to_string":
		for i := 0; i < n; i++ {
			sinkString = Bid64ToString(x)
		}
	default:
		panic("unknown Decimal64 benchmark row: " + name)
	}
}

func fairBID128BenchmarkRows(tb testing.TB) []fairBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return fairBID128BenchmarkRowsForInputs(tb, inputs)
}

func fairBID128BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []fairBenchmarkRow {
	tb.Helper()
	validateBenchmarkInputs(tb, inputs)
	x := exactBenchmarkDecimal128(tb, inputs.Decimal128.X)
	y := exactBenchmarkDecimal128(tb, inputs.Decimal128.Y)
	z := exactBenchmarkDecimal128(tb, inputs.Decimal128.Z)
	integerValue := Bid128FromInt64(inputs.IntegerOperand)
	names := []string{
		"add", "mul", "sub", "div", "fma", "sqrt", "remainder", "fmod",
		"quantize", "scaleb", "quiet_equal", "minnum", "maxnum",
		"from_int64", "to_int64", "to_decimal32", "to_decimal64", "parse", "to_string",
	}
	return makeFairBenchmarkRows(names, func(name string, n int) {
		runFairBID128BenchmarkRow(name, n, inputs, x, y, z, integerValue)
	})
}

func runFairBID128BenchmarkRow(name string, n int, inputs benchmarkInputs, x, y, z, integerValue BID_UINT128) {
	switch name {
	case "add":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = Bid128Add(x, y, 0, &flags)
			sinkFlags = flags
		}
	case "mul":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128Mul(x, y, 0)
		}
	case "sub":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = Bid128Sub(x, y, 0, &flags)
			sinkFlags = flags
		}
	case "div":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128Div(x, y, 0)
		}
	case "fma":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128Fma(x, y, z, 0)
		}
	case "sqrt":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128Sqrt(x, 0)
		}
	case "remainder":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128Rem(y, x)
		}
	case "fmod":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128Fmod(y, x)
		}
	case "quantize":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128Quantize(x, integerValue, 0)
		}
	case "scaleb":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = Bid128Scalbn(x, inputs.ScaleExponent, 0, &flags)
			sinkFlags = flags
		}
	case "quiet_equal":
		for i := 0; i < n; i++ {
			sinkInt, sinkFlags = Bid128QuietEqual(x, y)
		}
	case "minnum":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = Bid128Minnum(x, y, &flags)
			sinkFlags = flags
		}
	case "maxnum":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = Bid128Maxnum(x, y, &flags)
			sinkFlags = flags
		}
	case "from_int64":
		for i := 0; i < n; i++ {
			sink128 = Bid128FromInt64(inputs.IntegerOperand)
		}
	case "to_int64":
		for i := 0; i < n; i++ {
			sinkInt64, sinkFlags = Bid128ToInt64Rnint(integerValue)
		}
	case "to_decimal32":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = Bid128ToBid32(x, 0)
		}
	case "to_decimal64":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid128ToBid64(x, 0)
		}
	case "parse":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128FromString(inputs.Decimal128.X, 0)
		}
	case "to_string":
		for i := 0; i < n; i++ {
			sinkString = Bid128ToString(x)
		}
	default:
		panic("unknown Decimal128 benchmark row: " + name)
	}
}

func fairMixedBID64BenchmarkRows(tb testing.TB) []fairBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return fairMixedBID64BenchmarkRowsForInputs(tb, inputs)
}

func fairMixedBID64BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []fairBenchmarkRow {
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
	return makeFairBenchmarkRows(names, func(name string, n int) {
		runFairMixedBID64BenchmarkRow(name, n, dx, dy, qx, qy)
	})
}

func runFairMixedBID64BenchmarkRow(name string, n int, dx, dy uint64, qx, qy BID_UINT128) {
	switch name {
	case "dq_add":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64dqAdd(dx, qy, 0)
		}
	case "qd_add":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64qdAdd(qx, dy, 0)
		}
	case "qq_add":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64qqAdd(qx, qy, 0)
		}
	case "dq_sub":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64dqSub(dx, qy, 0)
		}
	case "qd_sub":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64qdSub(qx, dy, 0)
		}
	case "qq_sub":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64qqSub(qx, qy, 0)
		}
	case "dq_mul":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64dqMul(dx, qy, 0)
		}
	case "qd_mul":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64qdMul(qx, dy, 0)
		}
	case "qq_mul":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64qqMul(qx, qy, 0)
		}
	case "dq_div":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64dqDiv(dx, qy, 0)
		}
	case "qd_div":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64qdDiv(qx, dy, 0)
		}
	case "qq_div":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = Bid64qqDiv(qx, qy, 0)
		}
	default:
		panic("unknown mixed Decimal64 benchmark row: " + name)
	}
}

func fairMixedBID128BenchmarkRows(tb testing.TB) []fairBenchmarkRow {
	tb.Helper()
	inputs := loadBenchmarkInputs(tb)
	return fairMixedBID128BenchmarkRowsForInputs(tb, inputs)
}

func fairMixedBID128BenchmarkRowsForInputs(tb testing.TB, inputs benchmarkInputs) []fairBenchmarkRow {
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
	return makeFairBenchmarkRows(names, func(name string, n int) {
		runFairMixedBID128BenchmarkRow(name, n, dx, dy, qx, qy)
	})
}

func runFairMixedBID128BenchmarkRow(name string, n int, dx, dy uint64, qx, qy BID_UINT128) {
	switch name {
	case "dd_add":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128ddAdd(dx, dy, 0)
		}
	case "dq_add":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128dqAdd(dx, qy, 0)
		}
	case "qd_add":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128qdAdd(qx, dy, 0)
		}
	case "dd_sub":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128ddSub(dx, dy, 0)
		}
	case "dq_sub":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128dqSub(dx, qy, 0)
		}
	case "qd_sub":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128qdSub(qx, dy, 0)
		}
	case "dd_mul":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128ddMul(dx, dy, 0)
		}
	case "dq_mul":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128dqMul(dx, qy, 0)
		}
	case "qd_mul":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128qdMul(qx, dy, 0)
		}
	case "dd_div":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128ddDiv(dx, dy, 0)
		}
	case "dq_div":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128dqDiv(dx, qy, 0)
		}
	case "qd_div":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = Bid128qdDiv(qx, dy, 0)
		}
	default:
		panic("unknown mixed Decimal128 benchmark row: " + name)
	}
}
