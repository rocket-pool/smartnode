#!/bin/sh
# This script launches ETH1 clients for Rocket Pool's docker stack; only edit if you know what you're doing ;)

# Performance tuning for ARM systems
define_perf_prefix() {
    # Get the number of available cores
    CORE_COUNT=$(nproc)

    # Don't do performance tweaks on systems with 6+ cores
    if [ "$CORE_COUNT" -gt "5" ]; then
        echo "$CORE_COUNT cores detected, skipping performance tuning"
        return 0
    else
        echo "$CORE_COUNT cores detected, activating performance tuning"
    fi

    # Give the EC access to the last core
    CURRENT_CORE=$((CORE_COUNT - 1))
    CORE_STRING="$CURRENT_CORE"

    # If there are more than 2 cores, limit the EC to use all but the first one
    CURRENT_CORE=$((CURRENT_CORE - 1))
    while [ "$CURRENT_CORE" -gt "0" ]; do
        CORE_STRING="$CORE_STRING,$CURRENT_CORE"
        CURRENT_CORE=$((CURRENT_CORE - 1))
    done

    PERF_PREFIX="taskset -c $CORE_STRING ionice -c 3"
    echo "Performance tuning: $PERF_PREFIX"
}

# Set up the network-based flags
if [ "$NETWORK" = "mainnet" ]; then
    GETH_NETWORK=""
    RP_NETHERMIND_NETWORK="mainnet"
    BESU_NETWORK="--network=mainnet"
    RETH_NETWORK="--chain mainnet"
elif [ "$NETWORK" = "devnet" ]; then
    . "/devnet/nodevars_env.txt"
    GETH_NETWORK="--networkid 39438153"
    RP_NETHERMIND_NETWORK="private"
    BESU_NETWORK="--network=ephemery --bootnodes=$BOOTNODE_ENODE_LIST"
    RETH_NETWORK="--chain /devnet/genesis.json --bootnodes $BOOTNODE_ENODE_LIST"
elif [ "$NETWORK" = "testnet" ]; then
    GETH_NETWORK="--hoodi"
    RP_NETHERMIND_NETWORK="hoodi"
    BESU_NETWORK="--network=hoodi"
    RETH_NETWORK="--chain hoodi"
else
    echo "Unknown network [$NETWORK]"
    exit 1
fi


