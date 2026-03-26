# Task 2 — Local Setup Rebuild

## Goal
Get the full story-engine running locally on the new machine. Document every step so it becomes the definitive setup guide (migrates to `docs/how-to/`).

## Prerequisites
- WSL2 with Ubuntu 24.04
- Go 1.24+ (installed)
- PostgreSQL running (installed)
- Redis running (installed)
- protoc + Go plugins (installed)

## Current State
- PostgreSQL auth configured: `postgres` user, password `pass`, scram-sha-256
- pgvector 0.6.0 installed (postgresql-16-pgvector)
- Databases created: `storyblocks` (main), `storyblocks_llm` (gateway)
- Migrations run: main 071, gateway 011
- `.env` files updated for new machine
- oh-my-zsh `dotenv` plugin enabled for auto-loading `.env` files
- Setup doc created: `docs/how-to/301-local-dev-setup.md`
- DB name renamed from `storyengine` → `storyblocks` across all config
- Makefile run targets renamed: `http-blocks-api`, `grpc-blocks-api`, `http-blocks-llm`, `worker-blocks-llm`, `http-blocks-offline`

## Scope

- [x] PostgreSQL auth configured
- [x] pgvector installed
- [x] Databases created (storyblocks, storyblocks_llm)
- [x] Migrations run (main: 071, gateway: 011)
- [x] .env files configured
- [x] Setup doc created: `docs/how-to/301-local-dev-setup.md`
- [x] Main service HTTP starts (port 7979, default from config)
- [x] Main service gRPC starts (port 50051, changed from 9090 — Cockpit SSH tunnel)
- [x] Gateway HTTP starts (port 8081)
- [x] Gateway worker starts
- [ ] Health checks pass
- [x] End-to-end test: import CLI → main service API → Redis queue → gateway worker creates documents + embeddings
- [x] Gateway embeddings working (Ollama at `100.80.252.100:11434` via Tailscale, nomic-embed-text)
- [x] Canon retrieval endpoint returns assembled data (18 canonical docs, 7 states, 4 events)
- [x] MCP server tested end-to-end (get_canon returns full story canon via SSE transport)
- [ ] Test Obsidian plugin connection to running services

## Cleanup Done (post-setup)

