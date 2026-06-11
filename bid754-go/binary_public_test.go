package bid754

import (
	"math"
	"testing"

	bidgo "github.com/sky1core/bid754/bid-go"
)

func TestDecimalBIDToBinaryExactValues(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("1")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(1): %v", err)
	}
	if got, flags := d32.ToBinary32(RoundNearestEven); math.Float32bits(got) != 0x3f800000 || flags != 0 {
		t.Fatalf("Decimal32BID.ToBinary32(1): got=%08x flags=%s", math.Float32bits(got), flags)
	}
	if got, flags := d32.ToBinary64(RoundNearestEven); math.Float64bits(got) != 0x3ff0000000000000 || flags != 0 {
		t.Fatalf("Decimal32BID.ToBinary64(1): got=%016x flags=%s", math.Float64bits(got), flags)
	}

	d64, err := NewDecimal64BIDDirect("-0")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(-0): %v", err)
	}
	if got, flags := d64.ToBinary32(RoundNearestEven); math.Float32bits(got) != 0x80000000 || flags != 0 {
		t.Fatalf("Decimal64BID.ToBinary32(-0): got=%08x flags=%s", math.Float32bits(got), flags)
	}
	if got, flags := d64.ToBinary64(RoundNearestEven); math.Float64bits(got) != 0x8000000000000000 || flags != 0 {
		t.Fatalf("Decimal64BID.ToBinary64(-0): got=%016x flags=%s", math.Float64bits(got), flags)
	}

	d128, err := NewDecimal128BIDDirect("1")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(1): %v", err)
	}
	if got, flags := d128.ToBinary32(RoundNearestEven); math.Float32bits(got) != 0x3f800000 || flags != 0 {
		t.Fatalf("Decimal128BID.ToBinary32(1): got=%08x flags=%s", math.Float32bits(got), flags)
	}
	if got, flags := d128.ToBinary64(RoundNearestEven); math.Float64bits(got) != 0x3ff0000000000000 || flags != 0 {
		t.Fatalf("Decimal128BID.ToBinary64(1): got=%016x flags=%s", math.Float64bits(got), flags)
	}
}

func TestDecimalBIDToBinaryRoundingAndFlags(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("0.1")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(0.1): %v", err)
	}
	nearest, nearestFlags := d32.ToBinary32(RoundNearestEven)
	zero, zeroFlags := d32.ToBinary32(RoundTowardZero)
	if nearestFlags == 0 || zeroFlags == 0 {
		t.Fatalf("expected inexact flags for binary32 conversion of 0.1, got nearest=%s zero=%s", nearestFlags, zeroFlags)
	}
	if math.Float32bits(nearest) == math.Float32bits(zero) {
		t.Fatalf("expected different binary32 rounding results for 0.1, got same bits=%08x", math.Float32bits(nearest))
	}
}

func TestDecimalBIDToDecimal128AndBinary128Ports(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("123.45")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(123.45): %v", err)
	}
	got128, gotFlags := d32.ToDecimal128()
	want128, wantFlags := bidgo.Bid32ToBid128(d32.ToUint32())
	if gotFlags != bidgoExceptionFlags(wantFlags) {
		t.Fatalf("Decimal32BID.ToDecimal128 flags=%s want=%02x", gotFlags, wantFlags)
	}
	if got128 != decimal128BIDFromBidgo(want128) {
		t.Fatalf("Decimal32BID.ToDecimal128 bits mismatch")
	}
	gotBin128, gotBinFlags := d32.ToBinary128(RoundNearestEven)
	wantBin128, wantBinFlags := bidgo.Bid32ToBinary128(d32.ToUint32(), bidgoRoundingMode(RoundNearestEven))
	if gotBinFlags != bidgoExceptionFlags(wantBinFlags) {
		t.Fatalf("Decimal32BID.ToBinary128 flags=%s want=%02x", gotBinFlags, wantBinFlags)
	}
	if gotBin128 != binary128FromBidgo(wantBin128) {
		t.Fatalf("Decimal32BID.ToBinary128 bits mismatch")
	}

	d64, err := NewDecimal64BIDDirect("-789.125")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(-789.125): %v", err)
	}
	got128d64, gotFlagsd64 := d64.ToDecimal128()
	want128d64, wantFlagsd64 := bidgo.Bid64ToBid128(d64.ToUint64())
	if gotFlagsd64 != bidgoExceptionFlags(wantFlagsd64) {
		t.Fatalf("Decimal64BID.ToDecimal128 flags=%s want=%02x", gotFlagsd64, wantFlagsd64)
	}
	if got128d64 != decimal128BIDFromBidgo(want128d64) {
		t.Fatalf("Decimal64BID.ToDecimal128 bits mismatch")
	}
	gotBin128d64, gotBinFlagsd64 := d64.ToBinary128(RoundNearestEven)
	wantBin128d64, wantBinFlagsd64 := bidgo.Bid64ToBinary128(d64.ToUint64(), bidgoRoundingMode(RoundNearestEven))
	if gotBinFlagsd64 != bidgoExceptionFlags(wantBinFlagsd64) {
		t.Fatalf("Decimal64BID.ToBinary128 flags=%s want=%02x", gotBinFlagsd64, wantBinFlagsd64)
	}
	if gotBin128d64 != binary128FromBidgo(wantBin128d64) {
		t.Fatalf("Decimal64BID.ToBinary128 bits mismatch")
	}

	d128, err := NewDecimal128BIDDirect("1.234567890123456789012345678901234e+100")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(decimal128 sample): %v", err)
	}
	gotBin128d128, gotBinFlagsd128 := d128.ToBinary128(RoundNearestEven)
	wantBin128d128, wantBinFlagsd128 := bidgo.Bid128ToBinary128(decimal128BIDAsBidgo(d128), bidgoRoundingMode(RoundNearestEven))
	if gotBinFlagsd128 != bidgoExceptionFlags(wantBinFlagsd128) {
		t.Fatalf("Decimal128BID.ToBinary128 flags=%s want=%02x", gotBinFlagsd128, wantBinFlagsd128)
	}
	if gotBin128d128 != binary128FromBidgo(wantBin128d128) {
		t.Fatalf("Decimal128BID.ToBinary128 bits mismatch")
	}
}

