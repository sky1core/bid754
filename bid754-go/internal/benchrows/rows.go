package benchrows

import (
	"fmt"
	"math"

	"github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

// The Go mechanical-port benchmark rows and their observable sinks. The
// timing-loop bodies are moved verbatim from the previous in-package bidgo
// benchmark table (operation, operand binding, and one global sink store per
// iteration are unchanged); only the bidgo package qualifier is added.

var (
	sink64     uint64
	sink32     uint32
	sink128    bidgo.BID_UINT128
	sinkString string
	sinkFlags  uint32
	sinkInt64  int64
	sinkInt    int
)

// Row is one named Go-port benchmark row. Run executes the row's timing loop
// n times (benchmarks pass b.N; the preflight and the sink-discipline test
// pass 1).
type Row struct {
	Name string
	run  func(int)
}

// Run executes the row's timing loop with n iterations.
func (r Row) Run(n int) { r.run(n) }

func makeRows(names []string, run func(string, int)) []Row {
	rows := make([]Row, 0, len(names))
	for _, name := range names {
		name := name
		rows = append(rows, Row{
			Name: name,
			run:  func(n int) { run(name, n) },
		})
	}
	return rows
}

// FairBID32Rows returns the Decimal32 Go-port benchmark rows.
func FairBID32Rows(p Prepared) []Row {
	names := []string{
		"add", "mul", "sub", "div",
		"add_pure", "mul_pure", "sub_pure", "div_pure",
		"fma", "sqrt", "remainder", "fmod", "quantize", "scaleb",
		"quiet_equal", "minnum", "maxnum", "from_int64", "to_int64",
		"to_decimal64", "to_decimal128", "parse", "to_string",
	}
	return makeRows(names, func(name string, n int) {
		runFairBID32BenchmarkRow(name, n, p.Inputs, p.X32, p.Y32, p.Z32, p.Integer32)
	})
}

func runFairBID32BenchmarkRow(name string, n int, inputs Inputs, x, y, z, integerValue uint32) {
	switch name {
	case "add":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32AddWithFlags(x, y, 0)
		}
	case "mul":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32MulWithFlags(x, y, 0)
		}
	case "sub":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32SubWithFlags(x, y, 0)
		}
	case "div":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32DivWithFlags(x, y, 0)
		}
	case "add_pure":
		for i := 0; i < n; i++ {
			sink32 = bidgo.Bid32Add(x, y, 0)
		}
	case "mul_pure":
		for i := 0; i < n; i++ {
			sink32 = bidgo.Bid32Mul(x, y, 0)
		}
	case "sub_pure":
		for i := 0; i < n; i++ {
			sink32 = bidgo.Bid32Sub(x, y, 0)
		}
	case "div_pure":
		for i := 0; i < n; i++ {
			sink32 = bidgo.Bid32Div(x, y, 0)
		}
	case "fma":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32Fma(x, y, z, 0)
		}
	case "sqrt":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32Sqrt(x, 0)
		}
	case "remainder":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32Rem(y, x)
		}
	case "fmod":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32Fmod(y, x)
		}
	case "quantize":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32Quantize(x, integerValue, 0)
		}
	case "scaleb":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32ScalbnWithFlags(x, inputs.ScaleExponent, 0)
		}
	case "quiet_equal":
		for i := 0; i < n; i++ {
			sinkInt, sinkFlags = bidgo.Bid32QuietEqual(x, y)
		}
	case "minnum":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32MinNumWithFlags(x, y)
		}
	case "maxnum":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32MaxNumWithFlags(x, y)
		}
	case "from_int64":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32FromInt64(inputs.IntegerOperand, 0)
		}
	case "to_int64":
		for i := 0; i < n; i++ {
			sinkInt64, sinkFlags = bidgo.Bid32ToInt64Rnint(integerValue)
		}
	case "to_decimal64":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid32ToBid64(x)
		}
	case "to_decimal128":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid32ToBid128(x)
		}
	case "parse":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid32FromStringRaw(inputs.Decimal32.X, 0)
		}
	case "to_string":
		for i := 0; i < n; i++ {
			sinkString = bidgo.Bid32ToStringRaw(x)
		}
	default:
		panic("unknown Decimal32 benchmark row: " + name)
	}
}

