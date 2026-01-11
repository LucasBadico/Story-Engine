# Sync V2 Architecture Plan

> **Status**: Planning
> **Date**: 2025-01-09
> **Goal**: Refatorar completamente o sistema de sync do plugin Obsidian para ser modular, extensível e suportar World entities.

## Motivação

O sistema atual (`syncService.ts` com ~3370 linhas) se tornou difícil de manter:
- Monolítico - toda lógica em um arquivo
- Sem suporte para World entities (characters, locations, etc.)
- Parsing, file generation e sync misturados
- Estrutura de pastas rígida e hardcoded
- Duplicação de lógica entre entity types

## Objetivos

1. **Modularidade** - Um handler por entity type
2. **Extensibilidade** - Fácil adicionar novos entity types
3. **Separação de concerns** - Outline, Contents e Relations em arquivos separados
4. **World support** - Suporte completo para entities de World
5. **File naming inteligente** - Usar `order_num` + `title/intent` para identificação clara
6. **Content blocks multi-parágrafo** - Suporte a conteúdo rico com fences HTML

## Estratégia de Versões

- **Sync V1 (legado)** continua sendo o padrão (`syncVersion = "v1"`) até que o novo pipeline esteja estável. Todo o código atual em `syncService.ts` permanece operacional e é o caminho seguro para quem não quer mudanças ainda.
- **Sync V2 (modular)** fica atrás de um feature flag configurável no plugin do Obsidian (`Config → Sync Version`). Quando o usuário escolhe `v2`, o plugin passa a instanciar os novos handlers/geradores definidos neste documento.
- O `StoryEngineClient` e o restante do app precisam trabalhar em ambos os modos; somente a implementação interna de sync é trocada.
- O rollout prevê que V2 possa ser ativado por vault/workspace, permitindo comparar comportamento antes de migrar definitivamente.

### Plano de fallback

1. Carregar configurações → se `syncVersion === "v1"`, inicializar apenas serviços legados.
2. Se `syncVersion === "v2"`, inicializar toda a stack modular, mas manter V1 carregável para rollback rápido (sem precisar reinstalar o plugin).
3. Caso V2 encontre erro fatal durante sync, reverter automaticamente para V1 e notificar o usuário, mantendo arquivos intactos.

### Estado atual da implementação

- `SyncEngine` agora é uma interface única consumida pelo plugin/auto-sync. A versão legada (`SyncService`) implementa essa interface sem mudanças de comportamento.
- O factory `createSyncEngine` decide em tempo de execução qual engine carregar com base na configuração (`Sync Version` em Settings).
- O namespace `sync-v2/` já contém:
  - `types/sync.ts`: contratos de operação (`SyncOperation`, `SyncContext`, `SyncResult`) para padronizar handlers.
  - `core/SyncOrchestrator.ts`: ponto central que receberá fila de operações (`pull_story`, `push_story`, etc). Por enquanto, apenas captura payloads e retorna `not_implemented`, mas já consulta a API para validar dados.
  - `core/ModularSyncEngine.ts`: implementação concreta da interface `SyncEngine` que delega ao orchestrator e comunica status ao usuário.
- Com isso, conseguimos ativar o V2 via configurador, sem quebrar o vault: o usuário recebe avisos de que o pipeline ainda está em construção e pode voltar ao V1 instantaneamente.

## Garantias de Autoria Bidirecional

Para o Sync V2 os arquivos não serão apenas "preenchidos" pela API. Sempre que o escritor editar outline/contents/relations localmente, o sistema deve refletir no backend automaticamente.

- **Placeholders inteligentes** já existentes continuam sendo o ponto de entrada. Quando um placeholder é alterado (outline list item, fence de conteúdo etc.), o parser marca como `isModifiedPlaceholder` e agenda a criação da entidade via API.
- **Debounce e batching** reutilizam as regras atuais do `AutoSyncManager`: mudanças disparam um push somente após o tempo configurado (ex.: 1s typing pause / 5s idle / blur). Isso evita flood de requests enquanto o escritor digita.
- **Listas editáveis** (Outline e Relations): adicionar linhas novas cria operações `createEntity` ou `createRelation`. Remover linhas gera `delete` e mover linhas gera `reorder`.
- **Conteúdo textual** dentro das fences: editar texto continua gerando `updateContentBlock` com detecção de diffs multi-parágrafo.
- **Experiência offline**: no modo `local`, as criações ficam em fila até reconexão; no modo `remote`, o push é feito imediatamente após debounce.
- **Observabilidade**: cada operação bidirecional gera log estruturado para inspeção e facilita comparar comportamento V1 vs V2.

---

## Backend: EntityRelation

O sistema de sync utiliza a tabela `entity_relations` do backend para rastrear relações e citações entre entidades.

### Estrutura da EntityRelation

```protobuf
message EntityRelation {
  string id = 1;
  string tenant_id = 2;
  string world_id = 3;
  
  // Source: quem está citando/referenciando
  string source_type = 4;      // "scene", "beat", "content_block", "chapter", etc
  string source_id = 5;        // ID da entidade que cita
  
  // Target: quem está sendo citado/referenciado
  string target_type = 6;      // "character", "location", "artifact", "faction", etc
  string target_id = 7;        // ID da entidade citada
  
  // Tipo da relação
  string relation_type = 8;    // "citation", "pov", "setting", "owns", "member_of", etc
  
  // Contexto opcional (onde a relação ocorre)
  optional string context_type = 9;   // "chapter", "story", etc
  optional string context_id = 10;
  
  // Metadados flexíveis
  string attributes_json = 11; // JSON para dados adicionais
  string summary = 12;         // Resumo/descrição da relação
  
  // Para relações bidirecionais
  optional string mirror_id = 13;     // ID da relação inversa
  
  optional string created_by_user_id = 14;
  google.protobuf.Timestamp created_at = 15;
  google.protobuf.Timestamp updated_at = 16;
}
```

### Tipos de Relação (`relation_type`)

| relation_type | Direção | Exemplo | Uso |
|---------------|---------|---------|-----|
| `citation` | story → world | Scene cita Character | `.citations.md` |
| `pov` | scene → character | Scene tem POV de Character | `.relations.md` |
| `setting` | scene → location | Scene acontece em Location | `.relations.md` |
| `owns` | character → artifact | Character possui Artifact | `.relations.md` |
| `member_of` | character → faction | Character é membro de Faction | `.relations.md` |
| `ally` | character → character | Characters são aliados | `.relations.md` |
| `enemy` | character → character | Characters são inimigos | `.relations.md` |
| `located_in` | artifact → location | Artifact está em Location | `.relations.md` |
| `caused` | event → event | Event causou outro Event | `.relations.md` |

### Endpoints para Sync

**Para gerar `.citations.md`** (onde uma entity é citada):
```
ListRelationsByTarget(
  target_type = "character",
  target_id = character_id,
  relation_type = "citation"  // filtro opcional
)
```

**Para gerar `.relations.md`** (o que uma entity referencia):
```
ListRelationsBySource(
  source_type = "scene",
  source_id = scene_id
)
```
nota: no retorno excluir tipos citations.

**Por World** (todas as relações de um mundo):
```
ListRelationsByWorld(
  world_id = world_id,
  relation_type = "citation"  // filtro opcional
)
```

### Diferença: ContentAnchor vs EntityRelation

| Aspecto | ContentAnchor | EntityRelation |
|---------|---------------|----------------|
| **Propósito** | Onde um ContentBlock está pendurado | Citação/referência entre entities |
| **Direção** | ContentBlock → Entity (parent) | Entity → Entity |
| **Exemplo** | "Este texto pertence ao Beat X posicao 1" | "Scene Y cita Character Z" |
| **Uso no Sync** | Organizar contents por hierarquia | Gerar `.relations.md` e `.citations.md` |

---

## Estrutura de Arquivos no Vault

### Stories

```
StoryEngine/
├── stories/
│   └── {story-title}/
│       ├── story.outline.md          # Hierarquia: chapters > scenes > beats
│       ├── story.contents.md         # Content blocks por entity
│       ├── story.relations.md        # Relações com World entities
│       │
│       ├── chapters/
│       │   ├── ch-01-{title}.outline.md    # scenes > beats deste chapter
│       │   ├── ch-01-{title}.contents.md   # prose deste chapter
│       │   ├── ch-01-{title}.relations.md  # relações deste chapter
│       │   └── ...
│       │
│       ├── scenes/
│       │   ├── sc-01-{goal}.md             # detalhes + beats inline
│       │   ├── sc-02-{goal}.md
│       │   └── ...
│       │
│       ├── beats/
│       │   ├── bt-01-{intent}.md           # detalhes do beat
│       │   ├── bt-02-{intent}.md
│       │   └── ...
│       │
│       └── contents/
│           ├── texts/
│           │   └── {id}.md
│           ├── images/
│           │   └── {id}.md
│           └── ...
```

### Worlds

