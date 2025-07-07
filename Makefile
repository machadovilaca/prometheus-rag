.PHONY: lint test

all: lint test

lint:
	golangci-lint run ./...

test:
	go test -coverprofile=coverage.out ./...
