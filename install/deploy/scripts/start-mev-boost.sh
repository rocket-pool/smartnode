#!/bin/sh

# Set up the network-based flag
if [ "$NETWORK" = "mainnet" ]; then
    MEV_NETWORK="mainnet"
elif [ "$NETWORK" = "prater" ]; then
    MEV_NETWORK="goerli"
elif [ "$NETWORK" = "devnet" ]; then
    MEV_NETWORK="goerli"
else
    echo "Unknown network [$NETWORK]"
    exit 1
fi

# Run MEV-boost
exec /app/mev-boost -${MEV_NETWORK} -addr 0.0.0.0:${MEV_BOOST_PORT} -relay-check -relays ${MEV_BOOST_RELAYS}