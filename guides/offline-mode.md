# Guia do Modo Offline - Story Engine

## 📋 Visão Geral

O **Modo Offline** é uma versão simplificada do Story Engine que permite executar o servidor localmente usando SQLite como banco de dados, sem necessidade de configuração de PostgreSQL, gRPC, ou sistemas de autenticação. É ideal para quem deseja uma versão local com pouca complexidade.

## 🎯 Características Principais

- ✅ **Banco de dados SQLite**: Usa SQLite como banco de dados local (arquivo `.db`)
- ✅ **Tenant padrão automático**: Cria e utiliza automaticamente um tenant padrão
- ✅ **Sem autenticação**: Não requer headers de autenticação ou multi-tenant
- ✅ **HTTP apenas**: Servidor HTTP simples (sem gRPC)
- ✅ **Funcionalidades completas**: Suporta todas as funcionalidades de Story e World Building
- ❌ **Sem funcionalidades de usuário**: Não inclui repositórios de User/Membership/RPG

## 🚀 Como Executar

### Pré-requisitos

1. **SQLite3** (geralmente já incluído no sistema)

> **Nota**: Você pode usar o modo offline de duas formas:
> - **Opção 1 (Recomendado)**: Baixar o binário pré-compilado (não requer Go instalado)
> - **Opção 2**: Compilar a partir do código fonte (requer Go 1.21+ instalado)

### Opção 1: Usando Binário Pré-compilado

```bash
# Baixar o binário para sua plataforma (exemplo para Linux)
# (URL será disponibilizada quando os binários estiverem disponíveis)

# Executar o binário
./story-engine-offline

# O servidor iniciará na porta padrão 8080
```

### Opção 2: Compilar a Partir do Código

Se você preferir compilar a partir do código fonte:

```bash
# Pré-requisitos: Go 1.21+ instalado
# Instalar dependências
cd main-service
go mod download

# Executar o servidor offline
go run cmd/api-offline/main.go
```

O servidor iniciará na porta padrão **8080** (ou a porta configurada via `HTTP_PORT`).

### Com Variáveis de Ambiente

```bash
# Definir variáveis de ambiente (opcional)
export DB_PATH=./story-engine.db
export HTTP_PORT=8080

# Executar (usando binário ou código fonte)
./story-engine-offline
# ou
go run cmd/api-offline/main.go
```

## ⚙️ Configuração

### Variáveis de Ambiente

O modo offline utiliza as seguintes variáveis de ambiente (todos opcionais):

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `DB_PATH` | `./story-engine.db` | Caminho do arquivo SQLite |
| `HTTP_PORT` | `8080` | Porta do servidor HTTP |

> **Nota**: O modo offline sempre usa SQLite como banco de dados. O `DB_DRIVER` não é necessário, pois o código chama diretamente `database.NewSQLite()`.

### Exemplo de Configuração Completa

```bash
# Variáveis de ambiente (opcional)
export DB_PATH=./data/story-engine.db
export HTTP_PORT=8080
```

Ou usando um arquivo `.env` (se você tiver um loader de `.env` configurado):

```bash
# Arquivo .env (opcional)
DB_PATH=./data/story-engine.db
HTTP_PORT=8080
```

## 🏗️ Arquitetura

### O Que É Incluído

O modo offline inclui **apenas** os seguintes módulos:

#### World Building
- ✅ Worlds
- ✅ Locations (com hierarquia)
- ✅ Characters
- ✅ Character Traits
- ✅ Artifacts
- ✅ Artifact References
- ✅ Events
- ✅ Event Characters/Locations/Artifacts
- ✅ Factions (com hierarquia)
- ✅ Faction References
- ✅ Lores (com hierarquia)
- ✅ Lore References
- ✅ Traits
- ✅ Archetypes
- ✅ Archetype Traits