```
StoryEngine/
├── worlds/
│   └── {world-name}/
│       ├── world.outline.md          # Overview de todas entities
│       ├── world.contents.md         # Descrições e lore geral
│       ├── world.relations.md        # Relações entre entities do world
│       ├── world.citations.md        # Onde o world é usado nas stories
│       │
│       ├── characters/
│       │   ├── {name}.md             # Dados básicos
│       │   ├── {name}.contents.md    # Descrição detalhada, backstory
│       │   ├── {name}.relations.md   # Relações com outros characters, factions
│       │   ├── {name}.citations.md   # Scenes/beats onde aparece
│       │   └── ...
│       │
│       ├── locations/
│       │   ├── {name}.md             # Dados básicos (suporta hierarquia)
│       │   ├── {name}.contents.md    # Descrição detalhada
│       │   ├── {name}.relations.md   # Characters, artifacts, events ligados
│       │   ├── {name}.citations.md   # Scenes onde é cenário
│       │   └── ...
│       │
│       ├── factions/
│       │   ├── {name}.md             # Dados básicos (suporta hierarquia)
│       │   ├── {name}.contents.md    # História, crenças, estrutura
│       │   ├── {name}.relations.md   # Members, allies, enemies
│       │   ├── {name}.citations.md   # Onde é mencionada
│       │   └── ...
│       │
│       ├── artifacts/
│       │   ├── {name}.md
│       │   ├── {name}.contents.md    # Lore, história do artefato
│       │   ├── {name}.relations.md   # Owners, creators, locations
│       │   ├── {name}.citations.md   # Onde aparece nas histórias
│       │   └── ...
│       │
│       ├── events/
│       │   ├── {name}.md
│       │   ├── {name}.contents.md    # Descrição detalhada do evento
│       │   ├── {name}.relations.md   # Participants, locations, consequences
│       │   ├── {name}.citations.md   # Referências nas histórias
│       │   └── ...
│       │
│       └── lore/
│           ├── {name}.md             # Dados básicos (suporta hierarquia)
│           ├── {name}.contents.md    # Regras, explicações detalhadas
│           ├── {name}.relations.md   # Entidades que usam/seguem este lore
│           ├── {name}.citations.md   # Onde é aplicado/mencionado
│           └── ...
```

---

## Convenções de Nomenclatura

### Scenes e Beats

**Formato**: `{prefix}-{order_num:04d}-{sanitized_title}.md`

| Entity | Prefix | Exemplo |
|--------|--------|---------|
| Chapter | `ch` | `ch-01-the-beginning.md` |
| Scene | `sc` | `sc-01-meet-the-hero.md` |
| Beat | `bt` | `bt-01-introduction.md` |

**Regras**:
- `order_num` sempre com 4 dígitos (0001, 0002, ..., 999)
- Título sanitizado (lowercase, hífens, sem caracteres especiais)
- Máximo 40 caracteres no título
- **Quando `order_num` muda**: Renomear arquivo, manter `id` no frontmatter

**Exemplo de rename**:
```
# Antes: sc-03-departure.md
# Usuário move scene para posição 1
# Depois: sc-01-departure.md (mesmo ID no frontmatter)
```

### World Entities

**Formato**: `{sanitized_name}.md`

| Entity | Exemplo |
|--------|---------|
| Character | `john-smith.md` |
| Location | `crystal-mountains.md` |
| Faction | `order-of-light.md` |

---

## Formato dos Arquivos

### 1. Outline Files (`.outline.md`)

Mostram hierarquia editável de entities. **Editável pelo escritor**.

```markdown
---
id: {entity-id}
type: story-outline
synced_at: 2025-01-09T10:00:00Z
---

# {Story Title}

## Hierarchy

> [!tip] Como editar esta lista
> - **Reordenar**: Arraste itens para mudar a ordem
> - **Criar novo**: Edite a linha `_New..._` no final de cada seção
> - **Indentação**: Tab define hierarquia (chapter → scene → beat)
> - **Marcadores**: `+` tem conteúdo, `-` está vazio

- [[ch-01-the-beginning|Chapter 1: The Beginning]] +
	- [[sc-01-meet-hero|Scene 1: Meet the hero - Morning]] +
		- [[bt-01-intro|Beat 1: Introduction]] +
		- [[bt-02-conflict|Beat 2: First conflict]] -
		- _New beat: intent here..._
	- [[sc-02-call|Scene 2: The call - Afternoon]] -
		- _New beat: intent here..._
	- _New scene: goal - time_
- [[ch-02-journey|Chapter 2: The Journey]] -
	- _New scene: goal - time_
- _New chapter: title_
```

**Chapter outline** (`ch-01-title.outline.md`):
```markdown
---
id: {chapter-id}
type: chapter-outline
synced_at: 2025-01-09T10:00:00Z
---

# Chapter 1: The Beginning

## Scenes & Beats

> [!tip] Como editar esta lista
> - **Reordenar**: Arraste itens para mudar a ordem
> - **Criar novo**: Edite a linha `_New..._` no final de cada seção
> - **Indentação**: Tab = beat dentro de scene
> - **Marcadores**: `+` tem conteúdo, `-` está vazio

- [[sc-01-meet-hero|Scene 1: Meet the hero - Morning]] +
	- [[bt-01-intro|Beat 1: Introduction]] +
	- [[bt-02-conflict|Beat 2: First conflict]] -
	- _New beat: intent here..._
- [[sc-02-call|Scene 2: The call - Afternoon]] -
	- _New beat: intent here..._
- _New scene: goal - time_
```

### 2. Contents Files (`.contents.md`)

Mostram content blocks associados a cada entity usando **fences HTML**. 
Scenes e Beats também usam fences para hierarquia clara e editável.

**Formato da fence**:
```
<!--{type}-start:{order}:{name}:{id}-->
<!--{type}-end:{order}:{name}:{id}-->
```

| Campo | Formato | Descrição |
|-------|---------|-----------|
| `type` | string | `chapter`, `scene`, `beat`, `content` |
| `order` | 4 dígitos | Ordem da entity (ex: `0001`, `0042`) |
| `name` | kebab-case | Nome/título sanitizado para identificação |
| `id` | uuid | ID único da entity no backend |

**Reordenação**: Se uma fence mudar de posição no arquivo, o sync interpreta como mudança de `order_num`.

**Placeholders para novos itens**:
O arquivo sempre contém placeholders que mostram como adicionar novos itens. 
Estes são ignorados pelo sync até serem modificados pelo escritor.

```markdown
<!--content-start:0000:new-content:placeholder-->
_Write your content here..._
<!--content-end:0000:new-content:placeholder-->
```

| Identificador | Descrição |
|---------------|-----------|
| `order: 0000` | Ordem zero indica placeholder |
| `name: new-*` | Prefixo "new-" indica placeholder |
| `id: placeholder` | ID "placeholder" é sempre ignorado |

> 💡 **Atalho para o escritor**: comentários simples como `<!--new-content-->`, `<!--new-scene-->`, `<!--new-beat-->` e `<!--new-chapter-->` também são aceitos. O parser converte automaticamente esses marcadores em fences completas de placeholder, então basta inserir o comentário onde desejar criar um item novo.

Quando o escritor modifica o placeholder:
1. Parser detecta que conteúdo mudou (não é mais `_Write your content here..._`)
2. Sync cria nova entity no backend
3. Fence é atualizada com ID real e order calculado
4. Novo placeholder é adicionado no final

```markdown
---
id: {entity-id}
type: story-contents
synced_at: 2025-01-09T10:00:00Z
---

# {Story Title} - Contents

<!--chapter-start:0001:the-beginning:ch-uuid-001-->
## Chapter 1: The Beginning

<!--scene-start:0001:meet-the-hero:sc-uuid-001-->
### Scene 1: Meet the hero

<!--content-start:0001:the-sun-rose:cb-uuid-001-->
The sun rose over the mountains, casting long shadows across the valley.

John stood at the edge of the cliff, his heart pounding with anticipation.
This was the moment he had been waiting for.
<!--content-end:0001:the-sun-rose:cb-uuid-001-->

<!--content-start:0002:are-you-ready:cb-uuid-002-->
"Are you ready?" asked Maria, her voice barely a whisper.
<!--content-end:0002:are-you-ready:cb-uuid-002-->

<!--beat-start:0001:introduction:bt-uuid-001-->
#### Beat 1: Introduction

<!--content-start:0001:the-village-bell:cb-uuid-003-->
The village bell rang three times, signaling the start of the ceremony.

Everyone gathered in the square, their eyes fixed on the ancient tower.
<!--content-end:0001:the-village-bell:cb-uuid-003-->
<!--beat-end:0001:introduction:bt-uuid-001-->

<!--beat-start:0002:first-conflict:bt-uuid-002-->
#### Beat 2: First conflict

<!--content-start:0000:new-content:placeholder-->
_Write your content here..._
<!--content-end:0000:new-content:placeholder-->

<!--beat-end:0002:first-conflict:bt-uuid-002-->

<!--beat-start:0000:new-beat:placeholder-->
#### _New Beat Intent_

<!--content-start:0000:new-content:placeholder-->
_Write your content here..._
<!--content-end:0000:new-content:placeholder-->

<!--beat-end:0000:new-beat:placeholder-->

<!--scene-end:0001:meet-the-hero:sc-uuid-001-->

<!--scene-start:0002:the-call:sc-uuid-002-->
### Scene 2: The call

*No content blocks yet*
<!--scene-end:0002:the-call:sc-uuid-002-->

<!--chapter-end:0001:the-beginning:ch-uuid-001-->

---

<!--chapter-start:0002:the-journey:ch-uuid-002-->
## Chapter 2: The Journey

<!--scene-start:0000:new-scene:placeholder-->
### _New Scene Title_

_Describe what happens in this scene..._
<!--scene-end:0000:new-scene:placeholder-->

<!--chapter-end:0002:the-journey:ch-uuid-002-->

<!--chapter-start:0000:new-chapter:placeholder-->
## _New Chapter Title_

<!--scene-start:0000:new-scene:placeholder-->
### _New Scene Title_

_Describe what happens in this scene..._
<!--scene-end:0000:new-scene:placeholder-->

<!--chapter-end:0000:new-chapter:placeholder-->
```

**Tipos de fences HTML**:
| Fence | Exemplo |
|-------|---------|
| chapter | `<!--chapter-start:0001:the-beginning:ch-uuid-->` |
| scene | `<!--scene-start:0001:meet-hero:sc-uuid-->` |
| beat | `<!--beat-start:0001:intro:bt-uuid-->` |
| content | `<!--content-start:0001:the-sun:cb-uuid-->` |

