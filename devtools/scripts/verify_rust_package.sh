#!/usr/bin/env bash
set -euo pipefail

# Verify the bid754-rs package contents and dependency shape:
#   1. `cargo package --list` contains the required source and metadata files
#      and nothing outside the declared package layout;
#   2. [dependencies] contains only the exact pinned entries; and
#   3. the packaged crate remains a pure-Rust implementation with no native
#      link declarations in Cargo.toml or the packaged .rs files.
# `cargo publish --dry-run` is separate because Cargo.toml currently sets
# `publish = false`; changing publication status requires a later user decision.

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
crate_dir="$repo_root/bid754-rs"
cd "$crate_dir"

echo "==> cargo package --list exact file-set verification (bid754-rs)"

# --allow-dirty lets this check inspect the current working tree. --locked keeps
# dependency resolution aligned with the repository's other Rust commands.
package_list=$(cargo package --list --locked --allow-dirty 2>/dev/null)
if [ -z "$package_list" ]; then
  echo "ERROR: cargo package --list produced no output" >&2
  exit 1
fi

required_files=("Cargo.toml" "LICENSE" "README.md")
if [ -f "NOTICE" ]; then
  required_files+=("NOTICE")
fi

# Source anchors: src/lib.rs is the crate root and
# src/generated/api/mod.rs is the public-API module root.
required_source_files=("src/lib.rs" "src/generated/api/mod.rs")

# Cargo-managed bookkeeping files that accompany a packaged crate.
cargo_managed_files=(".cargo_vcs_info.json" "Cargo.lock" "Cargo.toml.orig")

echo "-- required file presence"
for f in "${required_files[@]}" "${required_source_files[@]}"; do
  if ! printf '%s\n' "$package_list" | grep -qxF "$f"; then
    echo "ERROR: required file missing from cargo package --list output: $f" >&2
    exit 1
  fi
  echo "  present: $f"
done

# Beyond the two named source anchors, the package must actually carry the
# body of src sources; a package that shipped only lib.rs would be broken.
shipped_src_count=$(printf '%s\n' "$package_list" | grep -cE '^src/.*\.rs$' || true)
if [ "$shipped_src_count" -lt 1 ]; then
  echo "ERROR: cargo package --list shipped no src/*.rs sources" >&2
  exit 1
fi
echo "  present: $shipped_src_count src/*.rs source file(s) shipped"

