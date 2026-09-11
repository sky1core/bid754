#!/usr/bin/env bash
set -euo pipefail
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
java_bin=$(command -v java)
javac_bin=$(command -v javac)
"$java_bin" -version
"$javac_bin" -version
mkdir -p "$root/test_results"
run_dir=$(mktemp -d "$root/test_results/bigdecimal.XXXXXX")
"$javac_bin" --release 17 -d "$run_dir/classes" "$root/devtools/java/BigDecimalProbe.java"
export BID754_BIGDECIMAL_JAVA="$java_bin"
export BID754_BIGDECIMAL_CLASSES="$run_dir/classes"
bash "$root/devtools/scripts/test_finite_paths.sh"