// FairBID64Rows returns the Decimal64 Go-port benchmark rows.
func FairBID64Rows(p Prepared) []Row {
	names := []string{
		"add", "mul", "sub", "div", "fma", "sqrt", "remainder", "fmod",
		"quantize", "scaleb", "quiet_equal", "minnum", "maxnum",
		"from_int64", "to_int64", "to_decimal32", "to_decimal128", "parse", "to_string",
	}
	return makeRows(names, func(name string, n int) {
		runFairBID64BenchmarkRow(name, n, p.Inputs, p.X64, p.Y64, p.Z64, p.Integer64)
	})
}

func runFairBID64BenchmarkRow(name string, n int, inputs Inputs, x, y, z, integerValue uint64) {
	switch name {
	case "add":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64AddWithFlags(x, y, 0)
		}
	case "mul":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64MulWithFlags(x, y, 0)
		}
	case "sub":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64SubWithFlags(x, y, 0)
		}
	case "div":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64DivWithFlags(x, y, 0)
		}
	case "fma":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64Fma(x, y, z, 0)
		}
	case "sqrt":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64Sqrt(x, 0)
		}
	case "remainder":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64Rem(y, x)
		}
	case "fmod":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64Fmod(y, x)
		}
	case "quantize":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64Quantize(x, integerValue, 0)
		}
	case "scaleb":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64Scalbn(x, inputs.ScaleExponent, 0)
		}
	case "quiet_equal":
		for i := 0; i < n; i++ {
			sinkInt, sinkFlags = bidgo.Bid64QuietEqual(x, y)
		}
	case "minnum":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64MinNum(x, y)
		}
	case "maxnum":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64MaxNum(x, y)
		}
	case "from_int64":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64FromInt64(inputs.IntegerOperand, 0)
		}
	case "to_int64":
		for i := 0; i < n; i++ {
			sinkInt64, sinkFlags = bidgo.Bid64ToInt64Rnint(integerValue)
		}
	case "to_decimal32":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid64ToBid32(x, 0)
		}
	case "to_decimal128":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid64ToBid128(x)
		}
	case "parse":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64FromString(inputs.Decimal64.X, 0)
		}
	case "to_string":
		for i := 0; i < n; i++ {
			sinkString = bidgo.Bid64ToString(x)
		}
	default:
		panic("unknown Decimal64 benchmark row: " + name)
	}
}

// FairBID128Rows returns the Decimal128 Go-port benchmark rows.
func FairBID128Rows(p Prepared) []Row {
	names := []string{
		"add", "mul", "sub", "div", "fma", "sqrt", "remainder", "fmod",
		"quantize", "scaleb", "quiet_equal", "minnum", "maxnum",
		"from_int64", "to_int64", "to_decimal32", "to_decimal64", "parse", "to_string",
	}
	return makeRows(names, func(name string, n int) {
		runFairBID128BenchmarkRow(name, n, p.Inputs, p.X128, p.Y128, p.Z128, p.Integer128)
	})
}

