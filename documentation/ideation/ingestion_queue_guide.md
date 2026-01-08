# Ingestion Queue Guide (Main Service)

This guide explains how to enqueue new/updated entities so the LLM gateway worker ingests them.

## When to Enqueue
- Any entity that should be searchable or used in entity extraction should enqueue on **create** and **update**.
- Current supported `source_type` values include: `world`, `character`, `location`, `event`, `artifact`, `faction`, `lore`, plus story-related types.

## Use Case Pattern
1) Add a queue field to the use case:
```
ingestionQueue queue.IngestionQueue
```

2) Add a setter:
```
func (uc *CreateXUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
}
```

3) Enqueue after successful create/update:
```
uc.enqueueIngestion(ctx, input.TenantID, entity.ID)
```

4) Add the helper:
```
func (uc *CreateXUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, entityID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "entity_type", entityID); err != nil {
		uc.logger.Error("failed to enqueue entity ingestion", "error", err, "entity_id", entityID, "tenant_id", tenantID)
	}
}
```

## Wiring (API Entrypoints)
After constructing the use case, inject the queue:
```
createXUseCase.SetIngestionQueue(ingestionQueue)
updateXUseCase.SetIngestionQueue(ingestionQueue)
```

Apply this in:
- `main-service/cmd/api-http/main.go`
- `main-service/cmd/api-grpc/main.go`

## Notes
- The queue can be nil when LLM gateway notifications are disabled; the helper safely no-ops.
- Use the correct `source_type` string (must match LLM gateway worker support).
