// main.go - C export wrapper for Pure Go BID implementation
// Build with: CGO_ENABLED=1 go build -buildmode=c-archive -o libbidgo.a
package main

/*
#include <stdint.h>
#include <stdlib.h>

// Intel BID compatible types
typedef uint64_t BID_UINT64;
typedef uint32_t BID_UINT32;
typedef int32_t BID_SINT32;
typedef int64_t BID_SINT64;
typedef struct { uint64_t w[2]; } BID_UINT128;
typedef unsigned int _IDEC_flags;
typedef unsigned int _IDEC_round;

// Global variables defined in stubs.c
extern _IDEC_round __bid_IDEC_glbround;
extern _IDEC_flags __bid_IDEC_glbflags;

// Functions defined in stubs.c
extern int __bid_getDecimalRoundingDirection(void);
extern void __bid_signalException(_IDEC_flags f);

// Inline helper functions for Go
static inline _IDEC_round get_rnd_mode(void) { return __bid_IDEC_glbround; }
static inline void or_flags(_IDEC_flags f) { __bid_signalException(f); }
*/
import "C"

import (
	"math"
	"math/big"
	"math/bits"
	"strconv"
	"strings"
	"unsafe"

	bidgo "github.com/sky1core/bid754/bid-go"
)

// clampMode ensures rounding mode is within valid range (0-5)
func bidSizeofLong() int {
	return int(C.sizeof_long)
}

func clampMode(mode int) int {
	if mode < 0 || mode > 5 {
		return 0
	}
	return mode
}

const bidInvalidException uint32 = 0x01
const bidZeroDivideException uint32 = 0x04
const bidOverflowException uint32 = 0x08
const bidUnderflowException uint32 = 0x10
const bidInexactException uint32 = 0x20

var bid32RoundConstTable = [6][19]uint64{
	{0, 5, 50, 500, 5000, 50000, 500000, 5000000, 50000000, 500000000, 5000000000, 50000000000, 500000000000, 5000000000000, 50000000000000, 500000000000000, 5000000000000000, 50000000000000000, 500000000000000000},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 9, 99, 999, 9999, 99999, 999999, 9999999, 99999999, 999999999, 9999999999, 99999999999, 999999999999, 9999999999999, 99999999999999, 999999999999999, 9999999999999999, 99999999999999999, 999999999999999999},
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 5, 50, 500, 5000, 50000, 500000, 5000000, 50000000, 500000000, 5000000000, 50000000000, 500000000000, 5000000000000, 50000000000000, 500000000000000, 5000000000000000, 50000000000000000, 500000000000000000},
	{0, 4, 49, 499, 4999, 49999, 499999, 4999999, 49999999, 499999999, 4999999999, 49999999999, 499999999999, 4999999999999, 49999999999999, 499999999999999, 4999999999999999, 49999999999999999, 499999999999999999},
}

var bid32ShortRecipScale = [...]int{1, 1, 5, 7, 11, 14, 17, 21, 24, 27, 31, 34, 37, 41, 44, 47, 51, 54}

var bid32Reciprocals10_64 = [...]uint64{
	1,
	0x3333333333333334,
	0x51eb851eb851eb86,
	0x20c49ba5e353f7cf,
	0x346dc5d63886594b,
	0x29f16b11c6d1e109,
	0x218def416bdb1a6e,
	0x35afe535795e90b0,
	0x2af31dc4611873c0,
	0x225c17d04dad2966,
	0x36f9bfb3af7b7570,
	0x2bfaffc2f2c92ac0,
	0x232f33025bd42233,
	0x384b84d092ed0385,
	0x2d09370d42573604,
	0x24075f3dceac2b37,
	0x39a5652fb1137857,
	0x2e1dea8c8da92d13,
}

func bid64AnyNaN(x, y uint64) bool {
	return bidgo.Bid64IsNaN(x) != 0 || bidgo.Bid64IsNaN(y) != 0
}

func bid64AnySNaN(x, y uint64) bool {
	return bidgo.Bid64IsSignaling(x) != 0 || bidgo.Bid64IsSignaling(y) != 0
}

func bid64LessNoNaN(x, y uint64) bool {
	result, _ := bidgo.Bid64SignalingLess(x, y)
	return result != 0
}

func bid64EqualNoNaN(x, y uint64) bool {
	if bid64LessNoNaN(x, y) {
		return false
	}
	return !bid64LessNoNaN(y, x)
}

type intConvMode int

const (
	intConvTrunc intConvMode = iota
	intConvFloor
	intConvCeil
	intConvNearestEven
	intConvNearestAway
)

func bidPow10Big(n int) *big.Int {
	if n <= 0 {
		return big.NewInt(1)
	}
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func parseBid64DecimalString(s string) (neg bool, coeff *big.Int, exp10 int, special string) {
	if s == "" {
		return false, big.NewInt(0), 0, ""
	}
	if s[0] == '+' {
		s = s[1:]
	} else if s[0] == '-' {
		neg = true
		s = s[1:]
	}
	if strings.Contains(s, "Inf") {
		return neg, big.NewInt(0), 0, "inf"
	}
	if strings.Contains(s, "NaN") {
		return neg, big.NewInt(0), 0, "nan"
	}

	exp10 = 0
	if idx := strings.IndexAny(s, "Ee"); idx >= 0 {
		if e, err := strconv.Atoi(s[idx+1:]); err == nil {
			exp10 = e
		}
		s = s[:idx]
	}

	fracDigits := 0
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		fracDigits = len(s) - dot - 1
		s = s[:dot] + s[dot+1:]
	}
	if s == "" {
		return neg, big.NewInt(0), exp10 - fracDigits, ""
	}
	coeff = new(big.Int)
	coeff.SetString(s, 10)
	return neg, coeff, exp10 - fracDigits, ""
}

func bid64RoundedMagnitude(x uint64, mode intConvMode) (neg bool, rounded *big.Int, inexact bool, invalid bool) {
	s := bidgo.Bid64ToString(x)
	neg, coeff, exp10, special := parseBid64DecimalString(s)
	if special != "" {
		return neg, big.NewInt(0), false, true
	}
	if coeff.Sign() == 0 {
		return neg, big.NewInt(0), false, false
	}
	if exp10 >= 0 {
		return neg, new(big.Int).Mul(coeff, bidPow10Big(exp10)), false, false
	}

	divisor := bidPow10Big(-exp10)
	quotient := new(big.Int)
	remainder := new(big.Int)
	quotient.QuoRem(coeff, divisor, remainder)
	inexact = remainder.Sign() != 0
	rounded = new(big.Int).Set(quotient)
	if !inexact {
		return neg, rounded, false, false
	}

	switch mode {
	case intConvFloor:
		if neg {
			rounded.Add(rounded, big.NewInt(1))
		}
	case intConvCeil:
		if !neg {
			rounded.Add(rounded, big.NewInt(1))
		}
	case intConvNearestAway:
		twiceRem := new(big.Int).Lsh(remainder, 1)
		if twiceRem.Cmp(divisor) >= 0 {
			rounded.Add(rounded, big.NewInt(1))
		}
	case intConvNearestEven:
		twiceRem := new(big.Int).Lsh(remainder, 1)
		cmp := twiceRem.Cmp(divisor)
		if cmp > 0 || (cmp == 0 && rounded.Bit(0) == 1) {
			rounded.Add(rounded, big.NewInt(1))
		}
	}
	return neg, rounded, true, false
}

func bid64ToSignedFixed(x uint64, mode intConvMode, bits int, exact bool) (uint64, uint32) {
	neg, rounded, inexact, invalid := bid64RoundedMagnitude(x, mode)
	sentinel := uint64(1) << (bits - 1)
	flags := uint32(0)
	if invalid {
		return sentinel, bidInvalidException
	}

	maxPos := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits-1)), big.NewInt(1))
	maxNegMag := new(big.Int).Lsh(big.NewInt(1), uint(bits-1))

	if neg {
		if rounded.Cmp(maxNegMag) > 0 {
			return sentinel, bidInvalidException
		}
		if exact && inexact {
			flags |= 0x20
		}
		if bits == 64 {
			if rounded.Cmp(maxNegMag) == 0 {
				return sentinel, flags
			}
			return uint64(-rounded.Int64()), flags
		}
		mag := rounded.Int64()
		return uint64(uint32(int32(-mag))), flags
	}

	if rounded.Cmp(maxPos) > 0 {
		return sentinel, bidInvalidException
	}
	if exact && inexact {
		flags |= 0x20
	}
	if bits == 64 {
		return uint64(rounded.Int64()), flags
	}
	return uint64(uint32(int32(rounded.Int64()))), flags
}

func bid64ToUnsignedFixed(x uint64, mode intConvMode, bits int, exact bool) (uint64, uint32) {
	neg, rounded, inexact, invalid := bid64RoundedMagnitude(x, mode)
	sentinel := uint64(1) << (bits - 1)
	if invalid {
		return sentinel, bidInvalidException
	}
	if neg && rounded.Sign() != 0 {
		return sentinel, bidInvalidException
	}

	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
	if rounded.Cmp(max) > 0 {
		return sentinel, bidInvalidException
	}

	flags := uint32(0)
	if exact && inexact {
		flags |= 0x20
	}
	return rounded.Uint64(), flags
}

func bidRoundingModeToIntConv(mode int) intConvMode {
	switch clampMode(mode) {
	case 1:
		return intConvFloor
	case 2:
		return intConvCeil
	case 3:
		return intConvTrunc
	case 4:
		return intConvNearestAway
	default:
		return intConvNearestEven
	}
}

func bid64CanonicalizeRoundSpecial(x uint64) (uint64, bool) {
	if bidgo.Bid64IsNaN(x) != 0 {
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000
		} else {
			x = x & 0xfe03ffffffffffff
		}
		if bidgo.Bid64IsSignaling(x) != 0 {
			return x & 0xfdffffffffffffff, true
		}
		return x, true
	}
	if bidgo.Bid64IsInf(x) != 0 {
		return (x & 0x8000000000000000) | 0x7800000000000000, true
	}
	return 0, false
}

