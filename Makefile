.PHONY: dev migrate migrate-down seed test build lint clean help

# Variables
API_DIR := apps/api
WEB_DIR := apps/web
MIGRATIONS_DIR := packages/database/migrations
DB_URL ?= postgresql://financeos:financeos@localhost:5432/financeos

# Default target
help:
	@echo "FinanceOS - Available commands:"
	@echo ""
	@echo "  make dev          - Start all services with docker-compose"
	@echo "  make migrate      - Run database migrations"
	@echo "  make migrate-down - Rollback last migration"
	@echo "  make seed         - Seed initial data"
	@echo "  make test         - Run all tests (Go + Flutter)"
	@echo "  make build        - Build production artifacts"
	@echo "  make lint         - Run linters (go vet + dart analyze)"
	@echo "  make clean        - Remove build artifacts"
	@echo ""

# Start development environment
dev:
	docker-compose up -d
	@echo "Services started:"
	@echo "  API:      http://localhost:8000"
	@echo "  Web:      http://localhost:3000"
	@echo "  Adminer:  http://localhost:8080"
	@echo "  Postgres: localhost:5432"
	@echo "  Redis:    localhost:6379"

dev-down:
	docker-compose down

dev-logs:
	docker-compose logs -f

# Run database migrations
migrate:
	@which migrate > /dev/null 2>&1 || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "Migrations applied successfully"

# Rollback last migration
migrate-down:
	@which migrate > /dev/null 2>&1 || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1
	@echo "Migration rolled back"

# Seed initial data
seed:
	@which migrate > /dev/null 2>&1 || (echo "Installing golang-migrate..." && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up
	@echo "Seed data applied"

# Run all tests
test: test-api test-web

test-api:
	@echo "Running Go tests..."
	cd $(API_DIR) && go test ./... -v -race -coverprofile=coverage.out
	@echo "Go tests completed"

test-web:
	@echo "Running Flutter tests..."
	cd $(WEB_DIR) && flutter test
	@echo "Flutter tests completed"

# Build production artifacts
build: build-api build-web

build-api:
	@echo "Building Go API..."
	cd $(API_DIR) && CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/server ./cmd/server
	@echo "API built: $(API_DIR)/bin/server"

build-web:
	@echo "Building Flutter web..."
	cd $(WEB_DIR) && flutter build web --release
	@echo "Flutter web built: $(WEB_DIR)/build/web"

# Run linters
lint: lint-api lint-web

lint-api:
	@echo "Running Go linter..."
	cd $(API_DIR) && go vet ./...
	@echo "Go vet completed"

lint-web:
	@echo "Running Dart analyzer..."
	cd $(WEB_DIR) && dart analyze
	@echo "Dart analyze completed"

# Clean build artifacts
clean:
	rm -rf $(API_DIR)/bin
	rm -rf $(API_DIR)/tmp
	rm -rf $(WEB_DIR)/build
	@echo "Build artifacts cleaned"

# Install dependencies
deps: deps-api deps-web

deps-api:
	cd $(API_DIR) && go mod download && go mod tidy

deps-web:
	cd $(WEB_DIR) && flutter pub get

# Format code
fmt: fmt-api fmt-web

fmt-api:
	cd $(API_DIR) && gofmt -w .

fmt-web:
	cd $(WEB_DIR) && dart format .
