package apiemit

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sky1core/bid754/devtools/internal/genmarker"
)

// The exhaustive set of emittable shape templates and the Go signature form each
// assumes lives in shapeSigs (apiemit.go); resolveClosure rejects a manifest
// emit rule naming a shape absent there. This is the exhaustive-shape check: wrapper
// bodies come only from these templates, never from free manifest text.

// portPath resolves a census bidgo function name to its generated Rust module
// and function. A shape referencing a function outside this table fails, so a
// new wrapper family must declare its module path here explicitly.
var portPath = map[string]struct{ module, fn string }{
	"Bid64FromString":   {"bid64_from_string", "bid64_from_string"},
	"Bid64ToString":     {"string64", "bid64_to_string"},
	"Bid64Add":          {"add64", "bid64_add"},
	"Bid64AddWithFlags": {"add64", "bid64_add_with_flags"},
	"Bid64dqAdd":        {"bid128_add", "bid64dq_add"},
	"Bid64qdAdd":        {"bid128_add", "bid64qd_add"},
	"Bid64qqAdd":        {"bid128_add", "bid64qq_add"},
	"Bid64dqSub":        {"bid128_add", "bid64dq_sub"},
	"Bid64qdSub":        {"bid128_add", "bid64qd_sub"},
	"Bid64qqSub":        {"bid128_add", "bid64qq_sub"},
	"Bid64dqMul":        {"bid128_mul", "bid64dq_mul"},
	"Bid64qdMul":        {"bid128_mul", "bid64qd_mul"},
	"Bid64qqMul":        {"bid128_mul", "bid64qq_mul"},
	"Bid64dqDiv":        {"div64", "bid64dq_div"},
	"Bid64qdDiv":        {"div64", "bid64qd_div"},
	"Bid64qqDiv":        {"div64", "bid64qq_div"},
	"Bid64ddqFma":       {"bid128_fma", "bid64ddq_fma"},
	"Bid64dqdFma":       {"bid128_fma", "bid64dqd_fma"},
	"Bid64dqqFma":       {"bid128_fma", "bid64dqq_fma"},
	"Bid64qddFma":       {"bid128_fma", "bid64qdd_fma"},
	"Bid64qdqFma":       {"bid128_fma", "bid64qdq_fma"},
	"Bid64qqdFma":       {"bid128_fma", "bid64qqd_fma"},
	"Bid64qqqFma":       {"bid128_fma", "bid64qqq_fma"},
	"Bid64qSqrt":        {"sqrt64", "bid64q_sqrt"},

	// Decimal64 arithmetic, miscellaneous, predicate, and conversion families.
	"Bid64Abs":                      {"noncomp64", "bid64_abs"},
	"Bid64Class":                    {"noncomp64", "bid64_class"},
	"Bid64Copy":                     {"noncomp64", "bid64_copy"},
	"Bid64CopySign":                 {"noncomp64", "bid64_copy_sign"},
	"Bid64Div":                      {"div64", "bid64_div"},
	"Bid64DivWithFlags":             {"div64", "bid64_div_with_flags"},
	"Bid64Fma":                      {"fma64", "bid64_fma"},
	"Bid64Fmod":                     {"fmod64", "bid64_fmod"},
	"Bid64ILogb":                    {"logb64", "bid64_i_logb"},
	"Bid64IsCanonical":              {"noncomp64", "bid64_is_canonical"},
	"Bid64IsFinite":                 {"noncomp64", "bid64_is_finite"},
	"Bid64IsInf":                    {"noncomp64", "bid64_is_inf"},
	"Bid64IsNaN":                    {"noncomp64", "bid64_is_na_n"},
	"Bid64IsNormal":                 {"noncomp64", "bid64_is_normal"},
	"Bid64IsSigned":                 {"noncomp64", "bid64_is_signed"},
	"Bid64IsSignaling":              {"noncomp64", "bid64_is_signaling"},
	"Bid64IsSubnormal":              {"noncomp64", "bid64_is_subnormal"},
	"Bid64IsZero":                   {"noncomp64", "bid64_is_zero"},
	"Bid64Logb":                     {"logb64", "bid64_logb"},
	"Bid64MaxNum":                   {"minmax64", "bid64_max_num"},
	"Bid64MaxNumMag":                {"minmax64", "bid64_max_num_mag"},
	"Bid64MinNum":                   {"minmax64", "bid64_min_num"},
	"Bid64MinNumMag":                {"minmax64", "bid64_min_num_mag"},
	"Bid64Mul":                      {"mul64", "bid64_mul"},
	"Bid64MulWithFlags":             {"mul64", "bid64_mul_with_flags"},
	"Bid64Negate":                   {"noncomp64", "bid64_negate"},
	"Bid64NextDown":                 {"next64", "bid64_next_down"},
	"Bid64NextUp":                   {"next64", "bid64_next_up"},
	"Bid64NextToward":               {"nexttoward64", "bid64_next_toward"},
	"Bid64Quantize":                 {"quantize64", "bid64_quantize"},
	"Bid64Radix":                    {"noncomp64", "bid64_radix"},
	"Bid64Rem":                      {"rem64", "bid64_rem"},
	"Bid64RoundIntegralExact":       {"round_integral64", "bid64_round_integral_exact"},
	"Bid64RoundIntegralNearestAway": {"round_integral64", "bid64_round_integral_nearest_away"},
	"Bid64RoundIntegralNearestEven": {"round_integral64", "bid64_round_integral_nearest_even"},
	"Bid64RoundIntegralNegative":    {"round_integral64", "bid64_round_integral_negative"},
	"Bid64RoundIntegralPositive":    {"round_integral64", "bid64_round_integral_positive"},
	"Bid64RoundIntegralZero":        {"round_integral64", "bid64_round_integral_zero"},
	"Bid64SameQuantum":              {"noncomp64", "bid64_same_quantum"},
	"Bid64Scalbn":                   {"scalb64", "bid64_scalbn"},
	"Bid64Scalbln":                  {"scalb64", "bid64_scalbln"},
	"Bid64Sqrt":                     {"sqrt64", "bid64_sqrt"},
	"Bid64Sub":                      {"add64", "bid64_sub"},
	"Bid64SubWithFlags":             {"add64", "bid64_sub_with_flags"},
	"Bid64ToBinary32":               {"to_binary64", "bid64_to_binary32"},
	"Bid64ToBinary64":               {"to_binary64", "bid64_to_binary64"},
	"Bid64ToBinary128":              {"to_binary64", "bid64_to_binary128"},
	"Bid64ToBid128":                 {"to_bid12864", "bid64_to_bid128"},
	"Bid64ToBid32":                  {"to_bid3264", "bid64_to_bid32"},
	"Bid64TotalOrder":               {"noncomp64", "bid64_total_order"},
	"Bid64TotalOrderMag":            {"noncomp64", "bid64_total_order_mag"},
	"Bid64FromInt32":                {"convert64", "bid64_from_int32"},
	"Bid64FromUint32":               {"convert64", "bid64_from_uint32"},
	"Bid64FromInt64":                {"convert64", "bid64_from_int64"},
	"Bid64FromUint64":               {"convert64", "bid64_from_uint64"},

	// Decimal64 quiet comparisons.
	"Bid64QuietEqual":            {"compare64", "bid64_quiet_equal"},
	"Bid64QuietNotEqual":         {"compare64", "bid64_quiet_not_equal"},
	"Bid64QuietGreater":          {"compare64", "bid64_quiet_greater"},
	"Bid64QuietGreaterEqual":     {"compare64", "bid64_quiet_greater_equal"},
	"Bid64QuietGreaterUnordered": {"compare64", "bid64_quiet_greater_unordered"},
	"Bid64QuietLess":             {"compare64", "bid64_quiet_less"},
	"Bid64QuietLessEqual":        {"compare64", "bid64_quiet_less_equal"},
	"Bid64QuietLessUnordered":    {"compare64", "bid64_quiet_less_unordered"},
	"Bid64QuietNotGreater":       {"compare64", "bid64_quiet_not_greater"},
	"Bid64QuietNotLess":          {"compare64", "bid64_quiet_not_less"},
	"Bid64QuietOrdered":          {"compare64", "bid64_quiet_ordered"},
	"Bid64QuietUnordered":        {"compare64", "bid64_quiet_unordered"},

	// Decimal64 signaling comparisons; Equal/NotEqual are composed.
	"Bid64SignalingGreater":          {"compare64", "bid64_signaling_greater"},
	"Bid64SignalingGreaterEqual":     {"compare64", "bid64_signaling_greater_equal"},
	"Bid64SignalingGreaterUnordered": {"compare64", "bid64_signaling_greater_unordered"},
	"Bid64SignalingLess":             {"noncomp64", "bid64_signaling_less"},
	"Bid64SignalingLessEqual":        {"compare64", "bid64_signaling_less_equal"},
	"Bid64SignalingLessUnordered":    {"compare64", "bid64_signaling_less_unordered"},
	"Bid64SignalingNotGreater":       {"compare64", "bid64_signaling_not_greater"},
	"Bid64SignalingNotLess":          {"compare64", "bid64_signaling_not_less"},

	// Decimal64 ConvertToInt<N>/ConvertToUint<N>(Exact) mode-dispatch leaves.
	"Bid64ToInt8Rnint":     {"to_int_small", "bid64_to_int8_rnint"},
	"Bid64ToInt8Rninta":    {"to_int_small", "bid64_to_int8_rninta"},
	"Bid64ToInt8Int":       {"to_int_small", "bid64_to_int8_int"},
	"Bid64ToInt8Ceil":      {"to_int_small", "bid64_to_int8_ceil"},
	"Bid64ToInt8Floor":     {"to_int_small", "bid64_to_int8_floor"},
	"Bid64ToInt8Xrnint":    {"to_int_small", "bid64_to_int8_xrnint"},
	"Bid64ToInt8Xrninta":   {"to_int_small", "bid64_to_int8_xrninta"},
	"Bid64ToInt8Xint":      {"to_int_small", "bid64_to_int8_xint"},
	"Bid64ToInt8Xceil":     {"to_int_small", "bid64_to_int8_xceil"},
	"Bid64ToInt8Xfloor":    {"to_int_small", "bid64_to_int8_xfloor"},
	"Bid64ToInt16Rnint":    {"to_int_small", "bid64_to_int16_rnint"},
	"Bid64ToInt16Rninta":   {"to_int_small", "bid64_to_int16_rninta"},
	"Bid64ToInt16Int":      {"to_int_small", "bid64_to_int16_int"},
	"Bid64ToInt16Ceil":     {"to_int_small", "bid64_to_int16_ceil"},
	"Bid64ToInt16Floor":    {"to_int_small", "bid64_to_int16_floor"},
	"Bid64ToInt16Xrnint":   {"to_int_small", "bid64_to_int16_xrnint"},
	"Bid64ToInt16Xrninta":  {"to_int_small", "bid64_to_int16_xrninta"},
	"Bid64ToInt16Xint":     {"to_int_small", "bid64_to_int16_xint"},
	"Bid64ToInt16Xceil":    {"to_int_small", "bid64_to_int16_xceil"},
	"Bid64ToInt16Xfloor":   {"to_int_small", "bid64_to_int16_xfloor"},
	"Bid64ToInt32Rnint":    {"to_int32", "bid64_to_int32_rnint"},
	"Bid64ToInt32Rninta":   {"to_int32_rninta", "bid64_to_int32_rninta"},
	"Bid64ToInt32Int":      {"to_int32_int", "bid64_to_int32_int"},
	"Bid64ToInt32Ceil":     {"to_int32_ceil", "bid64_to_int32_ceil"},
	"Bid64ToInt32Floor":    {"to_int32_floor", "bid64_to_int32_floor"},
	"Bid64ToInt32Xrnint":   {"to_int32", "bid64_to_int32_xrnint"},
	"Bid64ToInt32Xrninta":  {"to_int32_rninta", "bid64_to_int32_xrninta"},
	"Bid64ToInt32Xint":     {"to_int32_int", "bid64_to_int32_xint"},
	"Bid64ToInt32Xceil":    {"to_int32_ceil", "bid64_to_int32_xceil"},
	"Bid64ToInt32Xfloor":   {"to_int32_floor", "bid64_to_int32_xfloor"},
	"Bid64ToInt64Rnint":    {"to_int64_signed", "bid64_to_int64_rnint"},
	"Bid64ToInt64Rninta":   {"to_int64_signed", "bid64_to_int64_rninta"},
	"Bid64ToInt64Int":      {"to_int64_int", "bid64_to_int64_int"},
	"Bid64ToInt64Ceil":     {"to_int64_ceil", "bid64_to_int64_ceil"},
	"Bid64ToInt64Floor":    {"to_int64_floor", "bid64_to_int64_floor"},
	"Bid64ToInt64Xrnint":   {"to_int64_signed", "bid64_to_int64_xrnint"},
	"Bid64ToInt64Xrninta":  {"to_int64_signed", "bid64_to_int64_xrninta"},
	"Bid64ToInt64Xint":     {"to_int64_int", "bid64_to_int64_xint"},
	"Bid64ToInt64Xceil":    {"to_int64_ceil", "bid64_to_int64_xceil"},
	"Bid64ToInt64Xfloor":   {"to_int64_floor", "bid64_to_int64_xfloor"},
	"Bid64ToUint8Rnint":    {"to_uint_small", "bid64_to_uint8_rnint"},
	"Bid64ToUint8Rninta":   {"to_uint_small", "bid64_to_uint8_rninta"},
	"Bid64ToUint8Int":      {"to_uint_small", "bid64_to_uint8_int"},
	"Bid64ToUint8Ceil":     {"to_uint_small", "bid64_to_uint8_ceil"},
	"Bid64ToUint8Floor":    {"to_uint_small", "bid64_to_uint8_floor"},
	"Bid64ToUint8Xrnint":   {"to_uint_small", "bid64_to_uint8_xrnint"},
	"Bid64ToUint8Xrninta":  {"to_uint_small", "bid64_to_uint8_xrninta"},
	"Bid64ToUint8Xint":     {"to_uint_small", "bid64_to_uint8_xint"},
	"Bid64ToUint8Xceil":    {"to_uint_small", "bid64_to_uint8_xceil"},
	"Bid64ToUint8Xfloor":   {"to_uint_small", "bid64_to_uint8_xfloor"},
	"Bid64ToUint16Rnint":   {"to_uint_small", "bid64_to_uint16_rnint"},
	"Bid64ToUint16Rninta":  {"to_uint_small", "bid64_to_uint16_rninta"},
	"Bid64ToUint16Int":     {"to_uint_small", "bid64_to_uint16_int"},
	"Bid64ToUint16Ceil":    {"to_uint_small", "bid64_to_uint16_ceil"},
	"Bid64ToUint16Floor":   {"to_uint_small", "bid64_to_uint16_floor"},
	"Bid64ToUint16Xrnint":  {"to_uint_small", "bid64_to_uint16_xrnint"},
	"Bid64ToUint16Xrninta": {"to_uint_small", "bid64_to_uint16_xrninta"},
	"Bid64ToUint16Xint":    {"to_uint_small", "bid64_to_uint16_xint"},
	"Bid64ToUint16Xceil":   {"to_uint_small", "bid64_to_uint16_xceil"},
	"Bid64ToUint16Xfloor":  {"to_uint_small", "bid64_to_uint16_xfloor"},
	"Bid64ToUint32Rnint":   {"to_uint32_rnint", "bid64_to_uint32_rnint"},
	"Bid64ToUint32Rninta":  {"to_uint32_rninta", "bid64_to_uint32_rninta"},
	"Bid64ToUint32Int":     {"to_uint32_int", "bid64_to_uint32_int"},
	"Bid64ToUint32Ceil":    {"to_uint32_ceil", "bid64_to_uint32_ceil"},
	"Bid64ToUint32Floor":   {"to_uint32_floor", "bid64_to_uint32_floor"},
	"Bid64ToUint32Xrnint":  {"to_uint32_rnint", "bid64_to_uint32_xrnint"},
	"Bid64ToUint32Xrninta": {"to_uint32_rninta", "bid64_to_uint32_xrninta"},
	"Bid64ToUint32Xint":    {"to_uint32_int", "bid64_to_uint32_xint"},
	"Bid64ToUint32Xceil":   {"to_uint32_ceil", "bid64_to_uint32_xceil"},
	"Bid64ToUint32Xfloor":  {"to_uint32_floor", "bid64_to_uint32_xfloor"},
	"Bid64ToUint64Rnint":   {"to_uint64_rnint", "bid64_to_uint64_rnint"},
	"Bid64ToUint64Rninta":  {"to_uint64_rninta", "bid64_to_uint64_rninta"},
	"Bid64ToUint64Int":     {"to_uint64_int", "bid64_to_uint64_int"},
	"Bid64ToUint64Ceil":    {"to_uint64_ceil", "bid64_to_uint64_ceil"},
	"Bid64ToUint64Floor":   {"to_uint64_floor", "bid64_to_uint64_floor"},
	"Bid64ToUint64Xrnint":  {"to_uint64_rnint", "bid64_to_uint64_xrnint"},
	"Bid64ToUint64Xrninta": {"to_uint64_rninta", "bid64_to_uint64_xrninta"},
	"Bid64ToUint64Xint":    {"to_uint64_int", "bid64_to_uint64_xint"},
	"Bid64ToUint64Xceil":   {"to_uint64_ceil", "bid64_to_uint64_xceil"},
	"Bid64ToUint64Xfloor":  {"to_uint64_floor", "bid64_to_uint64_xfloor"},

	// Decimal32 arithmetic, miscellaneous, predicate, and conversion families.
	// (single-call families). Unlike Decimal64, whose bidgo functions land in
	// many small per-family generated modules, Decimal32's generated tree
	// consolidates most of these into a handful of files (bid32_exports,
	// bid32_status, bid32_noncomp, bid32_misc, bid32_to_int); each module
	// value below was verified against the actual generated::* source (not
	// assumed from the Decimal64 naming precedent).
	"Bid32Abs":                       {"bid32_noncomp", "bid32_abs"},
	"Bid32Add":                       {"bid32_exports", "bid32_add"},
	"Bid32AddWithFlags":              {"bid32_status", "bid32_add_with_flags"},
	"Bid32Class":                     {"bid32_misc", "bid32_class"},
	"Bid32Copy":                      {"bid32_noncomp", "bid32_copy"},
	"Bid32CopySign":                  {"bid32_noncomp", "bid32_copy_sign"},
	"Bid32Div":                       {"bid32_exports", "bid32_div"},
	"Bid32DivWithFlags":              {"bid32_status", "bid32_div_with_flags"},
	"Bid32Fma":                       {"bid32_fma", "bid32_fma"},
	"Bid32Fmod":                      {"bid32_misc", "bid32_fmod"},
	"Bid32FromInt32":                 {"bid32_from_int", "bid32_from_int32"},
	"Bid32FromInt64":                 {"bid32_to_int", "bid32_from_int64"},
	"Bid32FromStringRaw":             {"bid32_string", "bid32_from_string_raw"},
	"Bid32FromUint32":                {"bid32_from_int", "bid32_from_uint32"},
	"Bid32FromUint64":                {"bid32_to_int", "bid32_from_uint64"},
	"Bid32ILogb":                     {"bid32_logb", "bid32_i_logb"},
	"Bid32IsCanonical":               {"bid32_noncomp", "bid32_is_canonical"},
	"Bid32IsFinite":                  {"bid32_noncomp", "bid32_is_finite"},
	"Bid32IsInf":                     {"bid32_exports", "bid32_is_inf"},
	"Bid32IsNaN":                     {"bid32_exports", "bid32_is_na_n"},
	"Bid32IsNormal":                  {"bid32_noncomp", "bid32_is_normal"},
	"Bid32IsSignaling":               {"bid32_noncomp", "bid32_is_signaling"},
	"Bid32IsSigned":                  {"bid32_noncomp", "bid32_is_signed"},
	"Bid32IsSubnormal":               {"bid32_noncomp", "bid32_is_subnormal"},
	"Bid32IsZero":                    {"bid32_exports", "bid32_is_zero"},
	"Bid32Logb":                      {"bid32_logb", "bid32_logb"},
	"Bid32MaxNumMagWithFlags":        {"bid32_status", "bid32_max_num_mag_with_flags"},
	"Bid32MaxNumWithFlags":           {"bid32_status", "bid32_max_num_with_flags"},
	"Bid32MinNumMagWithFlags":        {"bid32_status", "bid32_min_num_mag_with_flags"},
	"Bid32MinNumWithFlags":           {"bid32_status", "bid32_min_num_with_flags"},
	"Bid32Mul":                       {"bid32_exports", "bid32_mul"},
	"Bid32MulWithFlags":              {"bid32_status", "bid32_mul_with_flags"},
	"Bid32Negate":                    {"bid32_noncomp", "bid32_negate"},
	"Bid32NextDown":                  {"bid32_next", "bid32_next_down"},
	"Bid32NextToward":                {"bid32_misc", "bid32_next_toward"},
	"Bid32NextUp":                    {"bid32_next", "bid32_next_up"},
	"Bid32Quantize":                  {"bid32_quantize", "bid32_quantize"},
	"Bid32QuietEqual":                {"compare32", "bid32_quiet_equal"},
	"Bid32QuietGreater":              {"compare32", "bid32_quiet_greater"},
	"Bid32QuietGreaterEqual":         {"compare32", "bid32_quiet_greater_equal"},
	"Bid32QuietGreaterUnordered":     {"compare32", "bid32_quiet_greater_unordered"},
	"Bid32QuietLess":                 {"compare32", "bid32_quiet_less"},
	"Bid32QuietLessEqual":            {"compare32", "bid32_quiet_less_equal"},
	"Bid32QuietLessUnordered":        {"compare32", "bid32_quiet_less_unordered"},
	"Bid32QuietNotEqual":             {"compare32", "bid32_quiet_not_equal"},
	"Bid32QuietNotGreater":           {"compare32", "bid32_quiet_not_greater"},
	"Bid32QuietNotLess":              {"compare32", "bid32_quiet_not_less"},
	"Bid32QuietOrdered":              {"compare32", "bid32_quiet_ordered"},
	"Bid32QuietUnordered":            {"compare32", "bid32_quiet_unordered"},
	"Bid32Radix":                     {"bid32_noncomp", "bid32_radix"},
	"Bid32Rem":                       {"bid32_rem", "bid32_rem"},
	"Bid32RoundIntegralExact":        {"bid32_round_integral", "bid32_round_integral_exact"},
	"Bid32RoundIntegralNearestAway":  {"bid32_round_integral", "bid32_round_integral_nearest_away"},
	"Bid32RoundIntegralNearestEven":  {"bid32_round_integral", "bid32_round_integral_nearest_even"},
	"Bid32RoundIntegralNegative":     {"bid32_round_integral", "bid32_round_integral_negative"},
	"Bid32RoundIntegralPositive":     {"bid32_round_integral", "bid32_round_integral_positive"},
	"Bid32RoundIntegralZero":         {"bid32_round_integral", "bid32_round_integral_zero"},
	"Bid32SameQuantum":               {"bid32_exports", "bid32_same_quantum"},
	"Bid32Scalbn":                    {"bid32_scalb", "bid32_scalbn"},
	"Bid32ScalblnWithFlags":          {"bid32_status", "bid32_scalbln_with_flags"},
	"Bid32SignalingGreater":          {"compare32", "bid32_signaling_greater"},
	"Bid32SignalingGreaterEqual":     {"compare32", "bid32_signaling_greater_equal"},
	"Bid32SignalingGreaterUnordered": {"compare32", "bid32_signaling_greater_unordered"},
	"Bid32SignalingLess":             {"compare32", "bid32_signaling_less"},
	"Bid32SignalingLessEqual":        {"compare32", "bid32_signaling_less_equal"},
	"Bid32SignalingLessUnordered":    {"compare32", "bid32_signaling_less_unordered"},
	"Bid32SignalingNotGreater":       {"compare32", "bid32_signaling_not_greater"},
	"Bid32SignalingNotLess":          {"compare32", "bid32_signaling_not_less"},
	"Bid32Sqrt":                      {"bid32_sqrt", "bid32_sqrt"},
	"Bid32Sub":                       {"bid32_exports", "bid32_sub"},
	"Bid32SubWithFlags":              {"bid32_status", "bid32_sub_with_flags"},
	"Bid32ToBid128":                  {"bid32_misc", "bid32_to_bid128"},
	"Bid32ToBid64":                   {"bid32_to_bid64", "bid32_to_bid64"},
	"Bid32ToBinary128":               {"bid32_misc", "bid32_to_binary128"},
	"Bid32ToBinary32":                {"bid32_misc", "bid32_to_binary32"},
	"Bid32ToBinary64":                {"bid32_misc", "bid32_to_binary64"},
	"Bid32ToString":                  {"bid32_exports", "bid32_to_string"},
	"Bid32TotalOrder":                {"bid32_noncomp", "bid32_total_order"},
	"Bid32TotalOrderMag":             {"bid32_noncomp", "bid32_total_order_mag"},

	// Decimal32 ConvertToInt<N>/ConvertToUint<N>(Exact) mode-dispatch leaves.
	// leaves (8 widths x 5 modes x {plain,exact}). Unlike Decimal64 (many
	// small per-mode-family files), every one of these lands in the single
	// bid32_to_int module.
	"Bid32ToInt16Ceil":     {"bid32_to_int", "bid32_to_int16_ceil"},
	"Bid32ToInt16Floor":    {"bid32_to_int", "bid32_to_int16_floor"},
	"Bid32ToInt16Int":      {"bid32_to_int", "bid32_to_int16_int"},
	"Bid32ToInt16Rnint":    {"bid32_to_int", "bid32_to_int16_rnint"},
	"Bid32ToInt16Rninta":   {"bid32_to_int", "bid32_to_int16_rninta"},
	"Bid32ToInt16Xceil":    {"bid32_to_int", "bid32_to_int16_xceil"},
	"Bid32ToInt16Xfloor":   {"bid32_to_int", "bid32_to_int16_xfloor"},
	"Bid32ToInt16Xint":     {"bid32_to_int", "bid32_to_int16_xint"},
	"Bid32ToInt16Xrnint":   {"bid32_to_int", "bid32_to_int16_xrnint"},
	"Bid32ToInt16Xrninta":  {"bid32_to_int", "bid32_to_int16_xrninta"},
	"Bid32ToInt32Ceil":     {"bid32_to_int", "bid32_to_int32_ceil"},
	"Bid32ToInt32Floor":    {"bid32_to_int", "bid32_to_int32_floor"},
	"Bid32ToInt32Int":      {"bid32_to_int", "bid32_to_int32_int"},
	"Bid32ToInt32Rnint":    {"bid32_to_int", "bid32_to_int32_rnint"},
	"Bid32ToInt32Rninta":   {"bid32_to_int", "bid32_to_int32_rninta"},
	"Bid32ToInt32Xceil":    {"bid32_to_int", "bid32_to_int32_xceil"},
	"Bid32ToInt32Xfloor":   {"bid32_to_int", "bid32_to_int32_xfloor"},
	"Bid32ToInt32Xint":     {"bid32_to_int", "bid32_to_int32_xint"},
	"Bid32ToInt32Xrnint":   {"bid32_to_int", "bid32_to_int32_xrnint"},
	"Bid32ToInt32Xrninta":  {"bid32_to_int", "bid32_to_int32_xrninta"},
	"Bid32ToInt64Ceil":     {"bid32_to_int", "bid32_to_int64_ceil"},
	"Bid32ToInt64Floor":    {"bid32_to_int", "bid32_to_int64_floor"},
	"Bid32ToInt64Int":      {"bid32_to_int", "bid32_to_int64_int"},
	"Bid32ToInt64Rnint":    {"bid32_to_int", "bid32_to_int64_rnint"},
	"Bid32ToInt64Rninta":   {"bid32_to_int", "bid32_to_int64_rninta"},
	"Bid32ToInt64Xceil":    {"bid32_to_int", "bid32_to_int64_xceil"},
	"Bid32ToInt64Xfloor":   {"bid32_to_int", "bid32_to_int64_xfloor"},
	"Bid32ToInt64Xint":     {"bid32_to_int", "bid32_to_int64_xint"},
	"Bid32ToInt64Xrnint":   {"bid32_to_int", "bid32_to_int64_xrnint"},
	"Bid32ToInt64Xrninta":  {"bid32_to_int", "bid32_to_int64_xrninta"},
	"Bid32ToInt8Ceil":      {"bid32_to_int", "bid32_to_int8_ceil"},
	"Bid32ToInt8Floor":     {"bid32_to_int", "bid32_to_int8_floor"},
	"Bid32ToInt8Int":       {"bid32_to_int", "bid32_to_int8_int"},
	"Bid32ToInt8Rnint":     {"bid32_to_int", "bid32_to_int8_rnint"},
	"Bid32ToInt8Rninta":    {"bid32_to_int", "bid32_to_int8_rninta"},
	"Bid32ToInt8Xceil":     {"bid32_to_int", "bid32_to_int8_xceil"},
	"Bid32ToInt8Xfloor":    {"bid32_to_int", "bid32_to_int8_xfloor"},
	"Bid32ToInt8Xint":      {"bid32_to_int", "bid32_to_int8_xint"},
	"Bid32ToInt8Xrnint":    {"bid32_to_int", "bid32_to_int8_xrnint"},
	"Bid32ToInt8Xrninta":   {"bid32_to_int", "bid32_to_int8_xrninta"},
	"Bid32ToUint16Ceil":    {"bid32_to_int", "bid32_to_uint16_ceil"},
	"Bid32ToUint16Floor":   {"bid32_to_int", "bid32_to_uint16_floor"},
	"Bid32ToUint16Int":     {"bid32_to_int", "bid32_to_uint16_int"},
	"Bid32ToUint16Rnint":   {"bid32_to_int", "bid32_to_uint16_rnint"},
	"Bid32ToUint16Rninta":  {"bid32_to_int", "bid32_to_uint16_rninta"},
	"Bid32ToUint16Xceil":   {"bid32_to_int", "bid32_to_uint16_xceil"},
	"Bid32ToUint16Xfloor":  {"bid32_to_int", "bid32_to_uint16_xfloor"},
	"Bid32ToUint16Xint":    {"bid32_to_int", "bid32_to_uint16_xint"},
	"Bid32ToUint16Xrnint":  {"bid32_to_int", "bid32_to_uint16_xrnint"},
	"Bid32ToUint16Xrninta": {"bid32_to_int", "bid32_to_uint16_xrninta"},
	"Bid32ToUint32Ceil":    {"bid32_to_int", "bid32_to_uint32_ceil"},
	"Bid32ToUint32Floor":   {"bid32_to_int", "bid32_to_uint32_floor"},
	"Bid32ToUint32Int":     {"bid32_to_int", "bid32_to_uint32_int"},
	"Bid32ToUint32Rnint":   {"bid32_to_int", "bid32_to_uint32_rnint"},
	"Bid32ToUint32Rninta":  {"bid32_to_int", "bid32_to_uint32_rninta"},
	"Bid32ToUint32Xceil":   {"bid32_to_int", "bid32_to_uint32_xceil"},
	"Bid32ToUint32Xfloor":  {"bid32_to_int", "bid32_to_uint32_xfloor"},
	"Bid32ToUint32Xint":    {"bid32_to_int", "bid32_to_uint32_xint"},
	"Bid32ToUint32Xrnint":  {"bid32_to_int", "bid32_to_uint32_xrnint"},
	"Bid32ToUint32Xrninta": {"bid32_to_int", "bid32_to_uint32_xrninta"},
	"Bid32ToUint64Ceil":    {"bid32_to_int", "bid32_to_uint64_ceil"},
	"Bid32ToUint64Floor":   {"bid32_to_int", "bid32_to_uint64_floor"},
	"Bid32ToUint64Int":     {"bid32_to_int", "bid32_to_uint64_int"},
	"Bid32ToUint64Rnint":   {"bid32_to_int", "bid32_to_uint64_rnint"},
	"Bid32ToUint64Rninta":  {"bid32_to_int", "bid32_to_uint64_rninta"},
	"Bid32ToUint64Xceil":   {"bid32_to_int", "bid32_to_uint64_xceil"},
	"Bid32ToUint64Xfloor":  {"bid32_to_int", "bid32_to_uint64_xfloor"},
	"Bid32ToUint64Xint":    {"bid32_to_int", "bid32_to_uint64_xint"},
	"Bid32ToUint64Xrnint":  {"bid32_to_int", "bid32_to_uint64_xrnint"},
	"Bid32ToUint64Xrninta": {"bid32_to_int", "bid32_to_uint64_xrninta"},
	"Bid32ToUint8Ceil":     {"bid32_to_int", "bid32_to_uint8_ceil"},
	"Bid32ToUint8Floor":    {"bid32_to_int", "bid32_to_uint8_floor"},
	"Bid32ToUint8Int":      {"bid32_to_int", "bid32_to_uint8_int"},
	"Bid32ToUint8Rnint":    {"bid32_to_int", "bid32_to_uint8_rnint"},
	"Bid32ToUint8Rninta":   {"bid32_to_int", "bid32_to_uint8_rninta"},
	"Bid32ToUint8Xceil":    {"bid32_to_int", "bid32_to_uint8_xceil"},
	"Bid32ToUint8Xfloor":   {"bid32_to_int", "bid32_to_uint8_xfloor"},
	"Bid32ToUint8Xint":     {"bid32_to_int", "bid32_to_uint8_xint"},
	"Bid32ToUint8Xrnint":   {"bid32_to_int", "bid32_to_uint8_xrnint"},
	"Bid32ToUint8Xrninta":  {"bid32_to_int", "bid32_to_uint8_xrninta"},

	// Decimal128 arithmetic, miscellaneous, predicate, and conversion families.
	// Every module value below follows the generated bid128 signatures. Those
	// signatures mix tuple returns with a status-pointer output parameter;
	// portPfpsf records the latter functions explicitly.
	"Bid128Abs":                       {"bid128_noncomp", "bid128_abs"},
	"Bid128Add":                       {"bid128_add", "bid128_add"},
	"Bid128ddAdd":                     {"bid128_add", "bid128dd_add"},
	"Bid128dqAdd":                     {"bid128_add", "bid128dq_add"},
	"Bid128qdAdd":                     {"bid128_add", "bid128qd_add"},
	"Bid128ddSub":                     {"bid128_add", "bid128dd_sub"},
	"Bid128dqSub":                     {"bid128_add", "bid128dq_sub"},
	"Bid128qdSub":                     {"bid128_add", "bid128qd_sub"},
	"Bid128Class":                     {"bid128_misc", "bid128_class"},
	"Bid128Copy":                      {"bid128_noncomp", "bid128_copy"},
	"Bid128CopySign":                  {"bid128_noncomp", "bid128_copy_sign"},
	"Bid128Div":                       {"bid128_div", "bid128_div"},
	"Bid128ddDiv":                     {"bid128_div", "bid128dd_div"},
	"Bid128dqDiv":                     {"bid128_div", "bid128dq_div"},
	"Bid128qdDiv":                     {"bid128_div", "bid128qd_div"},
	"Bid128Fma":                       {"bid128_fma", "bid128_fma"},
	"Bid128dddFma":                    {"bid128_fma", "bid128ddd_fma"},
	"Bid128ddqFma":                    {"bid128_fma", "bid128ddq_fma"},
	"Bid128dqdFma":                    {"bid128_fma", "bid128dqd_fma"},
	"Bid128dqqFma":                    {"bid128_fma", "bid128dqq_fma"},
	"Bid128qddFma":                    {"bid128_fma", "bid128qdd_fma"},
	"Bid128qdqFma":                    {"bid128_fma", "bid128qdq_fma"},
	"Bid128qqdFma":                    {"bid128_fma", "bid128qqd_fma"},
	"Bid128dSqrt":                     {"bid128_sqrt", "bid128d_sqrt"},
	"Bid128Fmod":                      {"bid128_rem", "bid128_fmod"},
	"Bid128FromInt32":                 {"bid128_from_int", "bid128_from_int32"},
	"Bid128FromInt64":                 {"bid128_from_int", "bid128_from_int64"},
	"Bid128FromString":                {"bid128_string", "bid128_from_string"},
	"Bid128FromUint32":                {"bid128_from_int", "bid128_from_uint32"},
	"Bid128FromUint64":                {"bid128_from_int", "bid128_from_uint64"},
	"Bid128Ilogb":                     {"bid128_misc", "bid128_ilogb"},
	"Bid128IsCanonical":               {"bid128_noncomp", "bid128_is_canonical"},
	"Bid128IsFinite":                  {"bid128_noncomp", "bid128_is_finite"},
	"Bid128IsInf":                     {"bid128_internal", "bid128_is_inf"},
	"Bid128IsNaN":                     {"bid128_internal", "bid128_is_na_n"},
	"Bid128IsNormal":                  {"bid128_noncomp", "bid128_is_normal"},
	"Bid128IsSignaling":               {"bid128_noncomp", "bid128_is_signaling"},
	"Bid128IsSigned":                  {"bid128_noncomp", "bid128_is_signed"},
	"Bid128IsSubnormal":               {"bid128_noncomp", "bid128_is_subnormal"},
	"Bid128IsZero":                    {"bid128_internal", "bid128_is_zero"},
	"Bid128Logb":                      {"bid128_misc", "bid128_logb"},
	"Bid128Maxnum":                    {"bid128_minmax", "bid128_maxnum"},
	"Bid128MaxnumMag":                 {"bid128_minmax", "bid128_maxnum_mag"},
	"Bid128Minnum":                    {"bid128_minmax", "bid128_minnum"},
	"Bid128MinnumMag":                 {"bid128_minmax", "bid128_minnum_mag"},
	"Bid128Mul":                       {"bid128_mul", "bid128_mul"},
	"Bid128ddMul":                     {"bid128_mul", "bid128dd_mul"},
	"Bid128dqMul":                     {"bid128_mul", "bid128dq_mul"},
	"Bid128qdMul":                     {"bid128_mul", "bid128qd_mul"},
	"Bid128Negate":                    {"bid128_noncomp", "bid128_negate"},
	"Bid128NextDown":                  {"bid128_next", "bid128_next_down"},
	"Bid128NextToward":                {"bid128_next", "bid128_next_toward"},
	"Bid128NextUp":                    {"bid128_next", "bid128_next_up"},
	"Bid128Quantize":                  {"bid128_quantize", "bid128_quantize"},
	"Bid128QuietEqual":                {"bid128_compare", "bid128_quiet_equal"},
	"Bid128QuietGreater":              {"bid128_compare", "bid128_quiet_greater"},
	"Bid128QuietGreaterEqual":         {"bid128_compare", "bid128_quiet_greater_equal"},
	"Bid128QuietGreaterUnordered":     {"bid128_compare", "bid128_quiet_greater_unordered"},
	"Bid128QuietLess":                 {"bid128_compare", "bid128_quiet_less"},
	"Bid128QuietLessEqual":            {"bid128_compare", "bid128_quiet_less_equal"},
	"Bid128QuietLessUnordered":        {"bid128_compare", "bid128_quiet_less_unordered"},
	"Bid128QuietNotEqual":             {"bid128_compare", "bid128_quiet_not_equal"},
	"Bid128QuietNotGreater":           {"bid128_compare", "bid128_quiet_not_greater"},
	"Bid128QuietNotLess":              {"bid128_compare", "bid128_quiet_not_less"},
	"Bid128QuietOrdered":              {"bid128_compare", "bid128_quiet_ordered"},
	"Bid128QuietUnordered":            {"bid128_compare", "bid128_quiet_unordered"},
	"Bid128Radix":                     {"bid128_noncomp", "bid128_radix"},
	"Bid128Rem":                       {"bid128_rem", "bid128_rem"},
	"Bid128RoundIntegralExact":        {"bid128_round_integral", "bid128_round_integral_exact"},
	"Bid128RoundIntegralNearestAway":  {"bid128_round_integral", "bid128_round_integral_nearest_away"},
	"Bid128RoundIntegralNearestEven":  {"bid128_round_integral", "bid128_round_integral_nearest_even"},
	"Bid128RoundIntegralNegative":     {"bid128_round_integral", "bid128_round_integral_negative"},
	"Bid128RoundIntegralPositive":     {"bid128_round_integral", "bid128_round_integral_positive"},
	"Bid128RoundIntegralZero":         {"bid128_round_integral", "bid128_round_integral_zero"},
	"Bid128SameQuantum":               {"bid128_noncomp", "bid128_same_quantum"},
	"Bid128Scalbn":                    {"bid128_misc", "bid128_scalbn"},
	"Bid128Scalbln":                   {"bid128_misc", "bid128_scalbln"},
	"Bid128SignalingGreater":          {"bid128_compare", "bid128_signaling_greater"},
	"Bid128SignalingGreaterEqual":     {"bid128_compare", "bid128_signaling_greater_equal"},
	"Bid128SignalingGreaterUnordered": {"bid128_compare", "bid128_signaling_greater_unordered"},
	"Bid128SignalingLess":             {"bid128_compare", "bid128_signaling_less"},
	"Bid128SignalingLessEqual":        {"bid128_compare", "bid128_signaling_less_equal"},
	"Bid128SignalingLessUnordered":    {"bid128_compare", "bid128_signaling_less_unordered"},
	"Bid128SignalingNotGreater":       {"bid128_compare", "bid128_signaling_not_greater"},
	"Bid128SignalingNotLess":          {"bid128_compare", "bid128_signaling_not_less"},
	"Bid128Sqrt":                      {"bid128_sqrt", "bid128_sqrt"},
	"Bid128Sub":                       {"bid128_add", "bid128_sub"},
	"Bid128ToBid32":                   {"bid128_conversions", "bid128_to_bid32"},
	"Bid128ToBid64":                   {"bid128_conversions", "bid128_to_bid64"},
	"Bid128ToBinary128":               {"to_binary64", "bid128_to_binary128"},
	"Bid128ToBinary32":                {"bid128_to_binary", "bid128_to_binary32"},
	"Bid128ToBinary64":                {"bid128_to_binary", "bid128_to_binary64"},
	"Bid128ToString":                  {"bid128_string", "bid128_to_string"},
	"Bid128TotalOrder":                {"bid128_compare", "bid128_total_order"},
	"Bid128TotalOrderMag":             {"bid128_compare", "bid128_total_order_mag"},

	// Decimal128 ConvertToInt<N>/ConvertToUint<N>(Exact) mode-dispatch
	// leaves (8 widths x 5 modes x {plain,exact}). Every one of these lands in
	// the single bid128_to_int module (verified against the generated source,
	// mirroring Decimal32's single-module consolidation rather than
	// Decimal64's many-small-files layout); every leaf here is tuple-return
	// (never pfpsf), also verified against the generated source.
	"Bid128ToInt8Rnint":     {"bid128_to_int", "bid128_to_int8_rnint"},
	"Bid128ToInt8Rninta":    {"bid128_to_int", "bid128_to_int8_rninta"},
	"Bid128ToInt8Int":       {"bid128_to_int", "bid128_to_int8_int"},
	"Bid128ToInt8Ceil":      {"bid128_to_int", "bid128_to_int8_ceil"},
	"Bid128ToInt8Floor":     {"bid128_to_int", "bid128_to_int8_floor"},
	"Bid128ToInt8Xrnint":    {"bid128_to_int", "bid128_to_int8_xrnint"},
	"Bid128ToInt8Xrninta":   {"bid128_to_int", "bid128_to_int8_xrninta"},
	"Bid128ToInt8Xint":      {"bid128_to_int", "bid128_to_int8_xint"},
	"Bid128ToInt8Xceil":     {"bid128_to_int", "bid128_to_int8_xceil"},
	"Bid128ToInt8Xfloor":    {"bid128_to_int", "bid128_to_int8_xfloor"},
	"Bid128ToInt16Rnint":    {"bid128_to_int", "bid128_to_int16_rnint"},
	"Bid128ToInt16Rninta":   {"bid128_to_int", "bid128_to_int16_rninta"},
	"Bid128ToInt16Int":      {"bid128_to_int", "bid128_to_int16_int"},
	"Bid128ToInt16Ceil":     {"bid128_to_int", "bid128_to_int16_ceil"},
	"Bid128ToInt16Floor":    {"bid128_to_int", "bid128_to_int16_floor"},
	"Bid128ToInt16Xrnint":   {"bid128_to_int", "bid128_to_int16_xrnint"},
	"Bid128ToInt16Xrninta":  {"bid128_to_int", "bid128_to_int16_xrninta"},
	"Bid128ToInt16Xint":     {"bid128_to_int", "bid128_to_int16_xint"},
	"Bid128ToInt16Xceil":    {"bid128_to_int", "bid128_to_int16_xceil"},
	"Bid128ToInt16Xfloor":   {"bid128_to_int", "bid128_to_int16_xfloor"},
	"Bid128ToInt32Rnint":    {"bid128_to_int", "bid128_to_int32_rnint"},
	"Bid128ToInt32Rninta":   {"bid128_to_int", "bid128_to_int32_rninta"},
	"Bid128ToInt32Int":      {"bid128_to_int", "bid128_to_int32_int"},
	"Bid128ToInt32Ceil":     {"bid128_to_int", "bid128_to_int32_ceil"},
	"Bid128ToInt32Floor":    {"bid128_to_int", "bid128_to_int32_floor"},
	"Bid128ToInt32Xrnint":   {"bid128_to_int", "bid128_to_int32_xrnint"},
	"Bid128ToInt32Xrninta":  {"bid128_to_int", "bid128_to_int32_xrninta"},
	"Bid128ToInt32Xint":     {"bid128_to_int", "bid128_to_int32_xint"},
	"Bid128ToInt32Xceil":    {"bid128_to_int", "bid128_to_int32_xceil"},
	"Bid128ToInt32Xfloor":   {"bid128_to_int", "bid128_to_int32_xfloor"},
	"Bid128ToInt64Rnint":    {"bid128_to_int", "bid128_to_int64_rnint"},
	"Bid128ToInt64Rninta":   {"bid128_to_int", "bid128_to_int64_rninta"},
	"Bid128ToInt64Int":      {"bid128_to_int", "bid128_to_int64_int"},
	"Bid128ToInt64Ceil":     {"bid128_to_int", "bid128_to_int64_ceil"},
	"Bid128ToInt64Floor":    {"bid128_to_int", "bid128_to_int64_floor"},
	"Bid128ToInt64Xrnint":   {"bid128_to_int", "bid128_to_int64_xrnint"},
	"Bid128ToInt64Xrninta":  {"bid128_to_int", "bid128_to_int64_xrninta"},
	"Bid128ToInt64Xint":     {"bid128_to_int", "bid128_to_int64_xint"},
	"Bid128ToInt64Xceil":    {"bid128_to_int", "bid128_to_int64_xceil"},
	"Bid128ToInt64Xfloor":   {"bid128_to_int", "bid128_to_int64_xfloor"},
	"Bid128ToUint8Rnint":    {"bid128_to_int", "bid128_to_uint8_rnint"},
	"Bid128ToUint8Rninta":   {"bid128_to_int", "bid128_to_uint8_rninta"},
	"Bid128ToUint8Int":      {"bid128_to_int", "bid128_to_uint8_int"},
	"Bid128ToUint8Ceil":     {"bid128_to_int", "bid128_to_uint8_ceil"},
	"Bid128ToUint8Floor":    {"bid128_to_int", "bid128_to_uint8_floor"},
	"Bid128ToUint8Xrnint":   {"bid128_to_int", "bid128_to_uint8_xrnint"},
	"Bid128ToUint8Xrninta":  {"bid128_to_int", "bid128_to_uint8_xrninta"},
	"Bid128ToUint8Xint":     {"bid128_to_int", "bid128_to_uint8_xint"},
	"Bid128ToUint8Xceil":    {"bid128_to_int", "bid128_to_uint8_xceil"},
	"Bid128ToUint8Xfloor":   {"bid128_to_int", "bid128_to_uint8_xfloor"},
	"Bid128ToUint16Rnint":   {"bid128_to_int", "bid128_to_uint16_rnint"},
	"Bid128ToUint16Rninta":  {"bid128_to_int", "bid128_to_uint16_rninta"},
	"Bid128ToUint16Int":     {"bid128_to_int", "bid128_to_uint16_int"},
	"Bid128ToUint16Ceil":    {"bid128_to_int", "bid128_to_uint16_ceil"},
	"Bid128ToUint16Floor":   {"bid128_to_int", "bid128_to_uint16_floor"},
	"Bid128ToUint16Xrnint":  {"bid128_to_int", "bid128_to_uint16_xrnint"},
	"Bid128ToUint16Xrninta": {"bid128_to_int", "bid128_to_uint16_xrninta"},
	"Bid128ToUint16Xint":    {"bid128_to_int", "bid128_to_uint16_xint"},
	"Bid128ToUint16Xceil":   {"bid128_to_int", "bid128_to_uint16_xceil"},
	"Bid128ToUint16Xfloor":  {"bid128_to_int", "bid128_to_uint16_xfloor"},
	"Bid128ToUint32Rnint":   {"bid128_to_int", "bid128_to_uint32_rnint"},
	"Bid128ToUint32Rninta":  {"bid128_to_int", "bid128_to_uint32_rninta"},
	"Bid128ToUint32Int":     {"bid128_to_int", "bid128_to_uint32_int"},
	"Bid128ToUint32Ceil":    {"bid128_to_int", "bid128_to_uint32_ceil"},
	"Bid128ToUint32Floor":   {"bid128_to_int", "bid128_to_uint32_floor"},
	"Bid128ToUint32Xrnint":  {"bid128_to_int", "bid128_to_uint32_xrnint"},
	"Bid128ToUint32Xrninta": {"bid128_to_int", "bid128_to_uint32_xrninta"},
	"Bid128ToUint32Xint":    {"bid128_to_int", "bid128_to_uint32_xint"},
	"Bid128ToUint32Xceil":   {"bid128_to_int", "bid128_to_uint32_xceil"},
	"Bid128ToUint32Xfloor":  {"bid128_to_int", "bid128_to_uint32_xfloor"},
	"Bid128ToUint64Rnint":   {"bid128_to_int", "bid128_to_uint64_rnint"},
	"Bid128ToUint64Rninta":  {"bid128_to_int", "bid128_to_uint64_rninta"},
	"Bid128ToUint64Int":     {"bid128_to_int", "bid128_to_uint64_int"},
	"Bid128ToUint64Ceil":    {"bid128_to_int", "bid128_to_uint64_ceil"},
	"Bid128ToUint64Floor":   {"bid128_to_int", "bid128_to_uint64_floor"},
	"Bid128ToUint64Xrnint":  {"bid128_to_int", "bid128_to_uint64_xrnint"},
	"Bid128ToUint64Xrninta": {"bid128_to_int", "bid128_to_uint64_xrninta"},
	"Bid128ToUint64Xint":    {"bid128_to_int", "bid128_to_uint64_xint"},
	"Bid128ToUint64Xceil":   {"bid128_to_int", "bid128_to_uint64_xceil"},
	"Bid128ToUint64Xfloor":  {"bid128_to_int", "bid128_to_uint64_xfloor"},
}