func TestDecimalBIDNextTowardPorts(t *testing.T) {
	target, err := NewDecimal128BIDDirect("1000000")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(1000000): %v", err)
	}

	d32, err := NewDecimal32BIDDirect("999999")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(999999): %v", err)
	}
	got32, got32Flags := d32.NextToward(target)
	want32, want32Flags := bidgo.Bid32NextToward(d32.ToUint32(), decimal128BIDAsBidgo(target))
	if got32Flags != bidgoExceptionFlags(want32Flags) {
		t.Fatalf("Decimal32BID.NextToward flags=%s want=%02x", got32Flags, want32Flags)
	}
	if got32.ToUint32() != want32 {
		t.Fatalf("Decimal32BID.NextToward bits=%08x want=%08x", got32.ToUint32(), want32)
	}

	d64, err := NewDecimal64BIDDirect("999999999999999")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(999999999999999): %v", err)
	}
	got64, got64Flags := d64.NextToward(target)
	want64, want64Flags := bidgo.Bid64NextToward(d64.ToUint64(), decimal128BIDAsBidgo(target))
	if got64Flags != bidgoExceptionFlags(want64Flags) {
		t.Fatalf("Decimal64BID.NextToward flags=%s want=%02x", got64Flags, want64Flags)
	}
	if got64.ToUint64() != want64 {
		t.Fatalf("Decimal64BID.NextToward bits=%016x want=%016x", got64.ToUint64(), want64)
	}
}

func TestDecimalBIDNextTowardMapsBIDStatusBits(t *testing.T) {
	target, err := NewDecimal128BIDDirect("10")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(10): %v", err)
	}
	d64, err := NewDecimal64BIDDirect("0")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(0): %v", err)
	}

	_, flags := d64.NextToward(target)
	want := FlagInexact | FlagUnderflow
	if flags != want {
		t.Fatalf("Decimal64BID.NextToward(0 -> 10) flags=%s want=%s", flags, want)
	}
}

func TestDecimalBIDNextPlusMinusPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("1")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(1): %v", err)
	}
	got32Plus, got32PlusFlags := d32.NextPlus()
	want32Plus, want32PlusFlags := bidgo.Bid32NextUp(d32.ToUint32())
	if got32PlusFlags != bidgoExceptionFlags(want32PlusFlags) || got32Plus.ToUint32() != want32Plus {
		t.Fatalf("Decimal32BID.NextPlus = %08x/%s want %08x/%02x", got32Plus.ToUint32(), got32PlusFlags, want32Plus, want32PlusFlags)
	}
	got32Minus, got32MinusFlags := d32.NextMinus()
	want32Minus, want32MinusFlags := bidgo.Bid32NextDown(d32.ToUint32())
	if got32MinusFlags != bidgoExceptionFlags(want32MinusFlags) || got32Minus.ToUint32() != want32Minus {
		t.Fatalf("Decimal32BID.NextMinus = %08x/%s want %08x/%02x", got32Minus.ToUint32(), got32MinusFlags, want32Minus, want32MinusFlags)
	}

	d64, err := NewDecimal64BIDDirect("-1")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(-1): %v", err)
	}
	got64Plus, got64PlusFlags := d64.NextPlus()
	want64Plus, want64PlusFlags := bidgo.Bid64NextUp(d64.ToUint64())
	if got64PlusFlags != bidgoExceptionFlags(want64PlusFlags) || got64Plus.ToUint64() != want64Plus {
		t.Fatalf("Decimal64BID.NextPlus = %016x/%s want %016x/%02x", got64Plus.ToUint64(), got64PlusFlags, want64Plus, want64PlusFlags)
	}
	got64Minus, got64MinusFlags := d64.NextMinus()
	want64Minus, want64MinusFlags := bidgo.Bid64NextDown(d64.ToUint64())
	if got64MinusFlags != bidgoExceptionFlags(want64MinusFlags) || got64Minus.ToUint64() != want64Minus {
		t.Fatalf("Decimal64BID.NextMinus = %016x/%s want %016x/%02x", got64Minus.ToUint64(), got64MinusFlags, want64Minus, want64MinusFlags)
	}

	d128, err := NewDecimal128BIDDirect("1")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(1): %v", err)
	}
	got128Plus, got128PlusFlags := d128.NextPlus()
	want128Plus, want128PlusFlags := bidgo.Bid128NextUp(decimal128BIDAsBidgo(d128))
	if got128PlusFlags != bidgoExceptionFlags(want128PlusFlags) || got128Plus != decimal128BIDFromBidgo(want128Plus) {
		t.Fatalf("Decimal128BID.NextPlus mismatch")
	}
	got128Minus, got128MinusFlags := d128.NextMinus()
	want128Minus, want128MinusFlags := bidgo.Bid128NextDown(decimal128BIDAsBidgo(d128))
	if got128MinusFlags != bidgoExceptionFlags(want128MinusFlags) || got128Minus != decimal128BIDFromBidgo(want128Minus) {
		t.Fatalf("Decimal128BID.NextMinus mismatch")
	}
}

func TestDecimalBIDFMAPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("2")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(2): %v", err)
	}
	mul32, err := NewDecimal32BIDDirect("3")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(3): %v", err)
	}
	add32, err := NewDecimal32BIDDirect("4")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(4): %v", err)
	}
	got32, got32Flags := d32.FMA(mul32, add32)
	want32, want32Flags := bidgo.Bid32Fma(d32.ToUint32(), mul32.ToUint32(), add32.ToUint32(), defaultBIDRoundingMode)
	if got32Flags != bidgoExceptionFlags(want32Flags) || got32.ToUint32() != want32 {
		t.Fatalf("Decimal32BID.FMA = %08x/%s want %08x/%02x", got32.ToUint32(), got32Flags, want32, want32Flags)
	}

	d64, err := NewDecimal64BIDDirect("2")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(2): %v", err)
	}
	mul64, err := NewDecimal64BIDDirect("3")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(3): %v", err)
	}
	add64, err := NewDecimal64BIDDirect("4")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(4): %v", err)
	}
	got64, got64Flags := d64.FMA(mul64, add64)
	want64, want64Flags := bidgo.Bid64Fma(d64.ToUint64(), mul64.ToUint64(), add64.ToUint64(), defaultBIDRoundingMode)
	if got64Flags != bidgoExceptionFlags(want64Flags) || got64.ToUint64() != want64 {
		t.Fatalf("Decimal64BID.FMA = %016x/%s want %016x/%02x", got64.ToUint64(), got64Flags, want64, want64Flags)
	}

	d128, err := NewDecimal128BIDDirect("2")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(2): %v", err)
	}
	mul128, err := NewDecimal128BIDDirect("3")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(3): %v", err)
	}
	add128, err := NewDecimal128BIDDirect("4")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(4): %v", err)
	}
	got128, got128Flags := d128.FMA(mul128, add128)
	want128, want128Flags := bidgo.Bid128Fma(decimal128BIDAsBidgo(d128), decimal128BIDAsBidgo(mul128), decimal128BIDAsBidgo(add128), defaultBIDRoundingMode)
	if got128Flags != bidgoExceptionFlags(want128Flags) || got128 != decimal128BIDFromBidgo(want128) {
		t.Fatalf("Decimal128BID.FMA mismatch")
	}
}

func TestDecimalBIDMinMaxPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("-2")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(-2): %v", err)
	}
	other32, err := NewDecimal32BIDDirect("1")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(1): %v", err)
	}
	got32, got32Flags := d32.MinNum(other32)
	want32, want32Flags := bidgo.Bid32MinNumWithFlags(d32.ToUint32(), other32.ToUint32())
	if got32Flags != bidgoExceptionFlags(want32Flags) || got32.ToUint32() != want32 {
		t.Fatalf("Decimal32BID.MinNum = %08x/%s want %08x/%02x", got32.ToUint32(), got32Flags, want32, want32Flags)
	}

	d64, err := NewDecimal64BIDDirect("-2")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(-2): %v", err)
	}
	other64, err := NewDecimal64BIDDirect("1")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(1): %v", err)
	}
	check64 := func(name string, got Decimal64BID, gotFlags ExceptionFlags, want uint64, wantFlags uint32) {
		t.Helper()
		if gotFlags != bidgoExceptionFlags(wantFlags) || got.ToUint64() != want {
			t.Fatalf("Decimal64BID.%s = %016x/%s want %016x/%02x", name, got.ToUint64(), gotFlags, want, wantFlags)
		}
	}
	got64Min, got64MinFlags := d64.MinNum(other64)
	want64Min, want64MinFlags := bidgo.Bid64MinNum(d64.ToUint64(), other64.ToUint64())
	check64("MinNum", got64Min, got64MinFlags, want64Min, want64MinFlags)
	got64Max, got64MaxFlags := d64.MaxNum(other64)
	want64Max, want64MaxFlags := bidgo.Bid64MaxNum(d64.ToUint64(), other64.ToUint64())
	check64("MaxNum", got64Max, got64MaxFlags, want64Max, want64MaxFlags)
	got64MinMag, got64MinMagFlags := d64.MinNumMag(other64)
	want64MinMag, want64MinMagFlags := bidgo.Bid64MinNumMag(d64.ToUint64(), other64.ToUint64())
	check64("MinNumMag", got64MinMag, got64MinMagFlags, want64MinMag, want64MinMagFlags)
	got64MaxMag, got64MaxMagFlags := d64.MaxNumMag(other64)
	want64MaxMag, want64MaxMagFlags := bidgo.Bid64MaxNumMag(d64.ToUint64(), other64.ToUint64())
	check64("MaxNumMag", got64MaxMag, got64MaxMagFlags, want64MaxMag, want64MaxMagFlags)

	d128, err := NewDecimal128BIDDirect("-2")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(-2): %v", err)
	}
	other128, err := NewDecimal128BIDDirect("1")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(1): %v", err)
	}
	var want128Flags uint32
	want128 := bidgo.Bid128Minnum(decimal128BIDAsBidgo(d128), decimal128BIDAsBidgo(other128), &want128Flags)
	got128, got128Flags := d128.MinNum(other128)
	if got128Flags != bidgoExceptionFlags(want128Flags) || got128 != decimal128BIDFromBidgo(want128) {
		t.Fatalf("Decimal128BID.MinNum mismatch")
	}
}

func TestDecimalBIDCompareTotalPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("-2")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(-2): %v", err)
	}
	other32, err := NewDecimal32BIDDirect("1")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(1): %v", err)
	}
	if got, want := d32.CompareTotal(other32), totalOrderComparison(bidgo.Bid32TotalOrder(d32.ToUint32(), other32.ToUint32()), bidgo.Bid32TotalOrder(other32.ToUint32(), d32.ToUint32())); got != want {
		t.Fatalf("Decimal32BID.CompareTotal=%d want %d", got, want)
	}
	if got, want := d32.CompareTotalMag(other32), totalOrderComparison(bidgo.Bid32TotalOrderMag(d32.ToUint32(), other32.ToUint32()), bidgo.Bid32TotalOrderMag(other32.ToUint32(), d32.ToUint32())); got != want {
		t.Fatalf("Decimal32BID.CompareTotalMag=%d want %d", got, want)
	}

	d64, err := NewDecimal64BIDDirect("7.0")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(7.0): %v", err)
	}
	other64, err := NewDecimal64BIDDirect("7")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(7): %v", err)
	}
	if got, want := d64.CompareTotal(other64), totalOrderComparison(bidgo.Bid64TotalOrder(d64.ToUint64(), other64.ToUint64()), bidgo.Bid64TotalOrder(other64.ToUint64(), d64.ToUint64())); got != want {
		t.Fatalf("Decimal64BID.CompareTotal=%d want %d", got, want)
	}
	if got, want := d64.CompareTotalMag(other64), totalOrderComparison(bidgo.Bid64TotalOrderMag(d64.ToUint64(), other64.ToUint64()), bidgo.Bid64TotalOrderMag(other64.ToUint64(), d64.ToUint64())); got != want {
		t.Fatalf("Decimal64BID.CompareTotalMag=%d want %d", got, want)
	}

	d128, err := NewDecimal128BIDDirect("-NaN41")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(-NaN41): %v", err)
	}
	other128, err := NewDecimal128BIDDirect("+NaN42")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(+NaN42): %v", err)
	}
	if got, want := d128.CompareTotal(other128), totalOrderComparison(bidgo.Bid128TotalOrder(decimal128BIDAsBidgo(d128), decimal128BIDAsBidgo(other128)), bidgo.Bid128TotalOrder(decimal128BIDAsBidgo(other128), decimal128BIDAsBidgo(d128))); got != want {
		t.Fatalf("Decimal128BID.CompareTotal=%d want %d", got, want)
	}
	if got, want := d128.CompareTotalMag(other128), totalOrderComparison(bidgo.Bid128TotalOrderMag(decimal128BIDAsBidgo(d128), decimal128BIDAsBidgo(other128)), bidgo.Bid128TotalOrderMag(decimal128BIDAsBidgo(other128), decimal128BIDAsBidgo(d128))); got != want {
		t.Fatalf("Decimal128BID.CompareTotalMag=%d want %d", got, want)
	}
}

func TestDecimalBIDLogBPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("100")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(100): %v", err)
	}
	got32, got32Flags := d32.LogB()
	want32, want32Flags := bidgo.Bid32Logb(d32.ToUint32())
	if got32Flags != bidgoExceptionFlags(want32Flags) || got32.ToUint32() != want32 {
		t.Fatalf("Decimal32BID.LogB = %08x/%s want %08x/%02x", got32.ToUint32(), got32Flags, want32, want32Flags)
	}

	d64, err := NewDecimal64BIDDirect("100")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(100): %v", err)
	}
	got64, got64Flags := d64.LogB()
	want64, want64Flags := bidgo.Bid64Logb(d64.ToUint64())
	if got64Flags != bidgoExceptionFlags(want64Flags) || got64.ToUint64() != want64 {
		t.Fatalf("Decimal64BID.LogB = %016x/%s want %016x/%02x", got64.ToUint64(), got64Flags, want64, want64Flags)
	}

	d128, err := NewDecimal128BIDDirect("100")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(100): %v", err)
	}
	var want128Flags uint32
	want128 := bidgo.Bid128Logb(decimal128BIDAsBidgo(d128), &want128Flags)
	got128, got128Flags := d128.LogB()
	if got128Flags != bidgoExceptionFlags(want128Flags) || got128 != decimal128BIDFromBidgo(want128) {
		t.Fatalf("Decimal128BID.LogB mismatch")
	}
}

func TestDecimalBIDScaleBPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("7.50")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(7.50): %v", err)
	}
	got32, got32Flags := d32.ScaleB(2)
	want32, want32Flags := bidgo.Bid32Scalbn(d32.ToUint32(), 2, defaultBIDRoundingMode)
	if got32Flags != bidgoExceptionFlags(want32Flags) || got32.ToUint32() != want32 {
		t.Fatalf("Decimal32BID.ScaleB = %08x/%s want %08x/%02x", got32.ToUint32(), got32Flags, want32, want32Flags)
	}

	d64, err := NewDecimal64BIDDirect("7.50")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(7.50): %v", err)
	}
	got64, got64Flags := d64.ScaleB(2)
	want64, want64Flags := bidgo.Bid64Scalbn(d64.ToUint64(), 2, defaultBIDRoundingMode)
	if got64Flags != bidgoExceptionFlags(want64Flags) || got64.ToUint64() != want64 {
		t.Fatalf("Decimal64BID.ScaleB = %016x/%s want %016x/%02x", got64.ToUint64(), got64Flags, want64, want64Flags)
	}

	d128, err := NewDecimal128BIDDirect("7.50")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(7.50): %v", err)
	}
	var want128Flags uint32
	want128 := bidgo.Bid128Scalbn(decimal128BIDAsBidgo(d128), 2, defaultBIDRoundingMode, &want128Flags)
	got128, got128Flags := d128.ScaleB(2)
	if got128Flags != bidgoExceptionFlags(want128Flags) || got128 != decimal128BIDFromBidgo(want128) {
		t.Fatalf("Decimal128BID.ScaleB mismatch")
	}
}

func TestDecimalBIDRemainderPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("2")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(2): %v", err)
	}
	other32, err := NewDecimal32BIDDirect("3")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(3): %v", err)
	}
	got32, got32Flags := d32.Remainder(other32)
	want32, want32Flags := bidgo.Bid32Rem(d32.ToUint32(), other32.ToUint32())
	if got32Flags != bidgoExceptionFlags(want32Flags) || got32.ToUint32() != want32 {
		t.Fatalf("Decimal32BID.Remainder = %08x/%s want %08x/%02x", got32.ToUint32(), got32Flags, want32, want32Flags)
	}

	d64, err := NewDecimal64BIDDirect("2")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(2): %v", err)
	}
	other64, err := NewDecimal64BIDDirect("3")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(3): %v", err)
	}
	got64, got64Flags := d64.Remainder(other64)
	want64, want64Flags := bidgo.Bid64Rem(d64.ToUint64(), other64.ToUint64())
	if got64Flags != bidgoExceptionFlags(want64Flags) || got64.ToUint64() != want64 {
		t.Fatalf("Decimal64BID.Remainder = %016x/%s want %016x/%02x", got64.ToUint64(), got64Flags, want64, want64Flags)
	}

	d128, err := NewDecimal128BIDDirect("2")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(2): %v", err)
	}
	other128, err := NewDecimal128BIDDirect("3")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(3): %v", err)
	}
	got128, got128Flags := d128.Remainder(other128)
	want128, want128Flags := bidgo.Bid128Rem(decimal128BIDAsBidgo(d128), decimal128BIDAsBidgo(other128))
	if got128Flags != bidgoExceptionFlags(want128Flags) || got128 != decimal128BIDFromBidgo(want128) {
		t.Fatalf("Decimal128BID.Remainder mismatch")
	}
}

func TestDecimalBIDFmodPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("2")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(2): %v", err)
	}
	other32, err := NewDecimal32BIDDirect("3")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(3): %v", err)
	}
	got32, got32Flags := d32.Fmod(other32)
	want32, want32Flags := bidgo.Bid32Fmod(d32.ToUint32(), other32.ToUint32())
	if got32Flags != bidgoExceptionFlags(want32Flags) || got32.ToUint32() != want32 {
		t.Fatalf("Decimal32BID.Fmod = %08x/%s want %08x/%02x", got32.ToUint32(), got32Flags, want32, want32Flags)
	}

	d64, err := NewDecimal64BIDDirect("2")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(2): %v", err)
	}
	other64, err := NewDecimal64BIDDirect("3")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(3): %v", err)
	}
	got64, got64Flags := d64.Fmod(other64)
	want64, want64Flags := bidgo.Bid64Fmod(d64.ToUint64(), other64.ToUint64())
	if got64Flags != bidgoExceptionFlags(want64Flags) || got64.ToUint64() != want64 {
		t.Fatalf("Decimal64BID.Fmod = %016x/%s want %016x/%02x", got64.ToUint64(), got64Flags, want64, want64Flags)
	}

	d128, err := NewDecimal128BIDDirect("2")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(2): %v", err)
	}
	other128, err := NewDecimal128BIDDirect("3")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(3): %v", err)
	}
	got128, got128Flags := d128.Fmod(other128)
	want128, want128Flags := bidgo.Bid128Fmod(decimal128BIDAsBidgo(d128), decimal128BIDAsBidgo(other128))
	if got128Flags != bidgoExceptionFlags(want128Flags) || got128 != decimal128BIDFromBidgo(want128) {
		t.Fatalf("Decimal128BID.Fmod mismatch")
	}
}

func TestDecimalBIDSqrtPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("4")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(4): %v", err)
	}
	got32, got32Flags := d32.Sqrt()
	want32, want32Flags := bidgo.Bid32Sqrt(d32.ToUint32(), defaultBIDRoundingMode)
	if got32Flags != bidgoExceptionFlags(want32Flags) || got32.ToUint32() != want32 {
		t.Fatalf("Decimal32BID.Sqrt = %08x/%s want %08x/%02x", got32.ToUint32(), got32Flags, want32, want32Flags)
	}

	d64, err := NewDecimal64BIDDirect("4")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(4): %v", err)
	}
	got64, got64Flags := d64.Sqrt()
	want64, want64Flags := bidgo.Bid64Sqrt(d64.ToUint64(), defaultBIDRoundingMode)
	if got64Flags != bidgoExceptionFlags(want64Flags) || got64.ToUint64() != want64 {
		t.Fatalf("Decimal64BID.Sqrt = %016x/%s want %016x/%02x", got64.ToUint64(), got64Flags, want64, want64Flags)
	}

	d128, err := NewDecimal128BIDDirect("4")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(4): %v", err)
	}
	got128, got128Flags := d128.Sqrt()
	want128, want128Flags := bidgo.Bid128Sqrt(decimal128BIDAsBidgo(d128), defaultBIDRoundingMode)
	if got128Flags != bidgoExceptionFlags(want128Flags) || got128 != decimal128BIDFromBidgo(want128) {
		t.Fatalf("Decimal128BID.Sqrt mismatch")
	}

	negative64, err := NewDecimal64BIDDirect("-1")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(-1): %v", err)
	}
	_, flags := negative64.Sqrt()
	if flags != FlagInvalidOperation {
		t.Fatalf("Decimal64BID.Sqrt(-1) flags = %s, want %s", flags, FlagInvalidOperation)
	}
}

