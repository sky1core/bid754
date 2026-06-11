// bid64_pure.go - Pure Go implementation of IEEE 754-2019 decimal64 BID encoding
//
// This file provides a CGO-free implementation of decimal64 arithmetic using
// Binary Integer Decimal (BID) encoding. It can be used in environments where
// CGO is not available or for comparison/testing purposes.
//
// Key features:
// - No CGO dependency
// - 128-bit arithmetic for full precision
// - IEEE 754-2019 compliant operations (Add, Sub, Mul, Div)
// - Compatible with Intel BID backend results
//
// Note: Cohort selection may differ from Intel BID, but mathematical results
// are equivalent.

package bidgo

import (
	"math"
	"math/bits"
	"strconv"
	"strings"
)

// Decimal64Pure is a pure Go implementation of IEEE 754-2019 decimal64 with BID encoding.
// It is binary compatible with Decimal64BID and can be used interchangeably for arithmetic.
type Decimal64Pure uint64

// BID64 인코딩 상수
const (
	bid64SignMask            uint64 = 0x8000000000000000 // 부호 비트
	bid64SpecialEncodingMask uint64 = 0x6000000000000000 // 특수 인코딩 (coefficient >= 2^53)
	bid64InfinityMask        uint64 = 0x7800000000000000 // Infinity
	bid64NaNMask             uint64 = 0x7c00000000000000 // NaN
	bid64SNaNMask            uint64 = 0x7e00000000000000 // Signaling NaN pattern
	bid64QuietBit            uint64 = 0x0200000000000000 // QNaN vs SNaN 구분 비트
	bid64LargeCoeffMask      uint64 = 0x0007ffffffffffff // 51비트 coefficient (특수 인코딩)
	bid64SmallCoeffMask      uint64 = 0x001fffffffffffff // 53비트 coefficient (일반 인코딩)
	bid64LargeCoeffHighBit   uint64 = 0x0020000000000000 // 2^53 (특수 인코딩 시 암묵적)
	bid64ExponentMask        uint64 = 0x3ff              // 10비트 지수
	bid64ExponentBias        int    = 398                // 지수 바이어스
	bid64MaxExponent         int    = 767                // 최대 지수
	bid64MaxCoefficient      uint64 = 9999999999999999   // 10^16 - 1

	// Intel BID 상수 (bid_internal.h) - 기계적 포팅용 별칭
	bid64LargestBID64     uint64 = 0x77fb86f26fc0ffff // LARGEST_BID64
	bid64InfMask          uint64 = 0x7800000000000000 // MASK_INF
	bid64SteeringBitsMask uint64 = 0x6000000000000000 // MASK_STEERING_BITS
	bid64BinaryExp1Mask   uint64 = 0x7fe0000000000000 // MASK_BINARY_EXPONENT1
	bid64BinaryExp2Mask   uint64 = 0x1ff8000000000000 // MASK_BINARY_EXPONENT2
	bid64BinarySig1Mask   uint64 = 0x001fffffffffffff // MASK_BINARY_SIG1
	bid64BinarySig2Mask   uint64 = 0x0007ffffffffffff // MASK_BINARY_SIG2
	bid64BinaryOr2        uint64 = 0x0020000000000000 // MASK_BINARY_OR2

	// 지수 시프트 위치
	bid64SmallExpShift = 53 // 일반 인코딩에서 지수 시프트
	bid64LargeExpShift = 51 // 특수 인코딩에서 지수 시프트
)

// getUnderflowResult returns the correct result for underflow based on rounding mode
// Returns smallest subnormal (±1E-398) or zero based on rounding direction
func getUnderflowResult(sign int, mode RoundingMode) Decimal64Pure {
	negative := sign != 0

	// 반올림 방향에 따라 최소 subnormal 또는 zero 반환
	switch mode {
	case RoundTowardPositive: // ceiling: +∞ 방향
		// 양수면 올림 → 최소 양수 subnormal
		if !negative {
			return Decimal64Pure(encodeBID64(0, 0, 1))
		}
		// 음수면 내림(0 방향) → -0
		return Decimal64Pure(encodeBID64(1, 0, 0))

	case RoundTowardNegative: // floor: -∞ 방향
		// 음수면 올림 → 최소 음수 subnormal
		if negative {
			return Decimal64Pure(encodeBID64(1, 0, 1))
		}
		// 양수면 내림(0 방향) → +0
		return Decimal64Pure(encodeBID64(0, 0, 0))

	default:
		// half_even, half_up, half_down, truncate: 0 반환
		return Decimal64Pure(encodeBID64(sign, 0, 0))
	}
}

// getOverflowResult returns the correct result for overflow based on rounding mode
// Ported from Intel fast_get_BID64_check_OF (bid_internal.h:1097-1117)
func getOverflowResult(sign int, mode RoundingMode) Decimal64Pure {
	sgn := uint64(0)
	if sign != 0 {
		sgn = bid64SignMask
	}

	// Default: Infinity
	r := sgn | bid64InfinityMask

	switch mode {
	case RoundTowardNegative: // BID_ROUNDING_DOWN (mode 1)
		// Round toward -∞: positive overflow → max positive finite
		if sgn == 0 {
			r = bid64LargestBID64
		}
	case RoundTowardZero: // BID_ROUNDING_TO_ZERO (mode 3)
		// Round toward 0: overflow → max finite with same sign
		r = sgn | bid64LargestBID64
	case RoundTowardPositive: // BID_ROUNDING_UP (mode 2)
		// Round toward +∞: negative overflow → max negative finite
		if sgn != 0 {
			r = bid64SignMask | bid64LargestBID64
		}
		// RoundNearestEven (mode 0) and RoundNearestAway (mode 4): Infinity (default)
	}

	return Decimal64Pure(r)
}

// 10의 거듭제곱 테이블 (0~19)
var pow10 = [...]uint64{
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
	10000000000000000,
	100000000000000000,
	1000000000000000000,
	10000000000000000000,
}

// BID64 subnormal 관련 상수
const (
	bid64MinExponent = 0    // 바이어스 적용된 최소 지수
	bid64Emin        = -383 // 실제 최소 지수 (1 - 398)
	bid64Etiny       = -398 // 가장 작은 지수 (subnormal)
)

// roundWithMode는 지정된 반올림 모드로 반올림을 수행
// coeff: 현재 계수, remainder: 나머지, divisor: 나눈 값
// mode: 반올림 모드, negative: 음수 여부 (ceiling/floor에 필요)
// 반환: 반올림된 계수
func roundWithMode(coeff uint64, remainder uint64, divisor uint64, mode RoundingMode, negative bool) uint64 {
	if remainder == 0 {
		return coeff // 정확한 값, 반올림 불필요
	}

	// 올바른 반올림 판단: remainder/divisor와 0.5 비교
	// remainder * 2와 divisor를 비교하면 됨 (오버플로 방지 필요)
	// remainder * 2 < divisor: 반 미만
	// remainder * 2 > divisor: 반 초과
	// remainder * 2 == divisor: 정확히 반 (짝수 divisor만 가능)
	var isExactlyHalf, isAboveHalf bool

	// 오버플로 방지: remainder가 매우 클 수 있음
	// remainder < divisor 이므로 remainder * 2는 2*divisor 미만
	// 하지만 remainder가 2^63 이상이면 오버플로 가능
	if remainder > (^uint64(0) >> 1) {
		// 오버플로 가능: 다른 방법 사용
		// remainder > divisor/2 이면 반 초과
		// (정수 나눗셈 고려)
		halfDiv := divisor / 2
		if divisor%2 == 0 {
			// 짝수: 정확히 half 계산 가능
			if remainder > halfDiv {
				isAboveHalf = true
			} else if remainder == halfDiv {
				isExactlyHalf = true
			}
		} else {
			// 홀수: halfDiv = (divisor-1)/2, 실제 half = halfDiv + 0.5
			// remainder > halfDiv 이면 반 초과 (remainder >= halfDiv+1 이면)
			if remainder > halfDiv {
				isAboveHalf = true
			}
			// 홀수 divisor에서 "정확히 반"은 불가능
		}
	} else {
		doubled := remainder * 2
		if doubled > divisor {
			isAboveHalf = true
		} else if doubled == divisor {
			isExactlyHalf = true
		}
	}

	switch mode {
	case RoundNearestEven: // half_even (은행원 반올림)
		if isAboveHalf {
			return coeff + 1
		} else if isExactlyHalf {
			// 정확히 반값 - 짝수로 반올림
			if (coeff & 1) == 1 {
				return coeff + 1
			}
		}
		return coeff

	case RoundNearestAway: // half_up (0.5를 0에서 멀어지는 방향)
		if isAboveHalf || isExactlyHalf {
			return coeff + 1
		}
		return coeff

	case RoundNearestDown: // half_down (0.5를 0 방향)
		if isAboveHalf {
			return coeff + 1
		}
		return coeff

	case RoundTowardPositive: // ceiling (양의 무한대 방향)
		if !negative {
			return coeff + 1 // 양수면 올림
		}
		return coeff // 음수면 내림 (0 방향)

	case RoundTowardNegative: // floor (음의 무한대 방향)
		if negative {
			return coeff + 1 // 음수면 올림 (0에서 멀어짐)
		}
		return coeff // 양수면 내림

	case RoundTowardZero: // down (truncation)
		return coeff // 항상 내림

	default:
		// 기본값: half_even
		if isAboveHalf {
			return coeff + 1
		} else if isExactlyHalf && (coeff&1) == 1 {
			return coeff + 1
		}
		return coeff
	}
}

// roundToNearestEven은 round-to-nearest-even (은행원 반올림)을 수행 - 호환성 유지
// coeff: 현재 계수, remainder: 나머지, divisor: 나눈 값
// 반환: 반올림된 계수
func roundToNearestEven(coeff uint64, remainder uint64, divisor uint64) uint64 {
	return roundWithMode(coeff, remainder, divisor, RoundNearestEven, false)
}

// divideAndRound는 128비트 (hi:lo)를 divisor로 나누고 round-to-nearest-even 적용
func divideAndRound128(hi, lo, divisor uint64) (quotient uint64, newHi uint64) {
	if hi >= divisor {
		// 오버플로 - 먼저 hi를 나눔
		hiQ := hi / divisor
		hiR := hi % divisor
		loQ, loR := bits.Div64(hiR, lo, divisor)
		quotient = loQ
		newHi = hiQ
		// 반올림
		quotient = roundToNearestEven(quotient, loR, divisor)
		if quotient == 0 && newHi > 0 {
			// carry 처리
		}
	} else {
		q, r := bits.Div64(hi, lo, divisor)
		quotient = roundToNearestEven(q, r, divisor)
		newHi = 0
	}
	return
}

// normalizeCoefficient는 계수를 16자리 이하로 정규화하면서 round-to-nearest-even 적용
func normalizeCoefficient(coeff uint64, exp int) (uint64, int) {
	for coeff > bid64MaxCoefficient {
		remainder := coeff % 10
		coeff = coeff / 10
		coeff = roundToNearestEven(coeff, remainder, 10)
		exp++
	}
	return coeff, exp
}

// maximizeCoefficient는 계수를 16자리까지 최대화하고 지수를 최소화
// IEEE 754 preferred exponent: 가능한 가장 작은 지수 사용 (계수 최대화)
func maximizeCoefficient(coeff uint64, exp int) (uint64, int) {
	if coeff == 0 {
		return coeff, exp
	}
	// 계수를 16자리까지 확장 (지수가 허용하는 한)
	for coeff < pow10[15] && exp > 0 {
		coeff *= 10
		exp--
	}
	return coeff, exp
}

// rescaleToSmaller는 값을 더 작은 지수로 스케일 (coefficient를 키움)
// targetExp가 currentExp보다 작아야 함
func rescaleToSmaller(original uint64, sign int, coeff uint64, currentExp, targetExp int) uint64 {
	diff := currentExp - targetExp
	if diff <= 0 {
		return original
	}
	// coefficient를 10^diff로 곱함 (16자리 제한 내에서)
	for i := 0; i < diff; i++ {
		if coeff > bid64MaxCoefficient/10 {
			// 16자리 초과 - 더 이상 스케일 불가
			// 현재까지의 스케일로 인코딩
			return encodeBID64(sign, currentExp-i, coeff)
		}
		coeff *= 10
	}
	return encodeBID64(sign, targetExp, coeff)
}

// decodeBID64는 BID64를 부호, 지수, 계수로 디코딩
// 반환: sign (0 또는 1), exponent (바이어스 적용 전), coefficient, isSpecial (Inf/NaN)
func decodeBID64(x uint64) (sign int, exponent int, coefficient uint64, isInf bool, isNaN bool, isSNaN bool) {
	sign = int((x & bid64SignMask) >> 63)

	// Infinity/NaN 체크
	if (x & bid64InfinityMask) == bid64InfinityMask {
		if (x & bid64NaNMask) == bid64NaNMask {
			isNaN = true
			if (x & bid64SNaNMask) == bid64SNaNMask {
				isSNaN = true
			}
			return
		}
		isInf = true
		return
	}

	// 특수 인코딩 체크 (coefficient >= 2^53)
	if (x & bid64SpecialEncodingMask) == bid64SpecialEncodingMask {
		// 특수 인코딩: 상위 2비트가 11
		coefficient = (x & bid64LargeCoeffMask) | bid64LargeCoeffHighBit
		exponent = int((x >> bid64LargeExpShift) & bid64ExponentMask)
	} else {
		// 일반 인코딩
		coefficient = x & bid64SmallCoeffMask
		exponent = int((x >> bid64SmallExpShift) & bid64ExponentMask)
	}

	// Non-canonical 체크: coefficient > 10^16-1 이면 0으로 처리
	// IEEE 754-2019: non-canonical 인코딩의 값은 0 (부호는 유지)
	if coefficient > bid64MaxCoefficient {
		coefficient = 0
		// 부호는 유지, isNaN/isInf는 false 유지 (일반 0으로 처리)
	}

	return
}

// encodeBID64는 부호, 지수, 계수를 BID64로 인코딩
func encodeBID64(sign int, exponent int, coefficient uint64) uint64 {
	return encodeBID64WithMode(sign, exponent, coefficient, RoundNearestEven)
}

// encodeBID64WithMode encodes BID64 with specified rounding mode for subnormal handling
func encodeBID64WithMode(sign int, exponent int, coefficient uint64, mode RoundingMode) uint64 {
	var r uint64
	negative := sign != 0

	if negative {
		r = bid64SignMask
	}

	// 지수 범위 체크
	if exponent > bid64MaxExponent {
		// Overflow -> Infinity
		return r | bid64InfinityMask
	}

	// Underflow 처리: subnormal 지원
	// BID64에서 최소 바이어스 지수는 0, 실제 지수는 -398
	if exponent < 0 {
		// 지수를 0으로 올리면서 coefficient를 줄임 (subnormal)
		shift := -exponent
		if shift > 19 {
			// 너무 작음 -> 0
			return r
		}

		// Double-rounding 방지: 단일 반올림만 적용
		// coefficient를 10^shift로 나누고 마지막에만 반올림
		var sticky uint64 = 0
		var lastRem uint64 = 0
		for i := 0; i < shift && coefficient > 0; i++ {
			if lastRem != 0 {
				sticky = 1 // 이전 나머지가 있었으면 sticky bit 설정
			}
			lastRem = coefficient % 10
			coefficient = coefficient / 10
		}

		// 반올림 모드에 따른 처리
		if lastRem > 0 || sticky > 0 {
			// sticky가 있으면 실제 나머지가 더 있음을 반영
			adjustedRem := lastRem
			if sticky != 0 {
				if lastRem == 0 {
					adjustedRem = 1 // 0.xxx > 0이므로 1로 처리 (directed rounding용)
				} else if lastRem == 5 {
					adjustedRem = 6 // 5.xxx > 5이므로 6으로 처리
				}
			}
			coefficient = roundWithMode(coefficient, adjustedRem, 10, mode, negative)
		}

		if coefficient == 0 {
			return r // 언더플로 -> 0
		}
		exponent = 0
	}

	// coefficient가 10^16 이상이면 오버플로
	if coefficient > bid64MaxCoefficient {
		return r | bid64InfinityMask
	}

	// coefficient가 2^53 이상이면 특수 인코딩 사용
	if coefficient >= bid64LargeCoeffHighBit {
		// 특수 인코딩
		r |= bid64SpecialEncodingMask
		r |= (coefficient & bid64LargeCoeffMask)
		r |= uint64(exponent) << bid64LargeExpShift
	} else {
		// 일반 인코딩
		r |= coefficient
		r |= uint64(exponent) << bid64SmallExpShift
	}

	return r
}

