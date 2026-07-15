//go:build cgo && bid754_native
// +build cgo,bid754_native

package bid754

/*
#cgo CFLAGS: -DDECNUMDIGITS=34 -I${SRCDIR}/../devtools/third_party/intel_dfp/src -I${SRCDIR}/../devtools/third_party/intel_dfp/include
#cgo LDFLAGS: -ldecnumber -L${SRCDIR}/../devtools/third_party/intel_dfp/lib -lbid -lm

#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include "bid_conf.h"
#include "bid_functions.h"

static BID_UINT32 bid754_bench_c_bid32_x;
static BID_UINT32 bid754_bench_c_bid32_y;
static BID_UINT32 bid754_bench_c_bid32_z;
static BID_UINT64 bid754_bench_c_bid64_x;
static BID_UINT64 bid754_bench_c_bid64_y;
static BID_UINT64 bid754_bench_c_bid64_z;
static BID_UINT128 bid754_bench_c_bid128_x;
static BID_UINT128 bid754_bench_c_bid128_y;
static BID_UINT128 bid754_bench_c_bid128_z;
static BID_UINT32 bid754_bench_c_bid32_integer;
static BID_UINT64 bid754_bench_c_bid64_integer;
static BID_UINT128 bid754_bench_c_bid128_integer;
static BID_SINT64 bid754_bench_c_integer_operand;
static int bid754_bench_c_scale_exponent;
static char bid754_bench_c_bid32_x_text[128];
static char bid754_bench_c_bid32_y_text[128];
static char bid754_bench_c_bid32_z_text[128];
static char bid754_bench_c_bid64_x_text[128];
static char bid754_bench_c_bid64_y_text[128];
static char bid754_bench_c_bid64_z_text[128];
static char bid754_bench_c_bid128_x_text[128];
static char bid754_bench_c_bid128_y_text[128];
static char bid754_bench_c_bid128_z_text[128];

static volatile BID_UINT32 bid754_bench_c_sink32;
static volatile BID_UINT64 bid754_bench_c_sink64;
static volatile BID_UINT64 bid754_bench_c_sink128_low;
static volatile BID_UINT64 bid754_bench_c_sink128_high;
static volatile BID_SINT64 bid754_bench_c_sink_int64;
static volatile _IDEC_flags bid754_bench_c_sink_flags;
static char bid754_bench_c_sink_string[128];
static char bid754_bench_c_oracle_string[128];

static void bid754_bench_c_keep128(BID_UINT128 x) {
	bid754_bench_c_sink128_low = x.w[0];
	bid754_bench_c_sink128_high = x.w[1];
}

static int bid754_bench_c_copy_text(char *dst, size_t capacity, const char *src) {
	size_t len = strlen(src);
	if (len + 1 > capacity) {
		return 0;
	}
	memcpy(dst, src, len + 1);
	return 1;
}

static void bid754_bench_c_reset_sinks(void) {
	bid754_bench_c_sink32 = UINT32_C(0xa5a5a5a5);
	bid754_bench_c_sink64 = UINT64_C(0xa5a5a5a5a5a5a5a5);
	bid754_bench_c_sink128_low = UINT64_C(0xa5a5a5a5a5a5a5a5);
	bid754_bench_c_sink128_high = UINT64_C(0x5a5a5a5a5a5a5a5a);
	bid754_bench_c_sink_int64 = INT64_MIN;
	bid754_bench_c_sink_flags = (_IDEC_flags)UINT32_C(0xffffffff);
	memcpy(bid754_bench_c_sink_string, "<unset>", sizeof("<unset>"));
}

static BID_UINT32 bid754_bench_c_get_sink32(void) {
	return bid754_bench_c_sink32;
}

static BID_UINT64 bid754_bench_c_get_sink64(void) {
	return bid754_bench_c_sink64;
}

static BID_UINT64 bid754_bench_c_get_sink128_low(void) {
	return bid754_bench_c_sink128_low;
}

static BID_UINT64 bid754_bench_c_get_sink128_high(void) {
	return bid754_bench_c_sink128_high;
}

static BID_SINT64 bid754_bench_c_get_sink_int64(void) {
	return bid754_bench_c_sink_int64;
}

static _IDEC_flags bid754_bench_c_get_sink_flags(void) {
	return bid754_bench_c_sink_flags;
}

static const char *bid754_bench_c_get_sink_string(void) {
	return bid754_bench_c_sink_string;
}

static const char *bid754_bench_c_bid32_to_string_once(void) {
	_IDEC_flags flags = 0;
	bid32_to_string(bid754_bench_c_oracle_string, bid754_bench_c_bid32_x, &flags);
	return bid754_bench_c_oracle_string;
}

static const char *bid754_bench_c_bid64_to_string_once(void) {
	_IDEC_flags flags = 0;
	bid64_to_string(bid754_bench_c_oracle_string, bid754_bench_c_bid64_x, &flags);
	return bid754_bench_c_oracle_string;
}

static const char *bid754_bench_c_bid128_to_string_once(void) {
	_IDEC_flags flags = 0;
	bid128_to_string(bid754_bench_c_oracle_string, bid754_bench_c_bid128_x, &flags);
	return bid754_bench_c_oracle_string;
}

static int bid754_bench_c_init(
	const char *bid32_x_text,
	const char *bid32_y_text,
	const char *bid32_z_text,
	const char *bid64_x_text,
	const char *bid64_y_text,
	const char *bid64_z_text,
	const char *bid128_x_text,
	const char *bid128_y_text,
	const char *bid128_z_text,
	long long integer_operand,
	int scale_exponent
) {
	if (!bid754_bench_c_copy_text(bid754_bench_c_bid32_x_text, sizeof(bid754_bench_c_bid32_x_text), bid32_x_text) ||
		!bid754_bench_c_copy_text(bid754_bench_c_bid32_y_text, sizeof(bid754_bench_c_bid32_y_text), bid32_y_text) ||
		!bid754_bench_c_copy_text(bid754_bench_c_bid32_z_text, sizeof(bid754_bench_c_bid32_z_text), bid32_z_text) ||
		!bid754_bench_c_copy_text(bid754_bench_c_bid64_x_text, sizeof(bid754_bench_c_bid64_x_text), bid64_x_text) ||
		!bid754_bench_c_copy_text(bid754_bench_c_bid64_y_text, sizeof(bid754_bench_c_bid64_y_text), bid64_y_text) ||
		!bid754_bench_c_copy_text(bid754_bench_c_bid64_z_text, sizeof(bid754_bench_c_bid64_z_text), bid64_z_text) ||
		!bid754_bench_c_copy_text(bid754_bench_c_bid128_x_text, sizeof(bid754_bench_c_bid128_x_text), bid128_x_text) ||
		!bid754_bench_c_copy_text(bid754_bench_c_bid128_y_text, sizeof(bid754_bench_c_bid128_y_text), bid128_y_text) ||
		!bid754_bench_c_copy_text(bid754_bench_c_bid128_z_text, sizeof(bid754_bench_c_bid128_z_text), bid128_z_text)) {
		return 0;
	}

	_IDEC_flags flags = 0;
	bid754_bench_c_bid32_x = bid32_from_string(bid754_bench_c_bid32_x_text, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid32_isFinite(bid754_bench_c_bid32_x)) {
		return 0;
	}
	flags = 0;
	bid754_bench_c_bid32_y = bid32_from_string(bid754_bench_c_bid32_y_text, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid32_isFinite(bid754_bench_c_bid32_y)) {
		return 0;
	}
	flags = 0;
	bid754_bench_c_bid32_z = bid32_from_string(bid754_bench_c_bid32_z_text, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid32_isFinite(bid754_bench_c_bid32_z)) {
		return 0;
	}
	flags = 0;
	bid754_bench_c_bid64_x = bid64_from_string(bid754_bench_c_bid64_x_text, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid64_isFinite(bid754_bench_c_bid64_x)) {
		return 0;
	}
	flags = 0;
	bid754_bench_c_bid64_y = bid64_from_string(bid754_bench_c_bid64_y_text, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid64_isFinite(bid754_bench_c_bid64_y)) {
		return 0;
	}
	flags = 0;
	bid754_bench_c_bid64_z = bid64_from_string(bid754_bench_c_bid64_z_text, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid64_isFinite(bid754_bench_c_bid64_z)) {
		return 0;
	}
	flags = 0;
	bid754_bench_c_bid128_x = bid128_from_string(bid754_bench_c_bid128_x_text, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid128_isFinite(bid754_bench_c_bid128_x)) {
		return 0;
	}
	flags = 0;
	bid754_bench_c_bid128_y = bid128_from_string(bid754_bench_c_bid128_y_text, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid128_isFinite(bid754_bench_c_bid128_y)) {
		return 0;
	}
	flags = 0;
	bid754_bench_c_bid128_z = bid128_from_string(bid754_bench_c_bid128_z_text, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid128_isFinite(bid754_bench_c_bid128_z)) {
		return 0;
	}
	if (bid32_isSigned(bid754_bench_c_bid32_x) ||
		bid64_isSigned(bid754_bench_c_bid64_x) ||
		bid128_isSigned(bid754_bench_c_bid128_x) ||
		bid32_isZero(bid754_bench_c_bid32_x) || bid32_isZero(bid754_bench_c_bid32_y) ||
		bid64_isZero(bid754_bench_c_bid64_x) || bid64_isZero(bid754_bench_c_bid64_y) ||
		bid128_isZero(bid754_bench_c_bid128_x) || bid128_isZero(bid754_bench_c_bid128_y)) {
		return 0;
	}

	bid754_bench_c_integer_operand = (BID_SINT64)integer_operand;
	bid754_bench_c_scale_exponent = scale_exponent;
	flags = 0;
	bid754_bench_c_bid32_integer = bid32_from_int64(bid754_bench_c_integer_operand, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid32_isFinite(bid754_bench_c_bid32_integer)) {
		return 0;
	}
	flags = 0;
	bid754_bench_c_bid64_integer = bid64_from_int64(bid754_bench_c_integer_operand, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid64_isFinite(bid754_bench_c_bid64_integer)) {
		return 0;
	}
	bid754_bench_c_bid128_integer = bid128_from_int64(bid754_bench_c_integer_operand);
	if (!bid128_isFinite(bid754_bench_c_bid128_integer)) {
		return 0;
	}
	flags = 0;
	if (bid32_to_int64_rnint(bid754_bench_c_bid32_integer, &flags) != bid754_bench_c_integer_operand || flags != 0) {
		return 0;
	}
	flags = 0;
	if (bid64_to_int64_rnint(bid754_bench_c_bid64_integer, &flags) != bid754_bench_c_integer_operand || flags != 0) {
		return 0;
	}
	flags = 0;
	if (bid128_to_int64_rnint(bid754_bench_c_bid128_integer, &flags) != bid754_bench_c_integer_operand || flags != 0) {
		return 0;
	}
	flags = 0;
	BID_UINT32 scaled32 = bid32_scalbn(bid754_bench_c_bid32_x, bid754_bench_c_scale_exponent, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid32_isFinite(scaled32)) {
		return 0;
	}
	flags = 0;
	BID_UINT64 scaled64 = bid64_scalbn(bid754_bench_c_bid64_x, bid754_bench_c_scale_exponent, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid64_isFinite(scaled64)) {
		return 0;
	}
	flags = 0;
	BID_UINT128 scaled128 = bid128_scalbn(bid754_bench_c_bid128_x, bid754_bench_c_scale_exponent, BID_ROUNDING_TO_NEAREST, &flags);
	if (flags != 0 || !bid128_isFinite(scaled128)) {
		return 0;
	}
	return 1;
}

static void bid754_bench_c_bid32_add(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_add(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid32_mul(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_mul(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid32_div(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_div(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid32_fma(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_fma(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, bid754_bench_c_bid32_z, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid32_sqrt(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_sqrt(bid754_bench_c_bid32_x, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid32_parse(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink32 = bid32_from_string(bid754_bench_c_bid32_x_text, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid32_to_string(long long n) {
	char buf[128] = {0};
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid32_to_string(buf, bid754_bench_c_bid32_x, &flags);
		bid754_bench_c_sink32 = (BID_UINT32)buf[0];
		bid754_bench_c_sink_flags = flags;
	}
	if (n > 0) {
		bid754_bench_c_copy_text(bid754_bench_c_sink_string, sizeof(bid754_bench_c_sink_string), buf);
	}
}

static void bid754_bench_c_bid64_add(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_add(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid64_mul(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_mul(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid64_div(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_div(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid64_fma(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_fma(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, bid754_bench_c_bid64_z, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid64_sqrt(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_sqrt(bid754_bench_c_bid64_x, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid64_parse(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_sink64 = bid64_from_string(bid754_bench_c_bid64_x_text, BID_ROUNDING_TO_NEAREST, &flags);
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid64_to_string(long long n) {
	char buf[128] = {0};
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid64_to_string(buf, bid754_bench_c_bid64_x, &flags);
		bid754_bench_c_sink64 = (BID_UINT64)buf[0];
		bid754_bench_c_sink_flags = flags;
	}
	if (n > 0) {
		bid754_bench_c_copy_text(bid754_bench_c_sink_string, sizeof(bid754_bench_c_sink_string), buf);
	}
}

static void bid754_bench_c_bid128_add(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_add(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags));
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid128_mul(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_mul(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags));
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid128_div(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_div(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags));
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid128_fma(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_fma(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, bid754_bench_c_bid128_z, BID_ROUNDING_TO_NEAREST, &flags));
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid128_sqrt(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_sqrt(bid754_bench_c_bid128_x, BID_ROUNDING_TO_NEAREST, &flags));
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid128_parse(long long n) {
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid754_bench_c_keep128(bid128_from_string(bid754_bench_c_bid128_x_text, BID_ROUNDING_TO_NEAREST, &flags));
		bid754_bench_c_sink_flags = flags;
	}
}

static void bid754_bench_c_bid128_to_string(long long n) {
	char buf[128] = {0};
	for (long long i = 0; i < n; i++) {
		_IDEC_flags flags = 0;
		bid128_to_string(buf, bid754_bench_c_bid128_x, &flags);
		bid754_bench_c_sink128_low = (BID_UINT64)buf[0];
		bid754_bench_c_sink_flags = flags;
	}
	if (n > 0) {
		bid754_bench_c_copy_text(bid754_bench_c_sink_string, sizeof(bid754_bench_c_sink_string), buf);
	}
}

#define BID754_DEFINE_BENCH32(name, expression) \
	static void name(long long n) { \
		for (long long i = 0; i < n; i++) { \
			_IDEC_flags flags = 0; \
			bid754_bench_c_sink32 = (expression); \
			bid754_bench_c_sink_flags = flags; \
		} \
	}

#define BID754_DEFINE_BENCH64(name, expression) \
	static void name(long long n) { \
		for (long long i = 0; i < n; i++) { \
			_IDEC_flags flags = 0; \
			bid754_bench_c_sink64 = (expression); \
			bid754_bench_c_sink_flags = flags; \
		} \
	}

#define BID754_DEFINE_BENCH128(name, expression) \
	static void name(long long n) { \
		for (long long i = 0; i < n; i++) { \
			_IDEC_flags flags = 0; \
			bid754_bench_c_keep128((expression)); \
			bid754_bench_c_sink_flags = flags; \
		} \
	}

#define BID754_DEFINE_BENCH128_NO_FLAGS(name, expression) \
	static void name(long long n) { \
		for (long long i = 0; i < n; i++) { \
			bid754_bench_c_keep128((expression)); \
		} \
	}

#define BID754_DEFINE_BENCH_INT64(name, expression) \
	static void name(long long n) { \
		for (long long i = 0; i < n; i++) { \
			_IDEC_flags flags = 0; \
			bid754_bench_c_sink_int64 = (expression); \
			bid754_bench_c_sink_flags = flags; \
		} \
	}

BID754_DEFINE_BENCH32(bid754_bench_c_bid32_sub,
	bid32_sub(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid32_remainder,
	bid32_rem(bid754_bench_c_bid32_y, bid754_bench_c_bid32_x, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid32_fmod,
	bid32_fmod(bid754_bench_c_bid32_y, bid754_bench_c_bid32_x, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid32_quantize,
	bid32_quantize(bid754_bench_c_bid32_x, bid754_bench_c_bid32_integer, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid32_scaleb,
	bid32_scalbn(bid754_bench_c_bid32_x, bid754_bench_c_scale_exponent, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid32_quiet_equal,
	(BID_UINT32)bid32_quiet_equal(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid32_minnum,
	bid32_minnum(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid32_maxnum,
	bid32_maxnum(bid754_bench_c_bid32_x, bid754_bench_c_bid32_y, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid32_from_int64,
	bid32_from_int64(bid754_bench_c_integer_operand, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH_INT64(bid754_bench_c_bid32_to_int64,
	bid32_to_int64_rnint(bid754_bench_c_bid32_integer, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid32_to_decimal64,
	bid32_to_bid64(bid754_bench_c_bid32_x, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid32_to_decimal128,
	bid32_to_bid128(bid754_bench_c_bid32_x, &flags))

BID754_DEFINE_BENCH64(bid754_bench_c_bid64_sub,
	bid64_sub(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_remainder,
	bid64_rem(bid754_bench_c_bid64_y, bid754_bench_c_bid64_x, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_fmod,
	bid64_fmod(bid754_bench_c_bid64_y, bid754_bench_c_bid64_x, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_quantize,
	bid64_quantize(bid754_bench_c_bid64_x, bid754_bench_c_bid64_integer, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_scaleb,
	bid64_scalbn(bid754_bench_c_bid64_x, bid754_bench_c_scale_exponent, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_quiet_equal,
	(BID_UINT64)bid64_quiet_equal(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_minnum,
	bid64_minnum(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_maxnum,
	bid64_maxnum(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_from_int64,
	bid64_from_int64(bid754_bench_c_integer_operand, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH_INT64(bid754_bench_c_bid64_to_int64,
	bid64_to_int64_rnint(bid754_bench_c_bid64_integer, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid64_to_decimal32,
	bid64_to_bid32(bid754_bench_c_bid64_x, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid64_to_decimal128,
	bid64_to_bid128(bid754_bench_c_bid64_x, &flags))

BID754_DEFINE_BENCH128(bid754_bench_c_bid128_sub,
	bid128_sub(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_remainder,
	bid128_rem(bid754_bench_c_bid128_y, bid754_bench_c_bid128_x, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_fmod,
	bid128_fmod(bid754_bench_c_bid128_y, bid754_bench_c_bid128_x, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_quantize,
	bid128_quantize(bid754_bench_c_bid128_x, bid754_bench_c_bid128_integer, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_scaleb,
	bid128_scalbn(bid754_bench_c_bid128_x, bid754_bench_c_scale_exponent, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid128_quiet_equal,
	(BID_UINT64)bid128_quiet_equal(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_minnum,
	bid128_minnum(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_maxnum,
	bid128_maxnum(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, &flags))
BID754_DEFINE_BENCH128_NO_FLAGS(bid754_bench_c_bid128_from_int64,
	bid128_from_int64(bid754_bench_c_integer_operand))
BID754_DEFINE_BENCH_INT64(bid754_bench_c_bid128_to_int64,
	bid128_to_int64_rnint(bid754_bench_c_bid128_integer, &flags))
BID754_DEFINE_BENCH32(bid754_bench_c_bid128_to_decimal32,
	bid128_to_bid32(bid754_bench_c_bid128_x, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid128_to_decimal64,
	bid128_to_bid64(bid754_bench_c_bid128_x, BID_ROUNDING_TO_NEAREST, &flags))

BID754_DEFINE_BENCH64(bid754_bench_c_bid64_dq_add,
	bid64dq_add(bid754_bench_c_bid64_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_qd_add,
	bid64qd_add(bid754_bench_c_bid128_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_qq_add,
	bid64qq_add(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_dq_sub,
	bid64dq_sub(bid754_bench_c_bid64_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_qd_sub,
	bid64qd_sub(bid754_bench_c_bid128_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_qq_sub,
	bid64qq_sub(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_dq_mul,
	bid64dq_mul(bid754_bench_c_bid64_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_qd_mul,
	bid64qd_mul(bid754_bench_c_bid128_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_qq_mul,
	bid64qq_mul(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_dq_div,
	bid64dq_div(bid754_bench_c_bid64_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_qd_div,
	bid64qd_div(bid754_bench_c_bid128_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH64(bid754_bench_c_bid64_qq_div,
	bid64qq_div(bid754_bench_c_bid128_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))

BID754_DEFINE_BENCH128(bid754_bench_c_bid128_dd_add,
	bid128dd_add(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_dq_add,
	bid128dq_add(bid754_bench_c_bid64_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_qd_add,
	bid128qd_add(bid754_bench_c_bid128_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_dd_sub,
	bid128dd_sub(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_dq_sub,
	bid128dq_sub(bid754_bench_c_bid64_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_qd_sub,
	bid128qd_sub(bid754_bench_c_bid128_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_dd_mul,
	bid128dd_mul(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_dq_mul,
	bid128dq_mul(bid754_bench_c_bid64_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_qd_mul,
	bid128qd_mul(bid754_bench_c_bid128_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_dd_div,
	bid128dd_div(bid754_bench_c_bid64_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_dq_div,
	bid128dq_div(bid754_bench_c_bid64_x, bid754_bench_c_bid128_y, BID_ROUNDING_TO_NEAREST, &flags))
BID754_DEFINE_BENCH128(bid754_bench_c_bid128_qd_div,
	bid128qd_div(bid754_bench_c_bid128_x, bid754_bench_c_bid64_y, BID_ROUNDING_TO_NEAREST, &flags))

#undef BID754_DEFINE_BENCH32
#undef BID754_DEFINE_BENCH64
#undef BID754_DEFINE_BENCH128
#undef BID754_DEFINE_BENCH128_NO_FLAGS
#undef BID754_DEFINE_BENCH_INT64
*/
import "C"

