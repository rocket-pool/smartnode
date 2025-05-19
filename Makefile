VERSION=v$(shell cat shared/version.txt)
LOCAL_OS=$(shell go env GOOS)-$(shell go env GOARCH)

BUILD_DIR=build
BIN_DIR=${BUILD_DIR}/${VERSION}/bin
DOCKER_DIR=${BUILD_DIR}/${VERSION}/docker

CLI_TARGET_OOS:=linux darwin
ARCHS:=arm64 amd64

CLI_TARGET_STRINGS:=$(foreach oos,$(CLI_TARGET_OOS), $(foreach arch,$(ARCHS),${BIN_DIR}/rocketpool-cli-$(oos)-$(arch)))
DAEMON_TARGET_STRINGS:=$(foreach arch,$(ARCHS),${BIN_DIR}/rocketpool-daemon-linux-$(arch))
TREEGEN_TARGET_STRINGS:=$(foreach arch,$(ARCHS),${BIN_DIR}/treegen-linux-$(arch))

MODULES:=$(foreach path,$(shell find . -name go.mod),$(dir $(path)))
MODULE_GLOBS:=$(foreach module,$(MODULES),$(module)...)

define rocketpool-cli-template
.PHONY: ${BIN_DIR}/rocketpool-cli-$1-$2
${BIN_DIR}/rocketpool-cli-$1-$2: ${bin_deps}
	@echo "Building rocketpool-cli-$1-$2"
ifndef NO_DOCKER
	docker run --rm -v ./:/src --user $(shell id -u):$(shell id -g) -e CGO_ENABLED=0 \
		-e GOARCH=$2 -e GOOS=$1 --workdir /src -v ~/.cache:/.cache rocketpool/smartnode-builder:${VERSION} \
		go build -o $$@ rocketpool-cli/rocketpool-cli.go
else
	CGO_ENABLED=0 GOOS=$1 GOARCH=$2 go build -o $$@ ./rocketpool-cli/rocketpool-cli.go
endif
endef

# Must be first- so `make` runs this.
.PHONY: default
default: ${BUILD_DIR}/rocketpool-cli ${BUILD_DIR}/rocketpool-daemon ${BUILD_DIR}/treegen lint

.PHONY: all
all: ${BUILD_DIR}/rocketpool-cli ${BUILD_DIR}/rocketpool-daemon ${BUILD_DIR}/treegen lint

.PHONY: release
release: ${CLI_TARGET_STRINGS} ${DAEMON_TARGET_STRINGS} ${TREEGEN_TARGET_STRINGS} ${BUILD_DIR}/rocketpool-cli ${BUILD_DIR}/rocketpool-daemon ${BUILD_DIR}/treegen

# Target for build/rocketpool-cli which is a symlink to an os-specific build
${BUILD_DIR}/rocketpool-cli: ${BIN_DIR}/rocketpool-cli-${LOCAL_OS}
	ln -sf $(shell pwd)/${BIN_DIR}/rocketpool-cli-${LOCAL_OS} ${BUILD_DIR}/rocketpool-cli


# Target for build/rocketpool-daemon which is a symlink to an os-specific build
${BUILD_DIR}/rocketpool-daemon: ${BIN_DIR}/rocketpool-daemon-${LOCAL_OS}
	ln -sf $(shell pwd)/${BIN_DIR}/rocketpool-daemon-${LOCAL_OS} ${BUILD_DIR}/rocketpool-daemon

# Target for build/treegen which is a symlink to a version-specific build
${BUILD_DIR}/treegen: ${BIN_DIR}/treegen-${LOCAL_OS}
	ln -sf $(shell pwd)/${BIN_DIR}/treegen-${LOCAL_OS} ${BUILD_DIR}/treegen

# docker-builder container
.PHONY: docker-builder
docker-builder:
	VERSION=${VERSION} docker bake -f docker/daemon-bake.hcl builder

bin_deps = ${BIN_DIR}
ifndef NO_DOCKER
	bin_deps += docker-builder
endif

docker_build_cmd_amd64 = docker run --rm -v ./:/src --user $(shell id -u):$(shell id -g) -e CGO_ENABLED=1 -e CGO_C_FLAGS="-O -D__BLST_PORTABLE__" \
	-e GOARCH=amd64 -e GOOS=linux --workdir /src -v ~/.cache:/.cache rocketpool/smartnode-builder:${VERSION} \
	go build
