VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
BUILDTAGS := $(shell uname)

ldflags = -X github.com/hdac-io/friday/version.Name=friday \
	  -X github.com/hdac-io/friday/version.ServerName=nodef \
	  -X github.com/hdac-io/friday/version.ClientName=clif \
	  -X github.com/hdac-io/friday/version.Version=$(VERSION) \
	  -X github.com/hdac-io/friday/version.Commit=$(COMMIT) \
	  -X "github.com/hdac-io/friday/version.BuildTags=$(BUILDTAGS)"

.PHONY: install test

all: install

install: go.sum
	bash ./scripts/install_casperlabs_ee.sh
	go install -mod=readonly -ldflags '$(ldflags)' ./cmd/nodef
	go install -mod=readonly -ldflags '$(ldflags)' ./cmd/clif

go.sum: go.mod
	@go mod verify
	@go mod tidy

test:
	bash ./scripts/tests_with_cover.sh
