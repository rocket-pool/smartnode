#!/bin/bash

# Build the images
build_docker() {
    echo -n "Building Docker images... "
    docker buildx build --platform=linux/amd64 -t rocketpool/smartnode-base:$VERSION-amd64 --load -f docker/smartnode-base .
    docker buildx build --platform=linux/arm64 -t rocketpool/smartnode-base:$VERSION-arm64 --load -f docker/smartnode-base .
    docker push rocketpool/smartnode-base:$VERSION-amd64
    docker push rocketpool/smartnode-base:$VERSION-arm64
    echo "done!"
}

# Build the manifest
build_manifest() {
    echo -n "Building Docker manifest... "
    rm -f ~/.docker/manifests/docker.io_rocketpool_smartnode-base-$VERSION
    rm -f ~/.docker/manifests/docker.io_rocketpool_smartnode-base-latest
    docker manifest create rocketpool/smartnode-base:$VERSION --amend rocketpool/smartnode-base:$VERSION-amd64 --amend rocketpool/smartnode-base:$VERSION-arm64
    docker manifest create rocketpool/smartnode-base:latest --amend rocketpool/smartnode-base:$VERSION-amd64 --amend rocketpool/smartnode-base:$VERSION-arm64
    echo "done!"
    echo -n "Pushing to Docker Hub... "
    docker manifest push --purge rocketpool/smartnode-base:$VERSION
    docker manifest push --purge rocketpool/smartnode-base:latest
    echo "done!"
}

# Print usage
usage() {
    echo "Usage: build-release.sh [options] -v <version number>"
    echo "This script builds the base image for the Smartnode."
    echo "Options:"
    echo $'\t-a\tBuild all of the artifacts.'
    echo $'\t-d\tBuild the Docker image'
    echo $'\t-m\tBuild the Docker manifest'
    exit 0
}

# =================
# === Main Body ===
# =================

# Get the version
while getopts "admv:" FLAG; do
    case "$FLAG" in
        a) DOCKER=true MANIFEST=true ;;
        d) DOCKER=true ;;
        m) MANIFEST=true ;;
        v) VERSION="$OPTARG" ;;
        *) usage ;;
    esac
done
if [ -z "$VERSION" ]; then
    usage
fi

# Build the artifacts
if [ "$DOCKER" = true ]; then
    build_docker
fi
if [ "$MANIFEST" = true ]; then
    build_manifest
fi
