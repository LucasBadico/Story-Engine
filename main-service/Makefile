.PHONY: db-up db-down migrate-up migrate-down migrate-create test test-integration clean

# Database management
db-up:
	docker-compose up -d postgres
	@echo "Waiting for database to be ready..."
	@sleep 3

db-down:
	docker-compose down

db-reset: db-down db-up migrate-up

# Migration management
MIGRATE_CMD = migrate -path migrations -database "postgres://storyengine:storyengine_dev@localhost:5432/storyengine?sslmode=disable"

migrate-up:
	$(MIGRATE_CMD) up

migrate-down:
	$(MIGRATE_CMD) down

migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Testing
test:
	go test -v ./...

test-integration:
	go test -v -tags=integration ./...

# Development
clean:
	go clean -cache

# Install dependencies
deps:
	go mod download
	go mod tidy

