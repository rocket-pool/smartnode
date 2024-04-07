#!/bin/bash

# This script will build all of the artifacts involved in a new Smart Node release.

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
    echo -n "Building CLI binaries... "
    docker buildx build --rm -f docker/cli.dockerfile --output build/$VERSION --target cli . || fail "Error building CLI binaries."
    echo "done!"
}


# Builds the smart node distro packages
build_distro_packages() {
    echo -n "Building deb packages..."
    docker buildx build --rm -f install/packages/debian/package.dockerfile --output build/$VERSION --target package . || fail "Error building deb packages."
    rm build/$VERSION/*.build build/$VERSION/*.buildinfo build/$VERSION/*.changes
    echo "done!"
}


# Builds the .tar.xz file packages with the RP configuration files
build_install_packages() {
    echo -n "Building Smart Node installer packages... "
    tar cfJ build/$VERSION/smartnode-install.tar.xz install/deploy || fail "Error building installer package."
    cp install/install.sh build/$VERSION
    cp install/install-update-tracker.sh build/$VERSION
    echo "done!"

    echo -n "Building update tracker package... "
    tar cfJ build/$VERSION/rp-update-tracker.tar.xz install/rp-update-tracker || fail "Error building update tracker package."
    echo "done!"
}


# Builds the daemon binaries and Docker Smart Node images, and pushes them to Docker Hub
# NOTE: You must install qemu first; e.g. sudo apt-get install -y qemu qemu-user-static
build_daemon() {
    echo "Building Smart Node daemon binaries..."
    docker buildx build --rm --platform=linux/amd64,linux/arm64 -f docker/daemon-build.dockerfile --output build/$VERSION --target daemon . || fail "Error building Smart Node daemon binaries."
    echo "done!"

    # Flatted the folders to make it easier to upload artifacts to github
    mv build/$VERSION/linux_amd64/rocketpool-daemon build/$VERSION/rocketpool-daemon-linux-amd64
    mv build/$VERSION/linux_arm64/rocketpool-daemon build/$VERSION/rocketpool-daemon-linux-arm64

    # Clean up the empty directories
    rmdir build/$VERSION/linux_amd64 build/$VERSION/linux_arm64

    echo "Building Smart Node Docker image..."
    # If uploading, make and push a manifest
    if [ "$UPLOAD" = true ]; then
        docker buildx build --rm --platform=linux/amd64,linux/arm64 --build-arg BINARIES_PATH=build/$VERSION -t rocketpool/smartnode:$VERSION -f docker/daemon.dockerfile --push . || fail "Error building Smart Node Docker image."
    else
        docker buildx build --rm --load --build-arg BINARIES_PATH=build/$VERSION -t rocketpool/smartnode:$VERSION -f docker/daemon.dockerfile . || fail "Error building Smart Node Docker image."
    fi
    echo "done!"
}


# Tag the 'latest' Docker image and push it to Docker Hub
build_latest_docker_manifest() {
    echo -n "Tagging 'latest' Docker image... "
    docker tag rocketpool/smartnode:$VERSION rocketpool/smartnode:latest
    echo "done!"

    if [ "$UPLOAD" = true ]; then
        echo -n "Pushing to Docker Hub... "
        docker push rocketpool/smartnode:latest
        echo "done!"
    else
        echo "The image tag only exists locally. Rerun with -u to upload to Docker Hub."
    fi
}


# Print usage
usage() {
    echo "Usage: build-release.sh [options] -v <version number>"
    echo "This script assumes it is in a directory that contains subdirectories for all of the Smart Node repositories."
    echo "Options:"
    echo $'\t-a\tBuild all of the artifacts'
    echo $'\t-c\tBuild the CLI binaries for all platforms'
    echo $'\t-t\tBuild the distro packages (.deb)'
    echo $'\t-p\tBuild the Smart Node installer packages'
    echo $'\t-d\tBuild the Smart Node Daemon image'
    echo $'\t-l\tTag the Docker image as "latest"'
    echo $'\t-u\tWhen passed with a build, upload the resulting image tags to Docker Hub'
    exit 0
}


# =================
# === Main Body ===
# =================

# Parse arguments
while getopts "actpdluv:" FLAG; do
    case "$FLAG" in
        a) CLI=true DISTRO=true PACKAGES=true DAEMON=true ;;
        c) CLI=true ;;
        t) DISTRO=true ;;
        p) PACKAGES=true ;;
        d) DAEMON=true ;;
        l) LATEST_MANIFEST=true ;;
        u) UPLOAD=true ;;
        v) VERSION="$OPTARG" ;;
        *) usage ;;
    esac
done
if [ -z "$VERSION" ]; then
    usage
fi

# Cleanup old artifacts
rm -rf build/$VERSION/*
mkdir -p build/$VERSION

# Make a multiarch builder, ignore if it's already there
docker buildx create --name multiarch-builder --driver docker-container --use > /dev/null 2>&1

# Build the artifacts
if [ "$CLI" = true ]; then
    build_cli
fi
if [ "$DISTRO" = true ]; then
    build_distro_packages
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


# =======================
# === Manual Routines ===
# =======================

# Builds the deb package builder image
build_deb_builder() {
    echo -n "Building deb builder..."
    docker buildx build --rm --platform=linux/amd64,linux/arm64 -t rocketpool/smartnode-deb-builder:$VERSION -f install/packages/debian/builder.dockerfile --push . || fail "Error building deb builder."
    echo "done!"
}


# Builds the prune provisioner image and pushes it to Docker Hub
build_docker_prune_provision() {
    echo "Building Prune Provisioner image..."
    docker buildx build --platform=linux/amd64,linux/arm64 -t rocketpool/eth1-prune-provision:$VERSION -f util-containers/prune-provision.dockerfile --push . || fail "Error building Prune Provision image."
    echo "done!"
}


# Builds the EC migrator image and pushes it to Docker Hub
build_ec_migrator() {
    echo "Building EC Migrator image..."
    docker buildx build --platform=linux/amd64,linux/arm64 -t rocketpool/ec-migrator:$VERSION -f util-containers/ec-migrator.dockerfile --push . || fail "Error building EC Migrator image."
    echo "done!"
}
