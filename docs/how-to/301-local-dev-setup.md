# Local Dev Setup

> Get the full story-engine stack running locally on WSL2: PostgreSQL, pgvector, Redis, both Go services, and the gateway worker.

---

## Prerequisites

- WSL2 with Ubuntu 24.04
- Go 1.24+ installed and in PATH
- PostgreSQL installed with your OS user as a superuser (see [novelist-app setup](../../novelist-app/docs/how-to/310-publix.go-environment-setup.md))
- Redis installed (`redis-server --version`)
- protoc + Go plugins installed (see [300-main.grpc-toolchain-setup.md](300-main.grpc-toolchain-setup.md))
- `migrate` CLI installed (`go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest`)

### PostgreSQL Auth Model

This project uses the `postgres` user with password `pass`. All `DATABASE_URL` connections use `postgres://postgres:pass@...`.

Set the password if you haven't already:

```bash
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'pass';"
```

Verify:

```bash
psql -h 127.0.0.1 -U postgres -c "SELECT 1;"
# Enter password: pass
```

---

## 1. Start PostgreSQL

```bash
sudo service postgresql start

# Check version
psql -h 127.0.0.1 -U postgres -c "SHOW server_version;"
```

---

## 2. Install pgvector Extension

Required by the LLM Gateway for embedding similarity search.

```bash
# Check your postgres version first
psql -h 127.0.0.1 -U postgres -c "SHOW server_version;"

# Install the matching package
sudo apt update
sudo apt install postgresql-16-pgvector
# or: sudo apt install postgresql-17-pgvector
```

If not in apt, build from source: https://github.com/pgvector/pgvector#linux-and-mac

### Verify pgvector is available

```bash
# Check the extension is recognized by PostgreSQL
psql -h 127.0.0.1 -U postgres -c "SELECT default_version FROM pg_available_extensions WHERE name = 'vector';"
# Expected: a version like 0.8.0 (any non-empty result means it's installed)

# Test it actually works
psql -h 127.0.0.1 -U postgres -c "CREATE EXTENSION IF NOT EXISTS vector; SELECT '[1,2,3]'::vector;"
# Expected: [1,2,3]
```

If `pg_available_extensions` returns no rows, the package didn't install for your PostgreSQL version — check the version match.

---

## 3. Create Databases

```bash
# Main service database
psql -h 127.0.0.1 -U postgres -c "CREATE DATABASE storyblocks;"

# LLM Gateway database
psql -h 127.0.0.1 -U postgres -c "CREATE DATABASE llm_gateway;"

# Enable pgvector on gateway DB
psql -h 127.0.0.1 -U postgres -d llm_gateway -c "CREATE EXTENSION IF NOT EXISTS vector;"

# Verify pgvector is enabled on gateway DB
psql -h 127.0.0.1 -U postgres -d llm_gateway -c "SELECT extname, extversion FROM pg_extension WHERE extname = 'vector';"
```

---

## 4. Run Migrations

```bash
# Main service (currently up to 070)
cd main-service
make migrate-up

# LLM Gateway (currently up to 010)
cd llm-gateway-service
migrate -path ./migrations -database "postgresql://$USER@127.0.0.1:5432/llm_gateway?sslmode=disable" up
```

---

## 5. Configure .env Files

**main-service/.env**

```
DATABASE_URL=postgres://postgres:pass@127.0.0.1:5432/storyblocks?sslmode=disable
GRPC_PORT=50051
HTTP_PORT=8080
LLM_GATEWAY_ENABLED=true
REDIS_ADDR=localhost:6379
LLM_GATEWAY_REDIS_ADDR=localhost:6379
LLM_GATEWAY_REDIS_DB=0
```

**llm-gateway-service/.env**

```
DATABASE_URL=postgres://postgres:pass@127.0.0.1:5432/llm_gateway?sslmode=disable
HTTP_ADDR=:8081
MAIN_SERVICE_GRPC_ADDR=localhost:50051
MAIN_SERVICE_HTTP_ADDR=http://localhost:8080
EMBEDDING_PROVIDER=ollama
EMBEDDING_MODEL=nomic-embed-text
EMBEDDING_BASE_URL=http://localhost:11434
REDIS_ADDR=localhost:6379
```

---

## 6. Start Redis

```bash
sudo service redis-server start

# Verify
redis-cli ping
# Expected: PONG
```

---

## 7. Run Services

Four terminals (or use `tmux`/background processes):

```bash
# Terminal 1: Main service HTTP API
cd main-service && make http-blocks-api

# Terminal 2: Main service gRPC server
cd main-service && make grpc-blocks-api

# Terminal 3: LLM Gateway HTTP API
cd llm-gateway-service && make http-blocks-llm

# Terminal 4: LLM Gateway Worker (processes Redis queue)
cd llm-gateway-service && make worker-blocks-llm
```

---

## 8. Verify

```bash
# Health checks
curl http://localhost:8080/health
curl http://localhost:8081/health

# gRPC (if grpcurl installed)
grpcurl -plaintext localhost:50051 list
```

---

## After WSL Restart

WSL2 doesn't auto-start services. Run these after every `wsl --shutdown` or reboot:

```bash
sudo service postgresql start
sudo service redis-server start
```

### DNS breaks (`no such host`, `dial tcp: lookup ...`)

```bash
echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf
```

**Do NOT** edit `/etc/wsl.conf` with `generateResolvConf = false` — this breaks SSH.

---

## Quick Reference

| What | Command |
|------|---------|
| Start PostgreSQL | `sudo service postgresql start` |
| Start Redis | `sudo service redis-server start` |
| Fix DNS | `echo "nameserver 8.8.8.8" \| sudo tee /etc/resolv.conf` |
| Run main-service migrations | `cd main-service && make migrate-up` |
| Run gateway migrations | `cd llm-gateway-service && migrate -path ./migrations -database "postgresql://$USER@..." up` |
| Main HTTP API | `cd main-service && make http-blocks-api` (port 8080) |
| Main gRPC server | `cd main-service && make grpc-blocks-api` (port 50051) |
| Gateway HTTP API | `cd llm-gateway-service && make http-blocks-llm` (port 8081) |
| Gateway Worker | `cd llm-gateway-service && make worker-blocks-llm` |
