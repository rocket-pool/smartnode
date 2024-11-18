#!/bin/sh

parse_additional_flags() {
  # Initialize an empty string for additional arguments
  ADDITIONAL_ARGS=""

  # Check if the environment variable MEV_BOOST_ADDITIONAL_FLAGS is not empty
  if [ -n "$MEV_BOOST_ADDITIONAL_FLAGS" ]; then
    # Split the input string into an array of key-value pairs using comma as the delimiter
    IFS=',' read -r -a pairs <<EOF
$MEV_BOOST_ADDITIONAL_FLAGS
EOF

    # Iterate over each key-value pair
    for pair in "${pairs[@]}"; do
      # Extract the key and value from the current pair
      key=$(echo "$pair" | cut -d'=' -f1)
      value=$(echo "$pair" | cut -d'=' -f2)

      # Append the key-value pair to the ADDITIONAL_ARGS string
      ADDITIONAL_ARGS="$ADDITIONAL_ARGS -${key} ${value}"
    done
  fi
}

if [ "$NETWORK" = "mainnet" ]; then
    MEV_NETWORK="mainnet"
elif [ "$NETWORK" = "holesky" ]; then
    MEV_NETWORK="holesky"
elif [ "$NETWORK" = "devnet" ]; then
    MEV_NETWORK="holesky"
else
    echo "Unknown network [$NETWORK]"
    exit 1
fi

parse_additional_flags

exec /app/mev-boost -${MEV_NETWORK} -addr 0.0.0.0:${MEV_BOOST_PORT} -relay-check -relays ${MEV_BOOST_RELAYS} ${ADDITIONAL_ARGS}