func bid64UnpackFiniteForRound(x uint64) (uint64, int, uint64) {
	xSign := x & 0x8000000000000000
	if (x & bidgo.MASK_STEERING_BITS64) == bidgo.MASK_STEERING_BITS64 {
		exp := int((x&bidgo.MASK_BINARY_EXPONENT2_64)>>51) - 398
		coeff := (x & bidgo.MASK_BINARY_SIG2_64) | bidgo.MASK_BINARY_OR2_64
		if coeff > 9999999999999999 {
			coeff = 0
		}
		return xSign, exp, coeff
	}
	exp := int((x&bidgo.MASK_BINARY_EXPONENT1_64)>>53) - 398
	coeff := x & bidgo.MASK_BINARY_SIG1_64
	return xSign, exp, coeff
}

func bid64UnpackForLogb(x uint64) (uint64, int, uint64, bool) {
	signX := x & 0x8000000000000000
	if (x & 0x6000000000000000) == 0x6000000000000000 {
		coeff := (x & 0x0007ffffffffffff) | 0x0020000000000000
		if (x & 0x7800000000000000) == 0x7800000000000000 {
			coefficientX := x & 0xfe03ffffffffffff
			if (x & 0x0003ffffffffffff) >= 1000000000000000 {
				coefficientX = x & 0xfe00000000000000
			}
			if (x & 0x7c00000000000000) == 0x7800000000000000 {
				coefficientX = x & 0xf800000000000000
			}
			return signX, 0, coefficientX, false
		}
		if coeff >= 10000000000000000 {
			coeff = 0
		}
		exponentX := int((x >> 51) & 0x3ff)
		return signX, exponentX, coeff, coeff != 0
	}
	exponentX := int((x >> 53) & 0x3ff)
	coefficientX := x & 0x001fffffffffffff
	return signX, exponentX, coefficientX, coefficientX != 0
}

func bid64RoundIntegralWithMode(x uint64, mode intConvMode) uint64 {
	if special, ok := bid64CanonicalizeRoundSpecial(x); ok {
		return special
	}
	xSign, exp, coeff := bid64UnpackFiniteForRound(x)
	if coeff == 0 {
		if exp < 0 {
			exp = 0
		}
		return xSign | (uint64(exp+398) << 53)
	}
	if exp >= 0 {
		return x
	}
	switch mode {
	case intConvFloor:
		r, _ := bidgo.Bid64RoundIntegralNegative(x)
		return r
	case intConvCeil:
		r, _ := bidgo.Bid64RoundIntegralPositive(x)
		return r
	case intConvTrunc:
		r, _ := bidgo.Bid64RoundIntegralZero(x)
		return r
	case intConvNearestAway:
		r, _ := bidgo.Bid64RoundIntegralNearestAway(x)
		return r
	default:
		r, _ := bidgo.Bid64RoundIntegralNearestEven(x)
		return r
	}
}

func bid64OverflowResult(sign uint64, mode int) uint64 {
	res := sign | 0x7800000000000000
	switch clampMode(mode) {
	case 1:
		if sign == 0 {
			res = 0x77fb86f26fc0ffff
		}
	case 2:
		if sign != 0 {
			res = 0xf7fb86f26fc0ffff
		}
	case 3:
		res = sign | 0x77fb86f26fc0ffff
	}
	return res
}

func bid64UnderflowResult(sign uint64, mode int) uint64 {
	switch clampMode(mode) {
	case 1:
		if sign != 0 {
			return 0x8000000000000001
		}
		return 0
	case 2:
		if sign == 0 {
			return 0x0000000000000001
		}
		return 0x8000000000000000
	case 4:
		if sign == 0 {
			return 0x0000000000000001
		}
		return 0x8000000000000001
	default:
		return sign
	}
}

func bid64ScalbnLike(x uint64, n int64) (uint64, uint32) {
	if bidgo.Bid64IsNaN(x) != 0 {
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x &= 0xfe00000000000000
		} else {
			x &= 0xfe03ffffffffffff
		}
		if bidgo.Bid64IsSignaling(x) != 0 {
			return x & 0xfdffffffffffffff, bidInvalidException
		}
		return x, 0
	}
	if bidgo.Bid64IsInf(x) != 0 {
		return (x & 0x8000000000000000) | 0x7800000000000000, 0
	}
	xSign, exp, coeff := bid64UnpackFiniteForRound(x)
	if coeff == 0 {
		newExp := int64(exp) + n
		if newExp > 369 {
			newExp = 369
		}
		if newExp < -398 {
			newExp = -398
		}
		return xSign | (uint64(newExp+398) << 53), 0
	}
	if n > 10000 {
		return bid64OverflowResult(xSign, int(C.get_rnd_mode())), bidOverflowException | bidInexactException
	}
	if n < -10000 {
		return bid64UnderflowResult(xSign, int(C.get_rnd_mode())), bidUnderflowException | bidInexactException
	}
	s := bidgo.Bid64ToString(x)
	neg, coeffBig, exp10, special := parseBid64DecimalString(s)
	if special != "" {
		return x, 0
	}
	newExp := int64(exp10) + n
	if newExp > 10000 {
		return bid64OverflowResult(xSign, int(C.get_rnd_mode())), bidOverflowException | bidInexactException
	}
	if newExp < -10000 {
		return bid64UnderflowResult(xSign, int(C.get_rnd_mode())), bidUnderflowException | bidInexactException
	}
	ss := coeffBig.String() + "E" + strconv.FormatInt(newExp, 10)
	if neg {
		ss = "-" + ss
	}
	r, flags := bidgo.Bid64FromString(ss, clampMode(int(C.get_rnd_mode())))
	mode := clampMode(int(C.get_rnd_mode()))
	if r == xSign && bidgo.Bid64IsZero(r) != 0 {
		normCoeff := new(big.Int).Set(coeffBig)
		normExp := newExp
		ten := big.NewInt(10)
		rem := new(big.Int)
		for normCoeff.Sign() != 0 && normExp < -398 {
			q, rr := new(big.Int), new(big.Int)
			q.QuoRem(normCoeff, ten, rr)
			if rr.Sign() != 0 {
				break
			}
			normCoeff = q
			normExp++
			rem = rr
			_ = rem
		}
		if normExp < -398 {
			return bid64UnderflowResult(xSign, mode), bidUnderflowException | bidInexactException
		}
	}
	return r, flags
}

func bid64PackRaw(sign uint64, biasedExp uint64, coeff uint64) uint64 {
	if (coeff & bidgo.MASK_BINARY_OR2_64) != 0 {
		return sign | (biasedExp << 51) | bidgo.MASK_STEERING_BITS64 | (coeff & bidgo.MASK_BINARY_SIG2_64)
	}
	return sign | (biasedExp << 53) | coeff
}

func bid64CanonicalizeNaN(x uint64) (uint64, uint32) {
	if (x & 0x0003ffffffffffff) > 999999999999999 {
		x &= 0xfe00000000000000
	} else {
		x &= 0xfe03ffffffffffff
	}
	if bidgo.Bid64IsSignaling(x) != 0 {
		return x & 0xfdffffffffffffff, bidInvalidException
	}
	return x, 0
}

func bid64CoeffDigits(coeff uint64) int {
	return len(strconv.FormatUint(coeff, 10))
}

var bidTen2k64 = [...]uint64{
	1,
	10,
	100,
	1000,
	10000,
	100000,
	1000000,
	10000000,
	100000000,
	1000000000,
	10000000000,
	100000000000,
	1000000000000,
	10000000000000,
	100000000000000,
	1000000000000000,
}

func bid64NextUpCore(x uint64) (uint64, uint32) {
	if bidgo.Bid64IsNaN(x) != 0 {
		return bid64CanonicalizeNaN(x)
	}
	if bidgo.Bid64IsInf(x) != 0 {
		if (x & 0x8000000000000000) == 0 {
			return 0x7800000000000000, 0
		}
		return 0xf7fb86f26fc0ffff, 0
	}
	xSign := x & 0x8000000000000000
	var xExp uint64
	var c1 uint64
	if (x & bidgo.MASK_STEERING_BITS64) == bidgo.MASK_STEERING_BITS64 {
		xExp = (x & bidgo.MASK_BINARY_EXPONENT2_64) >> 51
		c1 = (x & bidgo.MASK_BINARY_SIG2_64) | bidgo.MASK_BINARY_OR2_64
		if c1 > 9999999999999999 {
			xExp = 0
			c1 = 0
		}
	} else {
		xExp = (x & bidgo.MASK_BINARY_EXPONENT1_64) >> 53
		c1 = x & bidgo.MASK_BINARY_SIG1_64
	}
	if c1 == 0 {
		return 0x0000000000000001, 0
	}
	if x == 0x77fb86f26fc0ffff {
		return 0x7800000000000000, 0
	}
	if x == 0x8000000000000001 {
		return 0x8000000000000000, 0
	}
	q1 := bid64CoeffDigits(c1)
	if q1 < 16 {
		if xExp > uint64(16-q1) {
			ind := 16 - q1
			c1 *= bidTen2k64[ind]
			xExp -= uint64(ind)
		} else {
			ind := int(xExp)
			c1 *= bidTen2k64[ind]
			xExp = 0
		}
	}
	if xSign == 0 {
		c1++
		if c1 == 0x002386f26fc10000 {
			c1 = 0x00038d7ea4c68000
			xExp++
		}
	} else {
		c1--
		if c1 == 0x00038d7ea4c67fff && xExp != 0 {
			c1 = 0x002386f26fc0ffff
			xExp--
		}
	}
	return bid64PackRaw(xSign, xExp, c1), 0
}

