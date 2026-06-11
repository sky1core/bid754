#!/bin/bash
# setup_c_libs.sh - C 라이브러리 설정 및 빌드 스크립트

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 스크립트 경로
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

# 플랫폼 감지
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)     OS_TYPE=linux;;
    Darwin*)    OS_TYPE=darwin;;
    CYGWIN*|MINGW*|MSYS*) OS_TYPE=windows;;
    *)          echo -e "${RED}지원하지 않는 OS: $OS${NC}"; exit 1;;
esac

case "$ARCH" in
    x86_64|amd64)  ARCH_TYPE=amd64;;
    aarch64|arm64) ARCH_TYPE=arm64;;
    *)             echo -e "${RED}지원하지 않는 아키텍처: $ARCH${NC}"; exit 1;;
esac

PLATFORM="${OS_TYPE}_${ARCH_TYPE}"

echo -e "${BLUE}=== bid754 C 라이브러리 설정 ===${NC}"
echo -e "${GREEN}플랫폼: $PLATFORM${NC}\n"

# IBM decNumber 확인
check_ibm_decnumber() {
    echo -e "${BLUE}IBM decNumber 확인 중...${NC}"
    
    if [ -f "$HOME/local/lib/libdecnumber.a" ] && [ -f "$HOME/local/include/libdecnumber/decNumber.h" ] && [ -f "$HOME/local/include/libdecnumber/dpd/decimal32.h" ]; then
        echo -e "${GREEN}✓ IBM decNumber 발견: $HOME/local/lib/libdecnumber.a${NC}"
        return 0
    elif [ -f "/usr/local/lib/libdecnumber.a" ] && [ -f "/usr/local/include/libdecnumber/decNumber.h" ] && [ -f "/usr/local/include/libdecnumber/dpd/decimal32.h" ]; then
        echo -e "${GREEN}✓ IBM decNumber 발견: /usr/local/lib/libdecnumber.a${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ IBM decNumber를 찾을 수 없습니다${NC}"
        echo -e "${YELLOW}설치 방법:${NC}"
        echo "  1. 자동 설치:"
        echo "     bash ./scripts/install_ibm_decnumber.sh"
        echo "  2. 수동 설치가 필요하면 scripts/install_ibm_decnumber.sh의"
        echo "     pinned decNumber-icu-368.zip URL과 SHA-256을 그대로 사용하십시오."
        return 1
    fi
}

ensure_intel_dfp_layout() {
    local intel_dir="$1"

    if [ ! -e "$intel_dir/src" ] && [ -d "$intel_dir/LIBRARY/src" ]; then
        ln -s "LIBRARY/src" "$intel_dir/src"
        echo -e "${GREEN}✓ Intel DFP layout fixed (src -> LIBRARY/src)${NC}"
    fi

    if [ ! -e "$intel_dir/include" ] && [ -d "$intel_dir/LIBRARY/float128" ]; then
        mkdir -p "$intel_dir/include"
        ln -s "../LIBRARY/float128" "$intel_dir/include/float128"
        echo -e "${GREEN}✓ Intel DFP layout fixed (include/float128 -> ../LIBRARY/float128)${NC}"
    fi

    if [ ! -e "$intel_dir/float128" ] && [ -d "$intel_dir/include/float128" ]; then
        ln -s "include/float128" "$intel_dir/float128"
        echo -e "${GREEN}✓ Intel DFP float128 레이아웃 보정 완료 (float128 -> include/float128)${NC}"
    fi
}

intel_dfp_build_dir() {
    local intel_dir="$1"
    if [ -f "$intel_dir/LIBRARY/makefile" ]; then
        echo "$intel_dir/LIBRARY"
    else
        echo "$intel_dir"
    fi
}

intel_dfp_build_stamp_matches() {
    local intel_dir="$1"
    local extra_cflags="$2"
    local opt_cflags="$3"
    local stamp="$intel_dir/lib/.libbid.build-flags"

    [ -f "$intel_dir/lib/libbid.a" ] || return 1
    [ -f "$stamp" ] || return 1
    grep -Fxq "CALL_BY_REF=0" "$stamp" &&
        grep -Fxq "GLOBAL_RND=0" "$stamp" &&
        grep -Fxq "GLOBAL_FLAGS=0" "$stamp" &&
        grep -Fxq "UNCHANGED_BINARY_FLAGS=0" "$stamp" &&
        grep -Fxq "CFLAGS_AUX=$extra_cflags" "$stamp" &&
        grep -Fxq "CFLAGS_OPT=$opt_cflags" "$stamp"
}

