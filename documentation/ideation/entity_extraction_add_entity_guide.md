# Guia: adicionar uma nova entidade no pipeline de extração

Objetivo: incluir um novo tipo de entidade (ex.: `event`) no fluxo de router → extractor → matcher.

## 1) Phase 1 — Router (tipos disponíveis)
- Atualize a descrição do tipo em `llm-gateway-service/internal/application/extract/entities/type_router.go` (mapa `descriptions`).
- Se necessário, ajuste o prompt em `llm-gateway-service/internal/application/extract/entities/prompts/phase1_entity_type_router.prompt`.

## 2) Phase 2 — Extractor (prompt + use case)
- Crie um prompt novo em `llm-gateway-service/internal/application/extract/entities/prompts/phase2_<tipo>_extractor.prompt`.
- Inclua o `go:embed` e um construtor no arquivo `llm-gateway-service/internal/application/extract/entities/entity_extractor.go`:
  - `NewPhase2<Tipo>ExtractorUseCase(...)`
  - `phase2PromptTemplateByType` deve retornar o novo template.
- Registre o extractor no entrypoint em `llm-gateway-service/internal/application/extract/entities/entrypoint.go` (mapa `extractors`).

## 3) Phase 2 — tipos padrão do pipeline
- Se o tipo deve estar ativo por padrão, inclua-o em
  `llm-gateway-service/internal/application/extract/orchestrator.go`
  na lista `entityTypes` default.
  - O orchestrator público agora é `ExtractOrchestrator` e o input é `ExtractRequest`.

## 4) Phase 3 — Matcher (mapeamento para SourceType)
- Adicione o mapeamento em
  `llm-gateway-service/internal/application/extract/entities/match.go`
  na função `mapEntityTypeToSourceType`.
- Confirme o `SourceType` em `llm-gateway-service/internal/core/memory/document.go`.

## 5) Testes e fixtures
- Atualize `llm-gateway-service/internal/application/extract/test/phase1_entity_type_router_integration_test.go`:
  - `allowed` deve incluir o novo tipo.
  - `EntityTypes` de input deve conter o novo tipo se quiser que o router o considere.
- Se desejar cobertura extra:
  - Adicione um caso simples no Phase 2 integration test com texto explícito do novo tipo.
  - Adicione um teste unitário de `phase2_entity_extractor.go` para garantir prompt e parsing.

## 6) API / front-end
- O endpoint de extração já retorna tipos genéricos. Garanta que o front end:
  - Mostre entidades desconhecidas sem quebrar.
  - Se houver ações de “create/update”, implemente o adapter específico do novo tipo.

## Checklist rápido
- Prompt Phase1 atualizado
- Prompt Phase2 criado + embed + switch
- Extractor registrado no Phase2 entrypoint
- Tipo padrão atualizado (se necessário)
- Matcher mapeado para SourceType
- Testes ajustados
