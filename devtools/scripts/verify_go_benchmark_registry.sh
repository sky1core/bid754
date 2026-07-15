#!/usr/bin/env bash
set -euo pipefail

# Execute the real Go benchmark binaries once per registered subbenchmark and
# exact-compare their normalized names with an independent closed-world list.
# The one-iteration output is used only as a registry signal; timings are
# discarded and must not be interpreted as measurements.

usage() {
    echo "usage: $0 <portable|native|all>" >&2
    exit 2
}

[ $# -eq 1 ] || usage
mode=$1
case "$mode" in
    portable|native|all) ;;
    *) usage ;;
esac

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)

normalize_benchmark_names() {
    awk '$1 ~ /^Benchmark/ {
        name = $1
        sub(/-[0-9]+$/, "", name)
        print name
    }'
}

report_execution_failure() {
    local label=$1 registry_output=$2
    echo "ERROR: $label benchmark registry execution failed" >&2
    # Never report the one-iteration timing rows. Preserve build, linker,
    # panic, and other diagnostic output needed to understand the failure.
    printf '%s\n' "$registry_output" | awk '$1 !~ /^Benchmark/' >&2
}

compare_registry() {
    local label=$1 wanted_count=$2 expected_names=$3 actual_names=$4
    local expected_count actual_count

    expected_count=$(printf '%s\n' "$expected_names" | awk 'NF { count++ } END { print count + 0 }')
    actual_count=$(printf '%s\n' "$actual_names" | awk 'NF { count++ } END { print count + 0 }')
    if [ "$expected_count" -ne "$wanted_count" ]; then
        echo "ERROR: $label expectation is internally inconsistent: got $expected_count names, want $wanted_count" >&2
        exit 1
    fi
    if [ "$actual_count" -eq 0 ]; then
        echo "ERROR: $label benchmark binary registered no matching subbenchmarks" >&2
        exit 1
    fi

    if ! diff -u \
        <(printf '%s\n' "$expected_names" | LC_ALL=C sort) \
        <(printf '%s\n' "$actual_names" | LC_ALL=C sort); then
        echo "ERROR: $label benchmark registry differs from the required $wanted_count-row exact set (actual count: $actual_count)" >&2
        exit 1
    fi

    echo "$label benchmark registry verified: $actual_count exact rows."
}

emit_rows() {
    local prefix=$1
    shift
    local op
    for op in "$@"; do
        printf '%s/%s\n' "$prefix" "$op"
    done
}

emit_mixed_rows() {
    local prefix=$1
    emit_rows "$prefix" \
        dq_add qd_add qq_add \
        dq_sub qd_sub qq_sub \
        dq_mul qd_mul qq_mul \
        dq_div qd_div qq_div
}

emit_mixed128_rows() {
    local prefix=$1
    emit_rows "$prefix" \
        dd_add dq_add qd_add \
        dd_sub dq_sub qd_sub \
        dd_mul dq_mul qd_mul \
        dd_div dq_div qd_div
}

emit_public_width_rows() {
    local prefix=$1 conversion1=$2 conversion2=$3
    emit_rows "$prefix" \
        add mul sub div \
        add_with_flags mul_with_flags sub_with_flags div_with_flags \
        fma sqrt remainder fmod quantize scaleb quiet_equal minnum maxnum \
        from_int64 to_int64 "$conversion1" "$conversion2" parse to_string
}

emit_c_width_rows() {
    local prefix=$1 conversion1=$2 conversion2=$3
    emit_rows "$prefix" \
        add mul sub div fma sqrt remainder fmod quantize scaleb quiet_equal \
        minnum maxnum from_int64 to_int64 "$conversion1" "$conversion2" \
        parse to_string
}

emit_direct_bid32_rows() {
    emit_rows BenchmarkFairBID32 \
        add mul sub div add_pure mul_pure sub_pure div_pure \
        fma sqrt remainder fmod quantize scaleb quiet_equal minnum maxnum \
        from_int64 to_int64 to_decimal64 to_decimal128 parse to_string
}

portable_expected_names() {
    emit_direct_bid32_rows
    emit_c_width_rows BenchmarkFairBID64 to_decimal32 to_decimal128
    emit_c_width_rows BenchmarkFairBID128 to_decimal32 to_decimal64
    emit_mixed_rows BenchmarkFairMixedBID64
    emit_mixed128_rows BenchmarkFairMixedBID128
}

native_expected_names() {
    emit_public_width_rows BenchmarkAlignedBID32 to_decimal64 to_decimal128
    emit_public_width_rows BenchmarkAlignedBID64 to_decimal32 to_decimal128
    emit_public_width_rows BenchmarkAlignedBID128 to_decimal32 to_decimal64
    emit_mixed_rows BenchmarkAlignedMixedBID64
    emit_mixed128_rows BenchmarkAlignedMixedBID128

    emit_c_width_rows BenchmarkIntelCBID32 to_decimal64 to_decimal128
    emit_c_width_rows BenchmarkIntelCBID64 to_decimal32 to_decimal128
    emit_c_width_rows BenchmarkIntelCBID128 to_decimal32 to_decimal64
    emit_mixed_rows BenchmarkIntelCMixedBID64
    emit_mixed128_rows BenchmarkIntelCMixedBID128
}

verify_portable_registry() {
    local registry_output actual_names expected_names

    echo "==> Go mechanical-port benchmark registry exact-set verification"
    if ! registry_output=$({
        cd "$repo_root/bid754-go"
        GOCACHE=${GOCACHE:-/tmp/go-cache} go test -count=1 \
            -run='^$' \
            -bench='^BenchmarkFair(BID|MixedBID)' \
            -benchtime=1x \
            -timeout=300s \
            ./internal/bidgo
    } 2>&1); then
        report_execution_failure "Go mechanical-port" "$registry_output"
        exit 1
    fi

    actual_names=$(printf '%s\n' "$registry_output" | normalize_benchmark_names)
    expected_names=$(portable_expected_names)
    compare_registry "Go mechanical-port" 85 "$expected_names" "$actual_names"
}

verify_native_registry() {
    local registry_output actual_names expected_names

    if [ ! -f "$repo_root/.env.sh" ]; then
        echo "ERROR: native Go benchmark registry verification requires $repo_root/.env.sh" >&2
        exit 1
    fi

    echo "==> Go public API + Intel C benchmark registry exact-set verification"
    if ! registry_output=$({
        cd "$repo_root"
        set +u
        source ./.env.sh
        set -u
        cd bid754-go
        GOCACHE=${GOCACHE:-/tmp/go-cache} go test -count=1 \
            -tags bid754_native \
            -run='^$' \
            -bench='^Benchmark(Aligned|IntelC)(BID|MixedBID)' \
            -benchtime=1x \
            -timeout=300s \
            .
    } 2>&1); then
        report_execution_failure "Go public API + Intel C" "$registry_output"
        exit 1
    fi

    actual_names=$(printf '%s\n' "$registry_output" | normalize_benchmark_names)
    expected_names=$(native_expected_names)
    compare_registry "Go public API + Intel C" 174 "$expected_names" "$actual_names"
}

case "$mode" in
    portable)
        verify_portable_registry
        ;;
    native)
        verify_native_registry
        ;;
    all)
        verify_portable_registry
        verify_native_registry
        ;;
esac
