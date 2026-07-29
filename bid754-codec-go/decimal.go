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
	// Payload is the NaN payload (only meaningful for QNaN/SNaN). It represents
	// the full BID128 110-bit NaN payload as an unsigned integer, subject to the
	// schema-wide value limit of 10^33 (the widest canonical BID128 NaN payload,
	// mirroring the 10^34-1 coefficient cap). BID32/BID64 payloads are
	// unaffected subsets of this field. A nil Payload on a QNaN/SNaN is the
	// documented default zero payload (it encodes as payload 0); this is an
	// explicit field default, not an implicit fallback. Decode returns a
	// non-nil Payload (zero for a bare or non-canonical NaN).
	Payload *big.Int
}

func hasNonzeroComponent(v *big.Int) bool {
	return v != nil && v.Sign() != 0
}

// validateKindFields rejects fields that the selected Kind cannot carry. A
// zero-valued unused field is accepted because several language bindings use
// non-nullable numeric fields whose declared default is zero; a nonzero value
// would be silently lost by the BID encoding and must therefore fail.
func validateKindFields(operation string, c Components) error {
	switch c.Kind {
	case Normal:
		if c.Coefficient != nil && c.Coefficient.Sign() == 0 {
			return fmt.Errorf("%s: normal value cannot carry a zero coefficient", operation)
		}
		if hasNonzeroComponent(c.Payload) {
			return fmt.Errorf("%s: normal value cannot carry NaN payload %s", operation, c.Payload.String())
		}
	case Zero:
		if hasNonzeroComponent(c.Coefficient) {
			return fmt.Errorf("%s: zero value cannot carry coefficient %s", operation, c.Coefficient.String())
		}
		if hasNonzeroComponent(c.Payload) {
			return fmt.Errorf("%s: zero value cannot carry NaN payload %s", operation, c.Payload.String())
		}
	case Infinity:
		if hasNonzeroComponent(c.Coefficient) {
			return fmt.Errorf("%s: infinity cannot carry coefficient %s", operation, c.Coefficient.String())
		}
		if c.Exponent != 0 {
			return fmt.Errorf("%s: infinity cannot carry exponent %d", operation, c.Exponent)
		}
		if hasNonzeroComponent(c.Payload) {
			return fmt.Errorf("%s: infinity cannot carry NaN payload %s", operation, c.Payload.String())
		}
	case QNaN, SNaN:
		if hasNonzeroComponent(c.Coefficient) {
			return fmt.Errorf("%s: NaN cannot carry coefficient %s", operation, c.Coefficient.String())
		}
		if c.Exponent != 0 {
			return fmt.Errorf("%s: NaN cannot carry exponent %d", operation, c.Exponent)
		}
	}
	return nil
}

// --- BID32 ---

const (
	bid32NaNMask   = 0x7c000000
	bid32SNaNMask  = 0x7e000000
	bid32InfMask   = 0x78000000
	bid32SignMask  = 0x80000000
	bid32SteerMask = 0x60000000
	bid32ExpMask32 = 0xff
	bid32MaxCoeff  = 9999999 // 10^7-1
	bid32Bias      = 101
	bid32MinExp    = -101    // unbiased exponent lower bound (biased 0)
	bid32MaxExp    = 90      // unbiased exponent upper bound (biased 191)
	bid32PayLimit  = 1000000 // 10^6; canonical NaN payload must be < this
)

// Decode32 extracts components from a BID32-encoded uint32.
func Decode32(v uint32) Components {
	fields := decode32Fields(v)
	c := Components{
		Sign:     fields.sign,
		Exponent: fields.exponent,
		Kind:     fields.kind,
	}
	switch fields.kind {
	case Normal:
		c.Coefficient = new(big.Int).SetUint64(uint64(fields.coefficient))
	case QNaN, SNaN:
		c.Payload = new(big.Int).SetUint64(uint64(fields.payload))
	}
	return c
}

// decoded32Fields is the allocation-free production representation used by
// Decode32 before it materializes the public big.Int-backed Components value.
// Keeping the bit-layout decision in this primitive lets the exhaustive long
// harness execute the exact production decode path over all 2^32 encodings
// without billions of heap allocations.
type decoded32Fields struct {
	sign        bool
	coefficient uint32
	exponent    int32
	kind        Kind
	payload     uint32
}

