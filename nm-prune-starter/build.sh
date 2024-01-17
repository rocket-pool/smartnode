#!/bin/bash

export CGO_ENABLED=0
cd /smartnode/nm-prune-starter

# Build x64 version
CGO_CFLAGS="-O -D__BLST_PORTABLE__" GOARCH=amd64 GOOS=linux go build -o nm-prune-starter-linux-amd64 nm-prune-starter.go

# Build the arm64 version
CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-cpp CGO_CFLAGS="-O -D__BLST_PORTABLE__" GOARCH=arm64 GOOS=linux go build -o nm-prune-starter-linux-arm64 nm-prune-starter.go