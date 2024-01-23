#!/bin/bash

# This script will build all of the artifacts involved in a new Rocket Pool smartnode release
# (except for the macOS daemons, which need to be built manually on a macOS system) and put
# them into a convenient folder for ease of uploading.

# NOTE: You MUST put this in a directory that has the `smartnode` and `smartnode-install`
# repositories cloned as subdirectories.


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
    cd smartnode || fail "Directory ${PWD}/smartnode/rocketpool-cli does not exist or you don't have permissions to access it."

    echo -n "Building CLI binaries... "
    docker run --rm -v $PWD:/smartnode rocketpool/smartnode-builder:latest /smartnode/rocketpool-cli/build.sh || fail "Error building CLI binaries."
    mv rocketpool-cli/rocketpool-cli-* ../$VERSION
    echo "done!"

    cd ..
}


# Builds the .tar.xz file packages with the RP configuration files
build_install_packages() {
    cd smartnode-install || fail "Directory ${PWD}/smartnode-install does not exist or you don't have permissions to access it."
    rm -f rp-smartnode-install.tar.xz

    echo -n "Building Smartnode installer packages... "
    tar cfJ rp-smartnode-install.tar.xz install || fail "Error building installer package."
    mv rp-smartnode-install.tar.xz ../$VERSION
    cp install.sh ../$VERSION
    cp install-update-tracker.sh ../$VERSION
    echo "done!"

    echo -n "Building update tracker package... "
    tar cfJ rp-update-tracker.tar.xz rp-update-tracker || fail "Error building update tracker package."
    mv rp-update-tracker.tar.xz ../$VERSION
    echo "done!"

    cd ..
}


# Builds the daemon binaries and Docker Smartnode images, and pushes them to Docker Hub
# NOTE: You must install qemu first; e.g. sudo apt-get install -y qemu qemu-user-static
build_daemon() {
    cd smartnode || fail "Directory ${PWD}/smartnode does not exist or you don't have permissions to access it."

    echo -n "Building Daemon binary... "
    ./daemon-build.sh || fail "Error building daemon binary."
    cp rocketpool/rocketpool-daemon-* ../$VERSION
    echo "done!"

    echo "Building Docker Smartnode image..."
    docker buildx build --platform=linux/amd64 -t rocketpool/smartnode:$VERSION-amd64 -f docker/rocketpool-dockerfile --load . || fail "Error building amd64 Docker Smartnode image."
    docker buildx build --platform=linux/arm64 -t rocketpool/smartnode:$VERSION-arm64 -f docker/rocketpool-dockerfile --load . || fail "Error building arm64 Docker Smartnode image."
    echo "done!"

    echo -n "Pushing to Docker Hub... "
    docker push rocketpool/smartnode:$VERSION-amd64 || fail "Error pushing amd64 Docker Smartnode image to Docker Hub."
    docker push rocketpool/smartnode:$VERSION-arm64 || fail "Error pushing arm Docker Smartnode image to Docker Hub."
    rm -f rocketpool/rocketpool-daemon-*
    echo "done!"

    cd ..
}


# Builds the Docker prune provisioner image and pushes it to Docker Hub
build_docker_prune_provision() {
    cd smartnode || fail "Directory ${PWD}/smartnode does not exist or you don't have permissions to access it."

    echo "Building Docker Prune Provisioner image..."
    docker buildx build --platform=linux/amd64 -t rocketpool/smartnode:$VERSION-amd64 -f docker/rocketpool-prune-provision --load . || fail "Error building amd64 Docker Prune Provision  image."
    docker buildx build --platform=linux/arm64 -t rocketpool/smartnode:$VERSION-arm64 -f docker/rocketpool-prune-provision --load . || fail "Error building arm64 Docker Prune Provision  image."
    echo "done!"

    echo -n "Pushing to Docker Hub... "
    docker push rocketpool/eth1-prune-provision:$VERSION-amd64 || fail "Error pushing amd64 Docker Prune Provision image to Docker Hub."
    docker push rocketpool/eth1-prune-provision:$VERSION-arm64 || fail "Error pushing arm Docker Prune Provision image to Docker Hub."
    echo "done!"

    cd ..
}


