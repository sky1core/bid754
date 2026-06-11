#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)
go_cache=${GOCACHE:-/tmp/go-cache}
wheel_dir=$(mktemp -d)
audit_tmp=$(mktemp -d)

cleanup_bidcodec_python_artifacts() {
  rm -rf "$repo_root/bid754-codec-py/.pytest_cache"
  find "$repo_root/bid754-codec-py" -type d -name __pycache__ -prune -exec rm -rf {} +
}

cleanup() {
  cleanup_bidcodec_python_artifacts
  rm -rf "$wheel_dir"
  rm -rf "$audit_tmp"
}
trap cleanup EXIT

if [ -d /opt/homebrew/opt/openjdk/bin ]; then
  export PATH="/opt/homebrew/opt/openjdk/bin:$PATH"
fi

cd "$repo_root"

vectors_path="$repo_root/bid754-codec-vectors/vectors.json"
audit_vectors_dir="$audit_tmp/bid754-codec-vectors"
mkdir -p "$audit_vectors_dir"
cp "$vectors_path" "$audit_vectors_dir/vectors.json"

echo "==> BID codec package license/provenance files"
test -f THIRD_PARTY_NOTICES.md
for package_dir in bid754-codec-go bid754-codec-rs bid754-codec-java bid754-codec-py bid754-codec-js bid754-codec-swift; do
  if [ ! -f "$package_dir/LICENSE" ]; then
    echo "missing package-local LICENSE: $package_dir/LICENSE" >&2
    exit 1
  fi
done

