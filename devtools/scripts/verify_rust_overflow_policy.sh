#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)

cd "$repo_root/bid754-rs"

echo "==> Rust generated overflow policy"

# Count only the generated implementation files. The public API wrapper subtree
# src/generated/api/ carries no wrapping arithmetic (it routes through the port)
# and re-denies the overflow lints, so it is excluded from the implementation
# file tripwire and verified separately below.
generated_count=$(find src/generated -type f -name '*.rs' -not -path '*/api/*' | wc -l | tr -d ' ')
if [ "$generated_count" != "102" ]; then
  echo "generated Rust implementation file count changed: got $generated_count, want 102 (api/ wrappers excluded)" >&2
  exit 1
fi

if grep -Fq 'overflow-checks = false' Cargo.toml; then
  echo "bid754-rs must not disable Rust overflow checks at the Cargo profile level" >&2
  exit 1
fi

# The overflow lints are allowed at the crate level for the generated
# implementation/compat modules, and the public API module re-denies them so the
# allowance does not cover the public surface. Both halves are required: the
# crate-level allowance must still exist for the implementation, and the public
# API module must re-deny (a module-level deny overrides the crate-level allow).
allow_block=$(sed -n '/^#!\[allow(/,/^)]/p' src/lib.rs)
for lint in arithmetic_overflow overflowing_literals; do
  if ! printf '%s\n' "$allow_block" | grep -Eq "(^|[^[:alnum:]_])${lint}([^[:alnum:]_]|$)"; then
    echo "missing explicit crate-level ${lint} allowance in bid754-rs/src/lib.rs" >&2
    exit 1
  fi
  if ! grep -Eq "#!\[deny\(.*${lint}.*\)\]" src/generated/api/mod.rs; then
    echo "public API module src/generated/api/mod.rs must re-deny ${lint}: the crate-level overflow allowance must not cover the public surface" >&2
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

echo "Rust overflow policy verification passed: generated Rust no longer requires Cargo-level overflow-checks=false."
