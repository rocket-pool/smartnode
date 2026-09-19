#!/bin/sh
# TODO(Hegota): Remove this detector when Smart Node drops Patricia state pruning.
# Shared by start-ec.sh and prune-eth1. Inspect persisted state, not just the flat
# directory: Nethermind 2.0 also creates empty flat column families on Patricia.
# Based on https://github.com/ethstaker/eth-docker/pull/2819.
DB_ROOT=${1:-/ethclient/nethermind/nethermind_db}

if [ ! -d "$DB_ROOT" ]; then
    echo none
    exit 0
fi

FLAT_FILES=$(find "$DB_ROOT" -mindepth 3 -maxdepth 3 -path '*/flat/*' -name '*.sst' -print -quit) || exit 1
if [ -n "$FLAT_FILES" ]; then
    echo flat
    exit 0
fi

# Full pruning can put Patricia state files in numbered subdirectories.
STATE_FILES=$(find "$DB_ROOT" -path '*/state/*' -name '*.sst' -print -quit) || exit 1
if [ -n "$STATE_FILES" ]; then
    echo patricia
else
    echo none
fi
