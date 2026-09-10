#!/usr/bin/env bash
set -euo pipefail

usage() {
    echo "usage: $0 <portable-arm64|portable-amd64|native-amd64|digest-s390x|all>" >&2
    echo "  (all = portable arm64/amd64 and native amd64 profiles; digest-s390x is explicit)" >&2
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
            gate_cmd='bash devtools/scripts/setup_generation_inputs.sh all && make verify-profile VERIFY_PROFILE=portable VERIFY_RESULTS_BASE=/verification-results'
            ;;
        portable-amd64)
            platform=linux/amd64; arch=amd64
            gate_cmd='bash devtools/scripts/setup_generation_inputs.sh all && make verify-profile VERIFY_PROFILE=portable VERIFY_RESULTS_BASE=/verification-results'
            ;;
        native-amd64)
            platform=linux/amd64; arch=amd64
            gate_cmd='bash devtools/scripts/setup_generation_inputs.sh && bash devtools/scripts/install_ibm_decnumber.sh && bash devtools/scripts/setup_c_libs.sh && make doctor && make verify-profile VERIFY_PROFILE=native VERIFY_RESULTS_BASE=/verification-results'
            ;;
        digest-s390x)
            platform=linux/s390x; arch=s390x
            gate_cmd='make verify-profile VERIFY_PROFILE=big-endian VERIFY_RESULTS_BASE=/verification-results'
            ;;
        *)
            usage
            ;;
    esac

    local image="bid754-verify:$arch"
    echo "==> [$leg_name] building $image ($platform)"
    docker build --platform "$platform" -t "$image" "$snapshot_dir/docker-context"

    local verification_results="$repo_root/test_results/verification-linux/$snapshot_id/$leg_name"
    mkdir -p "$verification_results"
    local log="test_results/latest_linux_${leg_name}_results.txt"
    echo "==> [$leg_name] running gates in $platform container (log: $log)"
    local cargo_registry_cache="$repo_root/.build/docker-cargo-registry"
    mkdir -p "$cargo_registry_cache"
    docker run --rm -i --platform "$platform" \
            -e GIT_AUTHOR_NAME=bid754-verification \
            -e GIT_AUTHOR_EMAIL=bid754-verification@example.invalid \
            -e GIT_COMMITTER_NAME=bid754-verification \
            -e GIT_COMMITTER_EMAIL=bid754-verification@example.invalid \
            -e BID754_SNAPSHOT_ID="$snapshot_id" \
            -e BID754_SNAPSHOT_ARCHIVE=/snapshot/source.tar \
            -v "$verification_results:/verification-results" \
            -v "$snapshot_dir:/snapshot:ro" \
            -v "$repo_root/devtools/third_party/intel_dfp:/host-cache/devtools/third_party/intel_dfp:ro" \
            -v "$repo_root/devtools/third_party/ibm_decnumber:/host-cache/devtools/third_party/ibm_decnumber:ro" \
            -v "$repo_root/devtools/tests:/host-cache/devtools/tests:ro" \
            -v "$cargo_registry_cache:/root/.cargo/registry" \
            "$image" \
            bash -o pipefail -ec '
                python3 -B /snapshot/receiver.py restore - \
                    --root /work --expected-id "$BID754_SNAPSHOT_ID"
                for f in /host-cache/devtools/third_party/intel_dfp/IntelRDFPMathLib20U4.tar.gz \
                         /host-cache/devtools/third_party/ibm_decnumber/decNumber-icu-368.zip \
                         /host-cache/devtools/tests/dectest.zip; do
                    if [ -f "$f" ]; then
                        cp "$f" "/work/${f#/host-cache/}"
                    fi
                done
                cd /work
                '"$gate_cmd"'
            ' < "$snapshot_archive" 2>&1 | tee "$log"
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
    portable-arm64|portable-amd64|native-amd64|digest-s390x|all) ;;
    *) usage ;;
esac

snapshot_dir=$(mktemp -d "${TMPDIR:-/tmp}/bid754-source-snapshot.XXXXXXXX")
cleanup_snapshot() {
    python3 -c 'import shutil, sys; shutil.rmtree(sys.argv[1])' "$snapshot_dir"
}
trap cleanup_snapshot EXIT
snapshot_archive="$snapshot_dir/source.tar"
snapshot_id=$(python3 -B devtools/scripts/lib/source_snapshot.py create "$snapshot_archive" --receiver "$snapshot_dir/receiver.py")
tree_id=$(python3 -B "$snapshot_dir/receiver.py" tree-id "$snapshot_archive" --expected-id "$snapshot_id")
mkdir "$snapshot_dir/docker-context"
python3 -B "$snapshot_dir/receiver.py" extract "$snapshot_archive" \
    --expected-id "$snapshot_id" --path devtools/docker/verify/Dockerfile > "$snapshot_dir/docker-context/Dockerfile"
python3 -B "$snapshot_dir/receiver.py" extract "$snapshot_archive" \
    --expected-id "$snapshot_id" --path devtools/rust-version > "$snapshot_dir/docker-context/rust-version"
chmod 444 "$snapshot_dir/docker-context/Dockerfile" "$snapshot_dir/docker-context/rust-version"

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
