VERSION=$(shell cat shared/version.txt)
LOCAL_OS=$(shell go env GOOS)-$(shell go env GOARCH)

BUILD_DIR=build
CLI_DIR=${BUILD_DIR}/${VERSION}/cli
DAEMON_DIR=${BUILD_DIR}/${VERSION}/daemon

CLI_TARGET_OOS:=linux darwin
ARCHS:=arm64 amd64

CLI_TARGET_STRINGS:=$(foreach oos,$(CLI_TARGET_OOS), $(foreach arch,$(ARCHS),${CLI_DIR}/rocketpool-cli-$(oos)-$(arch)))
DAEMON_TARGET_STRINGS:=$(foreach arch,$(ARCHS),${DAEMON_DIR}/rocketpool-daemon-linux-$(arch))

MODULES:=$(foreach path,$(shell find . -name go.mod),$(dir $(path)))
MODULE_GLOBS:=$(foreach module,$(MODULES),$(module)...)

define rocketpool-cli-template
.PHONY: ${CLI_DIR}/rocketpool-cli-$1-$2
${CLI_DIR}/rocketpool-cli-$1-$2: ${CLI_DIR}
	@echo "Building rocketpool-cli-$1-$2"
	CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build -o $$@ ./rocketpool-cli/rocketpool-cli.go
endef

.PHONY: all
all: ${BUILD_DIR}/rocketpool-cli ${BUILD_DIR}/rocketpool-daemon lint

.PHONY: release
release: clean docker ${CLI_TARGET_STRINGS} ${DAEMON_TARGET_STRINGS}

# Target for build/rocketpool-cli which is a symlink to an os-specific build
${BUILD_DIR}/rocketpool-cli: ${CLI_DIR}/rocketpool-cli-${LOCAL_OS}
	ln -sf $(shell pwd)/${CLI_DIR}/rocketpool-cli-${LOCAL_OS} ${BUILD_DIR}/rocketpool-cli


# Target for build/rocketpool-daemon which is a symlink to an os-specific build
${BUILD_DIR}/rocketpool-daemon: ${DAEMON_DIR}/rocketpool-daemon-${LOCAL_OS}
	ln -sf $(shell pwd)/${DAEMON_DIR}/rocketpool-daemon-${LOCAL_OS} ${BUILD_DIR}/rocketpool-daemon

# amd64 daemon build
.PHONY: ${DAEMON_DIR}/rocketpool-daemon-linux-amd64
${DAEMON_DIR}/rocketpool-daemon-linux-amd64: ${DAEMON_DIR}
	CGO_ENABLED=1 CGO_C_FLAGS="-O -D__BLST_PORTABLE__" GOARCH=amd64 GOOS=linux go build -o $@ rocketpool/rocketpool.go

# arm64 daemon build
.PHONY: ${DAEMON_DIR}/rocketpool-daemon-linux-arm64
${DAEMON_DIR}/rocketpool-daemon-linux-arm64: ${DAEMON_DIR}
	CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-cpp CGO_C_FLAGS="-O -D__BLST_PORTABLE__" GOARCH=arm64 GOOS=linux go build -o $@ rocketpool/rocketpool.go

${CLI_DIR}:
	mkdir -p ${CLI_DIR}
${DAEMON_DIR}:
	mkdir -p ${DAEMON_DIR}

$(foreach oos,$(CLI_TARGET_OOS),$(foreach arch,$(ARCHS),$(eval $(call rocketpool-cli-template,$(oos),$(arch)))))


# Docker containers
.PHONY: docker
docker:
	rm -f ~/.docker/manifests/docker.io_rocketpool_smartnode-latest
	rm -f ~/.docker/manifests/docker.io_rocketpool_smartnode-${VERSION}
	VERSION=${VERSION} docker bake -f docker/daemon-bake.hcl daemon
	docker manifest create rocketpool/smartnode:${VERSION} --amend rocketpool/smartnode:${VERSION}-amd64 --amend rocketpool/smartnode:${VERSION}-arm64
	docker manifest create rocketpool/smartnode:latest --amend rocketpool/smartnode:${VERSION}-amd64 --amend rocketpool/smartnode:${VERSION}-arm64

define lint-template 
.PHONY: lint-$1
lint-$1:
	docker run --rm -v .:/go/smartnode --workdir /go/smartnode/$1 golangci/golangci-lint:v2.1-alpine golangci-lint fmt --diff
endef
$(foreach module,$(MODULES),$(eval $(call lint-template,$(module))))
.PHONY: lint
lint: $(foreach module,$(MODULES),lint-$(module))

.PHONY: test
test:
	go test -test.timeout 20m $(MODULE_GLOBS)

.PHONY: clean
clean:
	rm -rf ${BUILD_DIR}
