.PHONY: install test

all: install

install: go.sum
	bash ./scripts/install_casperlabs_ee.sh
	go install -mod=readonly ./cmd/nodf
	go install -mod=readonly ./cmd/clif

go.sum: go.mod
	@go mod verify
	@go mod tidy

test:
	bash ./scripts/tests_with_cover.sh
