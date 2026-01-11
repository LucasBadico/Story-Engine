# Próximas Etapas - Sync V2

> **Status**: Atualizado em 2025-01-09
> **Fase Atual**: Fase 9 (em progresso)
> **Última Atualização**: Implementação de SceneHandler.push() com criação automática de POV/Location relations

## Resumo Executivo

### ✅ Fases Completas
- **Fase 1**: Foundation (Core + Types) - ✅ Completa
- **Fase 2**: Parsers - ✅ Completa
- **Fase 3**: Generators - ✅ Completa (incluindo FrontmatterGenerator)
- **Fase 4**: Story Handlers - ✅ Completa (pull implementado)
- **Fase 5**: File Renaming - ✅ Completa
- **Fase 6**: Sync Orchestrator - ✅ Completa (DiffEngine, ContentsReconciler, PushPlanner, PushExecutor)
- **Fase 7**: World Handlers - ✅ Completa
- **Fase 8**: Relations & Citations System - ✅ Completa (pull implementado)

### 🚧 Fases Em Progresso
- **Fase 9**: Push Relations & Auto Sync - 🚧 Parcialmente completa
  - ✅ Push de relations via `.relations.md` (RelationsPushHandler)
  - ✅ Criação automática de POV/Location relations em SceneHandler
  - ⏳ Citations automáticas em ContentBlocks (pendente)
  - ⏳ Refatorar AutoSyncManager (pendente)
  - ⏳ Implementar ConflictResolver (pendente)
  - ⏳ Integrar ApiUpdateNotifier (pendente)

### 📋 Fases Pendentes
- **Fase 10**: Migration & Testing
- **Fase 11**: Backups & Git Integration

---

## Prioridade: Alta (Próximas Implementações)

### 1. Fase 7: Migrar World Handlers para FrontmatterGenerator ✅

**Status**: Completo (FrontmatterGenerator criado e integrado)

**Tarefas:**
- [x] Migrar `CharacterHandler.renderCharacter()` para usar `FrontmatterGenerator`
  - ✅ Substituir renderização manual por `FrontmatterGenerator.generate()`
  - ✅ Mapear campos: `id`, `world_id`, `class_level`, `archetype_id`, `current_class_id`, `created_at`, `updated_at`
  - ✅ Adicionar tags: `story-engine/character`, `world/{world-name}`
- [x] Migrar `LocationHandler.renderLocation()` para usar `FrontmatterGenerator`
  - ✅ Mapear campos: `id`, `world_id`, `type`, `hierarchy_level`, `parent_id`, `created_at`, `updated_at`
  - ✅ Adicionar tags: `story-engine/location`, `world/{world-name}`
- [x] Migrar `FactionHandler.renderFaction()` para usar `FrontmatterGenerator`
  - ✅ Mapear campos: `id`, `world_id`, `type`, `hierarchy_level`, `parent_id`, `created_at`, `updated_at`
  - ✅ Adicionar tags: `story-engine/faction`, `world/{world-name}`
- [x] Migrar `ArtifactHandler.renderArtifact()` para usar `FrontmatterGenerator`
  - ✅ Mapear campos: `id`, `world_id`, `rarity`, `created_at`, `updated_at`
  - ✅ Adicionar tags: `story-engine/artifact`, `world/{world-name}`
- [x] Migrar `EventHandler.renderEvent()` para usar `FrontmatterGenerator`
  - ✅ Mapear campos: `id`, `world_id`, `type`, `importance`, `timeline`, `parent_id`, `created_at`, `updated_at`
  - ✅ Adicionar tags: `story-engine/event`, `world/{world-name}`
- [x] Migrar `LoreHandler.renderLore()` para usar `FrontmatterGenerator`
  - ✅ Mapear campos: `id`, `world_id`, `category`, `parent_id`, `hierarchy_level`, `created_at`, `updated_at`
  - ✅ Adicionar tags: `story-engine/lore`, `world/{world-name}`
- [x] Migrar `ArchetypeHandler.renderArchetype()` para usar `FrontmatterGenerator`
  - ✅ Mapear campos: `id`, `tenant_id`, `created_at`, `updated_at`
  - ✅ Adicionar tags: `story-engine/archetype`
- [x] Migrar `TraitHandler.renderTrait()` para usar `FrontmatterGenerator`
  - ✅ Mapear campos: `id`, `tenant_id`, `category`, `created_at`, `updated_at`
  - ✅ Adicionar tags: `story-engine/trait`
