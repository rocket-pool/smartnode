#!/bin/bash


##
# Performs initial node setup, including:
# - Seeding node account with ETH and RPL
# - Registering node
# - Making node trusted
# - Making node deposits with various staking durations
# - Making user deposits to launch node minipools
##


##
# Config
##


GROUP_NAME="Group0"


##
# Setup
##


# Get scripts path
SCRIPTS_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# Get node account address
NODE_STATUS="$( go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" node status )"
if [[ $NODE_STATUS =~ (0x[a-fA-F0-9]{40}) ]] ; then
    NODE_ACCOUNT_ADDRESS="${BASH_REMATCH[1]}"
else
    echo "Could not get node account address"
    exit 1
fi

# Configure staking durations (in blocks)
node "${SCRIPTS_PATH}/set-minipool-setting.js" setMinipoolStakingDuration 3m 5
node "${SCRIPTS_PATH}/set-minipool-setting.js" setMinipoolStakingDuration 6m 10

# Seed node account with ETH and RPL
node "${SCRIPTS_PATH}/send-ether.js" $NODE_ACCOUNT_ADDRESS 500000
node "${SCRIPTS_PATH}/mint-rpl.js" $NODE_ACCOUNT_ADDRESS 500000

# Register node
go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" node register

# Make node trusted
node "${SCRIPTS_PATH}/set-node-trusted.js" $NODE_ACCOUNT_ADDRESS true

# Make node deposits
go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" deposit reserve 3m
go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" deposit complete
go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" deposit reserve 6m
go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" deposit complete
go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" deposit reserve 12m
go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" deposit complete
go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" deposit reserve 3m
go run "${SCRIPTS_PATH}/../rocketpool-cli/rocketpool-cli.go" deposit complete

# Create group depositor
CREATE_GROUP_ACCESSOR_OUTPUT="$( node "${SCRIPTS_PATH}/create-group-accessor.js" $GROUP_NAME )"
echo $CREATE_GROUP_ACCESSOR_OUTPUT
if [[ $CREATE_GROUP_ACCESSOR_OUTPUT =~ (0x[a-fA-F0-9]{40}).$ ]] ; then
    GROUP_DEPOSITOR_ADDRESS="${BASH_REMATCH[1]}"
else
    echo "Could not create group accessor"
    exit 1
fi

# Make user deposits
node "${SCRIPTS_PATH}/user-deposit.js" $GROUP_DEPOSITOR_ADDRESS 3m 8
node "${SCRIPTS_PATH}/user-deposit.js" $GROUP_DEPOSITOR_ADDRESS 3m 8
node "${SCRIPTS_PATH}/user-deposit.js" $GROUP_DEPOSITOR_ADDRESS 6m 8
node "${SCRIPTS_PATH}/user-deposit.js" $GROUP_DEPOSITOR_ADDRESS 6m 8
node "${SCRIPTS_PATH}/user-deposit.js" $GROUP_DEPOSITOR_ADDRESS 12m 8
node "${SCRIPTS_PATH}/user-deposit.js" $GROUP_DEPOSITOR_ADDRESS 12m 8

