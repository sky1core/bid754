#!/bin/bash
# native 의존성 설치 후 전체 검증 게이트(make full-audit)를 실행하는 래퍼

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$(dirname "$(dirname "$SCRIPT_DIR")")"

echo "1. native 의존성 설치 (IBM decNumber + Intel BID)..."
make setup-native

echo "2. 전체 검증 게이트 실행..."
export GOCACHE="${TMPDIR:-/tmp}/bid754-gocache"
mkdir -p "$GOCACHE"
make full-audit

cat <<'EOF'
빌드/테스트 성공!

devtools/scripts/build_all.sh는 native 의존성 설치를 make setup-native로, 검증 경계를
make full-audit로 위임한다. full-audit는 native smoke, generated FFI,
generated readtest, generated decTest, Rust native, cexport quarantine guard
게이트까지 실행한다.
EOF
