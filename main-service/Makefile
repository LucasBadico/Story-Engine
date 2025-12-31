.PHONY: db-up db-down db-reset-volume db-truncate migrate-up migrate-down migrate-create test test-integration test-http clean

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
	@echo "Creating database if it doesn't exist..."
	@docker exec story-engine-postgres psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'storyengine'" | grep -q 1 && echo "Database 'storyengine' found in catalog" || (docker exec story-engine-postgres psql -U postgres -c "CREATE DATABASE storyengine" && echo "Database 'storyengine' created successfully")
	@echo "Verifying database is accessible..."
	@timeout=10; \
	while [ $$timeout -gt 0 ]; do \
		if docker exec story-engine-postgres psql -U postgres -d storyengine -c "SELECT 1" > /dev/null 2>&1; then \
			echo "Database 'storyengine' is accessible!"; \
			break; \
		fi; \
		echo "Waiting for database to be accessible... ($$timeout seconds remaining)"; \
		sleep 1; \
		timeout=$$((timeout-1)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "ERROR: Database 'storyengine' is not accessible"; \
		exit 1; \
	fi

db-down:
	cd .. && docker-compose down

db-reset-volume:
	cd .. && docker-compose down -v
	docker volume rm story-engine_postgres_data 2>/dev/null || true
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
	@echo "Creating database if it doesn't exist..."
	@docker exec story-engine-postgres psql -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'storyengine'" | grep -q 1 && echo "Database 'storyengine' found in catalog" || (docker exec story-engine-postgres psql -U postgres -c "CREATE DATABASE storyengine" && echo "Database 'storyengine' created successfully")
	@echo "Verifying database is accessible..."
	@timeout=10; \
	while [ $$timeout -gt 0 ]; do \
		if docker exec story-engine-postgres psql -U postgres -d storyengine -c "SELECT 1" > /dev/null 2>&1; then \
			echo "Database 'storyengine' is accessible!"; \
			break; \
		fi; \
		echo "Waiting for database to be accessible... ($$timeout seconds remaining)"; \
		sleep 1; \
		timeout=$$((timeout-1)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "ERROR: Database 'storyengine' is not accessible"; \
		exit 1; \
	fi

db-reset: db-reset-volume migrate-up

db-truncate:
	@echo "Truncating all tables..."
	@docker exec story-engine-postgres psql -U postgres -d storyengine -c "\
		TRUNCATE TABLE prose_blocks, beats, scenes, chapters, stories, \
		audit_logs, memberships, users, tenants \
		RESTART IDENTITY CASCADE;" && \
		echo "All tables truncated successfully" || \
		(echo "ERROR: Failed to truncate tables"; exit 1)

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
	go test -v ./... 2>&1 | tee /tmp/gotest.log
	
test-fmt:
	go test -json -v ./... 2>&1 | tee /tmp/gotest.log | gotestfmt


test-integration: db-up
	@echo "Running integration tests with database cloning..."
	go test -v -tags=integration ./... 2>&1 | tee /tmp/gotest-integration.log


test-integration-fmt: db-up
	@echo "Running integration tests with database cloning and formatting..."
	go test -json -v -tags=integration ./... 2>&1 | tee /tmp/gotest-integration.log | gotestfmt
# Development
clean:
	go clean -cache

# Install dependencies
deps:
	go mod download
	go mod tidy

# Proto generation
proto-gen:
	@export PATH="$(PATH):$(shell go env GOPATH)/bin"; \
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/**/*.proto

proto-clean:
	find proto -name "*.pb.go" -delete

# Run gRPC server
run-grpc:
	go run cmd/api-grpc/main.go

# Run HTTP server
run-http:
	go run cmd/api-http/main.go

# Test HTTP handlers
test-http: db-up
	@echo "Running HTTP handler tests with database cloning..."
	go test -v -tags=integration ./internal/transport/http/...

