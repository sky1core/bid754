package bid754

import (
	"testing"
	"unsafe"

	bidgo "github.com/sky1core/bid754/bid754-go/internal/bidgo"
)

type compareCase32 struct {
	name  string
	value Decimal32BID
}

type compareCase64 struct {
	name  string
	value Decimal64BID
}

type compareCase128 struct {
	name  string
	value Decimal128BID
}

type publicCompare32 func(Decimal32BID, Decimal32BID) (bool, ExceptionFlags)
type publicCompare64 func(Decimal64BID, Decimal64BID) (bool, ExceptionFlags)
type publicCompare128 func(Decimal128BID, Decimal128BID) (bool, ExceptionFlags)

type directCompare32 func(uint32, uint32) (int, uint32)
type directCompare64 func(uint64, uint64) (int, uint32)
type directCompare128 func(bidgo.BID_UINT128, bidgo.BID_UINT128) (int, uint32)

func rawDecimal128BID(lo, hi uint64) Decimal128BID {
	bits := bidUint128Words{w: [2]uint64{lo, hi}}
	return *(*Decimal128BID)(unsafe.Pointer(&bits))
}

func compareInputs32(t *testing.T) []compareCase32 {
	t.Helper()
	return []compareCase32{
		{name: "one", value: mustDecimal32BID(t, "1")},
		{name: "minus_two", value: mustDecimal32BID(t, "-2")},
		{name: "zero", value: mustDecimal32BID(t, "0")},
		{name: "negzero", value: mustDecimal32BID(t, "-0")},
		{name: "inf", value: mustDecimal32BID(t, "Infinity")},
		{name: "qnan", value: mustDecimal32BID(t, "NaN")},
		{name: "snan", value: mustDecimal32BID(t, "sNaN")},
		{name: "noncanonical", value: Decimal32BID(0x7c100000)},
	}
}

func compareInputs64(t *testing.T) []compareCase64 {
	t.Helper()
	return []compareCase64{
		{name: "one", value: mustDecimal64BID(t, "1")},
		{name: "minus_two", value: mustDecimal64BID(t, "-2")},
		{name: "zero", value: mustDecimal64BID(t, "0")},
		{name: "negzero", value: mustDecimal64BID(t, "-0")},
		{name: "inf", value: mustDecimal64BID(t, "Infinity")},
		{name: "qnan", value: mustDecimal64BID(t, "NaN")},
		{name: "snan", value: mustDecimal64BID(t, "sNaN")},
		{name: "noncanonical", value: Decimal64BID(0x7c04000000000000)},
	}
}

func compareInputs128(t *testing.T) []compareCase128 {
	t.Helper()
	return []compareCase128{
		{name: "one", value: mustDecimal128BID(t, "1")},
		{name: "minus_two", value: mustDecimal128BID(t, "-2")},
		{name: "zero", value: mustDecimal128BID(t, "0")},
		{name: "negzero", value: mustDecimal128BID(t, "-0")},
		{name: "inf", value: mustDecimal128BID(t, "Infinity")},
		{name: "qnan", value: mustDecimal128BID(t, "NaN")},
		{name: "snan", value: mustDecimal128BID(t, "sNaN")},
		{name: "noncanonical", value: rawDecimal128BID(0, 0x7c00400000000000)},
	}
}

func TestCompareInputAssumptions(t *testing.T) {
	inputs32 := compareInputs32(t)
	if bidgo.Bid32IsNaN32(inputs32[5].value.ToUint32()) == 0 || bidgo.Bid32IsSignaling(inputs32[5].value.ToUint32()) != 0 {
		t.Fatal("decimal32 qNaN assumption failed")
	}
	if bidgo.Bid32IsSignaling(inputs32[6].value.ToUint32()) == 0 {
		t.Fatal("decimal32 sNaN assumption failed")
	}
	if bidgo.Bid32IsCanonical(inputs32[7].value.ToUint32()) != 0 {
		t.Fatal("decimal32 noncanonical assumption failed")
	}

	inputs64 := compareInputs64(t)
	if bidgo.Bid64IsNaN(inputs64[5].value.ToUint64()) == 0 || bidgo.Bid64IsSignaling(inputs64[5].value.ToUint64()) != 0 {
		t.Fatal("decimal64 qNaN assumption failed")
	}
	if bidgo.Bid64IsSignaling(inputs64[6].value.ToUint64()) == 0 {
		t.Fatal("decimal64 sNaN assumption failed")
	}
	if bidgo.Bid64IsCanonical(inputs64[7].value.ToUint64()) != 0 {
		t.Fatal("decimal64 noncanonical assumption failed")
	}

	inputs128 := compareInputs128(t)
	if bidgo.Bid128IsNaN(decimal128BIDAsBidgo(inputs128[5].value)) == 0 || bidgo.Bid128IsSignaling(decimal128BIDAsBidgo(inputs128[5].value)) != 0 {
		t.Fatal("decimal128 qNaN assumption failed")
	}
	if bidgo.Bid128IsSignaling(decimal128BIDAsBidgo(inputs128[6].value)) == 0 {
		t.Fatal("decimal128 sNaN assumption failed")
	}
	if bidgo.Bid128IsCanonical(decimal128BIDAsBidgo(inputs128[7].value)) != 0 {
		t.Fatal("decimal128 noncanonical assumption failed")
	}
}

func TestComparePublicRoutingDecimal32(t *testing.T) {
	tests := []struct {
		name   string
		public publicCompare32
		direct directCompare32
	}{
		{name: "QuietEqual", public: Decimal32BID.QuietEqual, direct: bidgo.Bid32QuietEqual}, {name: "QuietNotEqual", public: Decimal32BID.QuietNotEqual, direct: bidgo.Bid32QuietNotEqual}, {name: "QuietGreater", public: Decimal32BID.QuietGreater, direct: bidgo.Bid32QuietGreater}, {name: "QuietGreaterEqual", public: Decimal32BID.QuietGreaterEqual, direct: bidgo.Bid32QuietGreaterEqual}, {name: "QuietGreaterUnordered", public: Decimal32BID.QuietGreaterUnordered, direct: bidgo.Bid32QuietGreaterUnordered}, {name: "QuietLess", public: Decimal32BID.QuietLess, direct: bidgo.Bid32QuietLess}, {name: "QuietLessEqual", public: Decimal32BID.QuietLessEqual, direct: bidgo.Bid32QuietLessEqual}, {name: "QuietLessUnordered", public: Decimal32BID.QuietLessUnordered, direct: bidgo.Bid32QuietLessUnordered}, {name: "QuietNotGreater", public: Decimal32BID.QuietNotGreater, direct: bidgo.Bid32QuietNotGreater}, {name: "QuietNotLess", public: Decimal32BID.QuietNotLess, direct: bidgo.Bid32QuietNotLess}, {name: "QuietOrdered", public: Decimal32BID.QuietOrdered, direct: bidgo.Bid32QuietOrdered}, {name: "QuietUnordered", public: Decimal32BID.QuietUnordered, direct: bidgo.Bid32QuietUnordered}, {name: "SignalingGreater", public: Decimal32BID.SignalingGreater, direct: bidgo.Bid32SignalingGreater}, {name: "SignalingGreaterEqual", public: Decimal32BID.SignalingGreaterEqual, direct: bidgo.Bid32SignalingGreaterEqual}, {name: "SignalingGreaterUnordered", public: Decimal32BID.SignalingGreaterUnordered, direct: bidgo.Bid32SignalingGreaterUnordered}, {name: "SignalingLess", public: Decimal32BID.SignalingLess, direct: bidgo.Bid32SignalingLess}, {name: "SignalingLessEqual", public: Decimal32BID.SignalingLessEqual, direct: bidgo.Bid32SignalingLessEqual}, {name: "SignalingLessUnordered", public: Decimal32BID.SignalingLessUnordered, direct: bidgo.Bid32SignalingLessUnordered}, {name: "SignalingNotGreater", public: Decimal32BID.SignalingNotGreater, direct: bidgo.Bid32SignalingNotGreater}, {name: "SignalingNotLess", public: Decimal32BID.SignalingNotLess, direct: bidgo.Bid32SignalingNotLess},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, left := range compareInputs32(t) {
				for _, right := range compareInputs32(t) {
					gotTruth, gotFlags := tt.public(left.value, right.value)
					wantTruth, wantFlags := tt.direct(left.value.ToUint32(), right.value.ToUint32())
					if gotTruth != (wantTruth != 0) || gotFlags != bidgoExceptionFlags(wantFlags) {
						t.Fatalf("%s(%s,%s) = %v/%s, want %v/%s", tt.name, left.name, right.name, gotTruth, gotFlags, wantTruth != 0, bidgoExceptionFlags(wantFlags))
					}
				}
			}
		})
	}
}

func TestComparePublicRoutingDecimal64(t *testing.T) {
	tests := []struct {
		name   string
		public publicCompare64
		direct directCompare64
	}{
		{name: "QuietEqual", public: Decimal64BID.QuietEqual, direct: bidgo.Bid64QuietEqual}, {name: "QuietNotEqual", public: Decimal64BID.QuietNotEqual, direct: bidgo.Bid64QuietNotEqual}, {name: "QuietGreater", public: Decimal64BID.QuietGreater, direct: bidgo.Bid64QuietGreater}, {name: "QuietGreaterEqual", public: Decimal64BID.QuietGreaterEqual, direct: bidgo.Bid64QuietGreaterEqual}, {name: "QuietGreaterUnordered", public: Decimal64BID.QuietGreaterUnordered, direct: bidgo.Bid64QuietGreaterUnordered}, {name: "QuietLess", public: Decimal64BID.QuietLess, direct: bidgo.Bid64QuietLess}, {name: "QuietLessEqual", public: Decimal64BID.QuietLessEqual, direct: bidgo.Bid64QuietLessEqual}, {name: "QuietLessUnordered", public: Decimal64BID.QuietLessUnordered, direct: bidgo.Bid64QuietLessUnordered}, {name: "QuietNotGreater", public: Decimal64BID.QuietNotGreater, direct: bidgo.Bid64QuietNotGreater}, {name: "QuietNotLess", public: Decimal64BID.QuietNotLess, direct: bidgo.Bid64QuietNotLess}, {name: "QuietOrdered", public: Decimal64BID.QuietOrdered, direct: bidgo.Bid64QuietOrdered}, {name: "QuietUnordered", public: Decimal64BID.QuietUnordered, direct: bidgo.Bid64QuietUnordered}, {name: "SignalingGreater", public: Decimal64BID.SignalingGreater, direct: bidgo.Bid64SignalingGreater}, {name: "SignalingGreaterEqual", public: Decimal64BID.SignalingGreaterEqual, direct: bidgo.Bid64SignalingGreaterEqual}, {name: "SignalingGreaterUnordered", public: Decimal64BID.SignalingGreaterUnordered, direct: bidgo.Bid64SignalingGreaterUnordered}, {name: "SignalingLess", public: Decimal64BID.SignalingLess, direct: bidgo.Bid64SignalingLess}, {name: "SignalingLessEqual", public: Decimal64BID.SignalingLessEqual, direct: bidgo.Bid64SignalingLessEqual}, {name: "SignalingLessUnordered", public: Decimal64BID.SignalingLessUnordered, direct: bidgo.Bid64SignalingLessUnordered}, {name: "SignalingNotGreater", public: Decimal64BID.SignalingNotGreater, direct: bidgo.Bid64SignalingNotGreater}, {name: "SignalingNotLess", public: Decimal64BID.SignalingNotLess, direct: bidgo.Bid64SignalingNotLess},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, left := range compareInputs64(t) {
				for _, right := range compareInputs64(t) {
					gotTruth, gotFlags := tt.public(left.value, right.value)
					wantTruth, wantFlags := tt.direct(left.value.ToUint64(), right.value.ToUint64())
					if gotTruth != (wantTruth != 0) || gotFlags != bidgoExceptionFlags(wantFlags) {
						t.Fatalf("%s(%s,%s) = %v/%s, want %v/%s", tt.name, left.name, right.name, gotTruth, gotFlags, wantTruth != 0, bidgoExceptionFlags(wantFlags))
					}
				}
			}
		})
	}
}

func TestComparePublicRoutingDecimal128(t *testing.T) {
	tests := []struct {
		name   string
		public publicCompare128
		direct directCompare128
	}{
		{name: "QuietEqual", public: Decimal128BID.QuietEqual, direct: bidgo.Bid128QuietEqual}, {name: "QuietNotEqual", public: Decimal128BID.QuietNotEqual, direct: bidgo.Bid128QuietNotEqual}, {name: "QuietGreater", public: Decimal128BID.QuietGreater, direct: bidgo.Bid128QuietGreater}, {name: "QuietGreaterEqual", public: Decimal128BID.QuietGreaterEqual, direct: bidgo.Bid128QuietGreaterEqual}, {name: "QuietGreaterUnordered", public: Decimal128BID.QuietGreaterUnordered, direct: bidgo.Bid128QuietGreaterUnordered}, {name: "QuietLess", public: Decimal128BID.QuietLess, direct: bidgo.Bid128QuietLess}, {name: "QuietLessEqual", public: Decimal128BID.QuietLessEqual, direct: bidgo.Bid128QuietLessEqual}, {name: "QuietLessUnordered", public: Decimal128BID.QuietLessUnordered, direct: bidgo.Bid128QuietLessUnordered}, {name: "QuietNotGreater", public: Decimal128BID.QuietNotGreater, direct: bidgo.Bid128QuietNotGreater}, {name: "QuietNotLess", public: Decimal128BID.QuietNotLess, direct: bidgo.Bid128QuietNotLess}, {name: "QuietOrdered", public: Decimal128BID.QuietOrdered, direct: bidgo.Bid128QuietOrdered}, {name: "QuietUnordered", public: Decimal128BID.QuietUnordered, direct: bidgo.Bid128QuietUnordered}, {name: "SignalingGreater", public: Decimal128BID.SignalingGreater, direct: bidgo.Bid128SignalingGreater}, {name: "SignalingGreaterEqual", public: Decimal128BID.SignalingGreaterEqual, direct: bidgo.Bid128SignalingGreaterEqual}, {name: "SignalingGreaterUnordered", public: Decimal128BID.SignalingGreaterUnordered, direct: bidgo.Bid128SignalingGreaterUnordered}, {name: "SignalingLess", public: Decimal128BID.SignalingLess, direct: bidgo.Bid128SignalingLess}, {name: "SignalingLessEqual", public: Decimal128BID.SignalingLessEqual, direct: bidgo.Bid128SignalingLessEqual}, {name: "SignalingLessUnordered", public: Decimal128BID.SignalingLessUnordered, direct: bidgo.Bid128SignalingLessUnordered}, {name: "SignalingNotGreater", public: Decimal128BID.SignalingNotGreater, direct: bidgo.Bid128SignalingNotGreater}, {name: "SignalingNotLess", public: Decimal128BID.SignalingNotLess, direct: bidgo.Bid128SignalingNotLess},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, left := range compareInputs128(t) {
				for _, right := range compareInputs128(t) {
					gotTruth, gotFlags := tt.public(left.value, right.value)
					wantTruth, wantFlags := tt.direct(decimal128BIDAsBidgo(left.value), decimal128BIDAsBidgo(right.value))
					if gotTruth != (wantTruth != 0) || gotFlags != bidgoExceptionFlags(wantFlags) {
						t.Fatalf("%s(%s,%s) = %v/%s, want %v/%s", tt.name, left.name, right.name, gotTruth, gotFlags, wantTruth != 0, bidgoExceptionFlags(wantFlags))
					}
				}
			}
		})
	}
}

func TestCompareFlagMatrix(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		one := mustDecimal32BID(t, "1")
		qnan := mustDecimal32BID(t, "NaN")
		snan := mustDecimal32BID(t, "sNaN")
		assertCompareFlags(t, "d32 quiet qNaN", compareFlags(one.QuietEqual(qnan)), 0)
		assertCompareFlags(t, "d32 quiet sNaN", compareFlags(one.QuietEqual(snan)), FlagInvalidOperation)
		assertCompareFlags(t, "d32 signaling qNaN", compareFlags(one.SignalingGreater(qnan)), FlagInvalidOperation)
		assertCompareFlags(t, "d32 signaling sNaN", compareFlags(one.SignalingGreater(snan)), FlagInvalidOperation)
		assertCompareFlags(t, "d32 derived qNaN", compareFlags(one.SignalingEqual(qnan)), FlagInvalidOperation)
		assertCompareFlags(t, "d32 derived sNaN", compareFlags(one.SignalingEqual(snan)), FlagInvalidOperation)
		assertCompareFlags(t, "d32 derived not equal qNaN", compareFlags(one.SignalingNotEqual(qnan)), FlagInvalidOperation)
		assertCompareFlags(t, "d32 derived not equal sNaN", compareFlags(one.SignalingNotEqual(snan)), FlagInvalidOperation)
	})
	t.Run("decimal64", func(t *testing.T) {
		one := mustDecimal64BID(t, "1")
		qnan := mustDecimal64BID(t, "NaN")
		snan := mustDecimal64BID(t, "sNaN")
		assertCompareFlags(t, "d64 quiet qNaN", compareFlags(one.QuietEqual(qnan)), 0)
		assertCompareFlags(t, "d64 quiet sNaN", compareFlags(one.QuietEqual(snan)), FlagInvalidOperation)
		assertCompareFlags(t, "d64 signaling qNaN", compareFlags(one.SignalingGreater(qnan)), FlagInvalidOperation)
		assertCompareFlags(t, "d64 signaling sNaN", compareFlags(one.SignalingGreater(snan)), FlagInvalidOperation)
		assertCompareFlags(t, "d64 derived qNaN", compareFlags(one.SignalingEqual(qnan)), FlagInvalidOperation)
		assertCompareFlags(t, "d64 derived sNaN", compareFlags(one.SignalingEqual(snan)), FlagInvalidOperation)
		assertCompareFlags(t, "d64 derived not equal qNaN", compareFlags(one.SignalingNotEqual(qnan)), FlagInvalidOperation)
		assertCompareFlags(t, "d64 derived not equal sNaN", compareFlags(one.SignalingNotEqual(snan)), FlagInvalidOperation)
	})
	t.Run("decimal128", func(t *testing.T) {
		one := mustDecimal128BID(t, "1")
		qnan := mustDecimal128BID(t, "NaN")
		snan := mustDecimal128BID(t, "sNaN")
		assertCompareFlags(t, "d128 quiet qNaN", compareFlags(one.QuietEqual(qnan)), 0)
		assertCompareFlags(t, "d128 quiet sNaN", compareFlags(one.QuietEqual(snan)), FlagInvalidOperation)
		assertCompareFlags(t, "d128 signaling qNaN", compareFlags(one.SignalingGreater(qnan)), FlagInvalidOperation)
		assertCompareFlags(t, "d128 signaling sNaN", compareFlags(one.SignalingGreater(snan)), FlagInvalidOperation)
		assertCompareFlags(t, "d128 derived qNaN", compareFlags(one.SignalingEqual(qnan)), FlagInvalidOperation)
		assertCompareFlags(t, "d128 derived sNaN", compareFlags(one.SignalingEqual(snan)), FlagInvalidOperation)
		assertCompareFlags(t, "d128 derived not equal qNaN", compareFlags(one.SignalingNotEqual(qnan)), FlagInvalidOperation)
		assertCompareFlags(t, "d128 derived not equal sNaN", compareFlags(one.SignalingNotEqual(snan)), FlagInvalidOperation)
	})
}

