#!/bin/sh

# Generates the ssz encoding methods for eth2 types with fastssz
# Install sszgen with `go get github.com/ferranbt/fastssz/sszgen`
rm -f ./shared/types/eth2/types_encoding.go
sszgen --path ./shared/types/eth2