local_build_cmd_amd64 = CGO_ENABLED=1 CGO_C_FLAGS="-O -D__BLST_PORTABLE__" GOARCH=amd64 GOOS=linux go build
docker_build_cmd_arm64 = docker run --rm -v ./:/src --user $(shell id -u):$(shell id -g) -e CGO_ENABLED=1 -e CGO_C_FLAGS="-O -D__BLST_PORTABLE__" \
	-e CC=aarch64-linux-gnu-gcc -e CXX=aarch64-linux-gnu-cpp -e CGO_C_FLAGS="-O -D__BLST_PORTABLE__" -e GOARCH=arm64 -e GOOS=linux \
	--workdir /src -v ~/.cache:/.cache rocketpool/smartnode-builder:${VERSION} \
	go build
local_build_cmd_arm64 = CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-cpp CGO_C_FLAGS="-O -D__BLST_PORTABLE__" GOARCH=arm64 GOOS=linux go build
# amd64 daemon build
.PHONY: ${BIN_DIR}/rocketpool-daemon-linux-amd64
${BIN_DIR}/rocketpool-daemon-linux-amd64: ${bin_deps}
ifndef NO_DOCKER
	${docker_build_cmd_amd64} -o $@ rocketpool/rocketpool.go
else
	${local_build_cmd_amd64} -o $@ rocketpool/rocketpool.go
endif

# arm64 daemon build
.PHONY: ${BIN_DIR}/rocketpool-daemon-linux-arm64
${BIN_DIR}/rocketpool-daemon-linux-arm64: ${bin_deps}
ifndef NO_DOCKER
	${docker_build_cmd_arm64} -o $@ rocketpool/rocketpool.go
else
	${local_build_cmd_arm64} -o $@ rocketpool/rocketpool.go
endif

${BIN_DIR}:
	mkdir -p ${BIN_DIR}
${DOCKER_DIR}:
	mkdir -p ${DOCKER_DIR}

$(foreach oos,$(CLI_TARGET_OOS),$(foreach arch,$(ARCHS),$(eval $(call rocketpool-cli-template,$(oos),$(arch)))))

# amd64 treegen build
.PHONY: ${BIN_DIR}/treegen-linux-amd64
${BIN_DIR}/treegen-linux-amd64: ${bin_deps}
ifndef NO_DOCKER
	${docker_build_cmd_amd64} -o $@ ./treegen/.
else
	${local_build_cmd_amd64} -o $@ ./treegen/.
endif

# arm64 treegen build
.PHONY: ${BIN_DIR}/treegen-linux-arm64
${BIN_DIR}/treegen-linux-arm64: ${bin_deps}
ifndef NO_DOCKER
	${docker_build_cmd_arm64} -o $@ ./treegen/.
else
	${local_build_cmd_arm64} -o $@ ./treegen/.
endif

# Docker containers
.PHONY: docker
docker: ${DOCKER_DIR}
	VERSION=${VERSION} docker bake -f docker/daemon-bake.hcl daemon

.PHONY: docker-load
docker-load: docker
	docker import - smartnode:${VERSION}-amd64 < ${DOCKER_DIR}/smartnode:${VERSION}-amd64.tar
	docker import - smartnode:${VERSION}-arm64 < ${DOCKER_DIR}/smartnode:${VERSION}-arm64.tar

.PHONY: docker-push
docker-push: docker-load
	echo
	echo -n "Publishing smartnode:${VERSION} containers. Continue? [yN]: " && read ans && if [ $${ans:-'N'} != 'y' ]; then exit 1; fi
	docker push rocketpool/smartnode:${VERSION}-amd64
	docker push rocketpool/smartnode:${VERSION}-arm64
	echo "Done!"

.PHONY: docker-latest
docker-latest: docker-push
	echo
	echo -n "Publishing smartnode:${VERSION} as latest. Continue? [yN]: " && read ans && if [ $${ans:-'N'} != 'y' ]; then exit 1; fi
	rm -rf ~/.docker/manifests/docker.io_rocketpool_smartnode-latest
	rm -rf ~/.docker/manifests/docker.io_rocketpool_smartnode-${VERSION}
	docker manifest create rocketpool/smartnode:${VERSION} --amend rocketpool/smartnode:${VERSION}-amd64 --amend rocketpool/smartnode:${VERSION}-arm64
	docker manifest create rocketpool/smartnode:latest --amend rocketpool/smartnode:${VERSION}-amd64 --amend rocketpool/smartnode:${VERSION}-arm64
	docker manifest push --purge rocketpool/smartnode:${VERSION}
	docker manifest push --purge rocketpool/smartnode:latest

define lint-template 
.PHONY: lint-$1
lint-$1:
	docker run -e GOCACHE=/go/.cache/go-build -e GOLANGCI_LINT_CACHE=/go/.cache/golangci-lint --user $(shell id -u):$(shell id -g) --rm -v ~/.cache:/go/.cache -v .:/smartnode --workdir /smartnode/$1 golangci/golangci-lint:v2.1-alpine golangci-lint fmt --diff
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
