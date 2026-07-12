#!/usr/bin/env bash
# Structural check for the verify-generated comparison set.
#
# verify-generated mirrors its cp/cmp file list by hand, so a new generated
# artifact that never gets added to that list remains outside reproducibility
# verification. This script ensures every tracked file that carries a
# standard generated-code marker line must be either
#   (a) actually compared by the verify-generated recipe (a cmp/diff line, not
#       merely a cp backup — a cp without a paired compare verifies nothing), or
#   (b) listed in devtools/scripts/generated_marker_exceptions.txt with a
#       documented reason.
# The comparison set is extracted from `make -n verify-generated` so the check
# follows the recipe instead of keeping a second hand-mirrored list.
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$repo_root"

exceptions_file="devtools/scripts/generated_marker_exceptions.txt"
marker_regex='^(//|#) Code generated .* DO NOT EDIT\.$'

fail() {
    echo "❌ $*" >&2
    exit 1
}

# --- 1. tracked files carrying a generated-code marker ---
marked_files=$(git grep -lIE "$marker_regex" -- . || true)
[ -n "$marked_files" ] || fail "no tracked files with a generated-code marker found; the marker regex is broken"

# --- 2. file set compared by the verify-generated recipe ---
recipe=$(make -s -n verify-generated 2>/dev/null) || fail "cannot expand the verify-generated recipe via make -n"

compared_tmp=$(mktemp)
trap 'rm -f "$compared_tmp"' EXIT

# A file only counts as compared when the recipe actually diffs it against the
# $tmpdir backup: cp lines alone are just backups, so extracting from cp would
# count a copied-but-never-compared artifact as covered.

# pairwise compares: cmp -s <tracked> $tmpdir/... (either operand order)
printf '%s\n' "$recipe" \
    | sed -E 's/^[[:space:]]+//' \
    | awk '$1 == "cmp" && $2 == "-s" {
        if ($3 ~ /^\$tmpdir\//) { print $4 } else if ($4 ~ /^\$tmpdir\//) { print $3 }
    }' >> "$compared_tmp"

# recursive directory compares: diff -r[u] <a> <b> -> every tracked file below
# the non-$tmpdir operand
while IFS= read -r dir; do
    [ -n "$dir" ] || continue
    git ls-files -- "$dir" >> "$compared_tmp"
done < <(printf '%s\n' "$recipe" \
    | sed -E 's/^[[:space:]]+//' \
    | awk '$1 == "diff" && $2 ~ /^-r/ {
        if ($3 ~ /^\$tmpdir\//) { print $4 } else if ($4 ~ /^\$tmpdir\//) { print $3 }
    }')

# loop compares: for f in <files>; do cmp -s bid754-go/$f $tmpdir/$f; done
while IFS= read -r list; do
    for f in $list; do
        printf 'bid754-go/%s\n' "$f" >> "$compared_tmp"
    done
done < <(printf '%s\n' "$recipe" \
    | sed -nE 's/.*for f in ([^;]*); do cmp -s bid754-go\/\$f .*/\1/p')

compared_files=$(sort -u "$compared_tmp")
compared_count=$(printf '%s\n' "$compared_files" | sed '/^$/d' | wc -l | tr -d ' ')
[ "$compared_count" -ge 50 ] || fail "extracted only $compared_count compared files from the verify-generated recipe; the recipe parser no longer matches the Makefile"

# --- 3. exceptions (annotated, hand-maintained) ---
exception_files=""
if [ -f "$exceptions_file" ]; then
    exception_files=$(sed -E 's/[[:space:]]*#.*$//' "$exceptions_file" | sed '/^[[:space:]]*$/d')
fi

in_set() { # $1=needle, $2=newline-separated haystack
    # No pipe here: under pipefail, grep -Fxq exiting early on a match would
    # SIGPIPE the printf side and turn a real match into exit 141.
    grep -Fxq -- "$1" <<<"$2"
}

# exceptions hygiene: every entry must be tracked, must carry a marker, and
# must not already be covered by the verify-generated comparison set.
while IFS= read -r entry; do
    [ -n "$entry" ] || continue
    git ls-files --error-unmatch "$entry" >/dev/null 2>&1 \
        || fail "exception entry is not a tracked file: $entry"
    in_set "$entry" "$marked_files" \
        || fail "exception entry no longer carries a generated-code marker (stale entry): $entry"
    if in_set "$entry" "$compared_files"; then
        fail "exception entry is already compared by verify-generated (redundant entry): $entry"
    fi
done <<< "$exception_files"

# --- 4. coverage check ---
uncovered=""
covered_count=0
exception_count=0
while IFS= read -r f; do
    [ -n "$f" ] || continue
    if in_set "$f" "$compared_files"; then
        covered_count=$((covered_count + 1))
    elif [ -n "$exception_files" ] && in_set "$f" "$exception_files"; then
        exception_count=$((exception_count + 1))
    else
        uncovered="$uncovered$f"$'\n'
    fi
done <<< "$marked_files"

marked_count=$(printf '%s\n' "$marked_files" | sed '/^$/d' | wc -l | tr -d ' ')
echo "generated-marker coverage: $marked_count marked files = $covered_count verified by verify-generated + $exception_count documented exception(s)"
if [ -n "$exception_files" ]; then
    echo "documented in $exceptions_file:"
    printf '%s\n' "$exception_files" | sed 's/^/  /'
fi

if [ -n "$uncovered" ]; then
    echo "❌ files carry a generated-code marker but are neither compared by verify-generated nor documented as exceptions:" >&2
    printf '%s' "$uncovered" | sed 's/^/  /' >&2
    echo "add them to the verify-generated cp+cmp/diff comparison set, or document them in $exceptions_file" >&2
    exit 1
fi

echo "✅ every generated-code marker file is covered"
