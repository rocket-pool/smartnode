#!/bin/bash

# This script will build all of the artifacts involved in a new Smart Node release.
# NOTE: You MUST put this in a directory that has the `smartnode` repository cloned as a subdirectory.

# =================
# === Functions ===
# =================

# Print a failure message to stderr and exit
fail() {
    MESSAGE=$1
    RED='\033[0;31m'
    RESET='\033[;0m'
    >&2 echo -e "\n${RED}**ERROR**\n$MESSAGE${RESET}\n"
    exit 1
}


# Builds all of the CLI binaries
build_cli() {
    cd smartnode || fail "Directory ${PWD}/smartnode does not exist or you don't have permissions to access it."

    echo -n "Building CLI binaries... "
    docker buildx build --rm -f docker/cli.dockerfile --output ../$VERSION --target cli . || fail "Error building CLI binaries."
    #rm -rf rocketpool-cli/build
    echo "done!"

    cd ..
}


# Builds the .tar.xz file packages with the RP configuration files
build_install_packages() {
    cd smartnode || fail "Directory ${PWD}/smartnode does not exist or you don't have permissions to access it."
    rm -f smartnode-install.tar.xz

    echo -n "Building Smart Node installer packages... "
    tar cfJ smartnode-install.tar.xz install/deploy || fail "Error building installer package."
    mv smartnode-install.tar.xz ../$VERSION
    cp install/install.sh ../$VERSION
    cp install/install-update-tracker.sh ../$VERSION
    echo "done!"

    echo -n "Building update tracker package... "
    tar cfJ rp-update-tracker.tar.xz install/rp-update-tracker || fail "Error building update tracker package."
    mv rp-update-tracker.tar.xz ../$VERSION
    echo "done!"

    cd ..
}


# Builds the daemon binaries and Docker Smart Node images, and pushes them to Docker Hub
# NOTE: You must install qemu first; e.g. sudo apt-get install -y qemu qemu-user-static
build_daemon() {
    cd smartnode || fail "Directory ${PWD}/smartnode does not exist or you don't have permissions to access it."

    # Make a multiarch builder, ignore if it's already there
    docker buildx create --name multiarch-builder --driver docker-container --use > /dev/null 2>&1

    echo "Building Smart Node daemon binaries..."
    docker buildx build --rm --platform=linux/amd64,linux/arm64 -f docker/daemon-build.dockerfile --output ../$VERSION --target daemon . || fail "Error building Smart Node daemon binaries."

    # Copy the daemon binaries to a build folder so the image can access them
    mkdir -p ./build
    cp ../$VERSION/linux_amd64/* ./build
    cp ../$VERSION/linux_arm64/* ./build
    echo "done!"

    echo "Building Smart Node Docker image..."
    docker buildx build --rm --platform=linux/amd64,linux/arm64 -t rocketpool/smartnode:$VERSION -f docker/daemon.dockerfile --push . || fail "Error building Smart Node Docker image."
    echo "done!"

    # Cleanup
    mv ../$VERSION/linux_amd64/* ../$VERSION
    mv ../$VERSION/linux_arm64/* ../$VERSION
    rm -rf ../$VERSION/linux_amd64/
    rm -rf ../$VERSION/linux_arm64/
    rm -rf ./build

    cd ..
}


# Tag the 'latest' Docker image and push it to Docker Hub
build_latest_docker_manifest() {
    echo -n "Tagging 'latest' Docker image... "
    docker tag rocketpool/smartnode:$VERSION rocketpool/smartnode:latest
    echo "done!"

    echo -n "Pushing to Docker Hub... "
    docker push rocketpool/smartnode:latest
    echo "done!"
}


# Print usage
usage() {
    echo "Usage: build-release.sh [options] -v <version number>"
    echo "This script assumes it is in a directory that contains subdirectories for all of the Smart Node repositories."
    echo "Options:"
    echo $'\t-a\tBuild all of the artifacts'
    echo $'\t-c\tBuild the CLI binaries for all platforms'
    echo $'\t-p\tBuild the Smart Node installer packages'
    echo $'\t-d\tBuild the Smart Node Daemon image, and push it to Docker Hub'
    echo $'\t-l\tTag the Docker image as "latest" and push it to Docker Hub'
    exit 0
}


# =================
# === Main Body ===
# =================

# Parse arguments
while getopts "acpdlv:" FLAG; do
    case "$FLAG" in
        a) CLI=true PACKAGES=true DAEMON=true LATEST_MANIFEST=true ;;
        c) CLI=true ;;
        p) PACKAGES=true ;;
        d) DAEMON=true ;;
        l) LATEST_MANIFEST=true ;;
        v) VERSION="$OPTARG" ;;
        *) usage ;;
    esac
done
if [ -z "$VERSION" ]; then
    usage
fi

# Cleanup old artifacts
rm -f ./$VERSION/*
mkdir -p ./$VERSION

# Build the artifacts
if [ "$CLI" = true ]; then
    build_cli
fi
if [ "$PACKAGES" = true ]; then
    build_install_packages
fi
if [ "$DAEMON" = true ]; then
    build_daemon
fi
if [ "$LATEST_MANIFEST" = true ]; then
    build_latest_docker_manifest
fi