import "unsafe"

type nativeCBenchmarkSnapshot struct {
	BID32      uint32
	BID64      uint64
	BID128Low  uint64
	BID128High uint64
	Int64      int64
	Flags      uint32
	String     string
}

func nativeBenchCBIDResetSinks() {
	C.bid754_bench_c_reset_sinks()
}

func nativeBenchCBIDSnapshot() nativeCBenchmarkSnapshot {
	return nativeCBenchmarkSnapshot{
		BID32:      uint32(C.bid754_bench_c_get_sink32()),
		BID64:      uint64(C.bid754_bench_c_get_sink64()),
		BID128Low:  uint64(C.bid754_bench_c_get_sink128_low()),
		BID128High: uint64(C.bid754_bench_c_get_sink128_high()),
		Int64:      int64(C.bid754_bench_c_get_sink_int64()),
		Flags:      uint32(C.bid754_bench_c_get_sink_flags()),
		String:     C.GoString(C.bid754_bench_c_get_sink_string()),
	}
}

func nativeBenchCBID32ToStringOnce() string {
	return C.GoString(C.bid754_bench_c_bid32_to_string_once())
}

func nativeBenchCBID64ToStringOnce() string {
	return C.GoString(C.bid754_bench_c_bid64_to_string_once())
}

func nativeBenchCBID128ToStringOnce() string {
	return C.GoString(C.bid754_bench_c_bid128_to_string_once())
}