func bid64NextDownCore(x uint64) (uint64, uint32) {
	if bidgo.Bid64IsNaN(x) != 0 {
		return bid64CanonicalizeNaN(x)
	}
	if bidgo.Bid64IsInf(x) != 0 {
		if (x & 0x8000000000000000) != 0 {
			return 0xf800000000000000, 0
		}
		return 0x77fb86f26fc0ffff, 0
	}
	xSign := x & 0x8000000000000000
	var xExp uint64
	var c1 uint64
	if (x & bidgo.MASK_STEERING_BITS64) == bidgo.MASK_STEERING_BITS64 {
		xExp = (x & bidgo.MASK_BINARY_EXPONENT2_64) >> 51
		c1 = (x & bidgo.MASK_BINARY_SIG2_64) | bidgo.MASK_BINARY_OR2_64
		if c1 > 9999999999999999 {
			xExp = 0
			c1 = 0
		}
	} else {
		xExp = (x & bidgo.MASK_BINARY_EXPONENT1_64) >> 53
		c1 = x & bidgo.MASK_BINARY_SIG1_64
	}
	if c1 == 0 {
		return 0x8000000000000001, 0
	}
	if x == 0xf7fb86f26fc0ffff {
		return 0xf800000000000000, 0
	}
	if x == 0x0000000000000001 {
		return 0x0000000000000000, 0
	}
	q1 := bid64CoeffDigits(c1)
	if q1 < 16 {
		if xExp > uint64(16-q1) {
			ind := 16 - q1
			c1 *= bidTen2k64[ind]
			xExp -= uint64(ind)
		} else {
			ind := int(xExp)
			c1 *= bidTen2k64[ind]
			xExp = 0
		}
	}
	if xSign != 0 {
		c1++
		if c1 == 0x002386f26fc10000 {
			c1 = 0x00038d7ea4c68000
			xExp++
		}
	} else {
		c1--
		if c1 == 0x00038d7ea4c67fff && xExp != 0 {
			c1 = 0x002386f26fc0ffff
			xExp--
		}
	}
	return bid64PackRaw(xSign, xExp, c1), 0
}

func bid128Pow10Big(n int) *big.Int {
	if n <= 0 {
		return big.NewInt(1)
	}
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
}

func bid128CoeffBig(hi, lo uint64) *big.Int {
	res := new(big.Int).SetUint64(hi)
	res.Lsh(res, 64)
	res.Or(res, new(big.Int).SetUint64(lo))
	return res
}

type bid128Decoded struct {
	sign   uint64
	exp    int
	coeff  *big.Int
	isNaN  bool
	isSNaN bool
	isInf  bool
	isZero bool
}

func bid128Decode(hi, lo uint64) bid128Decoded {
	d := bid128Decoded{sign: hi & 0x8000000000000000, coeff: big.NewInt(0)}
	if (hi & 0x7c00000000000000) == 0x7c00000000000000 {
		payloadHi := hi & 0x00003fffffffffff
		payloadLo := lo
		t33hi := uint64(0x0000314dc6448d93)
		t33lo := uint64(0x38c15b09ffffffff)
		if payloadHi > t33hi || (payloadHi == t33hi && payloadLo > t33lo) {
			payloadHi = 0
			payloadLo = 0
		}
		d.coeff = bid128CoeffBig(payloadHi, payloadLo)
		d.isNaN = true
		d.isSNaN = (hi & 0x7e00000000000000) == 0x7e00000000000000
		return d
	}
	if (hi & 0x7c00000000000000) == 0x7800000000000000 {
		d.isInf = true
		return d
	}
	d.exp = int((hi>>49)&0x3fff) - 6176
	coeffHi := hi & 0x0001ffffffffffff
	coeff := bid128CoeffBig(coeffHi, lo)
	if coeffHi > 0x0001ed09bead87c0 ||
		(coeffHi == 0x0001ed09bead87c0 && lo > 0x378d8e63ffffffff) ||
		((hi & 0x6000000000000000) == 0x6000000000000000) {
		coeff = big.NewInt(0)
	}
	d.coeff = coeff
	d.isZero = coeff.Sign() == 0
	return d
}

func bid128NaNToBid64(hi, lo uint64) (uint64, uint32) {
	payloadHi := hi & 0x00003fffffffffff
	payloadLo := lo
	t33hi := uint64(0x0000314dc6448d93)
	t33lo := uint64(0x38c15b09ffffffff)
	if payloadHi > t33hi || (payloadHi == t33hi && payloadLo > t33lo) {
		payloadHi = 0
		payloadLo = 0
	}
	payload := bid128CoeffBig(payloadHi, payloadLo)
	payload.Quo(payload, big.NewInt(1000000000000000000))
	return (hi & 0xfc00000000000000) | payload.Uint64(), func() uint32 {
		if (hi & 0x7e00000000000000) == 0x7e00000000000000 {
			return bidInvalidException
		}
		return 0
	}()
}

func bid64DecodeForCompare(x uint64) (sign uint64, exp int, coeff *big.Int, isZero bool) {
	sign, exp, c := bid64UnpackFiniteForRound(x)
	coeff = new(big.Int).SetUint64(c)
	return sign, exp, coeff, c == 0
}

func bid64CompareToBid128(x uint64, yHi, yLo uint64) int {
	x = bid64CanonicalizeNonCanonicalFinite(x)
	xSign, xExp, xCoeff, xZero := bid64DecodeForCompare(x)
	y := bid128Decode(yHi, yLo)
	if bidgo.Bid64IsInf(x) != 0 {
		if y.isInf {
			if xSign == y.sign {
				return 0
			}
			if xSign != 0 {
				return -1
			}
			return 1
		}
		if xSign != 0 {
			return -1
		}
		return 1
	}
	if y.isInf {
		if y.sign != 0 {
			return 1
		}
		return -1
	}
	if xZero && y.isZero {
		return 0
	}
	if xSign != y.sign {
		if xZero && y.isZero {
			return 0
		}
		if xSign != 0 {
			return -1
		}
		return 1
	}
	xc := new(big.Int).Set(xCoeff)
	yc := new(big.Int).Set(y.coeff)
	if xExp > y.exp {
		xc.Mul(xc, bid128Pow10Big(xExp-y.exp))
	} else if y.exp > xExp {
		yc.Mul(yc, bid128Pow10Big(y.exp-xExp))
	}
	cmp := xc.Cmp(yc)
	if xSign != 0 {
		cmp = -cmp
	}
	return cmp
}

func bid64CanonicalizeNonCanonicalFinite(x uint64) uint64 {
	if (x & 0x7800000000000000) == 0x7800000000000000 {
		return x
	}
	if (x & bidgo.MASK_STEERING_BITS64) == bidgo.MASK_STEERING_BITS64 {
		if ((x & bidgo.MASK_BINARY_SIG2_64) | bidgo.MASK_BINARY_OR2_64) > 9999999999999999 {
			return (x & 0x8000000000000000) | ((x & bidgo.MASK_BINARY_EXPONENT2_64) << 2)
		}
	}
	return x
}

func bid64UnpackIntel(x uint64) (sign uint64, exponent int, coefficient uint64, valid bool) {
	sign = x & 0x8000000000000000
	if (x & bidgo.SPECIAL_ENCODING_MASK64) == bidgo.SPECIAL_ENCODING_MASK64 {
		coefficient = (x & bidgo.LARGE_COEFF_MASK64) | bidgo.LARGE_COEFF_HIGH_BIT64
		if (x & bidgo.INFINITY_MASK64) == bidgo.INFINITY_MASK64 {
			exponent = 0
			coefficient = x & 0xfe03ffffffffffff
			if (x & 0x0003ffffffffffff) >= 1000000000000000 {
				coefficient = x & 0xfe00000000000000
			}
			if (x & bidgo.NAN_MASK64) == bidgo.INFINITY_MASK64 {
				coefficient = x & bidgo.SINFINITY_MASK64
			}
			return sign, exponent, coefficient, false
		}
		if coefficient >= 10000000000000000 {
			coefficient = 0
		}
		exponent = int((x >> bidgo.EXPONENT_SHIFT_LARGE64) & bidgo.EXPONENT_MASK64)
		return sign, exponent, coefficient, coefficient != 0
	}
	exponent = int((x >> bidgo.EXPONENT_SHIFT_SMALL64) & bidgo.EXPONENT_MASK64)
	coefficient = x & bidgo.SMALL_COEFF_MASK64
	return sign, exponent, coefficient, coefficient != 0
}

func bid64FastGet(sgn uint64, expon int, coeff uint64) uint64 {
	if coeff < (uint64(1) << bidgo.EXPONENT_SHIFT_SMALL64) {
		return (uint64(expon) << bidgo.EXPONENT_SHIFT_SMALL64) | (coeff | sgn)
	}
	if coeff == 10000000000000000 {
		return (uint64(expon+1) << bidgo.EXPONENT_SHIFT_SMALL64) | (1000000000000000 | sgn)
	}
	return (uint64(expon) << bidgo.EXPONENT_SHIFT_LARGE64) | (sgn | bidgo.SPECIAL_ENCODING_MASK64) | (coeff & (((uint64(1) << bidgo.EXPONENT_SHIFT_SMALL64) >> 2) - 1))
}

func bid32VeryFastGet(sgn uint32, expon int, coeff uint32) uint32 {
	if coeff < (1 << 23) {
		return (uint32(expon) << 23) | (coeff | sgn)
	}
	return (uint32(expon) << 21) | (sgn | 0x60000000) | (coeff & 0x001fffff)
}

