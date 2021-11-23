#!/bin/bash

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
wget https://golang.org/dl/go1.17.3.linux-$ARCH.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.17.3.linux-$ARCH.tar.gz
export PATH=$PATH:/usr/local/go/bin
cd /smartnode/rocketpool
go build -o rocketpool-daemon-linux-$ARCH rocketpool.go
