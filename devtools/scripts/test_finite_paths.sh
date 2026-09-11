#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
target=${CARGO_TARGET_DIR:-"$root/bid754-rs/target"}
mkdir -p "$target" "$root/test_results"
target=$(cd "$target" && pwd)
cargo build --locked --manifest-path "$root/bid754-rs/Cargo.toml" --features verification --example finite_probe --target-dir "$target"
export BID754_FINITE_RUST="$target/debug/examples/finite_probe"
run_dir=$(mktemp -d "$root/test_results/finite-paths.XXXXXX")
export BID754_FINITE_FAILURES="$run_dir/finding.json"
cd "$root/bid754-go"
go test -count=1 -v -run '^(TestFiniteArithmeticPaths|TestFinitePathCoverageContract|TestFinitePathInputContract|TestFiniteRustDeadline|TestBigDecimalOracle|TestBigDecimalFuzzOracleContract)$' .