func decode32Fields(v uint32) decoded32Fields {
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
		return decoded32Fields{sign: sign, kind: kind, payload: uint32(payload)}
	}
	// Infinity
	if v&bid32InfMask == bid32InfMask {
		return decoded32Fields{sign: sign, kind: Infinity}
	}

	var exp int
	var coeff uint32
	if v&bid32SteerMask == bid32SteerMask {
		// special encoding (implicit high bit)
		exp = int((v >> 21) & bid32ExpMask32)
		coeff = (v & 0x001fffff) | 0x00800000
		if coeff > bid32MaxCoeff {
			coeff = 0 // non-canonical
		}
	} else {
		exp = int((v >> 23) & bid32ExpMask32)
		coeff = v & 0x007fffff
	}

	if coeff == 0 {
		return decoded32Fields{sign: sign, exponent: int32(exp - bid32Bias), kind: Zero}
	}
	return decoded32Fields{
		sign:        sign,
		coefficient: coeff,
		exponent:    int32(exp - bid32Bias),
		kind:        Normal,
	}
}

// packBID32Fields is the allocation-free packing primitive used after the
// public Encode32 validator has proved every field representable. The caller
// must supply a valid decoded32Fields value.
func packBID32Fields(fields decoded32Fields) uint32 {
	var sgn uint32
	if fields.sign {
		sgn = bid32SignMask
	}
	switch fields.kind {
	case Infinity:
		return sgn | 0x78000000
	case QNaN:
		return sgn | 0x7c000000 | fields.payload
	case SNaN:
		return sgn | 0x7e000000 | fields.payload
	case Zero:
		exp := uint32(fields.exponent + bid32Bias)
		return sgn | (exp << 23)
	case Normal:
		exp := uint32(fields.exponent + bid32Bias)
		if fields.coefficient < 0x800000 {
			return sgn | (exp << 23) | fields.coefficient
		}
		return sgn | bid32SteerMask | (exp << 21) | (fields.coefficient & 0x001fffff)
	default:
		panic("packBID32Fields called with invalid kind")
	}
}

// bid32BiasedExp validates an unbiased BID32 exponent and returns its biased
// form. It rejects out-of-range exponents instead of clamping them.
func bid32BiasedExp(exp int32) (uint32, error) {
	if exp < bid32MinExp || exp > bid32MaxExp {
		return 0, fmt.Errorf("bid32 encode: exponent %d out of range [%d, %d]", exp, bid32MinExp, bid32MaxExp)
	}
	return uint32(exp + bid32Bias), nil
}

// bid32NaNPayloadBits validates a NaN payload for BID32 encoding and returns its
// packed low bits. A nil payload is the documented default zero. It rejects a
// negative payload and a payload at or above the canonical limit (10^6) instead
// of masking or clamping.
func bid32NaNPayloadBits(payload *big.Int) (uint32, error) {
	if payload == nil {
		return 0, nil
	}
	if payload.Sign() < 0 {
		return 0, fmt.Errorf("bid32 encode: NaN payload %s is negative", payload.String())
	}
	if payload.Cmp(bigBid32PayLimit()) >= 0 {
		return 0, fmt.Errorf("bid32 encode: NaN payload %s exceeds max %d", payload.String(), bid32PayLimit-1)
	}
	return uint32(payload.Uint64()) & 0x000fffff, nil
}

