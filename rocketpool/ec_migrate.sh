#!/bin/sh

EC_CHAINDATA_DIR=/ethclient
EXTERNAL_DIR=/mnt/external
RSYNC_CMD="rsync -a --progress"

if [ "$EC_MIGRATE_MODE" = "export" ]; then
    $RSYNC_CMD $EC_CHAINDATA_DIR/* $EXTERNAL_DIR
elif [ "$EC_MIGRATE_MODE" = "import" ]; then
    rm -rf $EC_CHAINDATA_DIR/*
    $RSYNC_CMD $EXTERNAL_DIR/* $EC_CHAINDATA_DIR
else
    echo "Unknown migrate mode \"$EC_MIGRATE_MODE\""
fi