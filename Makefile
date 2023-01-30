GO ?= go
GOLANGCI_LINT ?= $$($(GO) env GOPATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.50.0
GOGOPROTOBUF ?= protoc-gen-gogofaster
GOGOPROTOBUF_VERSION ?= v1.3.1

COMMIT ?= "$(shell git describe --long --dirty --always --match "" || true)"
VERSION ?= "$(shell git describe --tags --abbrev=0 || true)"
LDFLAGS ?= -s -w -X github.com/fairdatasociety/fairOS-dfs.commit="$(COMMIT)" -X github.com/fairdatasociety/fairOS-dfs.version="$(VERSION)"
DEST ?= "$(shell (go list ./... | grep -v wasm))"

.PHONY: all
all: build lint vet test-race binary

.PHONY: binary
binary: export CGO_ENABLED=1
binary: dist FORCE
	$(GO) version
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/dfs ./cmd/dfs
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/dfs-cli ./cmd/dfs-cli

dist:
	mkdir $@

.PHONY: lint
lint: linter
	$(GOLANGCI_LINT) run --skip-dirs wasm

.PHONY: linter
linter:
	test -f $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

.PHONY: swagger
swagger:
	which swag || ( echo "install swag for your system from https://github.com/swaggo/swag" && exit 1)
	swag init -g ./cmd/server.go -d cmd/dfs,pkg/api,cmd/common,pkg/dir,pkg/file,pkg/pod,pkg/user,pkg/collection -o ./swagger

.PHONY: vet
vet:
	$(GO) vet "$(DEST)"

.PHONY: test-race
test-race:
	$(GO) test -race -timeout 300000ms -v "$(DEST)"

.PHONY: test
test:
	$(GO) test -v "$(DEST)"

.PHONY: build
build:
	$(GO) build  "$(DEST)"

.PHONY: githooks
githooks:
	ln -f -s ../../.githooks/pre-push.bash .git/hooks/pre-push

.PHONY: protobuftools
protobuftools:
	which protoc || ( echo "install protoc for your system from https://github.com/protocolbuffers/protobuf/releases" && exit 1)
	which $(GOGOPROTOBUF) || ( cd /tmp && GO111MODULE=on $(GO) get -u github.com/gogo/protobuf/$(GOGOPROTOBUF)@$(GOGOPROTOBUF_VERSION) )

.PHONY: protobuf
protobuf: GOFLAGS=-mod=mod # use modules for protobuf file include option
protobuf: protobuftools
	$(GO) generate -run protoc "$(DEST)"

.PHONY: clean
clean:
	$(GO) clean
	rm -rf dist/

.PHONY: release
release:
	docker run --rm --privileged \
		--env-file .release-env \
		-v ~/go/pkg/mod:/go/pkg/mod \
		-v `pwd`:/go/src/github.com/fairDataSociety/fairOS-dfs \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/fairDataSociety/fairOS-dfs \
		ghcr.io/goreleaser/goreleaser-cross:v1.19.5 release --rm-dist

.PHONY: release-dry-run
release-dry-run:
	docker run --rm --privileged \
		-v ~/go/pkg/mod:/go/pkg/mod \
		-v ~/go/bin:/go/bin \
		-v `pwd`:/go/src/github.com/fairDataSociety/fairOS-dfs \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-w /go/src/github.com/fairDataSociety/fairOS-dfs \
		ghcr.io/goreleaser/goreleaser-cross:v1.19.5 release --rm-dist \
		--skip-validate=true \
		--skip-publish


BUILD_DIR := $(PWD)
BIN_PATH := $(BUILD_DIR)/dist
EXEC_NAME := fairos.wasm
GZIP_EXEC_NAME := $(EXEC_NAME).gz

.PHONY: wasm
wasm:
	@GOOS=js GOARCH=wasm $(GO) build -ldflags="-s -w" -o $(BIN_PATH)/$(EXEC_NAME)
	@gzip -9 -v -c $(BIN_PATH)/$(EXEC_NAME) > $(BIN_PATH)/$(GZIP_EXEC_NAME)

.PHONY: android
android:
	$(GO) get golang.org/x/mobile/bind
	gomobile init
	gomobile bind -androidapi 21 -o fairos.aar -target=android -ldflags "$(LDFLAGS)" github.com/fairdatasociety/fairOS-dfs/gomobile
	$(GO) mod tidy

FORCE:
