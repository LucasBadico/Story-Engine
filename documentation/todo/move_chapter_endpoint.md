# Add "Move Chapter" Endpoint (main-service)

This document outlines how to add a "move chapter" endpoint to main-service. There is no existing move-chapter endpoint in either HTTP or gRPC today; only MoveScene exists.

## Goal

Add an endpoint that changes a chapter's position/order within a story. The most common interpretation is:
- Move chapter relative to other chapters in the same story (reorder by `number` or explicit ordering).
- Keep story ownership the same (not moving between stories).

If you intend to move chapters between stories, call that out explicitly and adjust validation rules accordingly.

## Current State (no move endpoint)

- gRPC: `main-service/proto/chapter/chapter.proto` does not define MoveChapter.
- HTTP: `main-service/internal/transport/http/handlers/chapter_handler.go` has Create/Get/Update/List/Delete only.
- Use cases: `main-service/internal/application/story/chapter/` has create/get/update/delete/list only.

## Design Choices (pick one)

1) Reorder by absolute position:
   - Request contains `target_number` (or `target_position`).
   - Server shifts other chapters in the same story accordingly.

2) Reorder by "insert before/after":
   - Request contains `target_chapter_id` and a `position` enum (BEFORE/AFTER).

3) Reorder by "swap":
   - Request contains `target_chapter_id` and swaps `number`.

The current schema uses an integer `number` in `chapter` (see `main-service/internal/core/story/chapter.go`), so option (1) is the most consistent.

## Implementation Steps

### 1) Domain layer (story)

If needed, add a domain method to adjust chapter number with validation:
- `main-service/internal/core/story/chapter.go`
  - Example: `func (c *Chapter) UpdateNumber(number int) { ... }`
  - Ensure validation rules still pass (likely already enforced in `Validate()`).

### 2) Repository layer

Add a repository method to reorder chapters in a story:
- `main-service/internal/ports/repositories/chapter.go`
  - Add a method like:
    - `MoveInStory(ctx, tenantID, chapterID uuid.UUID, targetNumber int) error`
    - or `Reorder(ctx, tenantID, storyID uuid.UUID, chapterID uuid.UUID, targetNumber int) error`

Then implement it in:
- `main-service/internal/adapters/db/postgres/chapter_repository.go`
- `main-service/internal/adapters/db/sqlite/chapter_repository.go`

Implementation idea:
- Fetch chapter to get its `story_id` and current number.
- In a transaction, shift siblings in the range and update the target chapter number.
- Use existing ordering rules (see `ListByStoryOrdered`) to keep consistent.

### 3) Application layer (use case)

Add a new use case:
- `main-service/internal/application/story/chapter/move_chapter.go`

Suggested API:
```
type MoveChapterInput struct {
  TenantID uuid.UUID
  ChapterID uuid.UUID
  TargetNumber int
}
type MoveChapterOutput struct {
  Chapter *story.Chapter
}
```

Steps in Execute:
- Get chapter by ID (tenant + chapter).
- Validate `TargetNumber` range (>=1 and <= total chapters for story).
- Call repository reorder method.
- Return updated chapter (re-fetch or update in memory).

### 4) gRPC transport

Update proto:
- `main-service/proto/chapter/chapter.proto`
  - Add `MoveChapterRequest` / `MoveChapterResponse`.
  - Add `rpc MoveChapter(MoveChapterRequest) returns (MoveChapterResponse);`

Regenerate protobufs:
- `main-service/proto/chapter/chapter.pb.go`
- `main-service/proto/chapter/chapter_grpc.pb.go`

Add handler:
- `main-service/internal/transport/grpc/handlers/chapter_handler.go`
  - Add a new method similar to `MoveScene`.
  - Parse tenant ID, `id`, and `target_number`.
  - Call use case and return response.

Wire use case in:
- `main-service/cmd/api-grpc/main.go`

### 5) HTTP transport

Add endpoint:
- `main-service/internal/transport/http/handlers/chapter_handler.go`
  - Add `Move` handler, for example:
    - `PATCH /api/v1/chapters/{id}/move`
    - Body: `{ "target_number": 3 }`

Wire route:
- `main-service/cmd/api-http/main.go`

### 6) Tests

Add unit tests:
- Use case: `main-service/internal/application/story/chapter/move_chapter_test.go`
- gRPC handler: `main-service/internal/transport/grpc/handlers/chapter_handler_test.go`
- HTTP handler: `main-service/internal/transport/http/handlers/chapter_handler_test.go`

Test cases:
- Move chapter to a higher position (shift down).
- Move chapter to a lower position (shift up).
- Invalid target number.
- Chapter not found.
- Cross-tenant safety.

### 7) Documentation (optional but recommended)

Add to API docs:
- `documentation/Story_Engine_API.postman_collection.json`
- `documentation/guides/REST_API_Quick_Reference.md`

## Suggested gRPC/HTTP Shapes

gRPC:
```
message MoveChapterRequest {
  string id = 1;
  int32 target_number = 2;
}
message MoveChapterResponse {
  Chapter chapter = 1;
}
```

HTTP:
```
PATCH /api/v1/chapters/{id}/move
{
  "target_number": 3
}
```

## Notes

- If `number` is not strictly 1..N today, you may want to normalize ordering before move.
- Use a transaction to avoid conflicting updates when shifting ranges.
- If moving between stories is required, add `target_story_id` and update validations to enforce tenant ownership.
