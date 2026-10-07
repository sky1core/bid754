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
    echo "usage: $0 [all|intel|dectest|verify-intel|verify-dectest]" >&2
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

sha256_matches() {
    [ "$(sha256_file "$1")" = "$2" ]
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

# Reuse a cached archive only when it already matches its pinned sha256; a
# present-but-invalid copy is reported and deleted so the download re-runs
# (never silently validated). The checksum is verified on every path -- a cache
# hit through sha256_matches, a fresh download through verify_sha256 -- and a
# fresh download that still fails the pin is a hard error.
ensure_archive() {
    local url="$1"
    local output="$2"
    local want="$3"
    if [ -f "$output" ]; then
        if sha256_matches "$output" "$want"; then
            echo "using cached archive (sha256 verified): $output"
            return 0
        fi
        echo "cached archive failed sha256; deleting and re-downloading: $output" >&2
        rm -f "$output"
    fi
    download_file "$url" "$output"
    verify_sha256 "$output" "$want"
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
    ensure_archive "$INTEL_URL" "$INTEL_ARCHIVE" "$INTEL_SHA256"
}

intel_inputs_present() {
    [ -f "$INTEL_DIR/src/bid_conf.h" ] &&
        [ -f "$INTEL_DIR/TESTS/readtest.in" ] &&
        [ -f "$INTEL_DIR/README" ] &&
        grep -qi "$INTEL_VERSION_MARKER" "$INTEL_DIR/README"
}

validate_intel_inputs() {
    local tmpdir listing
    tmpdir="$(mktemp -d)" || return 1
    listing="$tmpdir-files"
    if ! tar -xzf "$INTEL_ARCHIVE" -C "$tmpdir"; then
        echo "failed to unpack pinned Intel DFP archive" >&2
        rm -rf "$tmpdir"
        return 1
    fi
    if ! find "$tmpdir" -type f | sort > "$listing" || [ ! -s "$listing" ]; then
        echo "failed to enumerate pinned Intel DFP inputs" >&2
        rm -rf "$tmpdir" "$listing"
        return 1
    fi

    local expected rel
    while IFS= read -r expected; do
        rel="${expected#$tmpdir/}"
        if [ ! -f "$INTEL_DIR/$rel" ]; then
            echo "Intel DFP $INTEL_VERSION input is missing from extracted tree: $rel" >&2
            rm -rf "$tmpdir" "$listing"
            return 1
        fi
        if ! cmp -s "$expected" "$INTEL_DIR/$rel"; then
            echo "Intel DFP $INTEL_VERSION input differs from pinned archive: $rel" >&2
            rm -rf "$tmpdir" "$listing"
            return 1
        fi
    done < "$listing"

    rm -rf "$tmpdir" "$listing"
    return 0
}

verify_intel_native_inputs() {
    require_tool shasum
    require_tool tar
    [ -f "$INTEL_ARCHIVE" ] || { echo "missing Intel DFP archive" >&2; return 1; }
    verify_sha256 "$INTEL_ARCHIVE" "$INTEL_SHA256"
    if ! validate_intel_inputs; then
        return 1
    fi
    if [ ! -L "$INTEL_DIR/src" ] || [ ! "$INTEL_DIR/src" -ef "$INTEL_DIR/LIBRARY/src" ]; then
        echo "Intel DFP src alias does not resolve to pinned LIBRARY/src" >&2
        return 1
    fi
    local stamp="$INTEL_DIR/lib/.libbid.build-flags"
    if [ ! -s "$INTEL_DIR/lib/libbid.a" ] || [ ! -f "$stamp" ]; then
        echo "Intel DFP native library or build stamp missing" >&2
        return 1
    fi
    local aux="" opt="-O3 -ffp-contract=off"
    if [ -n "${INTEL_DFP_OPT_CFLAGS:-}" ] && [ "$INTEL_DFP_OPT_CFLAGS" != "$opt" ]; then
        echo "Intel DFP native verification requires pinned CFLAGS_OPT=$opt" >&2
        return 1
    fi
    case "$(uname -m)" in
        arm64|aarch64) aux="-DBID_SIZE_LONG=8" ;;
    esac
    local expected actual
    expected="$(printf 'CALL_BY_REF=0\nGLOBAL_RND=0\nGLOBAL_FLAGS=0\nUNCHANGED_BINARY_FLAGS=0\nCFLAGS_AUX=%s\nCFLAGS_OPT=%s\n' "$aux" "$opt")"
    actual="$(cat "$stamp")"
    if [ "$actual" != "$expected" ]; then
        echo "Intel DFP native build stamp does not match pinned flags" >&2
        return 1
    fi
    echo "Intel DFP pinned archive, source, and native build stamp verified"
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
    ensure_archive "$DECTEST_URL" "$DECTEST_ARCHIVE" "$DECTEST_SHA256"
}