// encodeInfinity는 무한대를 인코딩
func encodeInfinity64(sign int) uint64 {
	if sign != 0 {
		return bid64SignMask | bid64InfinityMask
	}
	return bid64InfinityMask
}

// encodeNaN64는 NaN을 인코딩
func encodeNaN64(sign int, signaling bool) uint64 {
	var r uint64
	if signaling {
		r = bid64SNaNMask
	} else {
		r = bid64NaNMask
	}
	if sign != 0 {
		r |= bid64SignMask
	}
	return r
}

// quietNaN64는 sNaN을 qNaN으로 변환
// Intel BID 동작 분석 결과:
// - bits 49-48이 모두 1이면 비정규 → canonical qNaN 반환
// - 그 외에는 bits 49-0을 payload로 보존
func quietNaN64(x uint64) uint64 {
	sign := x & bid64SignMask

	// bits 49-48이 모두 1이면 비정규 NaN → canonical qNaN
	if (x & 0x0003000000000000) == 0x0003000000000000 {
		return sign | bid64NaNMask
	}

	// payload bits 49-0 보존
	payload := x & 0x0003ffffffffffff
	return sign | bid64NaNMask | payload
}

// adjustForTinyContribution은 작은 반대 부호 기여가 있을 때 반올림 모드에 따라 조정
// coeff: 지배적인 계수, exp: 지수, sign: 부호, mode: 반올림 모드
// 예: 100 - (아주작은양) = 100 - ε 에서, RoundDown이면 99.99...로 조정
func adjustForTinyContribution(coeff uint64, exp int, sign int, mode RoundingMode) Decimal64Pure {
	negative := sign != 0

	// 반올림 모드별 조정 결정
	shouldDecrease := false

	switch mode {
	case RoundTowardZero: // down: 0 방향으로
		// 양수든 음수든 절대값이 작아지는 방향 (magnitude 감소)
		shouldDecrease = true

	case RoundTowardNegative: // floor: 음의 무한대 방향
		// 양수면 값이 작아져야 함
		if !negative {
			shouldDecrease = true
		}

	case RoundTowardPositive: // ceiling: 양의 무한대 방향
		// 음수면 값이 커져야 함 (절대값 감소)
		if negative {
			shouldDecrease = true
		}

	default:
		// half_even, half_up, half_down 등: 작은 기여는 무시됨
		shouldDecrease = false
	}

	if shouldDecrease {
		// 1 ULP 감소
		// 계수가 10^n 형태면 1 빼면 자릿수가 줄어듬
		// 예: 1000000000000000 - 1 = 999999999999999 (16자리 → 15자리)
		// 올바른 1 ULP: 10000000000000000 - 1 = 9999999999999999, exp 조정

		// 계수의 자릿수 확인
		digitsCoeff := countDigits(coeff)

		// 10의 거듭제곱인지 확인 (예: 1, 10, 100, ...)
		isPowerOf10 := (coeff == pow10[digitsCoeff-1])

		if isPowerOf10 && digitsCoeff > 1 {
			// 10^n에서 1 ULP 감소: 10^(n+1) - 1로 변환, exp 조정
			// 예: 1e15 × 10^e → (1e16-1) × 10^(e-1)
			coeff = pow10[digitsCoeff] - 1
			exp--
		} else if coeff > 1 {
			coeff--
		} else {
			// coeff == 1이고 1자리: 0이 됨 → 0.9999... 로 변환
			coeff = bid64MaxCoefficient
			exp--
		}
	}

	// 지수 범위 확인
	if exp < 0 {
		exp = 0
	}

	return Decimal64Pure(encodeBID64(sign, exp, coeff))
}

// adjustForTinyContributionAdd는 같은 부호일 때 작은 기여로 인한 올림을 처리
// adjustForTinyContribution과 반대: 뺄셈이 아닌 덧셈에서 작은 기여 처리
func adjustForTinyContributionAdd(coeff uint64, exp int, sign int, mode RoundingMode) Decimal64Pure {
	negative := sign != 0

	// 반올림 모드별 올림 결정
	shouldIncrease := false

	switch mode {
	case RoundTowardPositive: // ceiling: 양의 무한대 방향
		// 양수면 값이 커져야 함
		if !negative {
			shouldIncrease = true
		}

	case RoundTowardNegative: // floor: 음의 무한대 방향
		// 음수면 절대값이 커져야 함 (더 음수 방향)
		if negative {
			shouldIncrease = true
		}

	case RoundTowardZero: // down: 0 방향으로
		// 절대값 증가 안함
		shouldIncrease = false

	default:
		// half_even, half_up, half_down 등: 작은 기여는 half 미만이므로 무시
		shouldIncrease = false
	}

	if shouldIncrease {
		// 1 ULP 증가
		coeff++
		// 오버플로 처리
		if coeff > bid64MaxCoefficient {
			coeff = coeff / 10
			exp++
		}
	}

	// 지수 범위 확인 (max biased exp = 767)
	if exp > bid64MaxExponent {
		// 오버플로 - 반올림 모드에 따라 처리
		return getOverflowResult(sign, mode)
	}

	return Decimal64Pure(encodeBID64(sign, exp, coeff))
}

// Add는 두 Decimal64Pure 값을 더함 (기본 반올림: half_even)
func (a Decimal64Pure) Add(b Decimal64Pure) Decimal64Pure {
	return a.AddWithMode(b, RoundNearestEven)
}

// AddWithMode는 지정된 반올림 모드로 두 Decimal64Pure 값을 더함
func (a Decimal64Pure) AddWithMode(b Decimal64Pure, mode RoundingMode) Decimal64Pure {
	// 디코딩
	signA, expA, coeffA, infA, nanA, snanA := decodeBID64(uint64(a))
	signB, expB, coeffB, infB, nanB, snanB := decodeBID64(uint64(b))

	// NaN 처리 - Intel BID 규칙: payload 보존하면서 qNaN 반환
	if snanA || nanA {
		return Decimal64Pure(quietNaN64(uint64(a)))
	}
	if snanB || nanB {
		return Decimal64Pure(quietNaN64(uint64(b)))
	}

	// Infinity 처리 - Intel BID는 canonical infinity 반환
	if infA || infB {
		if infA && infB {
			if signA != signB {
				return Decimal64Pure(encodeNaN64(0, false)) // Inf - Inf = NaN
			}
			return Decimal64Pure(encodeInfinity64(signA))
		}
		if infA {
			return Decimal64Pure(encodeInfinity64(signA))
		}
		return Decimal64Pure(encodeInfinity64(signB))
	}

	// 0 처리 - IEEE 754 preferred exponent: min(exp1, exp2)
	if coeffA == 0 && coeffB == 0 {
		// 0 + 0 = 0, 작은 지수 사용 (IEEE 754)
		resultExp := expA
		if expB < expA {
			resultExp = expB
		}
		// IEEE 754 signed zero 규칙 (Intel Bid64Add.c:206-214)
		// - 부호가 같으면: 그 부호 사용
		// - 부호가 다르면 (round-to-nearest): +0
		// - 부호가 다르면 (floor 모드): -0
		resultSign := 0
		if signA == signB {
			resultSign = signA
		} else if mode == RoundTowardNegative {
			resultSign = 1
		}
		return Decimal64Pure(encodeBID64(resultSign, resultExp, 0))
	}
	// 0이 아닌 쪽이 결과가 되지만, preferred exponent는 min(expA, expB)
	if coeffA == 0 {
		// 0 + b = b, but with preferred exponent = min(expA, expB)
		if expA < expB {
			// B를 더 작은 지수로 스케일
			return Decimal64Pure(rescaleToSmaller(uint64(b), signB, coeffB, expB, expA))
		}
		return b
	}
	if coeffB == 0 {
		// a + 0 = a, but with preferred exponent = min(expA, expB)
		if expB < expA {
			// A를 더 작은 지수로 스케일
			return Decimal64Pure(rescaleToSmaller(uint64(a), signA, coeffA, expA, expB))
		}
		return a
	}

	// IEEE 754 덧셈: 작은 지수로 정렬하고 128비트 중간 계산 사용
	// 목표: 최대 정밀도 유지 후 16자리로 반올림

	// 작은 지수 결정 (quantum)
	var resultExp int
	var hiA, loA, hiB, loB uint64 // 128비트 표현

	if expA <= expB {
		resultExp = expA
		loA = coeffA
		hiA = 0
		diff := expB - expA

		// 지수 차이가 크면 먼저 양쪽 정규화
		if diff > 17 {
			// 양쪽 모두 16자리로 정규화
			normCoeffA, normExpA := maximizeCoefficient(coeffA, expA)
			normCoeffB, normExpB := maximizeCoefficient(coeffB, expB)
			newDiff := normExpB - normExpA

			if newDiff > 17 {
				// 정규화 후에도 차이가 크면 B가 지배
				if signA != signB && coeffA > 0 {
					// A가 B와 반대 부호로 기여 -> 뺄셈 효과 (조정)
					return adjustForTinyContribution(normCoeffB, normExpB, signB, mode)
				}
				if signA == signB && coeffA > 0 {
					// A가 B와 같은 부호로 기여 -> 덧셈 효과 (올림 가능)
					return adjustForTinyContributionAdd(normCoeffB, normExpB, signB, mode)
				}
				return Decimal64Pure(encodeBID64(signB, normExpB, normCoeffB))
			}

			// 정규화된 값으로 계속
			coeffA, expA = normCoeffA, normExpA
			coeffB, expB = normCoeffB, normExpB
			resultExp = expA
			loA = coeffA
			diff = newDiff
		}

		hiB, loB = scaleUp128(coeffB, diff)
	} else {
		resultExp = expB
		loB = coeffB
		hiB = 0
		diff := expA - expB

		// 지수 차이가 크면 먼저 양쪽 정규화
		if diff > 17 {
			normCoeffA, normExpA := maximizeCoefficient(coeffA, expA)
			normCoeffB, normExpB := maximizeCoefficient(coeffB, expB)
			newDiff := normExpA - normExpB

			if newDiff > 17 {
				// A가 지배
				if signA != signB && coeffB > 0 {
					// B가 A와 반대 부호로 기여 -> 뺄셈 효과 (조정)
					return adjustForTinyContribution(normCoeffA, normExpA, signA, mode)
				}
				if signA == signB && coeffB > 0 {
					// B가 A와 같은 부호로 기여 -> 덧셈 효과 (올림 가능)
					return adjustForTinyContributionAdd(normCoeffA, normExpA, signA, mode)
				}
				return Decimal64Pure(encodeBID64(signA, normExpA, normCoeffA))
			}

			coeffA, expA = normCoeffA, normExpA
			coeffB, expB = normCoeffB, normExpB
			resultExp = expB
			loB = coeffB
			diff = newDiff
		}

		hiA, loA = scaleUp128(coeffA, diff)
	}

	// 부호 적용 후 128비트 덧셈/뺄셈
	var resultSign int
	var hiR, loR uint64

	// 부호에 따른 연산
	if signA == signB {
		// 같은 부호: 덧셈
		resultSign = signA
		loR, hiR = add128(hiA, loA, hiB, loB)
	} else {
		// 다른 부호: 뺄셈 (큰 것에서 작은 것 빼기)
		cmp := cmp128(hiA, loA, hiB, loB)
		if cmp >= 0 {
			resultSign = signA
			loR, hiR = sub128(hiA, loA, hiB, loB)
		} else {
			resultSign = signB
			loR, hiR = sub128(hiB, loB, hiA, loA)
		}
	}

	// 결과가 0인 경우 - IEEE 754 signed zero 규칙
	// Intel Bid64Add.c:206-214 참조:
	// - 부호가 같으면: 그 부호 사용
	// - 부호가 다르면 (round-to-nearest): +0
	// - 부호가 다르면 (floor 모드): -0
	if hiR == 0 && loR == 0 {
		zeroSign := 0
		if signA == signB {
			zeroSign = signA
		} else if mode == RoundTowardNegative {
			// floor 모드에서는 다른 부호일 때 -0
			zeroSign = 1
		}
		return Decimal64Pure(encodeBID64(zeroSign, resultExp, 0))
	}

	// 128비트 결과를 16자리로 정규화
	resultCoeff := normalize128to64WithMode(hiR, loR, &resultExp, mode, resultSign != 0)

	// 언더플로/오버플로 체크
	if resultExp > bid64MaxExponent {
		return getOverflowResult(resultSign, mode)
	}

	return Decimal64Pure(encodeBID64(resultSign, resultExp, resultCoeff))
}

// Sub는 두 Decimal64Pure 값을 뺌 (기본 반올림: half_even)
func (a Decimal64Pure) Sub(b Decimal64Pure) Decimal64Pure {
	return a.SubWithMode(b, RoundNearestEven)
}

// SubWithMode는 지정된 반올림 모드로 두 Decimal64Pure 값을 뺌
func (a Decimal64Pure) SubWithMode(b Decimal64Pure, mode RoundingMode) Decimal64Pure {
	// Intel BID 규칙: NaN이 아니면 y의 부호를 반전하고 Add 호출
	// NaN은 부호 반전 없이 전달

	// NaN 체크 - NaN이면 부호 반전 없이 Add로 전달
	if (uint64(b) & bid64NaNMask) == bid64NaNMask {
		return a.AddWithMode(b, mode)
	}

	// b의 부호 반전하고 Add 호출
	negB := uint64(b) ^ bid64SignMask
	return a.AddWithMode(Decimal64Pure(negB), mode)
}

// Mul은 두 Decimal64Pure 값을 곱함 (기본 반올림: half_even)
func (a Decimal64Pure) Mul(b Decimal64Pure) Decimal64Pure {
	return a.MulWithMode(b, RoundNearestEven)
}

