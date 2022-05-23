#!/bin/sh

EC_CHAINDATA_DIR=/ethclient
EXTERNAL_DIR=/mnt/external

if [ "$OPERATION" = "size" ]; then

    if [ ! -d $EXTERNAL_DIR ]; then
        echo "Source path is not a directory." 1>&2
        exit 1
    else 
        # Get the space used by the external directory
        du -s -B 1 $EXTERNAL_DIR | awk -F '\\s*' '{print $1}'
    fi

else 

    RSYNC_CMD="rsync -a --progress"

    if [ "$EC_MIGRATE_MODE" = "export" ]; then
        $RSYNC_CMD $EC_CHAINDATA_DIR/* $EXTERNAL_DIR
    elif [ "$EC_MIGRATE_MODE" = "import" ]; then
        rm -rf $EC_CHAINDATA_DIR/*
        $RSYNC_CMD $EXTERNAL_DIR/* $EC_CHAINDATA_DIR
    else
        echo "Unknown migrate mode \"$EC_MIGRATE_MODE\""
    fi

fi