// Encode32 encodes components into a BID32 uint32.
//
// It is a validating packing API: it rejects any Components field that is not
// representable in BID32 exactly as supplied, returning an error rather than
// silently truncating, masking, or clamping. In-range values encode unchanged.
//   - coefficient (Normal): 1 .. 9999999
//   - exponent (Zero/Normal): unbiased -101 .. 90
//   - payload (QNaN/SNaN): 0 .. 999999
func Encode32(c Components) (uint32, error) {
	if err := validateKindFields("bid32 encode", c); err != nil {
		return 0, err
	}
	switch c.Kind {
	case Infinity:
		return packBID32Fields(decoded32Fields{sign: c.Sign, kind: Infinity}), nil
	case QNaN:
		bits, err := bid32NaNPayloadBits(c.Payload)
		if err != nil {
			return 0, err
		}
		return packBID32Fields(decoded32Fields{sign: c.Sign, kind: QNaN, payload: bits}), nil
	case SNaN:
		bits, err := bid32NaNPayloadBits(c.Payload)
		if err != nil {
			return 0, err
		}
		return packBID32Fields(decoded32Fields{sign: c.Sign, kind: SNaN, payload: bits}), nil
	case Zero:
		if _, err := bid32BiasedExp(c.Exponent); err != nil {
			return 0, err
		}
		return packBID32Fields(decoded32Fields{sign: c.Sign, exponent: c.Exponent, kind: Zero}), nil
	case Normal:
		if c.Coefficient == nil {
			return 0, fmt.Errorf("bid32 encode: normal value has nil coefficient")
		}
		if c.Coefficient.Sign() < 0 {
			return 0, fmt.Errorf("bid32 encode: coefficient %s is negative", c.Coefficient.String())
		}
		if c.Coefficient.Cmp(bigBid32MaxCoeff()) > 0 {
			return 0, fmt.Errorf("bid32 encode: coefficient %s exceeds max %d", c.Coefficient.String(), bid32MaxCoeff)
		}
		if _, err := bid32BiasedExp(c.Exponent); err != nil {
			return 0, err
		}
		coeff := uint32(c.Coefficient.Uint64()) // safe: coefficient <= 9999999
		return packBID32Fields(decoded32Fields{
			sign: c.Sign, coefficient: coeff, exponent: c.Exponent, kind: Normal,
		}), nil
	default:
		return 0, fmt.Errorf("bid32 encode: unrecognized kind %d", uint8(c.Kind))
	}
}

// --- BID64 ---

const (
	bid64NaNMask   = 0x7c00000000000000
	bid64SNaNMask  = 0x7e00000000000000
	bid64InfMask   = 0x7800000000000000
	bid64SignMask  = 0x8000000000000000
	bid64SteerMask = 0x6000000000000000
	bid64ExpMask   = 0x3ff
	bid64MaxCoeff  = 9999999999999999 // 10^16-1
	bid64Bias      = 398
	bid64MinExp    = -398             // unbiased exponent lower bound (biased 0)
	bid64MaxExp    = 369              // unbiased exponent upper bound (biased 767)
	bid64PayLimit  = 1000000000000000 // 10^15; canonical NaN payload must be < this
)

// Decode64 extracts components from a BID64-encoded uint64.
func Decode64(v uint64) Components {
	fields := decode64Fields(v)
	c := Components{
		Sign:     fields.sign,
		Exponent: fields.exponent,
		Kind:     fields.kind,
	}
	switch fields.kind {
	case Normal:
		c.Coefficient = new(big.Int).SetUint64(fields.coefficient)
	case QNaN, SNaN:
		c.Payload = new(big.Int).SetUint64(fields.payload)
	}
	return c
}

// decoded64Fields is the allocation-free production representation used by
// Decode64 before it materializes the public big.Int-backed Components value.
// The long codec harness exercises this exact primitive for deterministic
// large-scale raw-word differential verification.
type decoded64Fields struct {
	sign        bool
	coefficient uint64
	exponent    int32
	kind        Kind
	payload     uint64
}

func decode64Fields(v uint64) decoded64Fields {
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
		return decoded64Fields{sign: sign, kind: kind, payload: payload}
	}
	if v&bid64InfMask == bid64InfMask {
		return decoded64Fields{sign: sign, kind: Infinity}
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
		return decoded64Fields{sign: sign, exponent: int32(exp - bid64Bias), kind: Zero}
	}
	return decoded64Fields{
		sign:        sign,
		coefficient: coeff,
		exponent:    int32(exp - bid64Bias),
		kind:        Normal,
	}
}

// packBID64Fields is the allocation-free packing primitive used after the
// public Encode64 validator has proved every field representable. The caller
// must supply a valid decoded64Fields value.
func packBID64Fields(fields decoded64Fields) uint64 {
	var sign uint64
	if fields.sign {
		sign = bid64SignMask
	}
	switch fields.kind {
	case Infinity:
		return sign | bid64InfMask
	case QNaN:
		return sign | bid64NaNMask | fields.payload
	case SNaN:
		return sign | bid64SNaNMask | fields.payload
	case Zero:
		exp := uint64(fields.exponent + bid64Bias)
		return sign | (exp << 53)
	case Normal:
		exp := uint64(fields.exponent + bid64Bias)
		if fields.coefficient < 0x20000000000000 {
			return sign | (exp << 53) | fields.coefficient
		}
		return sign | bid64SteerMask | (exp << 51) | (fields.coefficient & 0x0007ffffffffffff)
	default:
		panic("packBID64Fields called with invalid kind")
	}
}