**Benefícios das fences HTML**:
- Suporta múltiplos parágrafos
- Suporta markdown formatting dentro
- Hierarquia clara e aninhável
- Fácil de parsear (regex)
- Não interfere com syntax do Obsidian
- **Identificação visual**: ordem + nome legível para o escritor
- **Sync bidirecional**: ID garante rastreamento mesmo com rename
- **Reordenação natural**: mover fence = mudar order_num

### 3. Relations Files (`.relations.md`)

Mostram relações com World entities. **Editável pelo escritor**. 
Sincronizado via `ListRelationsBySource` da API `EntityRelationService`.

**Story relations** (`story.relations.md`):
```markdown
---
id: {story-id}
type: story-relations
synced_at: 2025-01-09T10:00:00Z
world_id: {world-id}
---

# {Story Title} - Relations

> [!tip] Como editar relações
> - **Adicionar**: Escreva na linha `_Add new..._` da seção apropriada
> - **Remover**: Delete a linha da relação
> - **Formato**: `[[entity-file|Entity Name]] - description`
> - Relações são sincronizadas como `EntityRelation` no backend

## World
[[{world-name}|{World Name}]]

## Main Characters
- [[john-smith|John Smith]] - Protagonist
- [[maria-santos|Maria Santos]] - Mentor
- _Add new character: [[file|Name]] - role_

## Key Locations
- [[crystal-mountains|Crystal Mountains]] - Starting area
- [[dark-forest|Dark Forest]] - Chapter 2
- _Add new location: [[file|Name]] - context_

## Referenced Factions
- [[order-of-light|Order of Light]]
- _Add new faction: [[file|Name]] - role_

## Timeline Events
- [[great-war|The Great War]] - Background
- _Add new event: [[file|Name]] - context_

## Artifacts
- _Add new artifact: [[file|Name]] - context_
```

**Chapter relations** (`ch-01-title.relations.md`):
```markdown
---
id: {chapter-id}
type: chapter-relations
synced_at: 2025-01-09T10:00:00Z
---

# Chapter 1: The Beginning - Relations

> [!tip] Como editar relações
> - **Adicionar**: Escreva na linha `_Add new..._` da seção apropriada
> - **Remover**: Delete a linha da relação
> - **Formato POV**: `[[character|Name]] - Scene X, Y`
> - **Formato Location**: Sub-location após `:`

## POV Characters
- [[john-smith|John Smith]] - Scenes 1, 2
- [[maria-santos|Maria Santos]] - Scene 3
- _Add POV: [[character|Name]] - Scene N_

## Locations
- [[crystal-mountains|Crystal Mountains]]
  - Scene 1: cliff_edge
  - Scene 2: village_square
- _Add location: [[location|Name]] - Scene N: sub_location_

## Characters Appearing
- [[john-smith|John Smith]]
- [[maria-santos|Maria Santos]]
- [[elder-kai|Elder Kai]]
- _Add character: [[character|Name]]_

## Artifacts Mentioned
- [[sword-of-light|Sword of Light]]
- _Add artifact: [[artifact|Name]]_
```

### 4. Citations Files (`.citations.md`)

Mostram onde uma World entity é citada nas histórias. **Auto-gerado, NÃO EDITAR**.
Gerado via `ListRelationsByTarget` da API `EntityRelationService`.

**Character citations** (`john-smith.citations.md`):
```markdown
---
id: {character-id}
type: character-citations
synced_at: 2025-01-09T10:00:00Z
---

# John Smith - Citations

> [!warning] ⚠️ Arquivo auto-gerado - NÃO EDITE
> Este arquivo é atualizado automaticamente durante o sync.
> Suas edições serão sobrescritas na próxima sincronização.
> 
> **Para adicionar citações**: Referencie este character em arquivos `.relations.md` 
> de chapters, scenes ou beats.

> [!info] Como funciona
> Gerado via `ListRelationsByTarget(target_type="character", target_id=this.id)`.
> Mostra todas as relações onde este Character é o **target** (está sendo referenciado).

## Stories

### [[the-great-adventure|The Great Adventure]]

#### POV (`relation_type: pov`)
- [[sc-01-meet-hero|Scene 1: Meet the hero]] (Chapter 1)
- [[sc-02-call|Scene 2: The call]] (Chapter 1)
- [[sc-05-revelation|Scene 5: The revelation]] (Chapter 2)

#### Citations (`relation_type: citation`)
- [[bt-03-dialogue|Beat 3: Dialogue with mentor]] (Scene 2)
  - *"John hesitated before speaking..."*
- [[bt-07-memory|Beat 7: Memory flashback]] (Scene 4)
  - *"Memories of John's childhood..."*
- [[cb-001|Content Block]] (Scene 1, Beat 1)
  - *"John stood at the edge..."*
- [[cb-015|Content Block]] (Scene 3, Beat 2)
  - *"The sword belonged to John's father..."*

### [[prequel-story|The Prequel]]

#### POV (`relation_type: pov`)
- [[sc-01-childhood|Scene 1: Childhood]] (Chapter 1)

---

## Summary

| Story | relation_type | Count |
|-------|---------------|-------|
| The Great Adventure | pov | 3 |
| The Great Adventure | citation | 4 |
| The Prequel | pov | 1 |
| **Total** | | **8** |
```

**Location citations** (`crystal-mountains.citations.md`):
```markdown
---
id: {location-id}
type: location-citations
synced_at: 2025-01-09T10:00:00Z
---

# Crystal Mountains - Citations

> [!warning] ⚠️ Arquivo auto-gerado - NÃO EDITE
> Este arquivo é atualizado automaticamente durante o sync.
> 
> **Para adicionar citações**: Use este location como `setting` em scenes,
> ou referencie em arquivos `.relations.md`.

## Stories

### [[the-great-adventure|The Great Adventure]]

#### Setting (`relation_type: setting`)
- [[sc-01-meet-hero|Scene 1: Meet the hero]]
  - *attributes: { sub_location: "cliff_edge" }*
- [[sc-03-training|Scene 3: Training]]
  - *attributes: { sub_location: "mountain_peak" }*
- [[sc-08-return|Scene 8: The return]]
  - *attributes: { sub_location: "village_entrance" }*

#### Citations (`relation_type: citation`)
- [[bt-05-description|Beat 5: Landscape description]]
  - *"The mountains loomed in the distance..."*
- [[cb-042|Content Block]]
  - *"The Crystal Mountains gleamed..."*

---

## Summary

| relation_type | Count |
|---------------|-------|
| setting | 3 |
| citation | 2 |
| **Total** | **5** |
```

**Artifact citations** (`sword-of-light.citations.md`):
```markdown
---
id: {artifact-id}
type: artifact-citations
synced_at: 2025-01-09T10:00:00Z
---

# Sword of Light - Citations

> [!warning] ⚠️ Arquivo auto-gerado - NÃO EDITE
> Este arquivo é atualizado automaticamente durante o sync.
> 
> **Para adicionar citações**: Referencie este artifact em arquivos `.relations.md`
> de scenes, chapters ou characters.

## Stories

### [[the-great-adventure|The Great Adventure]]

#### Appears In (`relation_type: appears_in`)
- [[sc-03-training|Scene 3: Training]]
  - *summary: "First appearance"*
- [[sc-10-final-battle|Scene 10: Final battle]]
  - *summary: "Key role in climax"*

#### Citations (`relation_type: citation`)
- [[cb-015|Content Block]]
  - *"The sword belonged to John's father..."*
- [[cb-089|Content Block]]
  - *"As John raised the Sword of Light..."*
- [[cb-102|Content Block]]
  - *"The blade glowed with ancient power..."*

---

## Summary

| relation_type | Count |
|---------------|-------|
| appears_in | 2 |
| citation | 3 |
| **Total** | **5** |
```

### 4. Entity Files

**Scene file** (`sc-01-meet-hero.md`):
```markdown
---
id: {scene-id}
story_id: {story-id}
chapter_id: {chapter-id}
order_num: 1
goal: Meet the hero
time_ref: Morning
pov_character_id: {character-id}
location_id: {location-id}
synced_at: 2025-01-09T10:00:00Z
tags:
  - story-engine/scene
  - story/{story-name}
---

# Scene 1: Meet the hero

**Goal**: Meet the hero
**Time**: Morning
**POV**: [[john-smith|John Smith]]
**Location**: [[crystal-mountains|Crystal Mountains - Cliff Edge]]

## Beats

- [[bt-01-intro|Beat 1: Introduction]] +
- [[bt-02-conflict|Beat 2: First conflict]] -

## Notes

_Espaço para anotações do autor_
```

**Beat file** (`bt-01-intro.md`):
```markdown
---
id: {beat-id}
scene_id: {scene-id}
order_num: 1
type: exposition
intent: Introduce protagonist
outcome: Reader understands John's motivation
synced_at: 2025-01-09T10:00:00Z
tags:
  - story-engine/beat
  - story/{story-name}
---

# Beat 1: Introduction

**Type**: Exposition
**Intent**: Introduce protagonist
**Outcome**: Reader understands John's motivation

## Content

<!-- content-start:cb-uuid-003 -->
The village bell rang three times...
<!-- content-end:cb-uuid-003 -->

## Notes

_Espaço para anotações_
```

---

## Estrutura de Código

