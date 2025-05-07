#!/bin/bash

# Build the images
build_docker() {
    local PUSH=$1
    echo -n "Building Docker images... "
    docker buildx build --no-cache --progress=plain --platform=linux/amd64 -t rocketpool/treegen:$VERSION-amd64 --load -f ./docker/treegen .
    docker buildx build --platform=linux/arm64 -t rocketpool/treegen:$VERSION-arm64 --load -f ./docker/treegen .
    if [ "$PUSH" = "true" ]; then
        echo -n "Pushing to Docker Hub... "
        docker push rocketpool/treegen:$VERSION-amd64
        docker push rocketpool/treegen:$VERSION-arm64
    fi
    echo "done!"
}

# Build the manifest
build_manifest() {
    local PUSH=$1
    echo -n "Building Docker manifest... "
    rm -f ~/.docker/manifests/docker.io_rocketpool_treegen-$VERSION
    rm -f ~/.docker/manifests/docker.io_rocketpool_treegen-latest
    docker manifest create rocketpool/treegen:$VERSION --amend rocketpool/treegen:$VERSION-amd64 --amend rocketpool/treegen:$VERSION-arm64
    docker manifest create rocketpool/treegen:latest --amend rocketpool/treegen:$VERSION-amd64 --amend rocketpool/treegen:$VERSION-arm64
    echo "done!"
    if [ "$PUSH" = "true" ]; then
        echo -n "Pushing to Docker Hub... "
        docker manifest push --purge rocketpool/treegen:$VERSION
        docker manifest push --purge rocketpool/treegen:latest
    fi
    echo "done!"
}

# Print usage
usage() {
    echo "Usage: build-release.sh [options] -v <version number>"
    echo "This script builds treegen's artifacts."
    echo "Options:"
    echo $'\t-a\tBuild all of the artifacts.'
    echo $'\t-d\tBuild the Alpine-based Docker containers'
    echo $'\t-m\tBuild the Docker manifest'
    echo $'\t-p\tPush the docker manifests and/or containers'
    exit 0
}

# =================
# === Main Body ===
# =================

# Get the version
while getopts "admpv:" FLAG; do
    case "$FLAG" in
        a) DOCKER=true MANIFEST=true ;;
        d) DOCKER=true ;;
        m) MANIFEST=true ;;
	p) PUSH=true ;;
        v) VERSION="$OPTARG" ;;
        *) usage ;;
    esac
done
if [ -z "$VERSION" ]; then
    usage
fi

# Build the artifacts
if [ "$DOCKER" = true ]; then
    build_docker $PUSH
fi
if [ "$MANIFEST" = true ]; then
    build_manifest $PUSH
fi
