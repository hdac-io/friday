.PHONY: install test

all: install

install: go.sum
	go install -mod=readonly ./cmd/fryd
	go install -mod=readonly ./cmd/friday

go.sum: go.mod
	@go mod verify
	@go mod tidy

test:
	bash ./scripts/tests_with_cover.sh
