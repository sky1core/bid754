#!/usr/bin/env bash
# verify_linux.sh - run the Linux verification legs locally in Docker, without
# GitHub Actions. The legs mirror the Linux jobs of the CI workflow:
#
#   portable-arm64   linux/arm64   make test-go-modules && make test-rust
#   portable-amd64   linux/amd64   make test-go-modules && make test-rust
#   native-amd64     linux/amd64   IBM decNumber + Intel BID C oracle build,
#                                  then make doctor, the native smoke/FFI
#                                  bit-compare/readtest/decTest gates, and the
#                                  Rust Tier 1 long exact and readtest gates
#
# One extra leg has no CI counterpart and is not part of "all":
#
#   digest-s390x     linux/s390x   big-endian regression leg under qemu:
#                                  make digest (cross-endian bit identity of
#                                  the core-op digest, compared by make
#                                  verify-digest), the bid754-codec-go module
#                                  tests (byte-level codec contract), and the
#                                  two generated Go-port runners (readtest
#                                  goport + public-API parity, full corpus).
#                                  The full portable gate set is intentionally
#                                  not run: qemu-s390x emulation makes it
#                                  impractically slow for a routine leg.
#
# Intentionally not run here (platform-independent; covered on the host by
# make verify-all): verify-generated, the BID codec 6-language consumers, and
# check-scripts.
#
# The repository working tree is injected as a tar stream of tracked plus
# untracked-but-not-ignored files, so host build artifacts (.env.sh, macOS
# libbid.a, test_results/, caches) never enter the container. Pinned
# upstream archives already cached under devtools/third_party/ are copied in when
# present; otherwise the setup scripts download them against pinned SHA-256.
set -euo pipefail

usage() {
    echo "usage: $0 <portable-arm64|portable-amd64|native-amd64|digest-s390x|all>" >&2
    echo "  (all = the three CI-mirroring legs; digest-s390x is run explicitly)" >&2
    exit 2
}

