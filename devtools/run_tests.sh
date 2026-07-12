#!/usr/bin/env bash
# Compatibility test entrypoint. Keep this script as a thin wrapper so it cannot
# underreport the current repository verification boundary.

set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
results_dir="$repo_root/test_results"
timestamp=$(date +"%Y%m%d_%H%M%S")
output="$results_dir/full_verify_${timestamp}.txt"
latest="$results_dir/latest_full_verify_results.txt"

mkdir -p "$results_dir"

{
  echo "=== bid754 full verification ==="
  echo "Start: $(date)"
  echo "System: $(uname -a)"
  echo "Go: $(go version)"
  if command -v cargo >/dev/null 2>&1; then
    echo "Cargo: $(cargo --version)"
  else
    echo "Cargo: missing"
  fi
  echo
  echo "This compatibility script delegates to: make verify-all"
  echo

  cd "$repo_root"
  make verify-all

  echo
  echo "Completed: $(date)"
} 2>&1 | tee "$output"

cp "$output" "$latest"
cp "$output" "$results_dir/latest_test_results.txt"
echo "Full verification results: $output"
echo "Latest full verification results: $latest"
