#!/bin/sh

# Initialize an empty string for additional arguments
ADDITIONAL_ARGS=""

# Check if the environment variable MEV_BOOST_ADDITIONAL_FLAGS is not empty
if [ -n "$MEV_BOOST_ADDITIONAL_FLAGS" ]; then
  ADDITIONAL_ARGS=$MEV_BOOST_ADDITIONAL_FLAGS
    
fi

# Set up the network-based flag
if [ "$NETWORK" = "mainnet" ]; then
    MEV_NETWORK="mainnet"
elif [ "$NETWORK" = "testnet" ]; then
    MEV_NETWORK="hoodi"
elif [ "$NETWORK" = "devnet" ]; then
    MEV_NETWORK="hoodi"
else
    echo "Unknown network [$NETWORK]"
    exit 1
fi

exec /app/mev-boost -${MEV_NETWORK} -addr 0.0.0.0:${MEV_BOOST_PORT} -relay-check -relays ${MEV_BOOST_RELAYS} ${ADDITIONAL_ARGS}
