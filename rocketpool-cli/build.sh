#!/bin/bash

export CGO_ENABLED=0
cd /smartnode/rocketpool-cli

# Build x64 version
GOOS=linux GOARCH=amd64 go build -o rocketpool-cli-linux-amd64 rocketpool-cli.go
GOOS=darwin GOARCH=amd64 go build -o rocketpool-cli-darwin-amd64 rocketpool-cli.go

# Build the arm64 version
GOOS=linux GOARCH=arm64 go build -o rocketpool-cli-linux-arm64 rocketpool-cli.go
GOOS=darwin GOARCH=arm64 go build -o rocketpool-cli-darwin-arm64 rocketpool-cli.go
