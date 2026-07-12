#!/usr/bin/env bash
# Print a tree identifier that binds PLATFORM-DIGEST results to the exact
# source state they were produced from: the HEAD commit id, with a "-dirty"
# suffix when the working tree differs from HEAD, or "unknown" when no git
# history is available (e.g. the synthetic no-commit index inside the
# verify-linux container; the host side stamps the real id there).
# verify_digest.sh refuses to compare digest files whose tree ids disagree or
# are dirty/unknown, so digests produced from different code states can never
# be reported as cross-platform agreement.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
if ! commit=$(git rev-parse HEAD 2>/dev/null); then
    echo "unknown"
    exit 0
fi
if [ -n "$(git status --porcelain 2>/dev/null)" ]; then
    echo "${commit}-dirty"
else
    echo "$commit"
fi
