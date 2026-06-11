// Package bidcodec provides BID (Binary Integer Decimal) encoding/decoding
// for IEEE 754 decimal floating-point interchange between languages.
//
// This package extracts {sign, coefficient, exponent} components from
// BID32/64/128 encoded bytes, enabling conversion to any language's
// native decimal library (BigDecimal, rust_decimal, decimal.Decimal, etc).
package bidcodec

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

// Kind classifies a decimal value.
type Kind uint8

const (
	Normal   Kind = iota // Finite non-zero number
	Zero                 // Positive or negative zero
	Infinity             // Positive or negative infinity
	QNaN                 // Quiet NaN
	SNaN                 // Signaling NaN
)

// Components holds the decomposed parts of a BID-encoded decimal.
// Any decimal value can be reconstructed from these fields:
//
//	value = (-1)^Sign * Coefficient * 10^Exponent
//
// For special values (Infinity, NaN), Coefficient and Exponent are zero,
// and NaN payload is stored in Payload.
type Components struct {
	Sign        bool     // true = negative
	Coefficient *big.Int // unsigned integer (nil for Infinity/NaN)
	Exponent    int32    // power of 10
	Kind        Kind     // Normal, Zero, Infinity, QNaN, SNaN
	Payload     uint64   // NaN payload (only meaningful for QNaN/SNaN)
}

// --- BID32 ---

const (
	bid32NaNMask   = 0x7c000000
	bid32SNaNMask  = 0x7e000000
	bid32InfMask   = 0x78000000
	bid32SignMask  = 0x80000000
	bid32SteerMask = 0x60000000
	bid32ExpMask32 = 0xff
	bid32MaxCoeff  = 9999999
	bid32Bias      = 101
)

// Decode32 extracts components from a BID32-encoded uint32.
func Decode32(v uint32) Components {
	sign := v&bid32SignMask != 0

	// NaN
	if v&bid32NaNMask == bid32NaNMask {
		kind := QNaN
		if v&bid32SNaNMask == bid32SNaNMask {
			kind = SNaN
		}
		payload := uint64(v & 0x000fffff)
		if payload > 999999 {
			payload = 0 // non-canonical
		}
		return Components{Sign: sign, Kind: kind, Payload: payload}
	}
	// Infinity
	if v&bid32InfMask == bid32InfMask {
		return Components{Sign: sign, Kind: Infinity}
	}

	var exp int
	var coeff uint32
	if v&bid32SteerMask == bid32SteerMask {
		// special encoding (implicit high bit)
		exp = int((v >> 21) & bid32ExpMask32)
		coeff = (v & 0x001fffff) | 0x00800000
		if coeff >= 10000000 {
			coeff = 0 // non-canonical
		}
	} else {
		exp = int((v >> 23) & bid32ExpMask32)
		coeff = v & 0x007fffff
	}

	if coeff == 0 {
		return Components{Sign: sign, Exponent: int32(exp - bid32Bias), Kind: Zero}
	}
	return Components{
		Sign:        sign,
		Coefficient: new(big.Int).SetUint64(uint64(coeff)),
		Exponent:    int32(exp - bid32Bias),
		Kind:        Normal,
	}
}

// Encode32 encodes components into a BID32 uint32.
// Coefficient must be <= 9999999. Exponent range: -101 to 90.
func Encode32(c Components) uint32 {
	var sgn uint32
	if c.Sign {
		sgn = bid32SignMask
	}
	switch c.Kind {
	case Infinity:
		return sgn | 0x78000000
	case QNaN:
		return sgn | 0x7c000000 | (uint32(c.Payload) & 0x000fffff)
	case SNaN:
		return sgn | 0x7e000000 | (uint32(c.Payload) & 0x000fffff)
	case Zero:
		exp := int(c.Exponent) + bid32Bias
		if exp < 0 {
			exp = 0
		} else if exp > 191 {
			exp = 191
		}
		return sgn | (uint32(exp) << 23)
	}
	// Normal
	coeff := uint32(c.Coefficient.Uint64())
	exp := int(c.Exponent) + bid32Bias
	if exp < 0 {
		exp = 0
	} else if exp > 191 {
		exp = 191
	}
	if coeff < 0x800000 {
		return sgn | (uint32(exp) << 23) | coeff
	}
	return sgn | (0x60000000) | (uint32(exp) << 21) | (coeff & 0x001fffff)
}

// --- BID64 ---