// boolResultPorts is the closed set of bidgo functions whose generated Rust
// translation already returns bool, unlike the Intel-mirroring int-returning
// convention (compared via "!= 0") every other portPath entry assumes.
// Decimal32's bid32_exports.go was authored with native bool returns for
// this specific subset (IsInf/IsNaN/IsZero/SameQuantum); the equivalent
// Decimal64 predicates (noncomp64.go) and every other Decimal32 predicate
// (bid32_noncomp.go) return int (mirrors Intel's C convention) and are not
// listed here. A predicate/same_quantum/sign-shaped op looks a bidgo
// function up here (via decOp.boolPort, set when the op is built) rather
// than guessing bool-vs-int from the function name, per the project's
// no-implicit-inference convention.
var boolResultPorts = map[string]bool{
	"Bid32IsInf":       true,
	"Bid32IsNaN":       true,
	"Bid32IsZero":      true,
	"Bid32SameQuantum": true,
}

// portPfpsf is the closed set of bidgo functions whose generated Rust
// translation takes a trailing `pfpsf: &mut u32` output parameter and returns
// only the value, instead of the (value, flags) tuple every other portPath
// entry uses. This is a
// per-bidgo-function fact read directly off the generated
// bid754-rs/src/generated/bid128_*.rs source, not derived from the apiemit
// shape or width: Decimal128 mixes both calling conventions even within a
// single shape family (e.g. binary_flags_no_round's Bid128Fmod/Bid128Rem are
// tuple-return but Bid128Maxnum/Bid128MaxnumMag/Bid128Minnum/Bid128MinnumMag
// in that same shape are pfpsf; unary_with_flags_no_round's
// Bid128NextUp/Bid128NextDown are tuple-return but Bid128Logb and every
// Bid128RoundIntegral* in that shape are pfpsf, and
// unary_int_with_flags_no_round's Bid128Ilogb is pfpsf while its 32/64
// siblings Bid32ILogb/Bid64ILogb are tuple-return), so this table -- not the
// shape name -- is the single source of truth a decOp consults (via
// decOp.pfpsf, set when the op is built, mirroring boolResultPorts/
// decOp.boolPort immediately above). Every 32/64 bidgo function, and every
// other Bid128* function not listed here (confirmed by reading every
// generated:: signature the Decimal128 expansion manifest rows reference), is
// tuple-return.
var portPfpsf = map[string]bool{
	"Bid128Add":                      true,
	"Bid128Sub":                      true,
	"Bid128Ilogb":                    true,
	"Bid128Logb":                     true,
	"Bid128Scalbn":                   true,
	"Bid128Scalbln":                  true,
	"Bid128Maxnum":                   true,
	"Bid128MaxnumMag":                true,
	"Bid128Minnum":                   true,
	"Bid128MinnumMag":                true,
	"Bid128RoundIntegralExact":       true,
	"Bid128RoundIntegralNearestAway": true,
	"Bid128RoundIntegralNearestEven": true,
	"Bid128RoundIntegralNegative":    true,
	"Bid128RoundIntegralPositive":    true,
	"Bid128RoundIntegralZero":        true,
	"Bid128ToBinary32":               true,
	"Bid128ToBinary64":               true,
}

