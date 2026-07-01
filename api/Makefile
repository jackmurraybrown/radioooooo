.PHONY: build test tidy lint run schema-docs

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

schema-docs:
	tbls doc --rm-dist
	mv .tbls-tmp/README.md internal/database/SCHEMA.md
	rm -rf .tbls-tmp