- [x] DB renamed: `storyengine` → `storyblocks`, `storyengine_llm` → `storyblocks_llm` (all Makefiles, .env, config defaults, test helpers, docker-compose)
- [x] DB password: all connection strings updated from `postgres:postgres` → `postgres:pass`
- [x] Makefile run targets renamed for DX: `run-http` → `http-blocks-api`, `run-grpc` → `grpc-blocks-api`, `run-http` → `http-blocks-llm`, `run-worker` → `worker-blocks-llm`, `run-offline` → `http-blocks-offline`
- [x] Migration 071 (main): Refactored `plot_developments` — dropped `plot_development_entities` table, added `notes` (text) + `timeline_event_id` (FK → timeline_events). Entity context now comes through linked timeline events.
- [x] Migration 011 (gateway): Dropped orphaned `reveals` table (replaced by plots system in task-1 refactor #21)
- [x] Removed all gateway reveals dead code: domain model, repository interface, postgres implementation, test file, canon retrieval (RevealEntry, revealRepo, loadReveals), cmd/http wiring, test helper truncate
- [x] Removed reveals from MCP server: client DTO, canon tool, writing_context tool
- [x] Proto recompiled for plot changes (removed entities field, added notes + timeline_event_id)

## Dev Log

| # | Date | Event |
|---|------|-------|
| 1 | 2026-03-16 | Task created. New machine — tools installed (go, postgres, redis, protoc, air, migrate, gotestfmt). Need: postgres auth, pgvector, databases, migrations, .env config. |
| 2 | 2026-03-16 | Session ended — user got pulled into other work. MCP server (task 3) built in parallel: 5 tools, 2 resources, gateway HTTP client, builds clean. Next session: full focus on local setup + testing ingestion pipeline + MCP. |
| 3 | 2026-03-18 | PostgreSQL auth fixed: `postgres` user password set to `pass` via `sudo -u postgres psql`. pgvector 0.6.0 installed (postgresql-16-pgvector). Databases created: `storyblocks`, `storyblocks_llm` with vector extension. All migrations run. |
| 4 | 2026-03-18 | DB renamed `storyengine` → `storyblocks` across all config (Makefiles, .env, Go config defaults, test helpers, docker-compose). Password updated `postgres` → `pass` in all connection strings. |
| 5 | 2026-03-18 | Makefile run targets renamed for DX: `http-blocks-api`, `grpc-blocks-api`, `http-blocks-llm`, `worker-blocks-llm`, `http-blocks-offline`. oh-my-zsh `dotenv` plugin enabled for auto .env loading. |
| 6 | 2026-03-18 | Setup doc created: `docs/how-to/301-local-dev-setup.md` with full setup steps, pgvector verification, WSL2 known issues. |
| 7 | 2026-03-18 | Plot refactor: migration 071 drops `plot_development_entities`, adds `notes` + `timeline_event_id` FK to `plot_developments`. Updated domain model (removed `PlotDevelopmentEntity`), repo, use case, HTTP/gRPC handlers, proto, gateway client, gateway ingester. All 3 services build + vet clean. |
| 8 | 2026-03-18 | Reveals cleanup: gateway migration 011 drops `reveals` table. Removed all dead code across gateway (4 files deleted, canon retrieval cleaned) and MCP server (client DTO, 2 tools cleaned). All 3 services build + vet clean. |
| 9 | 2026-03-18 | Plot ingestion rewritten: gateway `IngestPlotUseCase` now creates versioned plot documents (warm canon pattern — close old version, LLM-synthesize new version via Gemini) + individual development documents. New prompt `synthesize_plot_state.prompt` generates rich narrative summaries. New `SourceTypePlotDevelopment`. Executor wired in worker. Mechanical fallback when LLM unavailable. |
| 10 | 2026-03-18 | MCP plot tool: new `get_plots` tool filters canon by source_type=plot/plot_development, groups developments under parent plot. MCP now has 6 tools. All 3 services build + vet clean. |
| 11 | 2026-03-22 | **Import CLI refactored to use HTTP API** instead of direct DB/use case calls. Old CLI bypassed the API, so entities were never published to Redis ingestion queue. New CLI uses `MainServiceAPI` interface backed by `HTTPAPIClient` — calls REST endpoints which properly trigger `SetIngestionQueue` → Redis publish. CLI now needs only `--api-url` (default `http://localhost:7979`), no DB connection. 4 files changed: `api.go` (new interface), `http_api_client.go` (new HTTP impl), `import_chapter.go` (refactored from 15 deps to 1 interface), `cmd/import-chapter/main.go` (simplified, no DB wiring). |
| 12 | 2026-03-22 | Fresh import test via HTTP API: 76 entities created (5 chars, 2 locs, 7 artifacts, 1 chapter, 7 scenes, 20 beats, 20 content blocks, 4 char states, 3 rel states, 1 timeline event, 3 plots + 3 devs). **Redis queue now has 75 items** (vs 2 before refactor). Gateway worker created 22 embedding_documents (7 source types). |
| 13 | 2026-03-22 | **Embedding chunks = 0**: gateway worker creates documents but can't generate embeddings. `EMBEDDING_PROVIDER=ollama` configured but Ollama runs on Windows, not WSL2. Updated `.env` to use Tailscale hostname `win-3090:11434`. Ollama on Windows needs `OLLAMA_HOST=0.0.0.0` to listen on all interfaces. Doc created: `docs/how-to/302-ollama-wsl2-setup.md`. |
| 14 | 2026-03-22 | Discovered main service HTTP default port is 7979 (not 8080). Gateway HTTP on 8081. Main service gRPC default is 50051 (not tested yet). |
| 15 | 2026-03-25 | **Ollama confirmed working** via Tailscale IP `100.80.252.100:11434`. Hostname `win-3090` resolves to `127.0.1.1` (local loopback) because WSL2 hostname is also `win-3090` (inherited from Windows). Tailscale name is `wsl-3090` but `/etc/hostname` says `win-3090`. Gateway `.env` already uses Tailscale IP directly — no issue. |
| 16 | 2026-03-25 | **Redis DB changed 2 → 1**. DB 0 used by other projects (Cockpit integration via SSH tunnel on port 9090). Both `.env` files updated. |
| 17 | 2026-03-25 | **gRPC port changed 9090 → 50051**. Port 9090 occupied by SSH tunnel for Cockpit. Updated `GRPC_PORT=50051` in main-service `.env`, `MAIN_SERVICE_GRPC_ADDR=localhost:50051` in gateway `.env`. |
| 18 | 2026-03-25 | **Root Makefile created** with `make start` (all 4 services: api, grpc, gateway, worker with humanlog + prefixed logs), individual targets (`make api`, `make grpc`, `make gateway`, `make worker`), and `make stop` (kills by port via `fuser`). Each target sources `.env` before running. |
| 19 | 2026-03-25 | **Re-import test**: truncated both DBs, created tenant/world/story, ran importer — 37 entities created (5 chars, 2 locs, 7 artifacts, 1 chapter, 7 scenes, 1 content block, 4 char states, 3 rel states, 1 timeline event, 3 plots+devs). 36 items queued in Redis DB 1. Worker was not processing — discovered gRPC not running (port 9090 conflict). |

## Decisions & Rationale

| # | Decision | Chosen | Why |
|---|----------|--------|-----|
| 1 | DB name `storyblocks` | Renamed from `storyengine` | User preference — fresh start on new machine |
| 2 | Password `pass` | Simple password for local dev | Easy to type, local-only PostgreSQL |
| 3 | Makefile `*-blocks-*` naming | `http-blocks-api`, `grpc-blocks-api`, etc. | Better DX — clear which service you're running at a glance |
| 4 | Plot developments → timeline events | Drop `plot_development_entities`, add `timeline_event_id` FK | Timeline events already track who/where/when — no need to duplicate entity links on developments |
| 5 | Drop gateway `reveals` table | Migration 011 + dead code removal | Main service replaced reveals with plots in task-1 refactor #21. Nothing writes to reveals anymore. |
| 6 | LLM-synthesized plot documents | Gemini generates narrative summary per plot version | Richer content = better embeddings = better semantic retrieval. Mechanical fallback preserves functionality without LLM. |
| 7 | Versioned plot documents (warm canon) | Close old doc (`valid_to_chapter=N`), create new version (`valid_from_chapter=N`) | Same pattern as warm canon — `as_of_chapter` filtering works out of the box via `ListCanonical`. No new tables needed. |
| 8 | Separate `plot_development` source type | Development docs distinct from plot summary docs | Plot doc = synthesized state up to chapter N. Development doc = detailed record of a single moment. Different retrieval purposes. |
| 9 | Import CLI uses HTTP API, not direct DB | `MainServiceAPI` interface + `HTTPAPIClient` | API path is where `SetIngestionQueue` is wired. Direct use case calls skipped Redis publish for most entity types. |
| 10 | Ollama via Tailscale IP `100.80.252.100` | Direct IP instead of hostname `win-3090` | WSL2 hostname is also `win-3090` (inherited from Windows), so hostname resolves to `127.0.1.1` instead of Tailscale. |
| 11 | Embeddings stay on Ollama (not Gemini) | `nomic-embed-text` via Ollama on Windows GPU | No Gemini embedding provider implemented in gateway. Ollama is free + local + fast with GPU. |
| 12 | Redis DB 1 (not 0 or 2) | `REDIS_DB=1` / `LLM_GATEWAY_REDIS_DB=1` | DB 0 used by other projects (Cockpit). DB 2 was arbitrary — consolidated to 1. |
| 13 | gRPC port 50051 (not 9090) | `GRPC_PORT=50051` | Port 9090 occupied by SSH tunnel for Cockpit integration. 50051 is gRPC convention. |

## Difficulties & Solutions

| # | Problem | Solution | Time |
|---|---------|----------|------|
| 1 | Lost postgres password from previous machine | Used `sudo -u postgres psql` (peer auth over Unix socket) to reset password | 5 min |
| 2 | Gateway HTTP crashed: "Missing Gemini API key" | `.env` wasn't being loaded — added oh-my-zsh `dotenv` plugin to auto-source `.env` on `cd` | 5 min |
| 3 | Import CLI didn't publish entities to Redis | CLI called use cases directly, bypassing HTTP API where `SetIngestionQueue` is wired. Only story + world reached Redis (2 of 75). | Refactored CLI to HTTP API client — all 75 entities now queued. |
| 4 | Ollama unreachable from WSL2 | Ollama runs on Windows (`localhost:11434`), WSL2 can't reach Windows localhost. | Use Tailscale hostname `win-3090:11434`. Also need `OLLAMA_HOST=0.0.0.0` on Windows side. |
