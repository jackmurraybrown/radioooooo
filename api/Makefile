.PHONY: build test tidy lint run

build:
	go build ./...

test:
	gotestsum --format testname ./...

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

run:
	go run ./cmd/api
