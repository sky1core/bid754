#!/usr/bin/env bash
# Read-only helper: print the current content hash of every
# verification_artifact_sha256 bucket, as computed by the anchor test itself.
#
# The Go test devtools/internal/testgen/TestVerificationArtifactContentHashes is
# the single source of truth for bucket membership; this script only runs it and
# extracts the hashes it logs. It DOES NOT read or write
# devtools/verification_anchors.json. After an intentional regeneration, run this
# script, then update verification_anchors.json BY HAND, stating the input change
# that moved each hash.
#
# Generators, templates, and emitters must never touch verification_anchors.json;
# this convenience script is a manual dev tool, not part of any
# generation path.
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$repo_root/devtools"

# The test logs one "verification_artifact_sha256.<bucket> = <hash>" line per
# bucket whether it passes or fails, so this works both to seed new anchors and
# to read the new value behind a failing anchor. "|| true" keeps a failing test
# (stale anchor) from aborting the extraction.
out=$(GOCACHE=${GOCACHE:-/tmp/go-cache} go test -count=1 -run '^TestVerificationArtifactContentHashes$' -v ./internal/testgen 2>&1 || true)

hashes=$(printf '%s\n' "$out" \
    | sed -n 's/.*verification_artifact_sha256\.\([A-Za-z0-9_]*\) = \([0-9a-f]\{64\}\).*/\1 \2/p' \
    | sort)

if [ -z "$hashes" ]; then
    echo "❌ could not extract any bucket hashes; the test output was:" >&2
    printf '%s\n' "$out" >&2
    exit 1
fi

echo "# current verification_artifact_sha256 bucket hashes (paste into"
echo "# devtools/verification_anchors.json by hand, with the input change noted):"
printf '%s\n' "$hashes"
