#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
cd "$repo_root"

for cmd in go cargo javac java node npm python3 swiftc; do
    command -v "$cmd" >/dev/null || { echo "missing parser contract tool: $cmd" >&2; exit 1; }
done

work_dir=$(mktemp -d "${TMPDIR:-/tmp}/bid754-parser-contract.XXXXXXXX")
trap 'python3 -c '\''import shutil, sys; shutil.rmtree(sys.argv[1])'\'' "$work_dir"' EXIT

(
    cd bid754-codec-go
    GOCACHE="${GOCACHE:-/tmp/go-cache}" go test -count=1 -v -run '^TestParserResourceContract$' .
)
(
    cd bid754-codec-rs
    cargo test --locked --test parser_resource -- --nocapture --test-threads=1
)
javac -d "$work_dir/java" \
    bid754-codec-java/src/main/java/io/github/sky1core/bidcodec/*.java \
    bid754-codec-java/src/test/java/io/github/sky1core/bidcodec/ParserResourceRunner.java
java -cp "$work_dir/java" io.github.sky1core.bidcodec.ParserResourceRunner
(
    cd bid754-codec-js
    npm ci --no-audit --no-fund
    npm run build
    node parser_resource_test.mjs
)
PYTHONNOUSERSITE=1 python3 -B bid754-codec-py/tests/parser_resource_runner.py
swiftc -module-cache-path "$work_dir/swift-cache" \
    bid754-codec-swift/Sources/BidCodec/BidCodec.swift \
    bid754-codec-swift/Tests/ParserResourceRunner.swift -o "$work_dir/swift-parser-contract"
"$work_dir/swift-parser-contract"

python3 - "$work_dir/swift-allocation-mutant.swift" <<'PY'
from pathlib import Path
import sys

source = Path("bid754-codec-swift/Sources/BidCodec/BidCodec.swift").read_text()
target = "        var input = ASCIIParser(str)"
if source.count(target) != 1:
    raise SystemExit("Swift allocation mutation no longer matches the parser entrypoint")
mutation = "        let copiedInput = Array(str.utf8)\n        defer { withExtendedLifetime(copiedInput) {} }\n"
Path(sys.argv[1]).write_text(source.replace(target, mutation + target))
PY
swiftc -module-cache-path "$work_dir/swift-cache" \
    "$work_dir/swift-allocation-mutant.swift" \
    bid754-codec-swift/Tests/ParserResourceRunner.swift -o "$work_dir/swift-allocation-mutant"
mutation_status=0
"$work_dir/swift-allocation-mutant" --memory-probe finite > "$work_dir/swift-mutation.log" 2>&1 || mutation_status=$?
if [ "$mutation_status" -ne 1 ] || ! grep -q '^PARSER-RESOURCE FAIL swift: isolated parser RSS budget exceeded:' "$work_dir/swift-mutation.log"; then
    cat "$work_dir/swift-mutation.log"
    echo "Swift allocation mutation did not fail through the intended memory budget" >&2
    exit 1
fi
echo "PARSER-MEMORY-FAULT language=swift detection=rss-budget"