func nativeBenchCBIDInit(bid32X, bid32Y, bid32Z, bid64X, bid64Y, bid64Z, bid128X, bid128Y, bid128Z string, integerOperand int64, scaleExponent int) bool {
	cBid32X := C.CString(bid32X)
	cBid32Y := C.CString(bid32Y)
	cBid32Z := C.CString(bid32Z)
	cBid64X := C.CString(bid64X)
	cBid64Y := C.CString(bid64Y)
	cBid64Z := C.CString(bid64Z)
	cBid128X := C.CString(bid128X)
	cBid128Y := C.CString(bid128Y)
	cBid128Z := C.CString(bid128Z)
	defer C.free(unsafe.Pointer(cBid32X))
	defer C.free(unsafe.Pointer(cBid32Y))
	defer C.free(unsafe.Pointer(cBid32Z))
	defer C.free(unsafe.Pointer(cBid64X))
	defer C.free(unsafe.Pointer(cBid64Y))
	defer C.free(unsafe.Pointer(cBid64Z))
	defer C.free(unsafe.Pointer(cBid128X))
	defer C.free(unsafe.Pointer(cBid128Y))
	defer C.free(unsafe.Pointer(cBid128Z))
	return C.bid754_bench_c_init(cBid32X, cBid32Y, cBid32Z, cBid64X, cBid64Y, cBid64Z, cBid128X, cBid128Y, cBid128Z, C.longlong(integerOperand), C.int(scaleExponent)) != 0
}