// bid64BiasedExp validates an unbiased BID64 exponent and returns its biased
// form. It rejects out-of-range exponents instead of clamping them.
func bid64BiasedExp(exp int32) (uint64, error) {
	if exp < bid64MinExp || exp > bid64MaxExp {
		return 0, fmt.Errorf("bid64 encode: exponent %d out of range [%d, %d]", exp, bid64MinExp, bid64MaxExp)
	}
	return uint64(exp + bid64Bias), nil
}

// bid64NaNPayloadBits validates a NaN payload for BID64 encoding and returns its
// packed low bits. A nil payload is the documented default zero. It rejects a
// negative payload and a payload at or above the canonical limit (10^15) instead
// of masking or clamping.
func bid64NaNPayloadBits(payload *big.Int) (uint64, error) {
	if payload == nil {
		return 0, nil
	}
	if payload.Sign() < 0 {
		return 0, fmt.Errorf("bid64 encode: NaN payload %s is negative", payload.String())
	}
	if payload.Cmp(bigBid64PayLimit()) >= 0 {
		return 0, fmt.Errorf("bid64 encode: NaN payload %s exceeds max %d", payload.String(), uint64(bid64PayLimit-1))
	}
	return payload.Uint64() & 0x0003ffffffffffff, nil
}

// Encode64 encodes components into a BID64 uint64.
//
// It is a validating packing API: it rejects any Components field that is not
// representable in BID64 exactly as supplied, returning an error rather than
// silently truncating, masking, or clamping. In-range values encode unchanged.
//   - coefficient (Normal): 1 .. 9999999999999999
//   - exponent (Zero/Normal): unbiased -398 .. 369
//   - payload (QNaN/SNaN): 0 .. 999999999999999
func Encode64(c Components) (uint64, error) {
	if err := validateKindFields("bid64 encode", c); err != nil {
		return 0, err
	}
	switch c.Kind {
	case Infinity:
		return packBID64Fields(decoded64Fields{sign: c.Sign, kind: Infinity}), nil
	case QNaN:
		bits, err := bid64NaNPayloadBits(c.Payload)
		if err != nil {
			return 0, err
		}
		return packBID64Fields(decoded64Fields{sign: c.Sign, kind: QNaN, payload: bits}), nil
	case SNaN:
		bits, err := bid64NaNPayloadBits(c.Payload)
		if err != nil {
			return 0, err
		}
		return packBID64Fields(decoded64Fields{sign: c.Sign, kind: SNaN, payload: bits}), nil
	case Zero:
		if _, err := bid64BiasedExp(c.Exponent); err != nil {
			return 0, err
		}
		return packBID64Fields(decoded64Fields{sign: c.Sign, exponent: c.Exponent, kind: Zero}), nil
	case Normal:
		if c.Coefficient == nil {
			return 0, fmt.Errorf("bid64 encode: normal value has nil coefficient")
		}
		if c.Coefficient.Sign() < 0 {
			return 0, fmt.Errorf("bid64 encode: coefficient %s is negative", c.Coefficient.String())
		}
		if c.Coefficient.Cmp(bigBid64MaxCoeff()) > 0 {
			return 0, fmt.Errorf("bid64 encode: coefficient %s exceeds max %d", c.Coefficient.String(), uint64(bid64MaxCoeff))
		}
		if _, err := bid64BiasedExp(c.Exponent); err != nil {
			return 0, err
		}
		coeff := c.Coefficient.Uint64() // safe: coefficient <= 9999999999999999
		return packBID64Fields(decoded64Fields{
			sign: c.Sign, coefficient: coeff, exponent: c.Exponent, kind: Normal,
		}), nil
	default:
		return 0, fmt.Errorf("bid64 encode: unrecognized kind %d", uint8(c.Kind))
	}
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
	bid128MinExp    = -6176 // unbiased exponent lower bound (biased 0)
	bid128MaxExp    = 6111  // unbiased exponent upper bound (biased 12287)
	bid128Ten33Hi   = 0x0000314dc6448d93
	bid128Ten33Lo   = 0x38c15b0a00000000
	bid128Ten34Hi   = 0x0001ed09bead87c0
	bid128Ten34Lo   = 0x378d8e6400000000
	// bid128MaxCoeffDecimal is 10^34-1, the largest representable BID128
	// coefficient, used in reject error messages.
	bid128MaxCoeffDecimal = "9999999999999999999999999999999999"
)

