#!/usr/bin/env bash
set -euo pipefail
usage() {
    echo "Usage: $0 [--duration <positive integer>[s|m|h]] [--parallel <positive integer>]"
    echo "       $0 --replay <corpus-id|seed#N>"
    echo "Defaults: --duration 60s --parallel 2; maximum duration 24h, parallelism 256."
}
duration=60s
parallel=2
replay=
search_options=0
while (($#)); do
    case "$1" in
        --duration|--parallel|--replay)
            if (($# < 2)); then usage >&2; exit 2; fi
            case "$1" in
                --duration) duration=$2; search_options=1 ;;
                --parallel) parallel=$2; search_options=1 ;;
                --replay) if [[ -z "$2" ]]; then usage >&2; exit 2; fi; replay=$2 ;;
            esac
            shift 2
            ;;
        --help|-h) usage; exit 0 ;;
        *) usage >&2; exit 2 ;;
    esac
done
if [[ -n "$replay" ]] && { ((search_options)) || [[ ! "$replay" =~ ^([0-9a-f]{1,64}|seed#[0-9]{1,6})$ ]]; }; then
    echo "Replay requires a corpus ID or seed#N, without search options." >&2
    exit 2
fi
if [[ ! "$duration" =~ ^([1-9][0-9]{0,5})(s|m|h)$ ]]; then
    echo "Invalid duration: use positive whole seconds, minutes or hours (e.g. 60s)." >&2
    exit 2
fi
seconds=${BASH_REMATCH[1]}
case "${BASH_REMATCH[2]}" in
    m) seconds=$((seconds * 60)) ;;
    h) seconds=$((seconds * 3600)) ;;
esac
if ((seconds > 86400)) || [[ ! "$parallel" =~ ^[1-9][0-9]{0,2}$ ]] || ((parallel > 256)); then
    echo "Invalid budget: duration must be <=24h and parallelism in [1,256]." >&2
    exit 2
fi
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
if [[ -n "$replay" && "$replay" != seed#* && ! -f "$root/bid754-go/testdata/fuzz/FuzzFiniteArithmeticBigDecimal/$replay" ]]; then
    echo "Missing BigDecimal fuzz corpus entry: $replay" >&2
    exit 2
fi
java_bin=$(command -v java)
javac_bin=$(command -v javac)
"$java_bin" -version
"$javac_bin" -version
target=${CARGO_TARGET_DIR:-"$root/bid754-rs/target"}
mkdir -p "$target" "$root/test_results"
target=$(cd "$target" && pwd)
cargo build --locked --manifest-path "$root/bid754-rs/Cargo.toml" --features verification --example finite_probe --target-dir "$target"
run_dir=$(mktemp -d "$root/test_results/bigdecimal-fuzz.XXXXXX")
"$javac_bin" --release 17 -d "$run_dir/classes" "$root/devtools/java/BigDecimalProbe.java"
export BID754_BIGDECIMAL_JAVA="$java_bin"
export BID754_BIGDECIMAL_CLASSES="$run_dir/classes"
export BID754_FINITE_RUST="$target/debug/examples/finite_probe"
export BID754_BIGDECIMAL_FUZZ_FINDINGS="$run_dir"
echo "Oracle classes: $BID754_BIGDECIMAL_CLASSES"
echo "Coverage guidance: Go instrumentation only; Java and Rust are external comparison processes."
cd "$root/bid754-go"
if [[ -n "$replay" ]]; then
    go test -count=1 -v "-run=^FuzzFiniteArithmeticBigDecimal/$replay\$" -timeout=30s . 2>&1 | tee "$run_dir/replay.log"
    replay_passed=0
    while IFS= read -r line; do
        if [[ "$line" == "    --- PASS: FuzzFiniteArithmeticBigDecimal/$replay ("* ]]; then
            replay_passed=1
        fi
    done < "$run_dir/replay.log"
    if (( ! replay_passed )); then
        echo "Requested fuzz input did not execute successfully: $replay" >&2
        exit 1
    fi
else
    go test -count=1 -v '-run=^$' '-fuzz=^FuzzFiniteArithmeticBigDecimal$' \
        "-fuzztime=$duration" "-parallel=$parallel" "-timeout=$((seconds + 180))s" . 2>&1 | tee "$run_dir/fuzz.log"
    cd "$root/devtools"
    go run ./cmd/verifyplan evidence --root .. --gate bigdecimal-fuzz --log "$run_dir/fuzz.log"
fi