func runFairBID128BenchmarkRow(name string, n int, inputs Inputs, x, y, z, integerValue bidgo.BID_UINT128) {
	switch name {
	case "add":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = bidgo.Bid128Add(x, y, 0, &flags)
			sinkFlags = flags
		}
	case "mul":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128Mul(x, y, 0)
		}
	case "sub":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = bidgo.Bid128Sub(x, y, 0, &flags)
			sinkFlags = flags
		}
	case "div":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128Div(x, y, 0)
		}
	case "fma":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128Fma(x, y, z, 0)
		}
	case "sqrt":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128Sqrt(x, 0)
		}
	case "remainder":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128Rem(y, x)
		}
	case "fmod":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128Fmod(y, x)
		}
	case "quantize":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128Quantize(x, integerValue, 0)
		}
	case "scaleb":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = bidgo.Bid128Scalbn(x, inputs.ScaleExponent, 0, &flags)
			sinkFlags = flags
		}
	case "quiet_equal":
		for i := 0; i < n; i++ {
			sinkInt, sinkFlags = bidgo.Bid128QuietEqual(x, y)
		}
	case "minnum":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = bidgo.Bid128Minnum(x, y, &flags)
			sinkFlags = flags
		}
	case "maxnum":
		for i := 0; i < n; i++ {
			var flags uint32
			sink128 = bidgo.Bid128Maxnum(x, y, &flags)
			sinkFlags = flags
		}
	case "from_int64":
		for i := 0; i < n; i++ {
			sink128 = bidgo.Bid128FromInt64(inputs.IntegerOperand)
		}
	case "to_int64":
		for i := 0; i < n; i++ {
			sinkInt64, sinkFlags = bidgo.Bid128ToInt64Rnint(integerValue)
		}
	case "to_decimal32":
		for i := 0; i < n; i++ {
			sink32, sinkFlags = bidgo.Bid128ToBid32(x, 0)
		}
	case "to_decimal64":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid128ToBid64(x, 0)
		}
	case "parse":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128FromString(inputs.Decimal128.X, 0)
		}
	case "to_string":
		for i := 0; i < n; i++ {
			sinkString = bidgo.Bid128ToString(x)
		}
	default:
		panic("unknown Decimal128 benchmark row: " + name)
	}
}

// FairMixedBID64Rows returns the mixed-width Decimal64-result Go-port rows.
func FairMixedBID64Rows(p Prepared) []Row {
	names := []string{
		"dq_add", "qd_add", "qq_add",
		"dq_sub", "qd_sub", "qq_sub",
		"dq_mul", "qd_mul", "qq_mul",
		"dq_div", "qd_div", "qq_div",
	}
	return makeRows(names, func(name string, n int) {
		runFairMixedBID64BenchmarkRow(name, n, p.X64, p.Y64, p.X128, p.Y128)
	})
}

func runFairMixedBID64BenchmarkRow(name string, n int, dx, dy uint64, qx, qy bidgo.BID_UINT128) {
	switch name {
	case "dq_add":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64dqAdd(dx, qy, 0)
		}
	case "qd_add":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64qdAdd(qx, dy, 0)
		}
	case "qq_add":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64qqAdd(qx, qy, 0)
		}
	case "dq_sub":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64dqSub(dx, qy, 0)
		}
	case "qd_sub":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64qdSub(qx, dy, 0)
		}
	case "qq_sub":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64qqSub(qx, qy, 0)
		}
	case "dq_mul":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64dqMul(dx, qy, 0)
		}
	case "qd_mul":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64qdMul(qx, dy, 0)
		}
	case "qq_mul":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64qqMul(qx, qy, 0)
		}
	case "dq_div":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64dqDiv(dx, qy, 0)
		}
	case "qd_div":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64qdDiv(qx, dy, 0)
		}
	case "qq_div":
		for i := 0; i < n; i++ {
			sink64, sinkFlags = bidgo.Bid64qqDiv(qx, qy, 0)
		}
	default:
		panic("unknown mixed Decimal64 benchmark row: " + name)
	}
}

// FairMixedBID128Rows returns the mixed-width Decimal128-result Go-port rows.
func FairMixedBID128Rows(p Prepared) []Row {
	names := []string{
		"dd_add", "dq_add", "qd_add",
		"dd_sub", "dq_sub", "qd_sub",
		"dd_mul", "dq_mul", "qd_mul",
		"dd_div", "dq_div", "qd_div",
	}
	return makeRows(names, func(name string, n int) {
		runFairMixedBID128BenchmarkRow(name, n, p.X64, p.Y64, p.X128, p.Y128)
	})
}

