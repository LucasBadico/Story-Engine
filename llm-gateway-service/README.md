# Ingestion Service

Service responsible for ingesting story content and generating embeddings for semantic search.

## Features

- Text ingestion from stories, chapters, scenes, beats, and prose blocks
- Vector storage using pgvector
- Semantic retrieval for memory search

## Architecture

The ingestion service:
- Connects to the main-service via gRPC to fetch story content
- Generates embeddings using embedding providers (OpenAI, Ollama, etc.)
- Stores embeddings in PostgreSQL with pgvector extension
- Provides search capabilities for semantic memory retrieval

## Setup

### Prerequisites

- PostgreSQL with pgvector extension
- Access to main-service gRPC endpoint
- Embedding provider API key (OpenAI, Ollama, etc.)

### Database Setup

```bash
# Run migrations
make migrate-up
```

### Configuration

Set environment variables:
- `DATABASE_URL`: PostgreSQL connection string
- `MAIN_SERVICE_GRPC_ADDR`: gRPC address of main-service
- `EMBEDDING_PROVIDER`: Provider name (openai, ollama, etc.)
- `EMBEDDING_API_KEY`: API key for embedding provider
- `HTTP_ADDR`: HTTP server bind address (default `:8081`)

## Usage

```bash
# Run worker
make run-worker
```

```bash
# Run HTTP API
go run ./cmd/api
```

## Search API

```
POST /api/v1/search
X-Tenant-ID: <tenant-uuid>
Content-Type: application/json

{
  "query": "string",
  "limit": 10,
  "cursor": "base64"
}
```

Response:
```
{
  "chunks": [
    {
      "chunk_id": "uuid",
      "document_id": "uuid",
      "source_type": "story",
      "source_id": "uuid",
      "content": "...",
      "score": 0.87,
      "beat_type": "...",
      "beat_intent": "...",
      "characters": ["..."],
      "location_name": "...",
      "timeline": "...",
      "pov_character": "...",
      "content_kind": "..."
    }
  ],
  "next_cursor": "base64"
}
```

## Development

```bash
# Run tests
make test

# Run integration tests
make test-integration
```