- [ ] Atualizar testes para refletir uso do `FrontmatterGenerator` (se necessário)
- [x] Verificar consistência de campos entre todos os handlers

**Dependências:**
- FrontmatterGenerator criado ✅

**Complexidade**: Média (refatoração, mas direta)
**Impacto**: Médio - Melhora consistência e manutenibilidade

---

### 2. Fase 9: Citations Automáticas em ContentBlocks ✅

**Status**: Implementado (criação automática de citations ao fazer push)

**Tarefas:**
- [x] Implementar `ContentBlockHandler.push()` básico (atualizar conteúdo) ✅
- [x] Adicionar helper `detectEntityMentions()` para detectar menções via parsing de links ✅
  - ✅ Detectar links no formato `[[filename path]]` ou `[[filename path|display]]`
    - Exemplos: `[[worlds/eldoria/characters/aria-moon]]`, `[[worlds/eldoria/locations/crystal-cave]]`
    - Suporta todas as World entities: character, location, faction, artifact, event, lore
  - ✅ Resolver filename path para ID via parsing de frontmatter:
    - ✅ Suporta formato oficial (`[[worlds/eldoria/characters/aria-moon]]`)
    - ✅ Suporta formato Obsidian (`[[aria-moon]]` com inferência via metadataCache)
    - ✅ Inferência de tipo via frontmatter (tags `story-engine/{type}` ou campo `entity_type`)
    - ✅ Suporta campo ID customizado via settings (`frontmatterIdField`)
  - ✅ **Importante**: Links devem ser renderizados com filename path completo para evitar ambiguidade
    - Motivo: pode haver 2 characters com nomes diferentes no mesmo world
    - Formato: `worlds/{world-name}/{entity-type}/{entity-slug}`
- [x] Adicionar helper `resolveContentBlockHierarchy()` para obter ContentAnchors e determinar nível ✅
  - ✅ Obter ContentAnchors do ContentBlock via API
  - ✅ Determinar nível mais específico (Beat > Scene > Chapter)
  - ✅ Fallback para chapter_id direto se não houver anchors
- [x] Adicionar helper `buildHierarchyContext()` para construir string de contexto ✅
  - ✅ Formato: `"Chapter 1: Introduction > Scene 2: The Meeting > Beat 3: Confrontation"`
- [x] Adicionar helper `createCitationRelations()` para criar citations no nível correto ✅
  - ✅ Validar que target entity existe antes de criar (via API get)
  - ✅ Criar citation relation com `source_type` correto (beat/scene/chapter/content_block)
  - ✅ Incluir context string completa no campo `context`
  - ✅ Evitar duplicatas verificando relações existentes
- [x] Integrar com ContentAnchors API para determinar hierarquia ✅
- [x] Validar que target entity existe antes de criar citation ✅
- [x] Integrar detecção de menções e criação de citations no `ContentBlockHandler.push()` ✅
- [ ] Integrar com LLM/API para detecção automática (Fase 2 - quando LLM estiver pronto)
- [ ] Atualizar `.citations.md` files automaticamente quando citations são criadas (será feito via pull)

**Dependências:**
- ContentAnchors API (já disponível)
- EntityRelation API (já disponível)
- LLM/API para detecção automática (trabalhado em outra thread)

**Complexidade**: Média-Alta
**Impacto**: Alto - Permite rastrear automaticamente onde World entities são mencionadas

---

### 3. Fase 9: Refatorar AutoSyncManager ✅

**Status**: Completo (AutoSyncManagerV2 implementado e testado)

**Tarefas:**
- [x] Analisar `AutoSyncManager` atual (V1) ✅
- [x] Refatorar para usar Sync V2 handlers ✅
- [x] Implementar debounce e batching para operações ✅
  - ✅ 1s typing pause (debounce)
  - ✅ 5s idle (batching)
  - ✅ blur event (active-leaf-change)
- [x] Integrar com `SyncOrchestrator` do V2 ✅
- [x] Manter compatibilidade com V1 (fallback baseado em `settings.syncVersion`) ✅
- [x] Implementar fila de operações pendentes (pendingOperations map e operationQueue) ✅
- [x] Criar testes para AutoSyncManager V2 ✅
ependências:**
- Sync V2 handlers completos
- SyncOrchestrator pronto

**Complexidade**: Alta
**Impacto**: Alto - UX do auto-sync é crítica

---

### 4. Fase 9: Implementar ConflictResolver 🚧

