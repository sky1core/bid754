#!/usr/bin/env bash
set -euo pipefail

# Verify that every versioned package manifest in the tree carries the same
# project version. The Version constant in bid754-go/bid754.go is the single
# version source that scripts read (see verify_bidcodec_packages.sh); this
# verification pins the remaining manifests to it so a version bump cannot leave a
# manifest behind. This verification compares manifests only; it does not change any
# version.

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$repo_root"

# shellcheck source=devtools/scripts/lib/project_version.sh
source "$repo_root/devtools/scripts/lib/project_version.sh"

expected=$(read_go_version_constant bid754-go/bid754.go)
if [ -z "$expected" ]; then
  echo "ERROR: failed to read the Version constant from bid754-go/bid754.go" >&2
  exit 1
fi
echo "==> project version source bid754-go/bid754.go Version: $expected"

fail=0

check_version() {
  manifest="$1"
  value="$2"
  if [ -z "$value" ]; then
    echo "ERROR: failed to read a version from $manifest" >&2
    fail=1
    return
  fi
  if [ "$value" != "$expected" ]; then
    echo "ERROR: $manifest version $value != bid754-go/bid754.go Version $expected" >&2
    fail=1
    return
  fi
  echo "  $manifest: $value"
}

toml_package_version() {
  python3 -c 'import sys, tomllib
with open(sys.argv[1], "rb") as f:
    print(tomllib.load(f)["package"]["version"])' "$1"
}

check_version "bid754-rs/Cargo.toml" "$(toml_package_version bid754-rs/Cargo.toml)"
check_version "bid754-codec-rs/Cargo.toml" "$(toml_package_version bid754-codec-rs/Cargo.toml)"
check_version "bid754-codec-js/package.json" "$(python3 -c 'import json, sys
with open(sys.argv[1]) as f:
    print(json.load(f)["version"])' bid754-codec-js/package.json)"
check_version "bid754-codec-py/pyproject.toml" "$(read_pyproject_version bid754-codec-py/pyproject.toml)"
check_version "bid754-codec-java/build.gradle" "$(read_gradle_version bid754-codec-java/build.gradle)"

if [ "$fail" -ne 0 ]; then
  echo "package manifest versions diverge from bid754-go/bid754.go Version" >&2
  exit 1
fi
echo "✅ all package manifests carry version $expected"
