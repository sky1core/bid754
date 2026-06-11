#!/usr/bin/env bash
# Legacy test entrypoint. Keep this script as a thin wrapper so it cannot
# underreport the current repository verification boundary.

set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
results_dir="$repo_root/test_results"
timestamp=$(date +"%Y%m%d_%H%M%S")
output="$results_dir/full_audit_${timestamp}.txt"
latest="$results_dir/latest_full_audit_results.txt"

mkdir -p "$results_dir"

{
  echo "=== bid754 full audit ==="
  echo "Start: $(date)"
  echo "System: $(uname -a)"
  echo "Go: $(go version)"
  if command -v cargo >/dev/null 2>&1; then
    echo "Cargo: $(cargo --version)"
  else
    echo "Cargo: missing"
  fi
  echo
  echo "This legacy script delegates to: make full-audit"
  echo

  cd "$repo_root"
  make full-audit

  echo
  echo "Completed: $(date)"
} 2>&1 | tee "$output"

cp "$output" "$latest"
cp "$output" "$results_dir/latest_test_results.txt"
echo "Full audit results: $output"
echo "Latest full audit results: $latest"
