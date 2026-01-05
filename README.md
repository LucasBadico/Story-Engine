# Story Engine

**Story Engine** is an experimental, educational, and exploratory platform for
AI-assisted storytelling, focused on narrative structure, long-form continuity,
and creative tooling for authors.

The project explores how stories can be treated as **structured, evolving systems**
rather than isolated blocks of text — combining software architecture,
narrative design, and large language models (LLMs).

---

## Vision

Story Engine aims to investigate and demonstrate how AI can assist creative writing
**without replacing authorship**, by providing:

* structure instead of chaos
* memory instead of repetition
* control instead of randomness

At its core, the project treats stories as **living systems**:
with versions, history, rules, relationships, and state.

---

## Project Status & Intent

This is currently a **personal and experimental project**.

I’m actively using this repository as a foundation for:

* studying narrative systems and storytelling theory
* exploring LLM-assisted writing workflows
* experimenting with software architecture and design
* creating videos, classes, and technical content
* building tools that I personally want to use as a writer and developer

At this stage, the project is a **creative playground**.
The focus is on learning, iteration, and building interesting things
without external pressure or commercial commitments.

Contributions, discussions, and learning-oriented usage are welcome.

---

## Core Ideas

Story Engine is built around a few fundamental principles:

* **Stories are structured systems**, not just text
* **Every version of a story is a first-class entity**
* **Narrative evolution should be inspectable and reversible**
* **LLM usage must be deterministic, debuggable, and replaceable**
* **Domain logic must be isolated from infrastructure**

---

## What the System Does

Story Engine is designed to support professional-grade fiction workflows by:

* Managing stories as structured entities (`story → chapter → scene → beat`)
* Supporting **full story versioning by cloning**, with branching and history graphs
* Tracking characters, relationships, world rules, and narrative state
* Integrating with multiple LLM providers dynamically per workspace
* Preserving long-term narrative memory using embeddings (pgvector)
* Enabling controlled, multi-step chapter generation pipelines

The goal is not to “generate text”, but to **orchestrate narrative evolution**.

---

## Architecture Overview

The project follows **Clean Architecture / Hexagonal Architecture** principles:

```
Domain (core)
   ↑
Application (use cases)
   ↑
Ports (interfaces)
   ↑
Adapters (DB, LLMs, Embeddings)
   ↑
Transports (HTTP / gRPC)
```

### Key Architectural Principles

* The **domain does not depend on frameworks**
* HTTP and gRPC are **entry points**, not business logic
* Infrastructure is fully replaceable
* All layers are testable in isolation

---

## High-Level Structure

```
story-engine/
├── core / internal domain      # Narrative models and rules
├── application                 # Use cases and orchestration
├── adapters                    # DB, LLMs, embeddings
├── transport                   # HTTP / gRPC interfaces
├── proto                       # Protobuf contracts
├── migrations                  # Database schema
├── obsidian-plugin             # Authoring integration
├── web-app                     # UI experiments
├── docs                        # Architecture & decisions
└── scripts                     # Tooling and automation
```

> Some directories may evolve or be split as the project matures.

---

## Core Domains

### Multi-tenancy

* Workspaces isolate data, configuration, and LLM usage
* Users may belong to multiple workspaces
* Each workspace has its own active LLM profile

### Story & Versioning

* Every story version is a **complete clone**
* Versions form a graph (`root_story_id`, `previous_story_id`)
* Forking, branching, and promotion are first-class concepts
* Version numbers are UI-only; identity is structural

### LLM Integration

* Provider-agnostic design
* Multiple providers per workspace
* Dynamic model and parameter selection
* Separate LLM roles:

  * Planner
  * Writer
  * Summarizer / Extractor

### Memory & State

* Embedding-based long-term memory (pgvector)
* Structured narrative state for:

  * Character relationships
  * Emotional tension
  * Open plot threads
  * World consistency

---

## Interfaces & Integrations

* **HTTP API**

  * Public-facing
  * Ideal for UIs and plugins (e.g. Obsidian)

* **gRPC**

  * Internal service-to-service communication
  * High-performance orchestration

Both share the same application and domain layers.

---

## Tech Stack

* **Language:** Go (Golang)
* **Database:** PostgreSQL
* **Vector Store:** pgvector
* **APIs:** HTTP (REST) + gRPC
* **Migrations:** SQL-based
* **Auth:** Token-based (API keys / sessions)

---

## Roadmap (High-Level)

* [ ] Core domain models
* [ ] Story versioning (clone & promote)
* [ ] LLM profile management
* [ ] Chapter generation pipeline
* [ ] Embedding ingestion & retrieval
* [ ] Obsidian authoring integration
* [ ] UI & authoring tools

---

## Philosophy

Story Engine is not a “prompt wrapper”.

It is a **narrative system** — designed to treat stories as evolving structures
with history, rules, and memory — enabling AI to assist creativity
without replacing the author.

---

## License

This project is licensed under the **Elastic License v2**.

* The source code is available and can be used, modified, and studied by the community.
* You may use this project for personal, educational, or internal purposes.
* **You may NOT offer this software as a hosted or managed service (SaaS), or create a competing commercial service.**

See the `LICENSE` file for full details.