// PortPfpsf reports whether the generated Rust function a bidgo function name
// resolves to uses the pfpsf-output-parameter calling convention (true) or
// the tuple-return convention every other port function uses (false).
// Exported for the same reason and under the same independence argument as
// PortPathFor/PortBoolResult (independent mapping): this is a
// structural fact about the generated function's Rust signature, not a
// flag/rounding semantic mapping, so the Rust public-API parity generator
// (devtools/internal/testgen) reusing it does not relax the
// independent-mapping guarantee.
func PortPfpsf(bidgoFunction string) bool {
	return portPfpsf[bidgoFunction]
}

// widthSpec is the width-parameterization record: everything a
// Decimal<w> shape template (decimal_emit.go) or the shared collector
// (buildDecimalRs) needs to stay a single mechanical template reused across
// widths instead of a second hand-written implementation per width. Every
// field is an explicit literal on decimal32Width/decimal64Width/
// decimal128Width below (never derived from another field by string-splicing
// at a call site), per the project's no-implicit-fallback convention --
// except the small pure derivations exposed as methods (bidgoPrefix,
// castFromU64, castToU64, selfArg, wrapResult), which only ever recombine
// already-explicit fields.
//
// Decimal128's 16-byte layout does not fit the single bitsType primitive
// 32/64 use (bitsType/byteSize/nanBits/.../canonicalLimitLabel are therefore
// unused -- left zero-valued -- on decimal128Width; buildDecimalRs's header
// emission and the two NaN-literal helpers branch on is128 instead of
// reading those fields for width 128, see buildDecimal128Header and
// emitParseDecimalNaN128Fn/emitFormatDecimalNaN128Fn). is128 is the Decimal128
// discriminant every shape template's port-call/value-conversion codegen
// discriminator: selfArg/wrapResult are the single generation
// point for the [u8;16]<->BID_UINT128 conversion (reusing the existing
// super::types::bid_uint128_from_le_bytes/to_le_bytes helpers), and
// decOp.pfpsf (populated from portPfpsf, not from is128 -- Decimal128 mixes
// both port calling conventions per-function) is the single generation point
// for the pfpsf-output-parameter fork (flagsCallStmt).
type widthSpec struct {
	selfType string // "Decimal32" | "Decimal64" | "Decimal128" -- Rust value-type name = manifest rust_owner
	digits   string // "32" | "64" | "128" -- for doc-comment/Go-symbol-name references
	bitsType string // "u32" | "u64" -- the wrapped primitive; unused when is128
	byteSize string // "4" | "8" | "16" -- core::mem::size_of::<Self>() assertion
	is128    bool   // true only for decimal128Width -- see the widthSpec doc comment

	// Inclusive cohort-exponent (quantum) range encoded by this BID width.
	// Precision is the maximum decimal coefficient digit count. These are
	// explicit IEEE format constants used only to detect a silent from-string
	// cohort coercion when the mechanical port reports no status flags.
	minQuantum string
	maxQuantum string
	precision  string

	// NaN-literal encoding, mirrored byte-for-byte from the Go public-API
	// glue (bid754-go/types_bid_nan_payload.go: parseDecimal<w>BIDNaN /
	// formatDecimal<w>BIDNaN), not re-derived or approximated. Unused when
	// is128 (Decimal128's NaN payload does not fit a u64 -- see
	// emitParseDecimalNaN128Fn/emitFormatDecimalNaN128Fn).
	nanBits              string // Rust hex literal, quiet-NaN combination-field pattern
	snanBits             string // ditto, signaling-NaN combination-field pattern
	signBit              string // ditto, sign bit
	payloadMask          string // ditto, NaN payload extraction mask (format direction)
	parseMaxPayload      string // Rust literal (typed), inclusive max parseable payload
	formatCanonicalLimit string // Rust literal (typed), exclusive canonical-payload display limit
	canonicalLimitLabel  string // human doc-comment label for formatCanonicalLimit, e.g. "10^6"
}

