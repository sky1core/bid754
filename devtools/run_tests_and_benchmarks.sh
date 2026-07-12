#!/usr/bin/env bash
# Compatibility test+benchmark entrypoint. Test verification delegates to
# `make verify-all`; benchmark execution delegates to `make bench` when native
# prerequisites are present.

set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
results_dir="$repo_root/test_results"
timestamp=$(date +"%Y%m%d_%H%M%S")
test_output="$results_dir/full_verify_${timestamp}.txt"
bench_output="$results_dir/benchmark_results_${timestamp}.txt"
summary_output="$results_dir/performance_summary_${timestamp}.txt"

mkdir -p "$results_dir"

{
  echo "=== bid754 full verification ==="
  echo "Start: $(date)"
  echo "System: $(uname -a)"
  echo "Go: $(go version)"
  echo
  echo "This compatibility script delegates test verification to: make verify-all"
  echo

  cd "$repo_root"
  make verify-all

  echo
  echo "Completed: $(date)"
} 2>&1 | tee "$test_output"

cp "$test_output" "$results_dir/latest_full_verify_results.txt"
cp "$test_output" "$results_dir/latest_test_results.txt"

{
  echo "=== bid754 benchmark results ==="
  echo "Start: $(date)"
  echo "System: $(uname -a)"
  echo "Go: $(go version)"
  echo

  cd "$repo_root"
  if [ -f .env.sh ]; then
    make bench
  else
    echo "SKIP: .env.sh is missing, so the full C/Go/Rust benchmark matrix is unavailable"
  fi

  echo
  echo "Completed: $(date)"
} 2>&1 | tee "$bench_output"

cp "$bench_output" "$results_dir/latest_benchmark_results.txt"

{
  echo "=== bid754 performance summary ==="
  echo "Generated: $(date)"
  echo

  if grep -Eq '^(Benchmark|bid(32|64|128)/(add|mul|div|parse|to_string))' "$bench_output"; then
    echo "=== C/Go benchmark matrix ==="
    grep -E "Benchmark(IntelCBID|AlignedBID|FairBID).*-([0-9]+|[0-9]+\\s)" "$bench_output" || true
    echo
    echo "=== Rust Criterion matrix ==="
    grep -E "^(bid32|bid64|bid128)/(add|mul|div|parse|to_string)" "$bench_output" || true
  else
    echo "No benchmark measurements were produced."
  fi
} >"$summary_output"

cp "$summary_output" "$results_dir/latest_performance_summary.txt"

echo "Full verification results: $test_output"
echo "Benchmark results: $bench_output"
echo "Performance summary: $summary_output"
