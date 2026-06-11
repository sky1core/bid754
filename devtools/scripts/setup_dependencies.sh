#!/bin/bash
# setup_dependencies.sh - Setup C library dependencies for bid754

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/../.." && pwd )"

echo "Setting up bid754 dependencies..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)     OS_TYPE=linux;;
    Darwin*)    OS_TYPE=darwin;;
    CYGWIN*|MINGW*|MSYS*) OS_TYPE=windows;;
    *)          echo -e "${RED}Unsupported OS: $OS${NC}"; exit 1;;
esac

case "$ARCH" in
    x86_64|amd64)  ARCH_TYPE=amd64;;
    aarch64|arm64) ARCH_TYPE=arm64;;
    *)             echo -e "${RED}Unsupported architecture: $ARCH${NC}"; exit 1;;
esac

PLATFORM="${OS_TYPE}_${ARCH_TYPE}"
echo -e "${GREEN}Detected platform: $PLATFORM${NC}"

# Function to check if the native decTest decNumber prerequisite is installed
check_ibm_decnumber() {
    if [ -f "/usr/local/lib/libdecnumber.a" ] && [ -f "/usr/local/include/libdecnumber/decNumber.h" ] && [ -f "/usr/local/include/libdecnumber/dpd/decimal32.h" ]; then
        echo -e "${GREEN}✓ IBM decNumber native decTest prerequisite found${NC}"
        return 0
    elif [ -f "$HOME/local/lib/libdecnumber.a" ] && [ -f "$HOME/local/include/libdecnumber/decNumber.h" ] && [ -f "$HOME/local/include/libdecnumber/dpd/decimal32.h" ]; then
        echo -e "${GREEN}✓ IBM decNumber native decTest prerequisite found${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ IBM decNumber native decTest prerequisite not found${NC}"
        return 1
    fi
}

# Function to check if Intel DFP is ready
check_intel_dfp() {
    if [ -f "$PROJECT_ROOT/devtools/third_party/intel_dfp/lib/libbid.a" ]; then
        echo -e "${GREEN}✓ Intel DFP library found${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ Intel DFP library not found${NC}"
        return 1
    fi
}

ensure_intel_dfp_layout() {
    local intel_dir="$1"

    if [ ! -e "$intel_dir/src" ] && [ -d "$intel_dir/LIBRARY/src" ]; then
        ln -s "LIBRARY/src" "$intel_dir/src"
        echo -e "${GREEN}✓ Fixed Intel DFP layout: src -> LIBRARY/src${NC}"
    fi

    if [ ! -e "$intel_dir/include" ] && [ -d "$intel_dir/LIBRARY/float128" ]; then
        mkdir -p "$intel_dir/include"
        ln -s "../LIBRARY/float128" "$intel_dir/include/float128"
        echo -e "${GREEN}✓ Fixed Intel DFP layout: include/float128 -> ../LIBRARY/float128${NC}"
    fi

    if [ ! -e "$intel_dir/float128" ] && [ -d "$intel_dir/include/float128" ]; then
        ln -s "include/float128" "$intel_dir/float128"
        echo -e "${GREEN}✓ Fixed Intel DFP layout: float128 -> include/float128${NC}"
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

# Function to build Intel DFP
build_intel_dfp() {
    echo -e "${YELLOW}Building Intel DFP library...${NC}"
    
    INTEL_DIR="$PROJECT_ROOT/devtools/third_party/intel_dfp"

    # Extract if needed
    if [ ! -d "$INTEL_DIR/src" ]; then
        bash "$SCRIPT_DIR/setup_generation_inputs.sh" intel
    fi

    ensure_intel_dfp_layout "$INTEL_DIR"
    
    extra_cflags=""
    local opt_cflags="${INTEL_DFP_OPT_CFLAGS:--O3 -ffp-contract=off}"
    if [ "$ARCH_TYPE" = "arm64" ]; then
        extra_cflags="-DBID_SIZE_LONG=8"
        echo -e "${YELLOW}ARM64 detected: forcing BID_SIZE_LONG=8 to preserve the intended 64-bit BID build path${NC}"
    fi

    local build_dir
    build_dir="$(intel_dfp_build_dir "$INTEL_DIR")"

    # Build library
    cd "$build_dir"
    make clean >/dev/null 2>&1 || true
    case "$OS_TYPE" in
        linux)
            if command -v clang >/dev/null 2>&1; then
                make CC=clang CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 UNCHANGED_BINARY_FLAGS=0 CFLAGS_AUX="$extra_cflags" CFLAGS_OPT="$opt_cflags"
            elif command -v gcc >/dev/null 2>&1; then
                make CC=gcc CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 UNCHANGED_BINARY_FLAGS=0 CFLAGS_AUX="$extra_cflags" CFLAGS_OPT="$opt_cflags"
            else
                echo -e "${RED}No suitable compiler found${NC}"
                exit 1
            fi
            ;;
        darwin)
            make CC=clang CALL_BY_REF=0 GLOBAL_RND=0 GLOBAL_FLAGS=0 UNCHANGED_BINARY_FLAGS=0 CFLAGS_AUX="$extra_cflags" CFLAGS_OPT="$opt_cflags"
            ;;
        windows)
            echo -e "${YELLOW}Windows build requires manual setup${NC}"
            echo "Please run windowsbuild_cl.bat or windowsbuild_clang.bat"
            exit 1
            ;;
    esac
    
    # Create lib directory if it doesn't exist
    mkdir -p "$INTEL_DIR/lib"
    
    # Copy the built library
    if [ -f "$build_dir/libbid.a" ]; then
        cp "$build_dir/libbid.a" "$INTEL_DIR/lib/"
        write_intel_dfp_build_stamp "$INTEL_DIR" "$extra_cflags" "$opt_cflags"
        echo -e "${GREEN}✓ Intel DFP library built successfully${NC}"
    else
        echo -e "${RED}Failed to build Intel DFP library${NC}"
        exit 1
    fi
}

