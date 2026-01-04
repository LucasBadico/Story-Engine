# Status da Implementação SQLite - Story Engine

**Data:** 2025-01-XX  
**Última Atualização:** Sessão de implementação de repositórios SQLite

## ✅ O QUE JÁ FOI IMPLEMENTADO

### 1. Infraestrutura Base ✅
- [x] Interface genérica `repositories.Tx` (substituindo `pgx.Tx`)
- [x] Postgres transaction adapter atualizado para usar interface genérica
- [x] Configuração atualizada (`config.go`) com suporte a `DB_DRIVER` e `DB_PATH`
- [x] Abstração de database genérica (`platform/database/database.go`)
- [x] Implementação SQLite da abstração (`platform/database/sqlite.go`)
- [x] Wrapper SQLite para adapters (`adapters/db/sqlite/db.go`)
- [x] Transaction repository SQLite (`adapters/db/sqlite/transaction.go`)

### 2. Migrations SQLite ✅
- [x] `001_create_tenants.up.sql` - Tabela tenants
- [x] `002_create_world_tables.up.sql` - Tabelas de world building (worlds, locations, characters, etc.)
- [x] `003_create_story_tables.up.sql` - Tabelas de story (stories, chapters, scenes, beats, content_blocks)
- [x] `004_create_event_tables.up.sql` - Tabelas de eventos
- [x] `005_create_faction_lore_tables.up.sql` - Tabelas de factions e lores
- [x] `006_create_scene_references.up.sql` - Tabelas de referências de scenes

### 3. Repositórios SQLite Implementados ✅

#### Common/Infrastructure
- [x] `tenant_repository.go` - Gerenciamento de tenants
- [x] `noop_audit_log_repository.go` - No-op para audit logs (offline mode)
- [x] `transaction.go` - Transaction repository

#### Story Domain
- [x] `story_repository.go` - Stories com versionamento
- [x] `chapter_repository.go` - Chapters
- [x] `scene_repository.go` - Scenes
- [x] `beat_repository.go` - Beats
- [x] `content_block_repository.go` - Content blocks (prose/images/video/etc)
- [x] `content_block_reference_repository.go` - Referências de content blocks

#### World Building Domain
- [x] `world_repository.go` - Worlds

## ⏳ O QUE AINDA PRECISA SER IMPLEMENTADO

### 1. Repositórios World Building - Core (PRIORIDADE ALTA)

#### Location
- [ ] `location_repository.go`
  - Interface: `repositories.LocationRepository`
  - Métodos: Create, GetByID, ListByWorld, Update, Delete, GetChildren (hierarquia)
  - Referência: `main-service/internal/adapters/db/postgres/location_repository.go`

#### Character
- [ ] `character_repository.go`
  - Interface: `repositories.CharacterRepository`
  - Métodos: Create, GetByID, ListByWorld, Update, Delete, GetChildren
  - Referência: `main-service/internal/adapters/db/postgres/character_repository.go`
- [ ] `character_trait_repository.go` (tabela junction)
  - Interface: `repositories.CharacterTraitRepository`
  - Métodos: Create, Delete, ListByCharacter, ListByTrait
  - Referência: `main-service/internal/adapters/db/postgres/character_trait_repository.go`

#### Artifact
- [ ] `artifact_repository.go`
  - Interface: `repositories.ArtifactRepository`
  - Métodos: Create, GetByID, ListByWorld, Update, Delete, CountByWorld
  - Referência: `main-service/internal/adapters/db/postgres/artifact_repository.go`
- [ ] `artifact_reference_repository.go`
  - Interface: `repositories.ArtifactReferenceRepository`
  - Métodos: Create, GetByID, ListByArtifact, ListByEntity, Delete, DeleteByArtifact, DeleteByArtifactAndEntity
  - Referência: `main-service/internal/adapters/db/postgres/artifact_reference_repository.go`

#### Event
- [ ] `event_repository.go`
  - Interface: `repositories.EventRepository`
  - Métodos: Create, GetByID, ListByWorld, Update, Delete
  - Referência: `main-service/internal/adapters/db/postgres/event_repository.go`
