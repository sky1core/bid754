#!/bin/bash
# install_ibm_decnumber.sh - Build and install pinned IBM decNumber 3.68.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

PREFIX="${BID754_DECNUMBER_PREFIX:-$HOME/local}"
WORKDIR="${BID754_DECNUMBER_WORKDIR:-${TMPDIR:-/tmp}/bid754-decnumber-368}"
CACHE_DIR="${BID754_DECNUMBER_CACHE:-$PROJECT_ROOT/devtools/third_party/ibm_decnumber}"
DECNUMBER_URL="${BID754_DECNUMBER_URL:-https://www.speleotrove.com/decimal/decNumber-icu-368.zip}"
DECNUMBER_SHA256="14ec2cf30b58758493a7661b78b80abfb281652b61a425b85cda83173518fe25"
DECNUMBER_ARCHIVE="$CACHE_DIR/decNumber-icu-368.zip"
CC="${CC:-cc}"
AR="${AR:-ar}"
CFLAGS="${CFLAGS:-}"
JOBS="${BID754_JOBS:-2}"

echo "Installing IBM decNumber into: $PREFIX"
echo "Using workdir: $WORKDIR"
echo "Using archive: $DECNUMBER_ARCHIVE"

rm -rf "$WORKDIR"
mkdir -p "$WORKDIR" "$CACHE_DIR"

if [ ! -f "$DECNUMBER_ARCHIVE" ]; then
    echo "Downloading IBM decNumber 3.68 from $DECNUMBER_URL"
    curl -L --fail --show-error --silent "$DECNUMBER_URL" -o "$DECNUMBER_ARCHIVE"
fi

got_sha="$(shasum -a 256 "$DECNUMBER_ARCHIVE" | awk '{print $1}')"
if [ "$got_sha" != "$DECNUMBER_SHA256" ]; then
    echo "checksum mismatch for $DECNUMBER_ARCHIVE" >&2
    echo "  got:  $got_sha" >&2
    echo "  want: $DECNUMBER_SHA256" >&2
    exit 1
fi

unzip -q "$DECNUMBER_ARCHIVE" -d "$WORKDIR"
cd "$WORKDIR/decNumber"

# The current native decTest path runs on little-endian macOS/Linux runners.
# DECLITEND can be overridden through CFLAGS if a big-endian port is added.
"$CC" -std=c99 -O2 -DDECLITEND=1 $CFLAGS -c \
    decContext.c \
    decNumber.c \
    decimal32.c \
    decimal64.c \
    decimal128.c \
    decPacked.c
"$AR" rcs libdecnumber.a \
    decContext.o \
    decNumber.o \
    decimal32.o \
    decimal64.o \
    decimal128.o \
    decPacked.o

mkdir -p "$PREFIX/lib" "$PREFIX/include/libdecnumber/dpd"
cp libdecnumber.a "$PREFIX/lib/"
cp ./*.h "$PREFIX/include/libdecnumber/"
cp ./decimal32.h ./decimal64.h ./decimal128.h "$PREFIX/include/libdecnumber/dpd/"

cat <<EOF
IBM decNumber installation complete.

Source:
  IBM decNumber 3.68 ICU zip
  $DECNUMBER_URL
  sha256: $DECNUMBER_SHA256

Installed files:
  $PREFIX/lib/libdecnumber.a
  $PREFIX/include/libdecnumber/decNumber.h
  $PREFIX/include/libdecnumber/dpd/decimal32.h

Next steps:
  bash "$PROJECT_ROOT/devtools/scripts/setup_c_libs.sh"
  source "$PROJECT_ROOT/.env.sh"
  go test -tags bid754_native -short ./...
EOF
