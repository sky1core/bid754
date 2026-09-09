#!/usr/bin/env bash
# Print a tree identifier that binds PLATFORM-DIGEST results to the exact
# source state they were produced from: the HEAD commit id, with a content
# digest and "-dirty" suffix when the working tree differs from HEAD, or "unknown" when no git
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
    fingerprint=$(python3 - <<'PY'
import hashlib
import os
import stat
import subprocess

digest = hashlib.sha256()

def add(value):
    digest.update(len(value).to_bytes(8, "big"))
    digest.update(value)

add(subprocess.check_output(["git", "diff", "HEAD", "--no-ext-diff", "--no-textconv", "--no-renames", "--binary"]))
paths = subprocess.check_output(["git", "ls-files", "--others", "--exclude-standard", "-z"]).split(b"\0")
for path in sorted(filter(None, paths)):
    add(path)
    mode = os.lstat(path).st_mode
    add(str(stat.S_IMODE(mode)).encode())
    if stat.S_ISLNK(mode):
        add(b"symlink")
        add(os.readlink(path))
    else:
        add(b"file")
        with open(path, "rb") as source:
            add(source.read())
print(digest.hexdigest())
PY
)
    echo "${commit}-${fingerprint}-dirty"
else
    echo "$commit"
fi
