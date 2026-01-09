# Status da Implementação SQLite - Story Engine

**Data:** 2025-01-XX  
**Última Atualização:** Sessão de implementação de funcionalidades Offline Mode e estrutura de testes SQLite

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
- [x] `scene_reference_repository.go` - Scene References
- [x] `beat_repository.go` - Beats
- [x] `content_block_repository.go` - Content blocks (prose/images/video/etc)
- [x] `content_anchor_repository.go` - Referências de content blocks

#### World Building Domain
- [x] `world_repository.go` - Worlds
- [x] `location_repository.go` - Locations com hierarquia (CTEs recursivos)
- [x] `character_repository.go` - Characters
- [x] `character_trait_repository.go` - Junction table Character-Trait
- [x] `artifact_repository.go` - Artifacts
- [x] `artifact_reference_repository.go` - Artifact References
- [x] `event_repository.go` - Events
- [x] `event_character_repository.go` - Junction table Event-Character
- [x] `event_location_repository.go` - Junction table Event-Location
- [x] `event_artifact_repository.go` - Junction table Event-Artifact
- [x] `faction_repository.go` - Factions com hierarquia
- [x] `faction_reference_repository.go` - Faction References
- [x] `lore_repository.go` - Lores com hierarquia
- [x] `lore_reference_repository.go` - Lore References
- [x] `trait_repository.go` - Traits
- [x] `archetype_repository.go` - Archetypes
- [x] `archetype_trait_repository.go` - Junction table Archetype-Trait

## ⏳ O QUE AINDA PRECISA SER IMPLEMENTADO

### 1. Repositórios World Building - Core (PRIORIDADE ALTA)
✅ **COMPLETO** - Todos os repositórios Core foram implementados

### 2. Repositórios World Building - Extended (PRIORIDADE MÉDIA)
✅ **COMPLETO** - Todos os repositórios Extended foram implementados

### 3. Scene References (PRIORIDADE BAIXA)
✅ **COMPLETO** - Scene reference repository foi implementado

### 4. Funcionalidades Offline Mode (PRIORIDADE ALTA)
✅ **COMPLETO** - Todas as funcionalidades de Offline Mode foram implementadas

#### Default Tenant Setup ✅
- [x] Criar lógica para criar tenant padrão automaticamente
  - UUID fixo para tenant offline: `00000000-0000-0000-0000-000000000001`
  - Auto-criação na inicialização do modo offline
  - Localização: `main-service/internal/platform/tenant/offline_setup.go`

#### Offline Middleware ✅
- [x] Criar middleware que injeta tenant padrão no context
  - Localização: `main-service/internal/transport/http/middleware/offline_tenant.go`
  - Injeta o tenant padrão em todas as requisições

#### Entry Point Offline ✅
- [x] Criar `cmd/api-offline/main.go`
  - Inicializa SQLite database
  - Cria tenant padrão se não existir
  - Configura middleware de tenant offline
  - Inicializa apenas repositórios de Story + World Building
  - Inicializa HTTP server (sem gRPC)
  - NÃO inicializa repositórios de User/Membership/RPG

### 5. Testes de Integração SQLite (PRIORIDADE MÉDIA)

#### Test Helpers ✅
- [x] `test_helper.go` - SetupTestSQLiteDB, SetupTestDBFile, applyMigrations, TruncateTables

