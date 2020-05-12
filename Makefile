.PHONY: all build clean fmt test lint vendor

## overridable Makefile variables
# test to run
TESTSET = .
# benchmarks to run
BENCHSET ?= .

# version (defaults to short git hash)
VERSION ?= $(shell git rev-parse --short HEAD)

# use correct sed for platform
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    SED := gsed
else
    SED := sed
endif

PKG_NAME=github.com/Octops/gameserver-events-broadcaster

LDFLAGS := -X "${PKG_NAME}/internal/version.Version=${VERSION}"
LDFLAGS += -X "${PKG_NAME}/internal/version.BuildTS=$(shell date -u '+%Y-%m-%d %I:%M:%S')"
LDFLAGS += -X "${PKG_NAME}/internal/version.GitCommit=$(shell git rev-parse HEAD)"
LDFLAGS += -X "${PKG_NAME}/internal/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"

GO       := GO111MODULE=on GOPRIVATE=github.com/Octops GOSUMDB=off go
GOBUILD  := CGO_ENABLED=0 $(GO) build $(BUILD_FLAG)
GOTEST   := $(GO) test -gcflags='-l' -p 3

CURRENT_DIR := $(shell pwd)
FILES    := $(shell find internal cmd -name '*.go' -type f -not -name '*.pb.go' -not -name '*_generated.go' -not -name '*_test.go')
TESTS    := $(shell find internal cmd -name '*.go' -type f -not -name '*.pb.go' -not -name '*_generated.go' -name '*_test.go')

BROADCASTER_BIN := bin/broadcaster

DOCKER_IMAGE_TAG ?= octops/gameserver-events-broadcaster:v0.1-alpha

default: clean build

build: clean $(BROADCASTER_BIN)

$(BROADCASTER_BIN):
	CGO_ENABLED=0 GOOS=linux go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(BROADCASTER_BIN) .

dist:
	CGO_ENABLED=0 GOOS=linux go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(BROADCASTER_BIN) .
	CGO_ENABLED=0 GOOS=darwin go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(BROADCASTER_BIN)-darwin .
	#CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(BROADCASTER_BIN).exe $(PKG_NAME)/cmd/controller/
	#CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=6 go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(BROADCASTER_BIN)-armhf $(PKG_NAME)/cmd/controller/
	#CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags '$(LDFLAGS)' -a -installsuffix cgo -o $(BROADCASTER_BIN)-arm64 $(PKG_NAME)/cmd/controller/

clean:
	rm -f $(BROADCASTER_BIN)*

get:
	$(GO) get ./...
	$(GO) mod verify
	$(GO) mod tidy

update:
	$(GO) get -u -v all
	$(GO) mod verify
	$(GO) mod tidy

fmt:
	gofmt -s -l -w $(FILES) $(TESTS)

lint:
	golangci-lint run

test:
	$(GOTEST) -run=$(TESTSET) ./...
	@echo
	@echo Configured tests ran ok.

test-strict:
	$(GO) test -p 3 -run=$(TESTSET) -gcflags='-l -m' -race ./...
	@echo
	@echo Configured tests ran ok.

bench:
	DEBUG=0 $(GOTEST) -run=nothing -bench=$(BENCHSET) -benchmem ./...
	@echo
	@echo Configured benchmarks ran ok.

vendor:
	$(GO) mod vendor

docker:
	docker build -t $(DOCKER_IMAGE_TAG) .

push: docker
	docker push $(DOCKER_IMAGE_TAG)