- [ ] `event_character_repository.go` (tabela junction)
  - Interface: `repositories.EventCharacterRepository`
  - Métodos: Create, Delete, ListByEvent, ListByCharacter
  - Referência: `main-service/internal/adapters/db/postgres/event_character_repository.go`
- [ ] `event_location_repository.go` (tabela junction)
  - Interface: `repositories.EventLocationRepository`
  - Métodos: Create, Delete, ListByEvent, ListByLocation
  - Referência: `main-service/internal/adapters/db/postgres/event_location_repository.go`
- [ ] `event_artifact_repository.go` (tabela junction)
  - Interface: `repositories.EventArtifactRepository`
  - Métodos: Create, Delete, ListByEvent, ListByArtifact
  - Referência: `main-service/internal/adapters/db/postgres/event_artifact_repository.go`

### 2. Repositórios World Building - Extended (PRIORIDADE MÉDIA)

#### Faction
- [ ] `faction_repository.go`
  - Interface: `repositories.FactionRepository`
  - Métodos: Create, GetByID, ListByWorld, Update, Delete, GetChildren
  - Referência: `main-service/internal/adapters/db/postgres/faction_repository.go`
- [ ] `faction_reference_repository.go`
  - Interface: `repositories.FactionReferenceRepository`
  - Métodos: Create, GetByID, ListByFaction, ListByEntity, Update, Delete, DeleteByFactionAndEntity, DeleteByFaction
  - Referência: `main-service/internal/adapters/db/postgres/faction_reference_repository.go`

#### Lore
- [ ] `lore_repository.go`
  - Interface: `repositories.LoreRepository`
  - Métodos: Create, GetByID, ListByWorld, Update, Delete, GetChildren
  - Referência: `main-service/internal/adapters/db/postgres/lore_repository.go`
- [ ] `lore_reference_repository.go`
  - Interface: `repositories.LoreReferenceRepository`
  - Métodos: Create, GetByID, ListByLore, ListByEntity, Update, Delete, DeleteByLoreAndEntity, DeleteByLore
  - Referência: `main-service/internal/adapters/db/postgres/lore_reference_repository.go`

#### Trait
- [ ] `trait_repository.go`
  - Interface: `repositories.TraitRepository`
  - Métodos: Create, GetByID, ListByWorld, Update, Delete
  - Referência: `main-service/internal/adapters/db/postgres/trait_repository.go`

#### Archetype
- [ ] `archetype_repository.go`
  - Interface: `repositories.ArchetypeRepository`
  - Métodos: Create, GetByID, ListByWorld, Update, Delete
  - Referência: `main-service/internal/adapters/db/postgres/archetype_repository.go`
- [ ] `archetype_trait_repository.go` (tabela junction)
  - Interface: `repositories.ArchetypeTraitRepository`
  - Métodos: Create, Delete, ListByArchetype, ListByTrait
  - Referência: `main-service/internal/adapters/db/postgres/archetype_trait_repository.go`

### 3. Scene References (PRIORIDADE BAIXA)
- [ ] `scene_reference_repository.go`
  - Interface: `repositories.SceneReferenceRepository`
  - Métodos: Create, GetByID, ListByScene, ListByEntity, Delete, DeleteByScene
  - Referência: `main-service/internal/adapters/db/postgres/scene_reference_repository.go`

### 4. Funcionalidades Offline Mode (PRIORIDADE ALTA)

#### Default Tenant Setup
- [ ] Criar lógica para criar tenant padrão automaticamente
  - UUID fixo para tenant offline (ex: `00000000-0000-0000-0000-000000000001`)
  - Auto-criação na inicialização do modo offline
  - Localização: `main-service/internal/platform/tenant/default.go` (sugestão)

#### Offline Middleware
- [ ] Criar middleware que injeta tenant padrão no context
  - Localização: `main-service/internal/transport/http/middleware/offline_tenant.go` (sugestão)
  - Deve injetar o tenant padrão em todas as requisições

