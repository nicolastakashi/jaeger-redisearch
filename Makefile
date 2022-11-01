
RELEASE_VERSION ?=$(shell cat VERSION)
RELEASE=1
REVISION ?= $(shell git rev-parse HEAD)
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)
BINARY_FOLDER=bin
BINARY_NAME=jaeger-redisearch
ARTIFACT_NAME=ntakashi/$(BINARY_NAME)
GOCMD=go
GOMAIN=./cmd/main.go
GOBUILD=$(GOCMD) build
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
ENVVARS=GOOS=linux CGO_ENABLED=0

LDFLAGS=-w -extldflags "-static" \
		-X github.com/prometheus/common/version.Version=$(RELEASE_VERSION) \
		-X github.com/prometheus/common/version.Revision=$(REVISION) \
		-X github.com/prometheus/common/version.Branch=$(BRANCH) \
		-X github.com/prometheus/common/version.BuildUser=$(shell whoami) \
		-X "github.com/prometheus/common/version.BuildDate=$(shell date -u)"

docker-build:
	@DOCKER_BUILDKIT=1 docker build -t ${ARTIFACT_NAME}:${RELEASE_VERSION} -f ./build/package/Dockerfile --progress=plain .

docker-push:
	@DOCKER_BUILDKIT=1 docker push $(ARTIFACT_NAME):${RELEASE_VERSION}

build:
	$(GOCMD) build -ldflags '$(LDFLAGS)' -o $(BINARY_FOLDER)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) -v $(GOMAIN)

.PHONY: build-linux-amd64
build-linux-amd64:
	GOOS=linux CGO_ENABLED=0 GOARCH=amd64 $(MAKE) build

.PHONY: build-linux-arm64
build-linux-arm64:
	GOOS=linux CGO_ENABLED=0 GOARCH=arm64 $(MAKE) build

.PHONY: build-darwin-amd64
build-darwin-amd64:
	GOOS=darwin CGO_ENABLED=0 GOARCH=amd64 $(MAKE) build

.PHONY: build-darwin-arm64
build-darwin-arm64:
	GOOS=darwin CGO_ENABLED=0 GOARCH=arm64 $(MAKE) build

.PHONY: build-all-platforms
build-all-platforms: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64

.PHONY: tar
tar:
	tar -czvf $(BINARY_NAME)-$(GOOS)-$(GOARCH).tar.gz  $(BINARY_FOLDER)/$(BINARY_NAME)-$(GOOS)-$(GOARCH)

.PHONY: tar-linux-amd64
tar-linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) tar

.PHONY: tar-linux-arm64
tar-linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) tar

.PHONY: tar-darwin-amd64
tar-darwin-amd64:
	GOOS=darwin GOARCH=amd64 $(MAKE) tar

.PHONY: tar-darwin-arm64
tar-darwin-arm64:
	GOOS=darwin GOARCH=arm64 $(MAKE) tar

.PHONY: tar-all-platforms
tar-all-platforms: tar-linux-amd64 tar-linux-arm64 tar-darwin-amd64 tar-darwin-arm64

deps:
	$(ENVVARS) $(GOCMD) mod download

fmt:
	$(ENVVARS) $(GOCMD) fmt -x ./...

vet:
	$(ENVVARS) $(GOCMD) vet ./...

tests:
	$(ENVVARS) $(GOCMD) test ./...

all: fmt vet tests deps build

.PHONY: build

run-hotrod:
	@docker-compose up hotrod

run-redis:
	@docker-compose up redis

run-jaeger:
	@docker-compose up jaeger

.PHONY: run
run: build-all-platforms
	@docker-compose up

clean:
	@docker-compose down -v

.PHONY: integration-test
integration-test: build-all-platforms
	STORAGE=grpc-plugin \
	PLUGIN_BINARY_PATH=$(PWD)/bin/jaeger-redisearch-linux-amd64 \
	PLUGIN_CONFIG_PATH=$(PWD)/configs/config.yaml \
	go test ./integration