func bid32Get(sgn uint32, expon int, coeff uint64, rmode int) uint32 {
	if coeff > 9999999 {
		expon++
		coeff = 1000000
	}
	if uint(expon) > 191 {
		if expon < 0 {
			if expon+7 < 0 {
				if rmode == 1 && sgn != 0 {
					return 0x80000001
				}
				if rmode == 2 && sgn == 0 {
					return 0x00000001
				}
				return sgn
			}
			if sgn != 0 && uint(rmode-1) < 2 {
				rmode = 3 - rmode
			}
			extraDigits := -expon
			coeff += bid32RoundConstTable[rmode][extraDigits]
			hi, lo := bits.Mul64(coeff, bid32Reciprocals10_64[extraDigits])
			amount := bid32ShortRecipScale[extraDigits]
			coeff = hi >> amount
			if rmode == 0 && (coeff&1) != 0 {
				remainderH := hi & ((uint64(1) << amount) - 1)
				if remainderH == 0 && lo < bid32Reciprocals10_64[extraDigits] {
					coeff--
				}
			}
			return sgn | uint32(coeff)
		}
		if coeff == 0 && expon > 191 {
			expon = 191
		}
		for coeff < 1000000 && expon >= 192 {
			expon--
			coeff = (coeff << 3) + (coeff << 1)
		}
		if expon > 191 {
			r := sgn | 0x78000000
			switch rmode {
			case 1:
				if sgn == 0 {
					r = 0x77f8967f
				}
			case 3:
				r = sgn | 0x77f8967f
			case 2:
				if sgn != 0 {
					r = 0xf7f8967f
				}
			}
			return r
		}
	}
	return bid32VeryFastGet(sgn, expon, uint32(coeff))
}

func floorLog2Rat(num, den *big.Int) int {
	exp2 := num.BitLen() - den.BitLen()
	if exp2 >= 0 {
		t := new(big.Int).Lsh(new(big.Int).Set(den), uint(exp2))
		if num.Cmp(t) < 0 {
			exp2--
		}
	} else {
		t := new(big.Int).Lsh(new(big.Int).Set(num), uint(-exp2))
		if t.Cmp(den) < 0 {
			exp2--
		}
	}
	return exp2
}

func roundRatToInt(num, den *big.Int, sign uint64, mode int) (*big.Int, bool) {
	q := new(big.Int)
	r := new(big.Int)
	q.QuoRem(num, den, r)
	if r.Sign() == 0 {
		return q, false
	}
	inexact := true
	twoR := new(big.Int).Lsh(new(big.Int).Set(r), 1)
	switch mode {
	case 0:
		cmp := twoR.Cmp(den)
		if cmp > 0 || (cmp == 0 && q.Bit(0) == 1) {
			q.Add(q, big.NewInt(1))
		}
	case 4:
		if twoR.Cmp(den) >= 0 {
			q.Add(q, big.NewInt(1))
		}
	case 3:
	case 2:
		if sign == 0 {
			q.Add(q, big.NewInt(1))
		}
	case 1:
		if sign != 0 {
			q.Add(q, big.NewInt(1))
		}
	}
	return q, inexact
}

func bid64FiniteToBinaryBits(sign uint64, exp10 int, coeff uint64, p, bias, expBits, fracBits, totalBits int, mode int) (uint64, uint32) {
	num := new(big.Int).SetUint64(coeff)
	den := big.NewInt(1)
	if exp10 >= 0 {
		num.Mul(num, bid128Pow10Big(exp10))
	} else {
		den = bid128Pow10Big(-exp10)
	}
	emin := 1 - bias
	emax := bias
	signBit := uint(totalBits - 1)
	maxExpField := uint64((uint64(1) << uint(expBits)) - 1)
	exp2 := floorLog2Rat(num, den)
	flags := uint32(0)

	if exp2 < emin {
		scale := fracBits - emin
		scaledNum := new(big.Int).Lsh(new(big.Int).Set(num), uint(scale))
		m, inexact := roundRatToInt(scaledNum, den, sign, mode)
		if m.Sign() == 0 {
			if inexact {
				flags |= bidUnderflowException | bidInexactException
			}
			return sign << signBit, flags
		}
		limit := new(big.Int).Lsh(big.NewInt(1), uint(fracBits))
		if m.Cmp(limit) >= 0 {
			expBits := uint64(emin + bias)
			frac := new(big.Int).Sub(m, limit)
			if inexact {
				flags |= bidUnderflowException | bidInexactException
			}
			return (sign << signBit) | (expBits << uint(fracBits)) | frac.Uint64(), flags
		}
		if inexact {
			flags |= bidUnderflowException | bidInexactException
		}
		return (sign << signBit) | m.Uint64(), flags
	}

	scale := fracBits - exp2
	var scaledNum, scaledDen *big.Int
	if scale >= 0 {
		scaledNum = new(big.Int).Lsh(new(big.Int).Set(num), uint(scale))
		scaledDen = new(big.Int).Set(den)
	} else {
		scaledNum = new(big.Int).Set(num)
		scaledDen = new(big.Int).Lsh(new(big.Int).Set(den), uint(-scale))
	}
	m, inexact := roundRatToInt(scaledNum, scaledDen, sign, mode)
	limit := new(big.Int).Lsh(big.NewInt(1), uint(fracBits+1))
	hidden := new(big.Int).Lsh(big.NewInt(1), uint(fracBits))
	if m.Cmp(limit) >= 0 {
		m.Rsh(m, 1)
		exp2++
	}
	if exp2 > emax {
		flags = bidOverflowException | bidInexactException
		if (sign == 0 && (mode == 1 || mode == 3)) || (sign != 0 && (mode == 2 || mode == 3)) {
			maxFrac := uint64((uint64(1) << uint(fracBits)) - 1)
			return (sign << signBit) | ((maxExpField - 1) << uint(fracBits)) | maxFrac, flags
		}
		return (sign << signBit) | (maxExpField << uint(fracBits)), flags
	}
	if inexact {
		flags |= bidInexactException
	}
	frac := new(big.Int).Sub(m, hidden)
	return (sign << signBit) | (uint64(exp2+bias) << uint(fracBits)) | frac.Uint64(), flags
}

func bid64FiniteToBinary128Bits(sign uint64, exp10 int, coeff uint64, mode int) (uint64, uint64, uint32) {
	num := new(big.Int).SetUint64(coeff)
	den := big.NewInt(1)
	if exp10 >= 0 {
		num.Mul(num, bid128Pow10Big(exp10))
	} else {
		den = bid128Pow10Big(-exp10)
	}
	const bias = 16383
	const fracBits = 112
	const expBits = 15
	emin := 1 - bias
	emax := bias
	exp2 := floorLog2Rat(num, den)
	flags := uint32(0)

	pack := func(sign uint64, expField uint64, frac *big.Int) (uint64, uint64) {
		v := new(big.Int).SetUint64(sign)
		v.Lsh(v, 127)
		if expField != 0 {
			t := new(big.Int).SetUint64(expField)
			t.Lsh(t, fracBits)
			v.Or(v, t)
		}
		if frac != nil && frac.Sign() != 0 {
			v.Or(v, frac)
		}
		lo := v.Uint64()
		hi := new(big.Int).Rsh(v, 64).Uint64()
		return hi, lo
	}

	if exp2 < emin {
		scale := fracBits - emin
		scaledNum := new(big.Int).Lsh(new(big.Int).Set(num), uint(scale))
		m, inexact := roundRatToInt(scaledNum, den, sign, mode)
		if m.Sign() == 0 {
			if inexact {
				flags |= bidUnderflowException | bidInexactException
			}
			return sign << 63, 0, flags
		}
		limit := new(big.Int).Lsh(big.NewInt(1), fracBits)
		if m.Cmp(limit) >= 0 {
			frac := new(big.Int).Sub(m, limit)
			if inexact {
				flags |= bidUnderflowException | bidInexactException
			}
			hi, lo := pack(sign, 1, frac)
			return hi, lo, flags
		}
		if inexact {
			flags |= bidUnderflowException | bidInexactException
		}
		hi, lo := pack(sign, 0, m)
		return hi, lo, flags
	}

	scale := fracBits - exp2
	var scaledNum, scaledDen *big.Int
	if scale >= 0 {
		scaledNum = new(big.Int).Lsh(new(big.Int).Set(num), uint(scale))
		scaledDen = new(big.Int).Set(den)
	} else {
		scaledNum = new(big.Int).Set(num)
		scaledDen = new(big.Int).Lsh(new(big.Int).Set(den), uint(-scale))
	}
	m, inexact := roundRatToInt(scaledNum, scaledDen, sign, mode)
	limit := new(big.Int).Lsh(big.NewInt(1), fracBits+1)
	hidden := new(big.Int).Lsh(big.NewInt(1), fracBits)
	if m.Cmp(limit) >= 0 {
		m.Rsh(m, 1)
		exp2++
	}
	if exp2 > emax {
		flags = bidOverflowException | bidInexactException
		maxExpField := uint64((uint64(1) << expBits) - 1)
		if (sign == 0 && (mode == 1 || mode == 3)) || (sign != 0 && (mode == 2 || mode == 3)) {
			maxFrac := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), fracBits), big.NewInt(1))
			hi, lo := pack(sign, maxExpField-1, maxFrac)
			return hi, lo, flags
		}
		hi, lo := pack(sign, maxExpField, nil)
		return hi, lo, flags
	}
	if inexact {
		flags |= bidInexactException
	}
	frac := new(big.Int).Sub(m, hidden)
	hi, lo := pack(sign, uint64(exp2+bias), frac)
	return hi, lo, flags
}

func bid64RemCore(x, y uint64) (uint64, uint32) {
	return bidgo.Bid64Rem(x, y)
}

