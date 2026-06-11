package bid754

import (
	"testing"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

type predicateCase32 struct {
	name  string
	value Decimal32BID
}

type predicateCase64 struct {
	name  string
	value Decimal64BID
}

type predicateCase128 struct {
	name  string
	value Decimal128BID
}

func predicateInputs32(t *testing.T) []predicateCase32 {
	t.Helper()
	return []predicateCase32{
		{name: "normal", value: mustDecimal32BID(t, "1.5")},
		{name: "zero", value: mustDecimal32BID(t, "0")},
		{name: "inf", value: mustDecimal32BID(t, "Infinity")},
		{name: "qnan", value: mustDecimal32BID(t, "NaN")},
		{name: "snan", value: mustDecimal32BID(t, "sNaN")},
		{name: "subnormal", value: mustDecimal32BID(t, "1E-101")},
		{name: "noncanonical", value: Decimal32BID(0x7c100000)},
	}
}

func predicateInputs64(t *testing.T) []predicateCase64 {
	t.Helper()
	return []predicateCase64{
		{name: "normal", value: mustDecimal64BID(t, "1.5")},
		{name: "zero", value: mustDecimal64BID(t, "0")},
		{name: "inf", value: mustDecimal64BID(t, "Infinity")},
		{name: "qnan", value: mustDecimal64BID(t, "NaN")},
		{name: "snan", value: mustDecimal64BID(t, "sNaN")},
		{name: "subnormal", value: mustDecimal64BID(t, "1E-398")},
		{name: "noncanonical", value: Decimal64BID(0x7c04000000000000)},
	}
}

func predicateInputs128(t *testing.T) []predicateCase128 {
	t.Helper()
	return []predicateCase128{
		{name: "normal", value: mustDecimal128BID(t, "1.5")},
		{name: "zero", value: mustDecimal128BID(t, "0")},
		{name: "inf", value: mustDecimal128BID(t, "Infinity")},
		{name: "qnan", value: mustDecimal128BID(t, "NaN")},
		{name: "snan", value: mustDecimal128BID(t, "sNaN")},
		{name: "subnormal", value: mustDecimal128BID(t, "1E-6176")},
		{name: "noncanonical", value: rawDecimal128BID(0, 0x7c00400000000000)},
	}
}

func TestPredicatePublicRoutingDecimal32(t *testing.T) {
	tests := []struct {
		name   string
		public func(Decimal32BID) bool
		direct func(uint32) int
	}{
		{name: "IsNormal", public: Decimal32BID.IsNormal, direct: bidgo.Bid32IsNormal},
		{name: "IsFinite", public: Decimal32BID.IsFinite, direct: bidgo.Bid32IsFinite},
		{name: "IsSubnormal", public: Decimal32BID.IsSubnormal, direct: bidgo.Bid32IsSubnormal},
		{name: "IsSignaling", public: Decimal32BID.IsSignaling, direct: bidgo.Bid32IsSignaling},
		{name: "IsCanonical", public: Decimal32BID.IsCanonical, direct: bidgo.Bid32IsCanonical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range predicateInputs32(t) {
				got := tt.public(input.value)
				want := tt.direct(input.value.ToUint32()) != 0
				if got != want {
					t.Fatalf("%s(%s) = %v, want %v", tt.name, input.name, got, want)
				}
			}
		})
	}
	if got := mustDecimal32BID(t, "1").Radix(); got != bidgo.Bid32Radix() {
		t.Fatalf("Decimal32BID.Radix() = %d, want %d", got, bidgo.Bid32Radix())
	}
	inputs := predicateInputs32(t)
	if bidgo.Bid32IsSubnormal(inputs[5].value.ToUint32()) == 0 || bidgo.Bid32IsCanonical(inputs[6].value.ToUint32()) != 0 {
		t.Fatalf("decimal32 subnormal/canonical premise failed")
	}
}

