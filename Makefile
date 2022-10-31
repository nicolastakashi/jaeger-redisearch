
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
	$(ENVVARS) $(GOCMD) build -ldflags '$(LDFLAGS)' -o $(BINARY_FOLDER)/$(BINARY_NAME) -v $(GOMAIN)

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
run: build
	@docker-compose up

clean:
	@docker-compose down -v

.PHONY: integration-test
integration-test: build
	STORAGE=grpc-plugin \
	PLUGIN_BINARY_PATH=$(PWD)/bin/jaeger-redisearch \
	PLUGIN_CONFIG_PATH=$(PWD)/configs/config.yaml \
	go test ./integration -v