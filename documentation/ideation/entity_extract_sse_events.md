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
  "phase": "extractor",
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
- `extractor`
- `matcher`
- `payload`

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

### result
Final payload (same shape as `/api/v1/entity-extract` response).

Data:
```json
{
  "payload": {
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
}
```

### error
Emitted if the pipeline fails.

Data:
- error

## Client Notes
- You can safely update UI incrementally using `extract.candidate` and `match.*`.
- Use `result` as the final authoritative payload.
- If the client cancels the request, the server stops emitting further events.