func TestPredicatePublicRoutingDecimal64(t *testing.T) {
	tests := []struct {
		name   string
		public func(Decimal64BID) bool
		direct func(uint64) int
	}{
		{name: "IsNormal", public: Decimal64BID.IsNormal, direct: bidgo.Bid64IsNormal},
		{name: "IsFinite", public: Decimal64BID.IsFinite, direct: bidgo.Bid64IsFinite},
		{name: "IsSubnormal", public: Decimal64BID.IsSubnormal, direct: bidgo.Bid64IsSubnormal},
		{name: "IsSignaling", public: Decimal64BID.IsSignaling, direct: bidgo.Bid64IsSignaling},
		{name: "IsCanonical", public: Decimal64BID.IsCanonical, direct: bidgo.Bid64IsCanonical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range predicateInputs64(t) {
				got := tt.public(input.value)
				want := tt.direct(input.value.ToUint64()) != 0
				if got != want {
					t.Fatalf("%s(%s) = %v, want %v", tt.name, input.name, got, want)
				}
			}
		})
	}
	if got := mustDecimal64BID(t, "1").Radix(); got != bidgo.Bid64Radix() {
		t.Fatalf("Decimal64BID.Radix() = %d, want %d", got, bidgo.Bid64Radix())
	}
	inputs := predicateInputs64(t)
	if bidgo.Bid64IsSubnormal(inputs[5].value.ToUint64()) == 0 || bidgo.Bid64IsCanonical(inputs[6].value.ToUint64()) != 0 {
		t.Fatalf("decimal64 subnormal/canonical premise failed")
	}
}

func TestPredicatePublicRoutingDecimal128(t *testing.T) {
	tests := []struct {
		name   string
		public func(Decimal128BID) bool
		direct func(bidgo.BID_UINT128) int
	}{
		{name: "IsNormal", public: Decimal128BID.IsNormal, direct: bidgo.Bid128IsNormal},
		{name: "IsFinite", public: Decimal128BID.IsFinite, direct: bidgo.Bid128IsFinite},
		{name: "IsSubnormal", public: Decimal128BID.IsSubnormal, direct: bidgo.Bid128IsSubnormal},
		{name: "IsSignaling", public: Decimal128BID.IsSignaling, direct: bidgo.Bid128IsSignaling},
		{name: "IsCanonical", public: Decimal128BID.IsCanonical, direct: bidgo.Bid128IsCanonical},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range predicateInputs128(t) {
				got := tt.public(input.value)
				want := tt.direct(decimal128BIDAsBidgo(input.value)) != 0
				if got != want {
					t.Fatalf("%s(%s) = %v, want %v", tt.name, input.name, got, want)
				}
			}
		})
	}
	if got := mustDecimal128BID(t, "1").Radix(); got != bidgo.Bid128Radix() {
		t.Fatalf("Decimal128BID.Radix() = %d, want %d", got, bidgo.Bid128Radix())
	}
	inputs := predicateInputs128(t)
	if bidgo.Bid128IsSubnormal(decimal128BIDAsBidgo(inputs[5].value)) == 0 || bidgo.Bid128IsCanonical(decimal128BIDAsBidgo(inputs[6].value)) != 0 {
		t.Fatalf("decimal128 subnormal/canonical premise failed")
	}
}

func TestRoundIntegralDirectedRoutingDecimal32(t *testing.T) {
	tests := []struct {
		name   string
		public func(Decimal32BID) (Decimal32BID, ExceptionFlags)
		direct func(uint32) (uint32, uint32)
	}{
		{name: "NearestEven", public: Decimal32BID.RoundIntegralNearestEven, direct: bidgo.Bid32RoundIntegralNearestEven},
		{name: "NearestAway", public: Decimal32BID.RoundIntegralNearestAway, direct: bidgo.Bid32RoundIntegralNearestAway},
		{name: "Zero", public: Decimal32BID.RoundIntegralZero, direct: bidgo.Bid32RoundIntegralZero},
		{name: "Positive", public: Decimal32BID.RoundIntegralPositive, direct: bidgo.Bid32RoundIntegralPositive},
		{name: "Negative", public: Decimal32BID.RoundIntegralNegative, direct: bidgo.Bid32RoundIntegralNegative},
	}
	inputs := []struct {
		name  string
		value Decimal32BID
	}{
		{name: "positive half", value: mustDecimal32BID(t, "1.5")},
		{name: "negative frac", value: mustDecimal32BID(t, "-2.75")},
		{name: "zero", value: mustDecimal32BID(t, "0")},
		{name: "qnan", value: mustDecimal32BID(t, "NaN")},
		{name: "snan", value: mustDecimal32BID(t, "sNaN")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range inputs {
				got, gotFlags := tt.public(input.value)
				want, wantFlags := tt.direct(input.value.ToUint32())
				if got != Decimal32BID(want) || gotFlags != bidgoExceptionFlags(wantFlags) {
					t.Fatalf("%s(%s) = %08x/%s, want %08x/%s", tt.name, input.name, got.ToUint32(), gotFlags, want, bidgoExceptionFlags(wantFlags))
				}
			}
		})
	}
}