```
obsidian-plugin/src/sync/
├── index.ts                    # Re-exports públicos
│
├── types/
│   ├── sync.ts                 # SyncOperation, SyncResult, SyncContext
│   ├── entity.ts               # EntityType, EntityRef, EntityPayload
│   ├── file.ts                 # FileSpec, ParsedFile, FileContent
│   ├── relations.ts            # RelationType, RelationRef
│   └── citations.ts            # EntityCitation, CitationContext
│
├── core/
│   ├── SyncOrchestrator.ts     # Coordena pull/push de múltiplas entities
│   ├── EntityRegistry.ts       # Registry de handlers por tipo
│   ├── DiffEngine.ts           # Compara estado local vs remote
│   └── FileRenamer.ts          # Gerencia renomeação quando order muda
│
├── handlers/
│   ├── base/
│   │   ├── EntityHandler.ts    # Interface base
│   │   └── HierarchicalHandler.ts  # Para entities com children
│   │
│   ├── story/
│   │   ├── StoryHandler.ts
│   │   ├── ChapterHandler.ts
│   │   ├── SceneHandler.ts
│   │   └── BeatHandler.ts
│   │
│   ├── world/
│   │   ├── WorldHandler.ts
│   │   ├── CharacterHandler.ts
│   │   ├── LocationHandler.ts
│   │   ├── FactionHandler.ts
│   │   ├── ArtifactHandler.ts
│   │   ├── EventHandler.ts
│   │   └── LoreHandler.ts
│   │
│   └── content/
│       └── ContentBlockHandler.ts
│
├── generators/
│   ├── OutlineGenerator.ts     # Gera .outline.md
│   ├── ContentsGenerator.ts    # Gera .contents.md com fences
│   ├── RelationsGenerator.ts   # Gera .relations.md
│   ├── CitationsGenerator.ts   # Gera .citations.md (auto-generated)
│   └── FrontmatterGenerator.ts # Gera frontmatter YAML
│
├── parsers/
│   ├── OutlineParser.ts        # Parse .outline.md
│   ├── ContentsParser.ts       # Parse fences <!-- content-start:id -->
│   ├── RelationsParser.ts      # Parse .relations.md
│   ├── FrontmatterParser.ts    # Parse frontmatter YAML
│   └── EntityFileParser.ts     # Parse entity .md files
│
├── files/
│   ├── FileManager.ts          # CRUD de arquivos simplificado
│   ├── PathResolver.ts         # Resolve paths por entity type
│   └── FileNaming.ts           # Gera/parseia nomes de arquivo
│
└── auto/
    ├── AutoSyncManager.ts      # Detecta mudanças em arquivos
    ├── ApiUpdateNotifier.ts    # Recebe notificações da API
    └── ConflictResolver.ts     # Resolve conflitos local/remote
```

---

## Interfaces Principais

### EntityHandler

```typescript
interface EntityHandler<T extends BaseEntity> {
  readonly entityType: EntityType;
  readonly filePrefix: string;  // 'ch', 'sc', 'bt', etc.
  
  // === Sync Operations ===
  pull(id: string, context: SyncContext): Promise<T>;
  push(entity: T, context: SyncContext): Promise<void>;
  delete(id: string, context: SyncContext): Promise<void>;
  
  // === File Operations ===
  getFilePath(entity: T, context: SyncContext): string;
  generateFileName(entity: T): string;
  parseFileName(fileName: string): { orderNum: number; title: string } | null;
  
  // === Content Generation ===
  generateFileContent(entity: T, context: SyncContext): string;
  parseFileContent(content: string, frontmatter: Record<string, any>): Partial<T>;
  
  // === Outline Support ===
  generateOutlineEntry(entity: T, depth: number): string;
  parseOutlineEntry(line: string, depth: number): OutlineEntryData | null;
  
  // === Relations Support (optional) ===
  getRelations?(entity: T): Promise<EntityRelation[]>;
  generateRelationsSection?(entity: T, relations: EntityRelation[]): string;
  
  // === Citations Support (for World entities) ===
  getCitations?(entity: T): Promise<EntityCitation[]>;
  generateCitationsFile?(entity: T, citations: EntityCitation[]): string;
  
  // === Children (for hierarchical) ===
  getChildren?(parentId: string): Promise<EntityRef[]>;
  getChildHandler?(childType: EntityType): EntityHandler<any>;
}
```

### ContentsParser

```typescript
type FenceType = 'chapter' | 'scene' | 'beat' | 'content';

interface ParsedFence {
  type: FenceType;
  id: string;
  order: number;        // Ordem numérica (ex: 1, 2, 42)
  name: string;         // Nome legível em kebab-case
  content: string;      // Conteúdo dentro da fence (pode incluir fences aninhadas)
  innerText: string;    // Apenas texto, sem fences aninhadas
  startLine: number;
  endLine: number;
  positionInFile: number;  // Posição absoluta para detectar reordenação
  children: ParsedFence[];  // Fences aninhadas
}

interface FenceChange {
  id: string;
  type: FenceType;
  changeType: 'created' | 'updated' | 'deleted' | 'moved' | 'reordered';
  oldOrder?: number;
  newOrder?: number;
  oldParentId?: string;
  newParentId?: string;
}

interface HierarchicalContent {
  chapters: ParsedFence[];
  orphanScenes: ParsedFence[];   // Scenes sem chapter
  orphanBeats: ParsedFence[];    // Beats sem scene
  orphanContents: ParsedFence[]; // Contents soltos
}

interface ContentsParser {
  // Parse all fences hierarchically
  parseHierarchy(content: string): HierarchicalContent;
  
  // Parse fences of a specific type
  parseFencesByType(content: string, type: FenceType): ParsedFence[];
  
  // Detect changes between old and new content
  detectChanges(oldContent: string, newContent: string): FenceChange[];
  
  // Generate fence tags
  generateFenceStart(type: FenceType, order: number, name: string, id: string): string;
  generateFenceEnd(type: FenceType, order: number, name: string, id: string): string;
  
  // Generate complete fence with content
  generateFence(
    type: FenceType, 
    order: number, 
    name: string, 
    id: string, 
    innerContent: string
  ): string;
  
  // Update fence content (keeps order/name/id)
  updateFenceContent(content: string, id: string, newInnerContent: string): string;
  
  // Update fence metadata (order, name)
  updateFenceMeta(content: string, id: string, newOrder: number, newName: string): string;
  
  // Remove a fence (and its contents)
  removeFence(content: string, id: string): string;
  
  // Recalculate all orders based on position in file
  recalculateOrders(content: string, type: FenceType): string;
  
  // Sanitize name for fence (kebab-case, max 30 chars)
  sanitizeName(title: string): string;
  
  // === Placeholder handling ===
  
  // Check if fence is a placeholder
  isPlaceholder(fence: ParsedFence): boolean;
  // Returns true if: order === 0 || name.startsWith('new-') || id === 'placeholder'
  
  // Check if placeholder was modified (ready to become real entity)
  isModifiedPlaceholder(fence: ParsedFence): boolean;
  // Returns true if: isPlaceholder && content !== default placeholder content
  
  // Generate placeholder for a fence type
  generatePlaceholder(type: FenceType, parentId?: string): string;
  
  // Ensure file has placeholders at appropriate locations
  ensurePlaceholders(content: string): string;
  // Adds missing placeholders at end of each parent fence
  
  // Replace placeholder with real fence after entity creation
  replacePlaceholder(
    content: string, 
    placeholderPosition: number,
    realFence: { type: FenceType; order: number; name: string; id: string; content: string }
  ): string;
}

// Default placeholder contents
const PLACEHOLDER_DEFAULTS: Record<FenceType, string> = {
  chapter: '## _New Chapter Title_',
  scene: '### _New Scene Title_\n\n_Describe what happens in this scene..._',
  beat: '#### _New Beat Intent_',
  content: '_Write your content here..._'
};
```

**Detecção de reordenação**:
```typescript
// Quando o escritor move uma fence no arquivo:
// 1. Parser detecta que posição mudou
// 2. Compara com order atual
// 3. Se posição != order esperada → gera FenceChange com changeType: 'reordered'
// 4. Sync atualiza order_num no backend
```

### CitationsGenerator

Usa `EntityRelation` com `ListRelationsByTarget` para encontrar onde uma World entity é citada.

```typescript
// Mapeado de EntityRelation do backend
interface EntityRelationDTO {
  id: string;
  tenantId: string;
  worldId: string;
  sourceType: string;      // "scene", "beat", "content_block", "chapter"
  sourceId: string;
  targetType: string;      // "character", "location", etc
  targetId: string;
  relationType: string;    // "citation", "pov", "setting", etc
  contextType?: string;
  contextId?: string;
  attributesJson?: string;
  summary?: string;
  createdAt: string;
}

// Citation parsed for display
interface ParsedCitation {
  // Source info (quem está citando)
  sourceType: 'chapter' | 'scene' | 'beat' | 'content_block';
  sourceId: string;
  sourceTitle: string;
  
  // Story context
  storyId: string;
  storyTitle: string;
  chapterId?: string;
  chapterTitle?: string;
  
  // Relation details
  relationType: string;    // "citation", "pov", "setting"
  summary?: string;
  attributes?: Record<string, any>;
}

interface CitationsGenerator {
  // Fetch citations from API using EntityRelation
  fetchCitationsForEntity(
    targetType: WorldEntityType,
    targetId: string
  ): Promise<EntityRelationDTO[]>;
  
  // Parse raw relations into displayable citations
  parseCitations(
    relations: EntityRelationDTO[],
    context: SyncContext
  ): Promise<ParsedCitation[]>;
  
  // Generate the full .citations.md file
  generateCitationsFile(
    entity: WorldEntity,
    citations: ParsedCitation[]
  ): string;
  
  // Group citations by story for organized display
  groupByStory(citations: ParsedCitation[]): Map<string, ParsedCitation[]>;
  
  // Group by relation type within a story
  groupByRelationType(citations: ParsedCitation[]): Map<string, ParsedCitation[]>;
  
  // Generate summary table at end of file
  generateSummaryTable(citations: ParsedCitation[]): string;
}
```

### RelationsGenerator

Usa `EntityRelation` com `ListRelationsBySource` para encontrar o que uma entity referencia.

