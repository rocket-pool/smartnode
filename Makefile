VERSION=v$(shell cat shared/version.txt)
LOCAL_OS=$(shell go env GOOS)
LOCAL_ARCH=$(shell go env GOARCH)
# Needed for binary artifacts of format `rocketpool-cli-linux-amd64`
LOCAL_TARGET=${LOCAL_OS}-${LOCAL_ARCH}
# Needed for docker --platform arguments of format `linux/amd64`
LOCAL_PLATFORM=${LOCAL_OS}/${LOCAL_ARCH}

BUILD_DIR=build
BIN_DIR=${BUILD_DIR}/${VERSION}/bin
TOOLS_DIR=${BUILD_DIR}/${VERSION}/tools

CLI_TARGET_OOS:=linux darwin
ARCHS:=arm64 amd64

CLI_TARGET_STRINGS:=$(foreach oos,$(CLI_TARGET_OOS), $(foreach arch,$(ARCHS),${BIN_DIR}/rocketpool-cli-$(oos)-$(arch)))
DAEMON_TARGET_STRINGS:=$(foreach arch,$(ARCHS),${BIN_DIR}/rocketpool-daemon-linux-$(arch))
TREEGEN_TARGET_STRINGS:=$(foreach arch,$(ARCHS),${BIN_DIR}/treegen-linux-$(arch))

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
${BUILD_DIR}/rocketpool-cli: ${BIN_DIR}/rocketpool-cli-${LOCAL_TARGET}
	ln -sf $(shell pwd)/${BIN_DIR}/rocketpool-cli-${LOCAL_TARGET} ${BUILD_DIR}/rocketpool-cli


# Target for build/rocketpool-daemon which is a symlink to an os-specific build
${BUILD_DIR}/rocketpool-daemon: ${BIN_DIR}/rocketpool-daemon-${LOCAL_TARGET}
	ln -sf $(shell pwd)/${BIN_DIR}/rocketpool-daemon-${LOCAL_TARGET} ${BUILD_DIR}/rocketpool-daemon

# Target for build/treegen which is a symlink to a version-specific build
${BUILD_DIR}/treegen: ${BIN_DIR}/treegen-${LOCAL_TARGET}
	ln -sf $(shell pwd)/${BIN_DIR}/treegen-${LOCAL_TARGET} ${BUILD_DIR}/treegen

# docker-builder container
.PHONY: docker-builder
docker-builder: ${BUILD_DIR}/docker-buildx-builder
	VERSION=${VERSION} docker bake --builder smartnode-builder -f docker/daemon-bake.hcl builder

bin_deps = ${BIN_DIR}
ifndef NO_DOCKER
	bin_deps += docker-builder
endif

docker_build_cmd_amd64 = docker run --rm -v ./:/src --user $(shell id -u):$(shell id -g) -e CGO_ENABLED=1 -e CGO_CFLAGS="-O -D__BLST_PORTABLE__" \
	-e GOARCH=amd64 -e GOOS=linux --workdir /src -v ~/.cache:/.cache rocketpool/smartnode-builder:${VERSION} \
	go build
local_build_cmd_amd64 = CGO_ENABLED=1 CGO_CFLAGS="-O -D__BLST_PORTABLE__" GOARCH=amd64 GOOS=linux go build
docker_build_cmd_arm64 = docker run --rm -v ./:/src --user $(shell id -u):$(shell id -g) -e CGO_ENABLED=1 -e CGO_CFLAGS="-O -D__BLST_PORTABLE__" \
	-e CC=aarch64-linux-gnu-gcc -e CXX=aarch64-linux-gnu-cpp -e CGO_CFLAGS="-O -D__BLST_PORTABLE__" -e GOARCH=arm64 -e GOOS=linux \
	--workdir /src -v ~/.cache:/.cache rocketpool/smartnode-builder:${VERSION} \
	go build
local_build_cmd_arm64 = CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-cpp CGO_CFLAGS="-O -D__BLST_PORTABLE__" GOARCH=arm64 GOOS=linux go build
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

${BUILD_DIR}:
	mkdir -p ${BUILD_DIR}
${BIN_DIR}:
	mkdir -p ${BIN_DIR}
${TOOLS_DIR}:
	mkdir -p ${TOOLS_DIR}

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

# amd64 state-cli build
.PHONY: ${BIN_DIR}/state-cli-linux-amd64
${TOOLS_DIR}/state-cli-linux-amd64: ${bin_deps}
ifndef NO_DOCKER
	${docker_build_cmd_amd64} -o $@ ./shared/services/state/cli/.
else
	${local_build_cmd_amd64} -o $@ ./shared/services/state/cli/.
endif

# arm64 state-cli build
.PHONY: ${BIN_DIR}/state-cli-linux-arm64
${TOOLS_DIR}/state-cli-linux-arm64: ${bin_deps}
ifndef NO_DOCKER
	${docker_build_cmd_arm64} -o $@ ./shared/services/state/cli/.
else
	${local_build_cmd_arm64} -o $@ ./shared/services/state/cli/.
endif

# Multiarch builder
${BUILD_DIR}/docker-buildx-builder: ${BUILD_DIR}
	docker buildx create --name smartnode-builder --driver docker-container --platform linux/amd64,linux/arm64 || true
	touch ${BUILD_DIR}/docker-buildx-builder

# Docker containers
.PHONY: docker
docker: ${BUILD_DIR}/docker-buildx-builder
	# override the platform so we can load the resulting image into docker
	VERSION=${VERSION} docker bake --builder smartnode-builder -f docker/daemon-bake.hcl smartnode --set "smartnode.platform=${LOCAL_PLATFORM}"

.PHONY: docker-push
docker-push: ${BUILD_DIR}/docker-buildx-builder
	echo -n "Building ${VERSION} and publishing containers. Continue? [yN]: " && read ans && if [ $${ans:-'N'} != 'y' ]; then exit 1; fi
	# override the output type to push to dockerhub
	VERSION=${VERSION} docker bake --builder smartnode-builder -f docker/daemon-bake.hcl smartnode --set "smartnode.output=type=registry"
	echo "Done!"

.PHONY: docker-latest
docker-latest: ${BUILD_DIR}/docker-buildx-builder
	echo -n "Building ${VERSION}, tagging as latest, and publishing. Continue? [yN]: " && read ans && if [ $${ans:-'N'} != 'y' ]; then exit 1; fi
	# override the output type to push to dockerhub, and the tags array to tag latest
	VERSION=${VERSION} docker bake --builder smartnode-builder -f docker/daemon-bake.hcl smartnode --set "smartnode.output=type=registry" --set "smartnode.tags=rocketpool/smartnode:latest"

.PHONY: docker-prune
docker-prune:
	docker system prune -af
	docker buildx prune -af
	docker buildx rm smartnode-builder
	rm ${BUILD_DIR}/docker-buildx-builder

.PHONY: lint
lint:
ifndef NO_DOCKER
	docker run -e GOMODCACHE=/go/.cache/pkg/mod -e GOCACHE=/go/.cache/go-build -e GOLANGCI_LINT_CACHE=/go/.cache/golangci-lint --user $(shell id -u):$(shell id -g) --rm -v ~/.cache:/go/.cache -v .:/smartnode --workdir /smartnode/ golangci/golangci-lint:v2.1-alpine golangci-lint fmt --diff
endif

.PHONY: test
test:
	go test -test.timeout 20m $$(go list ./... | grep -v bindings)

.PHONY: clean
clean:
	rm -rf ${BUILD_DIR}
	docker buildx rm smartnode-builder
