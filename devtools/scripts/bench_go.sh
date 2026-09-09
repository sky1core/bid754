#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

if [ "$#" -lt 3 ]; then
    echo "usage: bench_go.sh bench-native|bench-bidgo count benchtime [go test build flags...]" >&2
    exit 2
fi
target=$1
count=$2
benchtime=$3
shift 3
if [[ ! "$count" =~ ^[1-9][0-9]*$ ]]; then
    echo "benchmark count must be a positive integer" >&2
    exit 2
fi
case "$target" in
    bench-native)
        CGO_CFLAGS=${CGO_CFLAGS-}
        CGO_LDFLAGS=${CGO_LDFLAGS-}
        GOFLAGS=${GOFLAGS-}
        source ./.env.sh
        package=.
        ;;
    bench-bidgo)
        package=./internal/bidgo
        ;;
    *)
        echo "unsupported benchmark target: $target" >&2
        exit 2
        ;;
esac
export GOCACHE=${GOCACHE:-/tmp/go-cache}
tree=$(bash devtools/scripts/print_tree_id.sh)
build=$(python3 - "$target" "$@" <<'PY'
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

environment = json.loads(subprocess.check_output([
    "go", "env", "-json", "GOFLAGS", "GOEXPERIMENT", "GOAMD64", "GOARM64",
    "CGO_ENABLED", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS", "CC", "CXX",
], cwd="bid754-go"))
inputs = {"environment": environment, "args": sys.argv[2:]}
inputs["runtime"] = {name: os.environ.get(name, "") for name in ["GODEBUG", "GOGC", "GOMEMLIMIT", "GOMAXPROCS"]}
inputs["operands"] = hashlib.sha256(Path("bid754-go/testdata/benchmark_inputs.json").read_bytes()).hexdigest()
if sys.argv[1] == "bench-native":
    inputs["libbid"] = hashlib.sha256(Path("devtools/third_party/intel_dfp/lib/libbid.a").read_bytes()).hexdigest()
print(hashlib.sha256(json.dumps(inputs, sort_keys=True).encode()).hexdigest())
PY
)
printf 'BENCH-META target=%s count=%s go=%s tree=%s date=%s benchtime=%s build=%s\n' \
    "$target" "$count" "$(cd bid754-go && go env GOVERSION)" "$tree" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$benchtime" "$build"
(cd bid754-go && go test "$@" -bench=. -benchmem -count="$count" -benchtime="$benchtime" -run='^$' -timeout 1800s "$package")
if [ "$tree" != "$(bash devtools/scripts/print_tree_id.sh)" ]; then
    echo "FAIL source tree changed during benchmark measurement" >&2
    exit 1
fi
