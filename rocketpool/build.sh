#!/bin/bash

export CGO_ENABLED=1
cd /smartnode/rocketpool

# Build x64 version
CGO_CFLAGS="-O -D__BLST_PORTABLE__" GOARCH=amd64 GOOS=linux go build -o rocketpool-daemon-linux-amd64 rocketpool.go

# Build the arm64 version
CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-cpp CGO_CFLAGS="-O -D__BLST_PORTABLE__" GOARCH=arm64 GOOS=linux go build -o rocketpool-daemon-linux-arm64 rocketpool.go