**Status**: Em Progresso (estrutura básica completa, falta integração)

**Tarefas:**
- [x] Definir tipos de conflitos ✅
  - ✅ Conflito de modificação simultânea (local vs remote)
  - ⏳ Conflito de renomeação (arquivo já existe) - TODO
  - ⏳ Conflito de deleção (entidade removida remotamente) - TODO
- [x] Implementar estratégias de resolução ✅
  - ✅ Manual (usuário escolhe - fallback para local por enquanto)
  - ✅ Automática local (always use local)
  - ✅ Automática remote (always use remote/service)
  - ✅ Last-write-wins (implementado, mas não mapeado em settings ainda)
  - ⏳ Merge inteligente (para conteúdo textual) - TODO
- [x] Criar testes básicos ✅
- [x] Integrar com DiffEngine para detectar conflitos em conteúdo ✅ (integrado no StoryHandler)
- [ ] Criar UI genérica para resolução de conflitos (melhorar ConflictModal existente)
- [x] Integrar ConflictResolver com SyncOrchestrator/handlers ✅ (integrado no StoryHandler.pull())
- [ ] Adicionar logs estruturados de conflitos

**Dependências:**
- DiffEngine pronto
- SyncOrchestrator pronto

**Complexidade**: Alta
**Impacto**: Alto - Crítico para operação em equipe

---

### 5. Fase 9: Integrar ApiUpdateNotifier ⚠️

**Status**: Pendente

**Tarefas:**
- [ ] Definir eventos de notificação da API:
  - Entity criada/atualizada/deletada
  - Relation criada/atualizada/deletada
- [ ] Implementar polling ou WebSocket para receber notificações
- [ ] Processar notificações e atualizar arquivos locais
- [ ] Debounce de notificações para evitar flood
- [ ] Resolver conflitos quando local e remote mudam simultaneamente
- [ ] Notificar usuário sobre mudanças remotas

**Dependências:**
- API com suporte a notificações (verificar se existe)
- ConflictResolver pronto

**Complexidade**: Média-Alta
**Impacto**: Médio-Alto - Melhora colaboração em tempo real

---

## Prioridade: Média (Próximas Fases)

### 6. Fase 10: Migration & Testing 📋

**Status**: Pendente

**Tarefas:**
- [ ] Script de migração de formato antigo (V1 → V2)
  - Converter estrutura de pastas
  - Migrar arquivos para novo formato
  - Validar integridade dos dados
- [ ] Testes end-to-end completos
  - Pull completo de story com world entities
  - Editar arquivos localmente
  - Push de mudanças
  - Validar sync bidirecional
- [ ] Testes de regressão (paridade com V1)
  - Comparar resultados V1 vs V2
  - Validar que nenhuma funcionalidade foi perdida
- [ ] Documentação de uso
  - Guia de migração V1 → V2
  - Manual de uso do Sync V2
  - Troubleshooting

**Dependências:**
- Todas as fases anteriores completas
- AutoSyncManager refatorado
- ConflictResolver implementado

**Complexidade**: Média
**Impacto**: Alto - Necessário para adoção em produção

---

### 7. Fase 11: Backups & Git Integration 📋

**Status**: Pendente

**Tarefas:**
- [ ] Implementar snapshots automáticos antes de cada `pull`/`push`
  - Estrutura: `backups/YYYY-MM-DD/HH-mm-ss/`
  - Manter apenas arquivos modificados
  - Manifesto JSON com lista de arquivos
- [ ] Criar flag nas configurações para ativar/desativar backup Git
- [ ] Integrar com CLI (`git status/add/commit`) quando flag estiver ativa
- [ ] Adicionar comandos para restaurar snapshot/Git commit dentro do plugin
- [ ] Retenção: manter snapshots dos últimos 7 dias
- [ ] Integração com Git (opcional, opt-in)

**Dependências:**
- FileManager pronto
- Sync V2 estável

**Complexidade**: Média
**Impacto**: Médio - Importante para segurança, mas não crítico para MVP

---

## Tarefas Menores/Opcionais (Melhorias Futuras)

### AutoSyncManager - Melhorias
- [ ] Implementar detecção de rede offline/online (para modo offline real)
  - Usar `navigator.onLine` + eventos `online`/`offline`
  - Health check da API para confirmar conectividade real
  - Fila persistente (localStorage/vault) para sobreviver reloads
  - Auto-processamento quando internet voltar
  - UI/notificação mostrando operações pendentes