// bidgoPrefix returns this width's bidgo function family prefix ("Bid32" /
// "Bid64" / "Bid128"), used to build the exact census bidgo_function name for
// a shape whose auxiliary port lookup (e.g. sign's IsZero, signaling_eq's
// LessEqual) is not itself a manifest row's own BidgoFunction.
func (w widthSpec) bidgoPrefix() string { return "Bid" + w.digits }

// castFromU64 returns the Rust cast suffix needed to narrow a shared u64
// value (the types.rs parse_uint_payload helper's result) down to this
// width's bitsType, or "" when bitsType is already u64 (Decimal64). Not
// called for width 128 (its NaN payload uses parse_u128_payload directly,
// never parse_uint_payload -- see emitParseDecimalNaN128Fn).
func (w widthSpec) castFromU64() string {
	if w.bitsType == "u64" {
		return ""
	}
	return " as " + w.bitsType
}

// castToU64 returns the Rust cast suffix needed to widen a bitsType value up
// to u64 (the shared types.rs payload_string helper operates on u64
// regardless of decimal width), or "" when bitsType is already u64. Not
// called for width 128 (see castFromU64).
func (w widthSpec) castToU64() string {
	if w.bitsType == "u64" {
		return ""
	}
	return " as u64"
}

// selfArg returns the Rust expression converting an already-bound Decimal<w>
// identifier (e.g. "self", "rhs", "mul", "sign_source") into the raw argument
// form a crate::generated::* port call expects: the wrapped primitive
// directly (name.0) for width 32/64, or the [u8;16]->BID_UINT128 conversion
// for width 128 -- the single generation point for that conversion, reusing
// the existing super::types::bid_uint128_from_le_bytes
// helper rather than duplicating the byte-splitting logic per call site.
// name.0 is directly accessible for every width here because every emitter
// in this file runs inside that width's own module (decimal128.rs accesses
// its own private tuple field the same way decimal64.rs/decimal32.rs do);
// this is NOT the same call as NextToward's `target.to_le_bytes()` (a
// *different* module's public accessor for a foreign-typed parameter that is
// always Decimal128 regardless of the receiver width w).
func (w widthSpec) selfArg(name string) string {
	if w.is128 {
		return "super::types::bid_uint128_from_le_bytes(" + name + ".0)"
	}
	return name + ".0"
}

// wrapResult returns the Rust expression constructing a Self value from a
// port call's raw result expression: the tuple constructor directly for
// width 32/64, or the BID_UINT128->[u8;16] conversion (same single
// generation point as selfArg, reusing bid_uint128_to_le_bytes) for width
// 128.
func (w widthSpec) wrapResult(rawExpr string) string {
	if w.is128 {
		return w.selfType + "(super::types::bid_uint128_to_le_bytes(" + rawExpr + "))"
	}
	return w.selfType + "(" + rawExpr + ")"
}

// canonicalQNaNResult returns the per-width private tuple construction used
// when a public raw parser must reject an unrepresentable NaN payload through
// its exception-flag channel. The bit patterns are the same canonical quiet
// NaNs used by the Go public wrappers.
func (w widthSpec) canonicalQNaNResult() string {
	if w.is128 {
		return w.wrapResult("crate::gen_types::BID_UINT128 { lo: 0, hi: " + decimal128QuietNaNHighBits + " }")
	}
	return w.wrapResult(w.nanBits)
}

// crossModuleArg is selfArg's counterpart for a caller OUTSIDE width w's own
// module (context.rs, the only such caller: it takes Decimal<w> values as
// plain parameters, not a receiver, so it cannot reach the private tuple
// field selfArg uses). It converts via the PUBLIC to_bits()/to_le_bytes()
// accessor instead: name.to_bits() for width 32/64, or the BID_UINT128
// conversion over the public to_le_bytes() for width 128 -- the same pattern
// emitNextTowardOp already established for its cross-module Decimal128
// `target` parameter, generalized here to every width via widthSpec instead
// of hardcoded to Decimal128.
func (w widthSpec) crossModuleArg(name string) string {
	if w.is128 {
		return "super::types::bid_uint128_from_le_bytes(" + name + ".to_le_bytes())"
	}
	return name + ".to_bits()"
}

// crossModuleWrap is wrapResult's counterpart for a caller outside width w's
// own module: it constructs Self via the PUBLIC from_bits()/from_le_bytes()
// constructor instead of the private tuple-struct literal.
func (w widthSpec) crossModuleWrap(rawExpr string) string {
	if w.is128 {
		return w.selfType + "::from_le_bytes(super::types::bid_uint128_to_le_bytes(" + rawExpr + "))"
	}
	return w.selfType + "::from_bits(" + rawExpr + ")"
}

var decimal64Width = widthSpec{
	selfType: "Decimal64", digits: "64", bitsType: "u64", byteSize: "8",
	minQuantum: "-398", maxQuantum: "369", precision: "16",
	nanBits: "0x7c00_0000_0000_0000", snanBits: "0x7e00_0000_0000_0000", signBit: "0x8000_0000_0000_0000",
	payloadMask:          "0x0003_ffff_ffff_ffff",
	parseMaxPayload:      "999_999_999_999_999u64",
	formatCanonicalLimit: "1_000_000_000_000_000u64",
	canonicalLimitLabel:  "10^15",
}

var decimal32Width = widthSpec{
	selfType: "Decimal32", digits: "32", bitsType: "u32", byteSize: "4",
	minQuantum: "-101", maxQuantum: "90", precision: "7",
	nanBits: "0x7c00_0000", snanBits: "0x7e00_0000", signBit: "0x8000_0000",
	payloadMask:          "0x000f_ffff",
	parseMaxPayload:      "999_999u64",
	formatCanonicalLimit: "1_000_000u64",
	canonicalLimitLabel:  "10^6",
}

// decimal128Width is the Decimal128 expansion width record. bitsType/nanBits/snanBits/
// signBit/payloadMask/parseMaxPayload/formatCanonicalLimit/
// canonicalLimitLabel are intentionally left zero-valued: is128 routes every
// call site that would read them to a width-128-specific path instead (the
// widthSpec doc comment explains why: no single bitsType primitive and no
// u64-fitting NaN payload).
var decimal128Width = widthSpec{
	selfType: "Decimal128", digits: "128", byteSize: "16", is128: true,
	minQuantum: "-6176", maxQuantum: "6111", precision: "34",
}

// Decimal128 NaN-literal payload constants (emitParseDecimalNaN128Fn/
// emitFormatDecimalNaN128Fn): mirrors the Go decimal128NaNPayloadLimit()
// (10^33, computed there via big.Int.Exp) and its implicit inclusive parse
// max (limit-1), verified against the Go source arithmetically.
// decimal128NaNPayloadMaskHi is the high-word payload
// extraction mask (mirrors the Go formatDecimal128BIDNaN's
// `bits.hi & 0x00003fffffffffff`): together with the full 64-bit low word
// this gives the ~110-bit total payload capacity.
const (
	decimal128QuietNaNHighBits     = "0x7c00_0000_0000_0000"
	decimal128SignalingNaNHighBits = "0x7e00_0000_0000_0000"
	decimal128SignHighBit          = "0x8000_0000_0000_0000"
	decimal128NaNMaxPayload        = "999_999_999_999_999_999_999_999_999_999_999u128"
	decimal128NaNCanonicalLimit    = "1_000_000_000_000_000_000_000_000_000_000_000u128"
	decimal128NaNPayloadMaskHi     = "0x0000_3fff_ffff_ffff"
)

// tmpl performs @TOKEN@ substitution on a Rust source template: safer than
// fmt.Sprintf for templates with several repeated placeholders (a positional
// %s repetition miscount silently produces a "%!(EXTRA ...)" compile
// failure instead of a clear generation-time error -- a lesson from the parity check
// Rust parity generator, which hit exactly this and switched to the same
// token-substitution style; see rust_public_parity_emit.go's rustTmpl). vals
// must have an even length: token, value, token, value, .... Every token
// must occur at least once in src and every remaining "@...@" after
// substitution is treated as a leftover mistake, so both directions fail
// loudly (a generator-authoring bug, not a runtime input error).
func tmpl(src string, vals ...string) string {
	if len(vals)%2 != 0 {
		panic("apiemit: tmpl requires an even number of token/value arguments")
	}
	out := src
	for i := 0; i < len(vals); i += 2 {
		token := "@" + vals[i] + "@"
		if !strings.Contains(out, token) {
			panic("apiemit: tmpl token " + token + " not present in template")
		}
		out = strings.ReplaceAll(out, token, vals[i+1])
	}
	if strings.Contains(out, "@") {
		panic("apiemit: tmpl left an unsubstituted @...@ token in the template")
	}
	return out
}

// PortPathFor returns the crate::generated module and function name apiemit
// resolves a bidgo function name to when it builds a public-API wrapper call
// site. It is exported for devtools/internal/testgen's Rust public-API parity
// generator (independent mapping), which must call the same
// port entrypoint the wrapper itself calls -- independently of the wrapper's
// own flag/rounding conversion -- to make a wrapper-vs-port divergence
// visible. Reusing this table (rather than a second hand-copied map) keeps
// the module/function resolution itself as a single source of truth: only the
// flag-bit and rounding-mode value mappings are required to be independent
// literals (see mapPortFlagsForParity in the Go parity generator for the
// precedent), not the port address lookup.
func PortPathFor(bidgoFunction string) (module, fn string, ok bool) {
	p, ok := portPath[bidgoFunction]
	if !ok {
		return "", "", false
	}
	return p.module, p.fn, true
}

// PortBoolResult reports whether the generated Rust function a bidgo
// function name resolves to already returns bool, rather than the
// Intel-mirroring numeric type (compared via "!= 0") every other port
// function uses. Exported for the same reason and under the same
// independence argument as PortPathFor (independent mapping):
// this is a structural fact about the generated function's Rust return type
// -- discoverable by reading its signature, true regardless of how any
// caller interprets it -- not a flag/rounding *semantic* mapping, so a
// parity generator reusing it does not relax the independent-mapping
// guarantee (see boolResultPorts's doc comment for which bidgo functions
// this is true for and why).
func PortBoolResult(bidgoFunction string) bool {
	return boolResultPorts[bidgoFunction]
}

const crlf = "\n"

