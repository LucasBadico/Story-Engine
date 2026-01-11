# LLM Gateway Service

The **LLM Gateway Service** is a specialized microservice responsible for ingesting story content from the Story Engine API (main-service) and generating vector embeddings for semantic search and AI-powered memory retrieval. This service enables LLMs to access contextually relevant story information through semantic similarity searches.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Core Capabilities](#core-capabilities)
4. [Supported Entity Types](#supported-entity-types)
5. [Connection with Main Service](#connection-with-main-service)
6. [Embedding Providers](#embedding-providers)
7. [Database Schema](#database-schema)
8. [Worker System](#worker-system)
9. [Semantic Search](#semantic-search)
10. [Configuration](#configuration)
11. [Usage with LLMs](#usage-with-llms)
12. [Development Setup](#development-setup)

---

## Overview

The LLM Gateway Service acts as a bridge between the Story Engine's structured narrative data and Large Language Models. It:

- **Ingests** story content (characters, events, locations, lore, etc.) from the main-service via gRPC
- **Generates embeddings** using configurable providers (OpenAI, Ollama)
- **Stores vectors** in PostgreSQL with pgvector extension
- **Provides semantic search** for memory retrieval with rich metadata filtering

### Key Benefits

- **Semantic Memory**: LLMs can retrieve contextually relevant story information based on meaning, not just keywords
- **Rich Metadata**: Chunks include structural metadata (beat types, characters, locations, timelines) for precise filtering
- **Multi-tenant Support**: Full tenant isolation for data security
- **Debounced Processing**: Efficient batch processing with configurable debounce intervals

---

## Architecture

The service follows **hexagonal architecture** (ports and adapters pattern):

```
┌─────────────────────────────────────────────────────────────────┐
│                      LLM Gateway Service                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    Application Layer                      │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │   │
│  │  │   Ingest     │  │   Search     │  │   Worker     │   │   │
│  │  │  Use Cases   │  │  Use Cases   │  │   System     │   │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                      Ports (Interfaces)                   │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌───────┐ │   │
│  │  │  Embedder  │ │  gRPC      │ │ Repositories│ │ Queue │ │   │
│  │  │  Port      │ │  Client    │ │             │ │       │ │   │
│  │  └────────────┘ └────────────┘ └────────────┘ └───────┘ │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                       Adapters                            │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌───────┐ │   │
│  │  │  OpenAI    │ │  Ollama    │ │  PostgreSQL│ │ Redis │ │   │
│  │  └────────────┘ └────────────┘ └────────────┘ └───────┘ │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
           │                              │
           ▼                              ▼
    ┌─────────────┐              ┌──────────────────┐
    │ Main-Service│◄────gRPC────►│  PostgreSQL      │
    │   (gRPC)    │              │  (pgvector)      │
    └─────────────┘              └──────────────────┘
```

### Core Components

| Component | Description |
|-----------|-------------|
| **Ingest Use Cases** | Process entities and generate embeddings |
| **Search Use Cases** | Perform semantic similarity searches |
| **Worker** | Debounced queue processor for batch ingestion |
| **Embedder** | Interface for embedding generation |
| **gRPC Client** | Communication with main-service |
| **Repositories** | Data persistence layer |

---

## Core Capabilities

### 1. Content Ingestion

The service can ingest the following entity types from the main-service:

- Stories, Chapters, Content Blocks (prose text)
- Worlds, Characters, Locations
- Events, Artifacts, Factions, Lore

Each entity is processed to:
1. Fetch full entity data via gRPC
2. Fetch related entities (traits, references, relationships)
3. Build enriched content with context
4. Generate vector embeddings
5. Store documents and chunks with metadata

### 2. Embedding Generation

Supports multiple embedding providers:

- **OpenAI** (`text-embedding-ada-002`): 1536 dimensions
- **Ollama** (local models): 
  - `nomic-embed-text`: 768 dimensions
  - `mxbai-embed-large`: 1024 dimensions
  - `all-minilm`: 384 dimensions

### 3. Semantic Search

Vector similarity search using pgvector with:

- Cosine similarity scoring
- Multi-dimensional filtering (source type, beat types, characters, locations, etc.)
- Rich metadata in results for context

### 4. Debounced Processing

Worker system that:
- Polls Redis queue for pending items
- Waits for stability period (debounce) before processing
- Processes items in configurable batch sizes
- Handles multiple tenants concurrently

---

## Supported Entity Types

| Entity Type | Source Type Code | Metadata Captured |
|-------------|------------------|-------------------|
| **Story** | `story` | Title, Status |
| **Chapter** | `chapter` | Number, Title, Status |
| **Content Block** | `content_block` | Type, Kind, Beat context, Characters, Location, Timeline, POV |
| **World** | `world` | Name, Description, Genre |
| **Character** | `character` | Name, Description, Traits, World context |
| **Location** | `location` | Name, Type, Description, Hierarchy |
| **Event** | `event` | Name, Type, Description, Timeline, Importance, Related entities |
| **Artifact** | `artifact` | Name, Description, Rarity, World context |
| **Faction** | `faction` | Name, Type, Description, Beliefs, Structure |
| **Lore** | `lore` | Name, Category, Description, Rules, Limitations |

---

## Connection with Main Service

### gRPC Communication

The LLM Gateway connects to the main-service via gRPC to fetch story content. The connection is established using the `MAIN_SERVICE_GRPC_ADDR` environment variable.

```go
// Port Interface
type MainServiceClient interface {
    GetStory(ctx context.Context, storyID uuid.UUID) (*Story, error)
    GetChapter(ctx context.Context, chapterID uuid.UUID) (*Chapter, error)
    GetScene(ctx context.Context, sceneID uuid.UUID) (*Scene, error)
    GetBeat(ctx context.Context, beatID uuid.UUID) (*Beat, error)
    GetContentBlock(ctx context.Context, contentBlockID uuid.UUID) (*ContentBlock, error)
    GetWorld(ctx context.Context, worldID uuid.UUID) (*World, error)
    GetCharacter(ctx context.Context, characterID uuid.UUID) (*Character, error)
    GetLocation(ctx context.Context, locationID uuid.UUID) (*Location, error)
    GetEvent(ctx context.Context, eventID uuid.UUID) (*Event, error)
    GetArtifact(ctx context.Context, artifactID uuid.UUID) (*Artifact, error)
    GetFaction(ctx context.Context, factionID uuid.UUID) (*Faction, error)
    GetLore(ctx context.Context, loreID uuid.UUID) (*Lore, error)
    // ... relationship queries
}
```

### Data Flow

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Redis Queue   │────►│  LLM Gateway    │────►│  Main-Service   │
│   (Push Item)   │     │   Worker        │◄────│    (gRPC)       │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Embedder       │
                        │  (OpenAI/Ollama)│
                        └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  PostgreSQL     │
                        │  (pgvector)     │
                        └─────────────────┘
```

### Triggering Ingestion

Items are added to the ingestion queue via Redis. The queue uses sorted sets with timestamps for debouncing:

```go
// Queue Interface
type IngestionQueue interface {
    // Push adds/updates item with current timestamp (debounce reset)
    Push(ctx context.Context, tenantID uuid.UUID, sourceType string, sourceID uuid.UUID) error
    
    // PopStable returns items not updated since stableAt
    PopStable(ctx context.Context, tenantID uuid.UUID, stableAt time.Time, limit int) ([]*QueueItem, error)
    
    // ListTenantsWithItems returns tenant IDs with pending items
    ListTenantsWithItems(ctx context.Context) ([]uuid.UUID, error)
}
```

**Redis Key Format**: `ingestion:queue:{tenant_id}`

**Member Format**: `{source_type}:{source_id}`

---

## Embedding Providers

### OpenAI Adapter

```go
// Configuration
EMBEDDING_PROVIDER=openai
EMBEDDING_API_KEY=sk-...
EMBEDDING_MODEL=text-embedding-ada-002

// Dimensions: 1536
```

### Ollama Adapter (Local/Self-Hosted)

```go
// Configuration
EMBEDDING_PROVIDER=ollama
EMBEDDING_BASE_URL=http://localhost:11434
EMBEDDING_MODEL=nomic-embed-text

// Dimensions by model:
// - nomic-embed-text: 768
// - mxbai-embed-large: 1024
// - all-minilm: 384
```

### Embedder Interface

```go
type Embedder interface {
    // EmbedText generates an embedding for the given text
    EmbedText(text string) ([]float32, error)
    
    // EmbedBatch generates embeddings for multiple texts
    EmbedBatch(texts []string) ([][]float32, error)
    
    // Dimension returns the dimension of embeddings produced
    Dimension() int
}
```

---

## Database Schema

### Documents Table

Stores metadata about ingested entities:

```sql
CREATE TABLE embedding_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    source_type VARCHAR(50) NOT NULL,  -- 'story', 'chapter', 'character', etc.
    source_id UUID NOT NULL,
    title VARCHAR(255),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(tenant_id, source_type, source_id)
);
```

### Chunks Table

Stores embeddings with rich metadata:

```sql
CREATE TABLE embedding_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES embedding_documents(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(1536),  -- Adjustable for different models
    token_count INT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Structural metadata
    scene_id UUID,
    beat_id UUID,
    beat_type VARCHAR(50),
    beat_intent TEXT,
    
    -- Narrative context
    characters JSONB,          -- Array of character names
    location_id UUID,
    location_name VARCHAR(255),
    timeline VARCHAR(255),
    pov_character VARCHAR(255),
    
    -- Content block metadata
    content_type VARCHAR(50),  -- text, image, video
    content_kind VARCHAR(50),  -- final, alt_a, draft
    
    -- World entity metadata
    world_id UUID,
    world_name VARCHAR(255),
    world_genre VARCHAR(100),
    entity_name VARCHAR(255),
    event_timeline VARCHAR(255),
    importance INT,
    
    -- Related entities
    related_characters JSONB,
    related_locations JSONB,
    related_artifacts JSONB,
    related_events JSONB,
    
    UNIQUE(document_id, chunk_index)
);

-- Vector similarity search index
CREATE INDEX idx_embedding_chunks_embedding 
ON embedding_chunks USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);
```

---

## Worker System

### Debounced Worker

The worker polls the Redis queue and processes items that have been stable (not updated) for a configurable period:

```go
// Worker configuration
WORKER_DEBOUNCE_MINUTES=5   // Wait 5 minutes after last update
WORKER_POLL_SECONDS=60       // Poll queue every 60 seconds
WORKER_BATCH_SIZE=10         // Process 10 items per batch
```

### Processing Flow

1. **Poll**: Worker checks all tenant queues every `WORKER_POLL_SECONDS`
2. **Filter**: Items with timestamp older than `WORKER_DEBOUNCE_MINUTES` are selected
3. **Pop**: Selected items are atomically removed from queue
4. **Process**: Each item is ingested based on its source type
5. **Store**: Documents and chunks are saved to PostgreSQL

### Supported Ingest Operations

```go
// Ingest use cases
switch sourceType {
case "story":
    ingestStory.Execute(ctx, input)
case "chapter":
    ingestChapter.Execute(ctx, input)
case "content_block":
    ingestContentBlock.Execute(ctx, input)
case "world":
    ingestWorld.Execute(ctx, input)
case "character":
    ingestCharacter.Execute(ctx, input)
case "location":
    ingestLocation.Execute(ctx, input)
case "event":
    ingestEvent.Execute(ctx, input)
case "artifact":
    ingestArtifact.Execute(ctx, input)
case "faction":
    ingestFaction.Execute(ctx, input)
case "lore":
    ingestLore.Execute(ctx, input)
}
```

---

## Semantic Search

### Search Interface

```go
type SearchMemoryInput struct {
    TenantID    uuid.UUID
    Query       string
    Limit       int
    SourceTypes []memory.SourceType  // Optional filter
    
    // Structural filters
    BeatTypes   []string
    Characters  []string
    SceneIDs    []uuid.UUID
    StoryID     *uuid.UUID
}

type SearchResult struct {
    ChunkID     uuid.UUID
    DocumentID  uuid.UUID
    SourceType  memory.SourceType
    SourceID    uuid.UUID
    Content     string
    Similarity  float64  // Cosine similarity score (0-1)
    
    // Metadata
    BeatType     *string
    BeatIntent   *string
    Characters   []string
    LocationName *string
    Timeline     *string
    POVCharacter *string
}
```

### How Search Works

1. **Query Embedding**: The query text is converted to a vector using the same embedder
2. **Similarity Search**: pgvector finds chunks with highest cosine similarity
3. **Filtering**: Optional filters narrow results by metadata
4. **Enrichment**: Results include full metadata for context

### Search Query Example (pgvector)

```sql
SELECT c.*, d.source_type, d.source_id
FROM embedding_chunks c
INNER JOIN embedding_documents d ON c.document_id = d.id
WHERE d.tenant_id = $1
  AND d.source_type IN ('content_block', 'character')
  AND c.beat_type = 'dialogue'
ORDER BY c.embedding <=> $2::vector  -- Cosine distance
LIMIT 10;
```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5432/storyengine?sslmode=disable` |
| `REDIS_ADDR` | Redis server address | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password | `` |
| `REDIS_DB` | Redis database number | `0` |
| `MAIN_SERVICE_GRPC_ADDR` | Main-service gRPC address | `localhost:50051` |
| `EMBEDDING_PROVIDER` | Embedding provider (`openai`, `ollama`) | `ollama` |
| `EMBEDDING_BASE_URL` | Ollama base URL | `http://localhost:11434` |
| `EMBEDDING_API_KEY` | OpenAI API key | `` |
| `EMBEDDING_MODEL` | Embedding model name | `nomic-embed-text` |
| `WORKER_DEBOUNCE_MINUTES` | Debounce interval | `5` |
| `WORKER_POLL_SECONDS` | Queue poll interval | `60` |
| `WORKER_BATCH_SIZE` | Items per batch | `10` |

---

## Usage with LLMs

### RAG (Retrieval-Augmented Generation)

The LLM Gateway enables RAG workflows for story-aware AI:

```
┌──────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   User       │────►│  LLM Application│────►│  LLM Gateway    │
│   Query      │     │                 │◄────│  (Search API)   │
└──────────────┘     └────────┬────────┘     └─────────────────┘
                              │
                              ▼
                     ┌─────────────────┐
                     │   LLM (GPT-4,   │
                     │   Claude, etc.) │
                     └─────────────────┘
```

### Example RAG Flow

1. **User asks**: "What happened to Elena in the forest?"
2. **Search**: Query LLM Gateway for relevant chunks
   ```
   SearchMemoryInput{
       Query: "Elena forest",
       SourceTypes: ["content_block", "event"],
       Characters: ["Elena"],
       Limit: 5
   }
   ```
3. **Results**: Service returns relevant prose chunks with metadata
4. **Augment**: Results are injected into LLM prompt as context
5. **Generate**: LLM produces answer grounded in story data

### Memory Use Cases

| Use Case | How to Use |
|----------|------------|
| **Character Memory** | Search with `SourceTypes: ["character"]` and character name filter |
| **Event Timeline** | Search with `SourceTypes: ["event"]` and timeline filter |
| **Scene Context** | Search with specific `SceneIDs` filter |
| **Beat-Specific** | Filter by `BeatTypes: ["dialogue", "action"]` |
| **Location-Based** | Filter by `LocationIDs` or location name in query |
| **World Lore** | Search `SourceTypes: ["lore", "world"]` |

### Prompt Engineering with Context

```python
# Example prompt template
context_chunks = llm_gateway.search(query, filters)

prompt = f"""You are a story assistant with access to the following narrative context:

{format_chunks(context_chunks)}

Based on this context, answer the following question:
{user_question}

If the answer isn't in the context, say so clearly."""
```

---

## Development Setup

### Prerequisites

- Go 1.21+
- PostgreSQL 16+ with pgvector extension
- Redis 7+
- Main-service running with gRPC enabled
- Ollama (for local embeddings) or OpenAI API key

### Quick Start

```bash
# Start infrastructure
cd /path/to/story-engine
docker-compose up -d

# Navigate to llm-gateway-service
cd llm-gateway-service

# Install dependencies
make deps

# Run migrations
make migrate-up

# Start worker
make run-worker
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires DB)
make test-integration

# With formatted output
make test-fmt
```

### Database Management

```bash
# Start database
make db-up

# Reset database (drops all data)
make db-reset

# Truncate tables (keeps schema)
make db-truncate

# Create new migration
make migrate-create
```

---

## Future Enhancements

### Planned Features

1. **HTTP/gRPC API**: Expose search functionality via API endpoints
2. **Webhook Integration**: Receive events from main-service on entity updates
3. **Chunk Splitting**: Improved chunking strategies for long content
4. **Hybrid Search**: Combine vector search with keyword/BM25 search
5. **Batch Embedding**: Optimize embedding generation for large batches
6. **Cache Layer**: Redis caching for frequently accessed chunks

### Integration Roadmap

1. **Phase 1** (Current): Worker-based ingestion, basic search
2. **Phase 2**: API endpoints for search, webhook triggers
3. **Phase 3**: Real-time streaming, advanced filtering
4. **Phase 4**: Multi-model support, fine-tuned embeddings

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| `pgvector extension not found` | Run `CREATE EXTENSION vector;` in PostgreSQL |
| `Connection refused to main-service` | Ensure main-service gRPC is running on configured port |
| `Embedding dimension mismatch` | Recreate chunks table with correct vector dimension |
| `Redis connection failed` | Check `REDIS_ADDR` and ensure Redis is running |
| `Empty search results` | Verify documents exist and embeddings were generated |

### Logs

The worker logs all operations:

```
INFO Starting ingestion service worker...
INFO Database connected
INFO Redis connected
INFO gRPC client connected addr=localhost:50051
INFO Using Ollama embedder model=nomic-embed-text
INFO Starting worker debounce_minutes=5 poll_seconds=60
INFO Processing stable items tenant_id=xxx count=3
INFO Successfully ingested character character_id=xxx
```

---

## References

- [pgvector Documentation](https://github.com/pgvector/pgvector)
- [OpenAI Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)
- [Ollama Documentation](https://ollama.ai/)
- [Story Engine Main Service](../main-service/README.md)






