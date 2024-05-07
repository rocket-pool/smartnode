#!/bin/sh
# This script launches the Execution Client for the Smart Node's Docker stack; only edit if you know what you're doing ;)

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
    GETH_NETWORK="--holesky"
    RP_NETHERMIND_NETWORK="holesky"
    BESU_NETWORK="--network=holesky"
    RETH_NETWORK="--chain holesky"
elif [ "$NETWORK" = "holesky" ]; then
    GETH_NETWORK="--holesky"
    RP_NETHERMIND_NETWORK="holesky"
    BESU_NETWORK="--network=holesky"
    RETH_NETWORK="--chain holesky"
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

        $PERF_PREFIX /usr/local/bin/geth snapshot prune-state $GETH_NETWORK --datadir /ethclient/geth ; rm /ethclient/prune.lock

    # Run Geth normally
    else

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

        if [ ! -z "$ETHSTATS_LABEL" ] && [ ! -z "$ETHSTATS_LOGIN" ]; then
            CMD="$CMD --ethstats $ETHSTATS_LABEL:$ETHSTATS_LOGIN"
        fi

        if [ ! -z "$EC_MAX_PEERS" ]; then
            CMD="$CMD --maxpeers $EC_MAX_PEERS"
        fi

        if [ ! -z "$RP_GETH_EVM_TIMEOUT" ]; then
            CMD="$CMD --rpc.evmtimeout ${RP_GETH_EVM_TIMEOUT}s"
        fi

        if [ "$RP_GETH_ARCHIVE_MODE" = "true" ]; then
            CMD="$CMD --syncmode=full --gcmode=archive"
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

    CMD="$PERF_PREFIX $NETHERMIND_BINARY \
        --config $RP_NETHERMIND_NETWORK \
        --Sync.SnapSync true \
        --datadir /ethclient/nethermind \
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
        --Pruning.FullPruningMaxDegreeOfParallelism 0 \
        --Pruning.FullPruningMemoryBudgetMb=$RP_NETHERMIND_FULL_PRUNE_MEMORY_BUDGET \
        $EC_ADDITIONAL_FLAGS"

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
        $PERF_PREFIX /opt/besu/bin/besu $BESU_NETWORK --data-path=/ethclient/besu storage x-trie-log prune ; rm /ethclient/prune.lock

    # Run Besu normally
    else
        CMD="$PERF_PREFIX /opt/besu/bin/besu \
            $BESU_NETWORK \
            --data-path=/ethclient/besu \
            --fast-sync-min-peers=3 \
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
            --Xsnapsync-synchronizer-flat-db-healing-enabled=true \
            $EC_ADDITIONAL_FLAGS"

        if [ "$BESU_ARCHIVE_MODE" = "true" ]; then
            CMD="$CMD --sync-mode=FULL --data-storage-format=FOREST"
        else 
            CMD="$CMD --sync-mode=SNAP --data-storage-format=BONSAI --Xbonsai-limit-trie-logs-enabled=true"
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
            CMD="$CMD --bonsai-maximum-back-layers-to-load=$BESU_MAX_BACK_LAYERS"
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
        --ws.origins="*" \
        --authrpc.addr 0.0.0.0 \
        --authrpc.port ${EC_ENGINE_PORT:-8551} \
        --authrpc.jwtsecret /secrets/jwtsecret \
        $EC_ADDITIONAL_FLAGS"

    if [ "$ENABLE_METRICS" = "true" ]; then
        CMD="$CMD --metrics 0.0.0.0:$EC_METRICS_PORT"
    fi

    if [ ! -z "$EC_P2P_PORT" ]; then
        CMD="$CMD --port $EC_P2P_PORT"
    fi

    exec ${CMD}
fi
