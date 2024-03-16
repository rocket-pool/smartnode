# The builder for building the CLIs
FROM golang:1.21-bookworm AS builder
COPY . /rocketpool
ENV CGO_ENABLED=0
WORKDIR /rocketpool/src/rocketpool-cli

# Build x64 version
RUN GOOS=linux GOARCH=amd64 go build -o /build/rocketpool-cli-linux-amd64
RUN GOOS=darwin GOARCH=amd64 go build -o /build/rocketpool-cli-darwin-amd64

# Build the arm64 version
RUN GOOS=linux GOARCH=arm64 go build -o /build/rocketpool-cli-linux-arm64
RUN GOOS=darwin GOARCH=arm64 go build -o /build/rocketpool-cli-darwin-arm64

# Copy the output
FROM scratch AS cli
COPY --from=builder /build /