[ $# -eq 1 ] || usage

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
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
            gate_cmd='bash devtools/scripts/setup_generation_inputs.sh all && make test-go-modules && make test-rust && make digest'
            ;;
        portable-amd64)
            platform=linux/amd64; arch=amd64
            gate_cmd='bash devtools/scripts/setup_generation_inputs.sh all && make test-go-modules && make test-rust && make digest'
            ;;
        native-amd64)
            platform=linux/amd64; arch=amd64
            # setup_generation_inputs.sh prepares both pinned inputs: the Intel
            # BID C sources (cgen/csymbols sync tests inside the -short run)
            # and the IBM decTest originals that the native decTest gate
            # parses next to the IBM decNumber oracle.
            gate_cmd='bash devtools/scripts/setup_generation_inputs.sh && bash devtools/scripts/install_ibm_decnumber.sh && bash devtools/scripts/setup_c_libs.sh && make doctor && make test-native-smoke && make test-native-ffi && make _test-native-tier1-arithmetic-long-full && make _test-native-tier1-compare-conversion-long-full && make _test-rust-native-tier1-arithmetic-long-full && make _test-rust-native-tier1-compare-conversion-long-full && make test-native-readtest && make test-native-dectest && make test-rust-native'
            ;;
        digest-s390x)
            platform=linux/s390x; arch=s390x
            # Local-only big-endian regression leg (no CI counterpart, so it
            # is not part of "all"). make digest emits the PLATFORM-DIGEST
            # line captured below as test_results/digest_linux_s390x.txt and
            # compared across platforms by make verify-digest; the
            # bid754-codec-go module tests close the byte-level BID codec
            # contract on a big-endian host; the two generated Go-port
            # runners (readtest goport dispatch and public-API parity, full
            # corpus, anchored -run match) close the 128-bit byte-image <->
            # word conversion glue that a native-endian reinterpretation
            # would byte-swap here. All gates need only the tracked tree
            # (generated testspec and codec vectors are checked in), so no
            # pinned-input setup runs, and the full test-go-modules gate is
            # intentionally excluded: under qemu-s390x emulation it is
            # impractically slow for a routine leg. GOCACHE is pinned to the
            # same path the Makefile GOENV uses so the digest and runner
            # builds share one in-container build cache.
            gate_cmd='make digest && (cd bid754-codec-go && GOCACHE="${GOCACHE:-/tmp/go-cache}" go test -count=1 ./...) && (cd bid754-go && GOCACHE="${GOCACHE:-/tmp/go-cache}" go test -count=1 -run "^(TestGeneratedReadCasesGoPort|TestGeneratedPublicAPIParity)$" .)'
            ;;
        *)
            usage
            ;;
    esac

    local image="bid754-verify:$arch"
    echo "==> [$leg_name] building $image ($platform)"
    docker build --platform "$platform" -t "$image" devtools/docker/verify

    mkdir -p test_results
    local log="test_results/latest_linux_${leg_name}_results.txt"
    # Tree id is computed on the host at (near) tar-snapshot time: the container
    # index has no commits, so the host stamps the digest file instead.
    local tree_id
    tree_id=$(bash devtools/scripts/print_tree_id.sh)
    echo "==> [$leg_name] running gates in $platform container (log: $log)"
    local cargo_registry_cache="$repo_root/.build/docker-cargo-registry"
    local tracked_index_rel=".build/verify-linux-tracked-index-${leg_name}"
    local tracked_index="$repo_root/$tracked_index_rel"
    mkdir -p "$cargo_registry_cache"
    # The injected tree deliberately omits the host .git directory, but the
    # generated-artifact checks require git grep/ls-files. Carry
    # the exact host stage entries as a NUL-delimited sidecar, then rebuild only
    # the index in the container. Using index-info preserves literal paths and
    # file/directory type changes without staging the untracked-but-not-ignored
    # work that the tar stream also carries.
    git ls-files --stage -z > "$tracked_index"
    # COPYFILE_DISABLE stops macOS bsdtar from adding AppleDouble (._*)
    # metadata entries, which would land as stale files in the container tree.
    # The .git exclusion also prevents an untracked nested repository emitted
    # as a directory by git ls-files from contributing repository metadata.
    { git ls-files -coz --exclude-standard; printf '%s\0' "$tracked_index_rel"; } | \
        COPYFILE_DISABLE=1 tar --exclude='.git' --null -T - -cf - | \
        docker run --rm -i --platform "$platform" \
            -e BID754_TRACKED_INDEX_REL="$tracked_index_rel" \
            -v "$repo_root/devtools/third_party/intel_dfp:/host-cache/devtools/third_party/intel_dfp:ro" \
            -v "$repo_root/devtools/third_party/ibm_decnumber:/host-cache/devtools/third_party/ibm_decnumber:ro" \
            -v "$repo_root/devtools/tests:/host-cache/devtools/tests:ro" \
            -v "$cargo_registry_cache:/root/.cargo/registry" \
            "$image" \
            bash -o pipefail -ec '
                tar -xf - -C /work
                git -C /work init -q
                git -C /work update-index -z --index-info \
                    < "/work/$BID754_TRACKED_INDEX_REL"
                if ! git -C /work ls-files --stage -z | \
                    cmp -s "/work/$BID754_TRACKED_INDEX_REL" -; then
                    echo "synthetic Git index differs from host index" >&2
                    exit 1
                fi
                for f in /host-cache/devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz \
                         /host-cache/devtools/third_party/ibm_decnumber/decNumber-icu-368.zip \
                         /host-cache/devtools/tests/dectest.zip; do
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
        {
            printf 'PLATFORM-DIGEST-TREE %s\n' "$tree_id"
            printf '%s\n' "$digest_line"
        } > "test_results/digest_linux_${arch}.txt"
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
