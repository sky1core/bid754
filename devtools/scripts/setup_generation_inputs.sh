#!/bin/bash
# Download and unpack authoritative generator inputs.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

INTEL_VERSION="v20U4"
INTEL_URL="https://www.netlib.org/misc/intel/IntelRDFPMathLib20U4.tar.gz"
INTEL_SHA256="1df86132e7a31fd74d784fee1c679b21a088f73a8ec979cfaf784c200392e125"
INTEL_ARCHIVE="$PROJECT_ROOT/devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz"
INTEL_VERSION_MARKER="release 2.0 Update 4"
INTEL_DIR="$PROJECT_ROOT/devtools/third_party/intel_dfp"

DECTEST_URL="https://speleotrove.com/decimal/dectest.zip"
DECTEST_SHA256="b70a224cd52e82b7a8150aedac5efa2d0cb3941696fd829bdbe674f9f65c3926"
DECTEST_ARCHIVE="$PROJECT_ROOT/devtools/tests/dectest.zip"
DECTEST_DIR="$PROJECT_ROOT/devtools/tests"

usage() {
    echo "usage: $0 [all|intel|dectest]" >&2
}

require_tool() {
    if ! command -v "$1" >/dev/null 2>&1; then
        echo "missing required tool: $1" >&2
        exit 1
    fi
}

sha256_file() {
    shasum -a 256 "$1" | awk '{print $1}'
}

download_file() {
    local url="$1"
    local output="$2"
    mkdir -p "$(dirname "$output")"
    echo "downloading $url"
    # netlib.org/speleotrove.com intermittently time out in CI; retry with
    # curl's default exponential backoff instead of failing on first connect.
    curl -L --fail --show-error --silent \
        --retry 5 --retry-connrefused --connect-timeout 30 \
        "$url" -o "$output"
}

verify_sha256() {
    local path="$1"
    local want="$2"
    local got
    got="$(sha256_file "$path")"
    if [ "$got" != "$want" ]; then
        echo "checksum mismatch for $path" >&2
        echo "  got:  $got" >&2
        echo "  want: $want" >&2
        exit 1
    fi
}

normalize_intel_layout() {
    if [ ! -e "$INTEL_DIR/src" ] && [ -d "$INTEL_DIR/LIBRARY/src" ]; then
        ln -s "LIBRARY/src" "$INTEL_DIR/src"
        echo "fixed Intel DFP layout: src -> LIBRARY/src"
    fi

    if [ ! -e "$INTEL_DIR/include" ] && [ -d "$INTEL_DIR/LIBRARY/float128" ]; then
        mkdir -p "$INTEL_DIR/include"
        ln -s "../LIBRARY/float128" "$INTEL_DIR/include/float128"
        echo "fixed Intel DFP layout: include/float128 -> ../LIBRARY/float128"
    fi

    if [ ! -e "$INTEL_DIR/float128" ] && [ -d "$INTEL_DIR/include/float128" ]; then
        ln -s "include/float128" "$INTEL_DIR/float128"
        echo "fixed Intel DFP layout: float128 -> include/float128"
    fi
}

ensure_intel_archive() {
    if [ ! -f "$INTEL_ARCHIVE" ]; then
        download_file "$INTEL_URL" "$INTEL_ARCHIVE"
    fi
    verify_sha256 "$INTEL_ARCHIVE" "$INTEL_SHA256"
}

intel_inputs_present() {
    [ -f "$INTEL_DIR/src/bid_conf.h" ] &&
        [ -f "$INTEL_DIR/TESTS/readtest.in" ] &&
        [ -f "$INTEL_DIR/README" ] &&
        grep -qi "$INTEL_VERSION_MARKER" "$INTEL_DIR/README"
}

validate_intel_inputs() {
    local tmpdir
    tmpdir="$(mktemp -d)"

    tar -xzf "$INTEL_ARCHIVE" -C "$tmpdir"

    local expected rel
    while IFS= read -r expected; do
        rel="${expected#$tmpdir/}"
        if [ ! -f "$INTEL_DIR/$rel" ]; then
            echo "Intel DFP $INTEL_VERSION input is missing from extracted tree: $rel" >&2
            rm -rf "$tmpdir"
            return 1
        fi
        if ! cmp -s "$expected" "$INTEL_DIR/$rel"; then
            echo "Intel DFP $INTEL_VERSION input differs from pinned archive: $rel" >&2
            rm -rf "$tmpdir"
            return 1
        fi
    done < <(find "$tmpdir" -type f | sort)

    rm -rf "$tmpdir"
    return 0
}

