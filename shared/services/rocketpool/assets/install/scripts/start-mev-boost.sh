#!/bin/sh

# Initialize an empty string for additional arguments
ADDITIONAL_ARGS=""

# Check if the environment variable MEV_BOOST_ADDITIONAL_FLAGS is not empty
if [ -n "$MEV_BOOST_ADDITIONAL_FLAGS" ]; then
  ADDITIONAL_ARGS=$MEV_BOOST_ADDITIONAL_FLAGS
    
fi

# Set up the network-based flag
if [ -z "$BEACON_NETWORK" ]; then
    echo "BEACON_NETWORK is not set"
    exit 1
fi
MEV_NETWORK="$BEACON_NETWORK"

exec /app/mev-boost -${MEV_NETWORK} -addr 0.0.0.0:${MEV_BOOST_PORT} -relay-check -relays ${MEV_BOOST_RELAYS} ${ADDITIONAL_ARGS}
