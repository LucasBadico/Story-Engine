# Frontmatter Fields Reference - Sync V2

> **Status**: Atualizado em 2025-01-09
> **Versão**: Sync V2

Este documento lista todos os campos adicionados no frontmatter para cada tipo de entidade no Sync V2.

---

## Story Entities

### Story (`story.md`)

**Base Fields:**
- `id` (string) - ID da story
- `title` (string) - Título da story
- `status` (string) - Status da story
- `version` (number) - Número da versão
- `root_story_id` (string) - ID da story raiz (para versionamento)
- `previous_version_id` (string | null) - ID da versão anterior
- `created_at` (string) - Data de criação (ISO 8601)
- `updated_at` (string) - Data de atualização (ISO 8601)

**Tags:**
- `story-engine/story`
- `story/{sanitized-story-title}`
- `date/{YYYY}/{MM}/{DD}` (baseado em `created_at`)

---

### Chapter (`ch-XXXX-{title}.md`)

**Base Fields:**
- `id` (string) - ID do chapter
- `story_id` (string) - ID da story pai
- `number` (number) - Número do chapter
- `title` (string) - Título do chapter
- `status` (string) - Status do chapter
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/chapter`
- `story/{sanitized-story-name}` (se fornecido)
- `date/{YYYY}/{MM}/{DD}` (baseado em `created_at`)

---

### Scene (`sc-XXXX-{goal}.md`)

**Base Fields:**
- `id` (string) - ID da scene
- `story_id` (string) - ID da story pai
- `chapter_id` (string | null) - ID do chapter pai (nullable)
- `order_num` (number) - Número de ordem
- `time_ref` (string) - Referência de tempo (ex: "Morning", "Evening")
- `goal` (string) - Objetivo da scene
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Extra Fields (condicionais):**
- `pov_character_id` (string | null) - ID do character POV (apenas se definido)
- `location_id` (string | null) - ID da location (apenas se definido)

**Tags:**
- `story-engine/scene`
- `story/{sanitized-story-name}` (se fornecido)
- `date/{YYYY}/{MM}/{DD}` (baseado em `created_at`)

---

### Beat (`bt-XXXX-{intent}.md`)

**Base Fields:**
- `id` (string) - ID do beat
- `scene_id` (string) - ID da scene pai
- `order_num` (number) - Número de ordem
- `type` (string) - Tipo do beat
- `intent` (string) - Intenção do beat
- `outcome` (string) - Resultado do beat
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/beat`
- `story/{sanitized-story-name}` (se fornecido)
- `date/{YYYY}/{MM}/{DD}` (baseado em `created_at`)

---

### ContentBlock (`cb-XXXX-{title}.md`)

**Base Fields:**
- `id` (string) - ID do content block
- `chapter_id` (string | null) - ID do chapter (nullable)
- `order_num` (number) - Número de ordem
- `type` (string) - Tipo: "text" | "image" | "video" | "audio" | "embed" | "link"
- `kind` (string) - Subtipo/kind do content block
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Fields Condicionais por Tipo:**

**Text (`type: "text"`):**
- `word_count` (number) - Número de palavras (se disponível no metadata)

**Image (`type: "image"`):**
- `original_url` (string) - URL original (se content é uma URL)
- `alt_text` (string) - Texto alternativo
- `caption` (string) - Legenda
- `width` (number) - Largura
- `height` (number) - Altura
- `mime_type` (string) - Tipo MIME
- `source` (string) - Fonte: "unsplash" | "internet link" | "local"
- `author_name` (string) - Nome do autor
- `attribution` (string) - Atribuição
- `attribution_url` (string) - URL de atribuição

**Video/Audio (`type: "video"` ou `type: "audio"`):**
- `provider` (string) - Provedor (ex: "youtube", "vimeo")
- `video_id` (string) - ID do vídeo
- `duration` (number) - Duração em segundos
- `thumbnail_url` (string) - URL da miniatura

**Embed (`type: "embed"`):**
- `provider` (string) - Provedor
- `embed_html` (string) - HTML do embed

**Link (`type: "link"`):**
- `link_title` (string) - Título do link
- `link_description` (string) - Descrição do link
- `link_image` (string) - Imagem do link
- `site_name` (string) - Nome do site

**Tags:**
- `story-engine/content-block`
- `story/{sanitized-story-name}` (se fornecido)
- `date/{YYYY}/{MM}/{DD}` (baseado em `created_at`)

---

## World Entities

### World (`world.md`)

**Base Fields:**
- `id` (string) - ID do world
- `genre` (string) - Gênero do world
- `rpg_system_id` (string | null) - ID do sistema RPG (nullable)
- `time_config` (string | null) - Configuração de tempo (JSON stringified, nullable)
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/world`
- `world/{sanitized-world-name}`
- `date/{YYYY}/{MM}/{DD}` (baseado em `created_at`)

---

### Character (`characters/{name}.md`)

**Base Fields:**
- `id` (string) - ID do character
- `world_id` (string) - ID do world pai
- `class_level` (number) - Nível da classe
- `archetype_id` (string | null) - ID do archetype (nullable)
- `current_class_id` (string | null) - ID da classe atual (nullable)
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/character`
- `world/{sanitized-world-name}`

**Nota**: CharacterHandler agora usa `FrontmatterGenerator` (migrado em 2025-01-09).

