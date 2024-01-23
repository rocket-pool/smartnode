# The builder for building the daemon
FROM golang:1.21-bookworm AS builder
ARG TARGETARCH
COPY . /rocketpool
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-O -D__BLST_PORTABLE__"
RUN cd /rocketpool/rocketpool-daemon && go build -o build/rocketpool-daemon-linux-${TARGETARCH}

# Copy the output
FROM scratch AS daemon
ARG TARGETARCH
COPY --from=builder /rocketpool/rocketpool-daemon/build/rocketpool-daemon-linux-${TARGETARCH} /
