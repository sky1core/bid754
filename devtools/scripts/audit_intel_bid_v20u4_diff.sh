#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cache_dir="${BID754_AUDIT_CACHE:-$repo_root/devtools/third_party/intel_dfp}"

u3_url="https://www.netlib.org/misc/intel/IntelRDFPMathLib20U3.tar.gz"
u4_url="https://www.netlib.org/misc/intel/IntelRDFPMathLib20U4.tar.gz"
u3_sha="13f6924b2ed71df9b137a7df98706a0dcc3b43c283a0e32f8b6eadca4305136a"
u4_sha="1df86132e7a31fd74d784fee1c679b21a088f73a8ec979cfaf784c200392e125"
u3_archive="$cache_dir/IntelRDFPMathLib20U3.tar.gz"
u4_archive="$cache_dir/IntelRDFPMathLib20U4.tar.gz"

expected_src_changes=240
expected_semantic_changes=11

semantic_review_files=(
  "bid128_string.c"
  "bid64_string.c"
  "bid_strtod.h"
  "bid32_sqrt.c"
  "bid64_sqrt.c"
  "bid128_sqrt.c"
  "bid64_div.c"
  "bid128_div.c"
  "bid128_fmod.c"
  "bid128_rem.c"
  "bid_functions.h"
)

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required tool: $1" >&2
    exit 1
  fi
}

sha256_file() {
  shasum -a 256 "$1" | awk '{print $1}'
}

download_if_missing() {
  local url=$1
  local output=$2
  mkdir -p "$(dirname "$output")"
  if [ ! -f "$output" ]; then
    echo "downloading $url"
    curl -L --fail --show-error --silent "$url" -o "$output"
  fi
}

verify_sha256() {
  local path=$1
  local want=$2
  local got
  got=$(sha256_file "$path")
  if [ "$got" != "$want" ]; then
    echo "checksum mismatch for $path" >&2
    echo "  got:  $got" >&2
    echo "  want: $want" >&2
    exit 1
  fi
}

collect_src_files() {
  local root=$1
  (cd "$root/LIBRARY/src" && find . -type f | sed 's#^\./##' | sort)
}

extract_archive() {
  local archive=$1
  local output=$2
  mkdir -p "$output"
  tar -xzf "$archive" -C "$output"
}

contains_line() {
  local needle=$1
  local file=$2
  rg -qx --fixed-strings "$needle" "$file"
}

require_tool awk
require_tool comm
require_tool curl
require_tool find
require_tool rg
require_tool shasum
require_tool sort
require_tool tar

download_if_missing "$u3_url" "$u3_archive"
download_if_missing "$u4_url" "$u4_archive"
verify_sha256 "$u3_archive" "$u3_sha"
verify_sha256 "$u4_archive" "$u4_sha"

tmpdir=$(mktemp -d)
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

u3_root="$tmpdir/u3"
u4_root="$tmpdir/u4"
extract_archive "$u3_archive" "$u3_root"
extract_archive "$u4_archive" "$u4_root"

u3_files="$tmpdir/u3_files.txt"
u4_files="$tmpdir/u4_files.txt"
all_files="$tmpdir/all_src_files.txt"
changed_files="$tmpdir/changed_src_files.txt"
semantic_files="$tmpdir/semantic_files.txt"
nonsemantic_files="$tmpdir/nonsemantic_files.txt"

collect_src_files "$u3_root" >"$u3_files"
collect_src_files "$u4_root" >"$u4_files"
comm -3 "$u3_files" "$u4_files" | sed 's/^\t//' >"$all_files"
comm -12 "$u3_files" "$u4_files" >>"$all_files"
sort -u "$all_files" -o "$all_files"

: >"$changed_files"
while IFS= read -r rel; do
  if [ ! -f "$u3_root/LIBRARY/src/$rel" ] || [ ! -f "$u4_root/LIBRARY/src/$rel" ]; then
    echo "$rel" >>"$changed_files"
  elif ! cmp -s "$u3_root/LIBRARY/src/$rel" "$u4_root/LIBRARY/src/$rel"; then
    echo "$rel" >>"$changed_files"
  fi
done <"$all_files"
sort -u "$changed_files" -o "$changed_files"

: >"$semantic_files"
for rel in "${semantic_review_files[@]}"; do
  echo "$rel" >>"$semantic_files"
  if ! contains_line "$rel" "$changed_files"; then
    echo "expected semantic-review file not changed in v20U4 diff: $rel" >&2
    exit 1
  fi
done
sort -u "$semantic_files" -o "$semantic_files"

comm -23 "$changed_files" "$semantic_files" >"$nonsemantic_files"

src_changes=$(wc -l <"$changed_files" | tr -d ' ')
semantic_changes=$(wc -l <"$semantic_files" | tr -d ' ')
nonsemantic_changes=$(wc -l <"$nonsemantic_files" | tr -d ' ')

if [ "$src_changes" -ne "$expected_src_changes" ]; then
  echo "unexpected LIBRARY/src changed-file count: got $src_changes, want $expected_src_changes" >&2
  exit 1
fi
if [ "$semantic_changes" -ne "$expected_semantic_changes" ]; then
  echo "unexpected semantic-review file count: got $semantic_changes, want $expected_semantic_changes" >&2
  exit 1
fi
if [ "$nonsemantic_changes" -ne $((expected_src_changes - expected_semantic_changes)) ]; then
  echo "unexpected non-semantic remainder count: got $nonsemantic_changes" >&2
  exit 1
fi

layout_checks=(
  "removed:LIBRARY/RUNOSXINTEL64"
  "removed:LIBRARY/macbuild"
  "added:LIBRARY/RUNLINUXMACOSINTEL64_CLANG"
  "added:LIBRARY/makefile"
)

for check in "${layout_checks[@]}"; do
  kind=${check%%:*}
  path=${check#*:}
  case "$kind" in
    removed)
      if [ -e "$u4_root/$path" ]; then
        echo "expected removed path still exists in v20U4: $path" >&2
        exit 1
      fi
      ;;
    added)
      if [ ! -e "$u4_root/$path" ]; then
        echo "expected added path missing in v20U4: $path" >&2
        exit 1
      fi
      ;;
    *)
      echo "internal error: unknown layout check $check" >&2
      exit 1
      ;;
  esac
done

echo "Intel BID v20U3 -> v20U4 diff audit OK"
echo "  LIBRARY/src changed files: $src_changes"
echo "  semantic-review files:     $semantic_changes"
echo "  reviewed remainder files:  $nonsemantic_changes"
echo
echo "Semantic-review allowlist:"
sed 's/^/  - /' "$semantic_files"
