#!/usr/bin/env bash
# verify_digest.sh - compare PLATFORM-DIGEST results across platforms
# (PLATFORM_SPEC section 4 item 2: direct cross-platform bit comparison).
#
# Inputs are test_results/digest_<os>_<arch>.txt files produced by
# `make digest` on this host and extracted from the `make verify-linux`
# portable legs. All digests must agree on case count and SHA-256.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/../.."

files=()
for f in test_results/digest_*.txt; do
    [ -f "$f" ] && files+=("$f")
done

if [ "${#files[@]}" -lt 2 ]; then
    echo "verify-digest: need PLATFORM-DIGEST results from at least two platforms; found ${#files[@]}" >&2
    echo "  produce them with 'make digest' (this host) and 'make verify-linux' (Linux legs)" >&2
    exit 1
fi

ref_sum=""
ref_cases=""
ref_file=""
for f in "${files[@]}"; do
    line=$(grep '^PLATFORM-DIGEST ' "$f" | tail -1 || true)
    if [ -z "$line" ]; then
        echo "verify-digest: $f has no PLATFORM-DIGEST line" >&2
        exit 1
    fi
    sum=${line##*sha256=}
    cases=$(printf '%s\n' "$line" | sed -n 's/.*cases=\([0-9]*\).*/\1/p')
    echo "$f: $line"
    if [ -z "$ref_sum" ]; then
        ref_sum="$sum"; ref_cases="$cases"; ref_file="$f"
    else
        if [ "$cases" != "$ref_cases" ]; then
            echo "verify-digest: case-count mismatch: $ref_file=$ref_cases vs $f=$cases" >&2
            exit 1
        fi
        if [ "$sum" != "$ref_sum" ]; then
            echo "verify-digest: DIGEST MISMATCH: $ref_file=$ref_sum vs $f=$sum" >&2
            exit 1
        fi
    fi
done

echo "verify-digest: ${#files[@]} platforms agree (cases=$ref_cases sha256=$ref_sum)"
