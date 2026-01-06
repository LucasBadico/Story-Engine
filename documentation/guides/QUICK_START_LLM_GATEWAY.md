# LLM Gateway Quick Start

This guide shows the minimal steps to run the LLM Gateway ingestion worker and verify that embeddings are stored.

## Prerequisites

- Go 1.21+
- PostgreSQL 16+ with pgvector extension
- Redis 7+
- main-service running with gRPC enabled
- Embedding provider access (Ollama or OpenAI)

## 1) Start infrastructure

From the repository root:

```bash
docker-compose up -d
```

## 2) Configure environment

Set environment variables (defaults shown):

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/storyengine?sslmode=disable"
export REDIS_ADDR="localhost:6379"
export REDIS_PASSWORD=""
export REDIS_DB="0"
export MAIN_SERVICE_GRPC_ADDR="localhost:50051"
export EMBEDDING_PROVIDER="ollama"
export EMBEDDING_BASE_URL="http://localhost:11434"
export EMBEDDING_API_KEY=""
export EMBEDDING_MODEL="nomic-embed-text"
```

If you use OpenAI, set `EMBEDDING_PROVIDER="openai"` and provide `EMBEDDING_API_KEY`.

## 3) Run migrations

```bash
cd llm-gateway-service
make migrate-up
```

If your `storyengine` database is already used by `main-service`, you may hit migration version conflicts. The simplest fix is to use a separate database:

```bash
createdb storyengine_llm
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/storyengine_llm?sslmode=disable"
make migrate-up
```

## 4) Run the ingestion worker

```bash
make run-worker
```

You should see logs like:

```
INFO Starting ingestion service worker...
INFO Database connected
INFO Redis connected
```

## 5) Enqueue an ingestion item

The worker consumes Redis sorted sets with this format:

- Key: `ingestion:queue:{tenant_id}`
- Member: `{source_type}:{source_id}`
- Score: Unix timestamp

Example:

```bash
redis-cli ZADD ingestion:queue:<tenant-id> $(date +%s) "story:<story-id>"
```

Valid source types:

```
story, chapter, scene, beat, content_block, world, character, location, event, artifact, faction, lore
```

## 6) Verify results

Check that documents and chunks were created:

```bash
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM embedding_documents;"
psql "$DATABASE_URL" -c "SELECT COUNT(*) FROM embedding_chunks;"
```

If counts increase after enqueuing, ingestion worked.

## Troubleshooting

- `pgvector extension not found`: run `CREATE EXTENSION vector;` on the database.
- `Connection refused to main-service`: confirm `MAIN_SERVICE_GRPC_ADDR` and that gRPC is running.
- `Redis connection failed`: confirm `REDIS_ADDR`.
- `Embedding dimension mismatch`: ensure the vector dimension matches the embedding model.
