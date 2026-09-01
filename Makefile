.PHONY: infra-up infra-down backend frontend test build

infra-up:
	docker compose up -d postgres redis

infra-down:
	docker compose down

backend:
	cd backend && go run ./cmd/api

frontend:
	cd frontend && npm run dev

test:
	cd backend && go test ./...
	cd frontend && npm run build

build:
	docker compose build