func nativeBenchCBID32Add(n int)          { C.bid754_bench_c_bid32_add(C.longlong(n)) }
func nativeBenchCBID32Mul(n int)          { C.bid754_bench_c_bid32_mul(C.longlong(n)) }
func nativeBenchCBID32Div(n int)          { C.bid754_bench_c_bid32_div(C.longlong(n)) }
func nativeBenchCBID32Fma(n int)          { C.bid754_bench_c_bid32_fma(C.longlong(n)) }
func nativeBenchCBID32Sqrt(n int)         { C.bid754_bench_c_bid32_sqrt(C.longlong(n)) }
func nativeBenchCBID32Parse(n int)        { C.bid754_bench_c_bid32_parse(C.longlong(n)) }
func nativeBenchCBID32ToString(n int)     { C.bid754_bench_c_bid32_to_string(C.longlong(n)) }
func nativeBenchCBID32Sub(n int)          { C.bid754_bench_c_bid32_sub(C.longlong(n)) }
func nativeBenchCBID32Remainder(n int)    { C.bid754_bench_c_bid32_remainder(C.longlong(n)) }
func nativeBenchCBID32Fmod(n int)         { C.bid754_bench_c_bid32_fmod(C.longlong(n)) }
func nativeBenchCBID32Quantize(n int)     { C.bid754_bench_c_bid32_quantize(C.longlong(n)) }
func nativeBenchCBID32ScaleB(n int)       { C.bid754_bench_c_bid32_scaleb(C.longlong(n)) }
func nativeBenchCBID32QuietEqual(n int)   { C.bid754_bench_c_bid32_quiet_equal(C.longlong(n)) }
func nativeBenchCBID32MinNum(n int)       { C.bid754_bench_c_bid32_minnum(C.longlong(n)) }
func nativeBenchCBID32MaxNum(n int)       { C.bid754_bench_c_bid32_maxnum(C.longlong(n)) }
func nativeBenchCBID32FromInt64(n int)    { C.bid754_bench_c_bid32_from_int64(C.longlong(n)) }
func nativeBenchCBID32ToInt64(n int)      { C.bid754_bench_c_bid32_to_int64(C.longlong(n)) }
func nativeBenchCBID32ToDecimal64(n int)  { C.bid754_bench_c_bid32_to_decimal64(C.longlong(n)) }
func nativeBenchCBID32ToDecimal128(n int) { C.bid754_bench_c_bid32_to_decimal128(C.longlong(n)) }