```typescript
// Parsed relation for display
interface ParsedRelation {
  // Target info (quem está sendo referenciado)
  targetType: string;      // "character", "location", "faction", etc
  targetId: string;
  targetName: string;
  
  // Relation details
  relationType: string;    // "pov", "setting", "owns", "member_of"
  summary?: string;
  attributes?: Record<string, any>;
  
  // Context (onde a relação foi declarada)
  contextType?: string;
  contextId?: string;
}

interface RelationsGenerator {
  // Fetch relations from API using EntityRelation
  fetchRelationsForEntity(
    sourceType: string,
    sourceId: string
  ): Promise<EntityRelationDTO[]>;
  
  // Parse raw relations into displayable format
  parseRelations(
    relations: EntityRelationDTO[],
    context: SyncContext
  ): Promise<ParsedRelation[]>;
  
  // Generate the full .relations.md file
  generateRelationsFile(
    entity: BaseEntity,
    relations: ParsedRelation[]
  ): string;
  
  // Group relations by target type for organized display
  groupByTargetType(relations: ParsedRelation[]): Map<string, ParsedRelation[]>;
  
  // Group by relation type
  groupByRelationType(relations: ParsedRelation[]): Map<string, ParsedRelation[]>;
}
```

### FileRenamer

```typescript
interface FileRenamer {
  // Check if file needs renaming based on entity data
  needsRename(
    currentPath: string, 
    entity: BaseEntity, 
    handler: EntityHandler<any>
  ): boolean;
  
  // Calculate new path
  getNewPath(
    currentPath: string, 
    entity: BaseEntity, 
    handler: EntityHandler<any>
  ): string;
  
  // Execute rename with all necessary updates
  rename(
    currentPath: string, 
    newPath: string, 
    context: SyncContext
  ): Promise<void>;
  
  // Update all references to renamed file
  updateReferences(
    oldPath: string, 
    newPath: string, 
    context: SyncContext
  ): Promise<void>;
}
```

---

## Parsing de Content Fences

### Regex Pattern

```typescript
const CONTENT_FENCE_PATTERN = /<!--\s*content-start:([a-zA-Z0-9-]+)\s*-->\n([\s\S]*?)\n<!--\s*content-end:\1\s*-->/g;
```

### Exemplo de Parsing

```typescript
function parseContentFences(content: string): ContentFence[] {
  const fences: ContentFence[] = [];
  const lines = content.split('\n');
  
  let match;
  while ((match = CONTENT_FENCE_PATTERN.exec(content)) !== null) {
    const id = match[1];
    const fenceContent = match[2];
    
    // Calculate line numbers
    const startIndex = match.index;
    const startLine = content.substring(0, startIndex).split('\n').length;
    const endLine = startLine + fenceContent.split('\n').length + 1;
    
    fences.push({
      id,
      content: fenceContent,
      startLine,
      endLine
    });
  }
  
  return fences;
}
```

---

## Fluxo de Rename

Quando um `order_num` muda:

```
1. User reorders scene in outline.md
2. OutlineParser detecta mudança de ordem
3. SyncOrchestrator identifica scenes afetadas
4. Para cada scene afetada:
   a. FileRenamer.needsRename() -> true
   b. FileRenamer.getNewPath() -> novo caminho
   c. FileRenamer.rename():
      - Renomeia arquivo físico
      - Atualiza frontmatter se necessário
   d. FileRenamer.updateReferences():
      - Atualiza links em outline.md
      - Atualiza links em contents.md
      - Atualiza links em relations.md
      - Atualiza links em parent files
5. Push mudanças para API
```

---

## Fases de Implementação

### Fase 1: Foundation (Core + Types)
- [x] Definir todas as interfaces em `types/`
- [x] Implementar `EntityRegistry`
- [x] Implementar `FileManager` simplificado
- [ ] Implementar `FrontmatterParser`
- [ ] Implementar `PathResolver`
- [ ] Implementar `FileNaming`

**Testes de integração necessários:** validar carregamento básico do plugin com settings defaults (modo local/remote) garantindo que a instância correta de `SyncEngine` é criada quando `syncVersion` ou `mode` mudam.

### Fase 2: Parsers
- [x] Implementar `OutlineParser`
- [x] Implementar `ContentsParser` (com fences)
- [x] Implementar `RelationsParser`
- [x] Implementar `EntityFileParser`

**Testes de integração necessários:** rodar parsers em arquivos reais de vault (outline/contents/relations) e checar se os diffs resultantes acionam as operações corretas (criação de entidades, reordenação, placeholders → entidades reais).

### Fase 3: Generators
- [x] Implementar `OutlineGenerator`
- [x] Implementar `ContentsGenerator`
- [x] Implementar `RelationsGenerator`
- [x] Implementar `FrontmatterGenerator` (centralizado para reutilização)

#### Estado atual

- `sync-v2/generators/OutlineGenerator.ts` produz o arquivo `story.outline.md`, reutilizando o mesmo formato documentado (help box opcional, placeholders e links sanitizados). Testes em `outlineGenerator.test.ts` validam a hierarquia e a presença dos placeholders.
- `sync-v2/generators/ContentsGenerator.ts` gera `story.contents.md` com fences HTML (`chapter/scene/beat/content`). Aceita maps de content blocks para capítulos, scenes e beats, garantindo placeholders quando arrays estão vazios. Coberto por `contentsGenerator.test.ts`.
- `sync-v2/generators/RelationsGenerator.ts` renderiza `*.relations.md`, agrupando por `targetType` com placeholders `_Add new ..._`. Testado em `relationsGenerator.test.ts`.
- `sync-v2/generators/CitationsGenerator.ts` gera `*.citations.md` (auto warning, agrupamento por história/tipo e tabela resumo) com testes em `citationsGenerator.test.ts`.
- `sync-v2/generators/FrontmatterGenerator.ts` centraliza geração de frontmatter YAML para todas as entities (story, chapter, scene, beat, world, character, location, etc.), incluindo tags Obsidian (entity type, story/world name, date). Testado em `frontmatterGenerator.test.ts`.

**Status da Fase 3**: ✅ Completa - Todos os generators implementados e testados.

**Testes de integração necessários:** executar os generators com payloads reais e comparar o output (`story.outline.md`, `story.contents.md`, etc.) com snapshots aprovados (incluindo cenários com placeholders e múltiplos níveis).

### Fase 4: Story Handlers
 - [x] Implementar `StoryHandler`
 - [x] Implementar `ChapterHandler`
 - [x] Implementar `SceneHandler`
 - [x] Implementar `BeatHandler`
 - [x] Implementar `ContentBlockHandler`

#### Estado atual

- Registry e `StoryHandler` stub criados para permitir integração incremental (F4 em andamento).
- StoryHandler agora gera `story.outline.md`, `story.contents.md` e `story.md` via FileManager + generators (push/delete ainda pendentes).
- `ChapterHandler`, `SceneHandler`, `BeatHandler` e `ContentBlockHandler` criados com fluxo de pull completo (escrevem arquivos unitários usando FileManager). Testes unitários cobrem cada handler.
- `SyncOrchestrator` agora registra todos os handlers e aceita operações `pull_chapter/scene/beat/content_block` (a UI vai acionar quando estiver pronta).

**Testes de integração necessários:** rodar `pull_story` completo em um vault fake, validando que todos os arquivos (story/outline/contents/chapters/scenes/beats/content blocks) são escritos corretamente e que o orchestrator consegue encadear múltiplos handlers.

### Fase 5: File Renaming
 - [x] Implementar `FileRenamer`
 - [x] Implementar lógica de update references
 - [x] Testes de rename em cascata

#### Escopo planejado

- Renomear arquivos quando `order_num` ou título sanitizado mudarem (chapters/scenes/beats/content blocks), preservando o frontmatter `id`.
- Atualizar referências em todos os arquivos relevantes (`story.outline.md`, `story.contents.md`, `chapter.*`, `scene.*`, `relations`, etc.).
- Manter manifest dos renames para facilitar rollback (integração futura com backup/git).
- Detectar conflitos (ex.: destino já existe) e notificar o usuário para resolução manual.

#### Estado atual

- `PathResolver` gera nomes/paths normalizados (`sc-XX-slug.md`, `bt-XX-slug.md`) de forma consistente com os novos handlers.
- `FileRenamer` renomeia arquivos via `FileManager.renameFile` e atualiza referências com expressões regulares configuráveis; já possui testes unitários (`fileRenamer.test.ts`).
- `FileManager` ganhou helpers genéricos (`readFile`, `writeFile`, `renameFile`) para facilitar mocks e permitir que o renamer opere sem acessar o Vault diretamente.
- Renames agora cobrem `chapter` e content blocks: ao detectar `reordered` no diff, o `StoryHandler` busca a entidade correta, monta o path com `PathResolver` (`ch-0001-*.md`, `cb-0001-*.md`) e dispara o `FileRenamer`. Tests adicionais garantem que capítulos e blocos seguem o novo padrão.

**Testes de integração necessários:** simular renomeações através do outline (mover scenes/beats), rodar sync e conferir se arquivos foram renomeados e se todos os links dentro do vault foram atualizados (outline/contents/relations).

### Fase 6: Sync Orchestrator
- [ ] Implementar `SyncOrchestrator`
- [ ] Implementar `DiffEngine`
- [ ] Integrar com handlers

#### Escopo planejado

- `DiffEngine` para comparar arquivos locais vs payload da API e gerar operações (create/update/delete/reorder).
- `SyncOrchestrator` passando a coordenar pull/push completos: aplica diffs, dispara handlers específicos e utiliza `FileRenamer` quando necessário.
- Registro de métricas/telemetria (tempo de diff, número de operações).
- Suporte a retry e relatórios de conflito (mostrar quais arquivos precisam de intervenção manual).

#### Estado atual