// emitRustAPI cleans and regenerates the src/generated/api subtree from the
// manifest emit rules. It writes the type layer plus, for every owner that has
// emit rules, the wrapper file. Every file carries the go2rs generated marker.
func emitRustAPI(apiDir string, manifest *manifestFile) error {
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		return fmt.Errorf("apiemit: mkdir %q: %w", apiDir, err)
	}
	// Self-subtree clean, mirroring go2rs cleanGeneratedDir but scoped to api/:
	// remove only the .rs files this emitter owns, never touching the parent
	// generated/*.rs implementation set.
	if err := cleanAPIDir(apiDir); err != nil {
		return err
	}

	files := map[string]string{
		"types.rs": buildTypesRs(),
	}

	d64, err := buildDecimal64Rs(manifest)
	if err != nil {
		return err
	}
	files["decimal64.rs"] = d64

	d32, err := buildDecimal32Rs(manifest)
	if err != nil {
		return err
	}
	files["decimal32.rs"] = d32

	d128, err := buildDecimal128Rs(manifest)
	if err != nil {
		return err
	}
	files["decimal128.rs"] = d128

	ctx, err := buildContextRs(manifest)
	if err != nil {
		return err
	}
	files["context.rs"] = ctx

	// Reject any owner in the manifest that has no wrapper file target yet:
	// Decimal64, Decimal32, Decimal128, and Context all carry emitted
	// wrappers. Any additional owner needs an explicit file builder.
	for _, r := range manifest.Emit {
		if r.Shape == "parse_fold" || r.Shape == "copy_fold" {
			continue
		}
		if r.RustOwner != "Decimal64" && r.RustOwner != "Decimal32" && r.RustOwner != "Decimal128" && r.RustOwner != "Context" {
			return fmt.Errorf("apiemit: emit rule for go_symbol %q targets rust_owner %q which has no wrapper file builder yet (Decimal64, Decimal32, Decimal128, and Context are the only owners with a file builder today)", r.GoSymbol, r.RustOwner)
		}
	}

	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	// Check generated API files before any byte is written if any
	// api/*.rs file (including mod.rs) reaches outside the expected
	// crate::* implementation-domain modules or contains a rejected
	// unsafe/panic/unwrap/expect token.
	scanTargets := make(map[string]string, len(files)+1)
	for name, content := range files {
		scanTargets[name] = content
	}
	scanTargets["mod.rs"] = buildModRs(names)
	if err := staticCheckAPIOutput(scanTargets); err != nil {
		return err
	}

	for _, name := range names {
		path := filepath.Join(apiDir, name)
		if err := os.WriteFile(path, []byte(files[name]), 0o644); err != nil {
			return fmt.Errorf("apiemit: write %q: %w", path, err)
		}
	}
	if err := os.WriteFile(filepath.Join(apiDir, "mod.rs"), []byte(scanTargets["mod.rs"]), 0o644); err != nil {
		return fmt.Errorf("apiemit: write api mod.rs: %w", err)
	}
	return nil
}

// cleanAPIDir removes the emitter-owned .rs files (including mod.rs) at the top
// of the api directory, mirroring go2rs cleanGeneratedDir. It intentionally
// touches only the api subtree and never the parent generated/ files.
func cleanAPIDir(apiDir string) error {
	entries, err := os.ReadDir(apiDir)
	if err != nil {
		return fmt.Errorf("apiemit: read api dir %q: %w", apiDir, err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".rs") {
			continue
		}
		if err := os.Remove(filepath.Join(apiDir, e.Name())); err != nil {
			return fmt.Errorf("apiemit: remove stale api artifact %q: %w", e.Name(), err)
		}
	}
	return nil
}

func marker() string { return genmarker.Line(MarkerTool) }

func buildModRs(fileNames []string) string {
	var b strings.Builder
	b.WriteString(marker() + crlf)
	b.WriteString(crlf)
	b.WriteString("//! Public Rust API surface for bid754, generated by devtools/tools/go2rs/apiemit." + crlf)
	b.WriteString("//!" + crlf)
	b.WriteString("//! Wrapper bodies only call into crate::generated::* port functions, convert" + crlf)
	b.WriteString("//! value types, and map exception flags; no arithmetic is reproduced here." + crlf)
	b.WriteString(crlf)
	// Structurally enforce the public-surface guarantees (docs/SPEC.md: no
	// panic/trap, no unsafe on a public path) and restore the overflow lints the
	// implementation modules allow crate-wide, so a real overflow in this
	// hand-shaped surface is not silenced.
	b.WriteString("#![forbid(unsafe_code)]" + crlf)
	b.WriteString("#![deny(arithmetic_overflow, overflowing_literals)]" + crlf)
	b.WriteString(crlf)
	for _, name := range fileNames {
		mod := strings.TrimSuffix(name, ".rs")
		b.WriteString("mod " + mod + ";" + crlf)
	}
	b.WriteString(crlf)
	b.WriteString("pub use types::{Binary128, DecimalClass, ExceptionFlags, InexactIntegerError, InvalidRoundingMode, ParseDecimalError, RoundingMode};" + crlf)
	b.WriteString("pub use decimal32::Decimal32;" + crlf)
	b.WriteString("pub use decimal64::Decimal64;" + crlf)
	b.WriteString("pub use decimal128::Decimal128;" + crlf)
	b.WriteString("pub use context::Context;" + crlf)
	return b.String()
}

func buildTypesRs() string {
	var b strings.Builder
	b.WriteString(marker() + crlf)
	b.WriteString(`//! Shared public value types for the bid754 Rust API.

use core::fmt;

/// IEEE 754 rounding-direction attribute: the closed set of the five IEEE
/// rounding modes. The non-IEEE decTest-compatibility mode used by internal
/// verification plumbing is intentionally not representable here.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
#[repr(u32)]
pub enum RoundingMode {
    /// roundTiesToEven (IEEE 754 default).
    NearestEven = 0,
    /// roundTiesToAway.
    NearestAway = 1,
    /// roundTowardZero.
    TowardZero = 2,
    /// roundTowardPositive.
    TowardPositive = 3,
    /// roundTowardNegative.
    TowardNegative = 4,
}

impl core::convert::TryFrom<u32> for RoundingMode {
    type Error = InvalidRoundingMode;

    fn try_from(value: u32) -> Result<RoundingMode, InvalidRoundingMode> {
        match value {
            0 => Ok(RoundingMode::NearestEven),
            1 => Ok(RoundingMode::NearestAway),
            2 => Ok(RoundingMode::TowardZero),
            3 => Ok(RoundingMode::TowardPositive),
            4 => Ok(RoundingMode::TowardNegative),
            other => Err(InvalidRoundingMode(other)),
        }
    }
}

/// Error returned by ` + "`RoundingMode::try_from`" + ` for a value outside the five modes.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct InvalidRoundingMode(pub u32);

impl fmt::Display for InvalidRoundingMode {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "invalid rounding mode: {}", self.0)
    }
}

impl std::error::Error for InvalidRoundingMode {}

/// Error returned when an integer cannot be represented exactly in the
/// requested fixed-width decimal type.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct InexactIntegerError {
    input: String,
    target: &'static str,
}

impl InexactIntegerError {
    pub(crate) fn new(input: String, target: &'static str) -> InexactIntegerError {
        InexactIntegerError { input, target }
    }

    /// Returns the integer spelling that could not be represented exactly.
    pub fn input(&self) -> &str {
        &self.input
    }

    /// Returns the target decimal type name.
    pub fn target(&self) -> &'static str {
        self.target
    }
}

impl fmt::Display for InexactIntegerError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "integer {} is not exactly representable as {}", self.input, self.target)
    }
}

impl std::error::Error for InexactIntegerError {}

/// IEEE 754 exception status flags raised by an operation, as a bit set. The
/// eight constants match the Go public ExceptionFlags vocabulary.
#[derive(Clone, Copy, Debug, Default, PartialEq, Eq, Hash)]
pub struct ExceptionFlags(u32);

impl ExceptionFlags {
    /// Inexact result.
    pub const INEXACT: ExceptionFlags = ExceptionFlags(1 << 0);
    /// Underflow.
    pub const UNDERFLOW: ExceptionFlags = ExceptionFlags(1 << 1);
    /// Overflow.
    pub const OVERFLOW: ExceptionFlags = ExceptionFlags(1 << 2);
    /// Division by zero.
    pub const DIVISION_BY_ZERO: ExceptionFlags = ExceptionFlags(1 << 3);
    /// Invalid operation.
    pub const INVALID_OPERATION: ExceptionFlags = ExceptionFlags(1 << 4);
    /// Subnormal result.
    pub const SUBNORMAL: ExceptionFlags = ExceptionFlags(1 << 5);
    /// Rounded result.
    pub const ROUNDED: ExceptionFlags = ExceptionFlags(1 << 6);
    /// Exponent clamped.
    pub const CLAMPED: ExceptionFlags = ExceptionFlags(1 << 7);

    /// The empty flag set.
    pub const fn empty() -> ExceptionFlags {
        ExceptionFlags(0)
    }

    /// Returns the raw bit representation.
    pub const fn bits(self) -> u32 {
        self.0
    }

    /// Reports whether no flags are set.
    pub const fn is_empty(self) -> bool {
        self.0 == 0
    }

    /// Reports whether every flag in ` + "`other`" + ` is set in ` + "`self`" + `.
    pub const fn contains(self, other: ExceptionFlags) -> bool {
        (self.0 & other.0) == other.0
    }

    /// Returns ` + "`self`" + ` with every flag in ` + "`other`" + ` cleared.
    pub const fn difference(self, other: ExceptionFlags) -> ExceptionFlags {
        ExceptionFlags(self.0 & !other.0)
    }

    /// Maps a bidgo/Intel status-flag word returned by a port function into the
    /// public flag set. Mirrors the Go bidgoExceptionFlags mapping; the bidgo
    /// bit values are the fixed Intel status-flag constants (INVALID=1,
    /// ZERO_DIVIDE=4, OVERFLOW=8, UNDERFLOW=16, INEXACT=32).
    pub(crate) fn from_bidgo(raw: u32) -> ExceptionFlags {
        const BID_INVALID: u32 = 1;
        const BID_ZERO_DIVIDE: u32 = 4;
        const BID_OVERFLOW: u32 = 8;
        const BID_UNDERFLOW: u32 = 16;
        const BID_INEXACT: u32 = 32;
        let mut out = ExceptionFlags::empty();
        if raw & BID_INEXACT != 0 {
            out |= ExceptionFlags::INEXACT;
        }
        if raw & BID_UNDERFLOW != 0 {
            out |= ExceptionFlags::UNDERFLOW;
        }
        if raw & BID_OVERFLOW != 0 {
            out |= ExceptionFlags::OVERFLOW;
        }
        if raw & BID_ZERO_DIVIDE != 0 {
            out |= ExceptionFlags::DIVISION_BY_ZERO;
        }
        if raw & BID_INVALID != 0 {
            out |= ExceptionFlags::INVALID_OPERATION;
        }
        out
    }
}

impl core::ops::BitOr for ExceptionFlags {
    type Output = ExceptionFlags;
    fn bitor(self, rhs: ExceptionFlags) -> ExceptionFlags {
        ExceptionFlags(self.0 | rhs.0)
    }
}

impl core::ops::BitOrAssign for ExceptionFlags {
    fn bitor_assign(&mut self, rhs: ExceptionFlags) {
        self.0 |= rhs.0;
    }
}

impl core::ops::BitAnd for ExceptionFlags {
    type Output = ExceptionFlags;
    fn bitand(self, rhs: ExceptionFlags) -> ExceptionFlags {
        ExceptionFlags(self.0 & rhs.0)
    }
}

impl fmt::Display for ExceptionFlags {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        if self.0 == 0 {
            return f.write_str("None");
        }
        let mut first = true;
        for (flag, name) in [
            (ExceptionFlags::INEXACT, "Inexact"),
            (ExceptionFlags::UNDERFLOW, "Underflow"),
            (ExceptionFlags::OVERFLOW, "Overflow"),
            (ExceptionFlags::DIVISION_BY_ZERO, "DivisionByZero"),
            (ExceptionFlags::INVALID_OPERATION, "InvalidOperation"),
            (ExceptionFlags::SUBNORMAL, "Subnormal"),
            (ExceptionFlags::ROUNDED, "Rounded"),
            (ExceptionFlags::CLAMPED, "Clamped"),
        ] {
            if self.contains(flag) {
                if !first {
                    f.write_str("|")?;
                }
                f.write_str(name)?;
                first = false;
            }
        }
        Ok(())
    }
}

/// IEEE 754-2019 class(x) result (clause 5.7.2), using the GDA decTest class
/// spellings via ` + "`Display`" + `.
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
pub enum DecimalClass {
    /// Signaling NaN.
    SignalingNaN,
    /// Quiet NaN.
    QuietNaN,
    /// Negative infinity.
    NegativeInfinity,
    /// Negative normal.
    NegativeNormal,
    /// Negative subnormal.
    NegativeSubnormal,
    /// Negative zero.
    NegativeZero,
    /// Positive zero.
    PositiveZero,
    /// Positive subnormal.
    PositiveSubnormal,
    /// Positive normal.
    PositiveNormal,
    /// Positive infinity.
    PositiveInfinity,
}

impl fmt::Display for DecimalClass {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            DecimalClass::SignalingNaN => "sNaN",
            DecimalClass::QuietNaN => "NaN",
            DecimalClass::NegativeInfinity => "-Infinity",
            DecimalClass::NegativeNormal => "-Normal",
            DecimalClass::NegativeSubnormal => "-Subnormal",
            DecimalClass::NegativeZero => "-Zero",
            DecimalClass::PositiveZero => "+Zero",
            DecimalClass::PositiveSubnormal => "+Subnormal",
            DecimalClass::PositiveNormal => "+Normal",
            DecimalClass::PositiveInfinity => "+Infinity",
        };
        f.write_str(s)
    }
}

/// Error returned when a string cannot be parsed as a decimal value.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ParseDecimalError {
    input: String,
}

impl ParseDecimalError {
    pub(crate) fn new(input: &str) -> ParseDecimalError {
        ParseDecimalError {
            input: input.to_string(),
        }
    }

    /// Returns the input string that failed to parse.
    pub fn input(&self) -> &str {
        &self.input
    }
}

impl fmt::Display for ParseDecimalError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "invalid decimal string: {}", self.input)
    }
}

impl std::error::Error for ParseDecimalError {}

/// A fixed-width IEEE 754 binary128 bit pattern (16 little-endian bytes).
#[derive(Clone, Copy, Debug, PartialEq, Eq, Hash)]
#[repr(transparent)]
pub struct Binary128([u8; 16]);

impl Binary128 {
    /// Reinterprets 16 little-endian bytes as a binary128 pattern.
    pub const fn from_le_bytes(bytes: [u8; 16]) -> Binary128 {
        Binary128(bytes)
    }

    /// Returns the binary128 bit pattern as 16 little-endian bytes.
    pub const fn to_le_bytes(self) -> [u8; 16] {
        self.0
    }
}

const _: () = assert!(core::mem::size_of::<Binary128>() == 16);

/// Maps a closed public ` + "`RoundingMode`" + ` to its bidgo-domain integer value
/// (mirrors the Go ` + "`bidgoRoundingMode`" + ` table). Unlike Go's open ` + "`int`" + `
/// RoundingMode, the public enum here cannot hold an invalid discriminant, so
/// there is no "invalid mode" branch: the match is exhaustive over the five
/// IEEE rounding directions (docs/IEEE754_SPEC.md).
pub(crate) fn to_bidgo_rounding(mode: RoundingMode) -> i64 {
    match mode {
        RoundingMode::NearestEven => 0,
        RoundingMode::TowardNegative => 1,
        RoundingMode::TowardPositive => 2,
        RoundingMode::TowardZero => 3,
        RoundingMode::NearestAway => 4,
    }
}

/// Maps a raw bidgo/Intel class(x) result (0-9) to the public ` + "`DecimalClass`" + `
/// (mirrors the Go ` + "`decimalClassFromBIDClass`" + ` switch; an out-of-range value,
/// which the port never produces, falls back to QuietNaN like the Go default
/// case).
pub(crate) fn decimal_class_from_bid_class(class: i64) -> DecimalClass {
    match class {
        0 => DecimalClass::SignalingNaN,
        1 => DecimalClass::QuietNaN,
        2 => DecimalClass::NegativeInfinity,
        3 => DecimalClass::NegativeNormal,
        4 => DecimalClass::NegativeSubnormal,
        5 => DecimalClass::NegativeZero,
        6 => DecimalClass::PositiveZero,
        7 => DecimalClass::PositiveSubnormal,
        8 => DecimalClass::PositiveNormal,
        9 => DecimalClass::PositiveInfinity,
        _ => DecimalClass::QuietNaN,
    }
}

/// Converts 16 little-endian bytes to the generated-domain BID_UINT128
/// representation (low word first), matching every guaranteed target
/// platform's endianness (docs/PLATFORM_SPEC.md). Built from plain
/// byte-slice reconstruction and safe conversions only, not a pointer cast
/// (contrast with the Go port's decimal128BIDAsBidgo, which reinterprets the
/// bytes directly through Go's low-level pointer-cast facility).
pub(crate) fn bid_uint128_from_le_bytes(bytes: [u8; 16]) -> crate::gen_types::BID_UINT128 {
    let mut lo = [0u8; 8];
    let mut hi = [0u8; 8];
    lo.copy_from_slice(&bytes[0..8]);
    hi.copy_from_slice(&bytes[8..16]);
    crate::gen_types::BID_UINT128 {
        lo: u64::from_le_bytes(lo),
        hi: u64::from_le_bytes(hi),
    }
}

/// Converts the generated-domain BID_UINT128 representation back to 16
/// little-endian bytes, via safe conversions only (contrast with the Go
/// port's decimal128BIDFromBidgo, which reinterprets the bytes directly
/// through Go's low-level pointer-cast facility).
pub(crate) fn bid_uint128_to_le_bytes(v: crate::gen_types::BID_UINT128) -> [u8; 16] {
    let mut bytes = [0u8; 16];
    bytes[0..8].copy_from_slice(&v.lo.to_le_bytes());
    bytes[8..16].copy_from_slice(&v.hi.to_le_bytes());
    bytes
}

/// Composes the IEEE totalOrder(Mag) comparison result from two directional
/// port calls (mirrors the Go totalOrderComparison helper): the port only
/// exposes "x <= y", so both directions are queried and combined into an
/// Ordering.
pub(crate) fn total_order_comparison(left_le: i64, right_le: i64) -> core::cmp::Ordering {
    if left_le != 0 && right_le != 0 {
        core::cmp::Ordering::Equal
    } else if left_le != 0 {
        core::cmp::Ordering::Less
    } else {
        core::cmp::Ordering::Greater
    }
}

/// One parsed public NaN string literal: sign, signaling flag, and the
/// decimal payload digits (mirrors the Go bidNaNLiteral struct).
/// Width-independent; the width-specific bit layout is applied by each
/// Decimal<w> module.
pub(crate) struct BidNaNLiteral {
    pub(crate) negative: bool,
    pub(crate) signaling: bool,
    pub(crate) payload: String,
}

/// Recognizes the public NaN literal grammar (mirrors the Go
/// parseBIDNaNLiteral): ` + "`[+|-] (nan | qnan | snan) [decimal digits]`" + `,
/// case-insensitive, after skipping leading ASCII space/tab only. Trailing,
/// internal, newline, and Unicode whitespace are rejected. The payload's
/// leading zeros are stripped; decTest-style quoting is not part of this
/// grammar.
pub(crate) fn parse_bid_nan_literal(input: &str) -> Option<BidNaNLiteral> {
    let trimmed = input.trim_start_matches(|c| c == ' ' || c == '\t');
    if trimmed.is_empty() {
        return None;
    }
    let mut negative = false;
    let rest = if let Some(r) = trimmed.strip_prefix('+') {
        r
    } else if let Some(r) = trimmed.strip_prefix('-') {
        negative = true;
        r
    } else {
        trimmed
    };
    if rest.is_empty() {
        return None;
    }
    let lower = rest.to_ascii_lowercase();
    let (signaling, payload) = if lower.starts_with("snan") {
        (true, &rest[4..])
    } else if lower.starts_with("qnan") {
        (false, &rest[4..])
    } else if lower.starts_with("nan") {
        (false, &rest[3..])
    } else {
        return None;
    };
    if !payload.is_empty() && !payload.bytes().all(|b| b.is_ascii_digit()) {
        return None;
    }
    let payload = payload.trim_start_matches('0').to_string();
    Some(BidNaNLiteral { negative, signaling, payload })
}

/// Parses a NaN literal's decimal payload digits (already validated as
/// digits-only by parse_bid_nan_literal), rejecting a value above max
/// (mirrors the Go parseUintPayload; an empty payload is canonical zero).
pub(crate) fn parse_uint_payload(payload: &str, max: u64) -> Option<u64> {
    if payload.is_empty() {
        return Some(0);
    }
    let value: u64 = payload.parse().ok()?;
    if value > max {
        return None;
    }
    Some(value)
}

/// Renders a NaN payload only when canonical: nonzero and strictly below
/// canonical_limit (mirrors the Go payloadString; a noncanonical payload is
/// suppressed to empty so String()->parse stays a round trip).
pub(crate) fn payload_string(payload: u64, canonical_limit: u64) -> String {
    if payload == 0 || payload >= canonical_limit {
        String::new()
    } else {
        payload.to_string()
    }
}

/// Decimal128's u128 sibling of parse_uint_payload: BID128's NaN payload
/// (up to 10^33-1, ~110 bits per the pinned Intel encoding) does not fit a
/// u64, unlike Decimal64/Decimal32's payloads, so width 128 uses this
/// instead (mirrors the Go parseDecimal128BIDNaN's math/big.Int payload
/// parse, but native u128 arithmetic suffices since the payload is always
/// well under u128::MAX -- no bignum type or dependency needed).
pub(crate) fn parse_u128_payload(payload: &str, max: u128) -> Option<u128> {
    if payload.is_empty() {
        return Some(0);
    }
    let value: u128 = payload.parse().ok()?;
    if value > max {
        return None;
    }
    Some(value)
}

/// Decimal128's u128 sibling of payload_string; see parse_u128_payload.
pub(crate) fn payload_string_u128(payload: u128, canonical_limit: u128) -> String {
    if payload == 0 || payload >= canonical_limit {
        String::new()
    } else {
        payload.to_string()
    }
}

/// Formats a NaN display string from its sign/signaling/payload components
/// (mirrors the Go formatBIDNaN): ` + "`[+|-] (NaN|SNaN) [payload]`" + `.
pub(crate) fn format_bid_nan(negative: bool, signaling: bool, payload: &str) -> String {
    let mut s = String::new();
    s.push(if negative { '-' } else { '+' });
    s.push_str(if signaling { "SNaN" } else { "NaN" });
    s.push_str(payload);
    s
}
`)
	return b.String()
}

