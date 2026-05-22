export PATH := $(PATH):`go env GOPATH`/bin
export GO111MODULE=on
LDFLAGS := -s -w
NOWEB_TAG = $(shell [ ! -d web/frps/dist ] || [ ! -d web/frpc/dist ] && echo ',noweb')
FRP_COMPAT_BASELINE_COUNT ?= 8
FRP_COMPAT_FLOOR_VERSION ?= 0.61.0

.PHONY: web frps-web frpc-web frps frpc e2e-compatibility-smoke e2e-compatibility e2e-compatibility-floor

all: env fmt web build

build: frps frpc

env:
	@go version

web: frps-web frpc-web

frps-web:
	$(MAKE) -C web/frps build

frpc-web:
	$(MAKE) -C web/frpc build

fmt:
	go fmt ./...

fmt-more:
	gofumpt -l -w .

gci:
	gci write -s standard -s default -s "prefix(github.com/fatedier/frp/)" ./

vet:
	go vet -tags "$(NOWEB_TAG)" ./...

frps:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags "frps$(NOWEB_TAG)" -o bin/frps ./cmd/frps

frpc:
	env CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -tags "frpc$(NOWEB_TAG)" -o bin/frpc ./cmd/frpc

test: gotest

gotest:
	go test -tags "$(NOWEB_TAG)" -v --cover ./assets/...
	go test -tags "$(NOWEB_TAG)" -v --cover ./cmd/...
	go test -tags "$(NOWEB_TAG)" -v --cover ./client/...
	go test -tags "$(NOWEB_TAG)" -v --cover ./server/...
	go test -tags "$(NOWEB_TAG)" -v --cover ./pkg/...

e2e:
	./hack/run-e2e.sh

e2e-trace:
	DEBUG=true LOG_LEVEL=trace ./hack/run-e2e.sh

e2e-compatibility-smoke: build
	FRP_COMPAT_BASELINE_COUNT=1 ./hack/run-e2e-compatibility.sh

e2e-compatibility: build
	FRP_COMPAT_BASELINE_COUNT="$(FRP_COMPAT_BASELINE_COUNT)" ./hack/run-e2e-compatibility.sh

e2e-compatibility-floor: build
	FRP_COMPAT_BASELINE_VERSIONS="$(FRP_COMPAT_FLOOR_VERSION)" ./hack/run-e2e-compatibility.sh

e2e-compatibility-last-frpc:
	if [ ! -d "./lastversion" ]; then \
		TARGET_DIRNAME=lastversion ./hack/download.sh; \
	fi
	FRPC_PATH="`pwd`/lastversion/frpc" ./hack/run-e2e.sh
	rm -r ./lastversion

e2e-compatibility-last-frps:
	if [ ! -d "./lastversion" ]; then \
		TARGET_DIRNAME=lastversion ./hack/download.sh; \
	fi
	FRPS_PATH="`pwd`/lastversion/frps" ./hack/run-e2e.sh
	rm -r ./lastversion

alltest: vet gotest e2e
	
clean:
	rm -f ./bin/frpc
	rm -f ./bin/frps
	rm -rf ./lastversion
	rm -rf ./.cache
	rm -rf ./.compat

# Docker targets
docker-build:
	docker build -t frpc-manager:latest -f dockerfiles/Dockerfile-for-frpc-manager .

docker-build-release:
	docker build -t frpc-manager:$(shell git describe --tags --always --dirty) -f dockerfiles/Dockerfile-for-frpc-manager .

# Packaging targets
package-linux: frpc
	@mkdir -p release/frpc_manager_linux_amd64
	@cp bin/frpc release/frpc_manager_linux_amd64/
	@cp package/systemd/frpc@.service release/frpc_manager_linux_amd64/
	@cp package/scripts/install-service.sh release/frpc_manager_linux_amd64/install.sh
	@cp package/scripts/frp-env.sh release/frpc_manager_linux_amd64/frp-env.sh
	@chmod +x release/frpc_manager_linux_amd64/frpc
	@chmod +x release/frpc_manager_linux_amd64/install.sh
	@tar -czf release/frpc_manager_linux_amd64.tar.gz -C release frpc_manager_linux_amd64
	@echo "Linux package created: release/frpc_manager_linux_amd64.tar.gz"

package-windows: frpc
	@mkdir -p release/frpc_manager_windows_amd64
	@cp bin/frpc release/frpc_manager_windows_amd64/frpc.exe
	@cp package/scripts/install-service.bat release/frpc_manager_windows_amd64/
	@cp package/scripts/uninstall-service.bat release/frpc_manager_windows_amd64/
	@echo "Windows package created: release/frpc_manager_windows_amd64/"

package-all: package-linux package-windows
	@echo "All packages created in release/"
