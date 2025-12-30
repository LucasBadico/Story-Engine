# AI Story Engine — Entidades, Multitenancy, Versionamento e Config LLM

## Objetivo
Definir o modelo de dados e os principais contratos de API para um sistema de storytelling assistido por LLM, com:
- **Multitenancy (workspace/tenant)**
- **Histórias e estrutura narrativa** (story → chapter → scene → beat)
- **Versionamento por clone** (grafo de versões)
- **Configuração dinâmica de LLM por workspace** (provider/model/apiKey/url)

---

## Decisões de arquitetura

### Multitenancy
- A unidade de isolamento é o **Workspace/Tenant**.
- Usuários pertencem a um ou mais workspaces via **Membership**.
- Chaves de API (do sistema) podem ser gerenciadas por workspace com auditoria.

### Versionamento de história (clone-everything)
- **Cada versão é uma `story` completa e independente**.
- Criar versão = **clonar** a história inteira (capítulos, cenas, beats, personagens, world rules, etc.).
- O histórico e o grafo são mantidos por **`root_story_id`** (grupo) e **`previous_story_id`** (pai principal).
- O autor pode criar forks e manter versões paralelas.

> Trade-off assumido conscientemente: copiar tudo aumenta volume, mas simplifica consistência, rollback e edição. Otimizações (dedupe/copy-on-write) podem ser introduzidas depois sem quebrar o modelo mental.

### Configuração dinâmica de LLM por workspace
- Cada workspace pode ter **um perfil ativo** e **vários perfis salvos**.
- Um perfil define provider + modelo(s) + URL + apiKey + defaults (temperature, max_tokens, etc.).
- Segredos (apiKey) devem ficar **criptografados** (idealmente via entidade `secret`).

---

## Entidades

## 1) Identidade, Acesso e Multitenancy

### `tenant` (workspace)
Representa o espaço isolado (multi-tenant).
- `id`
- `name`
- `created_at`, `created_by`
- `active_llm_profile_id` (FK → `workspace_llm_profile.id`, opcional)

### `user`
- `id`
- `email`
- `name`
- `status`
- `created_at`

### `membership`
Liga `user` ↔ `tenant`.
- `id`
- `tenant_id` (FK)
- `user_id` (FK)
- `role` (`owner|admin|editor|viewer`)
- `status`
- `created_at`

### `api_key`
Chave do **produto** (para integrar plugin Obsidian/clients).
- `id`
- `tenant_id` (FK)
- `name`
- `key_hash` (nunca armazenar plaintext)
- `scopes` (jsonb ou enum)
- `expires_at`, `revoked_at`
- `created_by_user_id`
- `last_used_at`

### `audit_log`
Registro de ações.
- `id`
- `tenant_id`
- `actor_user_id` (nullable para ações por apikey)
- `action`
- `entity_type`, `entity_id`
- `metadata` (jsonb)
- `created_at`

---

## 2) Produto Narrativo (história “de fato”)

### `story`
Cada `story` é um nó do grafo (versão completa).
- `id`
- `tenant_id` (FK)
- `title`
- `status` (`draft|published|archived`)
- `version_number` (int, **somente UI**, não chave lógica)
- `root_story_id` (FK → `story.id`)
- `previous_story_id` (FK → `story.id`, nullable)
- `created_by_user_id`
- `created_at`, `updated_at`

**Regras:**
- A história original: `previous_story_id = NULL`.
- Todas as versões do mesmo grupo compartilham `root_story_id`.

### `chapter`
- `id`
- `story_id` (FK)
- `number` (int)
- `title`
- `status`
- `created_at`, `updated_at`

### `scene`
- `id`
- `story_id` (FK) *(opcional, mas facilita filtros por story)*
- `chapter_id` (FK)
- `order` (int)
- `pov_character_id` (FK → `character.id`, nullable)
- `location_id` (FK → `location.id`, nullable)
- `time_ref` (texto / enum)
- `goal` (texto)
- `created_at`, `updated_at`

### `beat`
- `id`
- `scene_id` (FK)
- `order` (int)
- `type` (enum: `setup|turn|reveal|conflict|climax|resolution|hook` etc.)
- `intent` (texto)
- `outcome` (texto)
- `created_at`, `updated_at`

### `prose_block` (texto escrito)
Separar texto de `scene` ajuda a evoluir formatos e alternativas.
- `id`
- `scene_id` (FK)
- `kind` (`final|alt_a|alt_b|cleaned|localized`)
- `content` (markdown/text)
- `word_count`
- `created_at`, `updated_at`

---

## 3) Personagens e Mundo (por versão)

### `character`
- `id`
- `story_id` (FK)
- `name`
- `role` (ex: `protagonist|support|antagonist`)
- `traits` (jsonb)
- `voice_profile` (jsonb)
- `bio` (texto)