- `DiffEngine` implementado para `story.contents.md` com detecção de operações (`created`, `updated`, etc.) baseada nos fences do `ContentsParser`, além de capturar trechos desconhecidos (fora das fences) para garantir que não sejam perdidos.
- Testes unitários (`diff/__tests__/DiffEngine.test.ts`) validam tanto a criação de operações quanto a preservação de texto fora das regiões rastreadas.
- `SyncOrchestrator` passou a instanciar o `DiffEngine`, preparando a integração com os handlers nas próximas etapas.
- `ContentsReconciler` injeta blocos de “texto preservado” quando encontra trechos fora das fences; `StoryHandler` já usa o reconciler antes de sobrescrever `story.contents.md` e aciona `FileRenamer` para renomear scenes/beats quando o diff sinaliza reordenação.
- `OutlineReconciler` e `RelationsReconciler` criados para manter o mesmo comportamento em `story.outline.md` e `*.relations.md` (texto do escritor nunca é descartado).
- Testes no `StoryHandler` cobrem o fluxo de reconciliação + rename para scenes; diffs adicionais garantem conversão automática de `<!--new-*-->`.
- Push pipeline inicial: `PushPlanner` compara `story.contents.md` (local) com o conteúdo gerado pelo backend e produz um plano estrutural (reordena/move chapters/scenes/beats); `PushExecutor` traduz esse plano em chamadas `updateChapter/updateScene/moveScene/updateBeat/moveBeat`. `SyncOrchestrator.handlePushStory` agora lê `story.md` para descobrir o ID remoto, gera o diff e aplica as operações suportadas — mudanças não suportadas (placeholders, texto desconhecido) ficam listadas para tratamento posterior.
- Conflitos/Avisos: `ContentsReconciler` gera `warnings` sempre que preserva trechos desconhecidos; `PushPlanner` adiciona alertas para operações ainda não suportadas ou texto fora das fences. O `SyncOrchestrator` agrega esses avisos e o `ModularSyncEngine` mostra um `Notice` resumindo o que precisa de atenção (além do bloco `<!-- story-engine/untracked -->` dentro do arquivo).

**Testes de integração necessários:** pipeline completo (pull → editar → push) garantindo que o diff engine identifica mudanças, resolve conflitos e aplica atualizações. Incluir casos com erros de rede/api para validar retries.

### Fase 7: World Handlers
- [x] Implementar `WorldHandler`
- [x] Implementar `CharacterHandler`
- [x] Implementar `LocationHandler`
- [x] Implementar `FactionHandler`
- [x] Implementar `ArtifactHandler`
- [x] Implementar `EventHandler`
- [x] Implementar `LoreHandler`
- [x] Implementar `ArchetypeHandler`
- [x] Implementar `TraitHandler`
- [x] Migrar World handlers para usar `FrontmatterGenerator` ao invés de renderização manual
  - [x] `CharacterHandler.renderCharacter()` → usar `FrontmatterGenerator`
  - [x] `LocationHandler.renderLocation()` → usar `FrontmatterGenerator`
  - [x] `FactionHandler.renderFaction()` → usar `FrontmatterGenerator`
  - [x] `ArtifactHandler.renderArtifact()` → usar `FrontmatterGenerator`
  - [x] `EventHandler.renderEvent()` → usar `FrontmatterGenerator`
  - [x] `LoreHandler.renderLore()` → usar `FrontmatterGenerator`
  - [x] `ArchetypeHandler.renderArchetype()` → usar `FrontmatterGenerator`
  - [x] `TraitHandler.renderTrait()` → usar `FrontmatterGenerator`
  - [x] `WorldHandler` (via `writeWorldMetadata` no FileManager) → já usa `generateFrontmatter`, consistência verificada

**Testes de integração necessários:** sincronizar um world inteiro e validar que `world.*` + `character/location/...` arquivos são produzidos corretamente, incluindo relações/citações entre World/Story.

### Fase 8: Relations & Citations System
- [x] Definir tipos em `relations.ts` e `citations.ts`
- [x] Implementar client para `EntityRelationService` (HTTP)
  - [x] `listRelationsBySource` - buscar relações onde entity é source
  - [x] `listRelationsByTarget` - buscar relações onde entity é target
  - [x] `listRelationsByWorld` - buscar todas relações de um world
- [x] Implementar `RelationsGenerator`
  - [x] Gerar `.relations.md` para Story entities (via `StoryHandler`)
  - [x] Gerar `.relations.md` para World entities (via `WorldHandler`)
  - [x] Suporte para characters, locations, factions, artifacts, events, lore
  - [x] Seção "World" quando entity tem `world_id`
  - [x] Placeholders para adicionar novas relações
- [x] Implementar `CitationsGenerator`
  - [x] Gerar `.citations.md` para World entities (via `WorldHandler`)
  - [x] Agrupar citações por story e relation type
  - [x] Tabela resumo de citações
  - [x] Aviso sobre arquivo auto-gerado
- [ ] Criar EntityRelations ao sincronizar (PUSH - Fase 9)
  - [x] Ao adicionar relação em `.relations.md` → criar relation na API (implementado em `RelationsPushHandler`)
  - [x] Ao modificar relação existente → atualizar relation na API (implementado em `RelationsPushHandler`)
  - [x] Ao remover relação → deletar relation na API (implementado em `RelationsPushHandler`)
  - [x] Ao adicionar POV character em Scene → criar relation automaticamente (implementado em `SceneHandler.push()`)
  - [x] Ao adicionar Location em Scene → criar relation automaticamente (implementado em `SceneHandler.push()`)
  - [ ] Ao mencionar entity em ContentBlock → criar citation automaticamente (ver seção detalhada na Fase 9 - Opção 3: Nível Mais Específico + Context)
- [x] Auto-regenerate citations/relations on story/world pull (geração automática durante pull)

**Testes de integração necessários:** rodar um sync que gera `*.relations.md` e `*.citations.md` e comparar com snapshots; editar `.relations.md`, rodar push e confirmar que as mudanças chegam à API (EntityRelationService).

### Fase 9: Push Relations & Auto Sync
- [x] Implementar push de relations (PUSH)
  - [x] Parser para `.relations.md` (detectar adições/modificações/remoções)
  - [x] Criar EntityRelations via API quando adicionar relação
  - [x] Atualizar EntityRelations via API quando modificar relação
  - [x] Deletar EntityRelations via API quando remover relação
  - [x] Validar que target entity existe antes de criar relação
- [x] Criar EntityRelations automaticamente durante sync
  - [x] Ao adicionar POV character em Scene → criar relation automaticamente (implementado em `SceneHandler.push()`)
  - [x] Ao adicionar Location em Scene → criar relation automaticamente (implementado em `SceneHandler.push()`)
  - [ ] Ao mencionar entity em ContentBlock → criar citation automaticamente (ver seção abaixo)
- [ ] Refatorar `AutoSyncManager`
- [ ] Implementar `ConflictResolver`
- [ ] Integrar `ApiUpdateNotifier`

#### Criação Automática de Citations em ContentBlocks

**Contexto:**
- ContentBlocks estão organizados em uma hierarquia: Chapter → Scene → Beat → ContentBlock
- Um ContentBlock pode ter múltiplos ContentAnchors conectando-o a diferentes entidades (Scene, Beat, Chapter)
- ContentBlocks são armazenados em `stories/{story-title}/03-contents/{type}/{cb-XXXX-title}.md`
- Quando uma World Entity (Character, Location, Artifact, etc.) é mencionada em um ContentBlock, devemos criar uma citation

**Estratégia Escolhida: Opção 3 - Nível Mais Específico + Context**

**Decisões de Design:**
1. **Onde criar a citation?**
   - Criar no nível mais específico disponível via ContentAnchors:
     - Se ContentBlock tem anchor para Beat → criar citation no Beat (`source_type: "beat"`)
     - Se tem anchor para Scene (mas não Beat) → criar citation no Scene (`source_type: "scene"`)
     - Se tem anchor apenas para Chapter → criar citation no Chapter (`source_type: "chapter"`)
     - Exceção: se não tem anchors → criar no Chapter ou Story level
   
2. **Como incluir contexto?**
   - Usar o campo `context` na EntityRelation para incluir hierarquia completa:
     ```
     context: "Chapter 1: Introduction > Scene 2: The Meeting > Beat 3: Confrontation"
     ```
   - Isso permite navegação completa mesmo quando citation está no nível mais específico

3. **Detecção de menções:**
   - **Fase 1 (Manual)**: Usuário pode criar citations manualmente via `.relations.md`
   - **Fase 2 (LLM/API)**: Detectar menções automaticamente via LLM durante push do ContentBlock
     - LLM será trabalhado em thread separada
     - API endpoint para detecção de entidades mencionadas no texto

4. **Implementação no ContentBlockHandler.push():**
   ```typescript
   // 1. Detectar entidades mencionadas (via LLM/API ou parsing de links)
   //    - Detectar links no formato [[filename path]] ou [[filename path|display]]
   //    - Exemplos: [[worlds/eldoria/characters/aria-moon]], [[worlds/eldoria/locations/crystal-cave]]
   //    - Extrair: world-name, entity-type (characters/locations/factions/etc), entity-slug
   //    - Resolver filename path para entity ID via API
   // 2. Obter ContentAnchors do ContentBlock para determinar hierarquia
   // 3. Determinar nível mais específico (beat > scene > chapter)
   // 4. Para cada entidade mencionada:
   //    - Determinar source_type e source_id baseado em ContentAnchors
   //    - Construir context string com hierarquia completa
   //    - Criar citation relation (source_type → target_type="character/location/etc", relation_type="citation")
   //    - Validar que target entity existe antes de criar
   // 5. Renderizar links com filename path completo para evitar ambiguidade
   ```

