.PHONY: db-up db-down db-reset-volume db-truncate migrate-up migrate-down migrate-create test test-integration test-http clean run-offline

# PostgreSQL defaults (used by pg-test-clean)
PGH ?= localhost
PGP ?= 5432
PGU ?= postgres
PGD ?= postgres
PGPASS ?= postgres

# Database management
db-up:
	@echo "Checking if PostgreSQL is running locally..."
	@pg_isready -h localhost -p 5432 -U postgres > /dev/null 2>&1 || \
		(echo "ERROR: PostgreSQL is not running. Start it with 'brew services start postgresql@16' or similar"; exit 1)
	@echo "PostgreSQL is ready!"
	@echo "Creating database if it doesn't exist..."
	@psql -h localhost -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'storyengine'" | grep -q 1 && \
		echo "Database 'storyengine' already exists" || \
		(psql -h localhost -U postgres -c "CREATE DATABASE storyengine" && echo "Database 'storyengine' created")

db-down:
	@echo "PostgreSQL runs as a local service. Use 'brew services stop postgresql' if needed."

db-reset-volume:
	@echo "Dropping and recreating database..."
	@psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS storyengine" 2>/dev/null || true
	@psql -h localhost -U postgres -c "CREATE DATABASE storyengine"
	@echo "Database reset complete"

db-reset: db-reset-volume migrate-up

db-truncate:
	@echo "Truncating all tables..."
	@psql -h localhost -U postgres -d storyengine -c "\
		TRUNCATE TABLE prose_blocks, beats, scenes, chapters, stories, \
		audit_logs, memberships, users, tenants \
		RESTART IDENTITY CASCADE;" && \
		echo "All tables truncated successfully"

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

# Run HTTP/SQLite server
run-offline:
	go run cmd/api-offline/main.go

# Test HTTP handlers
test-http: db-up
	@echo "Running HTTP handler tests with database cloning..."
	go test -v -tags=integration ./internal/transport/http/...


# Clean only databases that start with test_
pg-test-clean:
	@echo ">> Dropping databases starting with test_"
	@PGPASSWORD=$(PGPASS) cmds=$$(PGPASSWORD=$(PGPASS) psql -h $(PGH) -p $(PGP) -U $(PGU) -d $(PGD) -Atc "\
SELECT 'DROP DATABASE IF EXISTS \"' || datname || '\";' \
FROM pg_database \
WHERE datistemplate = false \
  AND datname ~ E'^test_';"); \
	echo "$$cmds" | pbcopy 2>/dev/null || true; \
	echo "$$cmds" | while read -r cmd; do \
		db=$$(echo $$cmd | sed -n 's/.*DROP DATABASE IF EXISTS \"\(.*\)\";.*/\1/p'); \
		[ -z "$$db" ] && continue; \
		echo " - Cleaning $$db"; \
		PGPASSWORD=$(PGPASS) psql -h $(PGH) -p $(PGP) -U $(PGU) -d $(PGD) -Atc "\
SELECT pg_terminate_backend(pid) FROM pg_stat_activity \
WHERE datname='$$db' AND pid <> pg_backend_pid();" >/dev/null || true; \
		PGPASSWORD=$(PGPASS) psql -h $(PGH) -p $(PGP) -U $(PGU) -d $(PGD) -c "$$cmd"; \
	done
	@echo ">> Done!"
