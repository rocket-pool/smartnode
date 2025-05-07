BUILD_DIR=build
CLI_DIR=${BUILD_DIR}/cli
DAEMON_DIR=${BUILD_DIR}/daemon
TREEGEN_DIR=${BUILD_DIR}/treegen

CLI_TARGET_OOS:=linux darwin
ARCHS:=arm64 amd64

DOCKER_RUN_CMD=docker run --env OWNER=$(shell id -u):$(shell id -g) --rm -v /tmp/docker-go-build:/root/.cache/go-build 
DOCKER_BUILDER_TAG=rocketpool/smartnode-builder:latest

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
all: lint rocketpool-cli rocketpool-daemon

.PHONY: docker-builder
docker-builder: docker/smartnode-builder
	docker build -t $(DOCKER_BUILDER_TAG) -f docker/smartnode-builder .

.PHONY: rocketpool-cli
rocketpool-cli: $(CLI_TARGET_STRINGS)

.PHONY: rocketpool-daemon
rocketpool-daemon: ${DAEMON_DIR} docker-builder
	$(DOCKER_RUN_CMD) -v $(PWD):/src -v $(PWD)/${DAEMON_DIR}:/out ${DOCKER_BUILDER_TAG} /src/rocketpool/build.sh
.PHONY: treegen-bin
treegen-bin: ${TREEGEN_DIR}/treegen-linux-amd64 ${TREEGEN_DIR}/treegen-linux-arm64
.PHONY: treegen-docker
treegen-container: treegen-bin
	./treegen/build-release.sh -a -v $(TREEGEN_VERSION)
.PHONY: treegen-container-push
treegen-container-push: treegen-container
	./treegen/build-release.sh -p -a -v $(TREEGEN_VERSION)

${CLI_DIR}:
	mkdir -p ${CLI_DIR}
${DAEMON_DIR}:
	mkdir -p ${DAEMON_DIR}
${TREEGEN_DIR}:
	mkdir -p ${TREEGEN_DIR}
${TREEGEN_DIR}/treegen-linux-amd64 ${TREEGEN_DIR}/treegen-linux-arm64: $(shell find treegen -name "*.go") treegen/go.mod docker-builder
	$(DOCKER_RUN_CMD) -v $(PWD):/src -v $(PWD)/${TREEGEN_DIR}:/out ${DOCKER_BUILDER_TAG} /src/treegen/build_binaries.sh

$(foreach oos,$(CLI_TARGET_OOS),$(foreach arch,$(ARCHS),$(eval $(call rocketpool-cli-template,$(oos),$(arch)))))

.PHONY: clean
clean:
	rm -rf build

.PHONY: lint
lint:
	@echo $(MODULE_GLOBS)
	golangci-lint run --disable-all --enable goimports $(MODULE_GLOBS) 

.PHONY: test
test:
	go test -test.timeout 20m $(MODULE_GLOBS)
