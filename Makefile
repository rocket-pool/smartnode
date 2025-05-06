BUILD_DIR=build
CLI_DIR=${BUILD_DIR}/cli
DAEMON_DIR=${BUILD_DIR}/daemon

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
all: lint rocketpool-cli rocketpool-daemon

.PHONY: rocketpool-cli
rocketpool-cli: $(CLI_TARGET_STRINGS)

.PHONY: rocketpool-daemon
rocketpool-daemon: ${DAEMON_DIR}
	docker build -t rocketpool/smartnode-builder:latest -f docker/smartnode-builder .
	docker run --env OWNER=$(shell id -u):$(shell id -g) --rm -v $(PWD):/src -v $(PWD)/${DAEMON_DIR}:/out -v /tmp/docker-go-build:/root/.cache/go-build rocketpool/smartnode-builder:latest /src/rocketpool/build.sh

${CLI_DIR}:
	mkdir -p ${CLI_DIR}
${DAEMON_DIR}:
	mkdir -p ${DAEMON_DIR}

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
