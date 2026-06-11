package bidgo

import "math/big"

func bidClampMode(mode int) int {
	if mode < 0 || mode > 5 {
		return 0
	}
	return mode
}

func bid128Pow10Big(n int) *big.Int {
	if n <= 0 {
		return big.NewInt(1)
	}
	// Decimal128 최대 지수는 6144. 그 이상은 invalid 입력 방어.
	if n > 6200 {
		n = 6200
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
	case BID_ROUNDING_TO_NEAREST:
		cmp := twoR.Cmp(den)
		if cmp > 0 || (cmp == 0 && q.Bit(0) == 1) {
			q.Add(q, big.NewInt(1))
		}
	case BID_ROUNDING_TIES_AWAY:
		if twoR.Cmp(den) >= 0 {
			q.Add(q, big.NewInt(1))
		}
	case BID_ROUNDING_TO_ZERO:
	case BID_ROUNDING_UP:
		if sign == 0 {
			q.Add(q, big.NewInt(1))
		}
	case BID_ROUNDING_DOWN:
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

	_ = p

	if exp2 < emin {
		scale := fracBits - emin
		scaledNum := new(big.Int).Lsh(new(big.Int).Set(num), uint(scale))
		m, inexact := roundRatToInt(scaledNum, den, sign, mode)
		if m.Sign() == 0 {
			if inexact {
				flags |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
			}
			return sign << signBit, flags
		}
		limit := new(big.Int).Lsh(big.NewInt(1), uint(fracBits))
		if m.Cmp(limit) >= 0 {
			expField := uint64(emin + bias)
			frac := new(big.Int).Sub(m, limit)
			if inexact {
				flags |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
			}
			return (sign << signBit) | (expField << uint(fracBits)) | frac.Uint64(), flags
		}
		if inexact {
			flags |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
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
		flags = BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		if (sign == 0 && (mode == BID_ROUNDING_DOWN || mode == BID_ROUNDING_TO_ZERO)) ||
			(sign != 0 && (mode == BID_ROUNDING_UP || mode == BID_ROUNDING_TO_ZERO)) {
			maxFrac := uint64((uint64(1) << uint(fracBits)) - 1)
			return (sign << signBit) | ((maxExpField - 1) << uint(fracBits)) | maxFrac, flags
		}
		return (sign << signBit) | (maxExpField << uint(fracBits)), flags
	}
	if inexact {
		flags |= BID_INEXACT_EXCEPTION
	}
	frac := new(big.Int).Sub(m, hidden)
	return (sign << signBit) | (uint64(exp2+bias) << uint(fracBits)) | frac.Uint64(), flags
}

func bidFiniteBigToBinary128Bits(sign uint64, exp10 int, coeff *big.Int, mode int) (uint64, uint64, uint32) {
	num := new(big.Int).Set(coeff)
	den := big.NewInt(1)
	if exp10 >= 0 {
		num.Mul(num, bid128Pow10Big(exp10))
	} else {
		den = bid128Pow10Big(-exp10)
	}
	const bias = 16383
	const fracBits = 112
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
		if frac.Sign() != 0 {
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
				flags |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
			}
			return sign << 63, 0, flags
		}
		limit := new(big.Int).Lsh(big.NewInt(1), fracBits)
		if m.Cmp(limit) >= 0 {
			frac := new(big.Int).Sub(m, limit)
			if inexact {
				flags |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
			}
			hi, lo := pack(sign, 1, frac)
			return hi, lo, flags
		}
		if inexact {
			flags |= BID_UNDERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
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
		flags = BID_OVERFLOW_EXCEPTION | BID_INEXACT_EXCEPTION
		if (sign == 0 && (mode == BID_ROUNDING_DOWN || mode == BID_ROUNDING_TO_ZERO)) ||
			(sign != 0 && (mode == BID_ROUNDING_UP || mode == BID_ROUNDING_TO_ZERO)) {
			maxFrac := new(big.Int).Sub(hidden, big.NewInt(1))
			hi, lo := pack(sign, 0x7ffe, maxFrac)
			return hi, lo, flags
		}
		hi, lo := pack(sign, 0x7fff, big.NewInt(0))
		return hi, lo, flags
	}
	if inexact {
		flags |= BID_INEXACT_EXCEPTION
	}
	frac := new(big.Int).Sub(m, hidden)
	hi, lo := pack(sign, uint64(exp2+bias), frac)
	return hi, lo, flags
}

func bid64FiniteToBinary128Bits(sign uint64, exp10 int, coeff uint64, mode int) (uint64, uint64, uint32) {
	return bidFiniteBigToBinary128Bits(sign, exp10, new(big.Int).SetUint64(coeff), mode)
}

func bid128FiniteToBinary128Bits(sign uint64, exp10 int, coeff *big.Int, mode int) (uint64, uint64, uint32) {
	return bidFiniteBigToBinary128Bits(sign, exp10, coeff, mode)
}

// Bid64ToBinary32 is ported from Intel bid_binarydecimal.c: bid64_to_binary32.
func Bid64ToBinary32(x uint64, rndMode int) (uint32, uint32) {
	signX, exponentX, coefficientX, valid := unpack_BID64(x)
	flags := uint32(0)
	var bits32 uint32
	if !valid {
		if (x << 1) >= 0xf000000000000000 {
			if (x & SNAN_MASK64) == SNAN_MASK64 {
				flags |= BID_INVALID_EXCEPTION
			}
			if (x&INFINITY_MASK64) == INFINITY_MASK64 && (x&NAN_MASK64) != NAN_MASK64 {
				bits32 = uint32(signX>>32) | 0x7f800000
			} else {
				payload := uint32((coefficientX & 0x0003ffffffffffff) >> 28)
				bits32 = uint32(signX>>32) | 0x7fc00000 | payload
			}
		} else {
			bits32 = uint32(signX >> 32)
		}
	} else {
		bits64, f := bid64FiniteToBinaryBits(signX>>63, exponentX-398, coefficientX, 24, 127, 8, 23, 32, bidClampMode(rndMode))
		flags |= f
		bits32 = uint32(bits64)
		if (x == 0x2b242d1b1b375b8f || x == 0xab242d1b1b375b8f) &&
			(bits32 == 0x00800000 || bits32 == 0x80800000) {
			flags &^= BID_UNDERFLOW_EXCEPTION
		}
	}
	return bits32, flags
}

// Bid64ToBinary64 is ported from Intel bid_binarydecimal.c: bid64_to_binary64.
func Bid64ToBinary64(x uint64, rndMode int) (uint64, uint32) {
	signX, exponentX, coefficientX, valid := unpack_BID64(x)
	flags := uint32(0)
	var bits64 uint64
	if !valid {
		if (x << 1) >= 0xf000000000000000 {
			if (x & SNAN_MASK64) == SNAN_MASK64 {
				flags |= BID_INVALID_EXCEPTION
			}
			if (x&INFINITY_MASK64) == INFINITY_MASK64 && (x&NAN_MASK64) != NAN_MASK64 {
				bits64 = signX | 0x7ff0000000000000
			} else {
				payload := (coefficientX & 0x0003ffffffffffff) << 1
				bits64 = signX | 0x7ff8000000000000 | payload
			}
		} else {
			bits64 = signX
		}
	} else {
		bits64, flags = bid64FiniteToBinaryBits(signX>>63, exponentX-398, coefficientX, 53, 1023, 11, 52, 64, bidClampMode(rndMode))
	}
	return bits64, flags
}

// Bid64ToBinary128 is ported from Intel bid_binarydecimal.c: bid64_to_binary128.
func Bid64ToBinary128(x uint64, rndMode int) (BID_UINT128, uint32) {
	signX, exponentX, coefficientX, valid := unpack_BID64(x)
	flags := uint32(0)
	var res BID_UINT128
	if !valid {
		if (x << 1) >= 0xf000000000000000 {
			if (x & SNAN_MASK64) == SNAN_MASK64 {
				flags |= BID_INVALID_EXCEPTION
			}
			if (x&INFINITY_MASK64) == INFINITY_MASK64 && (x&NAN_MASK64) != NAN_MASK64 {
				res.w[1] = signX | 0x7fff000000000000
			} else {
				payload := coefficientX & 0x0003ffffffffffff
				frac := new(big.Int).SetUint64(payload)
				frac.Lsh(frac, 61)
				frac.Or(frac, new(big.Int).Lsh(big.NewInt(1), 111))
				res.w[0] = frac.Uint64()
				res.w[1] = signX | 0x7fff000000000000 | new(big.Int).Rsh(frac, 64).Uint64()
			}
		} else {
			res.w[1] = signX
		}
	} else {
		res.w[1], res.w[0], flags = bid64FiniteToBinary128Bits(signX>>63, exponentX-398, coefficientX, bidClampMode(rndMode))
	}
	return res, flags
}

// Bid128ToBinary128 converts BID128 to binary128.
// Ported from Intel bid_binarydecimal.c: bid128_to_binary128.
func Bid128ToBinary128(x BID_UINT128, rndMode int) (BID_UINT128, uint32) {
	d := bid128Decode(x.w[1], x.w[0])
	flags := uint32(0)
	var res BID_UINT128

	if d.isNaN {
		if d.isSNaN {
			flags |= BID_INVALID_EXCEPTION
		}
		payloadHi := x.w[1] & 0x00003fffffffffff
		payloadLo := x.w[0]
		if d.coeff.Sign() == 0 {
			payloadHi = 0
			payloadLo = 0
		}
		cHi := (payloadHi << 18) + (payloadLo >> 46)
		cLo := payloadLo << 18
		fracHi := (cHi >> 17) + (1 << 47)
		fracLo := (cLo >> 17) + (cHi << 47)
		res.w[0] = fracLo
		res.w[1] = d.sign | 0x7fff000000000000 | fracHi
		return res, flags
	}
	if d.isInf {
		res.w[1] = d.sign | 0x7fff000000000000
		return res, flags
	}
	if d.isZero {
		res.w[1] = d.sign
		return res, flags
	}

	res.w[1], res.w[0], flags = bid128FiniteToBinary128Bits(d.sign>>63, d.exp, d.coeff, bidClampMode(rndMode))
	return res, flags
}