func runFairMixedBID128BenchmarkRow(name string, n int, dx, dy uint64, qx, qy bidgo.BID_UINT128) {
	switch name {
	case "dd_add":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128ddAdd(dx, dy, 0)
		}
	case "dq_add":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128dqAdd(dx, qy, 0)
		}
	case "qd_add":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128qdAdd(qx, dy, 0)
		}
	case "dd_sub":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128ddSub(dx, dy, 0)
		}
	case "dq_sub":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128dqSub(dx, qy, 0)
		}
	case "qd_sub":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128qdSub(qx, dy, 0)
		}
	case "dd_mul":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128ddMul(dx, dy, 0)
		}
	case "dq_mul":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128dqMul(dx, qy, 0)
		}
	case "qd_mul":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128qdMul(qx, dy, 0)
		}
	case "dd_div":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128ddDiv(dx, dy, 0)
		}
	case "dq_div":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128dqDiv(dx, qy, 0)
		}
	case "qd_div":
		for i := 0; i < n; i++ {
			sink128, sinkFlags = bidgo.Bid128qdDiv(qx, dy, 0)
		}
	default:
		panic("unknown mixed Decimal128 benchmark row: " + name)
	}
}

// GroupRows returns the Go-port rows of one descriptor group.
func GroupRows(p Prepared, group string) ([]Row, error) {
	switch group {
	case "bid32":
		return FairBID32Rows(p), nil
	case "bid64":
		return FairBID64Rows(p), nil
	case "bid128":
		return FairBID128Rows(p), nil
	case "bid64_mixed":
		return FairMixedBID64Rows(p), nil
	case "bid128_mixed":
		return FairMixedBID128Rows(p), nil
	default:
		return nil, fmt.Errorf("unknown benchmark row group %q", group)
	}
}

// bid128Bits reads a bidgo 128-bit value as its low/high 64-bit words through
// the endian-agnostic word accessors the module root uses for Decimal128BID
// (types_bidgo_runtime.go). The word view keeps the low/high split correct on
// big-endian platforms, where a native-endian pointer reinterpretation would
// byte-swap each word.
func bid128Bits(x bidgo.BID_UINT128) (lo, hi uint64) {
	hi, lo = bidgo.Bid128Words(x)
	return lo, hi
}

func bid128FromBits(lo, hi uint64) bidgo.BID_UINT128 {
	return bidgo.Bid128FromWords(hi, lo)
}

// Observation is one normalized untimed benchmark row execution result:
// the value read from the row's declared sink and, when the row's status
// contract observes flags, the raw Intel status flags.
type Observation struct {
	Kind      string
	Bits32    uint32
	Bits64    uint64
	Bits128Lo uint64
	Bits128Hi uint64
	Int64     int64
	Predicate bool
	Text      string
	Flags     uint32
	HasFlags  bool
}

// Sink poison values. Every poison is unreachable for the exact finite
// benchmark operand contract, so a row that fails to write its declared sink
// leaves the poison in place and is detected.
const (
	poisonSink32     = uint32(0xa5a5a5a5)
	poisonSink64     = uint64(0xa5a5a5a5a5a5a5a5)
	poisonSink128Lo  = uint64(0xa5a5a5a5a5a5a5a5)
	poisonSink128Hi  = uint64(0x5a5a5a5a5a5a5a5a)
	poisonSinkInt64  = int64(math.MinInt64)
	poisonSinkInt    = -1
	poisonSinkFlags  = uint32(0xffffffff)
	poisonSinkString = "<benchrows:unwritten>"
)

func primeSinks() {
	sink32 = poisonSink32
	sink64 = poisonSink64
	sink128 = bid128FromBits(poisonSink128Lo, poisonSink128Hi)
	sinkString = poisonSinkString
	sinkInt64 = poisonSinkInt64
	sinkInt = poisonSinkInt
	sinkFlags = poisonSinkFlags
}

