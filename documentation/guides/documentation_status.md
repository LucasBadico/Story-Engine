# Documentation Status Map

This is a cross‑check of documentation vs the current codebase. It lists each document, a short summary, and whether the described functionality appears implemented, partially implemented, or missing.

Legend:
- **Implemented**: feature exists in codebase
- **Partial / Outdated**: feature exists but doc has gaps or the feature is incomplete
- **Planned / Not Implemented**: doc is ideation or the feature is not found in code
- **Reference**: external reference, not a feature spec

## Root Documentation

- `documentation/LLM_GATEWAY_SERVICE.md`
  - Summary: LLM Gateway ingestion, embeddings, search, worker, pgvector.
  - Status: **Partial / Outdated**.
  - Notes: Does not mention the current LLM executor for extraction, relation extraction pipeline, relation ingestion summaries, or relation map loading. Embedding providers (OpenAI/Ollama) still valid, but Gemini LLM usage is not documented here.

- `documentation/Story_Engine_API.postman_collection.json`
  - Summary: Postman collection for main-service REST API.
  - Status: **Partial / Outdated**.
  - Notes: Collection appears built around tenant_id in body/query; code now favors `X-Tenant-ID` header. Relation endpoints and static relation map endpoints are not represented.

- `documentation/Story_Engine_Local.postman_environment.json`
  - Summary: Postman env for local testing.
  - Status: **Partial / Outdated**.
  - Notes: Environment likely still usable, but missing relation maps endpoints and any LLM gateway endpoints.

- `documentation/Story_Workflows.postman_collection.json`
  - Summary: Postman workflows collection.
  - Status: **Partial / Outdated**.
  - Notes: Similar to API collection; verify headers and new endpoints.

## Guides

- `documentation/guides/QUICK_START_LLM_GATEWAY.md`
  - Summary: Run ingestion worker + verify embeddings.
  - Status: **Partial / Outdated**.
  - Notes: Still valid for ingestion basics, but does not mention relation ingestion, LLM executor, or new env vars (e.g., `MAIN_SERVICE_HTTP_ADDR`). Go version is older than current.

- `documentation/guides/REST_API_Quick_Reference.md`
  - Summary: Curl examples for REST API.
  - Status: **Outdated**.
  - Notes: Uses tenant_id in body/query. Code uses `X-Tenant-ID` header for most routes. Missing relation endpoints and static relation map endpoints.

- `documentation/guides/POSTMAN_SETUP.md`
  - Summary: How to import/use Postman collection.
  - Status: **Partial / Outdated**.
  - Notes: Mentions many endpoints and variables; likely mismatch with current auth/tenant headers and relation endpoints.

- `documentation/guides/offline-mode.md`
  - Summary: Offline SQLite server (no auth) with main-service endpoints.
  - Status: **Partial**.
  - Notes: Offline mode exists and matches most routes, but the guide does not reflect relation maps endpoints and recent migration changes. Still broadly accurate.

- `documentation/guides/entity_extraction_flow.md`
  - Summary: End‑to‑end extraction flow (ingestion → phases → stream → UI).
  - Status: **Implemented**.
  - Notes: This document reflects the current pipeline, SSE events, and Obsidian UI flow.

## Ideation (Plans / Proposals)

- `documentation/ideation/00-proposal.md`
  - Summary: Initial project proposal and architecture.
  - Status: **Partial / Outdated** (high‑level vision). Many ideas implemented, but the doc is historical.

- `documentation/ideation/01-entidades_multitenancy_versionamento_e_config_llm.md`
  - Summary: Entities, multitenancy, versioning, dynamic LLM profiles.
  - Status: **Partial / Not Implemented**.
  - Notes: Multitenancy and versioning exist; dynamic per‑workspace LLM profiles are not implemented.

- `documentation/ideation/02-project_phases_plan_story_engine.md`
  - Summary: Project phases plan.
  - Status: **Partial**.
  - Notes: Some phases (core entities, gRPC, Obsidian MVP, ingestion, extraction) are implemented. Later phases (narrative generation MVP) are not present.

- `documentation/ideation/relations_extraction_plan.md`
  - Summary: Detailed relation extraction phases and API payloads.
  - Status: **Mostly Implemented**.
  - Notes: Phase 4–8 pipeline, SSE result_entities/result_relations, relation maps, normalization and mirror behavior are implemented. Remaining gaps: existing-relations snapshot for dedup, deeper inference, and some prompt refinements.

- `documentation/ideation/Plano — Relações E Inferências No Extract (v2).md`
  - Summary: Revised relation extraction plan.
  - Status: **Mostly Implemented** (merged into v1 plan).
  - Notes: Core phases implemented. Some advanced inference/dedup rules still pending.

- `documentation/ideation/relations_discovery_ingestion_plan.md`
  - Summary: Unified relations table + ingestion pipeline plan.
  - Status: **Partial / Outdated**.
  - Notes: Unified entity_relations exists; relation ingestion exists in LLM gateway. Doc still references compile breakages and older schema fields.

- `documentation/ideation/ingestion_queue_guide.md`
  - Summary: Main-service queueing pattern for ingestion.
  - Status: **Implemented**.

- `documentation/ideation/ingestion_reindex_pipeline_arch.md`
  - Summary: Reindex architecture for ingestion.
  - Status: **Partial / Not Implemented**.
  - Notes: Not fully present in code (no end‑to‑end reindex pipeline).

- `documentation/ideation/basic_search_endpoint_plan.md`
  - Summary: Search endpoint in LLM gateway.
  - Status: **Implemented**.
  - Notes: Search endpoint exists in LLM gateway.