# Geth startup
if [ "$CLIENT" = "geth" ]; then

    # Performance tuning for ARM systems
    UNAME_VAL=$(uname -m)
    if [ "$UNAME_VAL" = "arm64" ] || [ "$UNAME_VAL" = "aarch64" ]; then

        # Install taskset and ionice
        apk add util-linux

        # Define the performance tuning prefix
        define_perf_prefix

    fi

    # Check for the prune flag and run that first if requested
    if [ -f "/ethclient/prune.lock" ]; then

        if [ "$EC_PRUNING_MODE" = "historyExpiry" ]; then
            $PERF_PREFIX /usr/local/bin/geth prune-history $GETH_NETWORK --datadir /ethclient/geth ; rm /ethclient/prune.lock    
        fi

        if [ "$EC_PRUNING_MODE" = "fullNode" ]; then
            $PERF_PREFIX /usr/local/bin/geth snapshot prune-state $GETH_NETWORK --datadir /ethclient/geth ; rm /ethclient/prune.lock
        fi

    # Run Geth normally
    else

        if [ "$NETWORK" = "devnet" ]; then
            geth init --datadir /ethclient/geth /devnet/genesis.json 
        fi

        CMD="$PERF_PREFIX /usr/local/bin/geth $GETH_NETWORK \
            --datadir /ethclient/geth \
            --http \
            --http.addr 0.0.0.0 \
            --http.port ${EC_HTTP_PORT:-8545} \
            --http.api eth,net,web3 \
            --http.corsdomain=* \
            --ws \
            --ws.addr 0.0.0.0 \
            --ws.port ${EC_WS_PORT:-8546} \
            --ws.api eth,net,web3 \
            --authrpc.addr 0.0.0.0 \
            --authrpc.port ${EC_ENGINE_PORT:-8551} \
            --authrpc.jwtsecret /secrets/jwtsecret \
            --authrpc.vhosts=* \
            --pprof \
            $EC_ADDITIONAL_FLAGS"

        if [ "$NETWORK" = "devnet" ]; then\
            CMD="$CMD --bootnodes $BOOTNODE_ENODE_LIST"
        fi
        if [ ! -z "$EC_SUGGESTED_BLOCK_GAS_LIMIT" ]; then
            CMD="$CMD --miner.gaslimit $EC_SUGGESTED_BLOCK_GAS_LIMIT"
        fi
        
        if [ "$EC_PRUNING_MODE" = "archive" ]; then
            CMD="$CMD --syncmode=full --gcmode=archive"
        fi

        if [ "$EC_PRUNING_MODE" = "historyExpiry" ]; then
            CMD="$CMD --history.chain postmerge"
        fi

        if [ ! -z "$GETH_EVM_TIMEOUT" ]; then
            CMD="$CMD --rpc.evmtimeout ${GETH_EVM_TIMEOUT}s"
        fi
        
        if [ ! -z "$ETHSTATS_LABEL" ] && [ ! -z "$ETHSTATS_LOGIN" ]; then
            CMD="$CMD --ethstats $ETHSTATS_LABEL:$ETHSTATS_LOGIN"
        fi

        if [ ! -z "$EC_MAX_PEERS" ]; then
            CMD="$CMD --maxpeers $EC_MAX_PEERS"
        fi

        if [ "$ENABLE_METRICS" = "true" ]; then
            CMD="$CMD --metrics --metrics.addr 0.0.0.0 --metrics.port $EC_METRICS_PORT"
        fi

        if [ ! -z "$EC_P2P_PORT" ]; then
            CMD="$CMD --port $EC_P2P_PORT"
        fi

        exec ${CMD} --http.vhosts '*'

    fi

fi


