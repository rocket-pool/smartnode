#!/bin/sh
set -e

# Grab the Testnet genesis state if needed
if [ "$NETWORK" = "testnet" ]; then
    echo "Prysm is configured to use Hoodi, genesis state required."
    if [ ! -f "/ethclient/hoodi-genesis.ssz" ]; then
        echo "Downloading from Github..."
        wget https://github.com/eth-clients/hoodi/raw/refs/heads/main/metadata/genesis.ssz -O /ethclient/hoodi-genesis.ssz
        echo "Download complete."
    else
        echo "Genesis state already downloaded, continuing."
    fi
elif [ "$NETWORK" = "devnet" ]; then
    echo "Prysm is configured to use Hoodi, genesis state required."
    if [ ! -f "/ethclient/hoodi-genesis.ssz" ]; then
        echo "Downloading from Github..."
        wget https://github.com/eth-clients/hoodi/raw/refs/heads/main/metadata/genesis.ssz -O /ethclient/hoodi-genesis.ssz
        echo "Download complete."
    else
        echo "Genesis state already downloaded, continuing."
    fi
else
    echo "Genesis download not required for $NETWORK"
fi
