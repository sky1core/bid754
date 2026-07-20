#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
go_cache=${GOCACHE:-/tmp/go-cache}
java_out=""
py_venv=""

cleanup_bidcodec_python_artifacts() {
  rm -rf "$repo_root/bid754-codec-py/.pytest_cache"
  find "$repo_root/bid754-codec-py" -type d -name __pycache__ -prune -exec rm -rf {} +
}

cleanup() {
  cleanup_bidcodec_python_artifacts
  if [ -n "$java_out" ]; then
    rm -rf "$java_out"
  fi
  if [ -n "$py_venv" ]; then
    rm -rf "$py_venv"
  fi
}
trap cleanup EXIT

if [ -d /opt/homebrew/opt/openjdk/bin ]; then
  export PATH="/opt/homebrew/opt/openjdk/bin:$PATH"
fi

cd "$repo_root"

require_file() {
  local path=$1
  if [ ! -f "$path" ]; then
    echo "missing required BID codec vector consumer: $path" >&2
    exit 1
  fi
}

require_cmd() {
  local cmd=$1
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "missing required command for BID codec verification: $cmd" >&2
    exit 1
  fi
}

require_vector_reference() {
  local path=$1
  if ! rg -q 'bid754-codec-vectors' "$path"; then
    echo "BID codec vector consumer does not read generated vectors: $path" >&2
    exit 1
  fi
  if ! rg -q 'decimal_string|DecimalString' "$path"; then
    echo "BID codec vector consumer does not verify generated decimal_string: $path" >&2
    exit 1
  fi
}

# The go_full consumer exercises the bid754-go public parse surface against the
# reject_vectors from_string channel and the string_vectors channel only, so it
# is checked for those channel references instead of the decimal_string field
# the raw `vectors` channel consumers verify.
require_go_full_reference() {
  local path=$1
  if ! rg -q 'bid754-codec-vectors' "$path"; then
    echo "go_full BID codec consumer does not read generated vectors: $path" >&2
    exit 1
  fi
  if ! rg -q 'reject_vectors' "$path" || ! rg -q 'string_vectors' "$path"; then
    echo "go_full BID codec consumer does not reference the reject/string channels: $path" >&2
    exit 1
  fi
}

make verify-generated
(cd devtools && GOCACHE="$go_cache" go test -count=1 ./internal/testgen -run TestBidCodecVectorGeneratorDoesNotImportBidCodecUnderTest)

require_cmd rg
require_cmd go
require_cmd cargo
require_cmd javac
require_cmd java
require_cmd python3
require_cmd npm
require_cmd swift

required_consumers=(
  "bid754-codec-go/vector_test.go"
  "bid754-codec-rs/tests/vectors.rs"
  "bid754-rs/tests/bid_codec_vectors.rs"
  "bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorTest.java"
  "bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorRunner.java"
  "bid754-codec-py/tests/test_vectors.py"
  "bid754-codec-js/src/vectors.test.ts"
  "bid754-codec-swift/Sources/BidCodecVectorRunner/main.swift"
)

for consumer in "${required_consumers[@]}"; do
  require_file "$consumer"
  require_vector_reference "$consumer"
done

require_file "bid754-go/generated_bid_codec_vectors_test.go"
require_go_full_reference "bid754-go/generated_bid_codec_vectors_test.go"

echo "==> Go BID codec vector tests: bid754-codec-go"
(cd bid754-codec-go && GOCACHE="$go_cache" go test -count=1 -tags bid754_bidcodec_vectors ./...)

echo "==> Rust BID codec vector tests: bid754-codec-rs"
(cd bid754-codec-rs && cargo test --locked)

echo "==> Rust bid754 BID codec vector tests: bid754-rs"
(cd bid754-rs && cargo test --locked --test bid_codec_vectors)

echo "==> Go bid754 public parse BID codec vector tests: bid754-go"
# A -run expression that matches zero tests (or silently drops one of the two
# generated tests) still exits 0, so this leg must fail on the "test not
# executed" failure mode itself instead of trusting the name filter: require
# an explicit PASS line for each generated go_full test in the verbose output.
# No pipe into grep here: pipefail would surface grep's SIGPIPE otherwise.
go_full_out=$(cd bid754-go && GOCACHE="$go_cache" go test -count=1 -v -tags bid754_bidcodec_vectors -run '^TestGoFullBidCodec' . 2>&1) || {
  printf '%s\n' "$go_full_out"
  exit 1
}
printf '%s\n' "$go_full_out"
for go_full_test in TestGoFullBidCodecRejectVectors TestGoFullBidCodecStringVectors; do
  if ! grep -qF -- "--- PASS: $go_full_test" <<<"$go_full_out"; then
    echo "go_full BID codec consumer did not execute $go_full_test" >&2
    exit 1
  fi
done

echo "==> Rust bid754 public parse BID codec vector tests: bid754-rs"
# Same "test not executed" guard as the go_full leg above: cargo exits 0 when a
# filter matches nothing, so require the explicit PASS line. This leg is the
# only one that reaches the generated Rust public parse path -- rust_full next
# door exercises the embedded Components codec -- so a silently skipped run
# would leave that surface unverified while the gate reported success.
rust_parse_out=$(cd bid754-rs && cargo test --locked --test bid_codec_parse_vectors 2>&1) || {
  printf '%s\n' "$rust_parse_out"
  exit 1
}
printf '%s\n' "$rust_parse_out"
if ! grep -qF -- "test test_rust_full_parse_reject_vectors ... ok" <<<"$rust_parse_out"; then
  echo "rust_full_parse BID codec consumer did not execute test_rust_full_parse_reject_vectors" >&2
  exit 1
fi

echo "==> Java BID codec vector tests: bid754-codec-java"
java_out=$(mktemp -d)
javac -d "$java_out" \
  bid754-codec-java/src/main/java/io/github/sky1core/bidcodec/*.java \
  bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/VectorRunner.java
java -cp "$java_out" io.github.sky1core.bidcodec.VectorRunner bid754-codec-vectors/vectors.json

echo "==> Python BID codec vector tests: bid754-codec-py"
py_venv=$(mktemp -d)
python3 -m venv "$py_venv"
"$py_venv/bin/python" -m pip install "pytest==9.0.2"
(cd bid754-codec-py && PYTHONNOUSERSITE=1 "$py_venv/bin/python" -m pytest)

echo "==> JavaScript/TypeScript BID codec vector tests: bid754-codec-js"
(cd bid754-codec-js && npm ci && npm run build && npm test)

echo "==> Swift BID codec vector tests: bid754-codec-swift"
swift run BidCodecVectorRunner bid754-codec-vectors/vectors.json