func TestRoundIntegralDirectedRoutingDecimal64(t *testing.T) {
	tests := []struct {
		name   string
		public func(Decimal64BID) (Decimal64BID, ExceptionFlags)
		direct func(uint64) (uint64, uint32)
	}{
		{name: "NearestEven", public: Decimal64BID.RoundIntegralNearestEven, direct: bidgo.Bid64RoundIntegralNearestEven},
		{name: "NearestAway", public: Decimal64BID.RoundIntegralNearestAway, direct: bidgo.Bid64RoundIntegralNearestAway},
		{name: "Zero", public: Decimal64BID.RoundIntegralZero, direct: bidgo.Bid64RoundIntegralZero},
		{name: "Positive", public: Decimal64BID.RoundIntegralPositive, direct: bidgo.Bid64RoundIntegralPositive},
		{name: "Negative", public: Decimal64BID.RoundIntegralNegative, direct: bidgo.Bid64RoundIntegralNegative},
	}
	inputs := []struct {
		name  string
		value Decimal64BID
	}{
		{name: "positive half", value: mustDecimal64BID(t, "1.5")},
		{name: "negative frac", value: mustDecimal64BID(t, "-2.75")},
		{name: "zero", value: mustDecimal64BID(t, "0")},
		{name: "qnan", value: mustDecimal64BID(t, "NaN")},
		{name: "snan", value: mustDecimal64BID(t, "sNaN")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range inputs {
				got, gotFlags := tt.public(input.value)
				want, wantFlags := tt.direct(input.value.ToUint64())
				if got != Decimal64BID(want) || gotFlags != bidgoExceptionFlags(wantFlags) {
					t.Fatalf("%s(%s) = %016x/%s, want %016x/%s", tt.name, input.name, got.ToUint64(), gotFlags, want, bidgoExceptionFlags(wantFlags))
				}
			}
		})
	}
}

func TestRoundIntegralDirectedRoutingDecimal128(t *testing.T) {
	tests := []struct {
		name   string
		public func(Decimal128BID) (Decimal128BID, ExceptionFlags)
		direct func(bidgo.BID_UINT128, *uint32) bidgo.BID_UINT128
	}{
		{name: "NearestEven", public: Decimal128BID.RoundIntegralNearestEven, direct: bidgo.Bid128RoundIntegralNearestEven},
		{name: "NearestAway", public: Decimal128BID.RoundIntegralNearestAway, direct: bidgo.Bid128RoundIntegralNearestAway},
		{name: "Zero", public: Decimal128BID.RoundIntegralZero, direct: bidgo.Bid128RoundIntegralZero},
		{name: "Positive", public: Decimal128BID.RoundIntegralPositive, direct: bidgo.Bid128RoundIntegralPositive},
		{name: "Negative", public: Decimal128BID.RoundIntegralNegative, direct: bidgo.Bid128RoundIntegralNegative},
	}
	inputs := []struct {
		name  string
		value Decimal128BID
	}{
		{name: "positive half", value: mustDecimal128BID(t, "1.5")},
		{name: "negative frac", value: mustDecimal128BID(t, "-2.75")},
		{name: "zero", value: mustDecimal128BID(t, "0")},
		{name: "qnan", value: mustDecimal128BID(t, "NaN")},
		{name: "snan", value: mustDecimal128BID(t, "sNaN")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, input := range inputs {
				got, gotFlags := tt.public(input.value)
				var wantFlags uint32
				want := tt.direct(decimal128BIDAsBidgo(input.value), &wantFlags)
				wantDecimal := decimal128BIDFromBidgo(want)
				if got != wantDecimal || gotFlags != bidgoExceptionFlags(wantFlags) {
					t.Fatalf("%s(%s) = %s/%s, want %s/%s", tt.name, input.name, got.String(), gotFlags, wantDecimal.String(), bidgoExceptionFlags(wantFlags))
				}
			}
		})
	}
}

