// bid32_pure.go - Pure Go implementation of IEEE 754-2019 decimal32 BID encoding
//
// This file provides a CGO-free implementation of decimal32 arithmetic using
// Binary Integer Decimal (BID) encoding.

package bidgo

import (
	"errors"
	"strconv"
	"strings"
)

var errInvalidFormat32 = errors.New("invalid decimal32 format")

// Decimal32Pure is a pure Go implementation of IEEE 754-2019 decimal32 with BID encoding.
type Decimal32Pure uint32

// BID32 인코딩 상수
const (
	bid32SignMask            uint32 = 0x80000000 // 부호 비트
	bid32SpecialEncodingMask uint32 = 0x60000000 // 특수 인코딩 (coefficient >= 2^23)
	bid32InfinityMask        uint32 = 0x78000000 // Infinity
	bid32NaNMask             uint32 = 0x7c000000 // NaN
	bid32SNaNMask            uint32 = 0x7e000000 // Signaling NaN pattern
	bid32QuietMask           uint32 = 0xfdffffff // QNaN mask
	bid32LargeCoeffMask      uint32 = 0x007fffff // 23비트 coefficient (특수 인코딩)
	bid32SmallCoeffMask      uint32 = 0x001fffff // 21비트 coefficient (일반 인코딩)
	bid32LargeCoeffHighBit   uint32 = 0x00800000 // 2^23 (특수 인코딩 시 암묵적)
	bid32ExponentMask        uint32 = 0xff       // 8비트 지수
	bid32ExponentBias        int    = 101        // 지수 바이어스
	bid32MaxExponent         int    = 191        // 최대 biased 지수
	bid32MaxCoefficient      uint32 = 9999999    // 10^7 - 1

	// Intel BID 상수 (bid_internal.h)
	bid32LargestBID32 uint32 = 0x77f8967f // LARGEST_BID32: +9999999 * 10^90

	// 지수 시프트 위치
	bid32SmallExpShift = 23 // 일반 인코딩에서 지수 시프트
	bid32LargeExpShift = 21 // 특수 인코딩에서 지수 시프트
)

// unpack_BID32 extracts sign, exponent, and coefficient from BID32
// Returns false if the value is NaN or Infinity
func unpack_BID32(x uint32) (sign uint32, exponent int, coefficient uint32, valid bool) {
	sign = x & 0x80000000

	if (x & SPECIAL_ENCODING_MASK32) == SPECIAL_ENCODING_MASK32 {
		// special encodings
		if (x & INFINITY_MASK32) == INFINITY_MASK32 {
			coefficient = x & 0xfe0fffff
			if (x & 0x000fffff) >= 1000000 {
				coefficient = x & 0xfe000000
			}
			if (x & NAN_MASK32) == INFINITY_MASK32 {
				coefficient = x & 0xf8000000
			}
			exponent = 0
			return sign, exponent, coefficient, false // NaN or Infinity
		}
		// coefficient
		coefficient = (x & SMALL_COEFF_MASK32) | LARGE_COEFF_HIGH_BIT32
		// check for non-canonical value
		if coefficient >= 10000000 {
			coefficient = 0
		}
		// get exponent
		exponent = int((x >> 21) & EXPONENT_MASK32)
	} else {
		// exponent
		exponent = int((x >> 23) & EXPONENT_MASK32)
		// coefficient
		coefficient = x & LARGE_COEFF_MASK32
	}

	return sign, exponent, coefficient, coefficient != 0
}

// encodeBID32 encodes sign, exponent, and coefficient into BID32
func encodeBID32(sign int, exponent int, coefficient uint32) uint32 {
	var sgn uint32
	if sign != 0 {
		sgn = bid32SignMask
	}

	// 지수 클램핑
	if exponent < 0 {
		exponent = 0
	} else if exponent > bid32MaxExponent {
		exponent = bid32MaxExponent
	}

	// 계수가 2^23 이상이면 특수 인코딩
	if coefficient >= bid32LargeCoeffHighBit {
		return sgn | bid32SpecialEncodingMask |
			(uint32(exponent) << bid32LargeExpShift) |
			(coefficient & bid32SmallCoeffMask)
	}

	// 일반 인코딩
	return sgn | (uint32(exponent) << bid32SmallExpShift) | coefficient
}

