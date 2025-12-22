#!/bin/sh
# This script launches ETH2 validator clients for Rocket Pool's docker stack; only edit if you know what you're doing ;)

GWW_GRAFFITI_FILE="/addons/gww/graffiti.txt"
echo -n "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef" > "/validators/token-file.txt"

# Set up the network-based flags
if [ "$NETWORK" = "mainnet" ]; then
    LH_NETWORK="mainnet"
    LODESTAR_NETWORK="mainnet"
    PRYSM_NETWORK="--mainnet"
    TEKU_NETWORK="mainnet"
elif [ "$NETWORK" = "devnet" ]; then
    LH_NETWORK="hoodi"
    LODESTAR_NETWORK="hoodi"
    PRYSM_NETWORK="--hoodi"
    TEKU_NETWORK="hoodi"
elif [ "$NETWORK" = "testnet" ]; then
    LH_NETWORK="hoodi"
    LODESTAR_NETWORK="hoodi"
    PRYSM_NETWORK="--hoodi"
    TEKU_NETWORK="hoodi"
else
    echo "Unknown network [$NETWORK]"
    exit 1
fi

# Report a missing fee recipient file
if [ ! -f "/validators/$FEE_RECIPIENT_FILE" ]; then
    echo "Fee recipient file not found, please wait for the rocketpool_node process to create one."
    exit 1
fi


# Lighthouse startup
if [ "$CC_CLIENT" = "lighthouse" ]; then

    # Set up the CC + fallback string
    CC_URL_STRING=$CC_API_ENDPOINT
    if [ ! -z "$FALLBACK_CC_API_ENDPOINT" ]; then
        CC_URL_STRING="$CC_API_ENDPOINT,$FALLBACK_CC_API_ENDPOINT"
    fi

    if [ "$NETWORK" != "devnet" ]; then
        CMD_LH_NETWORK="--network $LH_NETWORK"
    else
        CMD_LH_NETWORK="--testnet-dir /devnet"
    fi

    CMD="/usr/local/bin/lighthouse validator \
        --network $LH_NETWORK \
        --datadir /validators/lighthouse \
        --init-slashing-protection \
        --http \
        --http-address 0.0.0.0 \
        --http-port ${VC_KEYMANAGER_API_PORT:-5062} \
        --http-token-path  /validators/token-file.txt \
        --unencrypted-http-transport \
        --logfile-max-number 0 \
        --beacon-nodes $CC_URL_STRING \
        --suggested-fee-recipient $(cat /validators/$FEE_RECIPIENT_FILE) \
        $VC_ADDITIONAL_FLAGS"

    if [ ! -z "$VC_SUGGESTED_BLOCK_GAS_LIMIT" ]; then
            CMD="$CMD --gas-limit $VC_SUGGESTED_BLOCK_GAS_LIMIT"
    fi

    if [ "$DOPPELGANGER_DETECTION" = "true" ]; then
        CMD="$CMD --enable-doppelganger-protection"
    fi

    if [ "$ENABLE_MEV_BOOST" = "true" ]; then
        CMD="$CMD --builder-proposals --prefer-builder-proposals"
    fi

    if [ "$ENABLE_METRICS" = "true" ]; then
        CMD="$CMD --metrics --metrics-address 0.0.0.0 --metrics-port $VC_METRICS_PORT"
    fi

    if [ "$ENABLE_BITFLY_NODE_METRICS" = "true" ]; then
        CMD="$CMD --monitoring-endpoint $BITFLY_NODE_METRICS_ENDPOINT?apikey=$BITFLY_NODE_METRICS_SECRET&machine=$BITFLY_NODE_METRICS_MACHINE_NAME"
    fi

    if [ "$ADDON_GWW_ENABLED" = "true" ]; then
        echo "default: $GRAFFITI" > $GWW_GRAFFITI_FILE # Default graffiti value for Lighthouse
        exec ${CMD} --graffiti-file $GWW_GRAFFITI_FILE
    else
        exec ${CMD} --graffiti "$GRAFFITI"
    fi

fi