write_intel_dfp_build_stamp() {
    local intel_dir="$1"
    local extra_cflags="$2"
    local opt_cflags="$3"
    local stamp="$intel_dir/lib/.libbid.build-flags"

    {
        echo "CALL_BY_REF=0"
        echo "GLOBAL_RND=0"
        echo "GLOBAL_FLAGS=0"
        echo "UNCHANGED_BINARY_FLAGS=0"
        echo "CFLAGS_AUX=$extra_cflags"
        echo "CFLAGS_OPT=$opt_cflags"
    } > "$stamp"
}

# Intel DFP 빌드
build_intel_dfp() {
    echo -e "${BLUE}Intel DFP 라이브러리 빌드 중...${NC}"
    
    INTEL_DIR="$PROJECT_ROOT/devtools/third_party/intel_dfp"

    # 소스 압축 해제
    if [ ! -d "$INTEL_DIR/src" ]; then
        bash "$SCRIPT_DIR/setup_generation_inputs.sh" intel
    fi

    ensure_intel_dfp_layout "$INTEL_DIR"
    
    local extra_cflags=""
    local opt_cflags="${INTEL_DFP_OPT_CFLAGS:--O3 -ffp-contract=off}"
    if [ "$ARCH_TYPE" = "arm64" ]; then
        extra_cflags="-DBID_SIZE_LONG=8"
        echo -e "${YELLOW}ARM64 감지: Intel DFP를 64-bit long 전제로 빌드하기 위해 CFLAGS_AUX='$extra_cflags' 적용${NC}"
    fi

    if intel_dfp_build_stamp_matches "$INTEL_DIR" "$extra_cflags" "$opt_cflags"; then
        echo -e "${GREEN}✓ Intel DFP 라이브러리가 현재 빌드 플래그로 이미 빌드되어 있습니다${NC}"
        return 0
    elif [ -f "$INTEL_DIR/lib/libbid.a" ]; then
        echo -e "${YELLOW}⚠ Intel DFP 라이브러리 빌드 플래그 스탬프가 없거나 현재 설정과 다릅니다. 재빌드합니다.${NC}"
    fi

    local build_dir
    build_dir="$(intel_dfp_build_dir "$INTEL_DIR")"

    # 플랫폼별 빌드
    cd "$build_dir"
    make clean >/dev/null 2>&1 || true
    case "$OS_TYPE" in
        linux)
            echo -e "${YELLOW}Linux에서 빌드 중...${NC}"
            if command -v clang >/dev/null 2>&1; then
                make CC=clang CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 UNCHANGED_BINARY_FLAGS=0 CFLAGS_AUX="$extra_cflags" CFLAGS_OPT="$opt_cflags"
            elif command -v gcc >/dev/null 2>&1; then
                make CC=gcc CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 UNCHANGED_BINARY_FLAGS=0 CFLAGS_AUX="$extra_cflags" CFLAGS_OPT="$opt_cflags"
            else
                echo -e "${RED}컴파일러를 찾을 수 없습니다${NC}"
                return 1
            fi
            ;;
        darwin)
            echo -e "${YELLOW}macOS에서 빌드 중...${NC}"
            make CC=clang CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 UNCHANGED_BINARY_FLAGS=0 CFLAGS_AUX="$extra_cflags" CFLAGS_OPT="$opt_cflags"
            ;;
        windows)
            echo -e "${YELLOW}Windows는 수동 빌드가 필요합니다${NC}"
            echo "다음 스크립트 중 하나를 실행하세요:"
            echo "  - windowsbuild_cl.bat (MSVC)"
            echo "  - windowsbuild_clang.bat (Clang)"
            return 1
            ;;
    esac
    
    # lib 디렉토리 생성 및 라이브러리 복사
    mkdir -p "$INTEL_DIR/lib"
    if [ -f "$build_dir/libbid.a" ]; then
        cp "$build_dir/libbid.a" "$INTEL_DIR/lib/"
        write_intel_dfp_build_stamp "$INTEL_DIR" "$extra_cflags" "$opt_cflags"
        echo -e "${GREEN}✓ Intel DFP 라이브러리 빌드 완료${NC}"
        return 0
    else
        echo -e "${RED}Intel DFP 라이브러리 빌드 실패${NC}"
        return 1
    fi
}