// IsNaN returns true if d is NaN
func (d Decimal32Pure) IsNaN() bool {
	return (uint32(d) & bid32NaNMask) == bid32NaNMask
}

// IsInf returns true if d is infinity
func (d Decimal32Pure) IsInf() bool {
	return (uint32(d) & bid32NaNMask) == bid32InfinityMask
}

// IsZero returns true if d is zero (positive or negative)
func (d Decimal32Pure) IsZero() bool {
	return Bid32IsZero(uint32(d))
}

// Sign returns -1 for negative, 0 for zero, 1 for positive
func (d Decimal32Pure) Sign() int {
	if d.IsNaN() {
		return 0
	}
	if d.IsZero() {
		return 0
	}
	if (uint32(d) & bid32SignMask) != 0 {
		return -1
	}
	return 1
}

// Neg returns -d
func (d Decimal32Pure) Neg() Decimal32Pure {
	return Decimal32Pure(uint32(d) ^ bid32SignMask)
}

// Abs returns |d|
func (d Decimal32Pure) Abs() Decimal32Pure {
	return Decimal32Pure(uint32(d) &^ bid32SignMask)
}

// String returns the string representation of d
func (d Decimal32Pure) String() string {
	x := uint32(d)

	// 부호 추출
	negative := (x & bid32SignMask) != 0

	// NaN 처리
	if (x & bid32NaNMask) == bid32NaNMask {
		if (x & bid32SNaNMask) == bid32SNaNMask {
			if negative {
				return "-sNaN"
			}
			return "sNaN"
		}
		if negative {
			return "-NaN"
		}
		return "NaN"
	}

	// Infinity 처리
	if (x & bid32NaNMask) == bid32InfinityMask {
		if negative {
			return "-Infinity"
		}
		return "Infinity"
	}

	// 계수와 지수 추출
	_, exp, coeff, _ := unpack_BID32(x)
	exp -= bid32ExponentBias // unbiased exponent

	// 0 처리
	if coeff == 0 {
		if negative {
			return "-0"
		}
		return "0"
	}

	// 후행 0 제거하고 지수 조정
	for coeff%10 == 0 && coeff > 0 {
		coeff /= 10
		exp++
	}

	// 계수를 문자열로
	coeffStr := uitoa32(coeff)
	coeffLen := len(coeffStr)

	// 지수 조정 (소수점 위치)
	adjExp := exp + coeffLen - 1

	var result strings.Builder

	if negative {
		result.WriteByte('-')
	}

	// 지수 표기법 사용 여부 결정
	if adjExp < -6 || adjExp > 6 {
		// 지수 표기법
		result.WriteByte(coeffStr[0])
		if coeffLen > 1 {
			result.WriteByte('.')
			result.WriteString(coeffStr[1:])
		}
		result.WriteByte('E')
		if adjExp >= 0 {
			result.WriteByte('+')
		}
		result.WriteString(strconv.Itoa(adjExp))
	} else {
		// 일반 표기법
		if exp >= 0 {
			result.WriteString(coeffStr)
			for i := 0; i < exp; i++ {
				result.WriteByte('0')
			}
		} else {
			decimalPos := coeffLen + exp
			if decimalPos <= 0 {
				result.WriteString("0.")
				for i := 0; i < -decimalPos; i++ {
					result.WriteByte('0')
				}
				result.WriteString(coeffStr)
			} else {
				result.WriteString(coeffStr[:decimalPos])
				result.WriteByte('.')
				result.WriteString(coeffStr[decimalPos:])
			}
		}
	}

	return result.String()
}