# Nethermind startup
if [ "$CLIENT" = "nethermind" ]; then

    # Performance tuning for ARM systems
    UNAME_VAL=$(uname -m)
    if [ "$UNAME_VAL" = "arm64" ] || [ "$UNAME_VAL" = "aarch64" ]; then

        # Define the performance tuning prefix
        define_perf_prefix

    fi

    # Create the JWT secret
    if [ ! -f "/secrets/jwtsecret" ]; then
        openssl rand -hex 32 | tr -d "\n" > /secrets/jwtsecret
    fi

    # Set the JSON RPC logging level
    LOG_LINE=$(awk '/<logger name=\"\*\" minlevel=\"Off\" writeTo=\"seq\" \/>/{print NR}' /nethermind/NLog.config)
    sed -e "${LOG_LINE} i \    <logger name=\"JsonRpc\.\*\" final=\"true\"/>\\n" -i /nethermind/NLog.config
    sed -e "${LOG_LINE} i \    <logger name=\"JsonRpc\.\*\" minlevel=\"Warn\" writeTo=\"auto-colored-console-async\" final=\"true\"/>" -i /nethermind/NLog.config
    sed -e "${LOG_LINE} i \    <logger name=\"JsonRpc\.\*\" minlevel=\"Warn\" writeTo=\"file-async\" final=\"true\"/>" -i /nethermind/NLog.config

    # Remove the sync peers report but leave error messages
    sed -e "${LOG_LINE} i \    <logger name=\"Synchronization.Peers.SyncPeersReport\" maxlevel=\"Info\" final=\"true\"/>" -i /nethermind/NLog.config
    sed -i 's/<!-- \(<logger name=\"Synchronization\.Peers\.SyncPeersReport\".*\/>\).*-->/\1/g' /nethermind/NLog.config

    # Get the binary name (changed with v1.21, required for backwards compatibility)
    if [ -f "/nethermind/Nethermind.Runner" ]; then
        NETHERMIND_BINARY=/nethermind/Nethermind.Runner
    elif [ -f "/nethermind/nethermind" ]; then
        NETHERMIND_BINARY=/nethermind/nethermind
    else
        echo "Nethermind binary not found, cannot start Execution Client."
        exit 1
    fi

    if [ "$NETWORK" = "devnet" ]; then
        EPHEMERY_CONFIG="--config /devnet/nethermind-config.json"
    else
        EPHEMERY_CONFIG="--config $RP_NETHERMIND_NETWORK"
    fi

    CMD="$PERF_PREFIX $NETHERMIND_BINARY \
        $EPHEMERY_CONFIG \
        --data-dir /ethclient/nethermind \
        --JsonRpc.Enabled true \
        --JsonRpc.Host 0.0.0.0 \
        --JsonRpc.Port ${EC_HTTP_PORT:-8545} \
        --JsonRpc.EnginePort ${EC_ENGINE_PORT:-8551} \
        --JsonRpc.EngineHost 0.0.0.0 \
        --Init.WebSocketsEnabled true \
        --JsonRpc.WebSocketsPort ${EC_WS_PORT:-8546} \
        --JsonRpc.JwtSecretFile=/secrets/jwtsecret \
        --Pruning.FullPruningTrigger=VolumeFreeSpace \
        --Pruning.FullPruningThresholdMb=$RP_NETHERMIND_FULL_PRUNING_THRESHOLD_MB \
        --Pruning.FullPruningCompletionBehavior AlwaysShutdown \
        --Pruning.FullPruningMaxDegreeOfParallelism=$RP_NETHERMIND_FULL_PRUNING_MAX_DEGREE_PARALLELISM \
        $EC_ADDITIONAL_FLAGS"

    if [ ! -z "$EC_SUGGESTED_BLOCK_GAS_LIMIT" ]; then
            CMD="$CMD --Blocks.TargetBlockGasLimit $EC_SUGGESTED_BLOCK_GAS_LIMIT"
    fi

    if [ "$EC_PRUNING_MODE" = "archive" ]; then
        CMD="$CMD --Sync.DownloadBodiesInFastSync=false --Sync.DownloadReceiptsInFastSync=false --Sync.FastSync=false --Sync.SnapSync=false --Sync.FastBlocks=false --Sync.PivotNumber=0"
        CMD="$CMD --Pruning.Mode=None"
    fi

    if [ "$EC_PRUNING_MODE" = "fullNode" ]; then
        CMD="$CMD --Sync.AncientBodiesBarrier=0 --Sync.AncientReceiptsBarrier=0"
        CMD="$CMD --Pruning.Mode=Hybrid"
    fi

    if [ "$EC_PRUNING_MODE" = "historyExpiry" ]; then
        CMD="$CMD --Sync.AncientBodiesBarrier=15537394 --Sync.AncientReceiptsBarrier=15537394"
        CMD="$CMD --Pruning.Mode=Hybrid"
    fi
    
    # Add optional supplemental primary JSON-RPC modules
    if [ ! -z "$RP_NETHERMIND_ADDITIONAL_MODULES" ]; then
        RP_NETHERMIND_ADDITIONAL_MODULES=",${RP_NETHERMIND_ADDITIONAL_MODULES}"
    fi
    CMD="$CMD --JsonRpc.EnabledModules Eth,Net,Web3$RP_NETHERMIND_ADDITIONAL_MODULES"

    # Add optional supplemental JSON-RPC URLs
    if [ ! -z "$RP_NETHERMIND_ADDITIONAL_URLS" ]; then
        RP_NETHERMIND_ADDITIONAL_URLS=",${RP_NETHERMIND_ADDITIONAL_URLS}"
    fi
    CMD="$CMD --JsonRpc.AdditionalRpcUrls [\"http://127.0.0.1:7434|http|admin\"$RP_NETHERMIND_ADDITIONAL_URLS]"

    if [ ! -z "$ETHSTATS_LABEL" ] && [ ! -z "$ETHSTATS_LOGIN" ]; then
        CMD="$CMD --EthStats.Enabled true --EthStats.Name $ETHSTATS_LABEL --EthStats.Secret $(echo $ETHSTATS_LOGIN | cut -d "@" -f1) --EthStats.Server $(echo $ETHSTATS_LOGIN | cut -d "@" -f2)"
    fi

    if [ ! -z "$EC_CACHE_SIZE" ]; then
        CMD="$CMD --Init.MemoryHint ${EC_CACHE_SIZE}000000"
    fi

    if [ ! -z "$EC_MAX_PEERS" ]; then
        CMD="$CMD --Network.MaxActivePeers $EC_MAX_PEERS"
    fi

    if [ "$ENABLE_METRICS" = "true" ]; then
        CMD="$CMD --Metrics.Enabled true --Metrics.ExposePort $EC_METRICS_PORT"
    fi

    if [ ! -z "$EC_P2P_PORT" ]; then
        CMD="$CMD --Network.DiscoveryPort $EC_P2P_PORT --Network.P2PPort $EC_P2P_PORT"
    fi

    if [ ! -z "$RP_NETHERMIND_PRUNE_MEM_SIZE" ]; then
        CMD="$CMD --Pruning.CacheMb $RP_NETHERMIND_PRUNE_MEM_SIZE"
    fi

    if [ ! -z "$RP_NETHERMIND_FULL_PRUNE_MEMORY_BUDGET" ]; then
        CMD="$CMD --Pruning.FullPruningMemoryBudgetMb $RP_NETHERMIND_FULL_PRUNE_MEMORY_BUDGET"
    fi
    
    exec ${CMD}