func bid64QuantizeFlags(x, y uint64) uint32 {
	flags := uint32(0)
	if bidgo.Bid64IsSignaling(y) != 0 {
		flags |= bidInvalidException
	}
	if bidgo.Bid64IsNaN(y) != 0 {
		return flags
	}
	if bidgo.Bid64IsInf(y) != 0 {
		if bidgo.Bid64IsFinite(x) != 0 {
			flags |= bidInvalidException
		}
		return flags
	}
	if bidgo.Bid64IsNaN(x) != 0 {
		if bidgo.Bid64IsSignaling(x) != 0 {
			flags |= bidInvalidException
		}
		return flags
	}
	if bidgo.Bid64IsInf(x) != 0 {
		flags |= bidInvalidException
		return flags
	}
	_, expX, coeffX := bid64UnpackFiniteForRound(x)
	_, expY, _ := bid64UnpackFiniteForRound(y)
	if coeffX == 0 {
		return flags
	}
	digitsX := bid64CoeffDigits(coeffX)
	exponDiff := expX - expY
	totalDigits := digitsX + exponDiff
	if totalDigits > 16 {
		flags |= bidInvalidException
		return flags
	}
	if totalDigits < 0 {
		flags |= bidInexactException
		return flags
	}
	if exponDiff >= 0 {
		return flags
	}
	extraDigits := -exponDiff
	if extraDigits > 0 && extraDigits < len(bidTen2k64) {
		if coeffX%bidTen2k64[extraDigits] != 0 {
			flags |= bidInexactException
		}
	}
	return flags
}

func bid64NextAfterCore(x, y uint64) (uint64, uint32) {
	flags := uint32(0)
	if bidgo.Bid64IsNaN(x) != 0 {
		res, f := bid64CanonicalizeNaN(x)
		flags |= f
		if bidgo.Bid64IsSignaling(y) != 0 {
			flags |= bidInvalidException
		}
		return res, flags
	}
	if bidgo.Bid64IsNaN(y) != 0 {
		res, f := bid64CanonicalizeNaN(y)
		flags |= f
		return res, flags
	}
	if bidgo.Bid64IsInf(x) != 0 {
		x &= 0xf800000000000000
	}
	if bidgo.Bid64IsInf(y) != 0 {
		y &= 0xf800000000000000
	}
	x = bid64CanonicalizeNonCanonicalFinite(x)
	if bid64EqualNoNaN(x, y) {
		return (y & 0x8000000000000000) | (x & 0x7fffffffffffffff), flags
	}
	var res uint64
	if bid64LessNoNaN(y, x) {
		res, _ = bid64NextDownCore(x)
	} else {
		res, _ = bid64NextUpCore(x)
	}
	if bidgo.Bid64IsInf(x) == 0 && bidgo.Bid64IsInf(res) != 0 {
		flags |= bidInexactException | bidOverflowException
	}
	absRes := res & 0x7fffffffffffffff
	if bid64LessNoNaN(absRes, 0x00038d7ea4c68000) && !bid64EqualNoNaN(x, res) {
		flags |= bidInexactException | bidUnderflowException
	}
	return res, flags
}

func bid64AdjustedExponent(x uint64) (int, uint32, bool) {
	s := bidgo.Bid64ToString(x)
	_, coeff, exp10, special := parseBid64DecimalString(s)
	switch special {
	case "inf":
		return 0, bidInvalidException, false
	case "nan":
		return 0, bidInvalidException, false
	}
	if coeff.Sign() == 0 {
		return 0, bidInvalidException, false
	}
	return exp10 + len(coeff.String()) - 1, 0, true
}

func bid64FiniteCoeffExp(x uint64) (neg bool, coeff *big.Int, exp10 int) {
	sign, exponent, coefficient, _ := bid64UnpackIntel(x)
	return sign != 0, new(big.Int).SetUint64(coefficient), exponent - 398
}

func decimalCoeffExpString(neg bool, coeff *big.Int, exp10 int) string {
	var b strings.Builder
	if neg && coeff.Sign() != 0 {
		b.WriteByte('-')
	}
	b.WriteString(coeff.String())
	if exp10 != 0 {
		b.WriteByte('E')
		b.WriteString(strconv.Itoa(exp10))
	}
	return b.String()
}

func roundDecimalToBID64(sum *big.Int, exp int, signMask uint64, mode int) (uint64, uint32) {
	coeff := new(big.Int).Set(sum)
	flags := uint32(0)
	signBit := uint64(0)
	if signMask != 0 {
		signBit = 1
	}
	if coeff.Sign() == 0 {
		return bid64PackRaw(signMask, uint64(398), 0), 0
	}

	if exp < -398 {
		divisor := bidPow10Big(-398 - exp)
		rounded, inexact := roundRatToInt(coeff, divisor, signBit, mode)
		coeff = rounded
		exp = -398
		if inexact {
			flags |= bidInexactException
		}
		if coeff.Sign() == 0 {
			if flags != 0 {
				flags |= bidUnderflowException
			}
			return signMask, flags
		}
	}

	digits := len(coeff.String())
	if digits > 16 {
		divisor := bidPow10Big(digits - 16)
		rounded, inexact := roundRatToInt(coeff, divisor, signBit, mode)
		coeff = rounded
		exp += digits - 16
		if inexact {
			flags |= bidInexactException
		}
	}

	limit := new(big.Int).SetUint64(10000000000000000)
	ten := big.NewInt(10)
	if coeff.Cmp(limit) >= 0 {
		coeff.Quo(coeff, ten)
		exp++
	}

	if exp > 369 {
		flags |= bidOverflowException | bidInexactException
		if (signMask == 0 && (mode == 1 || mode == 3)) || (signMask != 0 && (mode == 2 || mode == 3)) {
			return bid64PackRaw(signMask, 767, 9999999999999999), flags
		}
		return signMask | 0x7800000000000000, flags
	}

	if exp == -398 && (flags&bidInexactException) != 0 {
		flags |= bidUnderflowException
	}

	return bid64PackRaw(signMask, uint64(exp+398), coeff.Uint64()), flags
}

// ============ IMPLEMENTED FUNCTIONS ============

func bid64ToBinary32(x uint64, rndMode int) (uint32, uint32) {
	return bidgo.Bid64ToBinary32(x, clampMode(rndMode))
}

func bid64ToBinary64(x uint64, rndMode int) (uint64, uint32) {
	return bidgo.Bid64ToBinary64(x, clampMode(rndMode))
}

func bid64ToBinary128(x uint64, rndMode int) (uint64, uint64, uint32) {
	res, flags := bidgo.Bid64ToBinary128(x, clampMode(rndMode))
	resWords := *(*[2]uint64)(unsafe.Pointer(&res))
	return resWords[1], resWords[0], flags
}

