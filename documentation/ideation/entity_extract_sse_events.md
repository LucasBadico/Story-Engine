# Entity Extract SSE Events

Endpoint:
- `POST /api/v1/entity-extract/stream`
- Headers: `Content-Type: application/json`, `X-Tenant-ID: <uuid>`
- Body: `{ "text": "...", "world_id": "<uuid>", "context": "", ... }`

Stream format:
- Content-Type: `text/event-stream`
- Each message:
  - `event: <type>`
  - `data: <json>`

Event JSON shape:
```json
{
  "type": "phase.start",
  "phase": "entities.candidates",
  "message": "extractor started",
  "data": { "chunks": 4 },
  "timestamp": "2026-01-08T18:00:00Z"
}
```

## Event Types

### request.received
Emitted when the server accepts the request.

Data:
- tenant_id
- world_id
- text_len

### pipeline.start
Emitted before any extraction work begins.

### router.chunk
Emitted after the router identifies entity types for a chunk.

Data:
- paragraph_id
- chunk_id
- types

### phase.start
Emitted when a phase starts.

Phase values:
- `entities.detect`
- `entities.candidates`
- `entities.resolve`
- `entities.result`
- `relations.discover`
- `relations.normalize`
- `relations.match`
- `relations.result`

### extract.candidate
Emitted for each extracted entity candidate.

Data:
- entity_type
- name
- evidence
- summary
- paragraph
- chunk

### match.found
Emitted when a match is confirmed.

Data:
- entity_type
- name
- source_type
- source_id
- similarity

### match.none
Emitted when no match is confirmed.

Data:
- entity_type
- name
- candidates

### phase.done
Emitted when a phase finishes.

Data (per phase):
- extractor: findings
- matcher: results

### result_entities
Emitted when entity extraction completes.

Data:
```json
{
  "entities": [
    {
      "type": "character",
      "name": "Aria",
      "summary": "...",
      "found": true,
      "match": { "source_type": "character", "source_id": "...", "similarity": 0.92 },
      "candidates": [ ... ]
    }
  ]
}
```

### result_relations
Emitted when relation extraction completes (if enabled).

Data:
```json
{
  "relations": [
    {
      "relation_type": "member_of",
      "source": { "type": "character", "name": "Aria" },
      "target": { "type": "faction", "name": "Order of the Sun" }
    }
  ]
}
```

### error
Emitted if the pipeline fails.

Data:
- error

## Client Notes
- You can safely update UI incrementally using `extract.candidate` and `match.*`.
- Use `result_entities`/`result_relations` as the authoritative payloads.
- If the client cancels the request, the server stops emitting further events.