func nativeBenchCBID64Add(n int)          { C.bid754_bench_c_bid64_add(C.longlong(n)) }
func nativeBenchCBID64Mul(n int)          { C.bid754_bench_c_bid64_mul(C.longlong(n)) }
func nativeBenchCBID64Div(n int)          { C.bid754_bench_c_bid64_div(C.longlong(n)) }
func nativeBenchCBID64Fma(n int)          { C.bid754_bench_c_bid64_fma(C.longlong(n)) }
func nativeBenchCBID64Sqrt(n int)         { C.bid754_bench_c_bid64_sqrt(C.longlong(n)) }
func nativeBenchCBID64Parse(n int)        { C.bid754_bench_c_bid64_parse(C.longlong(n)) }
func nativeBenchCBID64ToString(n int)     { C.bid754_bench_c_bid64_to_string(C.longlong(n)) }
func nativeBenchCBID64Sub(n int)          { C.bid754_bench_c_bid64_sub(C.longlong(n)) }
func nativeBenchCBID64Remainder(n int)    { C.bid754_bench_c_bid64_remainder(C.longlong(n)) }
func nativeBenchCBID64Fmod(n int)         { C.bid754_bench_c_bid64_fmod(C.longlong(n)) }
func nativeBenchCBID64Quantize(n int)     { C.bid754_bench_c_bid64_quantize(C.longlong(n)) }
func nativeBenchCBID64ScaleB(n int)       { C.bid754_bench_c_bid64_scaleb(C.longlong(n)) }
func nativeBenchCBID64QuietEqual(n int)   { C.bid754_bench_c_bid64_quiet_equal(C.longlong(n)) }
func nativeBenchCBID64MinNum(n int)       { C.bid754_bench_c_bid64_minnum(C.longlong(n)) }
func nativeBenchCBID64MaxNum(n int)       { C.bid754_bench_c_bid64_maxnum(C.longlong(n)) }
func nativeBenchCBID64FromInt64(n int)    { C.bid754_bench_c_bid64_from_int64(C.longlong(n)) }
func nativeBenchCBID64ToInt64(n int)      { C.bid754_bench_c_bid64_to_int64(C.longlong(n)) }
func nativeBenchCBID64ToDecimal32(n int)  { C.bid754_bench_c_bid64_to_decimal32(C.longlong(n)) }
func nativeBenchCBID64ToDecimal128(n int) { C.bid754_bench_c_bid64_to_decimal128(C.longlong(n)) }

