.PHONY: start stop api grpc gateway worker mcp

# Source .env before running each service
define load-env
	set -a && . $(1)/.env && set +a
endef

# Start all services with interleaved logs via humanlog
start:
	@echo "Starting all services..."
	@(cd main-service && $(call load-env,.) && go run cmd/api-http/main.go 2>&1 | humanlog | sed 's/^/[api]     /' & \
	  cd main-service && $(call load-env,.) && go run cmd/api-grpc/main.go 2>&1 | humanlog | sed 's/^/[grpc]    /' & \
	  cd llm-gateway-service && $(call load-env,.) && go run cmd/http/main.go 2>&1 | humanlog | sed 's/^/[gateway] /' & \
	  cd llm-gateway-service && $(call load-env,.) && go run cmd/worker/main.go 2>&1 | humanlog | sed 's/^/[worker]  /' & \
	  wait)

# Individual services
api:
	cd main-service && $(call load-env,.) && go run cmd/api-http/main.go 2>&1 | humanlog

grpc:
	cd main-service && $(call load-env,.) && go run cmd/api-grpc/main.go 2>&1 | humanlog

gateway:
	cd llm-gateway-service && $(call load-env,.) && go run cmd/http/main.go 2>&1 | humanlog

worker:
	cd llm-gateway-service && $(call load-env,.) && go run cmd/worker/main.go 2>&1 | humanlog

# MCP server — SSE on port 8082 (for LM Studio on Windows)
mcp:
	cd mcp-server && go run cmd/mcp/main.go --sse=:8082 --default-story-id="d7acd365-1304-4ed7-918e-789fff0e82ef" --default-tenant-id="0d0a0534-7f05-4def-89bf-dd431f2134fc" 2>&1

# Stop all services (kills by port + worker process)
stop:
	@echo "Stopping services..."
	@fuser -k 7979/tcp 2>/dev/null && echo "  api stopped" || echo "  api not running"
	@fuser -k 50051/tcp 2>/dev/null && echo "  grpc stopped" || echo "  grpc not running"
	@fuser -k 8081/tcp 2>/dev/null && echo "  gateway stopped" || echo "  gateway not running"
	@pkill -f "story-block/llm-gateway-service.*cmd/worker" 2>/dev/null && echo "  worker stopped" || echo "  worker not running"
	@fuser -k 8082/tcp 2>/dev/null && echo "  mcp stopped" || echo "  mcp not running"
	@echo "Done"
