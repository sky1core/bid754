#!/usr/bin/env bash
set -euo pipefail

version="9.4.1"
sha256="2ab2958f2a1e51120c326cad6f385153bb11ee93b3c216c5fccebfdfbb7ec6cb"
url="https://services.gradle.org/distributions/gradle-${version}-bin.zip"

cache_root="${BID754_GRADLE_CACHE:-${XDG_CACHE_HOME:-$HOME/.cache}/bid754/gradle}"
zip_path="$cache_root/gradle-${version}-bin.zip"
gradle_home="$cache_root/gradle-${version}"

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "missing required tool for pinned Gradle: $1" >&2
    exit 1
  fi
}

sha256_file() {
  shasum -a 256 "$1" | awk '{print $1}'
}

verify_zip() {
  local got
  got="$(sha256_file "$zip_path")"
  if [ "$got" != "$sha256" ]; then
    echo "checksum mismatch for $zip_path" >&2
    echo "  got:  $got" >&2
    echo "  want: $sha256" >&2
    exit 1
  fi
}

require_tool curl
require_tool shasum
require_tool unzip

if [ -z "${JAVA_HOME:-}" ]; then
  for candidate in \
    /opt/homebrew/opt/openjdk/libexec/openjdk.jdk/Contents/Home \
    /usr/local/opt/openjdk/libexec/openjdk.jdk/Contents/Home; do
    if [ -x "$candidate/bin/java" ]; then
      export JAVA_HOME="$candidate"
      export PATH="$JAVA_HOME/bin:$PATH"
      break
    fi
  done
fi

mkdir -p "$cache_root"

if [ ! -f "$zip_path" ]; then
  curl -L --fail --show-error --silent "$url" -o "$zip_path"
fi

verify_zip

if [ ! -x "$gradle_home/bin/gradle" ]; then
  rm -rf "$gradle_home"
  unzip -oq "$zip_path" -d "$cache_root"
fi

exec "$gradle_home/bin/gradle" "$@"