// The big.Int limits below are constructed fresh on every call instead of
// being shared package-level *big.Int values: big.Int methods write into
// their receiver, so a shared pointer would be mutable process-wide state
// that only convention keeps constant. A fresh value per call makes the
// immutability structural.

// ten34 returns 10^34, max coefficient + 1 for BID128, as
// 10^17 * 10^17 (each factor fits uint64, so this is one small
// multiplication rather than a generic exponentiation).
func ten34() *big.Int {
	t := new(big.Int).SetUint64(100000000000000000) // 10^17
	return t.Mul(t, t)
}

// ten33 returns 10^33, the schema-wide NaN payload limit (max canonical
// BID128 NaN payload + 1), as 10^16 * 10^17. A NaN payload must be strictly
// below this in every BID width, mirroring the 10^34-1 coefficient cap.
func ten33() *big.Int {
	t := new(big.Int).SetUint64(10000000000000000) // 10^16
	return t.Mul(t, new(big.Int).SetUint64(100000000000000000))
}

// bid128PayMax returns 10^33-1, the largest canonical BID128 NaN payload,
// used in reject error messages.
func bid128PayMax() *big.Int {
	return new(big.Int).Sub(ten33(), big.NewInt(1))
}

// Unsigned coefficient and NaN payload limits for encode validation.
func bigBid32MaxCoeff() *big.Int { return big.NewInt(bid32MaxCoeff) } // 10^7-1
func bigBid64MaxCoeff() *big.Int { return big.NewInt(bid64MaxCoeff) } // 10^16-1
func bigBid32PayLimit() *big.Int { return big.NewInt(bid32PayLimit) } // 10^6
func bigBid64PayLimit() *big.Int { return big.NewInt(bid64PayLimit) } // 10^15

// Decode128 extracts components from BID128 encoded as [2]uint64{lo, hi}.
func Decode128(lo, hi uint64) Components {
	fields := decode128Fields(lo, hi)
	c := Components{
		Sign:     fields.sign,
		Exponent: fields.exponent,
		Kind:     fields.kind,
	}
	switch fields.kind {
	case Normal:
		c.Coefficient = bid128BigIntFromWords(fields.coefficientLo, fields.coefficientHi)
	case QNaN, SNaN:
		c.Payload = bid128BigIntFromWords(fields.payloadLo, fields.payloadHi)
	}
	return c
}

// decoded128Fields is the allocation-free production representation used by
// Decode128. Coefficients and NaN payloads remain split into little-word-order
// uint64 pairs until the public Components value is materialized.
type decoded128Fields struct {
	sign          bool
	coefficientLo uint64
	coefficientHi uint64
	exponent      int32
	kind          Kind
	payloadLo     uint64
	payloadHi     uint64
}

func bid128WordsAtLeast(lo, hi, limitLo, limitHi uint64) bool {
	return hi > limitHi || hi == limitHi && lo >= limitLo
}

func bid128BigIntFromWords(lo, hi uint64) *big.Int {
	value := new(big.Int).SetUint64(hi)
	value.Lsh(value, 64)
	return value.Or(value, new(big.Int).SetUint64(lo))
}