#### Repositórios com Testes
- [x] `tenant_repository_test.go` - TenantRepository (exemplo completo)
- [x] `world_repository_test.go` - WorldRepository
- [x] `location_repository_test.go` - LocationRepository (incluir testes de hierarquia)
- [x] `character_repository_test.go` - CharacterRepository
- [x] `character_trait_repository_test.go` - CharacterTraitRepository (junction table)
- [x] `artifact_repository_test.go` - ArtifactRepository
- [x] `artifact_reference_repository_test.go` - ArtifactReferenceRepository
- [x] `event_repository_test.go` - EventRepository
- [x] `event_character_repository_test.go` - EventCharacterRepository (junction table)
- [x] `event_location_repository_test.go` - EventLocationRepository (junction table)
- [x] `event_artifact_repository_test.go` - EventArtifactRepository (junction table)
- [x] `faction_repository_test.go` - FactionRepository (incluir testes de hierarquia)
- [x] `faction_reference_repository_test.go` - FactionReferenceRepository
- [x] `lore_repository_test.go` - LoreRepository (incluir testes de hierarquia)
- [x] `lore_reference_repository_test.go` - LoreReferenceRepository
- [x] `trait_repository_test.go` - TraitRepository
- [x] `archetype_repository_test.go` - ArchetypeRepository
- [x] `archetype_trait_repository_test.go` - ArchetypeTraitRepository (junction table)
- [x] `story_repository_test.go` - StoryRepository (incluir testes de versionamento)
- [x] `chapter_repository_test.go` - ChapterRepository
- [x] `scene_repository_test.go` - SceneRepository
- [x] `scene_reference_repository_test.go` - SceneReferenceRepository
- [x] `beat_repository_test.go` - BeatRepository
- [x] `content_block_repository_test.go` - ContentBlockRepository
- [x] `content_anchor_repository_test.go` - ContentAnchorRepository

#### Instruções para Criar Testes

1. **Criar arquivo de teste**: `{repository}_test.go` no mesmo diretório do repositório
2. **Usar build tag**: Adicionar `//go:build integration` no topo do arquivo
3. **Seguir o padrão**:
   ```go
   func Test{Repository}_{Method}(t *testing.T) {
       db, cleanup := SetupTestSQLiteDB(t)
       defer cleanup()
       
       ctx := context.Background()
       repo := New{Repository}(db)
       
       // Testes aqui
   }
   ```
4. **Testar operações CRUD básicas**:
   - Create (sucesso e constraints)
   - GetByID (existente e não existente)
   - Update
   - Delete
   - List (quando aplicável)
5. **Testar casos especiais**:
   - Hierarquia (Location, Faction, Lore - GetChildren, GetAncestors, GetDescendants)
   - Versionamento (Story - versões)
   - Junction tables (ON CONFLICT DO NOTHING)
   - Foreign keys e constraints
6. **Executar testes**: `go test -tags=integration ./internal/adapters/db/sqlite -v -run Test{Repository}`

**Referências**:
- Exemplo: `main-service/internal/adapters/db/sqlite/tenant_repository_test.go`
- Padrão Postgres: `main-service/internal/adapters/db/postgres/user_repository_test.go`

### 6. Melhorias e Ajustes (PRIORIDADE BAIXA)
- [ ] Adicionar comentários nos entry points SAAS indicando modo multi-tenant Postgres
- [x] Criar função helper para executar migrations SQLite (similar ao golang-migrate)
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
- ✅ **Repositórios Core completos** - location, character, artifact, event e todas as junction tables foram implementados
- ✅ **Repositórios Extended completos** - faction, lore, trait, archetype e todas as junction tables/references foram implementados

## 🎯 PRÓXIMOS PASSOS SUGERIDOS

1. ✅ ~~Implementar repositórios Core (location, character, artifact, event)~~ **CONCLUÍDO**
2. ✅ ~~Implementar repositórios Extended (faction, lore, trait, archetype)~~ **CONCLUÍDO**
3. ✅ ~~Implementar funcionalidades de offline mode (default tenant, middleware, entry point)~~ **CONCLUÍDO**
4. ✅ ~~Criar estrutura de testes SQLite (test_helper.go e exemplo)~~ **CONCLUÍDO**
5. ✅ ~~Criar testes de integração para repositórios SQLite (seguir lista na seção 5)~~ **CONCLUÍDO** - Todos os 25 repositórios agora possuem testes de integração completos
6. Testar o modo offline básico (testar entry point `cmd/api-offline/main.go`)
7. ✅ ~~Adicionar documentação de uso do modo offline~~ **CONCLUÍDO** - Documentação completa em `documentation/guides/offline-mode.md`

