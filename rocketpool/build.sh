#!/bin/bash

GOOS=linux GOARCH=amd64 go build -o rocketpoold-linux-amd64 rocketpool.go
GOOS=darwin GOARCH=amd64 go build -o rocketpoold-darwin-amd64 rocketpool.go
