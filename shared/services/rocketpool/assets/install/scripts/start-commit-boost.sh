#!/bin/sh
# This script launches Commit-Boost PBS clients for Rocket Pool's docker stack; 

# Commit-Boost PBS doesn't accept command line arguments, everything must be in the config file which is
# specified by an environment variable.
export CB_CONFIG="/cb_config.toml"
if [ ! -f "$CB_CONFIG" ]; then
    echo "Commit-Boost config file not found at $CB_CONFIG. Please ensure the config file is correctly mounted."
    exit 1
fi
CMD="/usr/local/bin/commit-boost-pbs"
exec ${CMD}