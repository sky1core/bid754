package testgen

import (
	"encoding/binary"
	"math/big"
)

type bidCodecRefKind uint8

const (
	bidCodecRefNormal bidCodecRefKind = iota
	bidCodecRefZero
	bidCodecRefInfinity
	bidCodecRefQNaN
	bidCodecRefSNaN
)

type bidCodecRefComponents struct {
	Sign        bool
	Coefficient *big.Int
	Exponent    int32
	Kind        bidCodecRefKind
	Payload     uint64
}

const (
	bid32RefNaNMask   = 0x7c000000
	bid32RefSNaNMask  = 0x7e000000
	bid32RefInfMask   = 0x78000000
	bid32RefSignMask  = 0x80000000
	bid32RefSteerMask = 0x60000000
	bid32RefExpMask   = 0xff
	bid32RefMaxCoeff  = 9999999
	bid32RefBias      = 101

	bid64RefNaNMask   = 0x7c00000000000000
	bid64RefSNaNMask  = 0x7e00000000000000
	bid64RefInfMask   = 0x7800000000000000
	bid64RefSignMask  = 0x8000000000000000
	bid64RefSteerMask = 0x6000000000000000
	bid64RefExpMask   = 0x3ff
	bid64RefMaxCoeff  = 9999999999999999
	bid64RefBias      = 398

	bid128RefNaNMask   = 0x7c00000000000000
	bid128RefSNaNMask  = 0x7e00000000000000
	bid128RefInfMask   = 0x7800000000000000
	bid128RefSignMask  = 0x8000000000000000
	bid128RefSteerMask = 0x6000000000000000
	bid128RefExpMask   = 0x3fff
	bid128RefBias      = 6176
)

var (
	bid128RefTen33 = mustBigIntDecimal("1000000000000000000000000000000000")
	bid128RefTen34 = mustBigIntDecimal("10000000000000000000000000000000000")
)

func mustBigIntDecimal(s string) *big.Int {
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic("invalid decimal constant: " + s)
	}
	return v
}

func refDecode32(v uint32) bidCodecRefComponents {
	sign := v&bid32RefSignMask != 0
	if v&bid32RefNaNMask == bid32RefNaNMask {
		kind := bidCodecRefQNaN
		if v&bid32RefSNaNMask == bid32RefSNaNMask {
			kind = bidCodecRefSNaN
		}
		payload := uint64(v & 0x000fffff)
		if payload > 999999 {
			payload = 0
		}
		return bidCodecRefComponents{Sign: sign, Kind: kind, Payload: payload}
	}
	if v&bid32RefInfMask == bid32RefInfMask {
		return bidCodecRefComponents{Sign: sign, Kind: bidCodecRefInfinity}
	}

	var exp int
	var coeff uint32
	if v&bid32RefSteerMask == bid32RefSteerMask {
		exp = int((v >> 21) & bid32RefExpMask)
		coeff = (v & 0x001fffff) | 0x00800000
		if coeff > bid32RefMaxCoeff {
			coeff = 0
		}
	} else {
		exp = int((v >> 23) & bid32RefExpMask)
		coeff = v & 0x007fffff
	}
	if coeff == 0 {
		return bidCodecRefComponents{Sign: sign, Exponent: int32(exp - bid32RefBias), Kind: bidCodecRefZero}
	}
	return bidCodecRefComponents{
		Sign:        sign,
		Coefficient: new(big.Int).SetUint64(uint64(coeff)),
		Exponent:    int32(exp - bid32RefBias),
		Kind:        bidCodecRefNormal,
	}
}

func refEncode32(c bidCodecRefComponents) uint32 {
	var sign uint32
	if c.Sign {
		sign = bid32RefSignMask
	}
	switch c.Kind {
	case bidCodecRefInfinity:
		return sign | bid32RefInfMask
	case bidCodecRefQNaN:
		return sign | bid32RefNaNMask | (uint32(c.Payload) & 0x000fffff)
	case bidCodecRefSNaN:
		return sign | bid32RefSNaNMask | (uint32(c.Payload) & 0x000fffff)
	case bidCodecRefZero:
		exp := clampInt(int(c.Exponent)+bid32RefBias, 0, 191)
		return sign | (uint32(exp) << 23)
	}

	coeff := uint32(c.Coefficient.Uint64())
	exp := clampInt(int(c.Exponent)+bid32RefBias, 0, 191)
	if coeff < 0x800000 {
		return sign | (uint32(exp) << 23) | coeff
	}
	return sign | bid32RefSteerMask | (uint32(exp) << 21) | (coeff & 0x001fffff)
}

