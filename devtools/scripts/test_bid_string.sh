#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
go_cache=${GOCACHE:-/tmp/go-cache}

cd "$repo_root"

make verify-generated

if ! command -v rg >/dev/null 2>&1; then
  echo "rg is required for BID string vector consumer discovery" >&2
  exit 1
fi

vector_tests_file=$(mktemp)
go_modules_file=$(mktemp)
rust_tests_file=$(mktemp)
trap 'rm -f "$vector_tests_file" "$go_modules_file" "$rust_tests_file"' EXIT

rg_status=0
rg -l 'TestGeneratedBIDStringVectors|test_generated_bid_string_vectors' \
  -g '*string_vectors_test.go' \
  -g '*/tests/*string_vectors.rs' \
  --glob '!test_results/**' \
  . >"$vector_tests_file" || rg_status=$?
if [ "$rg_status" -ne 0 ] && [ "$rg_status" -ne 1 ]; then
  echo "failed to discover BID string vector consumer tests" >&2
  exit "$rg_status"
fi

if [ ! -s "$vector_tests_file" ]; then
  echo "no BID string vector consumer tests found" >&2
  exit 1
fi

find_parent_file() {
  local dir=$1
  local marker=$2
  while [ "$dir" != "." ] && [ "$dir" != "/" ]; do
    if [ -f "$dir/$marker" ]; then
      printf '%s\n' "$dir"
      return 0
    fi
    dir=$(dirname "$dir")
  done
  if [ -f "$marker" ]; then
    printf '.\n'
    return 0
  fi
  return 1
}

while IFS= read -r path; do
  case "$path" in
    *_test.go)
      module_dir=$(find_parent_file "$(dirname "$path")" go.mod) || {
        echo "BID string Go vector test has no parent go.mod: $path" >&2
        exit 1
      }
      printf '%s\n' "$module_dir" >>"$go_modules_file"
      ;;
    *.rs)
      crate_dir=$(find_parent_file "$(dirname "$path")" Cargo.toml) || {
        echo "BID string Rust vector test has no parent Cargo.toml: $path" >&2
        exit 1
      }
      test_name=$(basename "$path" .rs)
      printf '%s:%s\n' "$crate_dir" "$test_name" >>"$rust_tests_file"
      ;;
    *)
      echo "unsupported BID string vector consumer test language: $path" >&2
      exit 1
      ;;
  esac
done <"$vector_tests_file"

require_discovered_consumer() {
  local file=$1
  local expected=$2
  local label=$3
  if ! grep -Fxq "$expected" "$file"; then
    echo "missing required BID string vector consumer for $label: $expected" >&2
    exit 1
  fi
}

require_discovered_consumer "$go_modules_file" "." "Go mechanical port (bid-go package in the root module)"
require_discovered_consumer "$rust_tests_file" "./bid754-rs:bid_string_vectors" "Rust generated implementation"

sort -u "$go_modules_file" | while IFS= read -r module_dir; do
  [ -n "$module_dir" ] || continue
  echo "==> go BID string vector tests: $module_dir"
  (cd "$module_dir" && GOCACHE="$go_cache" go test ./... -run TestGeneratedBIDStringVectors)
done

sort -u "$rust_tests_file" | while IFS= read -r key; do
  [ -n "$key" ] || continue
  crate_dir=${key%:*}
  test_name=${key#*:}
  echo "==> rust BID string vector test: $crate_dir --test $test_name"
  (cd "$crate_dir" && cargo test --locked --test "$test_name")
done
