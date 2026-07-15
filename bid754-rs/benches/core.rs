use bid754::bid64_from_string_raw;
use bid754::generated::add64::bid64_add_with_flags;
use bid754::generated::bid128_add::{
    bid128_add, bid128_sub, bid128dd_add, bid128dd_sub, bid128dq_add, bid128dq_sub, bid128qd_add,
    bid128qd_sub, bid64dq_add, bid64dq_sub, bid64qd_add, bid64qd_sub, bid64qq_add, bid64qq_sub,
};
use bid754::generated::bid128_compare::bid128_quiet_equal;
use bid754::generated::bid128_conversions::{bid128_to_bid32, bid128_to_bid64};
use bid754::generated::bid128_div::{bid128_div, bid128dd_div, bid128dq_div, bid128qd_div};
use bid754::generated::bid128_fma::bid128_fma;
use bid754::generated::bid128_from_int::bid128_from_int64;
use bid754::generated::bid128_minmax::{bid128_maxnum, bid128_minnum};
use bid754::generated::bid128_misc::bid128_scalbn;
use bid754::generated::bid128_mul::{
    bid128_mul, bid128dd_mul, bid128dq_mul, bid128qd_mul, bid64dq_mul, bid64qd_mul, bid64qq_mul,
};
use bid754::generated::bid128_quantize::bid128_quantize;
use bid754::generated::bid128_rem::{bid128_fmod, bid128_rem};
use bid754::generated::bid128_sqrt::bid128_sqrt;
use bid754::generated::bid128_string::{bid128_from_string, bid128_to_string};
use bid754::generated::bid128_to_int::bid128_to_int64_rnint;
use bid754::generated::bid32_fma::bid32_fma;
use bid754::generated::bid32_misc::{bid32_fmod, bid32_to_bid128};
use bid754::generated::bid32_quantize::bid32_quantize;
use bid754::generated::bid32_rem::bid32_rem;
use bid754::generated::bid32_sqrt::bid32_sqrt;
use bid754::generated::bid32_status::{
    bid32_add_with_flags, bid32_div_with_flags, bid32_max_num_with_flags, bid32_min_num_with_flags,
    bid32_mul_with_flags, bid32_scalbn_with_flags, bid32_sub_with_flags,
};
use bid754::generated::bid32_string::{bid32_from_string_raw, bid32_to_string_raw};
use bid754::generated::bid32_to_bid64::bid32_to_bid64;
use bid754::generated::bid32_to_int::{bid32_from_int64, bid32_to_int64_rnint};
use bid754::generated::compare32::bid32_quiet_equal;
use bid754::generated::compare64::bid64_quiet_equal;
use bid754::generated::convert64::bid64_from_int64;
use bid754::generated::div64::{bid64_div_with_flags, bid64dq_div, bid64qd_div, bid64qq_div};
use bid754::generated::fma64::bid64_fma;
use bid754::generated::fmod64::bid64_fmod;
use bid754::generated::minmax64::{bid64_max_num, bid64_min_num};
use bid754::generated::mul64::bid64_mul_with_flags;
use bid754::generated::quantize64::bid64_quantize;
use bid754::generated::rem64::bid64_rem;
use bid754::generated::scalb64::bid64_scalbn;
use bid754::generated::sqrt64::bid64_sqrt;
use bid754::generated::string64::bid64_to_string;
use bid754::generated::to_bid12864::bid64_to_bid128;
use bid754::generated::to_bid3264::bid64_to_bid32;
use bid754::generated::to_int64_signed::bid64_to_int64_rnint;
use criterion::{black_box, criterion_group, criterion_main, Criterion};

include!("support/inputs.rs");
include!("support/rows.rs");

macro_rules! register_criterion_row {
    (
        $group:ident,
        $id:ident,
        $group_name:literal,
        $name:literal,
        $operands:tt,
        $result:ident,
        $rounding:ident,
        $status:ident,
        $call:expr
    ) => {
        $group.bench_function($name, |b| b.iter(|| black_box($call)));
    };
}

macro_rules! register_criterion_group {
    ($criterion:ident, $prepared:ident, $name:literal, $count:literal, $rows:ident) => {{
        let mut group = $criterion.benchmark_group($name);
        $rows!(register_criterion_row, group, $prepared);
        group.finish();
    }};
}

fn bench_all(c: &mut Criterion) {
    let prepared = prepare_benchmark_inputs();
    benchmark_groups!(register_criterion_group, c, prepared);
}

criterion_group!(benches, bench_all);
criterion_main!(benches);