echo "-- expected file set"
unexpected=0
while IFS= read -r entry; do
  [ -z "$entry" ] && continue
  case "$entry" in
    src/*) continue ;;
  esac
  allowed=0
  for f in "${required_files[@]}" "${cargo_managed_files[@]}"; do
    if [ "$entry" = "$f" ]; then
      allowed=1
      break
    fi
  done
  if [ "$allowed" -ne 1 ]; then
    echo "ERROR: unexpected file in cargo package --list output: $entry" >&2
    unexpected=1
  fi
done <<<"$package_list"
if [ "$unexpected" -ne 0 ]; then
  echo "cargo package --list included files outside src/** and the expected metadata files" >&2
  exit 1
fi
file_count=$(printf '%s\n' "$package_list" | grep -c .)
echo "  ✅ package contains only the expected file set ($file_count files total)"

echo "==> [dependencies] pin and FFI-absence verification (bid754-rs)"
# The Rust analogue of the Go zero-dependency/cgo-purity contract: the
# generated public API's only runtime dependencies are the pinned
# num-bigint/num-traits pair, and no FFI/native-link plumbing rides along.
python3 - "$crate_dir/Cargo.toml" "$package_list" <<'PY'
import hashlib
import io
import json
import os
from pathlib import Path, PurePosixPath
import subprocess
import sys
import tarfile
import tomllib

path = sys.argv[1]
with open(path, "rb") as f:
    manifest = tomllib.load(f)

package = manifest.get("package", {})
if package.get("links") is not None:
    print(f"ERROR: package.links={package.get('links')!r} declares a native library link", file=sys.stderr)
    sys.exit(1)
if package.get("build") is not None:
    print(f"ERROR: package.build={package.get('build')!r} declares a custom build script", file=sys.stderr)
    sys.exit(1)
print("  no links/build-script native-linkage fields")

if manifest.get("build-dependencies"):
    print(f"ERROR: unexpected [build-dependencies]: {sorted(manifest['build-dependencies'])}", file=sys.stderr)
    sys.exit(1)
if "target" in manifest:
    print(f"ERROR: unexpected [target.*] table(s): {sorted(manifest['target'])}", file=sys.stderr)
    sys.exit(1)
print("  no [build-dependencies] or [target.*] tables")

deps = manifest.get("dependencies", {})
expected = {"num-bigint": "=0.4.6", "num-traits": "=0.2.19"}

actual_names = set(deps.keys())
expected_names = set(expected.keys())
if actual_names != expected_names:
    missing = expected_names - actual_names
    extra = actual_names - expected_names
    if missing:
        print(f"ERROR: missing pinned [dependencies]: {sorted(missing)}", file=sys.stderr)
    if extra:
        print(f"ERROR: unexpected [dependencies] beyond the pinned set: {sorted(extra)}", file=sys.stderr)
    sys.exit(1)

for name, want_version in expected.items():
    entry = deps[name]
    got_version = entry if isinstance(entry, str) else entry.get("version")
    if got_version != want_version:
        print(f"ERROR: {name} version {got_version!r} != pinned {want_version!r}", file=sys.stderr)
        sys.exit(1)
    if isinstance(entry, dict) and entry.get("features"):
        print(f"ERROR: {name} declares non-default features {entry['features']}, expected a bare pinned dependency", file=sys.stderr)
        sys.exit(1)
    if isinstance(entry, dict) and ("path" in entry or "git" in entry):
        print(f"ERROR: {name} is a path/git dependency, expected a registry-pinned dependency", file=sys.stderr)
        sys.exit(1)
    print(f"  {name}: {got_version}")

print(f"  ✅ [dependencies] is exactly the pinned set {sorted(expected_names)}")

metadata = json.loads(subprocess.check_output(
    ["cargo", "metadata", "--locked", "--all-features", "--format-version", "1"], text=True))
manifest_path = Path(path).resolve()
packages = [p for p in metadata["packages"] if Path(p["manifest_path"]).resolve() == manifest_path]
if len(packages) != 1:
    sys.exit("ERROR: cargo metadata must identify exactly one package for the crate manifest")
resolved = {package["id"]: package for package in metadata["packages"]}
nodes = {node["id"]: node for node in metadata["resolve"]["nodes"]}
root_id = packages[0]["id"]
runtime_direct = [resolved[dep["pkg"]] for dep in nodes[root_id]["deps"]
                  if any(kind["kind"] is None for kind in dep["dep_kinds"])]
if len(runtime_direct) != len(expected) or {p["name"]: "=" + p["version"] for p in runtime_direct} != expected:
    sys.exit("ERROR: resolved dependency names and versions do not match the pinned runtime set")
pending = [root_id]
seen = {root_id}
cargo_home = Path(os.environ["CARGO_HOME"]) if "CARGO_HOME" in os.environ else Path.home() / ".cargo"
with (manifest_path.parent / "Cargo.lock").open("rb") as source:
    locked = {(p["name"], p["version"], p.get("source")): p for p in tomllib.load(source)["package"]}

def walk_error(error):
    raise error

def verify_source(package):
    identity = (package["name"], package["version"], package["source"])
    checksum = locked[identity].get("checksum")
    if not isinstance(checksum, str) or len(checksum) != 64:
        sys.exit(f"ERROR: resolved dependency {package['name']} has no locked archive checksum")
    stem = f"{package['name']}-{package['version']}"
    archive_data = None
    for candidate in sorted((cargo_home / "registry" / "cache").glob(f"*/{stem}.crate")):
        data = candidate.read_bytes()
        if hashlib.sha256(data).hexdigest() == checksum:
            archive_data = data
            break
    if archive_data is None:
        sys.exit(f"ERROR: resolved dependency {stem} has no cached archive matching Cargo.lock")
    expected_files = {}
    with tarfile.open(fileobj=io.BytesIO(archive_data), mode="r:*") as archive:
        for member in archive.getmembers():
            parts = PurePosixPath(member.name).parts
            if not parts or parts[0] != stem or ".." in parts:
                sys.exit(f"ERROR: resolved dependency {stem} has an invalid archive path")
            if member.isdir():
                continue
            if not member.isfile() or len(parts) < 2:
                sys.exit(f"ERROR: resolved dependency {stem} has an unsupported archive entry")
            expected_files["/".join(parts[1:])] = archive.extractfile(member).read()
    source_root = Path(package["manifest_path"]).resolve().parent
    actual_files = set()
    for parent, directories, files in os.walk(source_root, followlinks=False, onerror=walk_error):
        for name in directories + files:
            if (Path(parent) / name).is_symlink():
                sys.exit(f"ERROR: resolved dependency {stem} contains a source symlink")
        for name in files:
            relative = (Path(parent) / name).relative_to(source_root).as_posix()
            if relative not in {".cargo-ok", ".cargo-checksum.json"}:
                actual_files.add(relative)
    if actual_files != expected_files.keys():
        missing = sorted(expected_files.keys() - actual_files)
        extra = sorted(actual_files - expected_files.keys())
        sys.exit(f"ERROR: resolved dependency {stem} source file set differs from its locked archive: missing={missing}, extra={extra}")
    for relative, expected_bytes in expected_files.items():
        if (source_root / relative).read_bytes() != expected_bytes:
            sys.exit(f"ERROR: resolved dependency {stem} source differs from its locked archive: {relative}")