func TestDecimal64BIDReducePort(t *testing.T) {
	d, err := NewDecimal64BIDDirect("120.00")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(120.00): %v", err)
	}
	got, gotFlags := d.Reduce()
	want, wantFlags := bidgo.Bid64Reduce(d.ToUint64())
	if gotFlags != bidgoExceptionFlags(wantFlags) {
		t.Fatalf("Decimal64BID.Reduce flags=%s want=%02x", gotFlags, wantFlags)
	}
	if got.ToUint64() != want {
		t.Fatalf("Decimal64BID.Reduce bits=%016x want=%016x", got.ToUint64(), want)
	}
	if !compareDecimalResults(got.String(), "1.2E+2") {
		t.Fatalf("Decimal64BID.Reduce string=%s want value 1.2E+2", got.String())
	}
}

func TestDecimalBIDWidthConversionPorts(t *testing.T) {
	d32, err := NewDecimal32BIDDirect("123.45")
	if err != nil {
		t.Fatalf("NewDecimal32BIDDirect(123.45): %v", err)
	}
	got64, got64Flags := d32.ToDecimal64()
	want64, want64Flags := bidgo.Bid32ToBid64(d32.ToUint32())
	if got64Flags != bidgoExceptionFlags(want64Flags) || got64.ToUint64() != want64 {
		t.Fatalf("Decimal32BID.ToDecimal64 = %016x/%s want %016x/%02x", got64.ToUint64(), got64Flags, want64, want64Flags)
	}

	d64, err := NewDecimal64BIDDirect("1234567890123456")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect(1234567890123456): %v", err)
	}
	got32, got32Flags := d64.ToDecimal32(RoundNearestEven)
	want32, want32Flags := bidgo.Bid64ToBid32(d64.ToUint64(), bidgoRoundingMode(RoundNearestEven))
	if got32Flags != bidgoExceptionFlags(want32Flags) || got32.ToUint32() != want32 {
		t.Fatalf("Decimal64BID.ToDecimal32 = %08x/%s want %08x/%02x", got32.ToUint32(), got32Flags, want32, want32Flags)
	}
	if got32Flags&FlagInexact == 0 {
		t.Fatalf("Decimal64BID.ToDecimal32(1234567890123456) expected inexact rounding, flags=%s", got32Flags)
	}

	d128, err := NewDecimal128BIDDirect("1234567890123456789012345678901234")
	if err != nil {
		t.Fatalf("NewDecimal128BIDDirect(34-digit): %v", err)
	}
	got64n, got64nFlags := d128.ToDecimal64(RoundNearestEven)
	want64n, want64nFlags := bidgo.Bid128ToBid64(decimal128BIDAsBidgo(d128), bidgoRoundingMode(RoundNearestEven))
	if got64nFlags != bidgoExceptionFlags(want64nFlags) || got64n.ToUint64() != want64n {
		t.Fatalf("Decimal128BID.ToDecimal64 = %016x/%s want %016x/%02x", got64n.ToUint64(), got64nFlags, want64n, want64nFlags)
	}
	if got64nFlags&FlagInexact == 0 {
		t.Fatalf("Decimal128BID.ToDecimal64(34-digit) expected inexact rounding, flags=%s", got64nFlags)
	}

	got32n, got32nFlags := d128.ToDecimal32(RoundNearestEven)
	want32n, want32nFlags := bidgo.Bid128ToBid32(decimal128BIDAsBidgo(d128), bidgoRoundingMode(RoundNearestEven))
	if got32nFlags != bidgoExceptionFlags(want32nFlags) || got32n.ToUint32() != want32n {
		t.Fatalf("Decimal128BID.ToDecimal32 = %08x/%s want %08x/%02x", got32n.ToUint32(), got32nFlags, want32n, want32nFlags)
	}
}

