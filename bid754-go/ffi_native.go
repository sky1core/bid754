//go:build cgo && bid754_native

package bid754

/*
#cgo CFLAGS: -DDECNUMDIGITS=34 -I${SRCDIR}/third_party/intel_dfp/src -I${SRCDIR}/third_party/intel_dfp/include
#cgo LDFLAGS: -ldecnumber -L${SRCDIR}/third_party/intel_dfp/lib -lbid -lm

#include <stdint.h>
#include "bid_conf.h"
#include "bid_functions.h"

static BID_UINT32 bid754_ffi_bid32_add(BID_UINT32 a, BID_UINT32 b) {
	_IDEC_flags flags = 0;
	return bid32_add(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT32 bid754_ffi_bid32_sub(BID_UINT32 a, BID_UINT32 b) {
	_IDEC_flags flags = 0;
	return bid32_sub(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT32 bid754_ffi_bid32_mul(BID_UINT32 a, BID_UINT32 b) {
	_IDEC_flags flags = 0;
	return bid32_mul(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT32 bid754_ffi_bid32_div(BID_UINT32 a, BID_UINT32 b) {
	_IDEC_flags flags = 0;
	return bid32_div(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT32 bid754_ffi_bid32_quantize(BID_UINT32 a, BID_UINT32 b) {
	_IDEC_flags flags = 0;
	return bid32_quantize(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT32 bid754_ffi_bid32_round_integral_exact(BID_UINT32 a) {
	_IDEC_flags flags = 0;
	return bid32_round_integral_exact(a, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT64 bid754_ffi_bid64_add(BID_UINT64 a, BID_UINT64 b) {
	_IDEC_flags flags = 0;
	return bid64_add(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT64 bid754_ffi_bid64_sub(BID_UINT64 a, BID_UINT64 b) {
	_IDEC_flags flags = 0;
	return bid64_sub(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT64 bid754_ffi_bid64_mul(BID_UINT64 a, BID_UINT64 b) {
	_IDEC_flags flags = 0;
	return bid64_mul(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT64 bid754_ffi_bid64_div(BID_UINT64 a, BID_UINT64 b) {
	_IDEC_flags flags = 0;
	return bid64_div(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT64 bid754_ffi_bid64_quantize(BID_UINT64 a, BID_UINT64 b) {
	_IDEC_flags flags = 0;
	return bid64_quantize(a, b, BID_ROUNDING_TO_NEAREST, &flags);
}

static BID_UINT64 bid754_ffi_bid64_round_integral_exact(BID_UINT64 a) {
	_IDEC_flags flags = 0;
	return bid64_round_integral_exact(a, BID_ROUNDING_TO_NEAREST, &flags);
}
*/
import "C"

func nativeFFIBID32Binary(function string, a uint32, b uint32) uint32 {
	switch function {
	case "bid32_add":
		return uint32(C.bid754_ffi_bid32_add(C.BID_UINT32(a), C.BID_UINT32(b)))
	case "bid32_sub":
		return uint32(C.bid754_ffi_bid32_sub(C.BID_UINT32(a), C.BID_UINT32(b)))
	case "bid32_mul":
		return uint32(C.bid754_ffi_bid32_mul(C.BID_UINT32(a), C.BID_UINT32(b)))
	case "bid32_div":
		return uint32(C.bid754_ffi_bid32_div(C.BID_UINT32(a), C.BID_UINT32(b)))
	case "bid32_quantize":
		return uint32(C.bid754_ffi_bid32_quantize(C.BID_UINT32(a), C.BID_UINT32(b)))
	default:
		panic("unsupported 32-bit ffi binary function")
	}
}

func nativeFFIBID32Unary(function string, a uint32) uint32 {
	switch function {
	case "bid32_round_integral_exact":
		return uint32(C.bid754_ffi_bid32_round_integral_exact(C.BID_UINT32(a)))
	default:
		panic("unsupported 32-bit ffi unary function")
	}
}

func nativeFFIBID64Binary(function string, a uint64, b uint64) uint64 {
	switch function {
	case "bid64_add":
		return uint64(C.bid754_ffi_bid64_add(C.BID_UINT64(a), C.BID_UINT64(b)))
	case "bid64_sub":
		return uint64(C.bid754_ffi_bid64_sub(C.BID_UINT64(a), C.BID_UINT64(b)))
	case "bid64_mul":
		return uint64(C.bid754_ffi_bid64_mul(C.BID_UINT64(a), C.BID_UINT64(b)))
	case "bid64_div":
		return uint64(C.bid754_ffi_bid64_div(C.BID_UINT64(a), C.BID_UINT64(b)))
	case "bid64_quantize":
		return uint64(C.bid754_ffi_bid64_quantize(C.BID_UINT64(a), C.BID_UINT64(b)))
	default:
		panic("unsupported 64-bit ffi binary function")
	}
}

func nativeFFIBID64Unary(function string, a uint64) uint64 {
	switch function {
	case "bid64_round_integral_exact":
		return uint64(C.bid754_ffi_bid64_round_integral_exact(C.BID_UINT64(a)))
	default:
		panic("unsupported 64-bit ffi unary function")
	}
}
