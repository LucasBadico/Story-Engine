# Shared Go Module Plan (Code Reuse Across Services)

## Goal
Extract cross-cutting Go code into a standalone module outside service repos to avoid duplication between `main-service` and `llm-gateway-service`.

## What Can Be Shared Now
### httpkit/middleware
- `Middleware` type
- `Chain(...)` helper

### httpkit/cors
- `CORS(opts...)` middleware
- Default origin: `app://obsidian.md`
- Configurable origins, methods, headers, credentials

### httpkit/response
- `WriteJSON(w, status, payload)`
- `WriteError(w, status, msg)` (simple string error)

### Optional (later)
#### httpkit/tenant
- `ParseTenantID(r)` (returns `uuid.UUID` + error)
- No error-formatting logic to keep it generic

## What Should NOT Be Shared Yet
### logging
- `main-service` uses `slog`, `llm-gateway` uses stdlib `log`
- Shared interface only makes sense if both services agree on a single logger abstraction

### config
- Each service has its own config schema
- No overlap worth extracting now

## Proposed New Repository
- Repo: `github.com/story-engine/shared-go` (or equivalent)
- Standalone `go.mod`
- Minimal dependencies (stdlib only)

## Migration Plan
1) Create repo + module scaffolding.
2) Move `Middleware` + `Chain` into `httpkit/middleware`.
3) Move CORS into `httpkit/cors` with options struct.
4) Move JSON helpers into `httpkit/response`.
5) Update imports in both services.
6) Remove local duplicates.
7) Tag `v0.1.0` and pin versions in both services.

## Versioning Strategy
- Semantic versioning
- Start with `v0.1.0`
- Use tags for releases; services should pin versions

## Rollout Order
1) LLM gateway (simpler surface).
2) Main service (more complex error formatting).

## Follow-ups
- Decide if tenant parsing should be shared.
- Revisit a shared logger interface if services converge on one logging stack.
