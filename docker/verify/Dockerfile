# Local Linux verification image for bid754, used by scripts/verify_linux.sh
# to run the Linux gates without GitHub Actions. Toolchain surface mirrors the
# CI workflow's ubuntu-24.04 jobs for the Go/Rust/C gates; built per-arch via
# docker build --platform linux/amd64 | linux/arm64.
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
        unzip \
    && rm -rf /var/lib/apt/lists/*

# Go pinned to the go.mod toolchain directive (toolchain go1.25.11).
ARG GO_VERSION=1.25.11
ARG GO_SHA256_AMD64=34f14304e856893f4ba30c2cacfe93906e9de7915c5f6aaaf3a81cdccd7ba30b
ARG GO_SHA256_ARM64=c30bf9e156a54ea4e31fbbbf31a712b32734b58cc9a22426fa5ee632d0885124
RUN set -eu; \
    case "$TARGETARCH" in \
        amd64) go_sha="$GO_SHA256_AMD64" ;; \
        arm64) go_sha="$GO_SHA256_ARM64" ;; \
        *) echo "unsupported TARGETARCH: $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz" -o /tmp/go.tar.gz; \
    echo "$go_sha  /tmp/go.tar.gz" | shasum -a 256 -c -; \
    tar -C /usr/local -xzf /tmp/go.tar.gz; \
    rm /tmp/go.tar.gz

# Rust stable via rustup, the same channel the CI workflow installs.
RUN curl -fsSL https://sh.rustup.rs | sh -s -- -y --profile minimal --default-toolchain stable

ENV PATH=/usr/local/go/bin:/root/.cargo/bin:$PATH
# The installed Go must satisfy go.mod's toolchain directive by itself; fail
# loudly instead of silently downloading another toolchain at run time.
ENV GOTOOLCHAIN=local

WORKDIR /work