fi


# Besu startup
if [ "$CLIENT" = "besu" ]; then

    # Performance tuning for ARM systems
    UNAME_VAL=$(uname -m)
    if [ "$UNAME_VAL" = "arm64" ] || [ "$UNAME_VAL" = "aarch64" ]; then

        # Define the performance tuning prefix
        define_perf_prefix

    fi

    # Create the JWT secret
    if [ ! -f "/secrets/jwtsecret" ]; then
        openssl rand -hex 32 | tr -d "\n" > /secrets/jwtsecret
    fi

    # Check for the prune flag and run that first if requested
    if [ -f "/ethclient/prune.lock" ]; then


        $PERF_PREFIX /opt/besu/bin/besu $BESU_NETWORK --data-path=/ethclient/besu --history-expiry-prune storage trie-log prune ; rm /ethclient/prune.lock

    # Run Besu normally
    else

        CMD="$PERF_PREFIX /opt/besu/bin/besu \
            $BESU_NETWORK \
            --data-path=/ethclient/besu \
            --fast-sync-min-peers=2 \
            --rpc-http-enabled \
            --rpc-http-host=0.0.0.0 \
            --rpc-http-port=${EC_HTTP_PORT:-8545} \
            --rpc-ws-enabled \
            --rpc-ws-host=0.0.0.0 \
            --rpc-ws-port=${EC_WS_PORT:-8546} \
            --host-allowlist=* \
            --rpc-http-max-active-connections=1024 \
            --nat-method=docker \
            --p2p-host=$EXTERNAL_IP \
            --engine-rpc-enabled \
            --engine-rpc-port=${EC_ENGINE_PORT:-8551} \
            --engine-host-allowlist=* \
            --engine-jwt-secret=/secrets/jwtsecret \
            --Xbonsai-full-flat-db-enabled=true \
            $EC_ADDITIONAL_FLAGS"

        if [ ! -z "$EC_SUGGESTED_BLOCK_GAS_LIMIT" ]; then
            CMD="$CMD --target-gas-limit=$EC_SUGGESTED_BLOCK_GAS_LIMIT"
        fi
        
        if [ "$EC_PRUNING_MODE" = "archive" ]; then
            CMD="$CMD --sync-mode=FULL --data-storage-format=FOREST"
        fi

        if [ "$EC_PRUNING_MODE" = "fullNode" ]; then
            CMD="$CMD --snapsync-synchronizer-pre-checkpoint-headers-only-enabled=false --snapsync-server-enabled"
        fi

        if [ "$EC_PRUNING_MODE" = "historyExpiry" ]; then
            CMD="$CMD --history-expiry-prune"
        fi

        if [ ! -z "$ETHSTATS_LABEL" ] && [ ! -z "$ETHSTATS_LOGIN" ]; then
            CMD="$CMD --ethstats $ETHSTATS_LABEL:$ETHSTATS_LOGIN"
        fi

        if [ ! -z "$EC_MAX_PEERS" ]; then
            CMD="$CMD --max-peers=$EC_MAX_PEERS"
        fi

        if [ "$ENABLE_METRICS" = "true" ]; then
            CMD="$CMD --metrics-enabled --metrics-host=0.0.0.0 --metrics-port=$EC_METRICS_PORT"
        fi

        if [ ! -z "$EC_P2P_PORT" ]; then
            CMD="$CMD --p2p-port=$EC_P2P_PORT"
        fi

        if [ ! -z "$BESU_MAX_BACK_LAYERS" ]; then
            CMD="$CMD --bonsai-historical-block-limit=$BESU_MAX_BACK_LAYERS"
        fi

        if [ "$BESU_JVM_HEAP_SIZE" -gt "0" ]; then
            CMD="env JAVA_OPTS=\"-Xmx${BESU_JVM_HEAP_SIZE}m\" $CMD"
        fi

        exec ${CMD}
    fi
