#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)

cd "$repo_root/bid754-rs"

echo "==> Rust generated overflow policy"

generated_count=$(find src/generated -type f -name '*.rs' | wc -l | tr -d ' ')
if [ "$generated_count" != "104" ]; then
  echo "generated Rust file count changed: got $generated_count, want 104" >&2
  exit 1
fi

if grep -Fq 'overflow-checks = false' Cargo.toml; then
  echo "bid754-rs must not disable Rust overflow checks at the Cargo profile level" >&2
  exit 1
fi

allow_block=$(sed -n '/^#!\[allow(/,/^)]/p' src/lib.rs)
for lint in arithmetic_overflow overflowing_literals; do
  if ! printf '%s\n' "$allow_block" | grep -Eq "(^|[^[:alnum:]_])${lint}([^[:alnum:]_]|$)"; then
    echo "missing explicit crate-level ${lint} allowance in bid754-rs/src/lib.rs" >&2
    exit 1
  fi
done

if grep -Fq "if ((x - b'A') <= (b'Z' - b'A'))" src/generated/bid64_from_string.rs; then
  echo "byte tolower overflow sentinel regressed in bid64_from_string.rs" >&2
  exit 1
fi

echo "==> Rust generated implementation tests with default overflow policy"
cargo test --locked --quiet

echo "==> Rust generated implementation tests with overflow-checks=yes"
RUSTFLAGS='-C overflow-checks=yes' cargo test --locked --quiet

echo "Rust overflow policy audit passed: generated Rust no longer requires Cargo-level overflow-checks=false."
