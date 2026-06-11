.PHONY: api frontend

api:
	cd api && go run ./cmd/api

frontend:
	cd frontend && npm run dev

build-api:
	cd api && go build ./...

build-frontend:
	cd frontend && npm run build