func TestRoundIntegralDirectedFlags(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		frac := mustDecimal32BID(t, "1.5")
		_, exactFlags := frac.RoundIntegralExactWithFlags()
		if !exactFlags.HasFlag(FlagInexact) {
			t.Fatalf("Decimal32 RoundIntegralExactWithFlags(1.5) flags = %s, want Inexact", exactFlags)
		}
		for _, call := range []func() (Decimal32BID, ExceptionFlags){frac.RoundIntegralNearestEven, frac.RoundIntegralNearestAway, frac.RoundIntegralZero, frac.RoundIntegralPositive, frac.RoundIntegralNegative} {
			_, flags := call()
			if flags.HasFlag(FlagInexact) {
				t.Fatalf("Decimal32 directed RoundIntegral(1.5) flags = %s, want no Inexact", flags)
			}
		}
		snan := mustDecimal32BID(t, "sNaN")
		for _, call := range []func() (Decimal32BID, ExceptionFlags){snan.RoundIntegralNearestEven, snan.RoundIntegralNearestAway, snan.RoundIntegralZero, snan.RoundIntegralPositive, snan.RoundIntegralNegative} {
			_, flags := call()
			if flags != FlagInvalidOperation {
				t.Fatalf("Decimal32 directed RoundIntegral(sNaN) flags = %s, want %s", flags, FlagInvalidOperation)
			}
		}
	})
	t.Run("decimal64", func(t *testing.T) {
		frac := mustDecimal64BID(t, "1.5")
		_, exactFlags := frac.RoundIntegralExactWithFlags()
		if !exactFlags.HasFlag(FlagInexact) {
			t.Fatalf("Decimal64 RoundIntegralExactWithFlags(1.5) flags = %s, want Inexact", exactFlags)
		}
		for _, call := range []func() (Decimal64BID, ExceptionFlags){frac.RoundIntegralNearestEven, frac.RoundIntegralNearestAway, frac.RoundIntegralZero, frac.RoundIntegralPositive, frac.RoundIntegralNegative} {
			_, flags := call()
			if flags.HasFlag(FlagInexact) {
				t.Fatalf("Decimal64 directed RoundIntegral(1.5) flags = %s, want no Inexact", flags)
			}
		}
		snan := mustDecimal64BID(t, "sNaN")
		for _, call := range []func() (Decimal64BID, ExceptionFlags){snan.RoundIntegralNearestEven, snan.RoundIntegralNearestAway, snan.RoundIntegralZero, snan.RoundIntegralPositive, snan.RoundIntegralNegative} {
			_, flags := call()
			if flags != FlagInvalidOperation {
				t.Fatalf("Decimal64 directed RoundIntegral(sNaN) flags = %s, want %s", flags, FlagInvalidOperation)
			}
		}
	})
	t.Run("decimal128", func(t *testing.T) {
		frac := mustDecimal128BID(t, "1.5")
		nearest, flags := frac.RoundIntegralNearestEven()
		exact, exactFlags := frac.RoundIntegralExactWithFlags()
		if nearest != exact || !exactFlags.HasFlag(FlagInexact) {
			t.Fatalf("Decimal128 exact cross-check result/flags = %s/%s nearest=%s/%s", exact.String(), exactFlags, nearest.String(), flags)
		}
		for _, call := range []func() (Decimal128BID, ExceptionFlags){frac.RoundIntegralNearestEven, frac.RoundIntegralNearestAway, frac.RoundIntegralZero, frac.RoundIntegralPositive, frac.RoundIntegralNegative} {
			_, directedFlags := call()
			if directedFlags.HasFlag(FlagInexact) {
				t.Fatalf("Decimal128 directed RoundIntegral(1.5) flags = %s, want no Inexact", directedFlags)
			}
		}
		snan := mustDecimal128BID(t, "sNaN")
		for _, call := range []func() (Decimal128BID, ExceptionFlags){snan.RoundIntegralNearestEven, snan.RoundIntegralNearestAway, snan.RoundIntegralZero, snan.RoundIntegralPositive, snan.RoundIntegralNegative} {
			_, snanFlags := call()
			if snanFlags != FlagInvalidOperation {
				t.Fatalf("Decimal128 directed RoundIntegral(sNaN) flags = %s, want %s", snanFlags, FlagInvalidOperation)
			}
		}
	})
}