// uitoa32 converts uint32 to string
func uitoa32(n uint32) string {
	if n == 0 {
		return "0"
	}
	var buf [10]byte
	i := 10
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// ParseDecimal32Pure parses a string into Decimal32Pure
func ParseDecimal32Pure(s string) (Decimal32Pure, error) {
	return ParseDecimal32PureWithMode(s, RoundNearestEven)
}

// ParseDecimal32PureWithMode parses a string into Decimal32Pure with specified rounding mode
func ParseDecimal32PureWithMode(s string, mode RoundingMode) (Decimal32Pure, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, errInvalidFormat32
	}

	// 부호 처리
	sign := 0
	if s[0] == '-' {
		sign = 1
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}

	// 특수값 처리
	upper := strings.ToUpper(s)
	if upper == "INF" || upper == "INFINITY" {
		r := bid32InfinityMask
		if sign != 0 {
			r |= bid32SignMask
		}
		return Decimal32Pure(r), nil
	}
	if upper == "NAN" || upper == "QNAN" {
		r := bid32NaNMask
		if sign != 0 {
			r |= bid32SignMask
		}
		return Decimal32Pure(r), nil
	}
	if upper == "SNAN" {
		r := bid32SNaNMask
		if sign != 0 {
			r |= bid32SignMask
		}
		return Decimal32Pure(r), nil
	}

	// 일반 숫자 파싱
	var coefficient uint64
	var exponent int
	var hasDecimal bool
	var decimalPos int
	var digits int

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			if digits < 16 { // 충분한 정밀도
				coefficient = coefficient*10 + uint64(c-'0')
				digits++
			} else {
				exponent++
			}
			if hasDecimal {
				decimalPos++
			}
		} else if c == '.' {
			if hasDecimal {
				return 0, errInvalidFormat32
			}
			hasDecimal = true
		} else if c == 'e' || c == 'E' {
			// 지수 파싱
			expStr := s[i+1:]
			exp, err := strconv.Atoi(expStr)
			if err != nil {
				return 0, errInvalidFormat32
			}
			exponent += exp
			break
		} else {
			return 0, errInvalidFormat32
		}
	}

	if hasDecimal {
		exponent -= decimalPos
	}

	// 0 처리
	if coefficient == 0 {
		return Decimal32Pure(encodeBID32(sign, bid32ExponentBias, 0)), nil
	}

	// 7자리로 정규화
	for coefficient > uint64(bid32MaxCoefficient) {
		// 반올림 처리
		lastDigit := coefficient % 10
		coefficient /= 10
		exponent++
		if mode == RoundNearestEven || mode == RoundNearestAway {
			if lastDigit > 5 || (lastDigit == 5 && mode == RoundNearestAway) {
				coefficient++
			} else if lastDigit == 5 && mode == RoundNearestEven && (coefficient%2) == 1 {
				coefficient++
			}
		}
	}

	// 지수 바이어스 적용
	biasedExp := exponent + bid32ExponentBias

	// 오버플로우/언더플로우 체크
	if biasedExp > bid32MaxExponent {
		// 오버플로우 → Infinity
		r := bid32InfinityMask
		if sign != 0 {
			r |= bid32SignMask
		}
		return Decimal32Pure(r), nil
	}
	if biasedExp < 0 {
		// 언더플로우 처리 (subnormal 또는 0)
		for biasedExp < 0 && coefficient > 0 {
			coefficient /= 10
			biasedExp++
		}
		if coefficient == 0 {
			return Decimal32Pure(encodeBID32(sign, 0, 0)), nil
		}
	}

	return Decimal32Pure(encodeBID32(sign, biasedExp, uint32(coefficient))), nil
}