- `documentation/ideation/llm_executor_arch.md`
  - Summary: LLM executor architecture.
  - Status: **Implemented**.

- `documentation/ideation/llm_executor_guide.md`
  - Summary: Usage guide for LLM executor.
  - Status: **Implemented**.

- `documentation/ideation/entity_extract_sse_events.md`
  - Summary: SSE event types for extraction.
  - Status: **Partial / Outdated**.
  - Notes: Many events exist; result_entities and result_relations are used now. Relation-specific events were added (relation.success/error) and may not be documented.

- `documentation/ideation/entity_extraction_add_entity_guide.md`
  - Summary: Steps to add a new entity to extraction pipeline.
  - Status: **Mostly Implemented**.

- `documentation/ideation/lang_graph_para_entity_extraction_no_story_engine.md`
  - Summary: LangGraph orchestration plan.
  - Status: **Not Implemented**.

- `documentation/ideation/obsidian_plugin_auth_flow.md`
  - Summary: Auth flow ideas for Obsidian plugin.
  - Status: **Not Implemented**.

- `documentation/ideation/obsidian_plugin_ui_react_proposta_e_boas_praticas.md`
  - Summary: React-based plugin UI proposal.
  - Status: **Not Implemented** (current plugin uses native Obsidian view rendering).

- `documentation/ideation/arquitetura_apps_irmas_ui_package_web_obsidian.md`
  - Summary: Sister apps + shared UI package architecture.
  - Status: **Not Implemented**.

- `documentation/ideation/guia_de_uso_do_hero_ui_com_design_tokens.md`
  - Summary: UI tokens guide.
  - Status: **Not Implemented**.

- `documentation/ideation/sync_v2_architecture_plan.md`
  - Summary: Sync V2 architecture.
  - Status: **Partial / Outdated**.
  - Notes: Core Sync V2 stack exists (ModularSyncEngine, SyncOrchestrator, handlers, parsers, generators, diff/push, AutoSyncManagerV2, ConflictResolver, ApiUpdateNotifier). Doc still reflects early "planning" state and differs on file layout (current uses `00-chapters/01-scenes/02-beats/03-contents` and `worlds/{world}/...` with `_archetypes`/`_traits`). Automatic rollback to V1 on fatal error is not implemented; API-update deletions and relation events are still TODO.

- `documentation/ideation/proximas_etapas_sync_v2.md`
  - Summary: Next steps for Sync V2.
  - Status: **Partial / Outdated**.
  - Notes: Many items marked “in progress/pending” are implemented in this branch (AutoSyncManagerV2, ContentBlock citation detection/push, ConflictResolver basics, ApiUpdateNotifier + handler). Remaining gaps include deletion handling on API updates, relation-event handling, conflict UI, and migration/backups.

- `documentation/ideation/shared_go_module_plan.md`
  - Summary: Shared Go module plan.
  - Status: **Not Implemented**.

- `documentation/ideation/frontmatter_fields_reference.md`
  - Summary: Frontmatter mapping for Sync V2.
  - Status: **Mostly Implemented**.
  - Notes: FrontmatterGenerator and FileManager write most fields listed (including content metadata). Custom ID field via settings is supported but not documented. Some fields remain payload‑dependent (e.g., versioning fields, optional metadata) and may be null if API omits them.

- `documentation/ideation/auth_tenant_roles_arch.md`
  - Summary: Tenant roles/auth architecture.
  - Status: **Not Implemented**.

- `documentation/ideation/llm_executor_arch.md`
  - Summary: LLM executor architecture.
  - Status: **Implemented**.

## Task Related

- `documentation/task related/ENTITY_TEST_STATUS.md`
  - Summary: Status of entity extraction tests.
  - Status: **Partial / Outdated**.

- `documentation/task related/STATUS_SQLITE_IMPLEMENTATION.md`
  - Summary: SQLite implementation status.
  - Status: **Partial** (offline mode exists; migration status may have changed).

- `documentation/task related/RESUMO_SESSAO_SQLITE.md`
  - Summary: SQLite session summary.
  - Status: **Historical**.

- `documentation/task related/WORLD_VIEW_CHARACTER_DETAILS_PLAN.md`
  - Summary: World view/character details UI plan.
  - Status: **Not Implemented** (current UI differs).

- `documentation/task related/phase-2-implementation-plan.md`
  - Summary: Phase 2 implementation plan (entity extraction).
  - Status: **Implemented** (core extraction pipeline exists).

- `documentation/task related/suporte_a_sqlite_e_postgres_no_story_engine_hexagonal.md`
  - Summary: SQLite/Postgres support plan.
  - Status: **Partial**.

- `documentation/task related/TODO`
  - Summary: Active TODO list.
  - Status: **Mixed** (some items implemented; needs review).

- `documentation/todo/move_chapter_endpoint.md`
  - Summary: Plan for chapter move endpoint.
  - Status: **Not Implemented** (no evidence in main-service routes).

## References

- `documentation/referencies/*.pdf`
  - Summary: External RAG references.
  - Status: **Reference**.

## Key Gaps / Recommendations

1) **Update API docs** to reflect `X-Tenant-ID` headers and new relation endpoints/static relation map endpoints.
2) **LLM Gateway docs** should mention extraction + relation pipeline and LLM executor.
3) **Sync V2 docs** are largely aspirational; tag them as “proposal” to avoid confusion.
4) **Relation extraction docs** are mostly aligned but should be updated with recent changes:
   - batch parallelism
   - relation discovery recovery and repair
   - relation success/error SSE events
