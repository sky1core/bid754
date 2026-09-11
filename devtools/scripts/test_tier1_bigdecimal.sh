#!/usr/bin/env bash
set -euo pipefail
usage() {
    echo "Usage: $0 [--fuzz <positive integer>[s|m|h] [--parallel <1..256>]]"
    echo "       $0 --replay <finding.json> | --corpus <corpus-id|seed#N>"
}
fuzz=
parallel=2
parallel_set=0
replay=
corpus=
while (($#)); do
    case "$1" in
        --fuzz|--parallel|--replay|--corpus)
            if (($# < 2)) || [[ -z "$2" ]]; then usage >&2; exit 2; fi
            case "$1" in
                --fuzz) fuzz=$2 ;;
                --parallel) parallel=$2; parallel_set=1 ;;
                --replay) replay=$2 ;;
                --corpus) corpus=$2 ;;
            esac
            shift 2 ;;
        --help|-h) usage; exit 0 ;;
        *) usage >&2; exit 2 ;;
    esac
done
modes=0
for selection in "$fuzz" "$replay" "$corpus"; do
    if [[ -n "$selection" ]]; then modes=$((modes+1)); fi
done
if ((modes > 1)) || { ((parallel_set)) && [[ -z "$fuzz" ]]; }; then usage >&2; exit 2; fi
if [[ ! "$parallel" =~ ^[1-9][0-9]{0,2}$ ]] || ((parallel > 256)); then usage >&2; exit 2; fi
seconds=0
if [[ -n "$fuzz" ]]; then
    if [[ ! "$fuzz" =~ ^([1-9][0-9]{0,5})(s|m|h)$ ]]; then usage >&2; exit 2; fi
    seconds=${BASH_REMATCH[1]}
    case "${BASH_REMATCH[2]}" in m) seconds=$((seconds*60));; h) seconds=$((seconds*3600));; esac
    if ((seconds>86400)); then usage >&2; exit 2; fi
fi
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
if [[ -n "$replay" ]]; then
    if [[ ! -f "$replay" ]]; then echo "Missing finding: $replay" >&2; exit 2; fi
    replay=$(cd "$(dirname "$replay")" && pwd)/$(basename "$replay")
fi
if [[ -n "$corpus" ]]; then
    if [[ ! "$corpus" =~ ^([0-9a-f]{1,64}|seed#[0-9]{1,6})$ ]]; then usage >&2; exit 2; fi
    if [[ "$corpus" != seed#* && ! -f "$root/bid754-go/testdata/fuzz/FuzzTier1BigDecimal/$corpus" ]]; then
        echo "Missing Tier1 fuzz corpus: $corpus" >&2; exit 2
    fi
fi
java_bin=$(command -v java)
javac_bin=$(command -v javac)
"$java_bin" -version
"$javac_bin" -version
target=${CARGO_TARGET_DIR:-"$root/bid754-rs/target"}
mkdir -p "$target" "$root/test_results"
target=$(cd "$target" && pwd)
cargo build --locked --manifest-path "$root/bid754-rs/Cargo.toml" --features verification --example tier1_probe --target-dir "$target"
run_dir=$(mktemp -d "$root/test_results/tier1-bigdecimal.XXXXXX")
"$javac_bin" --release 17 -d "$run_dir/classes" "$root/devtools/java/Tier1BigDecimalProbe.java"
export BID754_BIGDECIMAL_JAVA="$java_bin"
export BID754_BIGDECIMAL_CLASSES="$run_dir/classes"
export BID754_TIER1_RUST="$target/debug/examples/tier1_probe"
export BID754_TIER1_FINDINGS="$run_dir"
unset BID754_TIER1_REPLAY
cd "$root/bid754-go"
if [[ -n "$fuzz" ]]; then
    go test -count=1 -v '-run=^$' '-fuzz=^FuzzTier1BigDecimal$' "-fuzztime=$fuzz" "-parallel=$parallel" "-timeout=$((seconds+180))s" . 2>&1 | tee "$run_dir/fuzz.log"
    cd "$root/devtools"
    go run ./cmd/verifyplan evidence --root .. --gate tier1-bigdecimal-fuzz --log "$run_dir/fuzz.log"
elif [[ -n "$corpus" ]]; then
    go test -count=1 -v "-run=^FuzzTier1BigDecimal/$corpus\$" -timeout=30s . 2>&1 | tee "$run_dir/replay.log"
    passed=0
    while IFS= read -r line; do
        if [[ "$line" == "    --- PASS: FuzzTier1BigDecimal/$corpus ("* ]]; then passed=1; fi
    done < "$run_dir/replay.log"
    if (( ! passed )); then echo "Requested Tier1 input did not execute: $corpus" >&2; exit 1; fi
else
    if [[ -n "$replay" ]]; then export BID754_TIER1_REPLAY="$replay"; fi
    {
        go test -count=1 -v '-run=^TestTier1' -timeout=180s .
        go test -count=1 -v -timeout=180s ./internal/tier1ref
    } 2>&1 | tee "$run_dir/test.log"
fi