func TestCompareDerivedSignalingEqual(t *testing.T) {
	t.Run("decimal32", func(t *testing.T) {
		cases := []struct {
			name  string
			left  Decimal32BID
			right Decimal32BID
			want  bool
			flags ExceptionFlags
		}{
			{name: "same finite", left: mustDecimal32BID(t, "1"), right: mustDecimal32BID(t, "1"), want: true},
			{name: "different finite", left: mustDecimal32BID(t, "1"), right: mustDecimal32BID(t, "2"), want: false},
			{name: "signed zero", left: mustDecimal32BID(t, "0"), right: mustDecimal32BID(t, "-0"), want: true},
			{name: "same infinity", left: mustDecimal32BID(t, "Infinity"), right: mustDecimal32BID(t, "Infinity"), want: true},
			{name: "qnan", left: mustDecimal32BID(t, "NaN"), right: mustDecimal32BID(t, "1"), want: false, flags: FlagInvalidOperation},
		}
		for _, tc := range cases {
			got, flags := tc.left.SignalingEqual(tc.right)
			gotNot, notFlags := tc.left.SignalingNotEqual(tc.right)
			if got != tc.want || flags != tc.flags || gotNot != !tc.want || notFlags != tc.flags {
				t.Fatalf("%s: SignalingEqual=%v/%s SignalingNotEqual=%v/%s, want %v/%s and %v/%s", tc.name, got, flags, gotNot, notFlags, tc.want, tc.flags, !tc.want, tc.flags)
			}
			if tc.flags == 0 {
				quiet, quietFlags := tc.left.QuietEqual(tc.right)
				if quiet != got || quietFlags != 0 {
					t.Fatalf("%s: QuietEqual=%v/%s, SignalingEqual=%v/%s", tc.name, quiet, quietFlags, got, flags)
				}
			}
		}
	})
	t.Run("decimal64", func(t *testing.T) {
		cases := []struct {
			name  string
			left  Decimal64BID
			right Decimal64BID
			want  bool
			flags ExceptionFlags
		}{
			{name: "same finite", left: mustDecimal64BID(t, "1"), right: mustDecimal64BID(t, "1"), want: true},
			{name: "different finite", left: mustDecimal64BID(t, "1"), right: mustDecimal64BID(t, "2"), want: false},
			{name: "signed zero", left: mustDecimal64BID(t, "0"), right: mustDecimal64BID(t, "-0"), want: true},
			{name: "same infinity", left: mustDecimal64BID(t, "Infinity"), right: mustDecimal64BID(t, "Infinity"), want: true},
			{name: "qnan", left: mustDecimal64BID(t, "NaN"), right: mustDecimal64BID(t, "1"), want: false, flags: FlagInvalidOperation},
		}
		for _, tc := range cases {
			got, flags := tc.left.SignalingEqual(tc.right)
			gotNot, notFlags := tc.left.SignalingNotEqual(tc.right)
			if got != tc.want || flags != tc.flags || gotNot != !tc.want || notFlags != tc.flags {
				t.Fatalf("%s: SignalingEqual=%v/%s SignalingNotEqual=%v/%s, want %v/%s and %v/%s", tc.name, got, flags, gotNot, notFlags, tc.want, tc.flags, !tc.want, tc.flags)
			}
			if tc.flags == 0 {
				quiet, quietFlags := tc.left.QuietEqual(tc.right)
				if quiet != got || quietFlags != 0 {
					t.Fatalf("%s: QuietEqual=%v/%s, SignalingEqual=%v/%s", tc.name, quiet, quietFlags, got, flags)
				}
			}
		}
	})
	t.Run("decimal128", func(t *testing.T) {
		cases := []struct {
			name  string
			left  Decimal128BID
			right Decimal128BID
			want  bool
			flags ExceptionFlags
		}{
			{name: "same finite", left: mustDecimal128BID(t, "1"), right: mustDecimal128BID(t, "1"), want: true},
			{name: "different finite", left: mustDecimal128BID(t, "1"), right: mustDecimal128BID(t, "2"), want: false},
			{name: "signed zero", left: mustDecimal128BID(t, "0"), right: mustDecimal128BID(t, "-0"), want: true},
			{name: "same infinity", left: mustDecimal128BID(t, "Infinity"), right: mustDecimal128BID(t, "Infinity"), want: true},
			{name: "qnan", left: mustDecimal128BID(t, "NaN"), right: mustDecimal128BID(t, "1"), want: false, flags: FlagInvalidOperation},
		}
		for _, tc := range cases {
			got, flags := tc.left.SignalingEqual(tc.right)
			gotNot, notFlags := tc.left.SignalingNotEqual(tc.right)
			if got != tc.want || flags != tc.flags || gotNot != !tc.want || notFlags != tc.flags {
				t.Fatalf("%s: SignalingEqual=%v/%s SignalingNotEqual=%v/%s, want %v/%s and %v/%s", tc.name, got, flags, gotNot, notFlags, tc.want, tc.flags, !tc.want, tc.flags)
			}
			if tc.flags == 0 {
				quiet, quietFlags := tc.left.QuietEqual(tc.right)
				if quiet != got || quietFlags != 0 {
					t.Fatalf("%s: QuietEqual=%v/%s, SignalingEqual=%v/%s", tc.name, quiet, quietFlags, got, flags)
				}
			}
		}
	})
}

func compareFlags(_ bool, flags ExceptionFlags) ExceptionFlags {
	return flags
}

func assertCompareFlags(t *testing.T, name string, got ExceptionFlags, want ExceptionFlags) {
	t.Helper()
	if got != want {
		t.Fatalf("%s flags = %s, want %s", name, got, want)
	}
}