# Lodestar startup
if [ "$CC_CLIENT" = "lodestar" ]; then

    # Remove any lock files that were left over accidentally after an unclean shutdown
    find /validators/lodestar/validators -name voting-keystore.json.lock -delete

    # Set up the CC + fallback string
    CC_URL_STRING=$CC_API_ENDPOINT
    if [ ! -z "$FALLBACK_CC_API_ENDPOINT" ]; then
        CC_URL_STRING="$CC_API_ENDPOINT,$FALLBACK_CC_API_ENDPOINT"
    fi

    if [ "$NETWORK" != "devnet" ]; then
        CMD_NETWORK="--network $LODESTAR_NETWORK"
    else
        CMD_NETWORK="--paramsFile /devnet/config.yaml"
    fi

    CMD="/usr/app/node_modules/.bin/lodestar validator \
        --network $LODESTAR_NETWORK \
        --dataDir /validators/lodestar \
        --beacon-nodes $CC_URL_STRING \
        $FALLBACK_CC_STRING \
        --keystoresDir /validators/lodestar/validators \
        --secretsDir /validators/lodestar/secrets \
        --suggestedFeeRecipient $(cat /validators/$FEE_RECIPIENT_FILE) \
        --keymanager true \
        --keymanager.port ${VC_KEYMANAGER_API_PORT:-5062} \
        --keymanager.address 0.0.0.0 \
        --keymanager.tokenFile /validators/token-file.txt \
        $VC_ADDITIONAL_FLAGS"

if [ ! -z "$VC_SUGGESTED_BLOCK_GAS_LIMIT" ]; then
            CMD="$CMD --defaultGasLimit $VC_SUGGESTED_BLOCK_GAS_LIMIT"
    fi

    if [ "$DOPPELGANGER_DETECTION" = "true" ]; then
        CMD="$CMD --doppelgangerProtection"
    fi

    if [ "$ENABLE_MEV_BOOST" = "true" ]; then
        CMD="$CMD --builder"
    fi

    if [ "$ENABLE_METRICS" = "true" ]; then
        CMD="$CMD --metrics --metrics.address 0.0.0.0 --metrics.port $VC_METRICS_PORT"
    fi

    if [ "$ENABLE_BITFLY_NODE_METRICS" = "true" ]; then
        CMD="$CMD --monitoring.endpoint $BITFLY_NODE_METRICS_ENDPOINT?apikey=$BITFLY_NODE_METRICS_SECRET&machine=$BITFLY_NODE_METRICS_MACHINE_NAME"
    fi

    exec ${CMD} --graffiti "$GRAFFITI"

fi


# Nimbus startup
if [ "$CC_CLIENT" = "nimbus" ]; then

    # Nimbus won't start unless the validator directories already exist
    mkdir -p /validators/nimbus/validators
    mkdir -p /validators/nimbus/secrets

    # Set up the fallback arg
    if [ ! -z "$FALLBACK_CC_API_ENDPOINT" ]; then
        FALLBACK_CC_ARG="--beacon-node=$FALLBACK_CC_API_ENDPOINT"
    fi

    CMD="/home/user/nimbus_validator_client \
        --non-interactive \
        --beacon-node=$CC_API_ENDPOINT $FALLBACK_CC_ARG \
        --data-dir=/ethclient/nimbus_vc \
        --validators-dir=/validators/nimbus/validators \
        --secrets-dir=/validators/nimbus/secrets \
        --keymanager \
        --keymanager-port=${VC_KEYMANAGER_API_PORT:-5062} \
        --keymanager-address=0.0.0.0 \
        --keymanager-token-file=/validators/token-file.txt \
        --doppelganger-detection=$DOPPELGANGER_DETECTION \
        --suggested-fee-recipient=$(cat /validators/$FEE_RECIPIENT_FILE) \
        --block-monitor-type=event \
        $VC_ADDITIONAL_FLAGS"

    if [ "$ENABLE_MEV_BOOST" = "true" ]; then
        CMD="$CMD --payload-builder"
    fi

    if [ "$ENABLE_METRICS" = "true" ]; then
        CMD="$CMD --metrics --metrics-address=0.0.0.0 --metrics-port=$VC_METRICS_PORT"
    fi

    # Graffiti breaks if it's in the CMD string instead of here because of spaces
    exec ${CMD} --graffiti="$GRAFFITI"

fi