const (
	bid64NaNMask   = 0x7c00000000000000
	bid64SNaNMask  = 0x7e00000000000000
	bid64InfMask   = 0x7800000000000000
	bid64SignMask  = 0x8000000000000000
	bid64SteerMask = 0x6000000000000000
	bid64ExpMask   = 0x3ff
	bid64MaxCoeff  = 9999999999999999
	bid64Bias      = 398
)

// Decode64 extracts components from a BID64-encoded uint64.
func Decode64(v uint64) Components {
	sign := v&bid64SignMask != 0

	if v&bid64NaNMask == bid64NaNMask {
		kind := QNaN
		if v&bid64SNaNMask == bid64SNaNMask {
			kind = SNaN
		}
		payload := v & 0x0003ffffffffffff
		if payload > 999999999999999 {
			payload = 0
		}
		return Components{Sign: sign, Kind: kind, Payload: payload}
	}
	if v&bid64InfMask == bid64InfMask {
		return Components{Sign: sign, Kind: Infinity}
	}

	var exp int
	var coeff uint64
	if v&bid64SteerMask == bid64SteerMask {
		exp = int((v >> 51) & bid64ExpMask)
		coeff = (v & 0x0007ffffffffffff) | 0x0020000000000000
		if coeff > bid64MaxCoeff {
			coeff = 0
		}
	} else {
		exp = int((v >> 53) & bid64ExpMask)
		coeff = v & 0x001fffffffffffff
	}

	if coeff == 0 {
		return Components{Sign: sign, Exponent: int32(exp - bid64Bias), Kind: Zero}
	}
	return Components{
		Sign:        sign,
		Coefficient: new(big.Int).SetUint64(coeff),
		Exponent:    int32(exp - bid64Bias),
		Kind:        Normal,
	}
}

// Encode64 encodes components into a BID64 uint64.
func Encode64(c Components) uint64 {
	var sgn uint64
	if c.Sign {
		sgn = bid64SignMask
	}
	switch c.Kind {
	case Infinity:
		return sgn | 0x7800000000000000
	case QNaN:
		return sgn | 0x7c00000000000000 | (c.Payload & 0x0003ffffffffffff)
	case SNaN:
		return sgn | 0x7e00000000000000 | (c.Payload & 0x0003ffffffffffff)
	case Zero:
		exp := int(c.Exponent) + bid64Bias
		if exp < 0 {
			exp = 0
		} else if exp > 767 {
			exp = 767
		}
		return sgn | (uint64(exp) << 53)
	}
	coeff := c.Coefficient.Uint64()
	exp := int(c.Exponent) + bid64Bias
	if exp < 0 {
		exp = 0
	} else if exp > 767 {
		exp = 767
	}
	if coeff < 0x20000000000000 {
		return sgn | (uint64(exp) << 53) | coeff
	}
	return sgn | bid64SteerMask | (uint64(exp) << 51) | (coeff & 0x0007ffffffffffff)
}

// --- BID128 ---

const (
	bid128NaNMask   = 0x7c00000000000000
	bid128SNaNMask  = 0x7e00000000000000
	bid128InfMask   = 0x7800000000000000
	bid128SignMask  = 0x8000000000000000
	bid128SteerMask = 0x6000000000000000
	bid128ExpMask   = 0x3fff
	bid128Bias      = 6176
)

// ten34 = 10^34, max coefficient + 1 for BID128
var ten34 = func() *big.Int {
	v, _ := new(big.Int).SetString("10000000000000000000000000000000000", 10)
	return v
}()