#### Story
- ✅ Stories (com versionamento)
- ✅ Chapters
- ✅ Scenes
- ✅ Scene References
- ✅ Beats
- ✅ Content Blocks
- ✅ Content Block References

### O Que Não É Incluído

- ❌ User Management
- ❌ Membership Management
- ❌ RPG System Management
- ❌ Authentication/Authorization
- ❌ Audit Logs (usando No-op implementation)
- ❌ gRPC Server

## 🔑 Tenant Padrão

### ID Fixo

O modo offline utiliza um **tenant padrão fixo** com:

- **ID**: `00000000-0000-0000-0000-000000000001`
- **Nome**: `default`

### Criação Automática

O tenant padrão é criado automaticamente na inicialização se não existir. O middleware `OfflineTenantMiddleware` injeta automaticamente este tenant em todas as requisições, então **não é necessário** enviar o header `X-Tenant-ID` nas requisições.

### Localização do Código

- Setup do tenant: `main-service/internal/platform/tenant/offline_setup.go`
- Middleware: `main-service/internal/transport/http/middleware/offline_tenant.go`

## 📡 API Endpoints

O modo offline expõe os mesmos endpoints HTTP da versão online (para os módulos incluídos). Todos os endpoints seguem o padrão `/api/v1/{resource}`.

### Exemplos de Endpoints Disponíveis

#### Worlds
```
POST   /api/v1/worlds
GET    /api/v1/worlds
GET    /api/v1/worlds/{id}
PUT    /api/v1/worlds/{id}
DELETE /api/v1/worlds/{id}
```

#### Locations
```
POST   /api/v1/worlds/{world_id}/locations
GET    /api/v1/worlds/{world_id}/locations
GET    /api/v1/locations/{id}
PUT    /api/v1/locations/{id}
DELETE /api/v1/locations/{id}
GET    /api/v1/locations/{id}/children
GET    /api/v1/locations/{id}/ancestors
GET    /api/v1/locations/{id}/descendants
PUT    /api/v1/locations/{id}/move
```

#### Stories
```
POST   /api/v1/stories
GET    /api/v1/stories
GET    /api/v1/stories/{id}
PUT    /api/v1/stories/{id}
POST   /api/v1/stories/{id}/clone
```

#### Health Check
```
GET    /health
```

> **Nota**: Para uma lista completa de endpoints, consulte a documentação da API REST ou o código em `main-service/cmd/api-offline/main.go` (linhas 241-372).

## 🔄 Migrações do Banco de Dados

As migrações SQLite são aplicadas automaticamente na inicialização. Os arquivos de migração estão localizados em:

```
main-service/internal/adapters/db/sqlite/migrations/
```

As migrações incluem:
- `001_create_tenants.up.sql` - Tabela de tenants
- `002_create_world_tables.up.sql` - Tabelas de world building
- `003_create_story_tables.up.sql` - Tabelas de story
- `004_create_event_tables.up.sql` - Tabelas de eventos
- `005_create_faction_lore_tables.up.sql` - Tabelas de factions e lores
- `006_create_scene_references.up.sql` - Tabelas de referências de scenes

## 💡 Exemplos de Uso

### Criar um World

```bash
curl -X POST http://localhost:8080/api/v1/worlds \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Meu Mundo de Fantasia",
    "description": "Um mundo épico para minhas histórias",
    "is_public": false
  }'
```

### Criar uma Location

```bash
curl -X POST http://localhost:8080/api/v1/worlds/{world_id}/locations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Floresta Encantada",
    "description": "Uma floresta mágica cheia de criaturas",
    "type": "forest"
  }'
```

### Criar uma Story

```bash
curl -X POST http://localhost:8080/api/v1/stories \
  -H "Content-Type: application/json" \
  -d '{
    "title": "A Jornada do Herói",
    "status": "draft",
    "world_id": "{world_id}"
  }'
```

### Listar Todos os Worlds

```bash
curl http://localhost:8080/api/v1/worlds
```