// buildDecimal32Rs and buildDecimal64Rs share their entire wrapper-collection
// and text-assembly logic via buildDecimalRs (defined below, next to
// buildDecimal64Rs for easy comparison); only the widthSpec parameter
// differs.
func buildDecimal32Rs(manifest *manifestFile) (string, error) {
	return buildDecimalRs(manifest, decimal32Width)
}

// buildDecimal128Rs shares buildDecimalRs's entire wrapper-collection and
// text-assembly logic with buildDecimal64Rs/buildDecimal32Rs; only the
// widthSpec parameter (is128=true)
// differs. The struct header (from_le_bytes/to_le_bytes over [u8; 16]
// instead of from_bits/to_bits over a primitive) and the two NaN-literal
// helpers are the only genuinely width-128-specific branches inside
// buildDecimalRs; every shape emitter is otherwise the same mechanical
// template reused across all three widths via widthSpec.selfArg/wrapResult
// and decOp.pfpsf.
func buildDecimal128Rs(manifest *manifestFile) (string, error) {
	return buildDecimalRs(manifest, decimal128Width)
}

// buildContextRs emits the Context type plus every context-arithmetic
// wrapper selected by the manifest for rust_owner="Context" (Add64/Add32/
// Add128BIDWithContext).
func buildContextRs(manifest *manifestFile) (string, error) {
	var add64Op, add32Op, add128Op *decOp
	for _, r := range manifest.Emit {
		if r.RustOwner != "Context" {
			continue
		}
		switch r.Shape {
		case "context_binary_with_flags":
			p, ok := portPath[r.BidgoFunction]
			if !ok {
				return "", fmt.Errorf("apiemit: no port path for bidgo_function %q (go_symbol %q)", r.BidgoFunction, r.GoSymbol)
			}
			o := decOp{method: r.RustSurface, port: p.fn, module: p.module, boolPort: boolResultPorts[r.BidgoFunction], pfpsf: portPfpsf[r.BidgoFunction]}
			switch r.GoSymbol {
			case "Add64BIDWithContext":
				add64Op = &o
			case "Add32BIDWithContext":
				add32Op = &o
			case "Add128BIDWithContext":
				add128Op = &o
			default:
				return "", fmt.Errorf("apiemit: unhandled Context context_binary_with_flags go_symbol %q (only Add64BIDWithContext/Add32BIDWithContext/Add128BIDWithContext are wired to a file builder today)", r.GoSymbol)
			}
		default:
			return "", fmt.Errorf("apiemit: unhandled Context shape %q for go_symbol %q", r.Shape, r.GoSymbol)
		}
	}

	var b strings.Builder
	b.WriteString(marker() + crlf)
	b.WriteString("//! Arithmetic context: an explicit rounding mode plus accumulated flags." + crlf)
	b.WriteString(crlf)
	if add64Op != nil {
		b.WriteString("use super::decimal64::Decimal64;" + crlf)
	}
	if add32Op != nil {
		b.WriteString("use super::decimal32::Decimal32;" + crlf)
	}
	if add128Op != nil {
		b.WriteString("use super::decimal128::Decimal128;" + crlf)
	}
	b.WriteString("use super::types::{ExceptionFlags, RoundingMode};" + crlf)
	b.WriteString(crlf)
	b.WriteString(`/// An explicit arithmetic context: a rounding mode and an accumulated set of
/// raised exception flags. Unlike the Go ArithmeticContext there is no heap or
/// pointer indirection; flag accumulation is via ` + "`&mut self`" + `.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct Context {
    /// The rounding mode applied by context operations.
    pub rounding: RoundingMode,
    /// The accumulated exception flags.
    pub flags: ExceptionFlags,
}

impl Context {
    /// A new context with the IEEE default rounding (ties to even) and no flags.
    pub const fn new() -> Context {
        Context {
            rounding: RoundingMode::NearestEven,
            flags: ExceptionFlags::empty(),
        }
    }

    /// A new context with the given rounding mode and no flags.
    pub const fn with_rounding(rounding: RoundingMode) -> Context {
        Context {
            rounding,
            flags: ExceptionFlags::empty(),
        }
    }

    /// Raises the given flags in the context (IEEE 754-2019 5.7.4 raiseFlags).
    pub fn set_flag(&mut self, flags: ExceptionFlags) {
        self.flags |= flags;
    }

    /// Reports whether any of the given flags is set (5.7.4 testFlags).
    pub fn has_flag(&self, flags: ExceptionFlags) -> bool {
        !(self.flags & flags).is_empty()
    }

    /// Clears the given flags (5.7.4 lowerFlags).
    pub fn clear_flag(&mut self, flags: ExceptionFlags) {
        self.flags = self.flags.difference(flags);
    }

    /// Clears all flags.
    pub fn clear_all_flags(&mut self) {
        self.flags = ExceptionFlags::empty();
    }

    /// Returns a snapshot of all currently raised flags (5.7.4 saveAllFlags).
    pub fn save_all_flags(&self) -> ExceptionFlags {
        self.flags
    }

    /// Restores the flags selected by ` + "`mask`" + ` to their values in ` + "`saved`" + `
    /// and preserves the rest (IEEE 754-2019 5.7.4 restoreFlags). This is the
    /// same masked write as the Go ArithmeticContext.RestoreFlags: the whole
    /// ExceptionFlags domain is public, so no implicit masking is applied and
    /// bits outside ` + "`mask`" + ` keep their current value.
    pub fn restore_flags(&mut self, saved: ExceptionFlags, mask: ExceptionFlags) {
        self.flags = self.flags.difference(mask) | (saved & mask);
    }
`)
	if add64Op != nil {
		emitContextAddOp(&b, *add64Op, decimal64Width)
	}
	if add32Op != nil {
		emitContextAddOp(&b, *add32Op, decimal32Width)
	}
	if add128Op != nil {
		emitContextAddOp(&b, *add128Op, decimal128Width)
	}
	b.WriteString(`}

impl Default for Context {
    fn default() -> Context {
        Context::new()
    }
}
`)
	return b.String(), nil
}

// emitContextAddOp renders the Context method for width w's
// context_binary_with_flags row (Add64BIDWithContext / Add32BIDWithContext /
// Add128BIDWithContext): applies the port op at the context's own rounding
// mode and accumulates the raised flags into self.flags. Context lives in a
// different module than decimal<w>.rs, so it uses the PUBLIC to_bits()/
// from_bits() (or to_le_bytes()/from_le_bytes() for width 128) accessors via
// crossModuleArg/crossModuleWrap, not the private tuple field selfArg/
// wrapResult use for a receiver method inside that width's own module.
// Add128BIDWithContext's port (Bid128Add) uses the status-pointer convention;
// flagsCallStmt handles that the same way every other op does.
func emitContextAddOp(b *strings.Builder, op decOp, w widthSpec) {
	stmt := flagsCallStmt(op, []string{w.crossModuleArg("a"), w.crossModuleArg("b"), "super::types::to_bidgo_rounding(self.rounding)"}, "bits", "raw")
	fmt.Fprintf(b, `
    /// Adds a and b at the context's rounding mode and accumulates the
    /// raised flags into self.flags (mirrors the Go Add%sBIDWithContext).
    /// Rust has no nil-context default (the Go global SetDefaultRounding is
    /// outside the declared Rust public surface); the mode always comes
    /// from self.rounding, which the closed RoundingMode enum guarantees is
    /// always one of the five defined IEEE modes.
    pub fn %s(&mut self, a: %s, b: %s) -> %s {
        %s
        self.flags |= ExceptionFlags::from_bidgo(raw);
        %s
    }
`, w.digits, op.method, w.selfType, w.selfType, w.selfType, stmt, w.crossModuleWrap("bits"))
}

// buildDecimal64Rs and buildDecimal32Rs (defined next to buildDecimal32Rs's
// thin wrapper above) share this entire implementation; only the widthSpec
// parameter differs. buildDecimalRs emits the Decimal<w> value type plus
// every wrapper method selected by the manifest emit rules for that width.
// This includes every emitted symbol owned by Decimal<w>BID and every emitted
// free constructor. Each wrapper body only calls a crate::generated::* port
// function, converts value types, and maps flags/rounding; NaN-literal
// parsing and NaN display formatting reproduce Go public-API plumbing
// bit-for-bit (types_bid_nan_payload.go), not arithmetic. Decimal64's output
// uses the same width-parameterized templates.
func buildDecimal64Rs(manifest *manifestFile) (string, error) {
	return buildDecimalRs(manifest, decimal64Width)
}

