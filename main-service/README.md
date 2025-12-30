# Story Engine

**Story Engine** is a modular, multi-tenant, AI-assisted storytelling platform designed to generate long-form serialized narratives with strong continuity, version control, and structured narrative evolution.

The system is built to support professional-grade fiction workflows, enabling authors to plan, generate, revise, fork, and evolve stories over time using Large Language Models (LLMs), embeddings, and a structured narrative state engine.

---

## ✨ Core Concepts

Story Engine is based on a few fundamental ideas:

* **Stories are structured systems**, not just text
* **Every story version is a first-class entity**
* **LLM usage must be deterministic, inspectable, and replaceable**
* **Domain logic must be isolated from infrastructure**

---

## 🧠 What the System Does

* Manages stories as structured entities (stories → chapters → scenes → beats)
* Supports **full story versioning by cloning** (with branching and history graphs)
* Tracks characters, relationships, world rules, and narrative state
* Integrates with multiple LLM providers dynamically per workspace
* Uses embeddings to preserve long-term narrative memory
* Enables controlled, multi-step chapter generation pipelines

---

## 🏗️ Architecture Overview

The project follows a **Clean Architecture / Hexagonal Architecture** approach:

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

### Key Principles

* The **domain does not depend on frameworks**
* HTTP and gRPC are **entry points**, not business logic
* Infrastructure is fully replaceable
* Everything is testable in isolation

---

## 📁 Project Structure

```
story-engine/
├── cmd/                # Entry points (HTTP, gRPC, workers)
├── internal/
│   ├── core/           # Domain models and rules
│   ├── application/    # Use cases and orchestration
│   ├── ports/          # Interfaces (repositories, LLMs, memory)
│   ├── adapters/       # Implementations (Postgres, OpenAI, etc.)
│   ├── transport/      # HTTP and gRPC handlers
│   └── platform/       # Cross-cutting concerns
├── proto/              # Protobuf definitions
├── migrations/         # Database migrations
├── docs/               # Architecture and decision records
└── scripts/            # Tooling and automation
```

---

## 🧩 Core Domains

### 🏢 Multi-tenancy

* Workspaces (tenants) isolate data, configuration, and LLM usage
* Users may belong to multiple workspaces
* Each workspace has its own active LLM profile

### 📚 Story & Versioning

* Every story version is a **complete clone**
* Versions are connected via a graph (`root_story_id`, `previous_story_id`)
* Authors can fork, branch, and promote versions freely
* Version numbers are UI-only; identity is structural

### 🤖 LLM Integration

* Provider-agnostic design
* Multiple providers supported per workspace
* Dynamic selection of models and parameters
* Separate LLM roles:

  * Planner
  * Writer
  * Summarizer / Extractor

### 🧠 Memory & State

* Embedding-based memory (pgvector)
* Structured narrative state for:

  * Relationships
  * Emotional tension
  * Open plot threads
  * World consistency

---

## 🔌 Interfaces & Transports

* **HTTP API**

  * Public-facing
  * Ideal for UI clients and plugins (e.g. Obsidian)

* **gRPC**

  * Internal service-to-service communication
  * High-performance orchestration

> HTTP and gRPC share the same application layer and domain logic.

---

## 🧪 Testing Philosophy

* Domain: pure unit tests
* Application: mocked ports
* Adapters: integration tests
* Transport: minimal, focused tests

---

## 🚀 Tech Stack

* **Language:** Go (Golang)
* **Database:** PostgreSQL
* **Vector Store:** pgvector
* **APIs:** HTTP (REST) + gRPC
* **Migrations:** SQL-based
* **Auth:** Token-based (API keys / sessions)

---

## 📌 Current Status

🚧 Early-stage architecture and foundations
The project is under active design and implementation.

---

## 🛣️ Roadmap (High-Level)

* [ ] Core domain models
* [ ] Story versioning (clone & promote)
* [ ] LLM profile management
* [ ] Chapter generation pipeline
* [ ] Embedding memory ingestion & retrieval
* [ ] Obsidian integration
* [ ] UI / Authoring tools

---

## 🤝 Philosophy

Story Engine is not a “prompt wrapper”.

It is a **narrative system**, designed to treat stories as evolving structures with history, rules, and memory — enabling AI to assist creativity without replacing authorship.

---

## 📄 License

TBD
