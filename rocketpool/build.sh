#!/bin/bash

GO_VERSION=1.18.4

# Get CPU architecture
UNAME_VAL=$(uname -m)
ARCH=""
case $UNAME_VAL in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    *)       fail "CPU architecture not supported: $UNAME_VAL" ;;
esac

apt update
apt dist-upgrade -y
apt install build-essential git wget -y
cd /tmp
wget https://golang.org/dl/go${GO_VERSION}.linux-$ARCH.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go${GO_VERSION}.linux-$ARCH.tar.gz
export PATH=$PATH:/usr/local/go/bin
cd /smartnode/rocketpool
CGO_CFLAGS="-O -D__BLST_PORTABLE__" go build -o rocketpool-daemon-linux-$ARCH rocketpool.go
