#!/usr/bin/env bash
# verify_linux.sh - run the Linux verification legs locally in Docker, without
# GitHub Actions. The legs mirror the Linux jobs of the CI workflow:
#
#   portable-arm64   linux/arm64   make test-go-modules && make test-rust
#   portable-amd64   linux/amd64   make test-go-modules && make test-rust
#   native-amd64     linux/amd64   IBM decNumber + Intel BID C oracle build,
#                                  then make doctor, the native smoke/FFI
#                                  bit-compare/readtest/decTest gates, and the
#                                  Rust native readtest gate
#
# Intentionally not run here (platform-independent; covered on the host by
# make full-audit): verify-generated, the BID codec 6-language consumers, and
# check-scripts.
#
# The repository working tree is injected as a tar stream of tracked plus
# untracked-but-not-ignored files, so host build artifacts (.env.sh, macOS
# libbid.a, test_results/, caches) never leak into the container. Pinned
# upstream archives already cached under third_party/ are copied in when
# present; otherwise the setup scripts download them against pinned SHA-256.
set -euo pipefail

usage() {
    echo "usage: $0 <portable-arm64|portable-amd64|native-amd64|all>" >&2
    exit 2
}

[ $# -eq 1 ] || usage

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

run_leg() {
    local leg_name="$1" platform arch gate_cmd
    case "$leg_name" in
        portable-arm64)
            platform=linux/arm64; arch=arm64
            # internal/cgen and internal/csymbols verify generated artifacts
            # against the extracted Intel BID C sources, so prepare the pinned
            # generation inputs first (the CI portable-test matrix job gets
            # them from its preceding verify-generated step; the CI arm64
            # portable job has no such step). make digest emits the
            # PLATFORM-DIGEST line consumed by make verify-digest.
            gate_cmd='bash scripts/setup_generation_inputs.sh intel && make test-go-modules && make test-rust && make digest'
            ;;
        portable-amd64)
            platform=linux/amd64; arch=amd64
            gate_cmd='bash scripts/setup_generation_inputs.sh intel && make test-go-modules && make test-rust && make digest'
            ;;
        native-amd64)
            platform=linux/amd64; arch=amd64
            # setup_generation_inputs.sh prepares both pinned inputs: the Intel
            # BID C sources (cgen/csymbols sync tests inside the -short run)
            # and the IBM decTest originals that the native decTest gate
            # parses next to the IBM decNumber oracle.
            gate_cmd='bash scripts/setup_generation_inputs.sh && bash scripts/install_ibm_decnumber.sh && bash scripts/setup_c_libs.sh && make doctor && make test-native-smoke && make test-native-ffi && make test-native-readtest && make test-native-dectest && make test-rust-native'
            ;;
        *)
            usage
            ;;
    esac

    local image="bid754-verify:$arch"
    echo "==> [$leg_name] building $image ($platform)"
    docker build --platform "$platform" -t "$image" docker/verify

    mkdir -p test_results
    local log="test_results/latest_linux_${leg_name}_results.txt"
    echo "==> [$leg_name] running gates in $platform container (log: $log)"
    # COPYFILE_DISABLE stops macOS bsdtar from adding AppleDouble (._*)
    # metadata entries, which would land as stale files in the container tree.
    git ls-files -coz --exclude-standard | COPYFILE_DISABLE=1 tar --null -T - -cf - | \
        docker run --rm -i --platform "$platform" \
            -v "$repo_root/third_party/intel_dfp:/host-cache/third_party/intel_dfp:ro" \
            -v "$repo_root/third_party/ibm_decnumber:/host-cache/third_party/ibm_decnumber:ro" \
            -v "$repo_root/tests:/host-cache/tests:ro" \
            -v bid754-cargo-registry:/root/.cargo/registry \
            "$image" \
            bash -o pipefail -ec '
                tar -xf - -C /work
                for f in /host-cache/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz \
                         /host-cache/third_party/ibm_decnumber/decNumber-icu-368.zip \
                         /host-cache/tests/dectest.zip; do
                    if [ -f "$f" ]; then
                        cp "$f" "/work/${f#/host-cache/}"
                    fi
                done
                cd /work
                '"$gate_cmd"'
            ' 2>&1 | tee "$log"
    # The portable legs emit a PLATFORM-DIGEST line; persist it for
    # make verify-digest (PLATFORM_SPEC section 4 item 2).
    digest_line=$(grep '^PLATFORM-DIGEST ' "$log" | tail -1 || true)
    if [ -n "$digest_line" ]; then
        printf '%s\n' "$digest_line" > "test_results/digest_linux_${arch}.txt"
        echo "==> [$leg_name] digest captured: test_results/digest_linux_${arch}.txt"
    fi
    echo "==> [$leg_name] PASS"
}

case "$1" in
    all)
        run_leg portable-arm64
        run_leg portable-amd64
        run_leg native-amd64
        ;;
    *)
        run_leg "$1"
        ;;
esac