echo "==> Go BID codec package tests"
(cd bid754-codec-go && GOCACHE="$go_cache" go test ./...)
(cd bid754-codec-go && GOCACHE="$go_cache" go test -tags bid754_bidcodec_vectors ./...)
go_module_version="v0.1.0"
# The bid754-codec-go module lives in a subdirectory of the repository, so
# module resolution follows the Go multi-module convention: the VCS root is
# the repository URL and the module version tag is prefixed with the module
# subdirectory ("bid754-codec-go/v0.1.0"). Mirror that exact layout here.
go_release="$audit_tmp/go-bidcodec-release"
mkdir -p "$go_release/bid754-codec-go"
cp bid754-codec-go/LICENSE bid754-codec-go/README.md bid754-codec-go/go.mod bid754-codec-go/*.go "$go_release/bid754-codec-go/"
git -C "$go_release" init -q
git -C "$go_release" add .
git -C "$go_release" -c user.name=bid754-audit -c user.email=bid754-audit@example.invalid commit -qm "audit go bidcodec release"
git -C "$go_release" tag "bid754-codec-go/$go_module_version"
go_gitconfig="$audit_tmp/go-gitconfig"
cat >"$go_gitconfig" <<EOF
[url "file://$go_release/"]
	insteadOf = https://github.com/sky1core/bid754
EOF
go_smoke="$audit_tmp/go-smoke"
mkdir -p "$go_smoke"
cat >"$go_smoke/go.mod" <<EOF
module bidcodec-smoke

go 1.21

require github.com/sky1core/bid754/bid754-codec-go $go_module_version
EOF
cat >"$go_smoke/main.go" <<'EOF'
package main

import (
	"fmt"

	bidcodec "github.com/sky1core/bid754/bid754-codec-go"
)

func main() {
	c := bidcodec.Decode32(0x32800001)
	if c.Kind != bidcodec.Normal || c.Coefficient.String() != "1" || bidcodec.ToString(c) != "+1E+0" {
		panic(fmt.Sprintf("unexpected decode: %+v", c))
	}
}
EOF
(cd "$go_smoke" && GIT_CONFIG_GLOBAL="$go_gitconfig" GIT_ALLOW_PROTOCOL=file GOPRIVATE=github.com/sky1core/bid754 GOPROXY=direct GOCACHE="$go_cache" go mod download github.com/sky1core/bid754/bid754-codec-go)
(cd "$go_smoke" && GIT_CONFIG_GLOBAL="$go_gitconfig" GIT_ALLOW_PROTOCOL=file GOPRIVATE=github.com/sky1core/bid754 GOPROXY=direct GOCACHE="$go_cache" go run .)
downloaded_go_version=$(cd "$go_smoke" && GIT_CONFIG_GLOBAL="$go_gitconfig" GIT_ALLOW_PROTOCOL=file GOPRIVATE=github.com/sky1core/bid754 GOPROXY=direct GOCACHE="$go_cache" go list -m -f '{{.Version}}{{if .Replace}} replace{{end}}' github.com/sky1core/bid754/bid754-codec-go)
if [ "$downloaded_go_version" != "$go_module_version" ]; then
  echo "unexpected Go BID codec module resolution: $downloaded_go_version" >&2
  exit 1
fi
cp bid754-codec-go/testdata/external_vector_test.go "$go_smoke/vector_test.go"
(cd "$go_smoke" && GIT_CONFIG_GLOBAL="$go_gitconfig" GIT_ALLOW_PROTOCOL=file GOPRIVATE=github.com/sky1core/bid754 GOPROXY=direct GOCACHE="$go_cache" go test -tags bid754_bidcodec_vectors ./...)

echo "==> Rust BID codec package, docs, and repo vectors"
(
  cd bid754-codec-rs
  cargo package --locked
  cargo doc --locked --no-deps
  cargo clippy --locked --all-targets -- -D warnings
  cargo test --locked
)
rust_smoke="$audit_tmp/rust-smoke"
mkdir -p "$rust_smoke/src"
cat >"$rust_smoke/Cargo.toml" <<EOF
[package]
name = "bid754-codec-smoke"
version = "0.0.0"
edition = "2021"

[dependencies]
bid754-codec = { path = "$repo_root/bid754-codec-rs" }
EOF
cat >"$rust_smoke/src/main.rs" <<'EOF'
fn main() {
    let c = bid754_codec::decode32(0x32800001);
    assert_eq!(c.kind, bid754_codec::Kind::Normal);
    assert_eq!(c.coefficient, 1);
    assert_eq!(bid754_codec::to_string(&c), "+1E+0");
}
EOF
(cd "$rust_smoke" && cargo run --quiet)
mkdir -p "$rust_smoke/tests"
cp bid754-codec-rs/tests/vectors.rs "$rust_smoke/tests/vectors.rs"
cat >>"$rust_smoke/Cargo.toml" <<'EOF'

[dev-dependencies]
serde = { version = "=1.0.228", features = ["derive"] }
serde_json = "=1.0.149"
EOF
(cd "$rust_smoke" && cargo test --quiet)

echo "==> Java BID codec package build"
(
  cd bid754-codec-java
  java_maven_repo="$audit_tmp/java-maven"
  "$repo_root/devtools/scripts/run_pinned_gradle.sh" -q clean build publishMavenJavaPublicationToAuditRepository -Pbid754AuditMavenRepo="$java_maven_repo"
  java_smoke="$audit_tmp/java-smoke"
  mkdir -p "$java_smoke"
  java_version=$(awk -F"'" '/^version = / { print $2; exit }' build.gradle)
  jar="build/libs/bid-codec-java-${java_version}.jar"
  pom="$java_maven_repo/dev/bid754/bidcodec/bid-codec-java/${java_version}/bid-codec-java-${java_version}.pom"
  if [ ! -f "$jar" ]; then
    echo "missing expected Java library jar: $jar" >&2
    exit 1
  fi
  if [ ! -f "$pom" ]; then
    echo "missing expected Java Maven publication POM: $pom" >&2
    exit 1
  fi
  unexpected_jars=$(find build/libs -maxdepth 1 -type f -name '*.jar' ! -name '*sources*' ! -name '*javadoc*' ! -name "$(basename "$jar")")
  if [ -n "$unexpected_jars" ]; then
    echo "unexpected Java library jar(s) after clean build:" >&2
    echo "$unexpected_jars" >&2
    exit 1
  fi
  cat >"$java_smoke/BidCodecSmoke.java" <<'EOF'
import dev.bid754.bidcodec.BidCodec;
import dev.bid754.bidcodec.DecimalKind;

public final class BidCodecSmoke {
    public static void main(String[] args) {
        var c = BidCodec.decode32(0x32800001);
        if (c.kind() != DecimalKind.NORMAL || !c.coefficient().toString().equals("1")
                || !BidCodec.toString(c).equals("+1E+0")) {
            throw new AssertionError("unexpected decode: " + c);
        }
    }
}
EOF
  javac -cp "$jar" "$java_smoke/BidCodecSmoke.java"
  java -cp "$jar:$java_smoke" BidCodecSmoke
  java_vectors="$audit_tmp/java-vectors"
  mkdir -p "$java_vectors"
  javac -cp "$jar" -d "$java_vectors" "$repo_root/bid754-codec-java/src/test/java/dev/bid754/bidcodec/VectorRunner.java"
  java -cp "$jar:$java_vectors" dev.bid754.bidcodec.VectorRunner "$vectors_path"
)

echo "==> Python BID codec wheel and typed marker"
(
  cd bid754-codec-py
  rm -rf build dist
  find . -maxdepth 1 -name '*.egg-info' -exec rm -rf {} +
  py_version=$(python3 - <<'PY'
import tomllib
from pathlib import Path

with Path("pyproject.toml").open("rb") as f:
    print(tomllib.load(f)["project"]["version"])
PY
)
  python3 -m pip wheel . --no-deps -w "$wheel_dir"
  PY_VERSION="$py_version" WHEEL_DIR="$wheel_dir" python3 - <<'PY'
import os
from pathlib import Path
from zipfile import ZipFile

version = os.environ["PY_VERSION"]
wheels = sorted(Path(os.environ["WHEEL_DIR"]).glob(f"bid754_codec-{version}-*.whl"), key=lambda p: p.stat().st_mtime)
if not wheels:
    raise SystemExit(f"missing Python wheel for bid754-codec version {version}")
wheel = wheels[-1]
with ZipFile(wheel) as zf:
    names = set(zf.namelist())
if "bid_codec/py.typed" not in names:
    raise SystemExit(f"missing bid_codec/py.typed in {wheel}")
print(f"wheel includes bid_codec/py.typed: {wheel.name}")
PY
  python3 -m compileall -q bid_codec
  py_venv="$audit_tmp/python-venv"
  python3 -m venv "$py_venv"
  "$py_venv/bin/python" -m pip install --no-index --find-links "$wheel_dir" "bid754-codec==$py_version"
  "$py_venv/bin/python" - <<'PY'
from bid_codec import Kind, decode32, to_string

c = decode32(0x32800001)
assert c.kind == Kind.NORMAL
assert c.coefficient == 1
assert to_string(c) == "+1E+0"
PY
  "$py_venv/bin/python" -m pip install "pytest==9.0.2"
  py_vectors="$audit_tmp/python-vectors"
  mkdir -p "$py_vectors/tests"
  cp "$repo_root/bid754-codec-py/tests/test_vectors.py" "$py_vectors/tests/test_vectors.py"
  (cd "$py_vectors" && "$py_venv/bin/python" -m pytest tests)
  rm -rf build dist
  find . -maxdepth 1 -name '*.egg-info' -exec rm -rf {} +
)

echo "==> JavaScript/TypeScript BID codec package build and pack"
(
  cd bid754-codec-js
  npm ci
  npm run build
  npm test
  pack_dir="$audit_tmp/npm-pack"
  mkdir -p "$pack_dir"
  npm pack --pack-destination "$pack_dir" >/dev/null
  tarball=$(find "$pack_dir" -maxdepth 1 -name '*.tgz' | head -n 1)
  if [ -z "$tarball" ]; then
    echo "missing npm package tarball" >&2
    exit 1
  fi
  js_smoke="$audit_tmp/js-smoke"
  mkdir -p "$js_smoke"
  cat >"$js_smoke/package.json" <<'EOF'
{"type":"module","dependencies":{}}
EOF
  (cd "$js_smoke" && npm install "$tarball" >/dev/null)
  (cd "$js_smoke" && node --input-type=module - <<'EOF'
import { Kind, decode32, toString } from "@bid754/bid-codec";

const c = decode32(0x32800001);
if (c.kind !== Kind.Normal || c.coefficient !== 1n || toString(c) !== "+1E+0") {
  throw new Error(`unexpected decode: kind=${c.kind} coefficient=${c.coefficient} string=${toString(c)}`);
}
EOF
  )
  cp "$repo_root/bid754-codec-js/vector_runner.mjs" "$js_smoke/vector-audit.mjs"
  (cd "$js_smoke" && node vector-audit.mjs "$vectors_path")
)

echo "==> Swift BID codec release build"
swift build -c release
swift_smoke="$audit_tmp/swift-smoke"
mkdir -p "$swift_smoke/Sources/VectorAudit"
cat >"$swift_smoke/Package.swift" <<EOF
// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "BidCodecVectorAudit",
    dependencies: [
        .package(name: "bid754", path: "$repo_root"),
    ],
    targets: [
        .executableTarget(
            name: "VectorAudit",
            dependencies: [
                .product(name: "BidCodec", package: "bid754"),
            ]
        ),
    ]
)
EOF
cp bid754-codec-swift/Sources/BidCodecVectorRunner/main.swift "$swift_smoke/Sources/VectorAudit/main.swift"
(cd "$swift_smoke" && swift run -c release VectorAudit "$vectors_path")

echo "==> Package audit complete (cross-language vector verification runs separately via make test-bidcodec)"