func nativeBenchCBID128Add(n int)         { C.bid754_bench_c_bid128_add(C.longlong(n)) }
func nativeBenchCBID128Mul(n int)         { C.bid754_bench_c_bid128_mul(C.longlong(n)) }
func nativeBenchCBID128Div(n int)         { C.bid754_bench_c_bid128_div(C.longlong(n)) }
func nativeBenchCBID128Fma(n int)         { C.bid754_bench_c_bid128_fma(C.longlong(n)) }
func nativeBenchCBID128Sqrt(n int)        { C.bid754_bench_c_bid128_sqrt(C.longlong(n)) }
func nativeBenchCBID128Parse(n int)       { C.bid754_bench_c_bid128_parse(C.longlong(n)) }
func nativeBenchCBID128ToString(n int)    { C.bid754_bench_c_bid128_to_string(C.longlong(n)) }
func nativeBenchCBID128Sub(n int)         { C.bid754_bench_c_bid128_sub(C.longlong(n)) }
func nativeBenchCBID128Remainder(n int)   { C.bid754_bench_c_bid128_remainder(C.longlong(n)) }
func nativeBenchCBID128Fmod(n int)        { C.bid754_bench_c_bid128_fmod(C.longlong(n)) }
func nativeBenchCBID128Quantize(n int)    { C.bid754_bench_c_bid128_quantize(C.longlong(n)) }
func nativeBenchCBID128ScaleB(n int)      { C.bid754_bench_c_bid128_scaleb(C.longlong(n)) }
func nativeBenchCBID128QuietEqual(n int)  { C.bid754_bench_c_bid128_quiet_equal(C.longlong(n)) }
func nativeBenchCBID128MinNum(n int)      { C.bid754_bench_c_bid128_minnum(C.longlong(n)) }
func nativeBenchCBID128MaxNum(n int)      { C.bid754_bench_c_bid128_maxnum(C.longlong(n)) }
func nativeBenchCBID128FromInt64(n int)   { C.bid754_bench_c_bid128_from_int64(C.longlong(n)) }
func nativeBenchCBID128ToInt64(n int)     { C.bid754_bench_c_bid128_to_int64(C.longlong(n)) }
func nativeBenchCBID128ToDecimal32(n int) { C.bid754_bench_c_bid128_to_decimal32(C.longlong(n)) }
func nativeBenchCBID128ToDecimal64(n int) { C.bid754_bench_c_bid128_to_decimal64(C.longlong(n)) }