#### Entry Point Offline
- [ ] Criar `cmd/api-offline/main.go`
  - Inicializa SQLite database
  - Cria tenant padrão se não existir
  - Configura middleware de tenant offline
  - Inicializa apenas repositórios de Story + World Building
  - Inicializa HTTP server (sem gRPC)
  - NÃO inicializa repositórios de User/Membership/RPG

### 5. Melhorias e Ajustes (PRIORIDADE BAIXA)
- [ ] Adicionar comentários nos entry points SAAS indicando modo multi-tenant Postgres
- [ ] Criar função helper para executar migrations SQLite (similar ao golang-migrate)
- [ ] Testes de integração para repositórios SQLite
- [ ] Documentação de uso do modo offline

## 📋 PADRÕES DE IMPLEMENTAÇÃO

### Estrutura de um Repositório SQLite

```go
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/{domain}"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

var _ repositories.{Repository}Interface = (*{Repository})(nil)

type {Repository} struct {
	db *DB
}

func New{Repository}(db *DB) *{Repository} {
	return &{Repository}{db: db}
}

// Métodos seguindo o padrão:
// - UUIDs armazenados como TEXT
// - Timestamps armazenados como TEXT (RFC3339)
// - JSON armazenado como TEXT
// - Nullable fields usando sql.NullString, sql.NullInt64, etc.
// - Conversão de UUIDs: .String() para INSERT, uuid.Parse() para SELECT
// - Conversão de timestamps: .Format(time.RFC3339) para INSERT, time.Parse() para SELECT
```

### Diferenças SQLite vs Postgres

1. **Placeholders**: `?` ao invés de `$1, $2, ...`
2. **UUIDs**: Armazenar como `TEXT`, converter com `uuid.Parse()` e `.String()`
3. **Timestamps**: Armazenar como `TEXT` (ISO8601/RFC3339), converter com `time.Parse(time.RFC3339)` e `.Format(time.RFC3339)`
4. **JSON**: Armazenar como `TEXT`, usar `json.Marshal/Unmarshal`
5. **Erros**: `sql.ErrNoRows` ao invés de `pgx.ErrNoRows`
6. **Tipos**: `*sql.Rows` e `*sql.Row` ao invés de `pgx.Rows` e `pgx.Row`

### Checklist para Cada Repositório

- [ ] Ler interface do repositório em `internal/ports/repositories/{entity}.go`
- [ ] Ler implementação Postgres em `internal/adapters/db/postgres/{entity}_repository.go`
- [ ] Ler entidade core em `internal/core/{domain}/{entity}.go`
- [ ] Criar arquivo `internal/adapters/db/sqlite/{entity}_repository.go`
- [ ] Implementar todos os métodos da interface
- [ ] Converter UUIDs e timestamps corretamente
- [ ] Implementar método `scan{Entity}s` para listas
- [ ] Testar compilação (`go build`)
- [ ] Verificar linter (`read_lints`)

## 🔄 COMO RETOMAR O TRABALHO

1. **Ler este documento** para entender o estado atual
2. **Verificar o que foi feito** listando arquivos em `main-service/internal/adapters/db/sqlite/`
3. **Escolher um repositório** da lista "O QUE AINDA PRECISA SER IMPLEMENTADO"
4. **Seguir o padrão** dos repositórios já implementados
5. **Atualizar este documento** ao completar cada repositório
6. **Marcar como completo** no TODO quando terminar um grupo

## 📝 NOTAS IMPORTANTES

- **Não criar repositórios de User/Membership/RPG** - não são necessários para modo offline
- **Todas as migrations já foram criadas** - não precisa criar novas migrations
- **Usar o padrão estabelecido** nos repositórios já criados
- **Testar compilação** após cada repositório criado
- **Priorizar repositórios Core** antes dos Extended

## 🎯 PRÓXIMOS PASSOS SUGERIDOS

1. Implementar repositórios Core (location, character, artifact, event)
2. Implementar funcionalidades de offline mode (default tenant, middleware, entry point)
3. Testar o modo offline básico
4. Implementar repositórios Extended (faction, lore, trait, archetype)
5. Adicionar testes e documentação