// MulWithMode는 지정된 반올림 모드로 두 Decimal64Pure 값을 곱함
func (a Decimal64Pure) MulWithMode(b Decimal64Pure, mode RoundingMode) Decimal64Pure {
	signA, expA, coeffA, infA, nanA, snanA := decodeBID64(uint64(a))
	signB, expB, coeffB, infB, nanB, snanB := decodeBID64(uint64(b))

	resultSign := signA ^ signB
	negative := resultSign != 0

	// NaN 처리 - Intel BID 규칙: payload 보존하면서 qNaN 반환
	if snanA || nanA {
		return Decimal64Pure(quietNaN64(uint64(a)))
	}
	if snanB || nanB {
		return Decimal64Pure(quietNaN64(uint64(b)))
	}

	// Infinity 처리
	if infA || infB {
		if infA && (coeffB == 0 && !infB) {
			return Decimal64Pure(encodeNaN64(0, false)) // Inf * 0 = NaN
		}
		if infB && (coeffA == 0 && !infA) {
			return Decimal64Pure(encodeNaN64(0, false)) // 0 * Inf = NaN
		}
		return Decimal64Pure(encodeInfinity64(resultSign))
	}

	// 0 체크 - IEEE 754 preferred exponent for Mul: exp1 + exp2
	if coeffA == 0 || coeffB == 0 {
		// preferred exponent = expA + expB - 2*bias = (expA - bias) + (expB - bias)
		resultExp := (expA - bid64ExponentBias) + (expB - bid64ExponentBias)
		biasedExp := resultExp + bid64ExponentBias
		// 지수 범위 확인
		if biasedExp < 0 {
			biasedExp = 0
		}
		if biasedExp > bid64MaxExponent {
			biasedExp = bid64MaxExponent
		}
		return Decimal64Pure(encodeBID64(resultSign, biasedExp, 0))
	}

	// 곱셈 (128비트 결과 가능)
	hi, lo := bits.Mul64(coeffA, coeffB)

	// 결과 지수
	resultExp := (expA - bid64ExponentBias) + (expB - bid64ExponentBias)

	// 128비트 결과를 정규화
	// IEEE 754: 중간 단계는 truncation, 최종 단계에서만 rounding
	var lastRemainder uint64
	var sticky uint64 // 중간 단계에서 버린 비트가 있는지 추적

	// Phase 1: 16자리까지 정규화
	for hi > 0 || lo > bid64MaxCoefficient {
		hiQ := hi / 10
		hiR := hi % 10
		loQ, loR := bits.Div64(hiR, lo, 10)

		if lastRemainder != 0 {
			sticky = 1
		}
		lastRemainder = loR

		hi = hiQ
		lo = loQ
		resultExp++
	}

	// Phase 2: 언더플로 처리 - 지수가 음수면 추가 나눗셈
	biasedExp := resultExp + bid64ExponentBias

	// Intel Bid64Mul.c:212 - 심각한 언더플로 체크
	// biasedExp + 16 < 0이면 결과가 너무 작아서 16자리 스케일업으로도 복구 불가
	// 하지만 반올림 모드에 따라 최소 subnormal 반환 필요
	if biasedExp+16 < 0 {
		// lo > 0이면 non-zero 곱셈 결과가 있었음 → 반올림 모드 적용
		if lo > 0 || lastRemainder > 0 || sticky > 0 {
			return getUnderflowResult(resultSign, mode)
		}
		return Decimal64Pure(encodeBID64(resultSign, 0, 0))
	}

	for biasedExp < 0 && lo > 0 {
		if lastRemainder != 0 {
			sticky = 1
		}
		lastRemainder = lo % 10
		lo = lo / 10
		biasedExp++
	}

	// Phase 3: 오버플로 처리 - Intel bid_internal.h의 fast_get_BID64_check_OF 참조
	// 계수가 작고 지수가 너무 크면, 계수를 스케일업하고 지수를 줄임
	// (예: 1e+381 → 1000000000000000e+366)
	for lo < 1000000000000000 && biasedExp > bid64MaxExponent {
		lo = lo * 10
		biasedExp--
	}

	// 여전히 오버플로면 반올림 모드에 따라 처리
	if biasedExp > bid64MaxExponent {
		return getOverflowResult(resultSign, mode)
	}

	// 완전 언더플로 (계수가 0이 됨)
	if lo == 0 {
		// 언더플로에서도 반올림 적용
		adjustedRemainder := lastRemainder
		if sticky != 0 && lastRemainder == 5 {
			adjustedRemainder = 6
		}
		rounded := roundWithMode(0, adjustedRemainder, 10, mode, negative)
		if rounded > 0 {
			return Decimal64Pure(encodeBID64(resultSign, 0, 1))
		}
		return Decimal64Pure(encodeBID64(resultSign, 0, 0))
	}

	// 최종 반올림
	if lastRemainder > 0 || sticky > 0 {
		adjustedRemainder := lastRemainder
		if sticky != 0 {
			// sticky가 있으면 정확히 0이나 5가 아닌 0.xxx 또는 5.xxx임
			// lastRemainder가 0이면 0.xxx이므로 1로 (0보다 큼을 표시)
			// lastRemainder가 5이면 5.xxx이므로 6으로 (5보다 큼을 표시)
			if lastRemainder == 0 {
				adjustedRemainder = 1
			} else if lastRemainder == 5 {
				adjustedRemainder = 6
			}
		}
		lo = roundWithMode(lo, adjustedRemainder, 10, mode, negative)
		if lo > bid64MaxCoefficient {
			lo /= 10
			biasedExp++
			if biasedExp > bid64MaxExponent {
				return getOverflowResult(resultSign, mode)
			}
		}
	}

	return Decimal64Pure(encodeBID64WithMode(resultSign, biasedExp, lo, mode))
}

// Div는 두 Decimal64Pure 값을 나눔 (기본 반올림: half_even)
func (a Decimal64Pure) Div(b Decimal64Pure) Decimal64Pure {
	return a.DivWithMode(b, RoundNearestEven)
}

// DivWithMode는 지정된 반올림 모드로 두 Decimal64Pure 값을 나눔
// Intel Bid64Div.c의 기계적 포팅 사용
func (a Decimal64Pure) DivWithMode(b Decimal64Pure, mode RoundingMode) Decimal64Pure {
	return Decimal64Pure(Bid64Div(uint64(a), uint64(b), roundingModeToBID(mode)))
}

// countDigits는 숫자의 10진 자릿수를 반환
func countDigits(n uint64) int {
	if n == 0 {
		return 1
	}
	count := 0
	for n > 0 {
		n /= 10
		count++
	}
	return count
}

// mul64by64는 두 64비트 수를 곱해 128비트 결과 반환
func mul64by64(a, b uint64) (hi, lo uint64) {
	return bits.Mul64(a, b)
}

// scaleUp128은 64비트 값을 10^n으로 곱해 128비트 결과 반환
func scaleUp128(val uint64, n int) (hi, lo uint64) {
	if n == 0 {
		return 0, val
	}
	if n > 38 {
		// 오버플로
		return ^uint64(0), ^uint64(0)
	}

	// 10의 거듭제곱으로 곱셈 (단계별로)
	hi, lo = 0, val
	for n > 0 {
		step := n
		if step > 19 {
			step = 19
		}
		multiplier := pow10[step]
		// 128비트 곱셈: (hi:lo) * multiplier
		// bits.Mul64 반환: (hi, lo) 순서
		prodHi, prodLo := bits.Mul64(lo, multiplier)
		// hi * multiplier 도 더해야 함 (단, 64비트 내에서)
		hiProd := hi * multiplier
		hi = prodHi + hiProd
		lo = prodLo
		n -= step
	}
	return hi, lo
}

// add128은 두 128비트 수를 더함 (캐리 포함)
func add128(hiA, loA, hiB, loB uint64) (lo, hi uint64) {
	lo, carry := bits.Add64(loA, loB, 0)
	hi, _ = bits.Add64(hiA, hiB, carry)
	return lo, hi
}

// sub128은 128비트 뺄셈 (a - b, a >= b 가정)
func sub128(hiA, loA, hiB, loB uint64) (lo, hi uint64) {
	lo, borrow := bits.Sub64(loA, loB, 0)
	hi, _ = bits.Sub64(hiA, hiB, borrow)
	return lo, hi
}

// cmp128은 두 128비트 수를 비교 (a > b: 1, a == b: 0, a < b: -1)
func cmp128(hiA, loA, hiB, loB uint64) int {
	if hiA > hiB {
		return 1
	}
	if hiA < hiB {
		return -1
	}
	if loA > loB {
		return 1
	}
	if loA < loB {
		return -1
	}
	return 0
}

// normalize128to64WithMode는 128비트 값을 16자리 uint64로 정규화
// 지정된 반올림 모드 적용 (sticky bit 포함)
func normalize128to64WithMode(hi, lo uint64, exp *int, mode RoundingMode, negative bool) uint64 {
	// 128비트 값이 16자리(10^16) 이내면 그대로 반환
	if hi == 0 && lo <= bid64MaxCoefficient {
		return lo
	}

	// sticky bit: 이전 나눗셈에서 나머지가 있었는지 추적
	var sticky uint64

	// 10으로 반복 나눗셈 (지정된 반올림 모드 적용)
	for hi > 0 || lo > bid64MaxCoefficient {
		// 128비트 (hi:lo)를 10으로 나눔
		hiQ := hi / 10
		hiR := hi % 10
		loQ, loR := bits.Div64(hiR, lo, 10)

		// 결과 업데이트 (반올림 전)
		hi = hiQ
		lo = loQ

		// 마지막 iteration인지 확인
		if hi == 0 && lo <= bid64MaxCoefficient {
			// 마지막 반올림 - sticky bit 고려
			// sticky가 있으면 loR을 조정하여 실제 나머지가 0.xxx 또는 5.xxx임을 반영
			adjustedRemainder := loR
			if sticky != 0 {
				if loR == 0 {
					adjustedRemainder = 1 // 0.xxx는 0보다 크므로 1로 취급
				} else if loR == 5 {
					adjustedRemainder = 6 // 5.xxx는 5보다 크므로 6으로 취급
				}
			}

			// 반올림 적용
			roundedLo := roundWithMode(lo, adjustedRemainder, 10, mode, negative)
			if roundedLo != lo {
				lo = roundedLo
				if lo > bid64MaxCoefficient {
					lo /= 10
					*exp++
				}
			}
		} else {
			// 중간 iteration - sticky bit 업데이트만
			if loR != 0 {
				sticky = 1
			}
		}
		*exp++
	}

	return lo
}

// normalize128to64는 128비트 값을 16자리 uint64로 정규화
// IEEE 754 round-to-nearest-even 적용 (sticky bit 포함)
func normalize128to64(hi, lo uint64, exp *int) uint64 {
	return normalize128to64WithMode(hi, lo, exp, RoundNearestEven, false)
}

