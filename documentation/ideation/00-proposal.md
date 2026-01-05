# AI-Assisted Storytelling System

## Project Summary

This project explores the creation of an AI-assisted storytelling system designed to generate long-form serialized novels with strong narrative continuity, character development, and dynamic plot progression.

At its core, the system combines a Large Language Model (LLM) with an embedding-based memory layer and a structured state engine:

- **Embedding Memory** is responsible for retaining and retrieving narrative context such as character profiles, relationships, world rules, themes, and past chapter summaries, ensuring consistency across episodes.
- **State Engine** tracks measurable and evolving story variables—such as character relationships, emotional tension, secrets, unresolved conflicts, and ongoing arcs—allowing the narrative to evolve in a controlled and coherent way.

Stories are generated through a **multi-step pipeline**:
1. Planning narrative beats and chapter goals
2. Retrieving relevant context from memory and state
3. Producing prose guided by structured constraints

This approach enables balanced pacing between dialogue, action, and internal conflict while maintaining long-term plot arcs and cliffhangers typical of serialized web novels.

The system is designed to be **modular, extensible, and reproducible**, supporting multiple genres, tones, and narrative rule sets. It is suitable for experimentation in AI-driven fiction, interactive storytelling, and automated novel generation workflows.

---

## Design Goals

- Long-term narrative continuity
- Controlled plot progression
- Character consistency and growth
- Genre and tone modularity
- Scalable chapter generation
- Reproducible and inspectable outputs

---

## High-Level Architecture

### Core Components

1. **Story Orchestrator**
   - Central coordinator
   - Controls generation flow per chapter
   - Enforces pipeline order (plan → retrieve → write → update)

2. **LLM Generation Engine**
   - Responsible for planning and prose generation
   - Operates strictly based on provided context and constraints

3. **Embedding Memory Layer**
   - Vector database for narrative knowledge
   - Stores:
     - Character profiles
     - World rules and lore
     - Relationship histories
     - Chapter summaries
     - Thematic notes

4. **State Engine**
   - Structured data store (JSON / graph-based)
   - Tracks:
     - Relationship values
     - Emotional tension levels
     - Active secrets
     - Open plot threads
     - Arc progression states

5. **Narrative Planner**
   - Generates chapter-level outlines
   - Defines:
     - POV
     - Core conflict
     - Emotional beats
     - Cliffhangers or resolutions

6. **Consistency Validator (Optional / Advanced)**
   - Post-generation checks
   - Detects contradictions or rule violations
   - Can trigger partial rewrites

---

## Data Flow per Chapter

1. **Chapter Request**
   - Input: story ID, chapter index, optional constraints

2. **Context Retrieval**
   - Query embedding memory for relevant narrative elements
   - Load current structured state snapshot

3. **Narrative Planning**
   - Generate a structured chapter plan
   - Output: beats, goals, pacing markers

4. **Prose Generation**
   - LLM writes chapter text using:
     - Retrieved memory
     - Current state
     - Chapter plan

5. **Post-Processing**
   - Summarize chapter
   - Extract new facts, events, emotional shifts

6. **State & Memory Update**
   - Update state variables
   - Store chapter summary and extracted facts in embeddings

---

## Suggested Technical Stack (Go + Postgres)

- **Language / Services**: **Go (Golang)**
  - HTTP API (REST) and/or gRPC for internal orchestration
  - Clean architecture: `core` (domain) + `services` + `adapters` (LLM, DB)

- **Database**: **PostgreSQL** as the primary store for:
  1) **Raw narrative data** (structured)
  2) **Vector store** (embeddings) using **pgvector**

- **Vector Search**:
  - `pgvector` for storing and querying embeddings (`cosine`, `inner product`, etc.)
  - Hybrid retrieval: `tsvector` (full-text) + vector similarity for best recall

- **LLM Integration**:
  - Provider-agnostic adapter (OpenAI / other)
  - Separate clients for: planning, prose, summarization/extraction

- **Orchestration / Jobs**:
  - Simple synchronous pipeline for MVP
  - Optional async job queue later (e.g., asynq / Redis) for long chapters and retries

- **Evaluation / Validation**:
  - Rule-based checks + optional LLM-based “critic” step for contradictions

---

## Interface & Integrations (Obsidian)

### Can we build an Obsidian plugin that connects via gRPC?
Yes, but **direct gRPC from an Obsidian plugin** can be tricky depending on the transport/runtime.

- Obsidian plugins run in an **Electron** environment (browser-like APIs + some Node capabilities), so **native gRPC** (HTTP/2 + protobuf, `grpc-go` style) is not always the smoothest path.

### Recommended approach

**Option A (Recommended for UX + simplicity):**
- Keep **gRPC internal** between backend services.
- Expose a **public HTTP interface** for the Obsidian plugin using:
  - REST/JSON for basic operations
  - **Streaming** via **SSE** or **WebSocket** for generation events (tokens, progress, step logs)

**Option B (Still gRPC, but plugin-friendly): Connect / gRPC-Web**
- Backend:
  - Implement **Connect** (Buf) or **gRPC-Web** endpoints over HTTP.
  - Same protobuf definitions, but transported in a way that browsers can call.
- Plugin:
  - Uses `fetch()` to call Connect/gRPC-Web endpoints.
  - Avoids native gRPC dependencies.

**Option C (Local sidecar bridge):**
- Run a local desktop helper (Go) that speaks **native gRPC** to the service and exposes **localhost HTTP/WebSocket** for Obsidian.
- Best when you need strict gRPC semantics locally and want plugin simplicity.

### Suggested plugin capabilities
- Sync notes ↔ story memory (push/pull)
- Trigger chapter generation with constraints (POV, tone, arc)
- Show progress stream (plan → retrieve → write → update)
- Insert generated chapter into a note (with metadata/frontmatter)

### Security notes
- Auth token stored in plugin settings
- Per-story API keys / scopes (read vs write)
- Rate limits and audit logs server-side

---

## Extensibility

- Genre modules (romance, fantasy, sci-fi, thriller)
- Tone profiles (dark, light, comedic, dramatic)
- Interactive branches (reader choices)
- Multi-POV orchestration
- Cross-story shared universes

---

## Future Directions

- Reinforcement learning for arc optimization
- Reader-feedback-driven state adjustments
- Automated cover, synopsis, and marketing text generation
- Multi-agent character-driven narrative systems

---

## Conclusion

This architecture balances creative freedom with structural control, enabling AI systems to produce coherent, engaging, and scalable serialized fiction. By separating memory, state, planning, and generation, the system supports long-term storytelling that evolves naturally while remaining technically inspectable and extensible.