dectest_inputs_present() {
    [ -f "$DECTEST_DIR/add.decTest" ] && [ -f "$DECTEST_DIR/dqAdd.decTest" ]
}

clear_dectest_inputs() {
    mkdir -p "$DECTEST_DIR"
    find "$DECTEST_DIR" -maxdepth 1 -type f -name "*.decTest" -delete
}

dectest_file_names() {
    local path
    while IFS= read -r path; do
        printf '%s\n' "${path##*/}" || return 1
    done
}

validate_dectest_inputs() {
    local tmpdir
    tmpdir="$(mktemp -d)" || return 1

    if ! unzip -oq "$DECTEST_ARCHIVE" -d "$tmpdir"; then
        echo "failed to unpack pinned IBM decTest archive" >&2
        rm -rf "$tmpdir"
        return 1
    fi

    if ! find "$tmpdir" -type f -name "*.decTest" -print | dectest_file_names | sort >"$tmpdir/expected.list" || [ ! -s "$tmpdir/expected.list" ]; then
        echo "failed to enumerate pinned IBM decTest archive inputs" >&2
        rm -rf "$tmpdir"
        return 1
    fi
    if ! find "$DECTEST_DIR" -maxdepth 1 -type f -name "*.decTest" -print | dectest_file_names | sort >"$tmpdir/actual.list" || [ ! -s "$tmpdir/actual.list" ]; then
        echo "failed to enumerate pinned IBM decTest extracted inputs" >&2
        rm -rf "$tmpdir"
        return 1
    fi

    if ! cmp -s "$tmpdir/expected.list" "$tmpdir/actual.list"; then
        echo "IBM decTest inputs do not match the pinned 2.62 file list" >&2
        diff -u "$tmpdir/expected.list" "$tmpdir/actual.list" >&2 || true
        rm -rf "$tmpdir"
        return 1
    fi

    local name expected_file
    while IFS= read -r name; do
        if ! expected_file="$(find "$tmpdir" -type f -name "$name" -print -quit)"; then
            echo "failed to locate pinned IBM decTest input: $name" >&2
            rm -rf "$tmpdir"
            return 1
        fi
        if [ -z "$expected_file" ] || ! cmp -s "$expected_file" "$DECTEST_DIR/$name"; then
            echo "IBM decTest input differs from pinned 2.62 archive: $name" >&2
            rm -rf "$tmpdir"
            return 1
        fi
    done <"$tmpdir/expected.list"

    rm -rf "$tmpdir"
    return 0
}

verify_dectest_inputs() {
    require_tool shasum
    require_tool unzip
    [ -f "$DECTEST_ARCHIVE" ] || { echo "missing IBM decTest archive" >&2; return 1; }
    verify_sha256 "$DECTEST_ARCHIVE" "$DECTEST_SHA256"
    if ! validate_dectest_inputs; then
        return 1
    fi
    echo "IBM decTest pinned archive and source verified"
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
    if [ ! -L "$INTEL_DIR/src" ] || [ ! "$INTEL_DIR/src" -ef "$INTEL_DIR/LIBRARY/src" ]; then
        echo "Intel DFP src alias does not resolve to pinned LIBRARY/src" >&2
        return 1
    fi
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
        verify-intel)
            verify_intel_native_inputs
            ;;
        verify-dectest)
            verify_dectest_inputs
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
