.PHONY: db-up db-down migrate-up migrate-down migrate-create test test-integration clean deps run-worker

# Database management
db-up:
	cd .. && docker-compose up -d postgres
	@echo "Waiting for PostgreSQL to be ready..."
	@timeout=30; \
	while [ $$timeout -gt 0 ]; do \
		if docker exec story-engine-postgres pg_isready -U postgres > /dev/null 2>&1; then \
			echo "PostgreSQL is ready!"; \
			break; \
		fi; \
		echo "Waiting... ($$timeout seconds remaining)"; \
		sleep 1; \
		timeout=$$((timeout-1)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "PostgreSQL failed to start"; \
		exit 1; \
	fi

db-down:
	cd .. && docker-compose down

# Migration management
MIGRATE_CMD = migrate -path ./migrations -database "postgres://postgres:postgres@localhost:5432/storyengine?sslmode=disable"

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

test-integration: db-up
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

# Development
clean:
	go clean -cache

deps:
	go mod download
	go mod tidy

# Run worker
run-worker:
	go run cmd/worker/main.go