// Decode128 extracts components from BID128 encoded as [2]uint64{lo, hi}.
func Decode128(lo, hi uint64) Components {
	sign := hi&bid128SignMask != 0

	if hi&bid128NaNMask == bid128NaNMask {
		kind := QNaN
		if hi&bid128SNaNMask == bid128SNaNMask {
			kind = SNaN
		}
		// payload: hi[45:0] and lo[63:0] = 110 bits
		payHi := hi & 0x00003fffffffffff
		coeff := new(big.Int).SetUint64(payHi)
		coeff.Lsh(coeff, 64)
		coeff.Or(coeff, new(big.Int).SetUint64(lo))
		ten33, _ := new(big.Int).SetString("1000000000000000000000000000000000", 10)
		if coeff.Cmp(ten33) >= 0 {
			return Components{Sign: sign, Kind: kind, Payload: 0}
		}
		return Components{Sign: sign, Kind: kind, Payload: lo} // simplified: lo only for payload
	}
	if hi&bid128InfMask == bid128InfMask {
		return Components{Sign: sign, Kind: Infinity}
	}

	var exp int
	var coeffHi uint64
	if hi&bid128SteerMask == bid128SteerMask {
		exp = int((hi >> 47) & bid128ExpMask)
		coeffHi = (hi & 0x00007fffffffffff) | 0x0020000000000000
	} else {
		exp = int((hi >> 49) & bid128ExpMask)
		coeffHi = hi & 0x0001ffffffffffff
	}

	coeff := new(big.Int).SetUint64(coeffHi)
	coeff.Lsh(coeff, 64)
	coeff.Or(coeff, new(big.Int).SetUint64(lo))

	if coeff.Cmp(ten34) >= 0 {
		coeff.SetUint64(0)
	}

	if coeff.Sign() == 0 {
		return Components{Sign: sign, Exponent: int32(exp - bid128Bias), Kind: Zero}
	}
	return Components{
		Sign:        sign,
		Coefficient: coeff,
		Exponent:    int32(exp - bid128Bias),
		Kind:        Normal,
	}
}

// Encode128 encodes components into BID128 as (lo, hi uint64).
func Encode128(c Components) (lo, hi uint64) {
	var sgn uint64
	if c.Sign {
		sgn = bid128SignMask
	}
	switch c.Kind {
	case Infinity:
		return 0, sgn | 0x7800000000000000
	case QNaN:
		return c.Payload, sgn | 0x7c00000000000000
	case SNaN:
		return c.Payload, sgn | 0x7e00000000000000
	case Zero:
		exp := int(c.Exponent) + bid128Bias
		if exp < 0 {
			exp = 0
		} else if exp > 12287 {
			exp = 12287
		}
		return 0, sgn | (uint64(exp) << 49)
	}
	// Normal: coefficient as 128 bits
	var coeffBytes [16]byte
	c.Coefficient.FillBytes(coeffBytes[:])
	coeffHi := binary.BigEndian.Uint64(coeffBytes[0:8])
	coeffLo := binary.BigEndian.Uint64(coeffBytes[8:16])

	exp := int(c.Exponent) + bid128Bias
	if exp < 0 {
		exp = 0
	} else if exp > 12287 {
		exp = 12287
	}

	lo = coeffLo
	hi = sgn | (uint64(exp) << 49) | (coeffHi & 0x0001ffffffffffff)
	return lo, hi
}

// --- Byte-level convenience ---

func requireByteLength(name string, b []byte, want int) error {
	if len(b) != want {
		return fmt.Errorf("%s: expected %d bytes, got %d", name, want, len(b))
	}
	return nil
}

// Decode32Bytes decodes 4 bytes (little-endian) as BID32.
func Decode32Bytes(b []byte) (Components, error) {
	if err := requireByteLength("Decode32Bytes", b, 4); err != nil {
		return Components{}, err
	}
	return Decode32(binary.LittleEndian.Uint32(b)), nil
}

// Decode64Bytes decodes 8 bytes (little-endian) as BID64.
func Decode64Bytes(b []byte) (Components, error) {
	if err := requireByteLength("Decode64Bytes", b, 8); err != nil {
		return Components{}, err
	}
	return Decode64(binary.LittleEndian.Uint64(b)), nil
}

// Decode128Bytes decodes 16 bytes (little-endian) as BID128.
func Decode128Bytes(b []byte) (Components, error) {
	if err := requireByteLength("Decode128Bytes", b, 16); err != nil {
		return Components{}, err
	}
	lo := binary.LittleEndian.Uint64(b[0:8])
	hi := binary.LittleEndian.Uint64(b[8:16])
	return Decode128(lo, hi), nil
}

// Encode32Bytes encodes components as 4 bytes (little-endian) BID32.
func Encode32Bytes(c Components) [4]byte {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], Encode32(c))
	return buf
}

// Encode64Bytes encodes components as 8 bytes (little-endian) BID64.
func Encode64Bytes(c Components) [8]byte {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], Encode64(c))
	return buf
}

// Encode128Bytes encodes components as 16 bytes (little-endian) BID128.
func Encode128Bytes(c Components) [16]byte {
	var buf [16]byte
	lo, hi := Encode128(c)
	binary.LittleEndian.PutUint64(buf[0:8], lo)
	binary.LittleEndian.PutUint64(buf[8:16], hi)
	return buf
}
