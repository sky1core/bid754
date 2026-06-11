#!/bin/bash
# Intel DFP Library 설치 스크립트

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
INTEL_DIR="$PROJECT_ROOT/devtools/third_party/intel_dfp"
ARCH="$(uname -m)"

INTEL_DFP_EXTRA_CFLAGS=""
INTEL_DFP_OPT_CFLAGS="${INTEL_DFP_OPT_CFLAGS:--O3 -ffp-contract=off}"
if [[ "$ARCH" == "arm64" || "$ARCH" == "aarch64" ]]; then
    # bid_conf.h does not consistently detect AArch64 as a 64-bit long target.
    # This override keeps ARM64 on the same intended 64-bit BID path as x86_64.
    INTEL_DFP_EXTRA_CFLAGS="-DBID_SIZE_LONG=8"
fi

INTEL_DFP_BUILD_STAMP="$INTEL_DIR/lib/.libbid.build-flags"

write_intel_dfp_build_stamp() {
    mkdir -p "$INTEL_DIR/lib"
    {
        echo "CALL_BY_REF=0"
        echo "GLOBAL_RND=0"
        echo "GLOBAL_FLAGS=0"
        echo "UNCHANGED_BINARY_FLAGS=0"
        echo "CFLAGS_AUX=$INTEL_DFP_EXTRA_CFLAGS"
        echo "CFLAGS_OPT=$INTEL_DFP_OPT_CFLAGS"
    } > "$INTEL_DFP_BUILD_STAMP"
}

intel_dfp_build_stamp_matches() {
    [ -f "$INTEL_DIR/lib/libbid.a" ] || return 1
    [ -f "$INTEL_DFP_BUILD_STAMP" ] || return 1
    grep -Fxq "CALL_BY_REF=0" "$INTEL_DFP_BUILD_STAMP" &&
        grep -Fxq "GLOBAL_RND=0" "$INTEL_DFP_BUILD_STAMP" &&
        grep -Fxq "GLOBAL_FLAGS=0" "$INTEL_DFP_BUILD_STAMP" &&
        grep -Fxq "UNCHANGED_BINARY_FLAGS=0" "$INTEL_DFP_BUILD_STAMP" &&
        grep -Fxq "CFLAGS_AUX=$INTEL_DFP_EXTRA_CFLAGS" "$INTEL_DFP_BUILD_STAMP" &&
        grep -Fxq "CFLAGS_OPT=$INTEL_DFP_OPT_CFLAGS" "$INTEL_DFP_BUILD_STAMP"
}

echo "Intel DFP Library 설치 중..."

# Intel DFP 라이브러리가 이미 있는지 확인
if intel_dfp_build_stamp_matches; then
    echo "Intel DFP 라이브러리가 이미 설치되어 있습니다."
    exit 0
elif [ -f "$INTEL_DIR/lib/libbid.a" ]; then
    echo "Intel DFP 라이브러리가 있지만 빌드 플래그 스탬프가 현재 설정과 다릅니다. 재빌드합니다."
fi

# 소스가 있는지 확인
if [ ! -d "$INTEL_DIR/src" ]; then
    bash "$SCRIPT_DIR/setup_generation_inputs.sh" intel
fi

if [ ! -e "$INTEL_DIR/float128" ] && [ -d "$INTEL_DIR/include/float128" ]; then
    ln -s "include/float128" "$INTEL_DIR/float128"
    echo "Intel DFP float128 레이아웃 보정 완료: float128 -> include/float128"
fi

if [ -f "$INTEL_DIR/LIBRARY/makefile" ]; then
    BUILD_DIR="$INTEL_DIR/LIBRARY"
else
    BUILD_DIR="$INTEL_DIR"
fi

# 빌드 디렉토리로 이동
cd "$BUILD_DIR"
make clean >/dev/null 2>&1 || true

# macOS에서 빌드
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "macOS에서 Intel DFP 빌드 중..."
    if command -v clang >/dev/null 2>&1; then
        make CC=clang CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 UNCHANGED_BINARY_FLAGS=0 CFLAGS_AUX="$INTEL_DFP_EXTRA_CFLAGS" CFLAGS_OPT="$INTEL_DFP_OPT_CFLAGS"
    else
        echo "오류: clang이 설치되어 있지 않습니다."
        exit 1
    fi
# Linux에서 빌드
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    echo "Linux에서 Intel DFP 빌드 중..."
    if command -v gcc >/dev/null 2>&1; then
        make CC=gcc CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 UNCHANGED_BINARY_FLAGS=0 CFLAGS_AUX="$INTEL_DFP_EXTRA_CFLAGS" CFLAGS_OPT="$INTEL_DFP_OPT_CFLAGS"
    elif command -v clang >/dev/null 2>&1; then
        make CC=clang CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 UNCHANGED_BINARY_FLAGS=0 CFLAGS_AUX="$INTEL_DFP_EXTRA_CFLAGS" CFLAGS_OPT="$INTEL_DFP_OPT_CFLAGS"
    else
        echo "오류: gcc 또는 clang이 설치되어 있지 않습니다."
        exit 1
    fi
else
    echo "지원하지 않는 운영체제: $OSTYPE"
    exit 1
fi

# 빌드된 라이브러리 확인
if [ -f "$BUILD_DIR/libbid.a" ]; then
    mkdir -p "$INTEL_DIR/lib"
    cp "$BUILD_DIR/libbid.a" "$INTEL_DIR/lib/"
    write_intel_dfp_build_stamp
    echo "Intel DFP 라이브러리 설치 완료: $INTEL_DIR/lib/libbid.a"
else
    echo "오류: 라이브러리 빌드에 실패했습니다."
    exit 1
fi