## 🔍 Diferenças do Modo Online

| Aspecto | Modo Offline | Modo Online |
|---------|--------------|-------------|
| **Banco de Dados** | SQLite (arquivo local) | PostgreSQL |
| **Tenant** | Tenant padrão fixo | Multi-tenant (header `X-Tenant-ID`) |
| **Autenticação** | Não requerida | Requerida (JWT/Auth) |
| **Protocolo** | HTTP apenas | HTTP + gRPC |
| **Módulos** | Story + World Building | Todos (incluindo User/Membership/RPG) |
| **Audit Logs** | No-op (não persiste) | Persiste em banco |
| **Uso** | Uso local/Standalone | Produção/Multi-usuário |

## 📁 Estrutura de Arquivos

### Arquivo Principal
- `main-service/cmd/api-offline/main.go` - Entry point do modo offline

### Componentes Relacionados
- `main-service/internal/platform/tenant/offline_setup.go` - Setup do tenant padrão
- `main-service/internal/transport/http/middleware/offline_tenant.go` - Middleware de tenant
- `main-service/internal/adapters/db/sqlite/` - Repositórios SQLite
- `main-service/internal/adapters/db/sqlite/migrations/` - Migrações SQLite

## ⚠️ Limitações e Considerações

### Limitações

1. **Sem Multi-tenancy**: Apenas um tenant (o padrão) é suportado
2. **Sem Autenticação**: Não há verificação de identidade do usuário
3. **Sem Funcionalidades de Usuário**: Módulos User/Membership/RPG não estão disponíveis
4. **Audit Logs**: Não são persistidos (usa No-op implementation)
5. **Escalabilidade**: SQLite não é adequado para uso em produção com múltiplos clientes simultâneos
6. **Backup**: Backup manual do arquivo `.db` é necessário

### Quando Usar

✅ **Ideal para:**
- Uso local/pessoal com pouca complexidade
- Uso standalone
- Demonstrações rápidas
- Aprendizado da API

❌ **Não recomendado para:**
- Produção com múltiplos usuários
- Requisitos de escalabilidade
- Ambientes que exigem multi-tenancy
- Sistemas que requerem autenticação

## 🛠️ Troubleshooting

### Problema: "failed to open SQLite database"

**Solução**: Verifique as permissões do diretório onde o arquivo `.db` será criado.

```bash
# Dar permissões de escrita
chmod 755 $(dirname $DB_PATH)
```

### Problema: "failed to apply migrations"

**Solução**: Se estiver usando o binário pré-compilado, o binário já inclui as migrações. Se estiver compilando a partir do código, verifique se as migrações estão presentes em `main-service/internal/adapters/db/sqlite/migrations/`.

### Problema: Porta já em uso

**Solução**: Use uma porta diferente:

```bash
export HTTP_PORT=8081
./story-engine-offline
# ou
go run cmd/api-offline/main.go
```

### Problema: Banco de dados corrompido

**Solução**: Remova o arquivo `.db` e reinicie o servidor (as migrações serão aplicadas novamente):

```bash
rm ./story-engine.db
./story-engine-offline
# ou
go run cmd/api-offline/main.go
```

## 📚 Recursos Adicionais

- [Documentação de Status SQLite](../../STATUS_SQLITE_IMPLEMENTATION.md) - Status da implementação SQLite
- [API REST Quick Reference](../REST_API_Quick_Reference.md) - Referência rápida da API
- Código fonte: `main-service/cmd/api-offline/main.go`

## 🤝 Contribuindo

Para adicionar novas funcionalidades ao modo offline:

1. Adicione os repositórios SQLite necessários em `main-service/internal/adapters/db/sqlite/`
2. Inicialize os repositórios em `cmd/api-offline/main.go`
3. Configure os use cases e handlers correspondentes
4. Registre as rotas HTTP apropriadas
5. Atualize esta documentação

---

**Última atualização**: Janeiro 2025

