#!/bin/bash

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o rocketpool-cli-linux-amd64 rocketpool-cli.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o rocketpool-cli-darwin-amd64 rocketpool-cli.go
#CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o rocketpool-cli-windows-amd64.exe rocketpool-cli.go

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o rocketpool-cli-linux-arm64 rocketpool-cli.go
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o rocketpool-cli-darwin-arm64 rocketpool-cli.go