clear_intel_inputs() {
    mkdir -p "$INTEL_DIR"
    find "$INTEL_DIR" -mindepth 1 -maxdepth 1 \
        ! -name ".gitkeep" \
        ! -name "README.md" \
        ! -name "download.sh" \
        ! -name "$(basename "$INTEL_ARCHIVE")" \
        -exec rm -rf {} +
}

ensure_dectest_archive() {
    if [ ! -f "$DECTEST_ARCHIVE" ]; then
        download_file "$DECTEST_URL" "$DECTEST_ARCHIVE"
    fi
    verify_sha256 "$DECTEST_ARCHIVE" "$DECTEST_SHA256"
}

dectest_inputs_present() {
    [ -f "$DECTEST_DIR/add.decTest" ] && [ -f "$DECTEST_DIR/dqAdd.decTest" ]
}

clear_dectest_inputs() {
    mkdir -p "$DECTEST_DIR"
    find "$DECTEST_DIR" -maxdepth 1 -type f -name "*.decTest" -delete
}

validate_dectest_inputs() {
    local tmpdir
    tmpdir="$(mktemp -d)"

    unzip -oq "$DECTEST_ARCHIVE" -d "$tmpdir"

    find "$tmpdir" -type f -name "*.decTest" -exec basename {} \; | sort >"$tmpdir/expected.list"
    find "$DECTEST_DIR" -maxdepth 1 -type f -name "*.decTest" -exec basename {} \; | sort >"$tmpdir/actual.list"

    if ! cmp -s "$tmpdir/expected.list" "$tmpdir/actual.list"; then
        echo "IBM decTest inputs do not match the pinned 2.62 file list" >&2
        diff -u "$tmpdir/expected.list" "$tmpdir/actual.list" >&2 || true
        rm -rf "$tmpdir"
        return 1
    fi

    local name expected_file
    while IFS= read -r name; do
        expected_file="$(find "$tmpdir" -type f -name "$name" -print -quit)"
        if [ -z "$expected_file" ] || ! cmp -s "$expected_file" "$DECTEST_DIR/$name"; then
            echo "IBM decTest input differs from pinned 2.62 archive: $name" >&2
            rm -rf "$tmpdir"
            return 1
        fi
    done <"$tmpdir/expected.list"

    rm -rf "$tmpdir"
    return 0
}

ensure_intel_dfp() {
    require_tool curl
    require_tool shasum
    require_tool tar

    ensure_intel_archive

    if intel_inputs_present && validate_intel_inputs; then
        echo "Intel DFP $INTEL_VERSION inputs already present and verified"
    else
        if [ -f "$INTEL_DIR/src/bid_conf.h" ] || [ -f "$INTEL_DIR/TESTS/readtest.in" ]; then
            echo "removing stale Intel DFP inputs before extracting $INTEL_VERSION"
            clear_intel_inputs
        fi
        echo "extracting Intel DFP $INTEL_VERSION inputs"
        mkdir -p "$INTEL_DIR"
        tar -xzf "$INTEL_ARCHIVE" -C "$INTEL_DIR"
    fi

    normalize_intel_layout
}

ensure_dectest() {
    require_tool curl
    require_tool shasum
    require_tool unzip

    ensure_dectest_archive

    if dectest_inputs_present && validate_dectest_inputs; then
        echo "IBM decTest 2.62 inputs already present and verified"
        return
    fi

    if dectest_inputs_present; then
        echo "removing stale IBM decTest inputs before extracting pinned 2.62"
        clear_dectest_inputs
    fi

    echo "extracting IBM decTest 2.62 inputs"
    mkdir -p "$DECTEST_DIR"
    unzip -oq "$DECTEST_ARCHIVE" -d "$DECTEST_DIR"
    validate_dectest_inputs
}

main() {
    local target="${1:-all}"
    case "$target" in
        all)
            ensure_intel_dfp
            ensure_dectest
            ;;
        intel)
            ensure_intel_dfp
            ;;
        dectest)
            ensure_dectest
            ;;
        *)
            usage
            exit 2
            ;;
    esac
}

main "$@"
