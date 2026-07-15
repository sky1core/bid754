// Shared benchmark-row declarations.
//
// Each macro passes one direct generated-Rust call expression to its emitter.
// The Criterion emitter places that expression directly inside `b.iter`; the
// native wiring test evaluates the same expression once. Keep expectations and
// Intel C oracle dispatch out of this file so a wiring error cannot skew both
// the observed and expected legs in the same way.

macro_rules! bid32_benchmark_rows {
    ($emit:ident, $state:ident, $prepared:ident) => {
        $emit!(
            $state,
            Bid32Add,
            "bid32",
            "add",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_add_with_flags(black_box($prepared.x32), black_box($prepared.y32), 0)
        );
        $emit!(
            $state,
            Bid32Mul,
            "bid32",
            "mul",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_mul_with_flags(black_box($prepared.x32), black_box($prepared.y32), 0)
        );
        $emit!(
            $state,
            Bid32Sub,
            "bid32",
            "sub",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_sub_with_flags(black_box($prepared.x32), black_box($prepared.y32), 0)
        );
        $emit!(
            $state,
            Bid32Div,
            "bid32",
            "div",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_div_with_flags(black_box($prepared.x32), black_box($prepared.y32), 0)
        );
        $emit!(
            $state,
            Bid32Fma,
            "bid32",
            "fma",
            [D32, D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_fma(
                black_box($prepared.x32),
                black_box($prepared.y32),
                black_box($prepared.z32),
                0,
            )
        );
        $emit!(
            $state,
            Bid32Sqrt,
            "bid32",
            "sqrt",
            [D32],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_sqrt(black_box($prepared.x32), 0)
        );
        $emit!(
            $state,
            Bid32Remainder,
            "bid32",
            "remainder",
            [D32, D32],
            D32,
            NotApplicable,
            TupleFlags,
            bid32_rem(black_box($prepared.y32), black_box($prepared.x32))
        );
        $emit!(
            $state,
            Bid32Fmod,
            "bid32",
            "fmod",
            [D32, D32],
            D32,
            NotApplicable,
            TupleFlags,
            bid32_fmod(black_box($prepared.y32), black_box($prepared.x32))
        );
        $emit!(
            $state,
            Bid32Quantize,
            "bid32",
            "quantize",
            [D32, D32],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_quantize(black_box($prepared.x32), black_box($prepared.integer32), 0,)
        );
        $emit!(
            $state,
            Bid32Scaleb,
            "bid32",
            "scaleb",
            [D32, I64],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_scalbn_with_flags(
                black_box($prepared.x32),
                black_box($prepared.scale_exponent),
                0,
            )
        );
        $emit!(
            $state,
            Bid32QuietEqual,
            "bid32",
            "quiet_equal",
            [D32, D32],
            Predicate,
            NotApplicable,
            TupleFlags,
            bid32_quiet_equal(black_box($prepared.x32), black_box($prepared.y32))
        );
        $emit!(
            $state,
            Bid32Minnum,
            "bid32",
            "minnum",
            [D32, D32],
            D32,
            NotApplicable,
            TupleFlags,
            bid32_min_num_with_flags(black_box($prepared.x32), black_box($prepared.y32))
        );
        $emit!(
            $state,
            Bid32Maxnum,
            "bid32",
            "maxnum",
            [D32, D32],
            D32,
            NotApplicable,
            TupleFlags,
            bid32_max_num_with_flags(black_box($prepared.x32), black_box($prepared.y32))
        );
        $emit!(
            $state,
            Bid32FromInt64,
            "bid32",
            "from_int64",
            [I64],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_from_int64(black_box($prepared.integer_operand), 0)
        );
        $emit!(
            $state,
            Bid32ToInt64,
            "bid32",
            "to_int64",
            [D32],
            I64,
            FixedNearestEven,
            TupleFlags,
            bid32_to_int64_rnint(black_box($prepared.integer32))
        );
        $emit!(
            $state,
            Bid32ToDecimal64,
            "bid32",
            "to_decimal64",
            [D32],
            D64,
            NotApplicable,
            TupleFlags,
            bid32_to_bid64(black_box($prepared.x32))
        );
        $emit!(
            $state,
            Bid32ToDecimal128,
            "bid32",
            "to_decimal128",
            [D32],
            D128,
            NotApplicable,
            TupleFlags,
            bid32_to_bid128(black_box($prepared.x32))
        );
        $emit!(
            $state,
            Bid32Parse,
            "bid32",
            "parse",
            [Text],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid32_from_string_raw($prepared.decimal32_x_text.as_str(), 0)
        );
        $emit!(
            $state,
            Bid32ToString,
            "bid32",
            "to_string",
            [D32],
            Text,
            NotApplicable,
            NoFlags,
            bid32_to_string_raw($prepared.x32)
        );
    };
}

macro_rules! bid64_benchmark_rows {
    ($emit:ident, $state:ident, $prepared:ident) => {
        $emit!(
            $state,
            Bid64Add,
            "bid64",
            "add",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64_add_with_flags(black_box($prepared.x64), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid64Mul,
            "bid64",
            "mul",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64_mul_with_flags(black_box($prepared.x64), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid64Sub,
            "bid64",
            "sub",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid754::generated::add64::bid64_sub_with_flags(
                black_box($prepared.x64),
                black_box($prepared.y64),
                0,
            )
        );
        $emit!(
            $state,
            Bid64Div,
            "bid64",
            "div",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64_div_with_flags(black_box($prepared.x64), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid64Fma,
            "bid64",
            "fma",
            [D64, D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64_fma(
                black_box($prepared.x64),
                black_box($prepared.y64),
                black_box($prepared.z64),
                0,
            )
        );
        $emit!(
            $state,
            Bid64Sqrt,
            "bid64",
            "sqrt",
            [D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64_sqrt(black_box($prepared.x64), 0)
        );
        $emit!(
            $state,
            Bid64Remainder,
            "bid64",
            "remainder",
            [D64, D64],
            D64,
            NotApplicable,
            TupleFlags,
            bid64_rem(black_box($prepared.y64), black_box($prepared.x64))
        );
        $emit!(
            $state,
            Bid64Fmod,
            "bid64",
            "fmod",
            [D64, D64],
            D64,
            NotApplicable,
            TupleFlags,
            bid64_fmod(black_box($prepared.y64), black_box($prepared.x64))
        );
        $emit!(
            $state,
            Bid64Quantize,
            "bid64",
            "quantize",
            [D64, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64_quantize(black_box($prepared.x64), black_box($prepared.integer64), 0,)
        );
        $emit!(
            $state,
            Bid64Scaleb,
            "bid64",
            "scaleb",
            [D64, I64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64_scalbn(
                black_box($prepared.x64),
                black_box($prepared.scale_exponent),
                0,
            )
        );
        $emit!(
            $state,
            Bid64QuietEqual,
            "bid64",
            "quiet_equal",
            [D64, D64],
            Predicate,
            NotApplicable,
            TupleFlags,
            bid64_quiet_equal(black_box($prepared.x64), black_box($prepared.y64))
        );
        $emit!(
            $state,
            Bid64Minnum,
            "bid64",
            "minnum",
            [D64, D64],
            D64,
            NotApplicable,
            TupleFlags,
            bid64_min_num(black_box($prepared.x64), black_box($prepared.y64))
        );
        $emit!(
            $state,
            Bid64Maxnum,
            "bid64",
            "maxnum",
            [D64, D64],
            D64,
            NotApplicable,
            TupleFlags,
            bid64_max_num(black_box($prepared.x64), black_box($prepared.y64))
        );
        $emit!(
            $state,
            Bid64FromInt64,
            "bid64",
            "from_int64",
            [I64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64_from_int64(black_box($prepared.integer_operand), 0)
        );
        $emit!(
            $state,
            Bid64ToInt64,
            "bid64",
            "to_int64",
            [D64],
            I64,
            FixedNearestEven,
            TupleFlags,
            bid64_to_int64_rnint(black_box($prepared.integer64))
        );
        $emit!(
            $state,
            Bid64ToDecimal32,
            "bid64",
            "to_decimal32",
            [D64],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid64_to_bid32(black_box($prepared.x64), 0)
        );
        $emit!(
            $state,
            Bid64ToDecimal128,
            "bid64",
            "to_decimal128",
            [D64],
            D128,
            NotApplicable,
            TupleFlags,
            bid64_to_bid128(black_box($prepared.x64))
        );
        $emit!(
            $state,
            Bid64Parse,
            "bid64",
            "parse",
            [Text],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64_from_string_raw($prepared.decimal64_x_text.as_str(), 0)
        );
        $emit!(
            $state,
            Bid64ToString,
            "bid64",
            "to_string",
            [D64],
            Text,
            NotApplicable,
            NoFlags,
            bid64_to_string($prepared.x64)
        );
    };
}

macro_rules! bid128_benchmark_rows {
    ($emit:ident, $state:ident, $prepared:ident) => {
        $emit!(
            $state,
            Bid128Add,
            "bid128",
            "add",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            OutFlags,
            {
                {
                    let mut flags = 0u32;
                    let got = bid128_add(
                        black_box($prepared.x128),
                        black_box($prepared.y128),
                        0,
                        &mut flags,
                    );
                    (got, flags)
                }
            }
        );
        $emit!(
            $state,
            Bid128Mul,
            "bid128",
            "mul",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128_mul(black_box($prepared.x128), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid128Sub,
            "bid128",
            "sub",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            OutFlags,
            {
                {
                    let mut flags = 0u32;
                    let got = bid128_sub(
                        black_box($prepared.x128),
                        black_box($prepared.y128),
                        0,
                        &mut flags,
                    );
                    (got, flags)
                }
            }
        );
        $emit!(
            $state,
            Bid128Div,
            "bid128",
            "div",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128_div(black_box($prepared.x128), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid128Fma,
            "bid128",
            "fma",
            [D128, D128, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128_fma(
                black_box($prepared.x128),
                black_box($prepared.y128),
                black_box($prepared.z128),
                0,
            )
        );
        $emit!(
            $state,
            Bid128Sqrt,
            "bid128",
            "sqrt",
            [D128],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128_sqrt(black_box($prepared.x128), 0)
        );
        $emit!(
            $state,
            Bid128Remainder,
            "bid128",
            "remainder",
            [D128, D128],
            D128,
            NotApplicable,
            TupleFlags,
            bid128_rem(black_box($prepared.y128), black_box($prepared.x128))
        );
        $emit!(
            $state,
            Bid128Fmod,
            "bid128",
            "fmod",
            [D128, D128],
            D128,
            NotApplicable,
            TupleFlags,
            bid128_fmod(black_box($prepared.y128), black_box($prepared.x128))
        );
        $emit!(
            $state,
            Bid128Quantize,
            "bid128",
            "quantize",
            [D128, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128_quantize(
                black_box($prepared.x128),
                black_box($prepared.integer128),
                0,
            )
        );
        $emit!(
            $state,
            Bid128Scaleb,
            "bid128",
            "scaleb",
            [D128, I64],
            D128,
            ExplicitNearestEven,
            OutFlags,
            {
                {
                    let mut flags = 0u32;
                    let got = bid128_scalbn(
                        black_box($prepared.x128),
                        black_box($prepared.scale_exponent),
                        0,
                        &mut flags,
                    );
                    (got, flags)
                }
            }
        );
        $emit!(
            $state,
            Bid128QuietEqual,
            "bid128",
            "quiet_equal",
            [D128, D128],
            Predicate,
            NotApplicable,
            TupleFlags,
            bid128_quiet_equal(black_box($prepared.x128), black_box($prepared.y128))
        );
        $emit!(
            $state,
            Bid128Minnum,
            "bid128",
            "minnum",
            [D128, D128],
            D128,
            NotApplicable,
            OutFlags,
            {
                {
                    let mut flags = 0u32;
                    let got = bid128_minnum(
                        black_box($prepared.x128),
                        black_box($prepared.y128),
                        &mut flags,
                    );
                    (got, flags)
                }
            }
        );
        $emit!(
            $state,
            Bid128Maxnum,
            "bid128",
            "maxnum",
            [D128, D128],
            D128,
            NotApplicable,
            OutFlags,
            {
                {
                    let mut flags = 0u32;
                    let got = bid128_maxnum(
                        black_box($prepared.x128),
                        black_box($prepared.y128),
                        &mut flags,
                    );
                    (got, flags)
                }
            }
        );
        $emit!(
            $state,
            Bid128FromInt64,
            "bid128",
            "from_int64",
            [I64],
            D128,
            NotApplicable,
            NoFlags,
            bid128_from_int64(black_box($prepared.integer_operand))
        );
        $emit!(
            $state,
            Bid128ToInt64,
            "bid128",
            "to_int64",
            [D128],
            I64,
            FixedNearestEven,
            TupleFlags,
            bid128_to_int64_rnint(black_box($prepared.integer128))
        );
        $emit!(
            $state,
            Bid128ToDecimal32,
            "bid128",
            "to_decimal32",
            [D128],
            D32,
            ExplicitNearestEven,
            TupleFlags,
            bid128_to_bid32(black_box($prepared.x128), 0)
        );
        $emit!(
            $state,
            Bid128ToDecimal64,
            "bid128",
            "to_decimal64",
            [D128],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid128_to_bid64(black_box($prepared.x128), 0)
        );
        $emit!(
            $state,
            Bid128Parse,
            "bid128",
            "parse",
            [Text],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128_from_string($prepared.decimal128_x_text.as_str(), 0)
        );
        $emit!(
            $state,
            Bid128ToString,
            "bid128",
            "to_string",
            [D128],
            Text,
            NotApplicable,
            NoFlags,
            bid128_to_string($prepared.x128)
        );
    };
}

macro_rules! bid64_mixed_benchmark_rows {
    ($emit:ident, $state:ident, $prepared:ident) => {
        $emit!(
            $state,
            Bid64MixedDqAdd,
            "bid64_mixed",
            "dq_add",
            [D64, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64dq_add(black_box($prepared.x64), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid64MixedQdAdd,
            "bid64_mixed",
            "qd_add",
            [D128, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64qd_add(black_box($prepared.x128), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid64MixedQqAdd,
            "bid64_mixed",
            "qq_add",
            [D128, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64qq_add(black_box($prepared.x128), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid64MixedDqSub,
            "bid64_mixed",
            "dq_sub",
            [D64, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64dq_sub(black_box($prepared.x64), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid64MixedQdSub,
            "bid64_mixed",
            "qd_sub",
            [D128, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64qd_sub(black_box($prepared.x128), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid64MixedQqSub,
            "bid64_mixed",
            "qq_sub",
            [D128, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64qq_sub(black_box($prepared.x128), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid64MixedDqMul,
            "bid64_mixed",
            "dq_mul",
            [D64, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64dq_mul(black_box($prepared.x64), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid64MixedQdMul,
            "bid64_mixed",
            "qd_mul",
            [D128, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64qd_mul(black_box($prepared.x128), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid64MixedQqMul,
            "bid64_mixed",
            "qq_mul",
            [D128, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64qq_mul(black_box($prepared.x128), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid64MixedDqDiv,
            "bid64_mixed",
            "dq_div",
            [D64, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64dq_div(black_box($prepared.x64), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid64MixedQdDiv,
            "bid64_mixed",
            "qd_div",
            [D128, D64],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64qd_div(black_box($prepared.x128), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid64MixedQqDiv,
            "bid64_mixed",
            "qq_div",
            [D128, D128],
            D64,
            ExplicitNearestEven,
            TupleFlags,
            bid64qq_div(black_box($prepared.x128), black_box($prepared.y128), 0)
        );
    };
}

macro_rules! bid128_mixed_benchmark_rows {
    ($emit:ident, $state:ident, $prepared:ident) => {
        $emit!(
            $state,
            Bid128MixedDdAdd,
            "bid128_mixed",
            "dd_add",
            [D64, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128dd_add(black_box($prepared.x64), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid128MixedDqAdd,
            "bid128_mixed",
            "dq_add",
            [D64, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128dq_add(black_box($prepared.x64), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid128MixedQdAdd,
            "bid128_mixed",
            "qd_add",
            [D128, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128qd_add(black_box($prepared.x128), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid128MixedDdSub,
            "bid128_mixed",
            "dd_sub",
            [D64, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128dd_sub(black_box($prepared.x64), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid128MixedDqSub,
            "bid128_mixed",
            "dq_sub",
            [D64, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128dq_sub(black_box($prepared.x64), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid128MixedQdSub,
            "bid128_mixed",
            "qd_sub",
            [D128, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128qd_sub(black_box($prepared.x128), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid128MixedDdMul,
            "bid128_mixed",
            "dd_mul",
            [D64, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128dd_mul(black_box($prepared.x64), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid128MixedDqMul,
            "bid128_mixed",
            "dq_mul",
            [D64, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128dq_mul(black_box($prepared.x64), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid128MixedQdMul,
            "bid128_mixed",
            "qd_mul",
            [D128, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128qd_mul(black_box($prepared.x128), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid128MixedDdDiv,
            "bid128_mixed",
            "dd_div",
            [D64, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128dd_div(black_box($prepared.x64), black_box($prepared.y64), 0)
        );
        $emit!(
            $state,
            Bid128MixedDqDiv,
            "bid128_mixed",
            "dq_div",
            [D64, D128],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128dq_div(black_box($prepared.x64), black_box($prepared.y128), 0)
        );
        $emit!(
            $state,
            Bid128MixedQdDiv,
            "bid128_mixed",
            "qd_div",
            [D128, D64],
            D128,
            ExplicitNearestEven,
            TupleFlags,
            bid128qd_div(black_box($prepared.x128), black_box($prepared.y64), 0)
        );
    };
}

macro_rules! benchmark_groups {
    ($emit:ident, $state:ident, $prepared:ident) => {
        $emit!($state, $prepared, "bid32", 19, bid32_benchmark_rows);
        $emit!($state, $prepared, "bid64", 19, bid64_benchmark_rows);
        $emit!($state, $prepared, "bid128", 19, bid128_benchmark_rows);
        $emit!(
            $state,
            $prepared,
            "bid64_mixed",
            12,
            bid64_mixed_benchmark_rows
        );
        $emit!(
            $state,
            $prepared,
            "bid128_mixed",
            12,
            bid128_mixed_benchmark_rows
        );
    };
}