func nativeBenchCBID64DQAdd(n int) { C.bid754_bench_c_bid64_dq_add(C.longlong(n)) }
func nativeBenchCBID64QDAdd(n int) { C.bid754_bench_c_bid64_qd_add(C.longlong(n)) }
func nativeBenchCBID64QQAdd(n int) { C.bid754_bench_c_bid64_qq_add(C.longlong(n)) }
func nativeBenchCBID64DQSub(n int) { C.bid754_bench_c_bid64_dq_sub(C.longlong(n)) }
func nativeBenchCBID64QDSub(n int) { C.bid754_bench_c_bid64_qd_sub(C.longlong(n)) }
func nativeBenchCBID64QQSub(n int) { C.bid754_bench_c_bid64_qq_sub(C.longlong(n)) }
func nativeBenchCBID64DQMul(n int) { C.bid754_bench_c_bid64_dq_mul(C.longlong(n)) }
func nativeBenchCBID64QDMul(n int) { C.bid754_bench_c_bid64_qd_mul(C.longlong(n)) }
func nativeBenchCBID64QQMul(n int) { C.bid754_bench_c_bid64_qq_mul(C.longlong(n)) }
func nativeBenchCBID64DQDiv(n int) { C.bid754_bench_c_bid64_dq_div(C.longlong(n)) }
func nativeBenchCBID64QDDiv(n int) { C.bid754_bench_c_bid64_qd_div(C.longlong(n)) }
func nativeBenchCBID64QQDiv(n int) { C.bid754_bench_c_bid64_qq_div(C.longlong(n)) }