// String은 Decimal64Pure를 문자열로 변환
func (d Decimal64Pure) String() string {
	sign, exp, coeff, isInf, isNaN, _ := decodeBID64(uint64(d))

	if isNaN {
		if sign != 0 {
			return "-NaN"
		}
		return "NaN"
	}

	if isInf {
		if sign != 0 {
			return "-Infinity"
		}
		return "Infinity"
	}

	if coeff == 0 {
		if sign != 0 {
			return "-0"
		}
		return "0"
	}

	// 실제 지수
	realExp := exp - bid64ExponentBias

	// 계수를 문자열로
	coeffStr := strconv.FormatUint(coeff, 10)

	// 지수 조정 (소수점 위치)
	// 과학적 표기법 vs 일반 표기법 결정
	adjExp := realExp + len(coeffStr) - 1

	var result strings.Builder
	if sign != 0 {
		result.WriteByte('-')
	}

	if adjExp < -6 || adjExp > len(coeffStr)+5 {
		// 과학적 표기법
		result.WriteByte(coeffStr[0])
		if len(coeffStr) > 1 {
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
		pointPos := len(coeffStr) + realExp
		if pointPos <= 0 {
			result.WriteString("0.")
			for i := 0; i < -pointPos; i++ {
				result.WriteByte('0')
			}
			result.WriteString(coeffStr)
		} else if pointPos >= len(coeffStr) {
			result.WriteString(coeffStr)
			for i := 0; i < pointPos-len(coeffStr); i++ {
				result.WriteByte('0')
			}
		} else {
			result.WriteString(coeffStr[:pointPos])
			result.WriteByte('.')
			result.WriteString(coeffStr[pointPos:])
		}
	}

	return result.String()
}

// IsZero는 값이 0인지 확인 (부호 무시)
func (d Decimal64Pure) IsZero() bool {
	x := uint64(d)
	// NaN/Inf 체크
	if (x & bid64InfinityMask) == bid64InfinityMask {
		return false
	}
	// coefficient 추출
	var coeff uint64
	if (x & bid64SpecialEncodingMask) == bid64SpecialEncodingMask {
		coeff = (x & bid64LargeCoeffMask) | bid64LargeCoeffHighBit
	} else {
		coeff = x & bid64SmallCoeffMask
	}
	return coeff == 0
}

// IsNaN은 값이 NaN인지 확인
func (d Decimal64Pure) IsNaN() bool {
	return (uint64(d) & bid64NaNMask) == bid64NaNMask
}

// IsSNaN은 값이 Signaling NaN인지 확인
// SNaN은 NaN이면서 quiet 비트가 clear된 상태
func (d Decimal64Pure) IsSNaN() bool {
	x := uint64(d)
	return (x&bid64NaNMask) == bid64NaNMask && (x&bid64QuietBit) == 0
}

// IsInf는 값이 무한대인지 확인
func (d Decimal64Pure) IsInf() bool {
	x := uint64(d)
	return (x&bid64InfinityMask) == bid64InfinityMask && (x&bid64NaNMask) != bid64NaNMask
}

// Sign은 부호를 반환: -1 (음수/음수부호), 0 (영), 1 (양수/양수부호)
// NaN과 Infinity도 부호 비트에 따라 -1 또는 1을 반환
func (d Decimal64Pure) Sign() int {
	if d.IsZero() {
		return 0
	}
	if (uint64(d) & bid64SignMask) != 0 {
		return -1
	}
	return 1
}

// IsNegative는 부호 비트가 설정되어 있는지 확인 (음수 또는 -0, -Inf, -NaN)
func (d Decimal64Pure) IsNegative() bool {
	return (uint64(d) & bid64SignMask) != 0
}

// Cmp는 두 값을 비교: -1 (d < other), 0 (d == other), 1 (d > other)
// NaN이 포함되면 -2 반환
func (d Decimal64Pure) Cmp(other Decimal64Pure) int {
	// NaN 처리
	if d.IsNaN() || other.IsNaN() {
		return -2
	}

	x := uint64(d)
	y := uint64(other)

	// 무한대 처리
	xIsInf := d.IsInf()
	yIsInf := other.IsInf()
	xSign := (x & bid64SignMask) != 0
	ySign := (y & bid64SignMask) != 0

	if xIsInf && yIsInf {
		if xSign == ySign {
			return 0
		}
		if xSign {
			return -1
		}
		return 1
	}
	if xIsInf {
		if xSign {
			return -1
		}
		return 1
	}
	if yIsInf {
		if ySign {
			return 1
		}
		return -1
	}

	// 0 처리
	xIsZero := d.IsZero()
	yIsZero := other.IsZero()
	if xIsZero && yIsZero {
		return 0
	}
	if xIsZero {
		if ySign {
			return 1
		}
		return -1
	}
	if yIsZero {
		if xSign {
			return -1
		}
		return 1
	}

	// 부호가 다르면 간단
	if xSign != ySign {
		if xSign {
			return -1
		}
		return 1
	}

	// 같은 부호 - 디코딩해서 비교
	_, expX, coeffX, _, _, _ := decodeBID64(x)
	_, expY, coeffY, _, _, _ := decodeBID64(y)

	// 같은 지수로 정규화해서 비교
	// 지수 차이에 따라 계수 스케일링
	expDiff := expX - expY
	var cmp int
	if expDiff == 0 {
		if coeffX > coeffY {
			cmp = 1
		} else if coeffX < coeffY {
			cmp = -1
		} else {
			cmp = 0
		}
	} else if expDiff > 0 {
		// x의 지수가 더 큼 -> x의 계수를 늘려서 비교
		if expDiff > 20 {
			cmp = 1 // x가 확실히 더 큼
		} else {
			scaledX := coeffX
			for i := 0; i < expDiff && scaledX <= bid64MaxCoefficient; i++ {
				scaledX *= 10
			}
			if scaledX > coeffY {
				cmp = 1
			} else if scaledX < coeffY {
				cmp = -1
			} else {
				cmp = 0
			}
		}
	} else {
		// y의 지수가 더 큼 -> y의 계수를 늘려서 비교
		expDiff = -expDiff
		if expDiff > 20 {
			cmp = -1 // y가 확실히 더 큼
		} else {
			scaledY := coeffY
			for i := 0; i < expDiff && scaledY <= bid64MaxCoefficient; i++ {
				scaledY *= 10
			}
			if coeffX > scaledY {
				cmp = 1
			} else if coeffX < scaledY {
				cmp = -1
			} else {
				cmp = 0
			}
		}
	}

	// 음수면 결과 반전
	if xSign {
		return -cmp
	}
	return cmp
}

// Eq는 동등 비교
func (d Decimal64Pure) Eq(other Decimal64Pure) bool {
	return d.Cmp(other) == 0
}

// Lt는 미만 비교
func (d Decimal64Pure) Lt(other Decimal64Pure) bool {
	return d.Cmp(other) == -1
}

// Lte는 이하 비교
func (d Decimal64Pure) Lte(other Decimal64Pure) bool {
	c := d.Cmp(other)
	return c == -1 || c == 0
}

// Gt는 초과 비교
func (d Decimal64Pure) Gt(other Decimal64Pure) bool {
	return d.Cmp(other) == 1
}

// Gte는 이상 비교
func (d Decimal64Pure) Gte(other Decimal64Pure) bool {
	c := d.Cmp(other)
	return c == 1 || c == 0
}

// Abs는 절대값을 반환
func (d Decimal64Pure) Abs() Decimal64Pure {
	return Decimal64Pure(uint64(d) &^ bid64SignMask)
}

// Neg는 부호를 반전
func (d Decimal64Pure) Neg() Decimal64Pure {
	return Decimal64Pure(uint64(d) ^ bid64SignMask)
}

// bid64MultFactor - Intel bid_mult_factor table for minmax
var bid64MultFactor = [16]uint64{
	1, 10, 100, 1000,
	10000, 100000, 1000000, 10000000,
	100000000, 1000000000, 10000000000, 100000000000,
	1000000000000, 10000000000000,
	100000000000000, 1000000000000000,
}

// Min은 두 값 중 작은 값을 반환 (bid64_minnum 기계적 포팅)
// Intel bid64_minmax.c 라인바이라인 포팅
func (d Decimal64Pure) Min(other Decimal64Pure) Decimal64Pure {
	x := uint64(d)
	y := uint64(other)

	// check for non-canonical x
	if (x & bid64NaNMask) == bid64NaNMask { // x is NaN
		x = x & 0xfe03ffffffffffff // clear G6-G12
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and payload
		}
	} else if (x & bid64InfMask) == bid64InfMask { // check for Infinity
		x = x & (bid64SignMask | bid64InfMask)
	} else { // x is not special
		// check for non-canonical values - treated as zero
		if (x & bid64SteeringBitsMask) == bid64SteeringBitsMask {
			if ((x & bid64BinarySig2Mask) | bid64BinaryOr2) > 9999999999999999 {
				// non-canonical
				x = (x & bid64SignMask) | ((x & bid64BinaryExp2Mask) << 2)
			}
		}
	}

	// check for non-canonical y
	if (y & bid64NaNMask) == bid64NaNMask { // y is NaN
		y = y & 0xfe03ffffffffffff // clear G6-G12
		if (y & 0x0003ffffffffffff) > 999999999999999 {
			y = y & 0xfe00000000000000 // clear G6-G12 and payload
		}
	} else if (y & bid64InfMask) == bid64InfMask { // check for Infinity
		y = y & (bid64SignMask | bid64InfMask)
	} else { // y is not special
		// check for non-canonical values - treated as zero
		if (y & bid64SteeringBitsMask) == bid64SteeringBitsMask {
			if ((y & bid64BinarySig2Mask) | bid64BinaryOr2) > 9999999999999999 {
				// non-canonical
				y = (y & bid64SignMask) | ((y & bid64BinaryExp2Mask) << 2)
			}
		}
	}

	// NaN (CASE1)
	if (x & bid64NaNMask) == bid64NaNMask { // x is NAN
		if (x & bid64SNaNMask) == bid64SNaNMask { // x is SNaN
			// if x is SNAN, then return quiet (x)
			x = x & 0xfdffffffffffffff // quietize x
			return Decimal64Pure(x)
		} else { // x is QNaN
			if (y & bid64NaNMask) == bid64NaNMask { // y is NAN
				return Decimal64Pure(x)
			} else {
				return Decimal64Pure(y)
			}
		}
	} else if (y & bid64NaNMask) == bid64NaNMask { // y is NaN, but x is not
		if (y & bid64SNaNMask) == bid64SNaNMask {
			y = y & 0xfdffffffffffffff // quietize y
			return Decimal64Pure(y)
		} else {
			// will return x (which is not NaN)
			return Decimal64Pure(x)
		}
	}

	// SIMPLE (CASE2)
	if x == y {
		return Decimal64Pure(x)
	}

	// INFINITY (CASE3)
	if (x & bid64InfMask) == bid64InfMask {
		if (x & bid64SignMask) == bid64SignMask { // x is neg infinity
			return Decimal64Pure(x)
		}
		// x is pos infinity, return y
		return Decimal64Pure(y)
	} else if (y & bid64InfMask) == bid64InfMask {
		if (y & bid64SignMask) == bid64SignMask {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// Extract exponent and significand
	var expX, expY int
	var sigX, sigY uint64

	if (x & bid64SteeringBitsMask) == bid64SteeringBitsMask {
		expX = int((x & bid64BinaryExp2Mask) >> 51)
		sigX = (x & bid64BinarySig2Mask) | bid64BinaryOr2
	} else {
		expX = int((x & bid64BinaryExp1Mask) >> 53)
		sigX = x & bid64BinarySig1Mask
	}

	if (y & bid64SteeringBitsMask) == bid64SteeringBitsMask {
		expY = int((y & bid64BinaryExp2Mask) >> 51)
		sigY = (y & bid64BinarySig2Mask) | bid64BinaryOr2
	} else {
		expY = int((y & bid64BinaryExp1Mask) >> 53)
		sigY = y & bid64BinarySig1Mask
	}

	// ZERO (CASE4)
	xIsZero := sigX == 0
	yIsZero := sigY == 0

	if xIsZero && yIsZero {
		return Decimal64Pure(y)
	} else if xIsZero {
		if (y & bid64SignMask) == bid64SignMask {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	} else if yIsZero {
		if (x & bid64SignMask) != bid64SignMask {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// OPPOSITE SIGN (CASE5)
	if ((x ^ y) & bid64SignMask) == bid64SignMask {
		if (y & bid64SignMask) == bid64SignMask {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sigX > sigY && expX >= expY {
		if (x & bid64SignMask) != bid64SignMask {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}
	if sigX < sigY && expX <= expY {
		if (x & bid64SignMask) == bid64SignMask {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if expX-expY > 15 {
		if (x & bid64SignMask) != bid64SignMask {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if expY-expX > 15 {
		if (x & bid64SignMask) == bid64SignMask {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if expX > expY {
		// adjust the x significand upwards (128-bit multiplication)
		hi, lo := bits.Mul64(sigX, bid64MultFactor[expX-expY])
		// if positive, return whichever significand is larger (converse if negative)
		if hi == 0 && lo == sigY {
			return Decimal64Pure(y)
		}
		if (hi > 0 || lo > sigY) != ((x & bid64SignMask) == bid64SignMask) {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// adjust the y significand upwards (128-bit multiplication)
	hi, lo := bits.Mul64(sigY, bid64MultFactor[expY-expX])
	// if positive, return whichever significand is larger (converse if negative)
	if hi == 0 && lo == sigX {
		return Decimal64Pure(y)
	}
	if (hi == 0 && sigX > lo) != ((x & bid64SignMask) == bid64SignMask) {
		return Decimal64Pure(y)
	}
	return Decimal64Pure(x)
}

// Max는 두 값 중 큰 값을 반환 (bid64_maxnum 기계적 포팅)
// Intel bid64_minmax.c 라인바이라인 포팅
func (d Decimal64Pure) Max(other Decimal64Pure) Decimal64Pure {
	x := uint64(d)
	y := uint64(other)

	// check for non-canonical x
	if (x & bid64NaNMask) == bid64NaNMask { // x is NaN
		x = x & 0xfe03ffffffffffff // clear G6-G12
		if (x & 0x0003ffffffffffff) > 999999999999999 {
			x = x & 0xfe00000000000000 // clear G6-G12 and payload
		}
	} else if (x & bid64InfMask) == bid64InfMask { // check for Infinity
		x = x & (bid64SignMask | bid64InfMask)
	} else { // x is not special
		// check for non-canonical values - treated as zero
		if (x & bid64SteeringBitsMask) == bid64SteeringBitsMask {
			if ((x & bid64BinarySig2Mask) | bid64BinaryOr2) > 9999999999999999 {
				// non-canonical
				x = (x & bid64SignMask) | ((x & bid64BinaryExp2Mask) << 2)
			}
		}
	}

	// check for non-canonical y
	if (y & bid64NaNMask) == bid64NaNMask { // y is NaN
		y = y & 0xfe03ffffffffffff // clear G6-G12
		if (y & 0x0003ffffffffffff) > 999999999999999 {
			y = y & 0xfe00000000000000 // clear G6-G12 and payload
		}
	} else if (y & bid64InfMask) == bid64InfMask { // check for Infinity
		y = y & (bid64SignMask | bid64InfMask)
	} else { // y is not special
		// check for non-canonical values - treated as zero
		if (y & bid64SteeringBitsMask) == bid64SteeringBitsMask {
			if ((y & bid64BinarySig2Mask) | bid64BinaryOr2) > 9999999999999999 {
				// non-canonical
				y = (y & bid64SignMask) | ((y & bid64BinaryExp2Mask) << 2)
			}
		}
	}

	// NaN (CASE1)
	if (x & bid64NaNMask) == bid64NaNMask { // x is NAN
		if (x & bid64SNaNMask) == bid64SNaNMask { // x is SNaN
			// if x is SNAN, then return quiet (x)
			x = x & 0xfdffffffffffffff // quietize x
			return Decimal64Pure(x)
		} else { // x is QNaN
			if (y & bid64NaNMask) == bid64NaNMask { // y is NAN
				return Decimal64Pure(x)
			} else {
				return Decimal64Pure(y)
			}
		}
	} else if (y & bid64NaNMask) == bid64NaNMask { // y is NaN, but x is not
		if (y & bid64SNaNMask) == bid64SNaNMask {
			y = y & 0xfdffffffffffffff // quietize y
			return Decimal64Pure(y)
		} else {
			// will return x (which is not NaN)
			return Decimal64Pure(x)
		}
	}

	// SIMPLE (CASE2)
	if x == y {
		return Decimal64Pure(x)
	}

	// INFINITY (CASE3)
	if (x & bid64InfMask) == bid64InfMask {
		if (x & bid64SignMask) == bid64SignMask { // x is neg infinity
			return Decimal64Pure(y)
		}
		// x is pos infinity
		return Decimal64Pure(x)
	} else if (y & bid64InfMask) == bid64InfMask {
		if (y & bid64SignMask) == bid64SignMask {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}

	// Extract exponent and significand
	var expX, expY int
	var sigX, sigY uint64

	if (x & bid64SteeringBitsMask) == bid64SteeringBitsMask {
		expX = int((x & bid64BinaryExp2Mask) >> 51)
		sigX = (x & bid64BinarySig2Mask) | bid64BinaryOr2
	} else {
		expX = int((x & bid64BinaryExp1Mask) >> 53)
		sigX = x & bid64BinarySig1Mask
	}

	if (y & bid64SteeringBitsMask) == bid64SteeringBitsMask {
		expY = int((y & bid64BinaryExp2Mask) >> 51)
		sigY = (y & bid64BinarySig2Mask) | bid64BinaryOr2
	} else {
		expY = int((y & bid64BinaryExp1Mask) >> 53)
		sigY = y & bid64BinarySig1Mask
	}

	// ZERO (CASE4)
	xIsZero := sigX == 0
	yIsZero := sigY == 0

	if xIsZero && yIsZero {
		return Decimal64Pure(y)
	} else if xIsZero {
		if (y & bid64SignMask) == bid64SignMask { // y is negative
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	} else if yIsZero {
		if (x & bid64SignMask) != bid64SignMask { // x is positive
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}

	// OPPOSITE SIGN (CASE5)
	if ((x ^ y) & bid64SignMask) == bid64SignMask {
		if (y & bid64SignMask) == bid64SignMask {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sigX > sigY && expX >= expY {
		if (x & bid64SignMask) != bid64SignMask {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}
	if sigX < sigY && expX <= expY {
		if (x & bid64SignMask) == bid64SignMask {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if expX-expY > 15 {
		if (x & bid64SignMask) != bid64SignMask {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if expY-expX > 15 {
		if (x & bid64SignMask) == bid64SignMask {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if expX > expY {
		// adjust the x significand upwards (128-bit multiplication)
		hi, lo := bits.Mul64(sigX, bid64MultFactor[expX-expY])
		// if positive, return whichever significand is larger (converse if negative)
		if hi == 0 && lo == sigY {
			return Decimal64Pure(y)
		}
		if (hi > 0 || lo > sigY) != ((x & bid64SignMask) == bid64SignMask) {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}

	// adjust the y significand upwards (128-bit multiplication)
	hi, lo := bits.Mul64(sigY, bid64MultFactor[expY-expX])
	// if positive, return whichever significand is larger (converse if negative)
	if hi == 0 && lo == sigX {
		return Decimal64Pure(y)
	}
	if (hi == 0 && sigX > lo) != ((x & bid64SignMask) == bid64SignMask) {
		return Decimal64Pure(x)
	}
	return Decimal64Pure(y)
}

// Quantize는 d를 y의 지수로 양자화 (bid64_quantize 기계적 포팅)
// Intel bid64_quantize.c 라인바이라인 포팅
func (d Decimal64Pure) Quantize(y Decimal64Pure) Decimal64Pure {
	return d.QuantizeWithMode(y, RoundNearestEven)
}

// QuantizeWithMode는 반올림 모드를 지정한 양자화
func (d Decimal64Pure) QuantizeWithMode(y Decimal64Pure, rndMode RoundingMode) Decimal64Pure {
	x := uint64(d)
	yval := uint64(y)

	const (
		nanMask64   = 0x7c00000000000000
		snanMask64  = 0x7e00000000000000
		quietMask64 = 0xfdffffffffffffff // Intel QUIET_MASK64 - clears sNaN bit
	)

	// unpack x
	signX, exponentX, coefficientX, validX := bid64Unpack(x)

	// unpack y
	signY, exponentY, coefficientY, validY := bid64Unpack(yval)
	_ = signY

	if !validY {
		// y is Inf or NaN or 0
		// x=Inf, y=Inf?
		if (coefficientX<<1) == 0xf000000000000000 && (coefficientY<<1) == 0xf000000000000000 {
			return Decimal64Pure(coefficientX)
		}

		// y is Inf or NaN?
		if (yval & 0x7800000000000000) == 0x7800000000000000 {
			// y is sNaN, or (y is Inf and x is finite) => InvalidOperation
			if (yval & nanMask64) != nanMask64 {
				coefficientY = 0
			}
			if (x & nanMask64) != nanMask64 {
				res := uint64(0x7c00000000000000) | (coefficientY & quietMask64)
				// if y is not NaN and x is Infinity
				if (yval&nanMask64) != nanMask64 && (x&nanMask64) == 0x7800000000000000 {
					res = x
				}
				return Decimal64Pure(res)
			}
		}
	}

	if !validX {
		// x is Inf or NaN or 0
		// x is Inf or NaN?
		if (x & 0x7800000000000000) == 0x7800000000000000 {
			if (x & nanMask64) != nanMask64 {
				coefficientX = 0
			}
			res := uint64(0x7c00000000000000) | (coefficientX & quietMask64)
			return Decimal64Pure(res)
		}

		// x is 0
		return Decimal64Pure(bid64VeryFastSmallMantissa(signX, exponentY, 0))
	}

	// get number of decimal digits in coefficient_x
	digitsX := bid64CountDigits(coefficientX)

	exponDiff := exponentX - exponentY
	totalDigits := digitsX + exponDiff

	// check range of scaled coefficient
	if uint32(totalDigits+1) <= 17 {
		if exponDiff >= 0 {
			coefficientX *= pow10[exponDiff]
			return Decimal64Pure(bid64VeryFast(signX, exponentY, coefficientX))
		}

		// must round off -expon_diff digits
		extraDigits := -exponDiff
		rmode := int(rndMode)
		if signX != 0 && (rmode == 1 || rmode == 2) {
			rmode = 3 - rmode
		}

		coefficientX += bid64RoundConstTable[rmode][extraDigits]

		// get P*(2^M[extra_digits])/10^extra_digits
		hi, lo := bits.Mul64(coefficientX, bid64Reciprocals10[extraDigits])

		// now get P/10^extra_digits: shift C64 right by M[extra_digits]-128
		amount := bid64ShortRecipScale[extraDigits]
		C64 := hi >> amount

		// for round-to-nearest-even, check if fractional part is exactly .5
		if rndMode == RoundNearestEven {
			if C64&1 != 0 {
				// check whether fractional part is exactly .5
				amount2 := uint(64 - amount)
				remainderH := (^uint64(0)) >> amount2
				remainderH = remainderH & hi

				if remainderH == 0 && lo < bid64Reciprocals10[extraDigits] {
					C64--
				}
			}
		}

		return Decimal64Pure(bid64VeryFastSmallMantissa(signX, exponentY, C64))
	}

	if totalDigits < 0 {
		// result is 0 or 1 depending on rounding mode
		C64 := uint64(0)
		rmode := int(rndMode)
		if signX != 0 && (rmode == 1 || rmode == 2) {
			rmode = 3 - rmode
		}
		if rmode == 2 { // RoundTowardPositive after sign adjustment
			C64 = 1
		}
		return Decimal64Pure(bid64VeryFastSmallMantissa(signX, exponentY, C64))
	}

	// more than 16 digits in coefficient => InvalidOperation => NaN
	return Decimal64Pure(0x7c00000000000000)
}

// bid64Unpack unpacks a BID64 value (mirrors Intel's unpack_BID64)
// Returns sign (0 or 1), biased exponent, coefficient, and valid flag
// For NaN/Inf, coeff includes the sign and type bits for proper handling
func bid64Unpack(x uint64) (sign int, exp int, coeff uint64, valid bool) {
	const (
		sinfMask64 = 0xf800000000000000 // SINFINITY_MASK64
	)

	sign = int((x >> 63) & 1)

	// Check for special encodings (11 in bits 62-61)
	if (x & bid64SteeringBitsMask) == bid64SteeringBitsMask {
		// Check for Infinity or NaN
		if (x & bid64InfMask) == bid64InfMask {
			// Infinity or NaN
			exp = 0
			// Intel: *pcoefficient_x = x & 0xfe03ffffffffffffull
			coeff = x & 0xfe03ffffffffffff
			// If payload >= 1000000000000000, clear it too
			if (x & 0x0003ffffffffffff) >= 1000000000000000 {
				coeff = x & 0xfe00000000000000
			}
			// For Infinity (not NaN), use signed infinity mask
			if (x & bid64NaNMask) == bid64InfMask {
				coeff = x & sinfMask64
			}
			return sign, exp, coeff, false
		}
		// Large coefficient encoding
		exp = int((x & bid64BinaryExp2Mask) >> 51)
		coeff = (x & bid64BinarySig2Mask) | bid64BinaryOr2
		if coeff > 9999999999999999 {
			// Non-canonical - treated as zero
			coeff = 0
		}
	} else {
		// Small coefficient encoding
		exp = int((x & bid64BinaryExp1Mask) >> 53)
		coeff = x & bid64BinarySig1Mask
	}

	if coeff == 0 {
		// Zero - valid but coefficient is zero
		return sign, exp, 0, false
	}

	return sign, exp, coeff, true
}

// bid64CountDigits counts the number of decimal digits in a coefficient
func bid64CountDigits(coeff uint64) int {
	if coeff == 0 {
		return 1
	}
	digits := 1
	for coeff >= 10 {
		coeff /= 10
		digits++
	}
	return digits
}

// bid64VeryFastSmallMantissa creates a BID64 from small mantissa (Intel: very_fast_get_BID64_small_mantissa)
// WARNING: Assumes coeff < 2^53, no overflow checking
func bid64VeryFastSmallMantissa(sign int, exp int, coeff uint64) uint64 {
	signBit := uint64(sign) << 63
	return signBit | (uint64(exp) << 53) | coeff
}

// bid64VeryFast creates a BID64, using large encoding if needed (Intel: very_fast_get_BID64)
func bid64VeryFast(sign int, exp int, coeff uint64) uint64 {
	signBit := uint64(sign) << 63
	const mask = uint64(1) << 53 // 2^53

	// check whether coefficient fits in 53 bits
	if coeff < mask {
		// small encoding
		return signBit | (uint64(exp) << 53) | coeff
	}

	// large encoding (special format)
	r := uint64(exp)
	r <<= 51 // EXPONENT_SHIFT_LARGE64
	r |= signBit | bid64SpecialEncodingMask
	// add coeff without leading bits (keep lower 51 bits)
	coeff &= (mask >> 2) - 1 // (2^53 >> 2) - 1 = 2^51 - 1
	r |= coeff
	return r
}

// bid64RoundConstTable - Intel's bid_round_const_table
var bid64RoundConstTable = [6][18]uint64{
	// BID_ROUNDING_TO_NEAREST (0)
	{0, 5, 50, 500, 5000, 50000, 500000, 5000000, 50000000, 500000000,
		5000000000, 50000000000, 500000000000, 5000000000000, 50000000000000,
		500000000000000, 5000000000000000, 50000000000000000},
	// BID_ROUNDING_DOWN (1)
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	// BID_ROUNDING_UP (2)
	{0, 9, 99, 999, 9999, 99999, 999999, 9999999, 99999999, 999999999,
		9999999999, 99999999999, 999999999999, 9999999999999, 99999999999999,
		999999999999999, 9999999999999999, 99999999999999999},
	// BID_ROUNDING_TO_ZERO (3)
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	// BID_ROUNDING_TIES_AWAY (4)
	{0, 5, 50, 500, 5000, 50000, 500000, 5000000, 50000000, 500000000,
		5000000000, 50000000000, 500000000000, 5000000000000, 50000000000000,
		500000000000000, 5000000000000000, 50000000000000000},
	// RoundNearestDown (5) - same as TO_NEAREST for positive
	{0, 5, 50, 500, 5000, 50000, 500000, 5000000, 50000000, 500000000,
		5000000000, 50000000000, 500000000000, 5000000000000, 50000000000000,
		500000000000000, 5000000000000000, 50000000000000000},
}

// bid64Reciprocals10 - Intel's bid_reciprocals10_64
var bid64Reciprocals10 = [18]uint64{
	0x0,                // unused
	0x3333333333333334, // 10^-1
	0x51eb851eb851eb86, // 10^-2
	0x20c49ba5e353f7cf, // 10^-3
	0x346dc5d63886594b, // 10^-4
	0x29f16b11c6d1e109, // 10^-5
	0x218def416bdb1a6e, // 10^-6
	0x35afe535795e90b0, // 10^-7
	0x2af31dc4611873c0, // 10^-8
	0x225c17d04dad2966, // 10^-9
	0x36f9bfb3af7b7570, // 10^-10
	0x2bfaffc2f2c92ac0, // 10^-11
	0x232f33025bd42233, // 10^-12
	0x384b84d092ed0385, // 10^-13
	0x2d09370d42573604, // 10^-14
	0x24075f3dceac2b37, // 10^-15
	0x39a5652fb1137857, // 10^-16
	0x2e1dea8c8da92d13, // 10^-17
}

// bid64ShortRecipScale - Intel's bid_short_recip_scale (bid_decimal_data.c)
var bid64ShortRecipScale = [18]uint{
	1,  // 0: dummy
	1,  // 1: 65-64
	5,  // 2: 69-64
	7,  // 3: 71-64
	11, // 4: 75-64
	14, // 5: 78-64
	17, // 6: 81-64
	21, // 7: 85-64
	24, // 8: 88-64
	27, // 9: 91-64
	31, // 10: 95-64
	34, // 11: 98-64
	37, // 12: 101-64
	41, // 13: 105-64
	44, // 14: 108-64
	47, // 15: 111-64
	51, // 16: 115-64
	54, // 17: 118-64
}

// FMA는 Fused Multiply-Add: d * y + z
// 중간 결과의 정밀도 손실 없이 계산
func (d Decimal64Pure) FMA(y, z Decimal64Pure) Decimal64Pure {
	// 간단한 구현: Mul 후 Add
	// 완전한 FMA는 중간 결과를 더 높은 정밀도로 유지해야 하지만,
	// 현재는 간단하게 구현
	return d.Mul(y).Add(z)
}

// Sqrt는 제곱근을 반환
func (d Decimal64Pure) Sqrt() Decimal64Pure {
	x := uint64(d)

	// NaN 처리
	if (x & bid64NaNMask) == bid64NaNMask {
		return Decimal64Pure(x | bid64QuietBit)
	}

	// 부호 비트 확인
	if (x & bid64SignMask) != 0 {
		// 음수의 제곱근 -> NaN
		_, _, coeff, _, _, _ := decodeBID64(x)
		if coeff != 0 {
			return Decimal64Pure(0x7c00000000000000) // QNaN
		}
		// -0의 제곱근은 -0
		return d
	}

	// Infinity 처리
	if (x & bid64InfinityMask) == bid64InfinityMask {
		return d // +Inf의 제곱근은 +Inf
	}

	// 0 처리
	_, exp, coeff, _, _, _ := decodeBID64(x)
	if coeff == 0 {
		return d
	}

	// Newton-Raphson 방법으로 제곱근 계산
	// 초기 추정값: float64로 근사
	realExp := exp - bid64ExponentBias
	approx := float64(coeff)
	for realExp > 0 {
		approx *= 10
		realExp--
	}
	for realExp < 0 {
		approx /= 10
		realExp++
	}
	sqrtApprox := sqrtFloat(approx)

	// 결과를 Decimal64Pure로 변환 후 Newton-Raphson 정제
	result, _ := ParseDecimal64Pure(formatFloat(sqrtApprox))

	// Newton-Raphson: x' = (x + n/x) / 2
	two, _ := ParseDecimal64Pure("2")
	for i := 0; i < 5; i++ {
		// result = (result + d/result) / 2
		div := d.Div(result)
		sum := result.Add(div)
		result = sum.Div(two)
	}

	return result
}

// sqrtFloat는 float64의 제곱근을 계산
func sqrtFloat(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x == 0 {
		return 0
	}
	// Newton-Raphson
	guess := x / 2
	for i := 0; i < 50; i++ {
		newGuess := (guess + x/guess) / 2
		if newGuess == guess {
			break
		}
		guess = newGuess
	}
	return guess
}

// formatFloat는 float64를 문자열로 변환
func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', 16, 64)
}

// Round는 가장 가까운 정수로 반올림 (짝수 우선)
func (d Decimal64Pure) Round() Decimal64Pure {
	x := uint64(d)

	// NaN/Infinity 처리
	if (x & bid64NaNMask) == bid64NaNMask {
		return Decimal64Pure(x | bid64QuietBit)
	}
	if (x & bid64InfinityMask) == bid64InfinityMask {
		return d
	}

	sign, exp, coeff, _, _, _ := decodeBID64(x)
	if coeff == 0 {
		return d
	}

	// 실제 지수
	realExp := exp - bid64ExponentBias

	// 이미 정수면 그대로 반환
	if realExp >= 0 {
		return d
	}

	// 소수점 위치 계산
	decimalPlaces := -realExp
	if decimalPlaces > 16 {
		// 아주 작은 수 -> 0으로 반올림
		return Decimal64Pure(encodeBID64(sign, bid64ExponentBias, 0))
	}

	scale := pow10[decimalPlaces]
	quotient := coeff / scale
	remainder := coeff % scale

	// Round to nearest even
	half := scale / 2
	if remainder > half || (remainder == half && (quotient&1) == 1) {
		quotient++
	}

	return Decimal64Pure(encodeBID64(sign, bid64ExponentBias, quotient))
}

// Floor는 음의 무한대 방향으로 내림
func (d Decimal64Pure) Floor() Decimal64Pure {
	x := uint64(d)

	// NaN/Infinity 처리
	if (x & bid64NaNMask) == bid64NaNMask {
		return Decimal64Pure(x | bid64QuietBit)
	}
	if (x & bid64InfinityMask) == bid64InfinityMask {
		return d
	}

	sign, exp, coeff, _, _, _ := decodeBID64(x)
	if coeff == 0 {
		return d
	}

	realExp := exp - bid64ExponentBias
	if realExp >= 0 {
		return d
	}

	decimalPlaces := -realExp
	if decimalPlaces > 16 {
		// 아주 작은 양수 -> 0, 음수 -> -1
		if sign == 0 {
			return Decimal64Pure(encodeBID64(0, bid64ExponentBias, 0))
		}
		return Decimal64Pure(encodeBID64(1, bid64ExponentBias, 1))
	}

	scale := pow10[decimalPlaces]
	quotient := coeff / scale
	remainder := coeff % scale

	// Floor: 음수이고 나머지가 있으면 내림
	if sign == 1 && remainder > 0 {
		quotient++
	}

	return Decimal64Pure(encodeBID64(sign, bid64ExponentBias, quotient))
}

// Ceil은 양의 무한대 방향으로 올림
func (d Decimal64Pure) Ceil() Decimal64Pure {
	x := uint64(d)

	// NaN/Infinity 처리
	if (x & bid64NaNMask) == bid64NaNMask {
		return Decimal64Pure(x | bid64QuietBit)
	}
	if (x & bid64InfinityMask) == bid64InfinityMask {
		return d
	}

	sign, exp, coeff, _, _, _ := decodeBID64(x)
	if coeff == 0 {
		return d
	}

	realExp := exp - bid64ExponentBias
	if realExp >= 0 {
		return d
	}

	decimalPlaces := -realExp
	if decimalPlaces > 16 {
		// 아주 작은 양수 -> 1, 음수 -> 0
		if sign == 0 {
			return Decimal64Pure(encodeBID64(0, bid64ExponentBias, 1))
		}
		return Decimal64Pure(encodeBID64(1, bid64ExponentBias, 0))
	}

	scale := pow10[decimalPlaces]
	quotient := coeff / scale
	remainder := coeff % scale

	// Ceil: 양수이고 나머지가 있으면 올림
	if sign == 0 && remainder > 0 {
		quotient++
	}

	return Decimal64Pure(encodeBID64(sign, bid64ExponentBias, quotient))
}

// Trunc는 0 방향으로 버림
func (d Decimal64Pure) Trunc() Decimal64Pure {
	x := uint64(d)

	// NaN/Infinity 처리
	if (x & bid64NaNMask) == bid64NaNMask {
		return Decimal64Pure(x | bid64QuietBit)
	}
	if (x & bid64InfinityMask) == bid64InfinityMask {
		return d
	}

	sign, exp, coeff, _, _, _ := decodeBID64(x)
	if coeff == 0 {
		return d
	}

	realExp := exp - bid64ExponentBias
	if realExp >= 0 {
		return d
	}

	decimalPlaces := -realExp
	if decimalPlaces > 16 {
		// 아주 작은 수 -> 0
		return Decimal64Pure(encodeBID64(sign, bid64ExponentBias, 0))
	}

	scale := pow10[decimalPlaces]
	quotient := coeff / scale

	return Decimal64Pure(encodeBID64(sign, bid64ExponentBias, quotient))
}

// Int64는 int64로 변환 (round-to-nearest-even)
// NaN/Infinity는 0을 반환
func (d Decimal64Pure) Int64() int64 {
	x := uint64(d)

	// NaN/Infinity는 0 반환
	if (x & bid64NaNMask) == bid64NaNMask {
		return 0
	}
	if (x & bid64InfinityMask) == bid64InfinityMask {
		// Overflow 케이스: 최대/최소값 반환 (Intel BID 동작 맞추기)
		if (x & bid64SignMask) != 0 {
			return -9223372036854775808 // INT64_MIN
		}
		return 9223372036854775807 // INT64_MAX
	}

	sign, exp, coeff, _, _, _ := decodeBID64(x)
	if coeff == 0 {
		return 0
	}

	realExp := exp - bid64ExponentBias

	// 정수 부분 계산
	var result int64
	if realExp >= 0 {
		// 큰 수
		if realExp <= 18 {
			result = int64(coeff) * int64(pow10[realExp])
		} else {
			// 오버플로우
			if sign == 0 {
				return 9223372036854775807
			}
			return -9223372036854775808
		}
	} else {
		// 소수점 이하 버림 필요
		decimalPlaces := -realExp
		if decimalPlaces > 16 {
			return 0
		}

		scale := pow10[decimalPlaces]
		quotient := coeff / scale
		remainder := coeff % scale

		// Round to nearest even
		half := scale / 2
		if remainder > half || (remainder == half && (quotient&1) == 1) {
			quotient++
		}

		result = int64(quotient)
	}

	if sign == 1 {
		return -result
	}
	return result
}

// Float64는 float64로 변환
func (d Decimal64Pure) Float64() float64 {
	x := uint64(d)

	// NaN 처리
	if (x & bid64NaNMask) == bid64NaNMask {
		if (x & bid64SignMask) != 0 {
			return -naN()
		}
		return naN()
	}

	// Infinity 처리
	if (x & bid64InfinityMask) == bid64InfinityMask {
		if (x & bid64SignMask) != 0 {
			return negInf()
		}
		return posInf()
	}

	sign, exp, coeff, _, _, _ := decodeBID64(x)
	if coeff == 0 {
		if sign == 1 {
			return negZero()
		}
		return 0.0
	}

	// 실제 지수
	realExp := exp - bid64ExponentBias

	// float64로 계산
	result := float64(coeff)
	if realExp > 0 {
		for realExp > 0 {
			result *= 10
			realExp--
		}
	} else {
		for realExp < 0 {
			result /= 10
			realExp++
		}
	}

	if sign == 1 {
		return -result
	}
	return result
}

// naN은 float64 NaN을 반환
func naN() float64 {
	return math.NaN()
}

// posInf는 양의 무한대를 반환
func posInf() float64 {
	return math.Inf(1)
}

// negInf는 음의 무한대를 반환
func negInf() float64 {
	return math.Inf(-1)
}

// negZero는 음의 0을 반환
func negZero() float64 {
	return math.Copysign(0, -1)
}

// NewDecimal64PureFromInt64는 int64에서 Decimal64Pure 생성
func NewDecimal64PureFromInt64(x int64) Decimal64Pure {
	if x == 0 {
		return Decimal64Pure(encodeBID64(0, bid64ExponentBias, 0))
	}

	sign := 0
	coeff := uint64(x)
	if x < 0 {
		sign = 1
		coeff = uint64(-x)
	}

	// 계수가 최대를 초과하면 스케일링
	exp := bid64ExponentBias
	for coeff > bid64MaxCoefficient && exp < bid64MaxExponent {
		coeff = (coeff + 5) / 10 // 반올림
		exp++
	}

	return Decimal64Pure(encodeBID64(sign, exp, coeff))
}

// NewDecimal64PureFromFloat64는 float64에서 Decimal64Pure 생성
func NewDecimal64PureFromFloat64(x float64) Decimal64Pure {
	// 특수값 처리
	if x != x { // NaN
		return Decimal64Pure(0x7c00000000000000)
	}
	if x > 1e308 { // +Inf
		return Decimal64Pure(0x7800000000000000)
	}
	if x < -1e308 { // -Inf
		return Decimal64Pure(0xf800000000000000)
	}
	if x == 0 {
		if 1/x < 0 { // -0
			return Decimal64Pure(encodeBID64(1, bid64ExponentBias, 0))
		}
		return Decimal64Pure(encodeBID64(0, bid64ExponentBias, 0))
	}

	// 문자열 변환을 통해 정확한 값 획득
	s := strconv.FormatFloat(x, 'g', 16, 64)
	result, _ := ParseDecimal64Pure(s)
	return result
}

// ParseDecimal64Pure는 문자열을 Decimal64Pure로 파싱
func ParseDecimal64Pure(s string) (Decimal64Pure, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, strconv.ErrSyntax
	}

	// 부호 처리
	sign := 0
	if s[0] == '-' {
		sign = 1
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}

	// 특수값 처리 (Intel bid64_from_string 동작 모방)
	upper := strings.ToLower(s)

	// snan으로 시작하면 SNaN (뒤 문자 무시 - Intel 동작)
	if strings.HasPrefix(upper, "snan") {
		return Decimal64Pure(encodeNaN64(sign, true)), nil
	}

	// nan으로 시작하면 NaN
	if strings.HasPrefix(upper, "nan") {
		return Decimal64Pure(encodeNaN64(sign, false)), nil
	}

	// infinity (대소문자 무관)
	if upper == "infinity" {
		return Decimal64Pure(encodeInfinity64(sign)), nil
	}

	// inf로 시작하는 경우 - inf만 있거나 infinity여야 함
	if strings.HasPrefix(upper, "inf") {
		// inf 뒤에 문자가 있으면 infinity로 완성되어야 함
		if len(upper) == 3 {
			return Decimal64Pure(encodeInfinity64(sign)), nil
		}
		// inf 뒤에 문자가 있는데 infinity가 아니면 NaN
		return Decimal64Pure(encodeNaN64(sign, false)), nil
	}

	// 지수 분리
	expPart := 0
	if idx := strings.IndexAny(s, "eE"); idx >= 0 {
		expStr := s[idx+1:]
		s = s[:idx]
		var err error
		expPart, err = strconv.Atoi(expStr)
		if err != nil {
			return 0, err
		}
	}

	// 소수점 처리
	var intPart, fracPart string
	if idx := strings.Index(s, "."); idx >= 0 {
		intPart = s[:idx]
		fracPart = s[idx+1:]
	} else {
		intPart = s
	}

	// 선행 0 제거
	intPart = strings.TrimLeft(intPart, "0")
	if intPart == "" {
		intPart = "0"
	}

	// 계수 조합
	coeffStr := intPart + fracPart
	coeffStr = strings.TrimLeft(coeffStr, "0")

	// 지수 조정 (0인 경우에도 필요)
	expPart -= len(fracPart)

	if coeffStr == "" {
		// 0 값이지만 원래 지수를 보존
		biasedExp := expPart + bid64ExponentBias
		if biasedExp < 0 {
			biasedExp = 0
		}
		if biasedExp > bid64MaxExponent {
			biasedExp = bid64MaxExponent
		}
		return Decimal64Pure(encodeBID64(sign, biasedExp, 0)), nil
	}

	// 16자리로 제한 (반올림 적용)
	if len(coeffStr) > 16 {
		roundDigit := coeffStr[16] - '0' // 17번째 자리

		// sticky bit: 18번째 이후에 0이 아닌 자릿수가 있는지 확인
		sticky := false
		for i := 17; i < len(coeffStr); i++ {
			if coeffStr[i] != '0' {
				sticky = true
				break
			}
		}

		expPart += len(coeffStr) - 16
		coeffStr = coeffStr[:16]

		coeff, err := strconv.ParseUint(coeffStr, 10, 64)
		if err != nil {
			return 0, err
		}

		// IEEE 754 round-to-nearest-even
		// roundDigit > 5: 올림
		// roundDigit < 5: 내림
		// roundDigit == 5:
		//   - sticky bit 있으면: 올림 (정확히 반보다 큼)
		//   - sticky bit 없으면: 짝수로 반올림 (정확히 반)
		if roundDigit > 5 || (roundDigit == 5 && (sticky || (coeff&1) == 1)) {
			coeff++
			// 오버플로 체크
			if coeff > bid64MaxCoefficient {
				coeff /= 10
				expPart++
			}
		}

		biasedExp := expPart + bid64ExponentBias
		if biasedExp > bid64MaxExponent {
			return Decimal64Pure(encodeInfinity64(sign)), nil
		}
		// biasedExp < 0 인 경우 encodeBID64가 subnormal 처리
		return Decimal64Pure(encodeBID64(sign, biasedExp, coeff)), nil
	}

	coeff, err := strconv.ParseUint(coeffStr, 10, 64)
	if err != nil {
		return 0, err
	}

	biasedExp := expPart + bid64ExponentBias
	if biasedExp > bid64MaxExponent {
		return Decimal64Pure(encodeInfinity64(sign)), nil
	}
	// biasedExp < 0 인 경우 encodeBID64가 subnormal 처리

	return Decimal64Pure(encodeBID64(sign, biasedExp, coeff)), nil
}

// ParseDecimal64PureWithMode는 문자열을 지정된 라운딩 모드로 Decimal64Pure로 파싱
func ParseDecimal64PureWithMode(s string, mode RoundingMode) (Decimal64Pure, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, strconv.ErrSyntax
	}

	// 부호 처리
	sign := 0
	if s[0] == '-' {
		sign = 1
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}

	negative := sign != 0

	// 특수값 처리 (Intel bid64_from_string 동작 모방)
	upper := strings.ToLower(s)

	if strings.HasPrefix(upper, "snan") {
		return Decimal64Pure(encodeNaN64(sign, true)), nil
	}
	if strings.HasPrefix(upper, "nan") {
		return Decimal64Pure(encodeNaN64(sign, false)), nil
	}
	if upper == "infinity" {
		return Decimal64Pure(encodeInfinity64(sign)), nil
	}
	if strings.HasPrefix(upper, "inf") {
		if len(upper) == 3 {
			return Decimal64Pure(encodeInfinity64(sign)), nil
		}
		return Decimal64Pure(encodeNaN64(sign, false)), nil
	}

	// 지수 분리
	expPart := 0
	if idx := strings.IndexAny(s, "eE"); idx >= 0 {
		expStr := s[idx+1:]
		s = s[:idx]
		var err error
		expPart, err = strconv.Atoi(expStr)
		if err != nil {
			return 0, err
		}
	}

	// 소수점 처리
	var intPart, fracPart string
	if idx := strings.Index(s, "."); idx >= 0 {
		intPart = s[:idx]
		fracPart = s[idx+1:]
	} else {
		intPart = s
	}

	intPart = strings.TrimLeft(intPart, "0")
	if intPart == "" {
		intPart = "0"
	}

	coeffStr := intPart + fracPart
	coeffStr = strings.TrimLeft(coeffStr, "0")
	expPart -= len(fracPart)

	if coeffStr == "" {
		biasedExp := expPart + bid64ExponentBias
		if biasedExp < 0 {
			biasedExp = 0
		}
		if biasedExp > bid64MaxExponent {
			biasedExp = bid64MaxExponent
		}
		return Decimal64Pure(encodeBID64(sign, biasedExp, 0)), nil
	}

	// 16자리로 제한 (라운딩 모드 적용)
	if len(coeffStr) > 16 {
		roundDigit := uint64(coeffStr[16] - '0')

		// sticky bit
		sticky := uint64(0)
		for i := 17; i < len(coeffStr); i++ {
			if coeffStr[i] != '0' {
				sticky = 1
				break
			}
		}

		expPart += len(coeffStr) - 16
		coeffStr = coeffStr[:16]

		coeff, err := strconv.ParseUint(coeffStr, 10, 64)
		if err != nil {
			return 0, err
		}

		// 라운딩 모드에 따른 반올림
		// sticky가 있으면 실제 나머지가 더 있음을 반영
		adjustedRemainder := roundDigit
		if sticky != 0 {
			if roundDigit == 0 {
				adjustedRemainder = 1 // 0.xxx > 0이므로 1로 처리 (directed rounding용)
			} else if roundDigit == 5 {
				adjustedRemainder = 6 // 5.xxx > 5이므로 6으로 처리
			}
		}
		coeff = roundWithMode(coeff, adjustedRemainder, 10, mode, negative)

		// 오버플로 체크
		if coeff > bid64MaxCoefficient {
			coeff /= 10
			expPart++
		}

		biasedExp := expPart + bid64ExponentBias
		if biasedExp > bid64MaxExponent {
			return getOverflowResult(sign, mode), nil
		}
		// subnormal/underflow는 encodeBID64WithMode가 처리
		return Decimal64Pure(encodeBID64WithMode(sign, biasedExp, coeff, mode)), nil
	}

	coeff, err := strconv.ParseUint(coeffStr, 10, 64)
	if err != nil {
		return 0, err
	}

	biasedExp := expPart + bid64ExponentBias
	if biasedExp > bid64MaxExponent {
		return getOverflowResult(sign, mode), nil
	}

	// subnormal/underflow는 encodeBID64WithMode가 처리
	return Decimal64Pure(encodeBID64WithMode(sign, biasedExp, coeff, mode)), nil
}

// ParseDecimal64PureWithFlags parses a decimal string and returns the result with flags
// This is the version that returns exception flags for Intel readtest compatibility
func ParseDecimal64PureWithFlags(s string, mode RoundingMode) (Decimal64Pure, uint32, error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return 0, 0, strconv.ErrSyntax
	}

	var flags uint32

	// 부호 처리
	sign := 0
	if s[0] == '-' {
		sign = 1
		s = s[1:]
	} else if s[0] == '+' {
		s = s[1:]
	}

	negative := sign != 0

	// 특수값 처리 (Intel bid64_from_string 동작 모방)
	upper := strings.ToLower(s)

	if strings.HasPrefix(upper, "snan") {
		return Decimal64Pure(encodeNaN64(sign, true)), 0, nil
	}
	if strings.HasPrefix(upper, "nan") {
		return Decimal64Pure(encodeNaN64(sign, false)), 0, nil
	}
	if upper == "infinity" {
		return Decimal64Pure(encodeInfinity64(sign)), 0, nil
	}
	if strings.HasPrefix(upper, "inf") {
		if len(upper) == 3 {
			return Decimal64Pure(encodeInfinity64(sign)), 0, nil
		}
		return Decimal64Pure(encodeNaN64(sign, false)), 0, nil
	}

	// 지수 분리
	expPart := 0
	if idx := strings.IndexAny(s, "eE"); idx >= 0 {
		expStr := s[idx+1:]
		s = s[:idx]
		var err error
		expPart, err = strconv.Atoi(expStr)
		if err != nil {
			return 0, 0, err
		}
	}

	// 소수점 처리
	var intPart, fracPart string
	if idx := strings.Index(s, "."); idx >= 0 {
		intPart = s[:idx]
		fracPart = s[idx+1:]
	} else {
		intPart = s
	}

	intPart = strings.TrimLeft(intPart, "0")
	if intPart == "" {
		intPart = "0"
	}

	coeffStr := intPart + fracPart
	coeffStr = strings.TrimLeft(coeffStr, "0")
	expPart -= len(fracPart)

	if coeffStr == "" {
		biasedExp := expPart + bid64ExponentBias
		if biasedExp < 0 {
			biasedExp = 0
		}
		if biasedExp > bid64MaxExponent {
			biasedExp = bid64MaxExponent
		}
		return Decimal64Pure(encodeBID64(sign, biasedExp, 0)), 0, nil
	}

	// 16자리로 제한 (라운딩 모드 적용)
	if len(coeffStr) > 16 {
		// Check if any truncated digit is non-zero -> INEXACT
		rounded := false
		for i := 16; i < len(coeffStr); i++ {
			if coeffStr[i] != '0' {
				rounded = true
				break
			}
		}
		if rounded {
			flags |= BID_INEXACT_EXCEPTION
		}

		roundDigit := uint64(coeffStr[16] - '0')

		// sticky bit
		sticky := uint64(0)
		for i := 17; i < len(coeffStr); i++ {
			if coeffStr[i] != '0' {
				sticky = 1
				break
			}
		}

		expPart += len(coeffStr) - 16
		coeffStr = coeffStr[:16]

		coeff, err := strconv.ParseUint(coeffStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}

		// 라운딩 모드에 따른 반올림
		adjustedRemainder := roundDigit
		if sticky != 0 {
			if roundDigit == 0 {
				adjustedRemainder = 1
			} else if roundDigit == 5 {
				adjustedRemainder = 6
			}
		}
		coeff = roundWithMode(coeff, adjustedRemainder, 10, mode, negative)

		// 오버플로 체크
		if coeff > bid64MaxCoefficient {
			coeff /= 10
			expPart++
		}

		biasedExp := expPart + bid64ExponentBias
		if biasedExp > bid64MaxExponent {
			return getOverflowResult(sign, mode), flags, nil
		}
		return Decimal64Pure(encodeBID64WithMode(sign, biasedExp, coeff, mode)), flags, nil
	}

	coeff, err := strconv.ParseUint(coeffStr, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	biasedExp := expPart + bid64ExponentBias
	if biasedExp > bid64MaxExponent {
		return getOverflowResult(sign, mode), flags, nil
	}

	return Decimal64Pure(encodeBID64WithMode(sign, biasedExp, coeff, mode)), flags, nil
}

// bid64 minmax constants (mechanically ported from Intel bid64_minmax.c)
const (
	maskNAN64          = uint64(0x7c00000000000000)
	maskSNAN64         = uint64(0x7e00000000000000)
	maskINF64          = uint64(0x7800000000000000)
	maskSIGN64         = uint64(0x8000000000000000)
	maskSTEERINGBITS64 = uint64(0x6000000000000000)
	maskBINARYEXP1_64  = uint64(0x7fe0000000000000)
	maskBINARYEXP2_64  = uint64(0x1ff8000000000000)
	maskBINARYSIG1_64  = uint64(0x001fffffffffffff)
	maskBINARYSIG2_64  = uint64(0x0007ffffffffffff)
	maskBINARYOR2_64   = uint64(0x0020000000000000)
	maxCoeff64         = uint64(9999999999999999)
	maxCoeff64Minus1   = uint64(999999999999999)
)

// bidMultFactor64 is lookup table for 10^n
var bidMultFactor64 = [16]uint64{
	1, 10, 100, 1000,
	10000, 100000, 1000000, 10000000,
	100000000, 1000000000, 10000000000, 100000000000,
	1000000000000, 10000000000000,
	100000000000000, 1000000000000000,
}

// canonicalizeNonSpecial64 canonicalizes a non-special (finite) BID64 value
// Returns the canonicalized value
func canonicalizeNonSpecial64(x uint64) uint64 {
	if (x & maskSTEERINGBITS64) == maskSTEERINGBITS64 {
		if ((x & maskBINARYSIG2_64) | maskBINARYOR2_64) > maxCoeff64 {
			// non-canonical
			return (x & maskSIGN64) | ((x & maskBINARYEXP2_64) << 2)
		}
	}
	return x
}

// MinNum returns the minimum of a and b (mechanically ported from Intel bid64_minnum)
// If one operand is NaN, returns the other operand
// If both are NaN, returns NaN
func (a Decimal64Pure) MinNum(b Decimal64Pure) Decimal64Pure {
	x := uint64(a)
	y := uint64(b)

	// check for non-canonical x
	if (x & maskNAN64) == maskNAN64 { // x is NaN
		x = x & 0xfe03ffffffffffff // clear G6-G12
		if (x & 0x0003ffffffffffff) > maxCoeff64Minus1 {
			x = x & 0xfe00000000000000 // clear G6-G12 and the payload bits
		}
	} else if (x & maskINF64) == maskINF64 { // check for Infinity
		x = x & (maskSIGN64 | maskINF64)
	} else { // x is not special
		x = canonicalizeNonSpecial64(x)
	}

	// check for non-canonical y
	if (y & maskNAN64) == maskNAN64 { // y is NaN
		y = y & 0xfe03ffffffffffff
		if (y & 0x0003ffffffffffff) > maxCoeff64Minus1 {
			y = y & 0xfe00000000000000
		}
	} else if (y & maskINF64) == maskINF64 { // check for Infinity
		y = y & (maskSIGN64 | maskINF64)
	} else { // y is not special
		y = canonicalizeNonSpecial64(y)
	}

	// NaN (CASE1)
	if (x & maskNAN64) == maskNAN64 { // x is NAN
		if (x & maskSNAN64) == maskSNAN64 { // x is SNaN
			// if x is SNAN, then return quiet (x)
			x = x & 0xfdffffffffffffff // quietize x
			return Decimal64Pure(x)
		}
		// x is QNaN
		if (y & maskNAN64) == maskNAN64 { // y is NAN
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	} else if (y & maskNAN64) == maskNAN64 { // y is NaN, but x is not
		if (y & maskSNAN64) == maskSNAN64 {
			y = y & 0xfdffffffffffffff // quietize y
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// SIMPLE (CASE2)
	if x == y {
		return Decimal64Pure(x)
	}

	// INFINITY (CASE3)
	if (x & maskINF64) == maskINF64 {
		if (x & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(x) // x is -Inf
		}
		return Decimal64Pure(y) // x is +Inf
	} else if (y & maskINF64) == maskINF64 {
		if (y & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(y) // y is -Inf
		}
		return Decimal64Pure(x)
	}

	// Extract exp and sig
	var expX, expY int
	var sigX, sigY uint64

	if (x & maskSTEERINGBITS64) == maskSTEERINGBITS64 {
		expX = int((x & maskBINARYEXP2_64) >> 51)
		sigX = (x & maskBINARYSIG2_64) | maskBINARYOR2_64
	} else {
		expX = int((x & maskBINARYEXP1_64) >> 53)
		sigX = x & maskBINARYSIG1_64
	}

	if (y & maskSTEERINGBITS64) == maskSTEERINGBITS64 {
		expY = int((y & maskBINARYEXP2_64) >> 51)
		sigY = (y & maskBINARYSIG2_64) | maskBINARYOR2_64
	} else {
		expY = int((y & maskBINARYEXP1_64) >> 53)
		sigY = y & maskBINARYSIG1_64
	}

	// ZERO (CASE4)
	xIsZero := sigX == 0
	yIsZero := sigY == 0

	if xIsZero && yIsZero {
		return Decimal64Pure(y) // if both zeros, return y
	} else if xIsZero {
		if (y & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(y) // y is negative
		}
		return Decimal64Pure(x)
	} else if yIsZero {
		if (x & maskSIGN64) != maskSIGN64 {
			return Decimal64Pure(y) // x is positive
		}
		return Decimal64Pure(x)
	}

	// OPPOSITE SIGN (CASE5)
	if ((x ^ y) & maskSIGN64) == maskSIGN64 {
		if (y & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(y) // y is negative
		}
		return Decimal64Pure(x)
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sigX > sigY && expX >= expY {
		if (x & maskSIGN64) != maskSIGN64 {
			return Decimal64Pure(y) // positive, y is smaller
		}
		return Decimal64Pure(x)
	}
	if sigX < sigY && expX <= expY {
		if (x & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(y) // negative, y is smaller (more negative)
		}
		return Decimal64Pure(x)
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if expX-expY > 15 {
		if (x & maskSIGN64) != maskSIGN64 {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if expY-expX > 15 {
		if (x & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if expX > expY {
		hi, lo := bits.Mul64(sigX, bidMultFactor64[expX-expY])
		// if equal, return y
		if hi == 0 && lo == sigY {
			return Decimal64Pure(y)
		}
		// compare
		xIsLarger := hi > 0 || lo > sigY
		if xIsLarger != ((x & maskSIGN64) == maskSIGN64) {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// adjust y significand upwards
	hi, lo := bits.Mul64(sigY, bidMultFactor64[expY-expX])
	if hi == 0 && lo == sigX {
		return Decimal64Pure(y)
	}
	xIsLarger := hi == 0 && sigX > lo
	if xIsLarger != ((x & maskSIGN64) == maskSIGN64) {
		return Decimal64Pure(y)
	}
	return Decimal64Pure(x)
}

// MaxNum returns the maximum of a and b (mechanically ported from Intel bid64_maxnum)
// If one operand is NaN, returns the other operand
// If both are NaN, returns NaN
func (a Decimal64Pure) MaxNum(b Decimal64Pure) Decimal64Pure {
	x := uint64(a)
	y := uint64(b)

	// check for non-canonical x
	if (x & maskNAN64) == maskNAN64 {
		x = x & 0xfe03ffffffffffff
		if (x & 0x0003ffffffffffff) > maxCoeff64Minus1 {
			x = x & 0xfe00000000000000
		}
	} else if (x & maskINF64) == maskINF64 {
		x = x & (maskSIGN64 | maskINF64)
	} else {
		x = canonicalizeNonSpecial64(x)
	}

	// check for non-canonical y
	if (y & maskNAN64) == maskNAN64 {
		y = y & 0xfe03ffffffffffff
		if (y & 0x0003ffffffffffff) > maxCoeff64Minus1 {
			y = y & 0xfe00000000000000
		}
	} else if (y & maskINF64) == maskINF64 {
		y = y & (maskSIGN64 | maskINF64)
	} else {
		y = canonicalizeNonSpecial64(y)
	}

	// NaN (CASE1)
	if (x & maskNAN64) == maskNAN64 {
		if (x & maskSNAN64) == maskSNAN64 {
			x = x & 0xfdffffffffffffff
			return Decimal64Pure(x)
		}
		if (y & maskNAN64) == maskNAN64 {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	} else if (y & maskNAN64) == maskNAN64 {
		if (y & maskSNAN64) == maskSNAN64 {
			y = y & 0xfdffffffffffffff
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// SIMPLE (CASE2)
	if x == y {
		return Decimal64Pure(x)
	}

	// INFINITY (CASE3)
	if (x & maskINF64) == maskINF64 {
		if (x & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(y) // x is -Inf
		}
		return Decimal64Pure(x) // x is +Inf
	} else if (y & maskINF64) == maskINF64 {
		if (y & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(x) // y is -Inf
		}
		return Decimal64Pure(y)
	}

	// Extract exp and sig
	var expX, expY int
	var sigX, sigY uint64

	if (x & maskSTEERINGBITS64) == maskSTEERINGBITS64 {
		expX = int((x & maskBINARYEXP2_64) >> 51)
		sigX = (x & maskBINARYSIG2_64) | maskBINARYOR2_64
	} else {
		expX = int((x & maskBINARYEXP1_64) >> 53)
		sigX = x & maskBINARYSIG1_64
	}

	if (y & maskSTEERINGBITS64) == maskSTEERINGBITS64 {
		expY = int((y & maskBINARYEXP2_64) >> 51)
		sigY = (y & maskBINARYSIG2_64) | maskBINARYOR2_64
	} else {
		expY = int((y & maskBINARYEXP1_64) >> 53)
		sigY = y & maskBINARYSIG1_64
	}

	// ZERO (CASE4)
	xIsZero := sigX == 0
	yIsZero := sigY == 0

	if xIsZero && yIsZero {
		return Decimal64Pure(y) // if both zeros, return y
	} else if xIsZero {
		if (y & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(x) // y is negative, x is max
		}
		return Decimal64Pure(y)
	} else if yIsZero {
		if (x & maskSIGN64) != maskSIGN64 {
			return Decimal64Pure(x) // x is positive
		}
		return Decimal64Pure(y)
	}

	// OPPOSITE SIGN (CASE5)
	if ((x ^ y) & maskSIGN64) == maskSIGN64 {
		if (y & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(x) // y is negative, x is max
		}
		return Decimal64Pure(y)
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sigX > sigY && expX >= expY {
		if (x & maskSIGN64) != maskSIGN64 {
			return Decimal64Pure(x) // positive, x is larger
		}
		return Decimal64Pure(y)
	}
	if sigX < sigY && expX <= expY {
		if (x & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(x) // negative, x is larger (less negative)
		}
		return Decimal64Pure(y)
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if expX-expY > 15 {
		if (x & maskSIGN64) != maskSIGN64 {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}
	// if exp_x is 15 less than exp_y, no need for compensation
	if expY-expX > 15 {
		if (x & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if expX > expY {
		hi, lo := bits.Mul64(sigX, bidMultFactor64[expX-expY])
		if hi == 0 && lo == sigY {
			return Decimal64Pure(y)
		}
		xIsLarger := hi > 0 || lo > sigY
		if xIsLarger != ((x & maskSIGN64) == maskSIGN64) {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}

	// adjust y significand upwards
	hi, lo := bits.Mul64(sigY, bidMultFactor64[expY-expX])
	if hi == 0 && lo == sigX {
		return Decimal64Pure(y)
	}
	xIsLarger := hi == 0 && sigX > lo
	if xIsLarger != ((x & maskSIGN64) == maskSIGN64) {
		return Decimal64Pure(x)
	}
	return Decimal64Pure(y)
}

// MinNumMag returns the operand with smaller absolute value (mechanically ported from Intel bid64_minnum_mag)
// If magnitudes are equal, returns the negative operand (if any), otherwise returns y
func (a Decimal64Pure) MinNumMag(b Decimal64Pure) Decimal64Pure {
	x := uint64(a)
	y := uint64(b)

	// check for non-canonical x
	if (x & maskNAN64) == maskNAN64 {
		x = x & 0xfe03ffffffffffff
		if (x & 0x0003ffffffffffff) > maxCoeff64Minus1 {
			x = x & 0xfe00000000000000
		}
	} else if (x & maskINF64) == maskINF64 {
		x = x & (maskSIGN64 | maskINF64)
	} else {
		x = canonicalizeNonSpecial64(x)
	}

	// check for non-canonical y
	if (y & maskNAN64) == maskNAN64 {
		y = y & 0xfe03ffffffffffff
		if (y & 0x0003ffffffffffff) > maxCoeff64Minus1 {
			y = y & 0xfe00000000000000
		}
	} else if (y & maskINF64) == maskINF64 {
		y = y & (maskSIGN64 | maskINF64)
	} else {
		y = canonicalizeNonSpecial64(y)
	}

	// NaN (CASE1)
	if (x & maskNAN64) == maskNAN64 {
		if (x & maskSNAN64) == maskSNAN64 {
			x = x & 0xfdffffffffffffff
			return Decimal64Pure(x)
		}
		if (y & maskNAN64) == maskNAN64 {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	} else if (y & maskNAN64) == maskNAN64 {
		if (y & maskSNAN64) == maskSNAN64 {
			y = y & 0xfdffffffffffffff
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// SIMPLE (CASE2)
	if x == y {
		return Decimal64Pure(x)
	}

	// INFINITY (CASE3)
	if (x & maskINF64) == maskINF64 {
		// x is infinity, its magnitude >= y
		// return x only if y is infinity and x is negative
		if (x&maskSIGN64) == maskSIGN64 && (y&maskINF64) == maskINF64 {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	} else if (y & maskINF64) == maskINF64 {
		// y is infinity, x must be smaller in magnitude
		return Decimal64Pure(x)
	}

	// Extract exp and sig
	var expX, expY int
	var sigX, sigY uint64

	if (x & maskSTEERINGBITS64) == maskSTEERINGBITS64 {
		expX = int((x & maskBINARYEXP2_64) >> 51)
		sigX = (x & maskBINARYSIG2_64) | maskBINARYOR2_64
	} else {
		expX = int((x & maskBINARYEXP1_64) >> 53)
		sigX = x & maskBINARYSIG1_64
	}

	if (y & maskSTEERINGBITS64) == maskSTEERINGBITS64 {
		expY = int((y & maskBINARYEXP2_64) >> 51)
		sigY = (y & maskBINARYSIG2_64) | maskBINARYOR2_64
	} else {
		expY = int((y & maskBINARYEXP1_64) >> 53)
		sigY = y & maskBINARYSIG1_64
	}

	// ZERO (CASE4)
	if sigX == 0 {
		return Decimal64Pure(x) // x is zero, its magnitude must be smaller
	}
	if sigY == 0 {
		return Decimal64Pure(y) // y is zero, its magnitude must be smaller
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sigX > sigY && expX >= expY {
		return Decimal64Pure(y)
	}
	if sigX < sigY && expX <= expY {
		return Decimal64Pure(x)
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if expX-expY > 15 {
		return Decimal64Pure(y)
	}
	if expY-expX > 15 {
		return Decimal64Pure(x)
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if expX > expY {
		hi, lo := bits.Mul64(sigX, bidMultFactor64[expX-expY])
		if hi == 0 && lo == sigY {
			// two numbers are equal, return the negative one, otherwise return x
			if (y & maskSIGN64) == maskSIGN64 {
				return Decimal64Pure(y)
			}
			return Decimal64Pure(x)
		}
		if hi != 0 || lo > sigY {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// adjust y significand upwards
	hi, lo := bits.Mul64(sigY, bidMultFactor64[expY-expX])
	if hi == 0 && lo == sigX {
		// two numbers are equal, return the negative one, otherwise return x
		if (y & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}
	if hi == 0 && sigX > lo {
		return Decimal64Pure(y)
	}
	return Decimal64Pure(x)
}

// MaxNumMag returns the operand with larger absolute value (mechanically ported from Intel bid64_maxnum_mag)
// If magnitudes are equal, returns the positive operand (if any), otherwise returns y
func (a Decimal64Pure) MaxNumMag(b Decimal64Pure) Decimal64Pure {
	x := uint64(a)
	y := uint64(b)

	// check for non-canonical x
	if (x & maskNAN64) == maskNAN64 {
		x = x & 0xfe03ffffffffffff
		if (x & 0x0003ffffffffffff) > maxCoeff64Minus1 {
			x = x & 0xfe00000000000000
		}
	} else if (x & maskINF64) == maskINF64 {
		x = x & (maskSIGN64 | maskINF64)
	} else {
		x = canonicalizeNonSpecial64(x)
	}

	// check for non-canonical y
	if (y & maskNAN64) == maskNAN64 {
		y = y & 0xfe03ffffffffffff
		if (y & 0x0003ffffffffffff) > maxCoeff64Minus1 {
			y = y & 0xfe00000000000000
		}
	} else if (y & maskINF64) == maskINF64 {
		y = y & (maskSIGN64 | maskINF64)
	} else {
		y = canonicalizeNonSpecial64(y)
	}

	// NaN (CASE1)
	if (x & maskNAN64) == maskNAN64 {
		if (x & maskSNAN64) == maskSNAN64 {
			x = x & 0xfdffffffffffffff
			return Decimal64Pure(x)
		}
		if (y & maskNAN64) == maskNAN64 {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	} else if (y & maskNAN64) == maskNAN64 {
		if (y & maskSNAN64) == maskSNAN64 {
			y = y & 0xfdffffffffffffff
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	}

	// SIMPLE (CASE2)
	if x == y {
		return Decimal64Pure(x)
	}

	// INFINITY (CASE3)
	if (x & maskINF64) == maskINF64 {
		// x is infinity, its magnitude >= y
		// return y only if both are infinity and x is negative
		if (x&maskSIGN64) == maskSIGN64 && (y&maskINF64) == maskINF64 {
			return Decimal64Pure(y)
		}
		return Decimal64Pure(x)
	} else if (y & maskINF64) == maskINF64 {
		// y is infinity, y must be larger in magnitude
		return Decimal64Pure(y)
	}

	// Extract exp and sig
	var expX, expY int
	var sigX, sigY uint64

	if (x & maskSTEERINGBITS64) == maskSTEERINGBITS64 {
		expX = int((x & maskBINARYEXP2_64) >> 51)
		sigX = (x & maskBINARYSIG2_64) | maskBINARYOR2_64
	} else {
		expX = int((x & maskBINARYEXP1_64) >> 53)
		sigX = x & maskBINARYSIG1_64
	}

	if (y & maskSTEERINGBITS64) == maskSTEERINGBITS64 {
		expY = int((y & maskBINARYEXP2_64) >> 51)
		sigY = (y & maskBINARYSIG2_64) | maskBINARYOR2_64
	} else {
		expY = int((y & maskBINARYEXP1_64) >> 53)
		sigY = y & maskBINARYSIG1_64
	}

	// ZERO (CASE4)
	if sigX == 0 {
		return Decimal64Pure(y) // x is zero, y must be larger in magnitude
	}
	if sigY == 0 {
		return Decimal64Pure(x) // y is zero, x must be larger in magnitude
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	if sigX > sigY && expX >= expY {
		return Decimal64Pure(x)
	}
	if sigX < sigY && expX <= expY {
		return Decimal64Pure(y)
	}

	// if exp_x is 15 greater than exp_y, no need for compensation
	if expX-expY > 15 {
		return Decimal64Pure(x)
	}
	if expY-expX > 15 {
		return Decimal64Pure(y)
	}

	// if |exp_x - exp_y| < 15, it comes down to the compensated significand
	if expX > expY {
		hi, lo := bits.Mul64(sigX, bidMultFactor64[expX-expY])
		if hi == 0 && lo == sigY {
			// two numbers are equal, return the positive one, otherwise return y
			if (y & maskSIGN64) == maskSIGN64 {
				return Decimal64Pure(x)
			}
			return Decimal64Pure(y)
		}
		if hi != 0 || lo > sigY {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}

	// adjust y significand upwards
	hi, lo := bits.Mul64(sigY, bidMultFactor64[expY-expX])
	if hi == 0 && lo == sigX {
		// two numbers are equal, return the positive one, otherwise return y
		if (y & maskSIGN64) == maskSIGN64 {
			return Decimal64Pure(x)
		}
		return Decimal64Pure(y)
	}
	if hi == 0 && sigX > lo {
		return Decimal64Pure(x)
	}
	return Decimal64Pure(y)
}

// TotalOrder implements IEEE 754 totalOrder predicate.
// Returns true if x <= y in the total order, where:
// -NaN < -sNaN < -Inf < negative finite < -0 < +0 < positive finite < +Inf < +sNaN < +NaN
// Within same NaN type, payload ordering applies.
// Ported mechanically from Intel bid64_totalOrder (bid64_noncomp.c)
func (x Decimal64Pure) TotalOrder(y Decimal64Pure) bool {
	xv := uint64(x)
	yv := uint64(y)

	// NaN (CASE1)
	// if x and y are unordered numerically because either operand is NaN
	//    (1) totalOrder(-NaN, number) is true
	//    (2) totalOrder(number, +NaN) is true
	//    (3) if x and y are both NaN:
	//           i) negative sign bit < positive sign bit
	//           ii) signaling < quiet for +NaN, reverse for -NaN
	//           iii) lesser payload < greater payload for +NaN (reverse for -NaN)
	//           iv) else if bitwise identical (in canonical form), return 1
	if (xv & bid64NaNMask) == bid64NaNMask {
		// if x is -NaN
		if (xv & bid64SignMask) == bid64SignMask {
			// return true, unless y is -NaN also
			if (yv&bid64NaNMask) != bid64NaNMask || (yv&bid64SignMask) != bid64SignMask {
				return true // y is a number or +NaN
			}
			// x and y are both -NaN
			// if x and y are both -sNaN or both -qNaN, compare payloads
			// xnor: true if both are sNaN or both are qNaN
			xIsSNaN := (xv & bid64SNaNMask) == bid64SNaNMask
			yIsSNaN := (yv & bid64SNaNMask) == bid64SNaNMask
			if xIsSNaN == yIsSNaN {
				// compare payloads - for -NaN, larger payload comes first
				pyldX := xv & 0x0003ffffffffffff
				pyldY := yv & 0x0003ffffffffffff
				if pyldY > 999999999999999 || pyldY == 0 {
					return true // y's payload is 0 or non-canonical
				}
				if pyldX > 999999999999999 || pyldX == 0 {
					return false // x's payload is 0 or non-canonical
				}
				return pyldX >= pyldY
			}
			// one is -sNaN, one is -qNaN: -qNaN < -sNaN
			return yIsSNaN
		}
		// x is +NaN
		if (yv&bid64NaNMask) != bid64NaNMask || (yv&bid64SignMask) == bid64SignMask {
			return false // y is a number or -NaN
		}
		// x and y are both +NaN
		xIsSNaN := (xv & bid64SNaNMask) == bid64SNaNMask
		yIsSNaN := (yv & bid64SNaNMask) == bid64SNaNMask
		if xIsSNaN == yIsSNaN {
			// compare payloads - for +NaN, smaller payload comes first
			pyldX := xv & 0x0003ffffffffffff
			pyldY := yv & 0x0003ffffffffffff
			if pyldX > 999999999999999 || pyldX == 0 {
				return true // x's payload is 0 or non-canonical
			}
			if pyldY > 999999999999999 || pyldY == 0 {
				return false // y's payload is 0 or non-canonical
			}
			return pyldX <= pyldY
		}
		// +sNaN < +qNaN
		return xIsSNaN
	} else if (yv & bid64NaNMask) == bid64NaNMask {
		// x is not NaN, y is NaN
		// return true if y is positive
		return (yv & bid64SignMask) != bid64SignMask
	}

	// SIMPLE (CASE2)
	// if all bits are the same, these numbers are equal
	if xv == yv {
		return true
	}

	// OPPOSITE SIGNS (CASE3)
	// if signs are opposite, return true if x is negative
	xNeg := (xv & bid64SignMask) == bid64SignMask
	yNeg := (yv & bid64SignMask) == bid64SignMask
	if xNeg != yNeg {
		return xNeg
	}

	// INFINITY (CASE4)
	if (xv & bid64InfinityMask) == bid64InfinityMask {
		if xNeg {
			return true // -Inf <= anything with same sign
		}
		// +Inf: only true if y is also +Inf
		return (yv & bid64InfinityMask) == bid64InfinityMask
	} else if (yv & bid64InfinityMask) == bid64InfinityMask {
		// x is finite, y is Inf
		return !yNeg // true if y is +Inf
	}

	// Decode x
	var expX, expY int
	var sigX, sigY uint64
	var xIsZero, yIsZero bool

	if (xv & bid64SpecialEncodingMask) == bid64SpecialEncodingMask {
		expX = int((xv >> 51) & 0x3ff)
		sigX = (xv & 0x0007ffffffffffff) | 0x0020000000000000
		if sigX > 9999999999999999 || sigX == 0 {
			xIsZero = true
		}
	} else {
		expX = int((xv >> 53) & 0x3ff)
		sigX = xv & 0x001fffffffffffff
		if sigX == 0 {
			xIsZero = true
		}
	}

	if (yv & bid64SpecialEncodingMask) == bid64SpecialEncodingMask {
		expY = int((yv >> 51) & 0x3ff)
		sigY = (yv & 0x0007ffffffffffff) | 0x0020000000000000
		if sigY > 9999999999999999 || sigY == 0 {
			yIsZero = true
		}
	} else {
		expY = int((yv >> 53) & 0x3ff)
		sigY = yv & 0x001fffffffffffff
		if sigY == 0 {
			yIsZero = true
		}
	}

	// ZERO (CASE5)
	if xIsZero && yIsZero {
		if xNeg == yNeg {
			// same sign zeros: compare by exponent
			// totalOrder(x,y) iff exp_x <= exp_y for positive, >= for negative
			if expX == expY {
				return true
			}
			if xNeg {
				return expX >= expY
			}
			return expX <= expY
		}
		// different sign zeros: -0 < +0
		return xNeg
	}
	if xIsZero {
		// x is zero, y is not: x < y if y is positive
		return !yNeg
	}
	if yIsZero {
		// x is not zero, y is zero: x < y if x is negative
		return xNeg
	}

	// REDUNDANT REPRESENTATIONS (CASE6)
	// Both are non-zero finite numbers with same sign
	if sigX > sigY && expX >= expY {
		return xNeg // larger magnitude: true if negative
	}
	if sigX < sigY && expX <= expY {
		return !xNeg // smaller magnitude: true if positive
	}

	// Need to compare by scaling significands
	if expX-expY > 15 {
		// x is definitely larger in magnitude
		return xNeg
	}
	if expY-expX > 15 {
		// y is definitely larger in magnitude
		return !xNeg
	}

	// Compare with compensation
	if expX > expY {
		// Scale x up
		hi, lo := bits.Mul64(sigX, pow10[expX-expY])
		if hi == 0 && lo == sigY {
			// Same value, compare by exponent
			if xNeg {
				return expX >= expY
			}
			return expX <= expY
		}
		// Compare scaled values
		if hi == 0 && lo < sigY {
			return !xNeg
		}
		return xNeg
	}
	// Scale y up
	hi, lo := bits.Mul64(sigY, pow10[expY-expX])
	if hi == 0 && lo == sigX {
		// Same value, compare by exponent
		if xNeg {
			return expX >= expY
		}
		return expX <= expY
	}
	// Compare scaled values
	if hi > 0 || sigX < lo {
		return !xNeg
	}
	return xNeg
}

// TotalOrderMag implements IEEE 754 totalOrderMag predicate.
// Returns TotalOrder(|x|, |y|).
func (x Decimal64Pure) TotalOrderMag(y Decimal64Pure) bool {
	// Clear sign bits and compare
	xAbs := Decimal64Pure(uint64(x) &^ bid64SignMask)
	yAbs := Decimal64Pure(uint64(y) &^ bid64SignMask)
	return xAbs.TotalOrder(yAbs)
}

// SameQuantum returns true if x and y have the same quantum (exponent).
// Both NaN or both Inf return true, otherwise exponents must match.
// Ported mechanically from Intel bid64_sameQuantum (bid64_noncomp.c)
func (x Decimal64Pure) SameQuantum(y Decimal64Pure) bool {
	xv := uint64(x)
	yv := uint64(y)

	// if both operands are NaN, return true; if just one is NaN, return false
	xNaN := (xv & bid64NaNMask) == bid64NaNMask
	yNaN := (yv & bid64NaNMask) == bid64NaNMask
	if xNaN || yNaN {
		return xNaN && yNaN
	}

	// if both operands are INF, return true; if just one is INF, return false
	xInf := (xv & bid64InfinityMask) == bid64InfinityMask
	yInf := (yv & bid64InfinityMask) == bid64InfinityMask
	if xInf || yInf {
		return xInf && yInf
	}

	// decode exponents and return true if they match
	var expX, expY int
	if (xv & bid64SpecialEncodingMask) == bid64SpecialEncodingMask {
		expX = int((xv >> 51) & 0x3ff)
	} else {
		expX = int((xv >> 53) & 0x3ff)
	}
	if (yv & bid64SpecialEncodingMask) == bid64SpecialEncodingMask {
		expY = int((yv >> 51) & 0x3ff)
	} else {
		expY = int((yv >> 53) & 0x3ff)
	}
	return expX == expY
}

// Reduce removes trailing zeros from the coefficient and adjusts the exponent.
// For zero, returns 0 with exponent 0 (preserving sign).
// For NaN and Infinity, returns the value unchanged.
func (d Decimal64Pure) Reduce() Decimal64Pure {
	sign, exp, coeff, isInf, isNaN, _ := decodeBID64(uint64(d))

	// NaN and Infinity are unchanged
	if isNaN || isInf {
		return d
	}

	// Zero: return canonical zero (exponent 0 = biased 398)
	if coeff == 0 {
		return Decimal64Pure(encodeBID64(sign, bid64ExponentBias, 0))
	}

	// Remove trailing zeros
	for coeff%10 == 0 && coeff > 0 {
		coeff /= 10
		exp++
	}

	// Check for exponent overflow (should not happen with valid input)
	if exp > bid64MaxExponent {
		exp = bid64MaxExponent
	}

	return Decimal64Pure(encodeBID64(sign, exp, coeff))
}
