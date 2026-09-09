#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."
python3 - "$@" <<'PYTHON'
import argparse
from pathlib import Path
import re
import subprocess
import sys

parser = argparse.ArgumentParser(description="Compare platform bit digests from the same clean source tree")
parser.add_argument("--results-dir", type=Path, default=Path("test_results"))
parser.add_argument("--expected-tree")
parser.add_argument("--require-platform", action="append", default=[])
args = parser.parse_args()


def fail(message):
    sys.exit(f"verify-digest: {message}")


expected_tree = args.expected_tree
if expected_tree is None:
    expected_tree = subprocess.check_output(["bash", "devtools/scripts/print_tree_id.sh"], text=True).strip()
if not re.fullmatch(r"[0-9a-f]{40}|[0-9a-f]{64}", expected_tree):
    fail("expected tree must identify a clean checkout")
required = set(args.require_platform)
if any(not re.fullmatch(r"[a-z0-9]+/[a-z0-9]+", platform) for platform in required):
    fail("required platforms must use os/arch format")
files = sorted(args.results_dir.glob("digest_*.txt"))
if len(files) < 2:
    fail(f"need results from at least two platforms; found {len(files)}")
reference = None
platforms = set()
for path in files:
    lines = path.read_text().splitlines()
    if len(lines) != 2:
        fail(f"{path} must contain exactly one tree and one digest record")
    if lines[0] != f"PLATFORM-DIGEST-TREE {expected_tree}":
        fail(f"{path} tree mismatch: expected {expected_tree}, found {lines[0]!r}")
    match = re.fullmatch(r"PLATFORM-DIGEST goos=([a-z0-9]+) goarch=([a-z0-9]+) cases=([1-9][0-9]*) sha256=([0-9a-f]{64})", lines[1])
    if match is None:
        fail(f"{path} has an invalid digest record")
    goos, goarch, cases, checksum = match.groups()
    platform = f"{goos}/{goarch}"
    if path.name != f"digest_{goos}_{goarch}.txt":
        fail(f"{path} filename disagrees with recorded platform {platform}")
    if platform in platforms:
        fail(f"duplicate platform {platform}")
    platforms.add(platform)
    value = (cases, checksum)
    if reference is not None and value != reference:
        fail(f"{path} case-count or digest mismatch: {value} != {reference}")
    reference = value
    print(f"{path}: {lines[1]} (tree={expected_tree})")
missing = required - platforms
if missing:
    fail(f"missing required platforms: {', '.join(sorted(missing))}")
print(f"verify-digest: {len(platforms)} platforms agree (tree={expected_tree} cases={reference[0]} sha256={reference[1]})")
PYTHON
