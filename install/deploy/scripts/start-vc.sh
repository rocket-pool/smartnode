#!/bin/sh
# This script launches ETH2 validator clients for Rocket Pool's docker stack; only edit if you know what you're doing ;)

# Set up the network-based flags
if [ "$NETWORK" = "mainnet" ]; then
    LH_NETWORK="mainnet"
    LODESTAR_NETWORK="mainnet"
    PRYSM_NETWORK="--mainnet"
    TEKU_NETWORK="mainnet"
elif [ "$NETWORK" = "devnet" ]; then
    LH_NETWORK="holesky"
    LODESTAR_NETWORK="holesky"
    PRYSM_NETWORK="--holesky"
    TEKU_NETWORK="holesky"
elif [ "$NETWORK" = "holesky" ]; then
    LH_NETWORK="holesky"
    LODESTAR_NETWORK="holesky"
    PRYSM_NETWORK="--holesky"
    TEKU_NETWORK="holesky"
else
    echo "Unknown network [$NETWORK]"
    exit 1
fi

# Report a missing fee recipient file
if [ ! -f "$VALIDATORS_DIR/$FEE_RECIPIENT_FILE" ]; then
    echo "Fee recipient file not found, please wait for the rocketpool_node process to create one."
    exit 1
fi


# Lighthouse startup
if [ "$CLIENT" = "lighthouse" ]; then

    # Set up the CC + fallback string
    BN_URL_STRING=$BN_API_ENDPOINT
    if [ ! -z "$FALLBACK_BN_API_ENDPOINT" ]; then
        BN_URL_STRING="$BN_API_ENDPOINT,$FALLBACK_BN_API_ENDPOINT"
    fi

    CMD="/usr/local/bin/lighthouse validator \
        --network $LH_NETWORK \
        --datadir $VALIDATORS_DIR/lighthouse \
        --init-slashing-protection \
        --logfile-max-number 0 \
        --beacon-nodes $BN_URL_STRING \
        --suggested-fee-recipient $(cat $VALIDATORS_DIR/$FEE_RECIPIENT_FILE) \
        $VC_ADDITIONAL_FLAGS"

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
if [ "$CLIENT" = "lodestar" ]; then

    # Remove any lock files that were left over accidentally after an unclean shutdown
    find $VALIDATORS_DIR/lodestar/validators -name voting-keystore.json.lock -delete

    # Set up the CC + fallback string
    BN_URL_STRING=$BN_API_ENDPOINT
    if [ ! -z "$FALLBACK_BN_API_ENDPOINT" ]; then
        BN_URL_STRING="$BN_API_ENDPOINT,$FALLBACK_BN_API_ENDPOINT"
    fi

    CMD="/usr/app/node_modules/.bin/lodestar validator \
        --network $LODESTAR_NETWORK \
        --dataDir $VALIDATORS_DIR/lodestar \
        --beacon-nodes $BN_URL_STRING \
        $FALLBACK_BN_STRING \
        --keystoresDir $VALIDATORS_DIR/lodestar/validators \
        --secretsDir $VALIDATORS_DIR/lodestar/secrets \
        --suggestedFeeRecipient $(cat $VALIDATORS_DIR/$FEE_RECIPIENT_FILE) \
        $VC_ADDITIONAL_FLAGS"

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
if [ "$CLIENT" = "nimbus" ]; then

    # Nimbus won't start unless the validator directories already exist
    mkdir -p $VALIDATORS_DIR/nimbus/validators
    mkdir -p $VALIDATORS_DIR/nimbus/secrets

    # Set up the fallback arg
    if [ ! -z "$FALLBACK_BN_API_ENDPOINT" ]; then
        FALLBACK_BN_ARG="--beacon-node=$FALLBACK_BN_API_ENDPOINT"
    fi

    CMD="/home/user/nimbus_validator_client \
        --non-interactive \
        --beacon-node=$BN_API_ENDPOINT $FALLBACK_BN_ARG \
        --data-dir=/ethclient/nimbus_vc \
        --validators-dir=$VALIDATORS_DIR/nimbus/validators \
        --secrets-dir=$VALIDATORS_DIR/nimbus/secrets \
        --doppelganger-detection=$DOPPELGANGER_DETECTION \
        --suggested-fee-recipient=$(cat $VALIDATORS_DIR/$FEE_RECIPIENT_FILE) \
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
if [ "$CLIENT" = "prysm" ]; then

    # Make the Prysm dir
    mkdir -p $VALIDATORS_DIR/prysm-non-hd/

    # Get rid of the protocol prefix
    BN_RPC_ENDPOINT=$(echo $BN_RPC_ENDPOINT | sed -E 's/.*\:\/\/(.*)/\1/')
    if [ ! -z "$FALLBACK_BN_RPC_ENDPOINT" ]; then
        FALLBACK_BN_RPC_ENDPOINT=$(echo $FALLBACK_BN_RPC_ENDPOINT | sed -E 's/.*\:\/\/(.*)/\1/')
    fi

    # Set up the CC + fallback string
    BN_URL_STRING=$BN_RPC_ENDPOINT
    if [ ! -z "$FALLBACK_BN_RPC_ENDPOINT" ]; then
        BN_URL_STRING="$BN_RPC_ENDPOINT,$FALLBACK_BN_RPC_ENDPOINT"
    fi

    CMD="/app/cmd/validator/validator \
        --accept-terms-of-use \
        $PRYSM_NETWORK \
        --wallet-dir $VALIDATORS_DIR/prysm-non-hd \
        --wallet-password-file $VALIDATORS_DIR/prysm-non-hd/direct/accounts/secret \
        --beacon-rpc-provider $BN_URL_STRING \
        --suggested-fee-recipient $(cat $VALIDATORS_DIR/$FEE_RECIPIENT_FILE) \
        $VC_ADDITIONAL_FLAGS"

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
if [ "$CLIENT" = "teku" ]; then

    # Teku won't start unless the validator directories already exist
    mkdir -p $VALIDATORS_DIR/teku/keys
    mkdir -p $VALIDATORS_DIR/teku/passwords

    # Remove any lock files that were left over accidentally after an unclean shutdown
    rm -f $VALIDATORS_DIR/teku/keys/*.lock

    # Set up the CC + fallback string
    BN_URL_STRING=$BN_API_ENDPOINT
    if [ ! -z "$FALLBACK_BN_API_ENDPOINT" ]; then
        BN_URL_STRING="$BN_API_ENDPOINT,$FALLBACK_BN_API_ENDPOINT"
    fi

    CMD="/opt/teku/bin/teku validator-client \
        --network=$TEKU_NETWORK \
        --data-path=$VALIDATORS_DIR/teku \
        --validator-keys=$VALIDATORS_DIR/teku/keys:$VALIDATORS_DIR/teku/passwords \
        --beacon-node-api-endpoints=$BN_URL_STRING \
        --validators-keystore-locking-enabled=false \
        --log-destination=CONSOLE \
        --validators-proposer-default-fee-recipient=$(cat $VALIDATORS_DIR/$FEE_RECIPIENT_FILE) \
        $VC_ADDITIONAL_FLAGS"

    if [ "$TEKU_SHUT_DOWN_WHEN_SLASHED" = "true" ]; then
        CMD="$CMD --shut-down-when-validator-slashed-enabled=true"
    fi

    if [ "$DOPPELGANGER_DETECTION" = "true" ]; then
        CMD="$CMD --doppelganger-detection-enabled"
    fi

    if [ "$ENABLE_MEV_BOOST" = "true" ]; then
        CMD="$CMD --validators-builder-registration-default-enabled=true"
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

