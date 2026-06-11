#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
go_cache=${GOCACHE:-/tmp/go-cache}
java_out=""
py_venv=""

cleanup_bidcodec_python_artifacts() {
  rm -rf "$repo_root/bid-codec-py/.pytest_cache"
  find "$repo_root/bid-codec-py" -type d -name __pycache__ -prune -exec rm -rf {} +
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
  if ! rg -q 'bid-codec-vectors' "$path"; then
    echo "BID codec vector consumer does not read generated vectors: $path" >&2
    exit 1
  fi
  if ! rg -q 'decimal_string|DecimalString' "$path"; then
    echo "BID codec vector consumer does not verify generated decimal_string: $path" >&2
    exit 1
  fi
}

make verify-generated
GOCACHE="$go_cache" go test ./internal/testgen -run TestBidCodecVectorGeneratorDoesNotImportBidCodecUnderTest
bash ./scripts/audit_bidcodec_payload_scope.sh

require_cmd rg
require_cmd go
require_cmd cargo
require_cmd javac
require_cmd java
require_cmd python3
require_cmd npm
require_cmd swift

required_consumers=(
  "bidcodec/vector_test.go"
  "bid-codec-rs/tests/vectors.rs"
  "bid754-rs/tests/bid_codec_vectors.rs"
  "bid-codec-java/src/test/java/dev/bid754/bidcodec/VectorTest.java"
  "bid-codec-java/src/test/java/dev/bid754/bidcodec/VectorRunner.java"
  "bid-codec-py/tests/test_vectors.py"
  "bid-codec-js/src/vectors.test.ts"
  "bid-codec-swift/Sources/BidCodecVectorRunner/main.swift"
)

for consumer in "${required_consumers[@]}"; do
  require_file "$consumer"
  require_vector_reference "$consumer"
done

echo "==> Go BID codec vector tests: bidcodec"
(cd bidcodec && GOCACHE="$go_cache" go test -tags bid754_bidcodec_vectors ./...)

echo "==> Rust BID codec vector tests: bid-codec-rs"
(cd bid-codec-rs && cargo test --locked)

echo "==> Rust bid754 BID codec vector tests: bid754-rs"
(cd bid754-rs && cargo test --locked --test bid_codec_vectors)

echo "==> Java BID codec vector tests: bid-codec-java"
java_out=$(mktemp -d)
javac -d "$java_out" \
  bid-codec-java/src/main/java/dev/bid754/bidcodec/*.java \
  bid-codec-java/src/test/java/dev/bid754/bidcodec/VectorRunner.java
java -cp "$java_out" dev.bid754.bidcodec.VectorRunner bid-codec-vectors/vectors.json

echo "==> Python BID codec vector tests: bid-codec-py"
py_venv=$(mktemp -d)
python3 -m venv "$py_venv"
"$py_venv/bin/python" -m pip install "pytest==9.0.2"
(cd bid-codec-py && PYTHONNOUSERSITE=1 "$py_venv/bin/python" -m pytest)

echo "==> JavaScript/TypeScript BID codec vector tests: bid-codec-js"
(cd bid-codec-js && npm ci && npm run build && npm test)

echo "==> Swift BID codec vector tests: bid-codec-swift"
(cd bid-codec-swift && swift run BidCodecVectorRunner ../bid-codec-vectors/vectors.json)