func nativeBenchCBID128DDAdd(n int) { C.bid754_bench_c_bid128_dd_add(C.longlong(n)) }
func nativeBenchCBID128DQAdd(n int) { C.bid754_bench_c_bid128_dq_add(C.longlong(n)) }
func nativeBenchCBID128QDAdd(n int) { C.bid754_bench_c_bid128_qd_add(C.longlong(n)) }
func nativeBenchCBID128DDSub(n int) { C.bid754_bench_c_bid128_dd_sub(C.longlong(n)) }
func nativeBenchCBID128DQSub(n int) { C.bid754_bench_c_bid128_dq_sub(C.longlong(n)) }
func nativeBenchCBID128QDSub(n int) { C.bid754_bench_c_bid128_qd_sub(C.longlong(n)) }
func nativeBenchCBID128DDMul(n int) { C.bid754_bench_c_bid128_dd_mul(C.longlong(n)) }
func nativeBenchCBID128DQMul(n int) { C.bid754_bench_c_bid128_dq_mul(C.longlong(n)) }
func nativeBenchCBID128QDMul(n int) { C.bid754_bench_c_bid128_qd_mul(C.longlong(n)) }
func nativeBenchCBID128DDDiv(n int) { C.bid754_bench_c_bid128_dd_div(C.longlong(n)) }
func nativeBenchCBID128DQDiv(n int) { C.bid754_bench_c_bid128_dq_div(C.longlong(n)) }
func nativeBenchCBID128QDDiv(n int) { C.bid754_bench_c_bid128_qd_div(C.longlong(n)) }