- [ ] Adicionar logs estruturados
- [ ] Implementar métricas de performance
- **Prioridade**: Baixa (funcionalidade básica já implementada, melhorias incrementais)

### FrontmatterGenerator (Fase 3)
- [ ] Implementar `FrontmatterGenerator` centralizado (já implementado ✅)
- [ ] Reaproveitar lógica para story/chapter/scene/beat (já implementado ✅)
- **Prioridade**: Baixa (geradores específicos já funcionam)

---

## Roadmap Sugerido

### Sprint 1 (Alta Prioridade)
1. ✅ Implementar SceneHandler.push() com POV/Location relations (COMPLETO)
2. ✅ Criar FrontmatterGenerator (COMPLETO)
3. ✅ Migrar World handlers para usar FrontmatterGenerator (COMPLETO)
4. ✅ Implementar ContentBlockHandler.push() com citations automáticas (COMPLETO)

### Sprint 2 (Alta Prioridade)
5. ✅ Refatorar AutoSyncManager básico (COMPLETO)
6. ✅ Testes de integração para citations automáticas (COMPLETO)
7. ✅ Integrar AutoSyncManager com Sync V2 (COMPLETO)
8. ⏳ Implementar ConflictResolver básico
9. ⏳ Melhorar AutoSyncManager (offline detection, logs, métricas)

### Sprint 3 (Média Prioridade)
8. ⏳ Implementar ConflictResolver básico
9. ⏳ Integrar ApiUpdateNotifier
10. ⏳ Melhorar ConflictResolver (merge inteligente)
11. ⏳ Testes end-to-end básicos

### Sprint 4 (Média Prioridade)
12. ⏳ Script de migração V1 → V2
13. ⏳ Testes de regressão completos
14. ⏳ Documentação de uso

### Sprint 5 (Baixa Prioridade)
15. ⏳ Backups & Git Integration
16. ⏳ Otimizações e polish

---

## Bloqueios e Riscos

### Bloqueios Atuais
- **Nenhum bloqueio crítico**

### Riscos Identificados
1. **LLM/API para detecção automática**: Trabalhado em outra thread, pode atrasar Fase 2 das citations
   - **Mitigação**: Implementar Fase 1 (manual) primeiro, adicionar LLM depois
2. **ApiUpdateNotifier**: Pode não existir na API atual
   - **Mitigação**: Implementar polling como alternativa inicial
3. **Complexidade do ConflictResolver**: Pode ser mais complexo do que esperado
   - **Mitigação**: Começar com estratégias simples (manual, last-write-wins)

---

## Métricas de Progresso

### Fase 9 (Push Relations & Auto Sync)
- ✅ Push de relations: **100%** (5/5 tarefas)
- ✅ Auto-creation básica: **66%** (2/3 tarefas)
- ✅ Citations automáticas: **90%** (8/9 tarefas - faltando apenas LLM e auto-update de .citations.md)
- ✅ AutoSyncManager: **100%** (7/7 tarefas - implementado com testes completos)
- 🚧 ConflictResolver: **75%** (6/8 tarefas - estrutura básica + testes + integração com StoryHandler, falta UI e logs)
- ⏳ ApiUpdateNotifier: **0%** (0/1 tarefa)
- **Progresso Geral Fase 9**: ~82%

### Projeto Completo
- **Fases Completas**: 8/11 (73%)
- **Fases Em Progresso**: 1/11 (9%)
- **Fases Pendentes**: 2/11 (18%)
- **Progresso Geral**: ~75%

---

## Notas Importantes

1. **Citations Automáticas** é a próxima prioridade alta, com impacto significativo na funcionalidade
2. **AutoSyncManager** precisa ser refatorado para V2 antes de considerar V2 pronto para uso
3. **ConflictResolver** é crítico para operação em equipe, mas pode ser simplificado inicialmente
4. **Fase 10 (Migration & Testing)** deve ser iniciada apenas após Fase 9 estar completa
5. **Fase 11 (Backups)** pode ser adiada para depois do MVP, mas é importante para produção

---

## Referências

- [Plano Arquitetural Completo](./sync_v2_architecture_plan.md)
- [Fase 9: Push Relations & Auto Sync - Seção Detalhada](./sync_v2_architecture_plan.md#fase-9-push-relations--auto-sync)
- [Opção 3: Citations Automáticas - Decisões de Design](./sync_v2_architecture_plan.md#criação-automática-de-citations-em-contentblocks)

