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
        unzip \
    && rm -rf /var/lib/apt/lists/*

# Go pinned to the go.mod toolchain directive (toolchain go1.25.12).
ARG GO_VERSION=1.25.12
ARG GO_SHA256_AMD64=234828b7a89e0e303d2556310ee549fbcf253d28de937bac3da13d6294262ac1
ARG GO_SHA256_ARM64=8b5884aef89600aef5b0b051fb971f11f49bb996521e911f30f02a66884f7bd2
ARG GO_SHA256_S390X=09875de1d6cb3237437112271b9df3a96bac9fcd10cbf9bc777c00e67f4e3e3d
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

# Rust stable via rustup, the same channel the CI workflow installs. The
# s390x image skips Rust: its only consumer is the Go-only digest-s390x
# big-endian leg, and installing a toolchain under qemu-s390x emulation
# would add substantial image-build time for gates that never run there.
RUN set -eu; \
    if [ "$TARGETARCH" != "s390x" ]; then \
        curl -fsSL https://sh.rustup.rs | sh -s -- -y --profile minimal --default-toolchain stable; \
    fi

ENV PATH=/usr/local/go/bin:/root/.cargo/bin:$PATH
# The installed Go must satisfy go.mod's toolchain directive by itself; fail
# loudly instead of silently downloading another toolchain at run time.
ENV GOTOOLCHAIN=local

WORKDIR /work
