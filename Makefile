.PHONY: api frontend seed

api:
	cd api && go run ./cmd/api

frontend:
	cd frontend && npm run dev

seed:
	cd api && DATABASE_URL=postgres://radiooo:radiooo@localhost:5432/radiooo?sslmode=disable go run ./cmd/seed

build-api:
	cd api && go build ./...

build-frontend:
	cd frontend && npm run build