# Prysm startup
if [ "$CC_CLIENT" = "prysm" ]; then

    # Make the Prysm dir
    mkdir -p /validators/prysm-non-hd/

    # Set up the CC + fallback string
    CC_URL_STRING=$CC_RPC_ENDPOINT
    if [ ! -z "$FALLBACK_CC_RPC_ENDPOINT" ]; then
        CC_URL_STRING="$CC_RPC_ENDPOINT,$FALLBACK_CC_RPC_ENDPOINT"
    fi

    if [ "$NETWORK" != "devnet" ]; then
        CMD_NETWORK="$PRYSM_NETWORK"
    else
        CMD_NETWORK="--config-file=/devnet/config.yaml"
    fi

    CMD="/app/cmd/validator/validator \
        --accept-terms-of-use \
        $PRYSM_NETWORK \
        --datadir /validators/prysm-non-hd/direct \
        --wallet-dir /validators/prysm-non-hd \
        --rpc \
        --http-host 0.0.0.0 \
        --http-port ${VC_KEYMANAGER_API_PORT:-5062} \
        --keymanager-token-file /validators/token-file.txt \
        --wallet-password-file /validators/prysm-non-hd/direct/accounts/secret \
        --beacon-rpc-provider $CC_URL_STRING \
        --suggested-fee-recipient $(cat /validators/$FEE_RECIPIENT_FILE) \
        $VC_ADDITIONAL_FLAGS"

    if [ ! -z "$VC_SUGGESTED_BLOCK_GAS_LIMIT" ]; then
        CMD="$CMD --suggested-gas-limit=$VC_SUGGESTED_BLOCK_GAS_LIMIT"
    fi
    
    if [ "$ENABLE_MEV_BOOST" = "true" ]; then
        CMD="$CMD --enable-builder"
    fi

    if [ "$DOPPELGANGER_DETECTION" = "true" ]; then
        CMD="$CMD --enable-doppelganger"
    fi

    if [ "$ENABLE_METRICS" = "true" ]; then
        CMD="$CMD --monitoring-host 0.0.0.0 --monitoring-port $VC_METRICS_PORT"
    else
        CMD="$CMD --disable-account-metrics"
    fi

    if [ "$ADDON_GWW_ENABLED" = "true" ]; then
        echo "ordered:\n  - $GRAFFITI" > $GWW_GRAFFITI_FILE # Default graffiti value for Prysm
        exec ${CMD} --graffiti-file=$GWW_GRAFFITI_FILE
    else
        exec ${CMD} --graffiti "$GRAFFITI"
    fi

fi


# Teku startup
if [ "$CC_CLIENT" = "teku" ]; then

    # Teku won't start unless the validator directories already exist
    mkdir -p /validators/teku/keys
    mkdir -p /validators/teku/passwords

    # Remove any lock files that were left over accidentally after an unclean shutdown
    rm -f /validators/teku/keys/*.lock

    # Set up the CC + fallback string
    CC_URL_STRING=$CC_API_ENDPOINT
    if [ ! -z "$FALLBACK_CC_API_ENDPOINT" ]; then
        CC_URL_STRING="$CC_API_ENDPOINT,$FALLBACK_CC_API_ENDPOINT"
    fi

    CMD="/opt/teku/bin/teku validator-client \
        --network=$TEKU_NETWORK \
        --data-path=/validators/teku \
        --validator-keys=/validators/teku/keys:/validators/teku/passwords \
        --beacon-node-api-endpoints=$CC_URL_STRING \
        --validators-keystore-locking-enabled=false \
        --log-destination=CONSOLE \
        --validator-api-enabled=true \
        --validator-api-port=${VC_KEYMANAGER_API_PORT:-5062} \
        --validator-api-interface=0.0.0.0 \
        --validator-api-host-allowlist=* \
        --validator-api-bearer-file=/validators/token-file.txt \
        --Xvalidator-api-ssl-enabled=false \
        --Xvalidator-api-unsafe-hosts-enabled=true \
        --validators-proposer-default-fee-recipient=$(cat /validators/$FEE_RECIPIENT_FILE) \
        $VC_ADDITIONAL_FLAGS"

    if [ "$DOPPELGANGER_DETECTION" = "true" ]; then
        CMD="$CMD --doppelganger-detection-enabled"
    fi

    if [ "$ENABLE_MEV_BOOST" = "true" ]; then
        CMD="$CMD --validators-builder-registration-default-enabled=true"
        if [ ! -z "$BN_SUGGESTED_BLOCK_GAS_LIMIT" ]; then
            CMD="$CMD --validators-builder-registration-default-gas-limit=$BN_SUGGESTED_BLOCK_GAS_LIMIT"
        fi
    fi

    if [ "$TEKU_USE_SLASHING_PROTECTION" = "true" ]; then
        CMD="$CMD --shut-down-when-validator-slashed-enabled=true"
    fi

    if [ "$ENABLE_METRICS" = "true" ]; then
        CMD="$CMD --metrics-enabled=true --metrics-interface=0.0.0.0 --metrics-port=$VC_METRICS_PORT --metrics-host-allowlist=*"
    fi

    if [ "$ENABLE_BITFLY_NODE_METRICS" = "true" ]; then
        CMD="$CMD --metrics-publish-endpoint=$BITFLY_NODE_METRICS_ENDPOINT?apikey=$BITFLY_NODE_METRICS_SECRET&machine=$BITFLY_NODE_METRICS_MACHINE_NAME"
    fi

    if [ "$ADDON_GWW_ENABLED" = "true" ]; then
        echo "$GRAFFITI" > $GWW_GRAFFITI_FILE # Default graffiti value for Teku
        exec ${CMD} --validators-graffiti-file=$GWW_GRAFFITI_FILE
    else
        exec ${CMD} --validators-graffiti="$GRAFFITI"
    fi

fi