func TestDecimalBIDNarrowingConversionHonorsRoundingMode(t *testing.T) {
	d64, err := NewDecimal64BIDDirect("1234567890123456")
	if err != nil {
		t.Fatalf("NewDecimal64BIDDirect: %v", err)
	}
	nearest, _ := d64.ToDecimal32(RoundNearestEven)
	toward, _ := d64.ToDecimal32(RoundTowardZero)
	if nearest.ToUint32() == toward.ToUint32() {
		t.Fatalf("expected RoundNearestEven and RoundTowardZero to diverge, both = %08x", nearest.ToUint32())
	}
	wantToward, _ := bidgo.Bid64ToBid32(d64.ToUint64(), bidgoRoundingMode(RoundTowardZero))
	if toward.ToUint32() != wantToward {
		t.Fatalf("ToDecimal32(RoundTowardZero) = %08x want %08x", toward.ToUint32(), wantToward)
	}
}

func TestDecimalBIDSignalingEqualVectors(t *testing.T) {
	testCases := []struct {
		name      string
		a, b      string
		wantEq    bool
		wantFlags ExceptionFlags
	}{
		{name: "equal finite", a: "2", b: "2.0", wantEq: true, wantFlags: 0},
		{name: "unequal finite", a: "2", b: "3", wantEq: false, wantFlags: 0},
		{name: "equal infinities", a: "Inf", b: "Inf", wantEq: true, wantFlags: 0},
		{name: "opposite infinities", a: "Inf", b: "-Inf", wantEq: false, wantFlags: 0},
		{name: "quiet nan raises invalid", a: "NaN", b: "2", wantEq: false, wantFlags: FlagInvalidOperation},
		{name: "signaling nan raises invalid", a: "sNaN", b: "2", wantEq: false, wantFlags: FlagInvalidOperation},
		{name: "both quiet nan", a: "NaN", b: "NaN", wantEq: false, wantFlags: FlagInvalidOperation},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			a32, err := NewDecimal32BIDDirect(tc.a)
			if err != nil {
				t.Fatalf("NewDecimal32BIDDirect(%s): %v", tc.a, err)
			}
			b32, err := NewDecimal32BIDDirect(tc.b)
			if err != nil {
				t.Fatalf("NewDecimal32BIDDirect(%s): %v", tc.b, err)
			}
			a64, err := NewDecimal64BIDDirect(tc.a)
			if err != nil {
				t.Fatalf("NewDecimal64BIDDirect(%s): %v", tc.a, err)
			}
			b64, err := NewDecimal64BIDDirect(tc.b)
			if err != nil {
				t.Fatalf("NewDecimal64BIDDirect(%s): %v", tc.b, err)
			}
			a128, err := NewDecimal128BIDDirect(tc.a)
			if err != nil {
				t.Fatalf("NewDecimal128BIDDirect(%s): %v", tc.a, err)
			}
			b128, err := NewDecimal128BIDDirect(tc.b)
			if err != nil {
				t.Fatalf("NewDecimal128BIDDirect(%s): %v", tc.b, err)
			}
			check := func(width string, eq bool, eqFlags ExceptionFlags, ne bool, neFlags ExceptionFlags) {
				if eq != tc.wantEq || eqFlags != tc.wantFlags {
					t.Fatalf("%s SignalingEqual(%s,%s) = %v/%s want %v/%s", width, tc.a, tc.b, eq, eqFlags, tc.wantEq, tc.wantFlags)
				}
				if ne != !tc.wantEq || neFlags != tc.wantFlags {
					t.Fatalf("%s SignalingNotEqual(%s,%s) = %v/%s want %v/%s", width, tc.a, tc.b, ne, neFlags, !tc.wantEq, tc.wantFlags)
				}
			}
			eq32, eqFlags32 := a32.SignalingEqual(b32)
			ne32, neFlags32 := a32.SignalingNotEqual(b32)
			check("Decimal32BID", eq32, eqFlags32, ne32, neFlags32)
			eq64, eqFlags64 := a64.SignalingEqual(b64)
			ne64, neFlags64 := a64.SignalingNotEqual(b64)
			check("Decimal64BID", eq64, eqFlags64, ne64, neFlags64)
			eq128, eqFlags128 := a128.SignalingEqual(b128)
			ne128, neFlags128 := a128.SignalingNotEqual(b128)
			check("Decimal128BID", eq128, eqFlags128, ne128, neFlags128)
		})
	}
}