func decode128Fields(lo, hi uint64) decoded128Fields {
	sign := hi&bid128SignMask != 0

	if hi&bid128NaNMask == bid128NaNMask {
		kind := QNaN
		if hi&bid128SNaNMask == bid128SNaNMask {
			kind = SNaN
		}
		payHi := hi & 0x00003fffffffffff
		if bid128WordsAtLeast(lo, payHi, bid128Ten33Lo, bid128Ten33Hi) {
			lo, payHi = 0, 0
		}
		return decoded128Fields{sign: sign, kind: kind, payloadLo: lo, payloadHi: payHi}
	}
	if hi&bid128InfMask == bid128InfMask {
		return decoded128Fields{sign: sign, kind: Infinity}
	}

	var exp uint64
	if hi&bid128SteerMask == bid128SteerMask {
		// Intel unpack_BID128 classifies every finite steering-form encoding as
		// non-canonical. Its implied coefficient is outside the decimal128
		// canonical range, so it decodes to a zero while retaining the encoded
		// exponent cohort.
		exp = (hi >> 47) & bid128ExpMask
		return decoded128Fields{
			sign: sign, exponent: int32(exp) - bid128Bias, kind: Zero,
		}
	}

	exp = (hi >> 49) & bid128ExpMask
	coeffHi := hi & 0x0001ffffffffffff
	if bid128WordsAtLeast(lo, coeffHi, bid128Ten34Lo, bid128Ten34Hi) {
		lo, coeffHi = 0, 0
	}
	if lo == 0 && coeffHi == 0 {
		return decoded128Fields{sign: sign, exponent: int32(exp) - bid128Bias, kind: Zero}
	}
	return decoded128Fields{
		sign:          sign,
		coefficientLo: lo,
		coefficientHi: coeffHi,
		exponent:      int32(exp) - bid128Bias,
		kind:          Normal,
	}
}

// packBID128Fields is the allocation-free canonical packing primitive used
// after the public Encode128 validator has proved every field representable.
// BID128 canonical finite values always use the small-coefficient form;
// finite steering-form encodings are non-canonical in Intel BID C.
func packBID128Fields(fields decoded128Fields) (lo, hi uint64) {
	if fields.sign {
		hi = bid128SignMask
	}
	switch fields.kind {
	case Infinity:
		return 0, hi | bid128InfMask
	case QNaN:
		return fields.payloadLo, hi | bid128NaNMask | fields.payloadHi
	case SNaN:
		return fields.payloadLo, hi | bid128SNaNMask | fields.payloadHi
	case Zero:
		exp := uint64(fields.exponent + bid128Bias)
		return 0, hi | (exp << 49)
	case Normal:
		exp := uint64(fields.exponent + bid128Bias)
		return fields.coefficientLo, hi | (exp << 49) | fields.coefficientHi
	default:
		panic("packBID128Fields called with invalid kind")
	}
}

// bid128BiasedExp validates an unbiased BID128 exponent and returns its biased
// form. It rejects out-of-range exponents instead of clamping them.
func bid128BiasedExp(exp int32) (uint64, error) {
	if exp < bid128MinExp || exp > bid128MaxExp {
		return 0, fmt.Errorf("bid128 encode: exponent %d out of range [%d, %d]", exp, bid128MinExp, bid128MaxExp)
	}
	return uint64(exp + bid128Bias), nil
}

// bid128NaNPayloadWords validates a NaN payload for BID128 encoding and splits
// it into the (lo, hi) payload words. A nil payload is the documented default
// zero. It rejects a negative payload and a payload at or above the canonical
// limit (10^33, the no-silent-failure boundary) instead of masking or clamping.
// The returned hi holds only the payload's high bits (bits 64..109); the caller
// ORs in the NaN and sign bits.
func bid128NaNPayloadWords(payload *big.Int) (lo, hi uint64, err error) {
	if payload == nil {
		return 0, 0, nil
	}
	if payload.Sign() < 0 {
		return 0, 0, fmt.Errorf("bid128 encode: NaN payload %s is negative", payload.String())
	}
	if payload.Cmp(ten33()) >= 0 {
		return 0, 0, fmt.Errorf("bid128 encode: NaN payload %s exceeds max %s", payload.String(), bid128PayMax().String())
	}
	// payload < 10^33 < 2^110, so it fits 16 bytes and its high word occupies
	// only the 46-bit payload field (bits 64..109), never the NaN/sign bits, so
	// the caller can OR the mask in without collision or masking.
	var payBytes [16]byte
	payload.FillBytes(payBytes[:]) // safe: payload < 10^33 < 2^128
	hi = binary.BigEndian.Uint64(payBytes[0:8])
	lo = binary.BigEndian.Uint64(payBytes[8:16])
	return lo, hi, nil
}