# Main setup process
echo -e "\n${YELLOW}Checking dependencies...${NC}"

# Check IBM decNumber native decTest prerequisite
if ! check_ibm_decnumber; then
	    echo -e "${YELLOW}To install the IBM decNumber native decTest prerequisite:${NC}"
	    echo "1. Run the helper:"
	    echo "   bash ./scripts/install_ibm_decnumber.sh"
	    echo "2. Manual installs must use the same pinned decNumber-icu-368.zip URL and SHA-256 as that helper."
	fi

# Setup Intel DFP
if ! check_intel_dfp; then
    build_intel_dfp
else
    INTEL_DIR="$PROJECT_ROOT/devtools/third_party/intel_dfp"
    extra_cflags=""
    opt_cflags="${INTEL_DFP_OPT_CFLAGS:--O3 -ffp-contract=off}"
    if [ "$ARCH_TYPE" = "arm64" ]; then
        extra_cflags="-DBID_SIZE_LONG=8"
    fi
    if ! intel_dfp_build_stamp_matches "$INTEL_DIR" "$extra_cflags" "$opt_cflags"; then
        echo -e "${YELLOW}⚠ Intel DFP library build flag stamp is missing or stale; rebuilding${NC}"
        build_intel_dfp
    fi
fi

# Set up CGO flags
echo -e "\n${YELLOW}Setting up environment...${NC}"

# Create environment setup script
cat > "$PROJECT_ROOT/.env.sh" << EOF
#!/bin/bash
# Auto-generated environment setup for bid754

# Optional IBM decNumber paths
if [ -d "\$HOME/local" ]; then
    export CGO_CFLAGS="-I\$HOME/local/include/libdecnumber -I\$HOME/local/include \$CGO_CFLAGS"
    export CGO_LDFLAGS="-L\$HOME/local/lib \$CGO_LDFLAGS"
elif [ -d "/usr/local/include/libdecnumber" ]; then
    export CGO_CFLAGS="-I/usr/local/include/libdecnumber -I/usr/local/include \$CGO_CFLAGS"
    export CGO_LDFLAGS="-L/usr/local/lib \$CGO_LDFLAGS"
fi

# Intel DFP paths
export CGO_CFLAGS="-I$PROJECT_ROOT/devtools/third_party/intel_dfp/include \$CGO_CFLAGS"
export CGO_LDFLAGS="-L$PROJECT_ROOT/devtools/third_party/intel_dfp/lib \$CGO_LDFLAGS"
export GOFLAGS="-tags=bid754_native \$GOFLAGS"

# Platform-specific flags
case "$(uname -s)" in
    Darwin*)
        export CGO_LDFLAGS="-framework CoreFoundation \$CGO_LDFLAGS"
        ;;
esac

echo "bid754 environment configured"
EOF

chmod +x "$PROJECT_ROOT/.env.sh"

echo -e "${GREEN}✓ Setup complete!${NC}"
echo -e "\nTo use bid754, source the environment:"
echo -e "  ${YELLOW}source .env.sh${NC}"
echo -e "\nThen run the native workflow:"
echo -e "  ${YELLOW}go test -tags bid754_native -short ./...${NC}"
echo -e "\nIf you want to pass the tag explicitly, you can also run:"
echo -e "  ${YELLOW}go test -tags bid754_native ./...${NC}"
