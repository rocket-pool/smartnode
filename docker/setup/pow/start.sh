#!/bin/sh
# This script configures ETH1 clients for Rocket Pool's scalable docker stack; only edit if you know what you're doing ;)

# Arguments
CLIENT=$1          # Client type (geth/parity)
NETWORKID=$2       # ID of network to join
BOOTNODE=$3        # Initial bootnode(s) to connect to
ETHSTATSLABEL=$4   # The ID to show on the ETHSTATS page (with container ID appended)
ETHSTATSLOGIN=$5   # Login credentials for ethstats

# Get container ID
CONTAINERID="${HOSTNAME}"

# Create a container data directory if it doesn't exist
DATADIR="/ethclient/$CONTAINERID"
mkdir -p "$DATADIR"

# Geth startup
if [ $CLIENT == "geth" ]; then

    # Initialise
    CMD="/usr/local/bin/geth --datadir $DATADIR init /pow/genesis77.json"

    # Run
    CMD="$CMD && /usr/local/bin/geth --datadir $DATADIR --networkid $NETWORKID --bootnodes $BOOTNODE"
    CMD="$CMD --rpc --rpcaddr 0.0.0.0 --rpcport 8545 --rpcapi db,eth,net,web3,personal --rpcvhosts '*'"
    CMD="$CMD --ws --wsaddr 0.0.0.0 --wsport 8546 --wsapi db,eth,net,web3,personal --wsorigins '*'"

    # Add Ethstats to run
    if [ ! -z "$ETHSTATSLABEL" ] && [ ! -z "$ETHSTATSLOGIN" ]; then
        CMD="$CMD --ethstats $ETHSTATSLABEL-$CONTAINERID:$ETHSTATSLOGIN"
    fi

    # Run command
    eval "$CMD"

fi

# Parity startup
# TODO: implement
