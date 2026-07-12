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
python3 - "$crate_dir/Cargo.toml" <<'PY'
import sys
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
PY

echo "==> native-link construct inspection of packaged .rs files (bid754-rs)"
# The manifest check above cannot see FFI added directly inside a source file.
# Inspect every .rs file that `cargo package` actually ships (the package_list
# entries under src/, narrowed to real files in the working tree) for
# foreign-ABI / native-link constructs. Patterns tolerate whitespace variants:
#   extern[[:space:]]*"      -> extern "C" / "system" / "C-unwind" ... (explicit foreign ABI)
#   extern[[:space:]]*\{     -> extern { ... } (implicit C-ABI FFI block)
#   #\[[[:space:]]*link      -> #[link(...)], #[link_name = ...], #[linkage] (also inner #![link...])
#   link[[:space:]]*\([[:space:]]*name -> the link(name = "...") form, wherever it appears
# extern crate / extern "Rust" callback-only code is not the target, but a
# pure generated port should carry none of these anyway (current tree: 0).
ffi_pattern='extern[[:space:]]*"|extern[[:space:]]*\{|#\[[[:space:]]*link|link[[:space:]]*\([[:space:]]*name'
ffi_hits=0
ffi_checked=0
while IFS= read -r entry; do
  [ -z "$entry" ] && continue
  case "$entry" in
    *.rs) ;;
    *) continue ;;
  esac
  file="$crate_dir/$entry"
  [ -f "$file" ] || continue
  ffi_checked=$((ffi_checked + 1))
  if grep -nE "$ffi_pattern" "$file" >/dev/null 2>&1; then
    echo "ERROR: FFI/native-link construct found in a packaged source file: $entry" >&2
    grep -nE "$ffi_pattern" "$file" >&2 || true
    ffi_hits=$((ffi_hits + 1))
  fi
done <<<"$package_list"
if [ "$ffi_checked" -lt 1 ]; then
  echo "ERROR: no packaged .rs files were available for inspection" >&2
  exit 1
fi
if [ "$ffi_hits" -ne 0 ]; then
  echo "bid754-rs is a pure-Rust generated port; no FFI/native-link construct may ship in its package sources" >&2
  exit 1
fi
echo "  ✅ no extern-ABI / #[link] / link(name=…) construct in any of the $ffi_checked packaged .rs files"

echo "==> cargo publish --dry-run: not run while publish = false"
echo "    Publication status is a separate user-approved change."

echo "✅ verify-rust-package passed"
