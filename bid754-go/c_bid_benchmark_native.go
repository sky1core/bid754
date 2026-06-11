//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

/*
#cgo CFLAGS: -DDECNUMDIGITS=34 -I${SRCDIR}/third_party/intel_dfp/src -I${SRCDIR}/third_party/intel_dfp/include
#cgo LDFLAGS: -ldecnumber -L${SRCDIR}/third_party/intel_dfp/lib -lbid -lm

#include <stdint.h>
#include "bid_conf.h"
#include "bid_functions.h"

static BID_UINT32 bid754_bench_c_bid32_x;
static BID_UINT32 bid754_bench_c_bid32_y;
static BID_UINT64 bid754_bench_c_bid64_x;
static BID_UINT64 bid754_bench_c_bid64_y;
static BID_UINT128 bid754_bench_c_bid128_x;
static BID_UINT128 bid754_bench_c_bid128_y;

static volatile BID_UINT32 bid754_bench_c_sink32;
static volatile BID_UINT64 bid754_bench_c_sink64;
static volatile BID_UINT64 bid754_bench_c_sink128_low;
static volatile BID_UINT64 bid754_bench_c_sink128_high;

static void bid754_bench_c_keep128(BID_UINT128 x) {
	bid754_bench_c_sink128_low = x.w[0];
	bid754_bench_c_sink128_high = x.w[1];
}

static void bid754_bench_c_init(void) {
	_IDEC_flags flags = 0;
	bid754_bench_c_bid32_x = bid32_from_string((char*)"123.456", BID_ROUNDING_TO_NEAREST, &flags);
	flags = 0;
	bid754_bench_c_bid32_y = bid32_from_string((char*)"789.012", BID_ROUNDING_TO_NEAREST, &flags);
	flags = 0;
	bid754_bench_c_bid64_x = bid64_from_string((char*)"123456789.123456789", BID_ROUNDING_TO_NEAREST, &flags);
	flags = 0;
	bid754_bench_c_bid64_y = bid64_from_string((char*)"987654321.987654321", BID_ROUNDING_TO_NEAREST, &flags);
	flags = 0;
	bid754_bench_c_bid128_x = bid128_from_string((char*)"12345678901234567890.12345678901234", BID_ROUNDING_TO_NEAREST, &flags);
	flags = 0;
	bid754_bench_c_bid128_y = bid128_from_string((char*)"98765432109876543210.98765432109876", BID_ROUNDING_TO_NEAREST, &flags);
}

static void bid754_bench_c_bid32_add(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_add(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, BID_ROUNDING_TO_NEAREST, &flags);
	}
}

static void bid754_bench_c_bid32_mul(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_mul(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, BID_ROUNDING_TO_NEAREST, &flags);
	}
}

static void bid754_bench_c_bid32_div(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_div(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, BID_ROUNDING_TO_NEAREST, &flags);
	}
}

static void bid754_bench_c_bid32_parse(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_from_string((char*)"123.456", BID_ROUNDING_TO_NEAREST, &flags);
	}
}

static void bid754_bench_c_bid32_to_string(long long n) {
	char buf[128];
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid32_to_string(buf, bid754_bench_c_bid32_x, &flags);
		bid754_bench_c_sink32 = (BID_UINT32)buf[0];
	}
}

static void bid754_bench_c_bid64_add(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_add(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags);
	}
}

static void bid754_bench_c_bid64_mul(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_mul(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags);
	}
}

static void bid754_bench_c_bid64_div(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_div(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags);
	}
}

static void bid754_bench_c_bid64_parse(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_from_string((char*)"123456789.123456789", BID_ROUNDING_TO_NEAREST, &flags);
	}
}

static void bid754_bench_c_bid64_to_string(long long n) {
	char buf[128];
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid64_to_string(buf, bid754_bench_c_bid64_x, &flags);
		bid754_bench_c_sink64 = (BID_UINT64)buf[0];
	}
}

static void bid754_bench_c_bid128_add(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_add(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags));
	}
}

static void bid754_bench_c_bid128_mul(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_mul(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags));
	}
}

static void bid754_bench_c_bid128_div(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_div(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags));
	}
}

static void bid754_bench_c_bid128_parse(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_from_string((char*)"12345678901234567890.12345678901234", BID_ROUNDING_TO_NEAREST, &flags));
	}
}

static void bid754_bench_c_bid128_to_string(long long n) {
	char buf[128];
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid128_to_string(buf, bid754_bench_c_bid128_x, &flags);
		bid754_bench_c_sink128_low = (BID_UINT64)buf[0];
	}
}
*/
import "C"

func nativeBenchCBIDInit()            { C.bid754_bench_c_init() }
func nativeBenchCBID32Add(n int)      { C.bid754_bench_c_bid32_add(C.longlong(n)) }
func nativeBenchCBID32Mul(n int)      { C.bid754_bench_c_bid32_mul(C.longlong(n)) }
func nativeBenchCBID32Div(n int)      { C.bid754_bench_c_bid32_div(C.longlong(n)) }
func nativeBenchCBID32Parse(n int)    { C.bid754_bench_c_bid32_parse(C.longlong(n)) }
func nativeBenchCBID32ToString(n int) { C.bid754_bench_c_bid32_to_string(C.longlong(n)) }

func nativeBenchCBID64Add(n int)      { C.bid754_bench_c_bid64_add(C.longlong(n)) }
func nativeBenchCBID64Mul(n int)      { C.bid754_bench_c_bid64_mul(C.longlong(n)) }
func nativeBenchCBID64Div(n int)      { C.bid754_bench_c_bid64_div(C.longlong(n)) }
func nativeBenchCBID64Parse(n int)    { C.bid754_bench_c_bid64_parse(C.longlong(n)) }
func nativeBenchCBID64ToString(n int) { C.bid754_bench_c_bid64_to_string(C.longlong(n)) }

func nativeBenchCBID128Add(n int)      { C.bid754_bench_c_bid128_add(C.longlong(n)) }
func nativeBenchCBID128Mul(n int)      { C.bid754_bench_c_bid128_mul(C.longlong(n)) }
func nativeBenchCBID128Div(n int)      { C.bid754_bench_c_bid128_div(C.longlong(n)) }
func nativeBenchCBID128Parse(n int)    { C.bid754_bench_c_bid128_parse(C.longlong(n)) }
func nativeBenchCBID128ToString(n int) { C.bid754_bench_c_bid128_to_string(C.longlong(n)) }
