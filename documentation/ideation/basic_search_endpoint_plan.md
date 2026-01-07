Basic Semantic Search Endpoint Plan (Ideation)

Goal
- Expose a simple semantic search endpoint that queries embedding_chunks.
- Keep results ordered by similarity, with cursor-based pagination.
- Require tenant isolation via X-Tenant-ID header (JWT validation later).

API Shape (HTTP)
- POST /api/v1/search
- Headers: X-Tenant-ID: <uuid>
- Body:
  {
    "query": "string",
    "limit": 10,
    "cursor": "base64"
  }
- Response:
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

Score Definition
- Use cosine similarity.
- In pgvector: distance = (embedding <=> query_vector).
- score = 1 - distance (higher is better).

Cursor Model
- Cursor payload = { distance, chunk_id }.
- Query uses stable ordering: ORDER BY distance ASC, chunk_id ASC.
- Cursor filter:
  - (distance > last_distance) OR (distance = last_distance AND chunk_id > last_chunk_id)

Implementation Steps
1) Repository
- Extend ChunkRepository.SearchSimilar to accept cursor and return distance.
- Add SearchCursor + ScoredChunk to repositories layer.
- Update query to return distance and use stable ordering.

2) Use Case
- Extend SearchMemoryInput with cursor.
- Return NextCursor in SearchMemoryOutput.
- Convert distance to score in output.

3) HTTP Adapter
- Add minimal HTTP server cmd (new cmd/api).
- Implement search handler to validate X-Tenant-ID, parse body, decode cursor, call use case, encode cursor.
- Add /health endpoint.

4) Tests
- Update chunk repository tests for new signature and cursor ordering.
- Add handler unit tests for basic request/response + cursor.

5) Docs
- Update README with new API usage and env var for HTTP addr.