5. **Hierarquia de ContentAnchors:**
   - Um ContentBlock pode ter múltiplos ContentAnchors apontando para:
     - Beat (mais específico)
     - Scene (nível intermediário)
     - Chapter (nível mais geral)
   - Prioridade: Beat > Scene > Chapter
   - Se ContentBlock tem anchor para Beat, criar citation no Beat
   - Se não tem Beat mas tem Scene, criar no Scene
   - Se não tem Beat nem Scene, criar no Chapter

6. **Links no Markdown:**
   - Possibilitar criar links automáticos para entidades citadas:
     - Detectar links no formato `[[filename path]]` ou `[[filename path|display]]`
       - Exemplos:
         - `[[worlds/eldoria/characters/aria-moon]]` - Character
         - `[[worlds/eldoria/locations/crystal-cave]]` - Location
         - `[[worlds/eldoria/factions/the-guard]]` - Faction
         - `[[worlds/eldoria/artifacts/sword-of-light]]` - Artifact
         - `[[worlds/eldoria/events/the-great-war]]` - Event
         - `[[worlds/eldoria/lore/magic-system]]` - Lore
     - **Formato do filename path**: `worlds/{world-name}/{entity-type}/{entity-slug}`
       - `world-name`: Nome do world (sanitizado)
       - `entity-type`: Um de `characters`, `locations`, `factions`, `artifacts`, `events`, `lore`
       - `entity-slug`: Nome da entidade sanitizado (ex: `aria-moon`, `crystal-cave`)
     - **Importante**: Links devem ser renderizados com filename path completo para evitar ambiguidade
       - Motivo: pode haver 2 entities com nomes diferentes no mesmo world (ex: dois characters diferentes)
       - O filename path é único e não ambíguo
     - Resolver filename path para ID via API:
       - Extrair `world-name` e `entity-type` do path
       - Extrair `entity-slug` (nome sanitizado) do path
       - Buscar entidade no world correspondente via API usando o slug
     - Criar citation automaticamente quando link é detectado

**Tarefas de Implementação:**
- [ ] Implementar `ContentBlockHandler.push()` básico (atualizar conteúdo)
- [ ] Adicionar helper `detectEntityMentions()` para detectar menções via parsing de links
- [ ] Adicionar helper `resolveContentBlockHierarchy()` para obter ContentAnchors e determinar nível
- [ ] Adicionar helper `buildHierarchyContext()` para construir string de contexto com hierarquia completa
- [ ] Adicionar helper `createCitationRelations()` para criar citations no nível correto
- [ ] Integrar com ContentAnchors API para determinar hierarquia
- [ ] Validar que target entity existe antes de criar citation
- [ ] Integrar com LLM/API para detecção automática (Fase 2)
- [ ] Atualizar `.citations.md` files automaticamente quando citations são criadas

**Vantagens da Opção 3:**
- ✅ Sem duplicação: uma citation por menção no nível mais específico
- ✅ Contexto completo: campo `context` permite navegação completa
- ✅ Flexível: citations podem ser agregadas visualmente nos parents quando necessário
- ✅ Escalável: funciona tanto com detecção manual quanto LLM
- ✅ Precisa: citations estão no nível exato onde a menção ocorre

**Exemplo de Citation Criada:**
```typescript
{
  source_type: "beat",
  source_id: "beat-123",
  target_type: "character",
  target_id: "character-456",
  relation_type: "citation",
  context: "Chapter 1: Introduction > Scene 2: The Meeting > Beat 3: Confrontation",
  // ... outros campos
}
```

**Descobertas Estruturais:**

1. **ContentBlocks e ContentAnchors:**
   - ContentBlocks têm apenas `chapter_id` (opcional) - não têm `scene_id` ou `beat_id` diretamente
   - ContentAnchors são entidades separadas que conectam ContentBlock → Entity (Scene, Beat, Chapter)
   - Um ContentBlock pode ter múltiplos ContentAnchors apontando para diferentes entidades
   - Hierarquia: Chapter → Scene → Beat → ContentBlock (via ContentAnchors)
   - Prioridade para determinar nível: Beat > Scene > Chapter

2. **Citations vs Relations:**
   - **Citations** (`relation_type: "citation"`): Story entities (chapter, scene, beat, content_block) citando World entities (character, location, etc.)
   - **Relations** (outros tipos): Relações entre entities (pov, setting, owns, member_of, etc.)
   - Citations aparecem em `.citations.md` de World entities
   - Relations aparecem em `.relations.md` de Story e World entities

3. **POV e Location em Scenes:**
   - Scenes têm `pov_character_id` e `location_id` no frontmatter
   - Quando esses campos são alterados, criar relations automaticamente:
     - `pov_character_id` → relation type: "pov" (scene → character)
     - `location_id` → relation type: "setting" (scene → location)
   - Implementado em `SceneHandler.push()` com detecção de mudanças

**Testes de integração necessários:**
- Push relations: editar `.relations.md`, rodar push e confirmar que as mudanças chegam à API (create/update/delete EntityRelations)
- Auto sync: simular eventos de blur/idle/editor change e medir se os debounces disparam push/pull conforme configurado
- Fallback manual quando o auto-sync está desligado
- Citations automáticas: mencionar Character em ContentBlock e verificar se citation é criada no nível correto (Beat/Scene/Chapter) com context completo

### Fase 10: Migration & Testing
- [ ] Script de migração de formato antigo
- [ ] Testes end-to-end
- [ ] Documentação de uso

**Testes de integração necessários:** rodar o script de migração em um vault leg antigo, comparar o resultado com snapshots aprovados e executar um ciclo completo (pull → editar → push) para validar compatibilidade retroativa.

### Fase 11 (nova): Backups & Git Integration
- [ ] Implementar snapshots automáticos antes de cada `pull`/`push`
- [ ] Criar flag nas configurações para ativar/desativar backup Git
- [ ] Integrar com CLI (`git status/add/commit`) quando flag estiver ativa
- [ ] Adicionar comandos para restaurar snapshot/Git commit dentro do plugin

**Testes de integração necessários:** executar sync em modo snapshots e checar a estrutura `backups/YYYY-MM-DD/HH-mm-ss` com manifestos; para Git, usar repo temporário e garantir que commits/tags e restaurações funcionam.

### Entregas Detalhadas por Fase

| Fase | Entregáveis Principais | Responsáveis | Dependências | Testes Requeridos |
|------|------------------------|--------------|--------------|-------------------|
| 1. Foundation | Tipos em `sync/types`, `FileManager`/`PathResolver` revisados | Core/Infra | Manifest + settings | Unit: type guards, PathResolver cases |
| 2. Parsers | `OutlineParser`, `ContentsParser`, `RelationsParser`, `EntityFileParser` | Core Sync | Foundation pronta | Unit intensivos com fixtures Markdown; snapshot para placeholders |
| 3. Generators | Arquivos `.outline.md/.contents.md/.relations.md/.citations.md` | Core Sync | Parsers + tipos | Unit (string compare), smoke e snapshot |
| 4. Story Handlers | `Story/Chapter/Scene/Beat/ContentBlock Handler` e registry | Story squad | Fases 1-3 | Unit por handler, contract tests com orchestrator |
| 5. File Renaming | `FileRenamer` + update references | File Ops | Handlers prontos | Integration com vault fake; e2e rename cascata |
| 6. Sync Orchestrator | DiffEngine + pipeline orchestrator | Sync Core | Handlers + FileOps | Unit diff cases, integration pipeline, resilience tests |
| 7. World Handlers | World/Character/Location/... | World squad | Orchestrator ready | Unit + integration com relations/citations |
| 8. Relations & Citations | Client EntityRelation + generators (PULL) | World squad | World handlers | ✅ Unit tests para RelationsGenerator/CitationsGenerator; ✅ Unit tests para handlers integrando relations/citations; ⏳ Integration: push relations (Fase 9) |
| 9. Push Relations & Auto Sync | Push relations + `AutoSyncManager` refactor + API update hooks | Platform | Orchestrator + handlers | Integration (push relations, debounce), regression with V1 parity |
| 10. Migration & Testing | Scripts, docs, end-to-end | All squads | Todas as fases | Full suite: e2e, regression, manual UAT |

### Estratégia de Testes

1. **Unit**  
   - Preferir Vitest; cada parser/generator/handler precisa de fixtures cobrindo casos felizes e edge cases.  
   - Utilizar snapshots apenas para arquivos grandes (contents/outline). Sempre validar metadados (IDs, order).  
2. **Integration (Vault fake)**  
   - Criar helpers que simulam `TFile/TFolder` para testar `FileManager`, `FileRenamer`, `AutoSyncManager`.  
   - Casos obrigatórios: rename cascata, placeholders → entidades reais, diff bidirecional.  
3. **Contract/API**  
   - Mock programático de `StoryEngineClient` validando payloads enviados aos endpoints (create/update/delete).  
   - Fixtures baseados no Postman collection (`Story_Engine_API.postman_collection.json`).  
4. **End-to-End**  
   - Scripts CLI disparando `pull_story`/`push_story` com vault sandbox; comparar resultado com snapshots aprovados.  
   - Fluxos principais: criar story → editar outline/contents → push → verificar backend.  
5. **Regression/Parity**  
   - Conjunto de histórias pequenas sincronizadas tanto por V1 quanto V2; diffs devem ser idênticos (exceto campos novos).  
   - CI: rodar `npm run test` (unit), `npm run test:integration` (quando existir), e pipeline e2e opcional com flag.  
6. **Performance & Debounce**  
   - Medir tempo médio de parse e diff para stories com 1k+ fences. Meta: <150ms parse incremental.  
   - Testar AutoSync em idle vs typing contínuo para garantir que debounce não flooda a API.  
7. **Manual QA / Writer Experience**  
   - Checklist em `TESTING_WORKFLOW.md` atualizado com cenários V2 (toggle, fallback, mensagens).  
   - Validar tooltips/alerts dentro dos arquivos (`showHelpBox`) com e sem V2.  