# Builds the Docker Manifests and pushes them to Docker Hub
build_docker_manifest() {
    echo -n "Building Docker manifest... "
    rm -f ~/.docker/manifests/docker.io_rocketpool_smartnode-$VERSION
    docker manifest create rocketpool/smartnode:$VERSION --amend rocketpool/smartnode:$VERSION-amd64 --amend rocketpool/smartnode:$VERSION-arm64
    echo "done!"

    echo -n "Pushing to Docker Hub... "
    docker manifest push --purge rocketpool/smartnode:$VERSION
    echo "done!"
}


# Builds the 'latest' Docker Manifests and pushes them to Docker Hub
build_latest_docker_manifest() {
    echo -n "Building 'latest' Docker manifest... "
    rm -f ~/.docker/manifests/docker.io_rocketpool_smartnode-latest
    docker manifest create rocketpool/smartnode:latest --amend rocketpool/smartnode:$VERSION-amd64 --amend rocketpool/smartnode:$VERSION-arm64
    echo "done!"

    echo -n "Pushing to Docker Hub... "
    docker manifest push --purge rocketpool/smartnode:latest
    echo "done!"
}


# Builds the Docker Manifest for the prune provisioner and pushes it to Docker Hub
build_docker_prune_provision_manifest() {
    echo -n "Building Docker Prune Provision manifest... "
    rm -f ~/.docker/manifests/docker.io_rocketpool_eth1-prune-provision-$VERSION
    docker manifest create rocketpool/eth1-prune-provision:$VERSION --amend rocketpool/eth1-prune-provision:$VERSION-amd64 --amend rocketpool/eth1-prune-provision:$VERSION-arm64
    echo "done!"

    echo -n "Pushing to Docker Hub... "
    docker manifest push --purge rocketpool/eth1-prune-provision:$VERSION
    echo "done!"
}


# Print usage
usage() {
    echo "Usage: build-release.sh [options] -v <version number>"
    echo "This script assumes it is in a directory that contains subdirectories for all of the Rocket Pool repositories."
    echo "Options:"
    echo $'\t-a\tBuild all of the artifacts, except for the prune provisioner'
    echo $'\t-c\tBuild the CLI binaries for all platforms'
    echo $'\t-p\tBuild the Smartnode installer packages'
    echo $'\t-d\tBuild the Daemon binaries and Docker Smartnode images, and push them to Docker Hub'
    echo $'\t-x\tBuild the Docker POW Proxy image and push it to Docker Hub'
    echo $'\t-n\tBuild the Docker manifests (Smartnode and POW Proxy), and push them to Docker Hub'
    echo $'\t-r\tBuild the Docker Prune Provisioner image and push it to Docker Hub'
    echo $'\t-f\tBuild the Docker manifest for the Prune Provisioner and push it to Docker Hub'
    exit 0
}


# =================
# === Main Body ===
# =================

# Parse arguments
while getopts "acpdnlrfv:" FLAG; do
    case "$FLAG" in
        a) CLI=true PACKAGES=true DAEMON=true MANIFEST=true LATEST_MANIFEST=true ;;
        c) CLI=true ;;
        p) PACKAGES=true ;;
        d) DAEMON=true ;;
        n) MANIFEST=true ;;
        l) LATEST_MANIFEST=true ;;
        r) PRUNE=true ;;
        f) PRUNE_MANIFEST=true ;;
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
if [ "$MANIFEST" = true ]; then
    build_docker_manifest
fi
if [ "$LATEST_MANIFEST" = true ]; then
    build_latest_docker_manifest
fi
if [ "$PRUNE" = true ]; then
    build_docker_prune_provision
fi
if [ "$PRUNE_MANIFEST" = true ]; then
    build_docker_prune_provision_manifest
fi