while pending:
    for dep in nodes[pending.pop()]["deps"]:
        if not any(kind["kind"] != "dev" for kind in dep["dep_kinds"]):
            continue
        package_id = dep["pkg"]
        package = resolved[package_id]
        if package["source"] != "registry+https://github.com/rust-lang/crates.io-index":
            sys.exit(f"ERROR: resolved dependency {package['name']} {package['version']} is not from crates.io: {package['source']!r}")
        if package.get("links") is not None:
            sys.exit(f"ERROR: resolved dependency {package['name']} declares native linkage")
        if package_id not in seen:
            verify_source(package)
            seen.add(package_id)
            pending.append(package_id)
print(f"  ✅ resolved dependency graph: {len(seen) - 1} runtime/build dependencies match locked crates.io archives")
shipped = {manifest_path.parent / entry for entry in sys.argv[2].splitlines()}
targets = [target for target in packages[0]["targets"] if Path(target["src_path"]) in shipped]
if not targets:
    sys.exit("ERROR: no Cargo targets are included in the package")

print("==> compiler-enforced pure Rust packaged targets (default and all features)", flush=True)
for target in targets:
    kinds = target["kind"]
    if len(kinds) == 1 and kinds[0] in {"bin", "example", "test", "bench"}:
        selector = ["--" + kinds[0], target["name"]]
    elif kinds and set(kinds) <= {"lib", "rlib", "dylib", "cdylib", "staticlib", "proc-macro"}:
        selector = ["--lib"]
    else:
        sys.exit(f"ERROR: unsupported packaged Cargo target: {target['name']} ({kinds})")
    required = target.get("required-features", [])
    feature_sets = [["--features", ",".join(required)] if required else [], ["--all-features"]]
    for features in feature_sets:
        subprocess.run(["cargo", "rustc", "--locked", *selector, *features,
                        "--", "-F", "unsafe-code", "--cap-lints=forbid"], check=True)
print(f"  ✅ {len(targets)} packaged Cargo target(s) forbid unsafe code and foreign function declarations")
PY

echo "==> cargo publish --dry-run: not run while publish = false"
echo "    Publication status is a separate user-approved change."

echo "✅ verify-rust-package passed"
