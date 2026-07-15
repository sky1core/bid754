#!/usr/bin/env bash
set -euo pipefail

# Verify the benchmark binary's actual Criterion registry rather than only the
# shared row declarations. This catches an omitted criterion_group entry, a
# group rename, or an outer-registration wiring error without measuring timings.

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$repo_root/bid754-rs"

echo "==> Rust Criterion benchmark registry exact-set verification"

if ! registry_output=$(
  CARGO_TERM_COLOR=never cargo bench --locked --bench core -- \
    --list --format terse --color never 2>&1
); then
  echo "ERROR: cargo bench failed while listing the Rust Criterion benchmark registry" >&2
  printf '%s\n' "$registry_output" >&2
  exit 1
fi
actual_names=$(
  printf '%s\n' "$registry_output" |
    sed -n 's/: benchmark$//p'
)

if [ -z "$actual_names" ]; then
  echo "ERROR: Criterion listed no registered benchmarks" >&2
  printf '%s\n' "$registry_output" >&2
  exit 1
fi

expected_names=$(
  common_ops=(
    add mul sub div fma sqrt remainder fmod quantize scaleb quiet_equal
    minnum maxnum from_int64 to_int64 parse to_string
  )
  for width in bid32 bid64 bid128; do
    for op in "${common_ops[@]}"; do
      printf '%s/%s\n' "$width" "$op"
    done
  done
  printf '%s\n' \
    bid32/to_decimal64 bid32/to_decimal128 \
    bid64/to_decimal32 bid64/to_decimal128 \
    bid128/to_decimal32 bid128/to_decimal64

  for op in add sub mul div; do
    for operands in dq qd qq; do
      printf 'bid64_mixed/%s_%s\n' "$operands" "$op"
    done
    for operands in dd dq qd; do
      printf 'bid128_mixed/%s_%s\n' "$operands" "$op"
    done
  done
)

expected_count=$(printf '%s\n' "$expected_names" | grep -c .)
actual_count=$(printf '%s\n' "$actual_names" | grep -c .)

if [ "$expected_count" -ne 81 ]; then
  echo "ERROR: registry gate expectation is internally inconsistent: got $expected_count names, want 81" >&2
  exit 1
fi

if ! diff -u \
  <(printf '%s\n' "$expected_names" | LC_ALL=C sort) \
  <(printf '%s\n' "$actual_names" | LC_ALL=C sort); then
  echo "ERROR: Rust Criterion registry differs from the required 81-row exact set (actual count: $actual_count)" >&2
  exit 1
fi

echo "Rust Criterion benchmark registry verified: $actual_count exact rows."