func buildDecimalRs(manifest *manifestFile, w widthSpec) (string, error) {
	emitParse := false
	emitParseRaw := false
	emitParseWithFlags := false
	emitParseMode := false
	emitDisplay := false
	emitFromInt := false
	fromIntPortHasFlags := false
	emitFromI32Exact := false
	emitFromU32Exact := false
	emitFromI64Exact := false
	emitFromU64Exact := false
	emitFromI64Mode := false
	emitFromU64Mode := false
	emitFromI32Mode := false
	emitFromU32Mode := false

	var fromIntMethod, fromIntParamType, parseWithFlagsMethod string
	var fromIntOp decOp
	var parseModeMethod string
	var parseModeOp decOp
	var fromI32ExactOp, fromU32ExactOp decOp
	var fromI64ExactOp, fromU64ExactOp decOp
	var fromI64ModeOp, fromU64ModeOp, fromI32ModeOp, fromU32ModeOp decOp
	var parseRawOp, displayOp decOp

	var bins, binsFlags, binModeFlags []decOp
	var mixedBinModeFlags []mixedDecOp
	var mixedTernaryModeFlags []mixedTernaryDecOp
	var mixedUnaryModeFlags []mixedUnaryDecOp
	var unaryModeFlagsOps, ternaryModeFlagsOps, scalebModeOps []decOp
	var unaryOps, predicateOps []decOp
	var unaryFlagsNoRoundOps, unaryFlagsDefaultRoundOps []decOp
	var unaryIntFlagsNoRoundOps []decOp
	var binaryFlagsNoRoundOps, compareBoolFlagsOps []decOp
	var convOps []decConvOp

	var copysignOp, fmaOp, sameQuantumOp, classOp *decOp
	var totalCmpOp, totalCmpMagOp *decOp
	var quantizeDropOp, roundIntegralExactDropOp *decOp
	var nextTowardOp, scalebOp *decOp
	var toBinary32Op, toBinary64Op, toBinary128Op *decOp
	var toDecimal128Op, toDecimal32Op, toDecimal64Op, toDecimal64ModeOp *decOp
	var isNaNOp decOp

	var signMethod string
	var signIsZeroOp, signIsSignedOp decOp
	var radixMethod string
	var signalingEqMethod string
	var signalingEqGEOp, signalingEqLEOp decOp
	var signalingNotEqMethod string

	// Method names of the three quiet comparison surfaces the idiomatic
	// PartialEq/PartialOrd trait impls delegate to (captured from the emitted
	// compare_bool_flags rows by their census bidgo_function, so the traits use
	// whatever name the manifest assigns rather than a hardcoded string).
	var quietEqMethod, quietLtMethod, quietGtMethod string

	// This width's own bidgo function names for the census bidgo_function
	// equality checks and auxiliary portPath lookups below (class/sign/
	// signaling_eq_compose/predicate-IsNaN), computed once instead of
	// re-splicing "Bid"+w.digits+... at each call site.
	bidQuietEqual := w.bidgoPrefix() + "QuietEqual"
	bidQuietLess := w.bidgoPrefix() + "QuietLess"
	bidQuietGreater := w.bidgoPrefix() + "QuietGreater"
	bidClass := w.bidgoPrefix() + "Class"
	bidIsSigned := w.bidgoPrefix() + "IsSigned"
	bidIsZero := w.bidgoPrefix() + "IsZero"
	bidIsNaN := w.bidgoPrefix() + "IsNaN"
	bidSignalingGE := w.bidgoPrefix() + "SignalingGreaterEqual"
	bidSignalingLE := w.bidgoPrefix() + "SignalingLessEqual"

	for _, r := range manifest.Emit {
		if r.RustOwner != w.selfType {
			continue
		}
		switch r.Shape {
		case "parse":
			emitParse = true
			continue
		case "parse_fold", "copy_fold":
			// folded convenience twins (NewDecimal<w>BIDDirect, Copy); nothing
			// to emit beyond what parse/#[derive(Copy)] already provide.
			continue
		case "parse_with_flags":
			emitParseWithFlags = true
			parseWithFlagsMethod = r.RustSurface
			continue
		}

		p, ok := portPath[r.BidgoFunction]
		if !ok {
			return "", fmt.Errorf("apiemit: no port path for bidgo_function %q (go_symbol %q)", r.BidgoFunction, r.GoSymbol)
		}
		op := decOp{method: r.RustSurface, port: p.fn, module: p.module, boolPort: boolResultPorts[r.BidgoFunction], pfpsf: portPfpsf[r.BidgoFunction]}

		if r.Shape == "parse_raw" {
			emitParseRaw = true
			parseRawOp = op
			continue
		}
		if r.Shape == "parse_mode" {
			emitParseMode = true
			parseModeMethod = r.RustSurface
			parseModeOp = op
			continue
		}
		if r.Shape == "display" {
			emitDisplay = true
			displayOp = op
			continue
		}

		if ct, ok := convShapeTypes[r.Shape]; ok {
			convOps = append(convOps, decConvOp{method: r.RustSurface, rustType: ct.rustType, canonical: r.BidgoFunction, exact: ct.exact})
			continue
		}
		if operands, ok := mixedBinaryShapeOperands[r.Shape]; ok {
			left, leftOK := widthSpecForOwner(strings.TrimSuffix(operands[0], "BID"))
			right, rightOK := widthSpecForOwner(strings.TrimSuffix(operands[1], "BID"))
			if !leftOK || !rightOK {
				return "", fmt.Errorf("apiemit: mixed shape %q for go_symbol %q has unsupported operand types %v", r.Shape, r.GoSymbol, operands)
			}
			mixedBinModeFlags = append(mixedBinModeFlags, mixedDecOp{decOp: op, left: left, right: right})
			continue
		}
		if operands, ok := mixedTernaryShapeOperands[r.Shape]; ok {
			var widths [3]widthSpec
			for i, operand := range operands {
				operandWidth, widthOK := widthSpecForOwner(strings.TrimSuffix(operand, "BID"))
				if !widthOK {
					return "", fmt.Errorf("apiemit: mixed ternary shape %q for go_symbol %q has unsupported operand type %q", r.Shape, r.GoSymbol, operand)
				}
				widths[i] = operandWidth
			}
			mixedTernaryModeFlags = append(mixedTernaryModeFlags, mixedTernaryDecOp{decOp: op, operands: widths})
			continue
		}
		if operand, ok := mixedUnaryShapeOperands[r.Shape]; ok {
			operandWidth, widthOK := widthSpecForOwner(strings.TrimSuffix(operand, "BID"))
			if !widthOK {
				return "", fmt.Errorf("apiemit: mixed unary shape %q for go_symbol %q has unsupported operand type %q", r.Shape, r.GoSymbol, operand)
			}
			mixedUnaryModeFlags = append(mixedUnaryModeFlags, mixedUnaryDecOp{decOp: op, operand: operandWidth})
			continue
		}

		switch r.Shape {
		case "binary":
			bins = append(bins, op)
		case "binary_with_flags":
			binsFlags = append(binsFlags, op)
		case "binary_mode_flags":
			binModeFlags = append(binModeFlags, op)
		case "unary_mode_flags":
			unaryModeFlagsOps = append(unaryModeFlagsOps, op)
		case "ternary_mode_flags":
			ternaryModeFlagsOps = append(ternaryModeFlagsOps, op)
		case "scaleb_mode":
			scalebModeOps = append(scalebModeOps, op)
		case "unary":
			unaryOps = append(unaryOps, op)
		case "predicate":
			predicateOps = append(predicateOps, op)
			if r.BidgoFunction == bidIsNaN {
				isNaNOp = op
			}
		case "unary_with_flags_no_round":
			unaryFlagsNoRoundOps = append(unaryFlagsNoRoundOps, op)
		case "unary_int_with_flags_no_round":
			unaryIntFlagsNoRoundOps = append(unaryIntFlagsNoRoundOps, op)
		case "unary_with_flags_default_round":
			unaryFlagsDefaultRoundOps = append(unaryFlagsDefaultRoundOps, op)
		case "binary_flags_no_round":
			binaryFlagsNoRoundOps = append(binaryFlagsNoRoundOps, op)
		case "compare_bool_flags":
			compareBoolFlagsOps = append(compareBoolFlagsOps, op)
			switch r.BidgoFunction {
			case bidQuietEqual:
				quietEqMethod = r.RustSurface
			case bidQuietLess:
				quietLtMethod = r.RustSurface
			case bidQuietGreater:
				quietGtMethod = r.RustSurface
			}
		case "copysign":
			o := op
			copysignOp = &o
		case "fma":
			o := op
			fmaOp = &o
		case "same_quantum":
			o := op
			sameQuantumOp = &o
		case "class":
			if r.BidgoFunction != bidClass {
				return "", fmt.Errorf("apiemit: class shape go_symbol %q expects bidgo_function %q, got %q", r.GoSymbol, bidClass, r.BidgoFunction)
			}
			o := op
			classOp = &o
		case "total_cmp":
			o := op
			totalCmpOp = &o
		case "total_cmp_mag":
			o := op
			totalCmpMagOp = &o
		case "sign":
			if r.BidgoFunction != bidIsSigned {
				return "", fmt.Errorf("apiemit: sign shape go_symbol %q expects bidgo_function %q, got %q", r.GoSymbol, bidIsSigned, r.BidgoFunction)
			}
			isZeroP, ok := portPath[bidIsZero]
			if !ok {
				return "", fmt.Errorf("apiemit: sign shape needs portPath[%q]", bidIsZero)
			}
			signMethod = r.RustSurface
			signIsZeroOp = decOp{module: isZeroP.module, port: isZeroP.fn, boolPort: boolResultPorts[bidIsZero], pfpsf: portPfpsf[bidIsZero]}
			signIsSignedOp = op
		case "radix_const":
			radixMethod = r.RustSurface
		case "binary_drop_flags":
			o := op
			quantizeDropOp = &o
		case "unary_mode_drop_flags":
			o := op
			roundIntegralExactDropOp = &o
		case "next_toward":
			o := op
			nextTowardOp = &o
		case "scaleb":
			o := op
			scalebOp = &o
		case "signaling_eq_compose":
			if r.BidgoFunction != bidSignalingGE {
				return "", fmt.Errorf("apiemit: signaling_eq_compose shape go_symbol %q expects bidgo_function %q, got %q", r.GoSymbol, bidSignalingGE, r.BidgoFunction)
			}
			leP, ok := portPath[bidSignalingLE]
			if !ok {
				return "", fmt.Errorf("apiemit: signaling_eq_compose shape needs portPath[%q]", bidSignalingLE)
			}
			signalingEqMethod = r.RustSurface
			signalingEqGEOp = op
			signalingEqLEOp = decOp{module: leP.module, port: leP.fn, boolPort: boolResultPorts[bidSignalingLE], pfpsf: portPfpsf[bidSignalingLE]}
		case "signaling_not_eq_compose":
			signalingNotEqMethod = r.RustSurface
		case "to_binary32":
			o := op
			toBinary32Op = &o
		case "to_binary64":
			o := op
			toBinary64Op = &o
		case "to_binary128":
			o := op
			toBinary128Op = &o
		case "to_decimal128":
			o := op
			toDecimal128Op = &o
		case "to_decimal64":
			o := op
			toDecimal64Op = &o
		case "to_decimal32":
			o := op
			toDecimal32Op = &o
		case "to_decimal64_mode":
			o := op
			toDecimal64ModeOp = &o
		case "from_i32_exact":
			emitFromI32Exact = true
			fromI32ExactOp = op
		case "from_u32_exact":
			emitFromU32Exact = true
			fromU32ExactOp = op
		case "from_i64_exact":
			emitFromI64Exact = true
			fromI64ExactOp = op
		case "from_u64_exact":
			emitFromU64Exact = true
			fromU64ExactOp = op
		case "from_i32_exact_or_error", "from_i64_exact_or_error":
			emitFromInt = true
			fromIntMethod = r.RustSurface
			fromIntOp = op
			switch r.BidgoFunction {
			case "Bid32FromInt32", "Bid64FromInt64":
				fromIntPortHasFlags = true
			case "Bid128FromInt64":
				fromIntPortHasFlags = false
			default:
				return "", fmt.Errorf("apiemit: exact-or-error integer shape has unsupported port %q for go_symbol %q", r.BidgoFunction, r.GoSymbol)
			}
			if r.Shape == "from_i32_exact_or_error" {
				fromIntParamType = "i32"
			} else {
				fromIntParamType = "i64"
			}
		case "from_i64_mode":
			emitFromI64Mode = true
			fromI64ModeOp = op
		case "from_u64_mode":
			emitFromU64Mode = true
			fromU64ModeOp = op
		case "from_i32_mode":
			emitFromI32Mode = true
			fromI32ModeOp = op
		case "from_u32_mode":
			emitFromU32Mode = true
			fromU32ModeOp = op
		default:
			return "", fmt.Errorf("apiemit: unhandled shape %q for go_symbol %q", r.Shape, r.GoSymbol)
		}
	}

	if signalingNotEqMethod != "" && signalingEqMethod == "" {
		return "", fmt.Errorf("apiemit: signaling_not_eq_compose shape needs signaling_eq_compose to also be emitted")
	}

	needFromStr := emitParse
	needFlags := emitParseRaw || emitParseWithFlags || emitParseMode || len(binsFlags) > 0 || len(binModeFlags) > 0 || len(mixedBinModeFlags) > 0 || len(mixedTernaryModeFlags) > 0 || len(mixedUnaryModeFlags) > 0 ||
		len(unaryModeFlagsOps) > 0 || len(ternaryModeFlagsOps) > 0 || len(scalebModeOps) > 0 ||
		len(unaryFlagsNoRoundOps) > 0 || len(unaryFlagsDefaultRoundOps) > 0 ||
		len(unaryIntFlagsNoRoundOps) > 0 ||
		len(binaryFlagsNoRoundOps) > 0 || len(compareBoolFlagsOps) > 0 ||
		len(convOps) > 0 || fmaOp != nil || nextTowardOp != nil || scalebOp != nil ||
		signalingEqMethod != "" || signalingNotEqMethod != "" ||
		toBinary32Op != nil || toBinary64Op != nil || toBinary128Op != nil ||
		toDecimal128Op != nil || toDecimal32Op != nil || toDecimal64Op != nil || toDecimal64ModeOp != nil ||
		emitFromInt || emitFromI64Mode || emitFromU64Mode || emitFromI32Mode || emitFromU32Mode
	needRoundingMode := len(convOps) > 0 || len(binModeFlags) > 0 || len(mixedBinModeFlags) > 0 || len(mixedTernaryModeFlags) > 0 || len(mixedUnaryModeFlags) > 0 || emitParseMode ||
		len(unaryModeFlagsOps) > 0 || len(ternaryModeFlagsOps) > 0 || len(scalebModeOps) > 0 ||
		toBinary32Op != nil || toBinary64Op != nil ||
		toBinary128Op != nil || toDecimal32Op != nil || toDecimal64ModeOp != nil ||
		emitFromI64Mode || emitFromU64Mode || emitFromI32Mode || emitFromU32Mode
	needDecimalClass := classOp != nil
	needBinary128 := toBinary128Op != nil
	needDecimal128 := nextTowardOp != nil || toDecimal128Op != nil || len(mixedBinModeFlags) > 0 || len(mixedTernaryModeFlags) > 0 || len(mixedUnaryModeFlags) > 0
	needDecimal64 := toDecimal64Op != nil || toDecimal64ModeOp != nil || len(mixedBinModeFlags) > 0 || len(mixedTernaryModeFlags) > 0 || len(mixedUnaryModeFlags) > 0
	needDecimal32 := toDecimal32Op != nil
	needParseError := emitParse || emitParseWithFlags || emitParseMode
	needInexactIntegerError := emitFromInt

	var b strings.Builder
	b.WriteString(marker() + crlf)
	b.WriteString("//! The " + w.digits + "-bit decimal value type and its wrapper surface." + crlf)
	b.WriteString(crlf)

	var imports []string
	if emitDisplay || needFromStr {
		imports = append(imports, "use core::fmt;")
	}
	if needFromStr {
		imports = append(imports, "use core::str::FromStr;")
	}
	if emitParseRaw {
		imports = append(imports, "use num_bigint::BigInt;")
	}
	// A width never imports its own module (e.g. Decimal128's own file must
	// not `use super::decimal128::Decimal128` -- that is a self-import
	// error): NextToward's Decimal128 target is `Self` when the receiver is
	// itself Decimal128, so the type is already in scope.
	if needDecimal128 && w.selfType != "Decimal128" {
		imports = append(imports, "use super::decimal128::Decimal128;")
	}
	if needDecimal64 && w.selfType != "Decimal64" {
		imports = append(imports, "use super::decimal64::Decimal64;")
	}
	if needDecimal32 && w.selfType != "Decimal32" {
		imports = append(imports, "use super::decimal32::Decimal32;")
	}
	var typeUses []string
	if needFlags {
		typeUses = append(typeUses, "ExceptionFlags")
	}
	if needParseError {
		typeUses = append(typeUses, "ParseDecimalError")
	}
	if needInexactIntegerError {
		typeUses = append(typeUses, "InexactIntegerError")
	}
	if needRoundingMode {
		typeUses = append(typeUses, "RoundingMode")
	}
	if needDecimalClass {
		typeUses = append(typeUses, "DecimalClass")
	}
	if needBinary128 {
		typeUses = append(typeUses, "Binary128")
	}
	if len(typeUses) > 0 {
		sort.Strings(typeUses)
		imports = append(imports, "use super::types::{"+strings.Join(typeUses, ", ")+"};")
	}
	for _, imp := range imports {
		b.WriteString(imp + crlf)
	}
	b.WriteString(crlf)
	b.WriteString(`/// bidgo-domain rounding value for round-ties-to-even (the IEEE default);
/// most wrappers below pass this to the port directly. Explicit-mode
/// wrappers instead call super::types::to_bidgo_rounding, which maps the full
/// closed RoundingMode enum.
const BIDGO_ROUND_NEAREST_EVEN: i64 = 0;

`)
	if w.is128 {
		// Decimal128's raw representation is a 16-byte array, not a single
		// primitive int (widthSpec.bitsType is unused for is128 -- see its
		// doc comment), so it gets its own from_le_bytes/to_le_bytes
		// accessor pair instead of 32/64's from_bits/to_bits.
		b.WriteString(tmpl(`/// A 128-bit IEEE 754-2019 decimal (BID encoding): a fixed-width value type with
/// 1:1 byte correspondence to its 16-byte pattern. The public byte contract is
/// little-endian, matching every guaranteed target platform.
///
/// @B@==@B@ and @B@<@B@ / @B@>@B@ use IEEE 754 quiet comparison semantics (matching the
/// @B@f64@B@ precedent): a NaN is unordered and unequal to everything including
/// itself, and @B@-0 == +0@B@. This type is therefore deliberately not @B@Eq@B@ or
/// @B@Hash@B@ (quiet equality is not reflexive). For bit-identity comparison, a
/// total order, or use as a hash-map key, use @B@to_le_bytes()@B@ (or @B@total_cmp@B@).
#[derive(Clone, Copy, Debug)]
#[repr(transparent)]
pub struct Decimal128([u8; 16]);

const _: () = assert!(core::mem::size_of::<Decimal128>() == 16);

impl Decimal128 {
    /// Reinterprets 16 little-endian bytes as a @B@Decimal128@B@.
    pub const fn from_le_bytes(bytes: [u8; 16]) -> Decimal128 {
        Decimal128(bytes)
    }

    /// Returns the raw 128-bit BID pattern as 16 little-endian bytes.
    pub const fn to_le_bytes(self) -> [u8; 16] {
        self.0
    }
`, "B", "`"))
	} else {
		b.WriteString(tmpl(`/// A @DIGITS@-bit IEEE 754-2019 decimal (BID encoding): a fixed-width value type with
/// 1:1 byte correspondence to its @B@@BITS@@B@ bit pattern.
///
/// @B@==@B@ and @B@<@B@ / @B@>@B@ use IEEE 754 quiet comparison semantics (matching the
/// @B@f64@B@ precedent): a NaN is unordered and unequal to everything including
/// itself, and @B@-0 == +0@B@. This type is therefore deliberately not @B@Eq@B@ or
/// @B@Hash@B@ (quiet equality is not reflexive). For bit-identity comparison, a
/// total order, or use as a hash-map key, use @B@to_bits()@B@ (or @B@total_cmp@B@).
#[derive(Clone, Copy, Debug)]
#[repr(transparent)]
pub struct @SELF@(@BITS@);

const _: () = assert!(core::mem::size_of::<@SELF@>() == @BYTES@);

impl @SELF@ {
    /// Reinterprets a raw @DIGITS@-bit BID pattern as a @B@@SELF@@B@.
    pub const fn from_bits(bits: @BITS@) -> @SELF@ {
        @SELF@(bits)
    }

    /// Returns the raw @DIGITS@-bit BID pattern.
    pub const fn to_bits(self) -> @BITS@ {
        self.0
    }
`, "DIGITS", w.digits, "BITS", w.bitsType, "SELF", w.selfType, "BYTES", w.byteSize, "B", "`"))
	}

	if radixMethod != "" {
		emitRadixConstOp(&b, radixMethod, w)
	}

	// associated constants: ZERO/ONE/PI/E-style compile-time constants (declared API contract). These
	// come from manifest.Constants, an exhaustive set SEPARATE from
	// manifest.Emit above (resolveConstantsClosure in apiemit.go already
	// validated every row before Run() ever calls this function), filtered to
	// this width and ordered fixed (ZERO, ONE, PI, E) for a deterministic,
	// Go-source-order-matching diff regardless of manifest JSON array order.
	var constRows []constantRule
	for _, r := range manifest.Constants {
		if r.RustOwner == w.selfType {
			constRows = append(constRows, r)
		}
	}
	if len(constRows) > 0 {
		sort.Slice(constRows, func(i, j int) bool {
			return constEmitOrder[constRows[i].RustConst] < constEmitOrder[constRows[j].RustConst]
		})
		if err := emitConstOps(&b, constRows, w); err != nil {
			return "", err
		}
	}

	if emitParse {
		b.WriteString(tmpl(`
    /// Parses a decimal string literal exactly, returning an error for
    /// malformed or unrepresentable input. Use parse_with_flags or
    /// parse_with_mode when a rounded finite result is desired.
    ///
    /// Mirrors the Go NewDecimal@DIGITS@/NewDecimal@DIGITS@BIDDirect Ok/Err decision. A
    /// NaN-literal input (e.g. "NaN42" or "-SNaN") is recognized directly by
    /// the public NaN-literal grammar and never reaches the port, so the
    /// returned NaN's payload bits match the Go public NaN-literal
    /// construction exactly (see parse_decimal@DIGITS@_nan below).
    pub fn parse(s: &str) -> Result<@SELF@, ParseDecimalError> {
        let (value, flags) = @SELF@::parse_raw(s);
        if rejected_bid_string_input(s, result_is_nan(value.0), flags)
            || unrepresentable_bid_string_flags(flags)
        {
            return Err(ParseDecimalError::new(s));
        }
        Ok(value)
    }
`, "DIGITS", w.digits, "SELF", w.selfType))
	}
	if emitParseRaw {
		// parse_raw's port argument is the input string s (not a decimal
		// operand), so it needs no selfArg conversion; Bid128FromString is
		// tuple-return like every 32/64 from-string port (verified, not
		// pfpsf), so the tuple destructure stays valid for every width -- only
		// the result wrap (bits -> Self) needs to be width-aware.
		b.WriteString(tmpl(`
    /// Parses a decimal string through the Intel from_string path and returns
    /// the raised flags. A representable NaN-literal input (the public grammar
    /// recognized by parse_decimal@DIGITS@_nan) is constructed directly,
    /// without calling the port or raising flags. Malformed syntax, a NaN payload
    /// that does not fit this width, or a finite literal whose written quantum
    /// or coefficient the port would otherwise coerce with zero status returns
    /// canonical quiet NaN plus INVALID_OPERATION. Valid finite input retains
    /// the port's overflow/underflow/inexact result and flags.
    pub fn parse_raw(s: &str) -> (@SELF@, ExceptionFlags) {
        if let Some(value) = parse_decimal@DIGITS@_nan(s) {
            return (value, ExceptionFlags::empty());
        }
        if super::types::parse_bid_nan_literal(s).is_some() {
            return (@QNAN@, ExceptionFlags::INVALID_OPERATION);
        }
        let (bits, raw) = crate::generated::@MOD@::@PORT@(s, BIDGO_ROUND_NEAREST_EVEN);
        let value = @WRAPPED@;
        let flags = ExceptionFlags::from_bidgo(raw);
        if invalid_bid_string_input(s, result_is_nan(value.0)) {
            return (@QNAN@, ExceptionFlags::INVALID_OPERATION);
        }
        if raw == 0 && bid_finite_literal_cohort_unrepresentable(s, @MINQ@, @MAXQ@, @PRECISION@) {
            return (@QNAN@, ExceptionFlags::INVALID_OPERATION);
        }
        (value, flags)
    }
`, "DIGITS", w.digits, "SELF", w.selfType, "QNAN", w.canonicalQNaNResult(), "MINQ", w.minQuantum, "MAXQ", w.maxQuantum, "PRECISION", w.precision, "MOD", parseRawOp.module, "PORT", parseRawOp.port, "WRAPPED", w.wrapResult("bits")))
	}
	if emitParseWithFlags {
		emitParseWithFlagsOp(&b, parseWithFlagsMethod, w)
	}
	if emitParseMode {
		emitParseModeOp(&b, parseModeMethod, parseModeOp, w)
	}
	if emitFromInt {
		emitFromIntExactOrErrorOp(&b, fromIntMethod, fromIntOp, w, fromIntParamType, fromIntPortHasFlags)
	}
	if emitFromI64Mode {
		emitFromIntModeOp(&b, fromI64ModeOp, w, "i64")
	}
	if emitFromU64Mode {
		emitFromIntModeOp(&b, fromU64ModeOp, w, "u64")
	}
	if emitFromI32Mode {
		emitFromIntModeOp(&b, fromI32ModeOp, w, "i32")
	}
	if emitFromU32Mode {
		emitFromIntModeOp(&b, fromU32ModeOp, w, "u32")
	}

	sortDecOps(bins)
	for _, op := range bins {
		// "binary" is the one shape whose call SHAPE (not just its argument/
		// result conversion) genuinely forks by width: every 32/64 bidgo
		// function this shape uses (Bid64Add, Bid64Mul, ...) is a truly
		// separate, flags-less function distinct from its *WithFlags sibling,
		// but Decimal128's census maps Add and AddWithFlags (etc.) to the
		// SAME bidgo_function -- its generated:: signature always carries a
		// flags word (via pfpsf or tuple, per-function, verified), which this
		// flags-dropping shape must still satisfy and then discard.
		var body string
		if w.is128 {
			stmt := flagsCallStmt(op, []string{w.selfArg("self"), w.selfArg("rhs"), "BIDGO_ROUND_NEAREST_EVEN"}, "bits", "_raw")
			body = stmt + "\n        " + w.wrapResult("bits")
		} else {
			body = tmpl(`@SELF@(crate::generated::@MOD@::@PORT@(
            self.0,
            rhs.0,
            BIDGO_ROUND_NEAREST_EVEN,
        ))`, "SELF", w.selfType, "MOD", op.module, "PORT", op.port)
		}
		verb, err := methodVerb(op.method)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, `
    /// %s two decimals with the IEEE default round-ties-to-even rounding.
    // Inherent named operation (matches the Go .%s method); the std::ops
    // operator trait is outside the declared generated surface.
    #[allow(clippy::should_implement_trait)]
    pub fn %s(self, rhs: %s) -> %s {
        %s
    }
`, verb, goMethodName(op.method), op.method, w.selfType, w.selfType, body)
	}
	sortDecOps(binsFlags)
	for _, op := range binsFlags {
		base := strings.TrimSuffix(op.method, "_with_flags")
		stmt := flagsCallStmt(op, []string{w.selfArg("self"), w.selfArg("rhs"), "BIDGO_ROUND_NEAREST_EVEN"}, "bits", "raw")
		clause, err := binaryFlagsClause(base)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, `
    /// %s with default rounding, also returning the raised flags.
    pub fn %s(self, rhs: %s) -> (%s, ExceptionFlags) {
        %s
        (%s, ExceptionFlags::from_bidgo(raw))
    }
`, clause, op.method, w.selfType, w.selfType, stmt, w.wrapResult("bits"))
	}
	if err := emitBinaryModeFlagsOps(&b, binModeFlags, w); err != nil {
		return "", err
	}
	if err := emitMixedBinaryModeFlagsOps(&b, mixedBinModeFlags, w); err != nil {
		return "", err
	}
	if err := emitMixedTernaryModeFlagsOps(&b, mixedTernaryModeFlags, w); err != nil {
		return "", err
	}
	if err := emitMixedUnaryModeFlagsOps(&b, mixedUnaryModeFlags, w); err != nil {
		return "", err
	}
	if err := emitUnaryModeFlagsOps(&b, unaryModeFlagsOps, w); err != nil {
		return "", err
	}
	if err := emitTernaryModeFlagsOps(&b, ternaryModeFlagsOps, w); err != nil {
		return "", err
	}
	if err := emitScalebModeOps(&b, scalebModeOps, w); err != nil {
		return "", err
	}
	if fmaOp != nil {
		emitFMAOp(&b, *fmaOp, w)
	}

	emitUnaryOps(&b, unaryOps, w)
	if copysignOp != nil {
		emitCopySignOp(&b, *copysignOp, w)
	}

	emitBinaryFlagsNoRoundOps(&b, binaryFlagsNoRoundOps, w)
	if quantizeDropOp != nil {
		emitQuantizeDropOp(&b, *quantizeDropOp, w)
	}

	emitUnaryFlagsNoRoundOps(&b, unaryFlagsNoRoundOps, w)
	emitUnaryIntFlagsNoRoundOps(&b, unaryIntFlagsNoRoundOps, w)
	emitUnaryFlagsDefaultRoundOps(&b, unaryFlagsDefaultRoundOps, w)
	if roundIntegralExactDropOp != nil {
		emitRoundIntegralExactDropOp(&b, *roundIntegralExactDropOp, w)
	}

	emitPredicateOps(&b, predicateOps, w)
	if sameQuantumOp != nil {
		emitSameQuantumOp(&b, *sameQuantumOp, w)
	}
	if classOp != nil {
		emitClassOp(&b, *classOp, w)
	}
	if signMethod != "" {
		emitSignOp(&b, signMethod, signIsZeroOp, signIsSignedOp, w)
	}
	if totalCmpOp != nil {
		emitTotalCmpOp(&b, *totalCmpOp, w)
	}
	if totalCmpMagOp != nil {
		emitTotalCmpOp(&b, *totalCmpMagOp, w)
	}

	if scalebOp != nil {
		emitScaleBOp(&b, *scalebOp, w)
	}
	if nextTowardOp != nil {
		emitNextTowardOp(&b, *nextTowardOp, w)
	}

	emitCompareBoolFlagsOps(&b, compareBoolFlagsOps, w)
	if signalingEqMethod != "" {
		emitSignalingEqOp(&b, signalingEqMethod, signalingEqGEOp, signalingEqLEOp, w)
	}
	if signalingNotEqMethod != "" {
		emitSignalingNotEqOp(&b, signalingNotEqMethod, signalingEqMethod, w)
	}

	if toBinary32Op != nil {
		emitToBinary32Op(&b, *toBinary32Op, w)
	}
	if toBinary64Op != nil {
		emitToBinary64Op(&b, *toBinary64Op, w)
	}
	if toBinary128Op != nil {
		emitToBinary128Op(&b, *toBinary128Op, w)
	}
	if toDecimal128Op != nil {
		emitToDecimal128Op(&b, *toDecimal128Op)
	}
	if toDecimal64Op != nil {
		emitToDecimal64Op(&b, *toDecimal64Op)
	}
	if toDecimal32Op != nil {
		emitToDecimal32Op(&b, *toDecimal32Op, w)
	}
	if toDecimal64ModeOp != nil {
		emitToDecimal64ModeOp(&b, *toDecimal64ModeOp, w)
	}

	if err := emitConvOps(&b, convOps, w); err != nil {
		return "", err
	}

	b.WriteString("}" + crlf)

	if emitDisplay {
		b.WriteString(tmpl(`
impl fmt::Display for @SELF@ {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        // A NaN value is formatted directly from its bits (matches the Go
        // decimal@DIGITS@BIDStringPort NaN branch: sign, NaN/SNaN, canonical
        // payload). Every other value routes through the Intel
        // @PORT@ converter.
        if let Some(s) = format_decimal@DIGITS@_nan(self.0) {
            return f.write_str(&s);
        }
        f.write_str(&crate::generated::@MOD@::@PORT@(@ARG@))
    }
}
`, "SELF", w.selfType, "DIGITS", w.digits, "MOD", displayOp.module, "PORT", displayOp.port, "ARG", w.selfArg("self")))
	}
	if needFromStr {
		b.WriteString(tmpl(`
impl FromStr for @SELF@ {
    type Err = ParseDecimalError;
    fn from_str(s: &str) -> Result<@SELF@, ParseDecimalError> {
        @SELF@::parse(s)
    }
}
`, "SELF", w.selfType))
	}
	if emitFromI32Exact {
		emitFromI32ExactImpl(&b, fromI32ExactOp, w)
	}
	if emitFromU32Exact {
		emitFromU32ExactImpl(&b, fromU32ExactOp, w)
	}
	if emitFromI64Exact {
		emitFromI64ExactImpl(&b, fromI64ExactOp, w)
	}
	if emitFromU64Exact {
		emitFromU64ExactImpl(&b, fromU64ExactOp, w)
	}

	// The Decimal<w> struct no longer derives PartialEq/Eq/Hash (bit-equality
	// is wrong for an IEEE decimal float: it would make NaN == NaN and -0 !=
	// +0), so the idiomatic quiet == / quiet ordering must be emitted here
	// instead. This is required, not optional: without it the type would have
	// no PartialEq at all. Reject unresolved input if the three quiet comparison surfaces
	// the traits delegate to were not emitted.
	if quietEqMethod == "" || quietLtMethod == "" || quietGtMethod == "" {
		return "", fmt.Errorf("apiemit: %s dropped its PartialEq/Eq/Hash derive but the quiet comparison surfaces the trait impls need are missing (quiet_eq=%q quiet_lt=%q quiet_gt=%q); the QuietEqual/QuietLess/QuietGreater census symbols must be emitted", w.selfType, quietEqMethod, quietLtMethod, quietGtMethod)
	}
	emitPartialEqOrd(&b, quietEqMethod, quietLtMethod, quietGtMethod, w)

	if emitParse || emitParseWithFlags {
		if isNaNOp.module == "" {
			return "", fmt.Errorf("apiemit: %s emits parse/parse_with_flags but no %s predicate row was found to build result_is_nan", w.selfType, bidIsNaN)
		}
		emitResultIsNaNFn(&b, w, isNaNOp)
		emitInvalidBidStringInputFn(&b)
	}
	if emitParseRaw {
		if w.is128 {
			emitParseDecimalNaN128Fn(&b)
		} else {
			emitParseDecimalNaNFn(&b, w)
		}
	}
	if emitDisplay {
		if w.is128 {
			emitFormatDecimalNaN128Fn(&b)
		} else {
			emitFormatDecimalNaNFn(&b, w)
		}
	}

	return b.String(), nil
}

