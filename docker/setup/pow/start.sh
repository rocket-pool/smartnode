#!/bin/sh
# This script configures ETH1 clients for Rocket Pool's scalable docker stack; only edit if you know what you're doing ;)

# Get container ID
CONTAINERID="${HOSTNAME}"

# Create a container data directory if it doesn't exist
DATADIR="/ethclient/$CONTAINERID"
mkdir -p "$DATADIR"

# Geth startup
if [ $CLIENT == "geth" ]; then

    # Initialise
    CMD="/usr/local/bin/geth --datadir $DATADIR init /setup/genesis77.json"

    # Run
    CMD="$CMD && /usr/local/bin/geth --datadir $DATADIR --networkid $NETWORKID --bootnodes $BOOTNODE"
    CMD="$CMD --rpc --rpcaddr 0.0.0.0 --rpcport 8545 --rpcapi db,eth,net,web3,personal --rpcvhosts 'eth1.rpc.smartnode.localhost'"

    # Add Ethstats to run
    if [ ! -z "$ETHSTATSLABEL" ] && [ ! -z "$ETHSTATSLOGIN" ]; then
        CMD="$CMD --ethstats $ETHSTATSLABEL-$CONTAINERID:$ETHSTATSLOGIN"
    fi

    # Run command
    eval "$CMD"

fi

# Parity startup
# TODO: implement
