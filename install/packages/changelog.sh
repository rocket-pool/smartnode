#!/bin/bash

mark_for_release() {
    # Debian
    docker run -it --rm -w /rocketpool -v ./debian:/rocketpool --entrypoint dch rocketpool/smartnode-deb-builder:v1.0.1 -r -M
}

update_changelog() {
    # Debian
    docker run -it --rm -w /rocketpool -v ./debian:/rocketpool --entrypoint dch rocketpool/smartnode-deb-builder:v1.0.1 -M -v $VERSION
}

# Print usage
usage() {
    echo "Usage: changelog.sh [options] -v <version number>"
    echo "Options:"
    echo $'\t-u\tCreate a new entry to the changelog for the distro packages, or update the existing one if it isn\'t marked for release'
    echo $'\t-r\tMark the existing top-level changelog entry as ready for release'
    exit 0
}


# =================
# === Main Body ===
# =================

# Parse arguments
while getopts "urv:" FLAG; do
    case "$FLAG" in
        u) UPDATE=true ;;
        r) RELEASE=true ;;
        v) VERSION="$OPTARG" ;;
        *) usage ;;
    esac
done
if [ -z "$VERSION" ]; then
    usage
fi

# Run the command
if [ "$RELEASE" = true ]; then
    mark_for_release
elif [ "$UPDATE" = true ]; then
    update_changelog
else
    usage
fi