### Critérios de Aceite Gerais

- **Paridade funcional**: qualquer operação suportada no V1 deve ter caminho equivalente no V2 antes de torná-lo padrão.  
- **Observabilidade**: logs estruturados para cada operação (pull/push/diff) com IDs de entidade, duração e erros.
- **Rollback seguro**: alternar `Sync Version` deve ser imediato e sem efeitos colaterais nos arquivos existentes.  
- **Test Coverage**: mínimo 80% nas áreas novas (parsers/generators/handlers).  
- **Documentação**: cada módulo precisa de README curto indicando responsabilidades, APIs e testes relevantes.  
- **Debounce/Auto Push**: qualquer criação local deve respeitar as mesmas regras estabelecidas para V1 (1s pause, 5s idle, blur).  

## Estratégia de Backups e Integração com Git

### Objetivo

Garantir que antes de qualquer `pull` ou `push` do Sync V2 haja um snapshot recuperável dos arquivos, protegendo contra bugs de parser/filemanager que possam limpar trechos de texto e permitindo rollback imediato mesmo sem alternar versões. Quando possível, reaproveitar Git como mecanismo de histórico com opt-in explícito.

### Snapshot Mode (default)

- **Trigger**: imediatamente antes de `SyncOrchestrator.pull_story` ou `push_story`.
- **Destino**: pasta `.story-engine/backups/YYYY-MM-DD/HH-mm-ss/` contendo apenas os arquivos realmente modificados na operação (para reduzir espaço).
- **Conteúdo**: cópia dos arquivos tocados na operação + manifesto JSON (lista de arquivos, hash e origem da operação).
- **Retenção**: manter snapshots completos dos últimos 7 dias; após esse período, apenas registros existentes via Git (se habilitado) ficam disponíveis.
- **Config**: nova opção em Settings → Story Engine → “Automatic Backups” com valores `off | snapshots | git`.

### Git Mode (opt-in)

- Quando o usuário selecionar `git`, o plugin executa:
  1. `git status --porcelain` para conferir arquivos tocados.
  2. `git add <arquivos modificados pelo sync>`.
  3. `git commit -m "[StoryEngine Sync] pre-pull <story>/<operation>"`.
  4. Opcional: `git tag storyengine/<story-id>/<timestamp>` para facilitar rollback.
- Necessário expor campo “Git Binary Path” nas configs para ambientes não padrão.
- Em caso de erro (repo inexistente, sem permissões, etc.), revert para Snapshot mode e notificar o usuário.

### Restauração

- **Snapshots**: comando rápido “Restore last backup” que copia os arquivos do snapshot escolhido de volta para o vault.
- **Git**: abrir modal com últimos commits StoryEngine + botão `Reset files` (executa `git checkout <commit> -- <files>`).

### Atualizações no Plano

- Adicionar flag na seção de configurações (`settings.ts`) ao final do desenvolvimento.
- Orquestrador deve receber o modo de backup via `SyncContext` e executar a estratégia antes de chamar os generators/handlers.
- Testes:
  - Unit: garantir que o `BackupManager` grava manifestos corretos.
  - Integration: simular repo git fake e validar commits/tags.
  - Manual: fluxo completo de restore/snapshot e fallback quando Git falha.
---

## Comparação: Antes vs Depois

| Aspecto | Sync V1 | Sync V2 |
|---------|---------|---------|
| Linhas de código | ~3370 (1 arquivo) | ~200-400 por handler |
| Entity types | Story hierarchy only | Story + World |
| Adicionar entity | Modificar monólito | Criar novo handler |
| Content blocks | Single paragraph links | Multi-paragraph fences |
| Relations | Inline no arquivo | Arquivo dedicado `.relations.md` |
| Citations | Não existe | Auto-generated `.citations.md` |
| File naming | Timestamp-based | Order + Title |
| Rename support | Não existe | Automático com ref update |
| Testabilidade | Difícil | Handlers isolados |
| World entities | Não suportado | contents + relations + citations |

---

## Notas de Implementação

### Content Fences vs Links

**Por que fences ao invés de `[[link|preview]]`?**

1. **Multi-parágrafo**: Links só mostram preview de uma linha
2. **Edição inline**: Usuário pode editar conteúdo diretamente
3. **Sem arquivo extra**: Não precisa criar .md para cada content block
4. **Tracking**: ID explícito facilita sync bidirecional
5. **Markdown preservado**: Formatting funciona dentro da fence

**Trade-offs**:
- Arquivo contents.md fica maior
- Parsing mais complexo que links simples
- Menos "Obsidian native" (não usa wiki links)

### Order-based Naming

**Por que `order_num` ao invés de timestamp?**

1. **Ordenação visual**: Arquivos aparecem em ordem no file explorer
2. **Semântica clara**: `sc-01`, `sc-02` vs `sc-1704810000`
3. **Edição manual**: Fácil entender e renomear manualmente se necessário

**Trade-offs**:
- Precisa de rename quando ordem muda
- Potencial conflito se dois itens têm mesmo order
- Mais complexo que ID fixo

---

## Quick Reference: File Types

| File | Purpose | Editable | Placeholders |
|------|---------|----------|--------------|
| `*.outline.md` | Hierarquia editável | ✅ Yes | `_New chapter/scene/beat..._` |
| `*.contents.md` | Content blocks | ✅ Yes | `<!--*-start:0000:new-*:placeholder-->` |
| `*.relations.md` | Relações com World | ✅ Yes | `_Add new character/location..._` |
| `*.citations.md` | Onde é citado | ⛔ **NÃO EDITE** | N/A (auto-gerado) |
| `sc-NN-title.md` | Detalhes de scene | ✅ Yes | — |
| `bt-NN-title.md` | Detalhes de beat | ✅ Yes | — |

## Quick Reference: Naming

```
Stories:
  chapters/ch-01-title.outline.md
  chapters/ch-01-title.contents.md
  chapters/ch-01-title.relations.md
  scenes/sc-01-goal.md
  beats/bt-01-intent.md

Worlds:
  characters/{name}.md
  characters/{name}.contents.md
  characters/{name}.relations.md
  characters/{name}.citations.md    # auto-generated
  
  locations/{name}.md
  locations/{name}.contents.md
  locations/{name}.relations.md
  locations/{name}.citations.md     # auto-generated
  
  (same pattern for factions, artifacts, events, lore)
```

## Quick Reference: HTML Fences

**Formato**: `<!--{type}-{start|end}:{order}:{name}:{id}-->`

```markdown
<!--chapter-start:0001:the-beginning:ch-uuid-001-->
## Chapter 1: The Beginning

<!--scene-start:0001:meet-hero:sc-uuid-001-->
### Scene 1: Meet the hero

<!--content-start:0001:the-sun-rose:cb-uuid-001-->
First paragraph of content.

Second paragraph with **formatting**.
<!--content-end:0001:the-sun-rose:cb-uuid-001-->

<!--beat-start:0001:introduction:bt-uuid-001-->
#### Beat 1: Introduction

<!--content-start:0001:village-bell:cb-uuid-002-->
Beat content here...
<!--content-end:0001:village-bell:cb-uuid-002-->
<!--beat-end:0001:introduction:bt-uuid-001-->

<!--scene-end:0001:meet-hero:sc-uuid-001-->
<!--chapter-end:0001:the-beginning:ch-uuid-001-->
```

**Regex pattern** (genérico para todos os tipos):
```typescript
// Captura: type, action(start/end), order, name, id
const FENCE_PATTERN = /<!--(chapter|scene|beat|content)-(start|end):(\d{4}):([a-z0-9-]+):([a-zA-Z0-9-]+)-->/g;

// Exemplo de uso
const match = FENCE_PATTERN.exec(line);
if (match) {
  const [_, type, action, order, name, id] = match;
  // type: "scene"
  // action: "start"
  // order: "0001"
  // name: "meet-hero"
  // id: "sc-uuid-001"
}
```

**Reordenação**: Quando fences mudam de posição, recalcular `order` baseado na posição no arquivo.

**Placeholders** (ignorados até serem modificados):
```markdown
<!--beat-start:0000:new-beat:placeholder-->
#### _New Beat Intent_

<!--content-start:0000:new-content:placeholder-->
_Write your content here..._
<!--content-end:0000:new-content:placeholder-->

<!--beat-end:0000:new-beat:placeholder-->
```

Identificação de placeholder: `order === 0000` OR `id === "placeholder"` OR `name.startsWith("new-")`

## Quick Reference: EntityRelation Types

| relation_type | source_type | target_type | Descrição |
|---------------|-------------|-------------|-----------|
| `pov` | scene | character | Character é POV da scene |
| `setting` | scene | location | Scene acontece em location |
| `appears_in` | scene/beat | artifact | Artifact aparece na scene/beat |
| `citation` | content_block | any | ContentBlock menciona entity |
| `owns` | character | artifact | Character possui artifact |
| `member_of` | character | faction | Character é membro da faction |
| `ally` | character | character | Characters são aliados |
| `enemy` | character | character | Characters são inimigos |
| `located_in` | artifact/character | location | Entity está localizada em |
| `caused` | event | event | Event causou outro event |
| `follows` | character/faction | lore | Entity segue este lore |

**Queries comuns**:
```
# Para .citations.md (quem me cita)
ListRelationsByTarget(target_type, target_id)

# Para .relations.md (quem eu cito)
ListRelationsBySource(source_type, source_id)

# Todas relações de um world
ListRelationsByWorld(world_id)
```

---

## Referências

- [Obsidian Plugin API](https://docs.obsidian.md/Plugins/Getting+started/Build+a+plugin)
- [Story Engine API](../LLM_GATEWAY_SERVICE.md)
- [Current Sync Implementation](../../obsidian-plugin/src/sync/)

