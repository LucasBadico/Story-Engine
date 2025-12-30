# Story Engine — Project Phases Plan

This document defines the **phases, features, and deliverables** of the Story Engine project.
Each phase is designed to be **incremental, testable, and shippable**, allowing us to mark progress clearly and maintain momentum.

The goal is to reach a **functional MVP** with:
- Multi-tenant story management
- Versioned stories (clone-based)
- gRPC API
- Obsidian plugin integration
- Basic ingestion + embeddings

---

## Phase 0 — Project Foundation (Setup)

**Goal:** Prepare the repository and development environment.

### Deliverables
- [ ] Repository initialized
- [ ] Project folder structure created
- [ ] README.md (project overview)
- [ ] Makefile
- [ ] docker-compose (Postgres)

### Models
_None (infrastructure only)_

### Status
- ⬜ Not started

---

## Phase 1 — Core Domain + Database (Vertical Slice)

**Goal:** Establish the core domain, persistence layer, and integration-tested use cases.

### Features
- Multi-tenancy (tenant, user, membership)
- Story entity
- Clone-based story versioning

### Models
- `tenant`
- `user`
- `membership`
- `story`
- `chapter`
- `scene`
- `beat`
- `prose_block`

### Technical Scope
- PostgreSQL schema (migrations)
- sqlc-generated queries
- Repository implementations (Postgres)
- Application use cases
- Integration tests (real DB, no mocks)

### Key Use Cases
- [ ] CreateTenant
- [ ] CreateStory
- [ ] CloneStoryTx (transactional)

### Status
- ⬜ Not started

---

## Phase 2 — gRPC API Layer

**Goal:** Expose the core functionality via gRPC for programmatic access.

### Features
- gRPC server
- Auth context (tenant + user)
- Story management endpoints

### Models (Proto)
- `Tenant`
- `Story`
- `Chapter`
- `Scene`

### Endpoints
- [ ] CreateStory
- [ ] GetStory
- [ ] ListStories
- [ ] CloneStory
- [ ] ListStoryVersions

### Technical Scope
- Protobuf definitions
- gRPC handlers calling application layer
- Basic auth/interceptors

### Status
- ⬜ Not started

---

## Phase 3 — Obsidian Plugin (Client MVP)

**Goal:** Enable authors to interact with Story Engine from Obsidian.

### Features
- Workspace configuration
- gRPC (or Connect/gRPC-Web) client
- Story listing and selection
- Trigger story operations

### Plugin Capabilities
- [ ] Configure API endpoint & token
- [ ] List stories
- [ ] Create new story
- [ ] Clone story
- [ ] Insert story/chapter into note

### Models
- Plugin-level DTOs (mirroring gRPC models)

### Status
- ⬜ Not started

---

## Phase 4 — Ingestion & Embedding (Basic Memory)

**Goal:** Add long-term memory via embeddings with minimal viable functionality.

### Features
- Text ingestion
- Vector storage
- Semantic retrieval

### Models
- `embedding_document`
- `embedding_chunk`

### Technical Scope
- pgvector setup
- Embedding adapter
- Simple ingestion pipeline

### Use Cases
- [ ] IngestStoryText
- [ ] IngestChapterSummary
- [ ] SearchRelevantMemory

### Status
- ⬜ Not started

---

## Phase 5 — Narrative Generation MVP

**Goal:** Generate chapters using structured context and memory.

### Features
- LLM profile per workspace
- Basic chapter generation pipeline

### Models
- `llm_provider`
- `workspace_llm_profile`
- `generation_run`

### Pipeline Steps
- [ ] Load story state
- [ ] Retrieve embeddings
- [ ] Generate chapter prose
- [ ] Persist output

### Status
- ⬜ Not started

---

## Phase 6 — MVP Finalization

**Goal:** Deliver a complete, usable MVP.

### Features
- End-to-end flow (Obsidian → gRPC → DB → LLM → DB)
- Minimal error handling & logging
- Documentation for usage

### Checklist
- [ ] Story CRUD works
- [ ] Versioning works
- [ ] Obsidian plugin usable
- [ ] Embeddings working
- [ ] One-click chapter generation

### Status
- ⬜ Not started

---

## MVP Definition (Done Criteria)

The MVP is considered **DONE** when:

- A user can create a workspace
- Create a story
- Clone the story into versions
- Trigger generation from Obsidian
- Persist generated chapters
- Retrieve past context via embeddings

---

## Notes

- Each phase is intentionally isolated
- No premature optimization
- No mocks until necessary
- Real infrastructure from day one
- Tests favor integration over abstraction

---

_End of plan_