// toDecimal64 converts Decimal32Pure to Decimal64Pure for extended precision operations
func (d Decimal32Pure) toDecimal64() Decimal64Pure {
	x := uint32(d)

	// NaN 처리
	if (x & bid32NaNMask) == bid32NaNMask {
		if (x & bid32SNaNMask) == bid32SNaNMask {
			// sNaN
			r := bid64SNaNMask
			if (x & bid32SignMask) != 0 {
				r |= bid64SignMask
			}
			return Decimal64Pure(r)
		}
		// qNaN
		r := bid64NaNMask
		if (x & bid32SignMask) != 0 {
			r |= bid64SignMask
		}
		return Decimal64Pure(r)
	}

	// Infinity 처리
	if (x & bid32NaNMask) == bid32InfinityMask {
		r := bid64InfinityMask
		if (x & bid32SignMask) != 0 {
			r |= bid64SignMask
		}
		return Decimal64Pure(r)
	}

	// 일반 숫자
	sign, exp, coeff, _ := unpack_BID32(x)

	// 지수 변환: bid32 bias(101) -> bid64 bias(398)
	exp64 := exp - bid32ExponentBias + bid64ExponentBias

	signBit := 0
	if sign != 0 {
		signBit = 1
	}

	return Decimal64Pure(encodeBID64(signBit, exp64, uint64(coeff)))
}

// fromDecimal64 converts Decimal64Pure to Decimal32Pure with rounding
func fromDecimal64(d Decimal64Pure, mode RoundingMode) Decimal32Pure {
	res, _ := Bid64ToBid32(uint64(d), roundingModeToBID(mode))
	return Decimal32Pure(res)
}

// Add returns a + b
func (a Decimal32Pure) Add(b Decimal32Pure) Decimal32Pure {
	return a.AddWithMode(b, RoundNearestEven)
}

// AddWithMode returns a + b with specified rounding mode
func (a Decimal32Pure) AddWithMode(b Decimal32Pure, mode RoundingMode) Decimal32Pure {
	rndMode := roundingModeToBID(mode)
	return Decimal32Pure(bid32_add_pure(uint32(a), uint32(b), rndMode))
}

// Sub returns a - b
func (a Decimal32Pure) Sub(b Decimal32Pure) Decimal32Pure {
	return a.SubWithMode(b, RoundNearestEven)
}

// SubWithMode returns a - b with specified rounding mode
func (a Decimal32Pure) SubWithMode(b Decimal32Pure, mode RoundingMode) Decimal32Pure {
	rndMode := roundingModeToBID(mode)
	return Decimal32Pure(bid32_sub_pure(uint32(a), uint32(b), rndMode))
}

// Mul returns a * b
func (a Decimal32Pure) Mul(b Decimal32Pure) Decimal32Pure {
	return a.MulWithMode(b, RoundNearestEven)
}

// MulWithMode returns a * b with specified rounding mode
func (a Decimal32Pure) MulWithMode(b Decimal32Pure, mode RoundingMode) Decimal32Pure {
	rndMode := roundingModeToBID(mode)
	return Decimal32Pure(bid32_mul_pure(uint32(a), uint32(b), rndMode))
}

// Div returns a / b
func (a Decimal32Pure) Div(b Decimal32Pure) Decimal32Pure {
	return a.DivWithMode(b, RoundNearestEven)
}

// DivWithMode returns a / b with specified rounding mode
func (a Decimal32Pure) DivWithMode(b Decimal32Pure, mode RoundingMode) Decimal32Pure {
	rndMode := roundingModeToBID(mode)
	return Decimal32Pure(bid32_div_pure(uint32(a), uint32(b), rndMode))
}

// Cmp compares a and b: returns -1 if a < b, 0 if a == b, 1 if a > b
func (a Decimal32Pure) Cmp(b Decimal32Pure) int {
	a64 := a.toDecimal64()
	b64 := b.toDecimal64()
	return a64.Cmp(b64)
}

// Equal returns true if a == b
func (a Decimal32Pure) Equal(b Decimal32Pure) bool {
	return a.Cmp(b) == 0
}

// Less returns true if a < b
func (a Decimal32Pure) Less(b Decimal32Pure) bool {
	return a.Cmp(b) < 0
}
