# Suporte a SQLite e Postgres no Story Engine

## Contexto
O Story Engine é um backend web escrito em Go, integrado ao Obsidian, seguindo **Hexagonal Architecture**. O core da aplicação não conhece detalhes de infraestrutura, apenas **interfaces (ports)**. Atualmente existe um adapter `db/postgres` e queremos adicionar suporte a **SQLite** para viabilizar um modo *local-first* (um único binário, zero dependências externas).

Este documento descreve como suportar **SQLite e Postgres** simultaneamente, escolhendo o banco **em runtime ou em build**, sem violar a arquitetura.

---

## Objetivos

- Permitir dois modos de execução:
  - **Local-first**: SQLite (arquivo `.db`, ideal para usuários do Obsidian offline)
  - **Server / Pro**: Postgres (concorrência, pgvector, multi-tenant pesado)
- Manter o **core totalmente desacoplado** de SQL/driver
- Reaproveitar os mesmos **use cases** e **handlers**
- Garantir operações transacionais (ex: `CloneStoryTx`) em ambos os bancos

---

## Princípios de Arquitetura

1. **Core depende apenas de interfaces**
   - Nada de `database/sql`, `*sql.DB` ou drivers no core

2. **Um adapter por banco**
   - `internal/adapters/db/postgres`
   - `internal/adapters/db/sqlite`

3. **Mesmas interfaces, implementações diferentes**
   - Ambos implementam os mesmos repositórios e gerenciador de transações

4. **Escolha do banco fora do core**
   - Decisão ocorre no `main` (entrypoint)

---

## Estrutura de Pastas

```
internal/
  core/
    ports/
      repositories.go
      tx.go
    usecases/
      create_story.go
      clone_story_tx.go

  adapters/
    db/
      postgres/
        repo_story.go
        repo_chapter.go
        tx.go
        migrations/
      sqlite/
        repo_story.go
        repo_chapter.go
        tx.go
        migrations/

cmd/
  story-engine/
    main.go
```

---

## Interfaces (Ports) no Core

### Repositórios

O core define interfaces como:

- `StoryRepository`
- `ChapterRepository`
- `SceneRepository`
- `BeatRepository`
- etc.

Essas interfaces vivem em `internal/core/ports`.

### Transações (Porta Crítica)

Para suportar operações complexas (ex: clonar uma história inteira), o core define:

- `Tx` — representa uma transação ativa
- `TxManager` — responsável por iniciar/commit/rollback

O core **não sabe** se a transação é SQLite ou Postgres.

---

## Padrão Recomendado: Repositórios Atrelados à Transação

Para manter consistência, durante uma transação todos os repositórios devem operar sobre a **mesma conexão/tx**.

### Abordagem recomendada (sqlc-friendly)

- Definir uma interface `DBTX` (padrão do sqlc), implementada por `*sql.DB` e `*sql.Tx`
- Cada adapter cria repositórios a partir desse `DBTX`

Assim:
- Fora de transação → repos usam `*sql.DB`
- Dentro de transação → repos usam `*sql.Tx`

O core não percebe diferença.

---

## Adapters de Banco

### Adapter Postgres

Responsabilidades:
- Conectar via DSN
- Rodar migrations Postgres
- Implementar repositórios
- Implementar `TxManager`

Pode usar:
- `database/sql`
- `pgx` ou driver compatível
- `pgvector` (futuro)

### Adapter SQLite

Responsabilidades:
- Criar arquivo `.db` (ex: `~/.story-engine/story.db`)
- Rodar migrations SQLite
- Implementar os mesmos repositórios
- Implementar `TxManager`

Observações importantes:
- UUIDs normalmente armazenados como `TEXT`
- JSON armazenado como `TEXT`
- Constraints precisam ser explícitas (SQLite é permissivo)

---

## Configuração via Runtime (Recomendado)

### Variáveis / Config

- `DB_DRIVER=sqlite | postgres`
- `DB_DSN=...` (Postgres)
- `DB_PATH=...` (SQLite)
- `DATA_DIR=...`

### Fluxo no `main.go`

1. Carregar config
2. Ler `DB_DRIVER`
3. Inicializar adapter correspondente
4. Injetar repositórios e `TxManager` no app

Exemplo de regra:
- `sqlite` → cria `DATA_DIR/story-engine.db`
- `postgres` → conecta via DSN

---

## Build Flags (Opcional)

Além do runtime switch, é possível usar **build tags** para gerar binários diferentes:

- `-tags sqlite`
- `-tags postgres`

Isso permite:
- Evitar dependências desnecessárias
- Criar um binário *ultra-portável* para SQLite

### Quando usar

- Quando quiser distribuir dois binários distintos
- Quando quiser evitar CGO em ambientes específicos

---

## Migrations

Devido às diferenças de SQL, mantenha migrations separadas:

```
internal/adapters/db/postgres/migrations
internal/adapters/db/sqlite/migrations
```

Cada adapter é responsável por rodar suas próprias migrations no startup.

---

## Diferenças Importantes entre SQLite e Postgres

### UUID
- Postgres: tipo nativo `uuid`
- SQLite: `TEXT`

### JSON
- Postgres: `jsonb`
- SQLite: `TEXT`

### Enum
- Postgres: enum nativo ou `TEXT + constraint`
- SQLite: `TEXT` (validação no app)

**Regra geral:**
> O core valida regras de domínio; o banco apenas persiste.

---

## Embeddings e Features Avançadas

SQLite **não substitui pgvector**.

Estratégia recomendada:
- Definir uma porta no core: `EmbeddingStore`
- Adapter Postgres → pgvector
- Adapter SQLite → implementação simples ou feature desativada

Isso permite evoluir sem quebrar o modo local.

---

## Fluxo de Inicialização (Resumo)

1. `main` carrega config
2. Factory escolhe adapter (`sqlite` ou `postgres`)
3. Adapter inicializa DB + migrations
4. App monta use cases
5. Servidor HTTP/gRPC sobe

---

## Recomendação Final

- Use **runtime switch** como padrão (`DB_DRIVER`)
- SQLite como default para Obsidian / local-first
- Postgres para server e usuários avançados
- Core permanece limpo, testável e independente

Esse modelo mantém a Hexagonal Architecture intacta e permite distribuir o Story Engine como **um único binário pronto para uso**, sem sacrificar a escalabilidade futura.

