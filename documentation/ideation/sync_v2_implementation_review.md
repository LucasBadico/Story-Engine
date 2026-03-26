# Sync V2 - Revisão de Implementação

> **Status**: Análise (Atualizado)
> **Data**: 2026-01-16
> **Última Revisão**: 2026-01-16
> **Documento de Referência**: `sync_v2_architecture_plan.md`

## Resumo Executivo

Após análise detalhada do plano de arquitetura e da implementação atual em `obsidian-plugin/src/sync-v2/`, a implementação evoluiu significativamente desde a última revisão. A maioria dos itens críticos foi endereçada e o sistema está aproximadamente **85-90% completo**.

---

## Índice

1. [Aspectos Positivos](#aspectos-positivos)
2. [Problemas Identificados](#problemas-identificados)
3. [Análise de Testes](#análise-de-testes)
4. [Divergências do Plano](#divergências-do-plano)
5. [Ações Recomendadas](#ações-recomendadas)
6. [Priorização](#priorização)

---

## Aspectos Positivos

### ✅ Arquitetura Modular
- Excelente separação entre handlers (`handlers/story/`, `handlers/world/`)
- Generators independentes para cada tipo de arquivo (outline, contents, relations, citations)
- Parsers bem definidos com interface clara
- **NOVO**: `StoryHandler` refatorado em 4 serviços dedicados:
  - `StoryRelationsService` - Geração e push de relations
  - `StoryConflictService` - Detecção e resolução de conflitos
  - `StoryRenameService` - Renomeação de arquivos por reorder
  - `StoryFileService` - Operações de I/O

### ✅ Sistema de Tipos Robusto
- `types/sync.ts` define contratos claros (`SyncOperation`, `SyncResult`, `SyncContext`)
- Tipos para fences bem estruturados (`ParsedFence`, `FenceChange`)
- Discriminated unions para operações de sync

### ✅ Handlers Completos
- Todos os story handlers implementados (Story, Chapter, Scene, Beat, ContentBlock)
- Todos os world handlers implementados (Character, Location, Faction, Artifact, Event, Lore, Archetype, Trait)
- Handlers usam `FrontmatterGenerator` consistentemente
- **NOVO**: Todos os world handlers têm `push()` funcional

### ✅ Push Pipeline Funcional
- `PushPlanner` detecta reorders, moves de scenes/beats, **content updates**
- `PushExecutor` aplica ações na API
- `RelationsPushHandler` sincroniza relations bidirecionalmente
- **NOVO**: `OutlinePushHandler` para chapter reorders
- **NOVO**: `ContentCitationService` para citations automáticas

### ✅ AutoSync Implementado
- `AutoSyncManagerV2` com debounce de 1s (typing) e 5s (idle)
- Batching de operações por story folder
- Retry para erros de rede

### ✅ Sistema de Backup Implementado
- **NOVO**: `BackupManager` cria snapshots antes de pull/push
- Arquivos copiados para `.story-engine/backups/{timestamp}/`
- Manifesto JSON com lista de arquivos afetados

### ✅ Citations Automáticas Implementadas
- **NOVO**: `ContentBlockHandler.push()` detecta entity mentions
- **NOVO**: `detectEntityMentions()` e `resolveEntityMention()` helpers
- **NOVO**: `resolveContentBlockHierarchy()` para determinar nível da citation
- **NOVO**: `buildHierarchyContext()` para context string
- **NOVO**: `createCitationRelations()` cria relations no nível correto (Beat > Scene > Chapter)
- Testes unitários em `contentBlockHelpers.test.ts` e `ContentCitationService.test.ts`

---

## Problemas Identificados

### ~~🔴 CRÍTICO - Funcionalidades Incompletas~~ → Maioria Resolvida

#### ~~1. Citations Automáticas em ContentBlocks (Fase 9)~~ ✅ IMPLEMENTADO
**Status**: Completo

Implementação encontrada em:
- `ContentBlockHandler.push()` - Detecta menções e cria citations
- `utils/contentBlockHelpers.ts` - Helpers para hierarquia e criação
- `push/ContentCitationService.ts` - Serviço dedicado para citations

#### ~~2. Script de Migração V1→V2 (Fase 10)~~ ✅ DESCARTADO
**PARECER HUMANO**: Não precisamos migrar vaults de V1 para V2.

#### ~~3. Sistema de Backups (Fase 11)~~ ✅ IMPLEMENTADO
**Status**: Completo (versão simplificada)

Implementação em `backup/BackupManager.ts`:
- Cria snapshots em `.story-engine/backups/{timestamp}/`
- Manifesto JSON com arquivos afetados
- Integrado ao `SyncOrchestrator`

---

### 🟠 ALTO - Problemas de Código

#### ~~4. `StoryHandler` Muito Grande~~ ✅ REFATORADO
**Status**: Resolvido

`StoryHandler.ts` agora é um coordenador slim (~165 linhas) que delega para:
```
handlers/story/
├── StoryHandler.ts           # Coordenador (~165 linhas)
└── services/
    ├── StoryRelationsService.ts   # Geração e push de relations
    ├── StoryConflictService.ts    # Detecção/resolução de conflitos
    ├── StoryRenameService.ts      # Renomeação por reorder
    └── StoryFileService.ts        # I/O de arquivos
```

#### 5. Push Incompleto no StoryHandler - PARCIALMENTE RESOLVIDO
**Arquivo**: `StoryHandler.ts`

O push atual sincroniza:
- ✅ Relations (via `StoryRelationsService.pushRelations()`)
- ✅ Chapter reorders (via `OutlinePushHandler`)
- ❌ Scene/Beat reorders (não usa `PushPlanner` no `StoryHandler.push()`)
- ❌ Content text updates (não usa `content_update` actions)

**Lacuna**: O `PushPlanner` detecta `content_update` mas não está integrado no `StoryHandler.push()`.

#### ~~6. World Handlers sem Push~~ ✅ IMPLEMENTADO
**Status**: Todos os world handlers têm `push()` funcional:
- `CharacterHandler.push()` ✅
- `LocationHandler.push()` ✅
- `FactionHandler.push()` ✅
- `ArtifactHandler.push()` ✅
- `EventHandler.push()` ✅
- `LoreHandler.push()` ✅
- `ArchetypeHandler.push()` ✅
- `TraitHandler.push()` ✅
- `WorldHandler.push()` ✅

#### 7. ConflictResolver sem UI - PENDENTE
**Arquivo**: `conflict/ConflictResolver.ts` linha 260

```typescript
private async resolveManual(conflict: Conflict): Promise<ConflictResolution> {
    // TODO: Show conflict resolution modal
    // For now, default to local version
```

Quando `conflictResolution: "manual"`, não há modal; sempre usa versão local.

#### 8. AutoSync sempre "Online" - PENDENTE
**Arquivo**: `autoSync/AutoSyncManagerV2.ts` linha 432

```typescript
private isOnline(): boolean {
    // TODO: Implement actual network detection
    return true;
}
```

Não há detecção real de conectividade.

---

### 🟡 MÉDIO - Inconsistências

#### ~~9. Estrutura de Pastas Divergente~~ ✅ ACEITO
**Decisão**: Manter implementação com prefixos numéricos (`00-chapters/`, etc.)

#### 10. `TARGET_TYPE_MAP` Hardcoded - PENDENTE
**Arquivo**: `push/RelationsPushHandler.ts`

Mapeamento frágil baseado em strings de seção.

#### 11. Duplicação entre Generators - PENDENTE
Lógica similar de frontmatter/sanitização poderia ser extraída.

---

### 🆕 NOVOS PROBLEMAS IDENTIFICADOS

#### 12. Falta de Testes Unitários para Services do StoryHandler
**Arquivos**: `handlers/story/services/*.ts`

Os 4 novos serviços extraídos não têm testes unitários dedicados:
- `StoryRelationsService.ts` - sem testes
- `StoryConflictService.ts` - sem testes
- `StoryRenameService.ts` - sem testes
- `StoryFileService.ts` - sem testes

**Impacto**: Refatoração futura arriscada; cobertura depende dos testes do `StoryHandler.test.ts`.

#### 13. Push de Contents não Integrado
**Localização**: `StoryHandler.push()` vs `PushPlanner`

O `PushPlanner` detecta `content_update` actions mas o `StoryHandler.push()` não as utiliza. O fluxo de push de texto está incompleto.

**Fluxo esperado**:
1. `PushPlanner.buildPlan()` → detecta `content_update`
2. `PushExecutor.execute()` → chama `apiClient.updateContentBlock()`

**Fluxo atual**:
- `StoryHandler.push()` só faz relations e outline reorders

---

## Análise de Testes

### Cobertura Geral
- **37+ arquivos de teste** encontrados
- Cobertura boa para parsers e generators
- Cobertura razoável para handlers
- **NOVO**: Testes para `contentBlockHelpers.ts` e `ContentCitationService.ts`

### ~~🔴 Problemas nos Testes~~ → Parcialmente Resolvido

#### ~~1. Teste Skipado~~ ✅ CORRIGIDO
**Arquivo**: `StoryHandler.test.ts`

O teste `"preserves untracked segments when reconciling contents"` foi reabilitado após correção no `DiffEngine.findUntrackedSegments()` (strip frontmatter).

#### 2. Mocks Excessivos - PENDENTE
Os testes ainda usam mocks pesados, dificultando detecção de bugs de integração.

#### 3. Falta de Testes de Integração - PENDENTE
Não há pasta `__integration__` ou fixtures de vault real.

#### 4. Cenários de Erro Subexplorados - PENDENTE
Poucos testes para timeout/retry/concorrência.

#### 🆕 5. Serviços do StoryHandler sem Testes Dedicados
Os novos serviços em `handlers/story/services/` não têm arquivos de teste próprios.

### 🟢 Testes Bem Escritos

#### Bons Exemplos:
- `contentsParser.test.ts` - Cobertura abrangente
- `DiffEngine.test.ts` - Testes claros
- `AutoSyncManagerV2.test.ts` - Testes extensivos de timing
- `detectEntityMentions.test.ts` - Excelente cobertura
- **NOVO**: `contentBlockHelpers.test.ts` - Testes para helpers de hierarquia
- **NOVO**: `ContentCitationService.test.ts` - Testes para citations

---

## Divergências do Plano

| Aspecto | Plano | Implementação | Status |
|---------|-------|---------------|--------|
| Estrutura de pastas | `chapters/`, `scenes/` | `00-chapters/`, `01-scenes/` | ✅ Aceito |
| `FileNaming.ts` separado | Linha 863 | Embutido em `PathResolver` | ⚠️ Divergente |
| `FrontmatterParser.ts` | Linha 857 | `parseFrontmatter()` em utils | ⚠️ Divergente |
| Pull World com relations | Fase 7 | ✅ Implementado | ✅ OK |
| Push Relations | Fase 9 | ✅ Implementado | ✅ OK |
| Auto Citations | Fase 9 | ✅ Implementado | ✅ OK |
| Migration Script | Fase 10 | ❌ Não implementado | ✅ Descartado |
| Backups simples | Fase 11 | ✅ Implementado | ✅ OK |
| ConflictResolver Modal | Fase 9 | ❌ Não implementado | 🟠 Pendente |
| StoryHandler refatorado | Fase 6 | ✅ 4 serviços extraídos | ✅ OK |
| World handlers push | Fase 7 | ✅ Todos implementados | ✅ OK |

---

## Ações Recomendadas

### ✅ CONCLUÍDAS (desde última revisão)

| # | Ação | Status |
|---|------|--------|
| 1 | Corrigir teste skipado | ✅ Concluído |
| 2 | World handlers push | ✅ Concluído |
| 4 | Backups simples | ✅ Concluído |
| 5 | Citations automáticas | ✅ Concluído |
| 6 | Refatorar StoryHandler | ✅ Concluído |

---

### Imediatas (Sprint Atual)

#### 1. Testes Unitários para Serviços do StoryHandler
**Prioridade**: ALTA
**Esforço**: 1-2 dias

Criar arquivos de teste para:
```
handlers/story/services/__tests__/
├── StoryRelationsService.test.ts
├── StoryConflictService.test.ts
├── StoryRenameService.test.ts
└── StoryFileService.test.ts
```

#### 2. Integrar Push de Contents no StoryHandler
**Prioridade**: ALTA
**Esforço**: 1 dia

Usar `PushPlanner` e `PushExecutor` no `StoryHandler.push()` para:
- Detectar `content_update` actions
- Chamar `apiClient.updateContentBlock()`

---

### Curto Prazo (1-2 Sprints)

#### 3. UI de Resolução de Conflitos
**Prioridade**: BAIXA
**Esforço**: 3-5 dias

Criar modal para `conflictResolution: "manual"`:
- Diff visual
- Escolher local/remote/merge

#### 4. Detecção Real de Conectividade
**Prioridade**: BAIXA
**Esforço**: 4h

Implementar `isOnline()` no `AutoSyncManagerV2`.

---

### Médio Prazo (1-2 Meses)

#### 5. Testes de Integração
**Prioridade**: MÉDIA
**Esforço**: 1 semana

Criar suite com fixtures de vault real e testes e2e.

---

## Priorização

### Sprint Atual
| # | Ação | Esforço | Impacto |
|---|------|---------|---------|
| 1 | Testes unitários para serviços | 2d | Alto |
| 2 | Integrar push de contents | 1d | Alto |

### Próximo Sprint
| # | Ação | Esforço | Impacto |
|---|------|---------|---------|
| 3 | UI conflitos | 5d | Baixo |
| 4 | Detecção de conectividade | 4h | Baixo |

### Backlog
| # | Ação | Esforço | Impacto |
|---|------|---------|---------|
| 5 | Testes integração | 1sem | Médio |

### ✅ Concluídos
| Ação | Data |
|------|------|
| Corrigir teste skipado | 2026-01-16 |
| World handlers push | 2026-01-16 |
| Backups simples | 2026-01-16 |
| Citations automáticas | 2026-01-16 |
| Refatorar StoryHandler | 2026-01-16 |

### ✅ Descartados
| Ação | Motivo |
|------|--------|
| Script migração V1→V2 | PARECER HUMANO: Não necessário |

---

## Conclusão

A implementação do Sync V2 está **85-90% completa** em relação ao plano. Progresso significativo desde a última revisão:

### ✅ Fundamentos Sólidos
- Arquitetura modular com serviços bem separados
- Pull funcional para story e world
- Push funcional para relations, outline reorders e world entities
- Citations automáticas em ContentBlocks
- Sistema de backup simples
- StoryHandler refatorado em 4 serviços

### Lacunas Restantes
1. **Push de contents não integrado** - `content_update` detectado mas não executado
2. **Testes unitários para serviços** - Novos serviços sem testes dedicados
3. **ConflictResolver UI** - Usa versão local por padrão
4. **Detecção de conectividade** - Sempre assume online

### Próximos Passos Prioritários
1. Criar testes unitários para os 4 serviços do StoryHandler
2. Integrar push de contents usando `PushPlanner`/`PushExecutor`
3. (Opcional) UI de conflitos e detecção de rede

O sync bidirecional está quase completo - falta apenas a parte de push de texto de ContentBlocks estar integrada no fluxo principal.