# 환경 설정 파일 생성
create_env_file() {
    echo -e "${BLUE}환경 설정 파일 생성 중...${NC}"
    
cat > "$PROJECT_ROOT/.env.sh" << EOF
#!/bin/bash
# bid754 빌드 환경 설정 (자동 생성됨)

# 선택적 IBM decNumber 경로
if [ -d "\$HOME/local" ]; then
    export CGO_CFLAGS="-I\$HOME/local/include/libdecnumber -I\$HOME/local/include \$CGO_CFLAGS"
    export CGO_LDFLAGS="-L\$HOME/local/lib \$CGO_LDFLAGS"
elif [ -d "/usr/local/include/libdecnumber" ]; then
    export CGO_CFLAGS="-I/usr/local/include/libdecnumber -I/usr/local/include \$CGO_CFLAGS"
    export CGO_LDFLAGS="-L/usr/local/lib \$CGO_LDFLAGS"
fi

# Intel DFP 경로
export CGO_CFLAGS="-I$PROJECT_ROOT/devtools/third_party/intel_dfp/include \$CGO_CFLAGS"
export CGO_LDFLAGS="-L$PROJECT_ROOT/devtools/third_party/intel_dfp/lib \$CGO_LDFLAGS"
export GOFLAGS="-tags=bid754_native \$GOFLAGS"

# 플랫폼별 플래그
case "\$(uname -s)" in
    Darwin*)
        export CGO_LDFLAGS="-framework CoreFoundation \$CGO_LDFLAGS"
        ;;
esac

echo "bid754 환경 설정 완료"
EOF

    chmod +x "$PROJECT_ROOT/.env.sh"
    echo -e "${GREEN}✓ 환경 설정 파일 생성 완료: .env.sh${NC}"
}

# 메인 실행
main() {
    echo -e "${BLUE}1단계: IBM decNumber 확인${NC}"
    if check_ibm_decnumber; then
        IBM_STATUS=0
    else
        IBM_STATUS=1
    fi
    
    echo -e "\n${BLUE}2단계: Intel DFP 설정${NC}"
    # pinned Intel BID source tree 에서만 libbid.a 를 빌드한다.
    build_intel_dfp
    INTEL_STATUS=$?
    
    echo -e "\n${BLUE}3단계: 환경 설정${NC}"
    create_env_file
    
    # 최종 상태 출력
    echo -e "\n${BLUE}=== 설정 완료 ===${NC}"
    
    if [ $IBM_STATUS -eq 0 ]; then
        echo -e "${GREEN}✓ IBM decNumber: 준비됨${NC}"
    else
        echo -e "${YELLOW}⚠ IBM decNumber: 수동 설치 필요${NC}"
    fi
    
    if [ $INTEL_STATUS -eq 0 ]; then
        echo -e "${GREEN}✓ Intel DFP: 준비됨${NC}"
    else
        echo -e "${YELLOW}⚠ Intel DFP: 수동 빌드 필요${NC}"
    fi
    
echo -e "\n${YELLOW}다음 단계:${NC}"
    echo "1. 환경 설정 로드:"
    echo "   source .env.sh"
    echo "2. 기본 빌드/테스트:"
    echo "   go test ./..."
    echo "3. 실제 native CGO 백엔드 테스트(선택):"
    echo "   go test -tags bid754_native -short ./..."

    if [ $INTEL_STATUS -ne 0 ]; then
        echo -e "\n${RED}Intel DFP가 준비되지 않아 네이티브 워크플로를 실행할 수 없습니다.${NC}"
        exit 1
    fi

    if [ $IBM_STATUS -ne 0 ]; then
        echo -e "\n${YELLOW}주의: IBM decNumber가 없으면 native decTest 검증 일부를 실행할 수 없습니다.${NC}"
    fi
}

# 스크립트 실행
main