// ObserveOnce primes every sink with a poison value, executes the row's
// timing loop exactly once, and snapshots the sink declared by spec. It fails
// when the declared sink was not written, when any other value sink was
// written, or when the status-flag sink violates the declared status
// contract (flags_observed rows must write it; value_only rows must leave it
// untouched).
func ObserveOnce(row Row, spec DescriptorRow) (Observation, error) {
	if spec.Status != StatusFlagsObserved && spec.Status != StatusValueOnly {
		return Observation{}, fmt.Errorf("row %s/%s: status %q is not observable on the Go-port layer", spec.Group, spec.Name, spec.Status)
	}
	primeSinks()
	row.Run(1)

	observation := Observation{Kind: spec.Result}
	writtenSink := ""
	switch spec.Result {
	case ResultD32:
		if sink32 == poisonSink32 {
			return Observation{}, fmt.Errorf("row %s/%s did not write the Decimal32 sink", spec.Group, spec.Name)
		}
		observation.Bits32 = sink32
		writtenSink = "d32"
	case ResultD64:
		if sink64 == poisonSink64 {
			return Observation{}, fmt.Errorf("row %s/%s did not write the Decimal64 sink", spec.Group, spec.Name)
		}
		observation.Bits64 = sink64
		writtenSink = "d64"
	case ResultD128:
		lo, hi := bid128Bits(sink128)
		if lo == poisonSink128Lo && hi == poisonSink128Hi {
			return Observation{}, fmt.Errorf("row %s/%s did not write the Decimal128 sink", spec.Group, spec.Name)
		}
		observation.Bits128Lo = lo
		observation.Bits128Hi = hi
		writtenSink = "d128"
	case ResultI64:
		if sinkInt64 == poisonSinkInt64 {
			return Observation{}, fmt.Errorf("row %s/%s did not write the int64 sink", spec.Group, spec.Name)
		}
		observation.Int64 = sinkInt64
		writtenSink = "i64"
	case ResultPredicate:
		if sinkInt == poisonSinkInt {
			return Observation{}, fmt.Errorf("row %s/%s did not write the predicate sink", spec.Group, spec.Name)
		}
		observation.Predicate = sinkInt != 0
		writtenSink = "predicate"
	case ResultText:
		if sinkString == poisonSinkString {
			return Observation{}, fmt.Errorf("row %s/%s did not write the string sink", spec.Group, spec.Name)
		}
		observation.Text = sinkString
		writtenSink = "text"
	default:
		return Observation{}, fmt.Errorf("row %s/%s: unknown result kind %q", spec.Group, spec.Name, spec.Result)
	}

	if err := requireOtherSinksUntouched(spec, writtenSink); err != nil {
		return Observation{}, err
	}

	switch spec.Status {
	case StatusFlagsObserved:
		if sinkFlags == poisonSinkFlags {
			return Observation{}, fmt.Errorf("row %s/%s did not write the status-flag sink", spec.Group, spec.Name)
		}
		observation.Flags = sinkFlags
		observation.HasFlags = true
	case StatusValueOnly:
		if sinkFlags != poisonSinkFlags {
			return Observation{}, fmt.Errorf("row %s/%s is value-only but wrote status flags %#x", spec.Group, spec.Name, sinkFlags)
		}
	}
	return observation, nil
}

func requireOtherSinksUntouched(spec DescriptorRow, writtenSink string) error {
	if writtenSink != "d32" && sink32 != poisonSink32 {
		return fmt.Errorf("row %s/%s wrote the Decimal32 sink but declares result %q", spec.Group, spec.Name, spec.Result)
	}
	if writtenSink != "d64" && sink64 != poisonSink64 {
		return fmt.Errorf("row %s/%s wrote the Decimal64 sink but declares result %q", spec.Group, spec.Name, spec.Result)
	}
	if lo, hi := bid128Bits(sink128); writtenSink != "d128" && (lo != poisonSink128Lo || hi != poisonSink128Hi) {
		return fmt.Errorf("row %s/%s wrote the Decimal128 sink but declares result %q", spec.Group, spec.Name, spec.Result)
	}
	if writtenSink != "i64" && sinkInt64 != poisonSinkInt64 {
		return fmt.Errorf("row %s/%s wrote the int64 sink but declares result %q", spec.Group, spec.Name, spec.Result)
	}
	if writtenSink != "predicate" && sinkInt != poisonSinkInt {
		return fmt.Errorf("row %s/%s wrote the predicate sink but declares result %q", spec.Group, spec.Name, spec.Result)
	}
	if writtenSink != "text" && sinkString != poisonSinkString {
		return fmt.Errorf("row %s/%s wrote the string sink but declares result %q", spec.Group, spec.Name, spec.Result)
	}
	return nil
}
