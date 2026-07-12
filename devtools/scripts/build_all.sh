#!/bin/bash
# native 의존성 설치 후 전체 검증 게이트(make verify-all)를 실행하는 래퍼

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$(dirname "$(dirname "$SCRIPT_DIR")")"

echo "1. native 의존성 설치 (IBM decNumber + Intel BID)..."
make setup-native

echo "2. 전체 검증 게이트 실행..."
export GOCACHE="${TMPDIR:-/tmp}/bid754-gocache"
mkdir -p "$GOCACHE"
make verify-all

cat <<'EOF'
빌드/테스트 성공!

devtools/scripts/build_all.sh는 native 의존성 설치를 make setup-native로, 검증 경계를
make verify-all로 위임한다. verify-all는 native smoke, generated FFI,
generated readtest, generated decTest, Rust native, cexport disabled check
게이트까지 실행한다.
EOF
