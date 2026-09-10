# Local Linux verification image for bid754, used by devtools/scripts/verify_linux.sh
# to run the Linux gates without GitHub Actions. Toolchain surface mirrors the
# CI workflow's ubuntu-24.04 jobs for the Go/Rust/C gates; built per-arch via
# docker build --platform linux/amd64 | linux/arm64 | linux/s390x (the s390x
# image backs the local-only big-endian digest leg and is Go-only).
FROM ubuntu:24.04

ARG TARGETARCH

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        clang \
        curl \
        gcc \
        git \
        libc6-dev \
        make \
        perl \
        python3 \
        unzip \
    && rm -rf /var/lib/apt/lists/*

# Go pinned to the go.mod toolchain directive (toolchain go1.26.8).
ARG GO_VERSION=1.26.8
ARG GO_SHA256_AMD64=d0f743b33e8d8945e6b1f432edd15785c70507121d6e2a723b21285eddf8b57b
ARG GO_SHA256_ARM64=211ffced9dcb9633a55eac6364816ec0ddd951389a740e88fa8b3337971bdda0
ARG GO_SHA256_S390X=339d1a09716815f737cb5ddd355e04e339312b449c268cc417118f4b8137ddd1
RUN set -eu; \
    case "$TARGETARCH" in \
        amd64) go_sha="$GO_SHA256_AMD64" ;; \
        arm64) go_sha="$GO_SHA256_ARM64" ;; \
        s390x) go_sha="$GO_SHA256_S390X" ;; \
        *) echo "unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" -o /tmp/go.tar.gz; \
    echo "$go_sha  /tmp/go.tar.gz" | shasum -a 256 -c -; \
    tar -C /usr/local -xzf /tmp/go.tar.gz; \
    rm /tmp/go.tar.gz

# Rust uses the version selected by CI. The
# s390x image skips Rust: its only consumer is the Go-only digest-s390x
# big-endian leg, and installing a toolchain under qemu-s390x emulation
# would add substantial image-build time for gates that never run there.
# The minimal profile omits clippy, which make test-rust runs as the lint
# gate for the generated public API, so request the component explicitly.
COPY rust-version /tmp/bid754-rust-version
RUN set -eu; \
    if [ "$TARGETARCH" != "s390x" ]; then \
        curl -fsSL https://sh.rustup.rs | sh -s -- -y --profile minimal --default-toolchain "$(cat /tmp/bid754-rust-version)" --component clippy; \
    fi

ENV PATH=/usr/local/go/bin:/root/.cargo/bin:$PATH
# The installed Go must satisfy go.mod's toolchain directive by itself; fail
# loudly instead of silently downloading another toolchain at run time.
ENV GOTOOLCHAIN=local

WORKDIR /work
