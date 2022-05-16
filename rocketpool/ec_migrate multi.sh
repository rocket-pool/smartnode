#!/bin/sh

EC_CHAINDATA_DIR=/ethclient
EXTERNAL_DIR=/mnt/external

# Get the core count
CORE_COUNT=$(grep -c ^processor /proc/cpuinfo)
if [ -z "$CORE_COUNT" ]; then
    CORE_COUNT=1
elif [ "$CORE_COUNT" -lt 1 ]; then
    CORE_COUNT=1
fi

RSYNC_CMD="xargs -n1 -P${CORE_COUNT} -I% rsync -a --progress %"

if [ "$EC_MIGRATE_MODE" = "export" ]; then
    ls $EC_CHAINDATA_DIR/* | $RSYNC_CMD $EXTERNAL_DIR
elif [ "$EC_MIGRATE_MODE" = "import" ]; then
    rm -rf $EC_CHAINDATA_DIR/*
    ls $EXTERNAL_DIR/* | $RSYNC_CMD $EC_CHAINDATA_DIR
else
    echo "Unknown migrate mode \"$EC_MIGRATE_MODE\""
fi