//export __bid64_to_binary32
func __bid64_to_binary32(x C.BID_UINT64) C.float {
	bits32, flags := bid64ToBinary32(uint64(x), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.float(math.Float32frombits(bits32))
}

//export __bid64_to_binary64
func __bid64_to_binary64(x C.BID_UINT64) C.double {
	bits64, flags := bid64ToBinary64(uint64(x), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.double(math.Float64frombits(bits64))
}

//export __bid64_to_binary128
func __bid64_to_binary128(x C.BID_UINT64) C.BID_UINT128 {
	var r C.BID_UINT128
	hi, lo, flags := bid64ToBinary128(uint64(x), int(C.get_rnd_mode()))
	r.w[0] = C.uint64_t(lo)
	r.w[1] = C.uint64_t(hi)
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return r
}

//export __bid64_sqrt
func __bid64_sqrt(x C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64Sqrt(uint64(x), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_nan
func __bid64_nan(s *C.char) C.BID_UINT64 {
	res := uint64(0x7c00000000000000)
	if s == nil {
		return C.BID_UINT64(res)
	}
	x, _ := bidgo.Bid64FromString(C.GoString(s), 0)
	res |= x & 0x0003ffffffffffff
	return C.BID_UINT64(res)
}

//export __bid64_to_bid128
func __bid64_to_bid128(x C.BID_UINT64) C.BID_UINT128 {
	var r C.BID_UINT128
	res, flags := bidgo.Bid64ToBid128(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	resWords := *(*[2]uint64)(unsafe.Pointer(&res))
	r.w[0] = C.uint64_t(resWords[0])
	r.w[1] = C.uint64_t(resWords[1])
	return r
}

//export __bid64_to_bid32
func __bid64_to_bid32(x C.BID_UINT64) C.BID_UINT32 {
	res, flags := bidgo.Bid64ToBid32(uint64(x), clampMode(int(C.get_rnd_mode())))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT32(res)
}

//export __bid64_scalbn
func __bid64_scalbn(x C.BID_UINT64, n C.int) C.BID_UINT64 {
	r, flags := bidgo.Bid64Scalbn(uint64(x), int(n), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(r)
}

//export __bid64_scalbln
func __bid64_scalbln(x C.BID_UINT64, n C.long) C.BID_UINT64 {
	r, flags := bidgo.Bid64Scalbln(uint64(x), int64(n), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(r)
}

//export __bid64_ldexp
func __bid64_ldexp(x C.BID_UINT64, n C.int) C.BID_UINT64 {
	r, flags := bidgo.Bid64Ldexp(uint64(x), int(n), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(r)
}

//export __bid64_nextup
func __bid64_nextup(x C.BID_UINT64) C.BID_UINT64 {
	r, flags := bidgo.Bid64NextUp(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(r)
}

//export __bid64_nextdown
func __bid64_nextdown(x C.BID_UINT64) C.BID_UINT64 {
	r, flags := bidgo.Bid64NextDown(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(r)
}

//export __bid64_nextafter
func __bid64_nextafter(x C.BID_UINT64, y C.BID_UINT64) C.BID_UINT64 {
	r, flags := bidgo.Bid64NextAfter(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(r)
}

//export __bid64_rem
func __bid64_rem(x C.BID_UINT64, y C.BID_UINT64) C.BID_UINT64 {
	r, flags := bid64RemCore(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(r)
}

//export __bid64_fmod
func __bid64_fmod(x C.BID_UINT64, y C.BID_UINT64) C.BID_UINT64 {
	r, flags := bidgo.Bid64Fmod(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(r)
}

//export __bid64_nexttoward
func __bid64_nexttoward(x C.BID_UINT64, y C.BID_UINT128) C.BID_UINT64 {
	ux := uint64(x)
	yHi := uint64(y.w[1])
	yLo := uint64(y.w[0])
	if yHi == 0 && (yLo == 0x7800000000000000 || yLo == 0xf800000000000000) {
		yHi, yLo = yLo, yHi
	}
	flags := uint32(0)
	if bidgo.Bid64IsNaN(ux) != 0 {
		res, f := bid64CanonicalizeNaN(ux)
		flags |= f
		yd := bid128Decode(yHi, yLo)
		if yd.isSNaN {
			flags |= bidInvalidException
		}
		if flags != 0 {
			C.or_flags(C._IDEC_flags(flags))
		}
		return C.BID_UINT64(res)
	}
	if bidgo.Bid64IsInf(ux) != 0 {
		ux = (ux & 0x8000000000000000) | 0x7800000000000000
	} else {
		ux = bid64CanonicalizeNonCanonicalFinite(ux)
	}
	yd := bid128Decode(yHi, yLo)
	if yd.isNaN {
		res, f := bid128NaNToBid64(yHi, yLo)
		flags |= f
		if flags != 0 {
			C.or_flags(C._IDEC_flags(flags))
		}
		return C.BID_UINT64(res)
	}
	cmp := bid64CompareToBid128(ux, yHi, yLo)
	var res uint64
	if cmp == 0 {
		res = (yd.sign & 0x8000000000000000) | (ux & 0x7fffffffffffffff)
	} else if cmp > 0 {
		res, _ = bidgo.Bid64NextDown(ux)
	} else {
		res, _ = bidgo.Bid64NextUp(ux)
	}
	if bidgo.Bid64IsInf(ux) == 0 && bidgo.Bid64IsInf(res) != 0 {
		flags |= bidInexactException | bidOverflowException
	}
	absRes := res & 0x7fffffffffffffff
	if bid64LessNoNaN(absRes, 0x00038d7ea4c68000) && !bid64EqualNoNaN(ux, res) {
		flags |= bidInexactException | bidUnderflowException
	}
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_quantize
func __bid64_quantize(x C.BID_UINT64, y C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64Quantize(uint64(x), uint64(y), clampMode(int(C.get_rnd_mode())))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_fdim
func __bid64_fdim(x C.BID_UINT64, y C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64Fdim(uint64(x), uint64(y), clampMode(int(C.get_rnd_mode())))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_modf
func __bid64_modf(x C.BID_UINT64, iptr *C.BID_UINT64) C.BID_UINT64 {
	res, ires, flags := bidgo.Bid64Modf(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	*iptr = C.BID_UINT64(ires)
	return C.BID_UINT64(res)
}

//export __bid64_frexp
func __bid64_frexp(x C.BID_UINT64, exp *C.int) C.BID_UINT64 {
	res, e := bidgo.Bid64Frexp(uint64(x))
	*exp = C.int(e)
	return C.BID_UINT64(res)
}

//export __bid64_add
func __bid64_add(x, y C.BID_UINT64) C.BID_UINT64 {
	mode := clampMode(int(C.get_rnd_mode()))
	result, flags := bidgo.Bid64AddWithFlags(uint64(x), uint64(y), mode)
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_sub
func __bid64_sub(x, y C.BID_UINT64) C.BID_UINT64 {
	mode := clampMode(int(C.get_rnd_mode()))
	result, flags := bidgo.Bid64SubWithFlags(uint64(x), uint64(y), mode)
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_mul
func __bid64_mul(x, y C.BID_UINT64) C.BID_UINT64 {
	mode := clampMode(int(C.get_rnd_mode()))
	result, flags := bidgo.Bid64MulWithFlags(uint64(x), uint64(y), mode)
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_div
func __bid64_div(x, y C.BID_UINT64) C.BID_UINT64 {
	mode := clampMode(int(C.get_rnd_mode()))
	result, flags := bidgo.Bid64DivWithFlags(uint64(x), uint64(y), mode)
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_fma
func __bid64_fma(x, y, z C.BID_UINT64) C.BID_UINT64 {
	result, flags := bidgo.Bid64Fma(uint64(x), uint64(y), uint64(z), clampMode(int(C.get_rnd_mode())))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_from_string
func __bid64_from_string(s *C.char) C.BID_UINT64 {
	mode := clampMode(int(C.get_rnd_mode()))
	goStr := C.GoString(s)
	// Use mechanically ported bid64_from_string
	result, flags := bidgo.Bid64FromString(goStr, mode)
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_to_string
func __bid64_to_string(ps *C.char, x C.BID_UINT64) {
	result := bidgo.Bid64ToString(uint64(x))
	// Copy result to C buffer
	for i := 0; i < len(result); i++ {
		*(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(ps)) + uintptr(i))) = C.char(result[i])
	}
	*(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(ps)) + uintptr(len(result)))) = 0
}

//export __bid64_from_int32
func __bid64_from_int32(x C.BID_SINT32) C.BID_UINT64 {
	return C.BID_UINT64(bidgo.Bid64FromInt32(int32(x)))
}

//export __bid64_from_uint32
func __bid64_from_uint32(x C.BID_UINT32) C.BID_UINT64 {
	return C.BID_UINT64(bidgo.Bid64FromUint32(uint32(x)))
}

//export __bid64_from_int64
func __bid64_from_int64(x C.BID_SINT64) C.BID_UINT64 {
	mode := clampMode(int(C.get_rnd_mode()))
	result, flags := bidgo.Bid64FromInt64(int64(x), mode)
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_from_uint64
func __bid64_from_uint64(x C.BID_UINT64) C.BID_UINT64 {
	mode := clampMode(int(C.get_rnd_mode()))
	result, flags := bidgo.Bid64FromUint64(uint64(x), mode)
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

// ============ NON-COMPUTATIONAL FUNCTIONS ============

//export __bid64_isSigned
func __bid64_isSigned(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64IsSigned(uint64(x)))
}

//export __bid64_isNaN
func __bid64_isNaN(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64IsNaN(uint64(x)))
}

//export __bid64_isFinite
func __bid64_isFinite(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64IsFinite(uint64(x)))
}

//export __bid64_isInf
func __bid64_isInf(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64IsInf(uint64(x)))
}

//export __bid64_isSignaling
func __bid64_isSignaling(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64IsSignaling(uint64(x)))
}

//export __bid64_isCanonical
func __bid64_isCanonical(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64IsCanonical(uint64(x)))
}

//export __bid64_isZero
func __bid64_isZero(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64IsZero(uint64(x)))
}

//export __bid64_isNormal
func __bid64_isNormal(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64IsNormal(uint64(x)))
}

//export __bid64_isSubnormal
func __bid64_isSubnormal(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64IsSubnormal(uint64(x)))
}

//export __bid64_copy
func __bid64_copy(x C.BID_UINT64) C.BID_UINT64 {
	return C.BID_UINT64(bidgo.Bid64Copy(uint64(x)))
}

//export __bid64_negate
func __bid64_negate(x C.BID_UINT64) C.BID_UINT64 {
	return C.BID_UINT64(bidgo.Bid64Negate(uint64(x)))
}

//export __bid64_abs
func __bid64_abs(x C.BID_UINT64) C.BID_UINT64 {
	return C.BID_UINT64(bidgo.Bid64Abs(uint64(x)))
}

//export __bid64_copySign
func __bid64_copySign(x, y C.BID_UINT64) C.BID_UINT64 {
	return C.BID_UINT64(bidgo.Bid64CopySign(uint64(x), uint64(y)))
}

//export __bid64_sameQuantum
func __bid64_sameQuantum(x, y C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64SameQuantum(uint64(x), uint64(y)))
}

//export __bid64_class
func __bid64_class(x C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64Class(uint64(x)))
}

//export __bid64_radix
func __bid64_radix() C.int {
	return C.int(bidgo.Bid64Radix())
}

//export __bid64_totalOrder
func __bid64_totalOrder(x, y C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64TotalOrder(uint64(x), uint64(y)))
}

//export __bid64_totalOrderMag
func __bid64_totalOrderMag(x, y C.BID_UINT64) C.int {
	return C.int(bidgo.Bid64TotalOrderMag(uint64(x), uint64(y)))
}

//export __bid64_quantum
func __bid64_quantum(x C.BID_UINT64) C.BID_UINT64 {
	return C.BID_UINT64(bidgo.Bid64Quantum(uint64(x)))
}

//export __bid64_quantexp
func __bid64_quantexp(x C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64Quantexp(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_llquantexp
func __bid64_llquantexp(x C.BID_UINT64) C.longlong {
	result, flags := bidgo.Bid64LLQuantexp(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_signaling_less
func __bid64_signaling_less(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64SignalingLess(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_minnum
func __bid64_minnum(x, y C.BID_UINT64) C.BID_UINT64 {
	result, flags := bidgo.Bid64MinNum(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_maxnum
func __bid64_maxnum(x, y C.BID_UINT64) C.BID_UINT64 {
	result, flags := bidgo.Bid64MaxNum(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_minnum_mag
func __bid64_minnum_mag(x, y C.BID_UINT64) C.BID_UINT64 {
	result, flags := bidgo.Bid64MinNumMag(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_maxnum_mag
func __bid64_maxnum_mag(x, y C.BID_UINT64) C.BID_UINT64 {
	result, flags := bidgo.Bid64MaxNumMag(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_quiet_equal
func __bid64_quiet_equal(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietEqual(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_not_equal
func __bid64_quiet_not_equal(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietNotEqual(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_less
func __bid64_quiet_less(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietLess(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_less_equal
func __bid64_quiet_less_equal(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietLessEqual(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_greater
func __bid64_quiet_greater(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietGreater(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_greater_equal
func __bid64_quiet_greater_equal(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietGreaterEqual(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_not_greater
func __bid64_quiet_not_greater(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietNotGreater(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_not_less
func __bid64_quiet_not_less(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietNotLess(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_ordered
func __bid64_quiet_ordered(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietOrdered(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_unordered
func __bid64_quiet_unordered(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietUnordered(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_greater_unordered
func __bid64_quiet_greater_unordered(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietGreaterUnordered(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_quiet_less_unordered
func __bid64_quiet_less_unordered(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64QuietLessUnordered(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_signaling_greater
func __bid64_signaling_greater(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64SignalingGreater(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_signaling_greater_equal
func __bid64_signaling_greater_equal(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64SignalingGreaterEqual(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_signaling_greater_unordered
func __bid64_signaling_greater_unordered(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64SignalingGreaterUnordered(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_signaling_less_equal
func __bid64_signaling_less_equal(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64SignalingLessEqual(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_signaling_less_unordered
func __bid64_signaling_less_unordered(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64SignalingLessUnordered(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_signaling_not_greater
func __bid64_signaling_not_greater(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64SignalingNotGreater(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

//export __bid64_signaling_not_less
func __bid64_signaling_not_less(x, y C.BID_UINT64) C.int {
	result, flags := bidgo.Bid64SignalingNotLess(uint64(x), uint64(y))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(result)
}

func bid64Llrint(x uint64, rndMode int) (int64, uint32) {
	return bidgo.Bid64Llrint(x, clampMode(rndMode))
}

func bid64Lrint(x uint64, rndMode int) (int64, uint32) {
	return bidgo.Bid64Lrint(x, clampMode(rndMode))
}

func bid64Llround(x uint64) (int64, uint32) {
	return bidgo.Bid64Llround(x)
}

func bid64Lround(x uint64) (int64, uint32) {
	return bidgo.Bid64Lround(x)
}

//export __bid64_llrint
func __bid64_llrint(x C.BID_UINT64) C.longlong {
	result, flags := bid64Llrint(uint64(x), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_lrint
func __bid64_lrint(x C.BID_UINT64) C.long {
	result, flags := bid64Lrint(uint64(x), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.long(result)
}

//export __bid64_llround
func __bid64_llround(x C.BID_UINT64) C.longlong {
	result, flags := bid64Llround(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_lround
func __bid64_lround(x C.BID_UINT64) C.long {
	result, flags := bid64Lround(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.long(result)
}

func bid64ToInt32Rnint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestEven, 32, false)
	return int64(int32(result)), flags
}

func bid64ToInt32Rninta(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestAway, 32, false)
	return int64(int32(result)), flags
}

func bid64ToInt32Floor(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvFloor, 32, false)
	return int64(int32(result)), flags
}

func bid64ToInt32Ceil(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvCeil, 32, false)
	return int64(int32(result)), flags
}

func bid64ToInt32Int(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvTrunc, 32, false)
	return int64(int32(result)), flags
}

//export __bid64_to_int32_rnint
func __bid64_to_int32_rnint(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Rnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

//export __bid64_to_int32_rninta
func __bid64_to_int32_rninta(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Rninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

//export __bid64_to_int32_floor
func __bid64_to_int32_floor(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Floor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

//export __bid64_to_int32_ceil
func __bid64_to_int32_ceil(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Ceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

//export __bid64_to_int32_int
func __bid64_to_int32_int(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Int(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

func bid64ToInt64Rnint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestEven, 64, false)
	return int64(result), flags
}

func bid64ToInt64Rninta(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestAway, 64, false)
	return int64(result), flags
}

func bid64ToInt64Floor(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvFloor, 64, false)
	return int64(result), flags
}

func bid64ToInt64Ceil(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvCeil, 64, false)
	return int64(result), flags
}

func bid64ToInt64Int(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvTrunc, 64, false)
	return int64(result), flags
}

//export __bid64_to_int64_rnint
func __bid64_to_int64_rnint(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Rnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_to_int64_rninta
func __bid64_to_int64_rninta(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Rninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_to_int64_floor
func __bid64_to_int64_floor(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Floor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_to_int64_ceil
func __bid64_to_int64_ceil(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Ceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_to_int64_int
func __bid64_to_int64_int(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Int(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

func bid64ToUint32Rnint(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestEven, 32, false)
	return uint64(uint32(result)), flags
}

func bid64ToUint32Rninta(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestAway, 32, false)
	return uint64(uint32(result)), flags
}

func bid64ToUint32Floor(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvFloor, 32, false)
	return uint64(uint32(result)), flags
}

func bid64ToUint32Ceil(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvCeil, 32, false)
	return uint64(uint32(result)), flags
}

func bid64ToUint32Int(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvTrunc, 32, false)
	return uint64(uint32(result)), flags
}

//export __bid64_to_uint32_rnint
func __bid64_to_uint32_rnint(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Rnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

//export __bid64_to_uint32_rninta
func __bid64_to_uint32_rninta(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Rninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

//export __bid64_to_uint32_floor
func __bid64_to_uint32_floor(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Floor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

//export __bid64_to_uint32_ceil
func __bid64_to_uint32_ceil(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Ceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

//export __bid64_to_uint32_int
func __bid64_to_uint32_int(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Int(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

func bid64ToUint64Rnint(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvNearestEven, 64, false)
}

func bid64ToUint64Rninta(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvNearestAway, 64, false)
}

func bid64ToUint64Floor(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvFloor, 64, false)
}

func bid64ToUint64Ceil(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvCeil, 64, false)
}

func bid64ToUint64Int(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvTrunc, 64, false)
}

//export __bid64_to_uint64_rnint
func __bid64_to_uint64_rnint(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Rnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_to_uint64_rninta
func __bid64_to_uint64_rninta(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Rninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_to_uint64_floor
func __bid64_to_uint64_floor(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Floor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_to_uint64_ceil
func __bid64_to_uint64_ceil(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Ceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_to_uint64_int
func __bid64_to_uint64_int(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Int(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

func bid64ToInt32Xrnint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestEven, 32, true)
	return int64(int32(result)), flags
}

func bid64ToInt32Xrninta(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestAway, 32, true)
	return int64(int32(result)), flags
}

func bid64ToInt32Xfloor(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvFloor, 32, true)
	return int64(int32(result)), flags
}

func bid64ToInt32Xceil(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvCeil, 32, true)
	return int64(int32(result)), flags
}

func bid64ToInt32Xint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvTrunc, 32, true)
	return int64(int32(result)), flags
}

//export __bid64_to_int32_xrnint
func __bid64_to_int32_xrnint(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Xrnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

//export __bid64_to_int32_xrninta
func __bid64_to_int32_xrninta(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Xrninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

//export __bid64_to_int32_xfloor
func __bid64_to_int32_xfloor(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Xfloor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

//export __bid64_to_int32_xceil
func __bid64_to_int32_xceil(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Xceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

//export __bid64_to_int32_xint
func __bid64_to_int32_xint(x C.BID_UINT64) C.int {
	result, flags := bid64ToInt32Xint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(int32(result))
}

func bid64ToInt64Xrnint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestEven, 64, true)
	return int64(result), flags
}

func bid64ToInt64Xrninta(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestAway, 64, true)
	return int64(result), flags
}

func bid64ToInt64Xfloor(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvFloor, 64, true)
	return int64(result), flags
}

func bid64ToInt64Xceil(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvCeil, 64, true)
	return int64(result), flags
}

func bid64ToInt64Xint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvTrunc, 64, true)
	return int64(result), flags
}

//export __bid64_to_int64_xrnint
func __bid64_to_int64_xrnint(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Xrnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_to_int64_xrninta
func __bid64_to_int64_xrninta(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Xrninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_to_int64_xfloor
func __bid64_to_int64_xfloor(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Xfloor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_to_int64_xceil
func __bid64_to_int64_xceil(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Xceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

//export __bid64_to_int64_xint
func __bid64_to_int64_xint(x C.BID_UINT64) C.longlong {
	result, flags := bid64ToInt64Xint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.longlong(result)
}

func bid64ToUint32Xrnint(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestEven, 32, true)
	return uint64(uint32(result)), flags
}

func bid64ToUint32Xrninta(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestAway, 32, true)
	return uint64(uint32(result)), flags
}

func bid64ToUint32Xfloor(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvFloor, 32, true)
	return uint64(uint32(result)), flags
}

func bid64ToUint32Xceil(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvCeil, 32, true)
	return uint64(uint32(result)), flags
}

func bid64ToUint32Xint(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvTrunc, 32, true)
	return uint64(uint32(result)), flags
}

//export __bid64_to_uint32_xrnint
func __bid64_to_uint32_xrnint(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Xrnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

//export __bid64_to_uint32_xrninta
func __bid64_to_uint32_xrninta(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Xrninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

//export __bid64_to_uint32_xfloor
func __bid64_to_uint32_xfloor(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Xfloor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

//export __bid64_to_uint32_xceil
func __bid64_to_uint32_xceil(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Xceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

//export __bid64_to_uint32_xint
func __bid64_to_uint32_xint(x C.BID_UINT64) C.uint {
	result, flags := bid64ToUint32Xint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uint(uint32(result))
}

func bid64ToUint64Xrnint(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvNearestEven, 64, true)
}

func bid64ToUint64Xrninta(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvNearestAway, 64, true)
}

func bid64ToUint64Xfloor(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvFloor, 64, true)
}

func bid64ToUint64Xceil(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvCeil, 64, true)
}

func bid64ToUint64Xint(x uint64) (uint64, uint32) {
	return bid64ToUnsignedFixed(x, intConvTrunc, 64, true)
}

//export __bid64_to_uint64_xrnint
func __bid64_to_uint64_xrnint(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Xrnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_to_uint64_xrninta
func __bid64_to_uint64_xrninta(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Xrninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_to_uint64_xfloor
func __bid64_to_uint64_xfloor(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Xfloor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_to_uint64_xceil
func __bid64_to_uint64_xceil(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Xceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

//export __bid64_to_uint64_xint
func __bid64_to_uint64_xint(x C.BID_UINT64) C.BID_UINT64 {
	result, flags := bid64ToUint64Xint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(result)
}

func bid64ToInt8Rnint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestEven, 8, false)
	return int64(int8(result)), flags
}

func bid64ToInt8Rninta(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestAway, 8, false)
	return int64(int8(result)), flags
}

func bid64ToInt8Floor(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvFloor, 8, false)
	return int64(int8(result)), flags
}

func bid64ToInt8Ceil(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvCeil, 8, false)
	return int64(int8(result)), flags
}

func bid64ToInt8Int(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvTrunc, 8, false)
	return int64(int8(result)), flags
}

func bid64ToInt8Xrnint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestEven, 8, true)
	return int64(int8(result)), flags
}

func bid64ToInt8Xrninta(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestAway, 8, true)
	return int64(int8(result)), flags
}

func bid64ToInt8Xfloor(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvFloor, 8, true)
	return int64(int8(result)), flags
}

func bid64ToInt8Xceil(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvCeil, 8, true)
	return int64(int8(result)), flags
}

func bid64ToInt8Xint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvTrunc, 8, true)
	return int64(int8(result)), flags
}

//export __bid64_to_int8_rnint
func __bid64_to_int8_rnint(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Rnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

//export __bid64_to_int8_rninta
func __bid64_to_int8_rninta(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Rninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

//export __bid64_to_int8_floor
func __bid64_to_int8_floor(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Floor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

//export __bid64_to_int8_ceil
func __bid64_to_int8_ceil(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Ceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

//export __bid64_to_int8_int
func __bid64_to_int8_int(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Int(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

//export __bid64_to_int8_xrnint
func __bid64_to_int8_xrnint(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Xrnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

//export __bid64_to_int8_xrninta
func __bid64_to_int8_xrninta(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Xrninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

//export __bid64_to_int8_xfloor
func __bid64_to_int8_xfloor(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Xfloor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

//export __bid64_to_int8_xceil
func __bid64_to_int8_xceil(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Xceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

//export __bid64_to_int8_xint
func __bid64_to_int8_xint(x C.BID_UINT64) C.schar {
	result, flags := bid64ToInt8Xint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.schar(int8(result))
}

func bid64ToInt16Rnint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestEven, 16, false)
	return int64(int16(result)), flags
}

func bid64ToInt16Rninta(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestAway, 16, false)
	return int64(int16(result)), flags
}

func bid64ToInt16Floor(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvFloor, 16, false)
	return int64(int16(result)), flags
}

func bid64ToInt16Ceil(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvCeil, 16, false)
	return int64(int16(result)), flags
}

func bid64ToInt16Int(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvTrunc, 16, false)
	return int64(int16(result)), flags
}

func bid64ToInt16Xrnint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestEven, 16, true)
	return int64(int16(result)), flags
}

func bid64ToInt16Xrninta(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvNearestAway, 16, true)
	return int64(int16(result)), flags
}

func bid64ToInt16Xfloor(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvFloor, 16, true)
	return int64(int16(result)), flags
}

func bid64ToInt16Xceil(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvCeil, 16, true)
	return int64(int16(result)), flags
}

func bid64ToInt16Xint(x uint64) (int64, uint32) {
	result, flags := bid64ToSignedFixed(x, intConvTrunc, 16, true)
	return int64(int16(result)), flags
}

//export __bid64_to_int16_rnint
func __bid64_to_int16_rnint(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Rnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

//export __bid64_to_int16_rninta
func __bid64_to_int16_rninta(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Rninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

//export __bid64_to_int16_floor
func __bid64_to_int16_floor(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Floor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

//export __bid64_to_int16_ceil
func __bid64_to_int16_ceil(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Ceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

//export __bid64_to_int16_int
func __bid64_to_int16_int(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Int(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

//export __bid64_to_int16_xrnint
func __bid64_to_int16_xrnint(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Xrnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

//export __bid64_to_int16_xrninta
func __bid64_to_int16_xrninta(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Xrninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

//export __bid64_to_int16_xfloor
func __bid64_to_int16_xfloor(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Xfloor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

//export __bid64_to_int16_xceil
func __bid64_to_int16_xceil(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Xceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

//export __bid64_to_int16_xint
func __bid64_to_int16_xint(x C.BID_UINT64) C.short {
	result, flags := bid64ToInt16Xint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.short(int16(result))
}

func bid64ToUint8Rnint(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestEven, 8, false)
	return uint64(uint8(result)), flags
}

func bid64ToUint8Rninta(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestAway, 8, false)
	return uint64(uint8(result)), flags
}

func bid64ToUint8Floor(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvFloor, 8, false)
	return uint64(uint8(result)), flags
}

func bid64ToUint8Ceil(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvCeil, 8, false)
	return uint64(uint8(result)), flags
}

func bid64ToUint8Int(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvTrunc, 8, false)
	return uint64(uint8(result)), flags
}

func bid64ToUint8Xrnint(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestEven, 8, true)
	return uint64(uint8(result)), flags
}

func bid64ToUint8Xrninta(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestAway, 8, true)
	return uint64(uint8(result)), flags
}

func bid64ToUint8Xfloor(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvFloor, 8, true)
	return uint64(uint8(result)), flags
}

func bid64ToUint8Xceil(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvCeil, 8, true)
	return uint64(uint8(result)), flags
}

func bid64ToUint8Xint(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvTrunc, 8, true)
	return uint64(uint8(result)), flags
}

//export __bid64_to_uint8_rnint
func __bid64_to_uint8_rnint(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Rnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

//export __bid64_to_uint8_rninta
func __bid64_to_uint8_rninta(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Rninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

//export __bid64_to_uint8_floor
func __bid64_to_uint8_floor(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Floor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

//export __bid64_to_uint8_ceil
func __bid64_to_uint8_ceil(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Ceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

//export __bid64_to_uint8_int
func __bid64_to_uint8_int(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Int(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

//export __bid64_to_uint8_xrnint
func __bid64_to_uint8_xrnint(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Xrnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

//export __bid64_to_uint8_xrninta
func __bid64_to_uint8_xrninta(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Xrninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

//export __bid64_to_uint8_xfloor
func __bid64_to_uint8_xfloor(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Xfloor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

//export __bid64_to_uint8_xceil
func __bid64_to_uint8_xceil(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Xceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

//export __bid64_to_uint8_xint
func __bid64_to_uint8_xint(x C.BID_UINT64) C.uchar {
	result, flags := bid64ToUint8Xint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.uchar(uint8(result))
}

func bid64ToUint16Rnint(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestEven, 16, false)
	return uint64(uint16(result)), flags
}

func bid64ToUint16Rninta(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestAway, 16, false)
	return uint64(uint16(result)), flags
}

func bid64ToUint16Floor(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvFloor, 16, false)
	return uint64(uint16(result)), flags
}

func bid64ToUint16Ceil(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvCeil, 16, false)
	return uint64(uint16(result)), flags
}

func bid64ToUint16Int(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvTrunc, 16, false)
	return uint64(uint16(result)), flags
}

func bid64ToUint16Xrnint(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestEven, 16, true)
	return uint64(uint16(result)), flags
}

func bid64ToUint16Xrninta(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvNearestAway, 16, true)
	return uint64(uint16(result)), flags
}

func bid64ToUint16Xfloor(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvFloor, 16, true)
	return uint64(uint16(result)), flags
}

func bid64ToUint16Xceil(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvCeil, 16, true)
	return uint64(uint16(result)), flags
}

func bid64ToUint16Xint(x uint64) (uint64, uint32) {
	result, flags := bid64ToUnsignedFixed(x, intConvTrunc, 16, true)
	return uint64(uint16(result)), flags
}

//export __bid64_to_uint16_rnint
func __bid64_to_uint16_rnint(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Rnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_to_uint16_rninta
func __bid64_to_uint16_rninta(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Rninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_to_uint16_floor
func __bid64_to_uint16_floor(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Floor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_to_uint16_ceil
func __bid64_to_uint16_ceil(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Ceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_to_uint16_int
func __bid64_to_uint16_int(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Int(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_to_uint16_xrnint
func __bid64_to_uint16_xrnint(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Xrnint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_to_uint16_xrninta
func __bid64_to_uint16_xrninta(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Xrninta(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_to_uint16_xfloor
func __bid64_to_uint16_xfloor(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Xfloor(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_to_uint16_xceil
func __bid64_to_uint16_xceil(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Xceil(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_to_uint16_xint
func __bid64_to_uint16_xint(x C.BID_UINT64) C.ushort {
	result, flags := bid64ToUint16Xint(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.ushort(uint16(result))
}

//export __bid64_round_integral_nearest_even
func __bid64_round_integral_nearest_even(x C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64RoundIntegralNearestEven(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_round_integral_nearest_away
func __bid64_round_integral_nearest_away(x C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64RoundIntegralNearestAway(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_round_integral_negative
func __bid64_round_integral_negative(x C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64RoundIntegralNegative(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_round_integral_positive
func __bid64_round_integral_positive(x C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64RoundIntegralPositive(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_round_integral_zero
func __bid64_round_integral_zero(x C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64RoundIntegralZero(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_round_integral_exact
func __bid64_round_integral_exact(x C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64RoundIntegralExact(uint64(x), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_nearbyint
func __bid64_nearbyint(x C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64NearbyInt(uint64(x), int(C.get_rnd_mode()))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

//export __bid64_ilogb
func __bid64_ilogb(x C.BID_UINT64) C.int {
	res, flags := bidgo.Bid64ILogb(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.int(res)
}

//export __bid64_logb
func __bid64_logb(x C.BID_UINT64) C.BID_UINT64 {
	res, flags := bidgo.Bid64Logb(uint64(x))
	if flags != 0 {
		C.or_flags(C._IDEC_flags(flags))
	}
	return C.BID_UINT64(res)
}

func main() {}