fi

# Reth startup
if [ "$CLIENT" = "reth" ]; then

    # Create the JWT secret
    if [ ! -f "/secrets/jwtsecret" ]; then
        echo -n "$(head -c 32 /dev/urandom | od -A n -t x1 | tr -d '[:space:]')" > /secrets/jwtsecret
    fi

    if [ "$NETWORK" = "devnet" ]; then
            reth init --datadir /ethclient/geth --chain /devnet/genesis.json
    fi

    CMD="$PERF_PREFIX /usr/local/bin/reth node $RETH_NETWORK \
        --datadir /ethclient/reth \
        --http \
        --http.addr 0.0.0.0 \
        --http.port ${EC_HTTP_PORT:-8545} \
        --http.api eth,net,web3 \
        --http.corsdomain="*" \
        --ws \
        --ws.addr 0.0.0.0 \
        --ws.port ${EC_WS_PORT:-8546} \
        --ws.api eth,net,web3 \
        --ws.origins '*' \
        --authrpc.addr 0.0.0.0 \
        --authrpc.port ${EC_ENGINE_PORT:-8551} \
        --authrpc.jwtsecret /secrets/jwtsecret \
        $EC_ADDITIONAL_FLAGS"

    if [ ! -z "$EC_SUGGESTED_BLOCK_GAS_LIMIT" ]; then
            CMD="$CMD --builder.gaslimit $EC_SUGGESTED_BLOCK_GAS_LIMIT"
    fi
    
    if [ "$ENABLE_METRICS" = "true" ]; then
        CMD="$CMD --metrics 0.0.0.0:$EC_METRICS_PORT"
    fi

    if [ "$EC_PRUNING_MODE" = "fullNode" ]; then
        CMD="$CMD --block-interval 5"
        CMD="$CMD --prune.receipts.before 0"
        CMD="$CMD --prune.senderrecovery.full"
        CMD="$CMD --prune.accounthistory.distance 10064"
        CMD="$CMD --prune.storagehistory.distance 100064"
    fi

    if [ "$EC_PRUNING_MODE" = "historyExpiry" ]; then
        CMD="$CMD --block-interval 5"
        CMD="$CMD --prune.senderrecovery.full"
        CMD="$CMD --prune.accounthistory.distance 10064"
        CMD="$CMD --prune.storagehistory.distance 10064"
        CMD="$CMD --prune.bodies.pre-merge"
        CMD="$CMD --prune.receipts.pre-merge"
        CMD="$CMD --prune.transactionlookup.distance=10064"
    fi

    if [ ! -z "$EC_MAX_PEERS" ]; then
        CMD="$CMD --max-outbound-peers=$EC_MAX_PEERS"
    fi

    if [ ! -z "$RETH_MAX_INBOUND_PEERS" ]; then
        CMD="$CMD --max-inbound-peers=$RETH_MAX_INBOUND_PEERS"
    fi

    if [ ! -z "$EC_P2P_PORT" ]; then
        CMD="$CMD --port $EC_P2P_PORT"
    fi

    exec ${CMD}

fi