### `character_relation`
- `id`
- `story_id` (FK)
- `a_character_id` (FK)
- `b_character_id` (FK)
- `type` (ex: `ally|enemy|romantic|family|mentor`)
- `intensity` (int)
- `notes` (texto)

### `location`
- `id`
- `story_id` (FK)
- `name`
- `description`
- `rules` (jsonb)

### `world_rule`
- `id`
- `story_id` (FK)
- `key`
- `value` (texto/jsonb)
- `priority`

---

## 4) Versionamento e Grafo

### Versionamento (clone)
- Criar versão clona **tudo** de uma `story` para uma nova `story`.
- A nova story recebe:
  - `previous_story_id = <id da story clonada>`
  - `root_story_id = <root da original>`
  - `version_number` atualizado (apenas UI)

### `story_edge` (opcional, para grafo completo/merge)
Se quiser múltiplos pais e merges.
- `from_story_id`
- `to_story_id`
- `edge_type` (`clone|fork|merge|import`)
- `created_at`, `created_by`

### `story_origin` (opcional, para “promote/detach”)
Para converter uma versão em nova história sem perder proveniência.
- `story_id` (nova história)
- `origin_story_id` (versão antiga)
- `origin_root_story_id` (grupo antigo)
- `origin_type` (`promoted`)
- `created_at`

---

## 5) Configuração dinâmica de LLM por workspace

### `llm_provider`
Catálogo de provedores suportados.
- `id`
- `key` (ex: `openai`, `anthropic`, `azure_openai`, `ollama`, `openrouter`)
- `name`
- `supports` (jsonb: chat, embeddings, tools, vision)

### `workspace_llm_profile`
Perfis configuráveis por workspace.
- `id`
- `tenant_id` (FK)
- `name` (ex: `Produção`, `Barato`, `Local`)
- `provider_id` (FK)
- `base_url` (nullable)
- `api_key_secret_id` (FK → `secret.id`)
- `default_model`
- `embedding_model` (opcional)
- `temperature` (default)
- `max_tokens` (default)
- `extra` (jsonb: headers, azure deployment, api_version, org, project, etc.)
- `created_at`, `created_by`, `updated_at`

### `secret`
Armazena segredos criptografados.
- `id`
- `tenant_id`
- `kind` (`llm_api_key` etc.)
- `ciphertext` (bytea)
- `key_id` (identificador da chave de criptografia)
- `created_at`, `rotated_at`, `revoked_at`

---

## APIs (alto nível)

## 1) Workspace / Auth
- `POST /auth/login`
- `GET /me`
- `POST /tenants`
- `POST /tenants/{id}/members`
- `POST /tenants/{id}/apikeys`

## 2) Story (produto)
- `POST /stories`
- `GET /stories/{id}`
- `POST /stories/{id}/chapters`
- `POST /chapters/{id}/scenes`
- `POST /scenes/{id}/beats`
- `PUT /scenes/{id}/prose` (ou `POST /prose-blocks`)

## 3) Versionamento
- `POST /stories/{id}/versions` (clone completo)
- `GET /stories/{id}/versions` (linha do tempo)
- `GET /stories/{id}/graph` (se usar `story_edge`)
- `POST /stories/{id}/promote` (gera nova raiz + `story_origin`)

## 4) LLM Profiles
- `POST /tenants/{id}/llm-profiles`
- `GET /tenants/{id}/llm-profiles`
- `PATCH /llm-profiles/{profile_id}`
- `POST /tenants/{id}/llm-profiles/{profile_id}/activate`
- `POST /llm-profiles/{profile_id}/test` (opcional, para UX)

---

## Notas de implementação

### Clone transacional
A operação de versionamento deve rodar em **uma transação**:
1) criar `story` nova
2) copiar `chapters` (gerando novo mapeamento de IDs)
3) copiar `scenes` (mapeando IDs do chapter)
4) copiar `beats` (mapeando IDs da scene)
5) copiar `characters/locations/world_rules` e quaisquer outros recursos
6) copiar `prose_block`

### Índices recomendados
- `story(tenant_id)`
- `story(root_story_id)`
- `story(previous_story_id)`
- `chapter(story_id, number)`
- `scene(chapter_id, order)`
- `beat(scene_id, order)`
- `workspace_llm_profile(tenant_id)`

### Segurança
- `api_key.key_hash` (hash seguro + salt)
- `secret.ciphertext` (criptografia at-rest)
- `audit_log` para ações sensíveis (troca de llm profile, revogação de keys, publish/promote)

---

## Próximos passos
1) Transformar este documento em **DDL Postgres** (tabelas/FKs/índices).
2) Definir contratos **gRPC + HTTP** para os endpoints principais.
3) Implementar os fluxos críticos:
   - `CloneStoryTx`
   - `PromoteStoryTx`
   - CRUD + Activate de `workspace_llm_profile`

