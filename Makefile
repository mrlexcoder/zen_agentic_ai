.PHONY: help build run docker-build docker-up docker-down test clean

help:
	@echo "AI Trading System - Make Commands"
	@echo ""
	@echo "Backend:"
	@echo "  make run-backend    - Run Go backend locally"
	@echo ""
	@echo "Frontend:"
	@echo "  make run-frontend   - Run frontend dev server"
	@echo "  make build-frontend - Build frontend for production"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build   - Build Docker images"
	@echo "  make docker-up      - Start all services"
	@echo "  make docker-down    - Stop all services"
	@echo ""
	@echo "All:"
	@echo "  make build          - Build everything"
	@echo "  make clean          - Clean build artifacts"

run-backend:
	cd backend && go run cmd/server/main.go

run-frontend:
	cd frontend && npm run dev

build-frontend:
	cd frontend && npm run build

docker-build:
	cd docker && docker-compose build

docker-up:
	cd docker && docker-compose up -d

docker-down:
	cd docker && docker-compose down

build:
	make build-frontend

clean:
	cd backend && go clean
	rm -rf frontend/dist