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

## Usage

```bash
# Run worker
make run-worker
```

## Development

```bash
# Run tests
make test

# Run integration tests
make test-integration
```

