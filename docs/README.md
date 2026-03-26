# docs/

Story Engine is a multi-service platform for structured story authoring with LLM-powered memory and retrieval. The monorepo has:

- **main-service/** — Go backend API (net/http stdlib, pgx/v5, gRPC, PostgreSQL) — source of truth for all entities
- **llm-gateway-service/** — Go LLM gateway (embeddings, extraction, retrieval, Gemini) — ingests from main service via gRPC + Redis
- **obsidian-plugin/** — TypeScript Obsidian plugin (React, SSE streaming) — authoring UI

---

## product/ (1xx)
| File | Description |
|------|-------------|

## tech/ (2xx)
| File | Description |
|------|-------------|

## how-to/ (3xx)
| File | Description |
|------|-------------|
| [300-main.grpc-toolchain-setup.md](how-to/300-main.grpc-toolchain-setup.md) | Install protoc, protoc-gen-go, protoc-gen-go-grpc on WSL2 and compile protos |
| [301-local-dev-setup.md](how-to/301-local-dev-setup.md) | Get the full stack running locally: PostgreSQL, pgvector, Redis, both Go services |
| [302-ollama-wsl2-setup.md](how-to/302-ollama-wsl2-setup.md) | Configure Ollama on Windows to serve embeddings to the LLM Gateway in WSL2 |

## Numbering scheme

`XYZ` — X = category (1 product, 2 tech, 3 how-to), Y = subdivision (0 main-service, 1 llm-gateway, 2 obsidian-plugin), Z = sequence