// Encode128 encodes components into BID128 as (lo, hi uint64).
//
// It is a validating packing API: it rejects any Components field that is not
// representable in BID128 exactly as supplied, returning an error rather than
// silently truncating, masking, clamping, or panicking. In-range values encode
// unchanged.
//   - coefficient (Normal): 1 .. 10^34-1
//   - exponent (Zero/Normal): unbiased -6176 .. 6111
//   - payload (QNaN/SNaN): 0 .. 10^33-1, split across the (lo, hi) payload
//     words; a payload at or above 10^33 is rejected (the no-silent-failure
//     boundary), and a nil payload is the documented default zero.
func Encode128(c Components) (lo, hi uint64, err error) {
	if err := validateKindFields("bid128 encode", c); err != nil {
		return 0, 0, err
	}
	switch c.Kind {
	case Infinity:
		lo, hi = packBID128Fields(decoded128Fields{sign: c.Sign, kind: Infinity})
		return lo, hi, nil
	case QNaN:
		payLo, payHi, err := bid128NaNPayloadWords(c.Payload)
		if err != nil {
			return 0, 0, err
		}
		lo, hi = packBID128Fields(decoded128Fields{
			sign: c.Sign, kind: QNaN, payloadLo: payLo, payloadHi: payHi,
		})
		return lo, hi, nil
	case SNaN:
		payLo, payHi, err := bid128NaNPayloadWords(c.Payload)
		if err != nil {
			return 0, 0, err
		}
		lo, hi = packBID128Fields(decoded128Fields{
			sign: c.Sign, kind: SNaN, payloadLo: payLo, payloadHi: payHi,
		})
		return lo, hi, nil
	case Zero:
		if _, err := bid128BiasedExp(c.Exponent); err != nil {
			return 0, 0, err
		}
		lo, hi = packBID128Fields(decoded128Fields{
			sign: c.Sign, exponent: c.Exponent, kind: Zero,
		})
		return lo, hi, nil
	case Normal:
		if c.Coefficient == nil {
			return 0, 0, fmt.Errorf("bid128 encode: normal value has nil coefficient")
		}
		if c.Coefficient.Sign() < 0 {
			return 0, 0, fmt.Errorf("bid128 encode: coefficient %s is negative", c.Coefficient.String())
		}
		// Reject coefficient >= 10^34 before FillBytes. Because 10^34 < 2^128,
		// this also prevents the 16-byte FillBytes below from panicking on a
		// coefficient that does not fit 128 bits.
		if c.Coefficient.Cmp(ten34()) >= 0 {
			return 0, 0, fmt.Errorf("bid128 encode: coefficient %s exceeds max %s", c.Coefficient.String(), bid128MaxCoeffDecimal)
		}
		if _, err := bid128BiasedExp(c.Exponent); err != nil {
			return 0, 0, err
		}
		var coeffBytes [16]byte
		c.Coefficient.FillBytes(coeffBytes[:]) // safe: coefficient < 10^34 < 2^128
		coeffHi := binary.BigEndian.Uint64(coeffBytes[0:8])
		coeffLo := binary.BigEndian.Uint64(coeffBytes[8:16])
		lo, hi = packBID128Fields(decoded128Fields{
			sign:          c.Sign,
			coefficientLo: coeffLo,
			coefficientHi: coeffHi,
			exponent:      c.Exponent,
			kind:          Normal,
		})
		return lo, hi, nil
	default:
		return 0, 0, fmt.Errorf("bid128 encode: unrecognized kind %d", uint8(c.Kind))
	}
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

// Encode32Bytes encodes components as 4 bytes (little-endian) BID32. It rejects
// invalid Components with the same contract as Encode32.
func Encode32Bytes(c Components) ([4]byte, error) {
	var buf [4]byte
	v, err := Encode32(c)
	if err != nil {
		return buf, err
	}
	binary.LittleEndian.PutUint32(buf[:], v)
	return buf, nil
}

// Encode64Bytes encodes components as 8 bytes (little-endian) BID64. It rejects
// invalid Components with the same contract as Encode64.
func Encode64Bytes(c Components) ([8]byte, error) {
	var buf [8]byte
	v, err := Encode64(c)
	if err != nil {
		return buf, err
	}
	binary.LittleEndian.PutUint64(buf[:], v)
	return buf, nil
}

// Encode128Bytes encodes components as 16 bytes (little-endian) BID128. It
// rejects invalid Components with the same contract as Encode128.
func Encode128Bytes(c Components) ([16]byte, error) {
	var buf [16]byte
	lo, hi, err := Encode128(c)
	if err != nil {
		return buf, err
	}
	binary.LittleEndian.PutUint64(buf[0:8], lo)
	binary.LittleEndian.PutUint64(buf[8:16], hi)
	return buf, nil
}
