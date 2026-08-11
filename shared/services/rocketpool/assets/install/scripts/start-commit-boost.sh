#!/bin/sh
# This script launches Commit-Boost PBS clients for Rocket Pool's docker stack; 

# PBS settings come from the config file (via CB_CONFIG); the only CLI arg is the service subcommand.
export CB_CONFIG="/cb_config.toml"
if [ ! -f "$CB_CONFIG" ]; then
    echo "Commit-Boost config file not found at $CB_CONFIG. Please ensure the config file is correctly mounted."
    exit 1
fi
CMD="/usr/local/bin/commit-boost"
exec ${CMD} pbs