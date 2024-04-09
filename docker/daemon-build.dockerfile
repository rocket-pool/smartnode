# The builder for building the daemon
FROM --platform=${BUILDPLATFORM} golang:1.21-bookworm AS builder
ARG TARGETOS TARGETARCH BUILDPLATFORM
COPY . /rocketpool
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"
RUN if [ "$BUILDPLATFORM" = "linux/amd64" -a "$TARGETARCH" = "arm64" ]; then \
        # Install the GCC cross compiler
        apt update && apt install -y gcc-aarch64-linux-gnu g++-aarch64-linux-gnu && \
        export CC=aarch64-linux-gnu-gcc && export CC_FOR_TARGET=gcc-aarch64-linux-gnu; \
    elif [ "$BUILDPLATFORM" = "linux/arm64" -a "$TARGETARCH" = "amd64" ]; then \
        apt update && apt install -y gcc-x86-64-linux-gnu g++-x86-64-linux-gnu && \
        export CC=x86_64-linux-gnu-gcc && export CC_FOR_TARGET=gcc-x86-64-linux-gnu; \
    fi && \
    cd /rocketpool/src/rocketpool-daemon && \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /build/rocketpool-daemon

# Copy the output
FROM scratch AS daemon
ARG TARGETOS TARGETARCH
COPY --from=builder /build/rocketpool-daemon /rocketpool-daemon