// goMethodName returns the PascalCase Go method name for a Rust wrapper method
// name, used only in the emitted comment that cites the Go counterpart.
func goMethodName(method string) string {
	if method == "" {
		return method
	}
	return strings.ToUpper(method[:1]) + method[1:]
}

// methodVerb returns a human-readable verb for a wrapper method name, used only
// in the emitted doc comment. An unregistered method fails the emit rather than
// falling back to a vague verb that would ship in the public rustdoc.
func methodVerb(method string) (string, error) {
	switch method {
	case "add":
		return "Adds", nil
	case "sub":
		return "Subtracts", nil
	case "mul":
		return "Multiplies", nil
	case "div":
		return "Divides", nil
	default:
		return "", fmt.Errorf("apiemit: no doc verb registered for binary method %q", method)
	}
}

// binaryFlagsClause returns the subject-and-object clause of the emitted doc
// comment for a binary flag-returning wrapper. quantize takes its rhs as an
// exponent template rather than as an addend, so a shared "two decimals"
// phrasing would describe it wrongly; an unknown method fails the emit rather
// than falling back to a vague clause that would ship in the public rustdoc.
func binaryFlagsClause(base string) (string, error) {
	switch base {
	case "add", "sub", "mul", "div":
		verb, err := methodVerb(base)
		if err != nil {
			return "", err
		}
		return verb + " two decimals", nil
	case "quantize":
		return "Rescales self to the quantum (target exponent) of rhs", nil
	default:
		return "", fmt.Errorf("apiemit: no doc clause registered for binary_with_flags method %q", base)
	}
}
