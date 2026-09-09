#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
if [ "$#" -eq 0 ]; then
    if [ -n "${BID754_SNAPSHOT_ARCHIVE:-}" ] || [ -n "${BID754_SNAPSHOT_ID:-}" ]; then
        : "${BID754_SNAPSHOT_ARCHIVE:?snapshot archive and ID must both be set}"
        : "${BID754_SNAPSHOT_ID:?snapshot archive and ID must both be set}"
        python3 -B devtools/scripts/lib/source_snapshot.py verify-source "$BID754_SNAPSHOT_ARCHIVE" \
            --expected-id "$BID754_SNAPSHOT_ID" --root . > /dev/null
        exec python3 -B devtools/scripts/lib/source_snapshot.py tree-id "$BID754_SNAPSHOT_ARCHIVE" \
            --expected-id "$BID754_SNAPSHOT_ID"
    fi
    exec python3 -B devtools/scripts/lib/source_snapshot.py current-tree-id
elif [ "$#" -eq 2 ] && [ "$1" = --snapshot ]; then
    exec python3 -B devtools/scripts/lib/source_snapshot.py tree-id "$2"
else
    echo "usage: $0 [--snapshot ARCHIVE]" >&2
    exit 2
fi