---

### Location (`locations/{name}.md`)

**Base Fields:**
- `id` (string) - ID da location
- `world_id` (string) - ID do world pai
- `type` (string) - Tipo da location
- `hierarchy_level` (number) - Nível na hierarquia
- `parent_id` (string | null) - ID do parent (nullable, para hierarquia)
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/location`
- `world/{sanitized-world-name}`

**Nota**: LocationHandler agora usa `FrontmatterGenerator` (migrado em 2025-01-09).

---

### Faction (`factions/{name}.md`)

**Base Fields:**
- `id` (string) - ID da faction
- `world_id` (string) - ID do world pai
- `type` (string | null) - Tipo da faction (nullable)
- `hierarchy_level` (number) - Nível na hierarquia
- `parent_id` (string | null) - ID do parent (nullable, para hierarquia)
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/faction`
- `world/{sanitized-world-name}`

**Nota**: FactionHandler agora usa `FrontmatterGenerator` (migrado em 2025-01-09).

---

### Artifact (`artifacts/{name}.md`)

**Base Fields:**
- `id` (string) - ID do artifact
- `world_id` (string) - ID do world pai
- `rarity` (string) - Raridade do artifact
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/artifact`
- `world/{sanitized-world-name}`

**Nota**: ArtifactHandler agora usa `FrontmatterGenerator` (migrado em 2025-01-09).

---

### Event (`events/{name}.md`)

**Base Fields:**
- `id` (string) - ID do event
- `world_id` (string) - ID do world pai
- `type` (string | null) - Tipo do event (nullable)
- `importance` (number) - Importância do event
- `timeline` (string | null) - Timeline do event (nullable)
- `parent_id` (string | null) - ID do parent (nullable, para hierarquia)
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/event`
- `world/{sanitized-world-name}`

**Nota**: EventHandler agora usa `FrontmatterGenerator` (migrado em 2025-01-09).

---

### Lore (`lore/{name}.md`)

**Base Fields:**
- `id` (string) - ID do lore
- `world_id` (string) - ID do world pai
- `category` (string | null) - Categoria do lore (nullable)
- `parent_id` (string | null) - ID do parent (nullable, para hierarquia)
- `hierarchy_level` (number) - Nível na hierarquia
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/lore`
- `world/{sanitized-world-name}`

**Nota**: LoreHandler agora usa `FrontmatterGenerator` (migrado em 2025-01-09).

---

### Archetype (`characters/_archetypes/{name}.md`)

**Base Fields:**
- `id` (string) - ID do archetype
- `tenant_id` (string) - ID do tenant
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/archetype`

**Nota**: ArchetypeHandler agora usa `FrontmatterGenerator` (migrado em 2025-01-09).

---

### Trait (`characters/_traits/{name}.md`)

**Base Fields:**
- `id` (string) - ID do trait
- `tenant_id` (string) - ID do tenant
- `category` (string) - Categoria do trait
- `created_at` (string) - Data de criação
- `updated_at` (string) - Data de atualização

**Tags:**
- `story-engine/trait`

**Nota**: TraitHandler agora usa `FrontmatterGenerator` (migrado em 2025-01-09).

---

## Tags Obsidian

Todas as entidades recebem tags no formato:

```
tags:
  - story-engine/{entity-type}
  - story/{sanitized-story-name}  # (apenas se entityType for story/chapter/scene/beat/content-block e storyName fornecido)
  - world/{sanitized-world-name}  # (apenas se worldName fornecido)
  - date/{YYYY}/{MM}/{DD}         # (apenas se date fornecido)
```

**Sanitização de Tags:**
- Converte para lowercase
- Normaliza (NFKD)
- Remove caracteres especiais (mantém apenas alphanuméricos, espaços e hífens)
- Substitui espaços por hífens
- Remove hífens duplicados
- Remove hífens no início/fim

---

## Observações Importantes

1. **Campos Null/Undefined**: Campos com valor `null` são renderizados como `field: null` no YAML
2. **Strings com Caracteres Especiais**: Strings contendo `:`, `"` ou `\n` são envolvidas em aspas e escapadas
3. **Timestamps ISO**: Campos de data (`created_at`, `updated_at`) são strings ISO 8601, geralmente com `:` então são envolvidos em aspas
4. **FrontmatterGenerator**: Todos os handlers (Story e World entities) agora usam o `FrontmatterGenerator` centralizado (migração concluída em 2025-01-09)
5. **Consistência**: Todos os handlers usam o mesmo gerador, garantindo formato consistente de frontmatter e tags

---

## Status da Migração

✅ **Completo** (2025-01-09):
- Todos os World handlers migrados para usar `FrontmatterGenerator`
- Todos os Story handlers (via FileManager) já usavam método similar, agora padronizado
- Tags geradas consistentemente: `story-engine/{entity-type}`, `story/{story-name}`, `world/{world-name}`, `date/{YYYY}/{MM}/{DD}`

---

## Próximos Passos (Opcional)

- [ ] Adicionar campo `tenant_id` em World entities que não têm (atualmente apenas Archetype e Trait têm)
- [ ] Considerar adicionar campo `synced_at` para tracking de última sincronização
- [ ] Verificar se testes precisam ser atualizados para refletir uso do FrontmatterGenerator