func refDecode64(v uint64) bidCodecRefComponents {
	sign := v&bid64RefSignMask != 0
	if v&bid64RefNaNMask == bid64RefNaNMask {
		kind := bidCodecRefQNaN
		if v&bid64RefSNaNMask == bid64RefSNaNMask {
			kind = bidCodecRefSNaN
		}
		payload := v & 0x0003ffffffffffff
		if payload > 999999999999999 {
			payload = 0
		}
		return bidCodecRefComponents{Sign: sign, Kind: kind, Payload: payload}
	}
	if v&bid64RefInfMask == bid64RefInfMask {
		return bidCodecRefComponents{Sign: sign, Kind: bidCodecRefInfinity}
	}

	var exp int
	var coeff uint64
	if v&bid64RefSteerMask == bid64RefSteerMask {
		exp = int((v >> 51) & bid64RefExpMask)
		coeff = (v & 0x0007ffffffffffff) | 0x0020000000000000
		if coeff > bid64RefMaxCoeff {
			coeff = 0
		}
	} else {
		exp = int((v >> 53) & bid64RefExpMask)
		coeff = v & 0x001fffffffffffff
	}
	if coeff == 0 {
		return bidCodecRefComponents{Sign: sign, Exponent: int32(exp - bid64RefBias), Kind: bidCodecRefZero}
	}
	return bidCodecRefComponents{
		Sign:        sign,
		Coefficient: new(big.Int).SetUint64(coeff),
		Exponent:    int32(exp - bid64RefBias),
		Kind:        bidCodecRefNormal,
	}
}

func refEncode64(c bidCodecRefComponents) uint64 {
	var sign uint64
	if c.Sign {
		sign = bid64RefSignMask
	}
	switch c.Kind {
	case bidCodecRefInfinity:
		return sign | bid64RefInfMask
	case bidCodecRefQNaN:
		return sign | bid64RefNaNMask | (c.Payload & 0x0003ffffffffffff)
	case bidCodecRefSNaN:
		return sign | bid64RefSNaNMask | (c.Payload & 0x0003ffffffffffff)
	case bidCodecRefZero:
		exp := clampInt(int(c.Exponent)+bid64RefBias, 0, 767)
		return sign | (uint64(exp) << 53)
	}

	coeff := c.Coefficient.Uint64()
	exp := clampInt(int(c.Exponent)+bid64RefBias, 0, 767)
	if coeff < 0x20000000000000 {
		return sign | (uint64(exp) << 53) | coeff
	}
	return sign | bid64RefSteerMask | (uint64(exp) << 51) | (coeff & 0x0007ffffffffffff)
}

func refDecode128(lo, hi uint64) bidCodecRefComponents {
	sign := hi&bid128RefSignMask != 0
	if hi&bid128RefNaNMask == bid128RefNaNMask {
		kind := bidCodecRefQNaN
		if hi&bid128RefSNaNMask == bid128RefSNaNMask {
			kind = bidCodecRefSNaN
		}
		payloadHi := hi & 0x00003fffffffffff
		payload := new(big.Int).SetUint64(payloadHi)
		payload.Lsh(payload, 64)
		payload.Or(payload, new(big.Int).SetUint64(lo))
		if payload.Cmp(bid128RefTen33) >= 0 {
			return bidCodecRefComponents{Sign: sign, Kind: kind, Payload: 0}
		}
		return bidCodecRefComponents{Sign: sign, Kind: kind, Payload: lo}
	}
	if hi&bid128RefInfMask == bid128RefInfMask {
		return bidCodecRefComponents{Sign: sign, Kind: bidCodecRefInfinity}
	}

	var exp int
	var coeffHi uint64
	if hi&bid128RefSteerMask == bid128RefSteerMask {
		exp = int((hi >> 47) & bid128RefExpMask)
		coeffHi = (hi & 0x00007fffffffffff) | 0x0020000000000000
	} else {
		exp = int((hi >> 49) & bid128RefExpMask)
		coeffHi = hi & 0x0001ffffffffffff
	}

	coeff := new(big.Int).SetUint64(coeffHi)
	coeff.Lsh(coeff, 64)
	coeff.Or(coeff, new(big.Int).SetUint64(lo))
	if coeff.Cmp(bid128RefTen34) >= 0 {
		coeff.SetUint64(0)
	}
	if coeff.Sign() == 0 {
		return bidCodecRefComponents{Sign: sign, Exponent: int32(exp - bid128RefBias), Kind: bidCodecRefZero}
	}
	return bidCodecRefComponents{
		Sign:        sign,
		Coefficient: coeff,
		Exponent:    int32(exp - bid128RefBias),
		Kind:        bidCodecRefNormal,
	}
}

func refEncode128(c bidCodecRefComponents) (lo, hi uint64) {
	var sign uint64
	if c.Sign {
		sign = bid128RefSignMask
	}
	switch c.Kind {
	case bidCodecRefInfinity:
		return 0, sign | bid128RefInfMask
	case bidCodecRefQNaN:
		return c.Payload, sign | bid128RefNaNMask
	case bidCodecRefSNaN:
		return c.Payload, sign | bid128RefSNaNMask
	case bidCodecRefZero:
		exp := clampInt(int(c.Exponent)+bid128RefBias, 0, 12287)
		return 0, sign | (uint64(exp) << 49)
	}

	var coeffBytes [16]byte
	c.Coefficient.FillBytes(coeffBytes[:])
	coeffHi := binary.BigEndian.Uint64(coeffBytes[0:8])
	coeffLo := binary.BigEndian.Uint64(coeffBytes[8:16])
	exp := clampInt(int(c.Exponent)+bid128RefBias, 0, 12287)
	return coeffLo, sign | (uint64(exp) << 49) | (coeffHi & 0x0001ffffffffffff)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
