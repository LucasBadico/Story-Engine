# Arquitetura: Apps irmãs + UI Package (Web + Obsidian)

## Objetivo

Ter **um único “app de UI” reutilizável** que roda em dois ambientes:

- **Frontend Web** (`web-app`)
- **Plugin do Obsidian** (`obsidian-plugin`)

E manter tudo como **apps irmãs** no repositório (mesmo nível), com um pacote compartilhado de UI (`ui-package`).

---

## Estrutura do repositório (apps irmãs)

```
repo/
  main-service/
  llm-gateway-service/
  obsidian-plugin/
  web-app/
  ui-package/
  shared/              (opcional, recomendado)
    contracts/         (types, DTOs, zod schemas)
    clients/           (SDKs HTTP/gRPC)
    config/            (eslint/tsconfig/prettier)
```

### Papel de cada item

- **main-service**: Core do Story Engine (HTTP/gRPC) + persistência.
- **llm-gateway-service**: gateway/roteamento de LLM, políticas, prompts, tools.
- **web-app**: app web (admin/dev/usuário final) consumindo APIs.
- **obsidian-plugin**: integração com Obsidian (views, commands, vault, settings) e consumo da mesma UI.
- **ui-package**: biblioteca de componentes + design tokens + estilos (sem dependência do Obsidian).
- **shared/** (opcional): contratos e SDKs compartilhados entre app e plugin.

---

## Princípio-chave

**UI não conhece o ambiente.**

A UI (componentes + páginas) depende apenas de **interfaces** (ports) e recebe implementações (adapters) do ambiente:

- No web: `fetch`, `localStorage`, notificações web
- No Obsidian: `app.vault`, `Notice`, `plugin.settings`, etc.

---

## Camadas (Clean-ish, pragmático)

### 1) UI Layer (`ui-package`)

- Componentes puros (React)
- Design tokens (cores, spacing, tipografia)
- CSS/Styles (tailwind ou css modules, etc.)
- Componentes não fazem:
  - `fetch`
  - chamadas Obsidian API
  - acesso a storage direto

### 2) App Layer (páginas + estado)

Existem duas abordagens. Recomendada: criar um pacote “app-core” (opcional) para evitar duplicar páginas.

**Opção A (mínimo que você pediu):**

- `web-app` contém páginas + estado
- `obsidian-plugin` importa as mesmas páginas a partir do `web-app` (não recomendado) ou duplica (pior)

**Opção B (recomendada e limpa):** Criar um pacote adicional (dentro do seu layout atual) **sem virar “app”**:

- `shared/app-core` (ou `ui-package` vira `ui-package` + `app-core` separado)

**Como você pediu só **``** e **``**:** A melhor forma ainda é:

- colocar as **páginas/fluxos reutilizáveis** dentro do `ui-package` (como `ui-package/app/*`)
- e manter `web-app` e `obsidian-plugin` só como “shells”

> Ou seja: `ui-package` vira mais um **UI Kit + App Shell**.

---

## Contratos (Ports) que precisam existir

A UI/App usa um conjunto de interfaces (ports). Exemplo:

### `StoryEngineClient`

- `listStories()`
- `getStory(id)`
- `updateStory(id, patch)`
- `listWorlds()`
- `getWorld(id)`

### `LLMGatewayClient` (se fizer sentido no client)

- `suggestBeats(sceneId)`
- `rewriteText(...)`

### `Storage`

- `get(key)` / `set(key, value)` (config, tokens, preferências)

### `Notifier`

- `success(msg)` / `error(msg)` / `info(msg)`

### `Navigator` (opcional)

- `openExternal(url)`
- `openNote(path)` (no Obsidian)

A UI recebe isso via Provider:

- `<AppProviders services={...}>`

---

## Fluxo de composição (Shell → App)

### Web (`web-app`)

1. Cria adapters:
   - `HttpStoryEngineClient(baseUrl)`
   - `BrowserStorage(localStorage)`
   - `ToastNotifier()`
2. Renderiza o App:
   - `createRoot(...).render(<StoryApp services={...} />)`

### Obsidian (`obsidian-plugin`)

1. Dentro da View do plugin, cria adapters:
   - `HttpStoryEngineClient(baseUrl)` **ou** `LocalEmbeddedClient()` se existir modo offline
   - `ObsidianStorage(plugin.settings)`
   - `ObsidianNotifier(new Notice())`
2. Monta o mesmo App dentro do container:
   - `createRoot(view.containerEl).render(<StoryApp services={...} />)`
3. No `onClose/unload`: desmonta root.

---

## Diagrama textual (alto nível)

```
                 +----------------------+
                 |     ui-package       |
                 |  components + pages  |
                 |  (no env coupling)   |
                 +----------+-----------+
                            |
           services/ports   |
                            v
      +---------------------+---------------------+
      |                                           |
+-----+---------+                         +-------+--------+
|  web-app |                         | obsidian-plugin |
|  Web Shell    |                         | Obsidian Shell  |
| adapters:     |                         | adapters:       |
| fetch/storage |                         | vault/Notice    |
+-----+---------+                         +-------+--------+
      |                                           |
      | HTTP/gRPC                                 | HTTP/gRPC (ou local)
      v                                           v
+------------------+                      +-------------------+
|   main-service   |                      | llm-gateway-serv.  |
+------------------+                      +-------------------+
```

---

## Offline (seu contexto: Obsidian sem serviço web)

Você comentou que muita gente usa Obsidian sem serviço web. Então aqui vão 3 modos (do mais simples ao mais completo):

### Modo 1 — Online obrigatório

- Plugin e web dependem de `main-service` rodando.

### Modo 2 — “Local server” empacotado

- Usuário baixa um binário (Go) que sobe HTTP local + DB embutido.
- Obsidian Plugin aponta para `http://127.0.0.1:PORT`.

### Modo 3 — Embedded (sem server)

- Plugin chama um core local (WASM/Go shared lib) + storage local.
- Mais complexo, mas remove completamente “serviço web”.

Nesta arquitetura, o Modo 2 encaixa muito bem sem mudar UI: só muda o adapter de client.

---

## Build e Dev Experience

### Desenvolvimento

- Rodar `web-app` em modo dev (Next.js) com mocks ou apontando para um `main-service` local
- Rodar `obsidian-plugin` linkado no vault e apontando para o mesmo `ui-package`
- `ui-package` com build incremental (library mode) para consumo pelos dois shells

### Next.js (no `web-app`)

- O `web-app` usa Next.js para rotas/páginas e pode servir como “demo” completa do produto fora do Obsidian.
- O `web-app` continua sendo um **shell**: ele instancia adapters (HTTP, storage, notifier) e renderiza o App/UI compartilhado.
- Preferir `app/` router do Next.js (quando fizer sentido) e manter o App compartilhado como componente client.

### Storybook (executável da UI)

Você quer o “executável” como Storybook para deixar os componentes claros.

Opções:

- Storybook dentro do `ui-package` (padrão)
- ou `apps/storybook` (se preferir separar)

Recomendação dentro do seu layout de apps irmãs:

- manter Storybook como script do `ui-package`, por ser parte do pacote.

## Checklist de decisão rápida

✅ Você vai conseguir reuso real se:

- UI não chama Obsidian API nem `fetch` direto
- Tudo que é IO entra por interfaces
- `web-app` e `obsidian-plugin` são só “shells”

⚠️ Você vai sofrer se:

- colocar lógica de Obsidian dentro de componentes
- usar caminhos/aliases diferentes sem padronizar tsconfig

---

## Próximos passos recomendados

1. Definir o `ui-package` (estrutura + tokens + estilos)
2. Definir ports (interfaces) e providers
3. Criar `web-app` como shell (adapters web)
4. Ajustar `obsidian-plugin` como shell (adapters Obsidian)
5. Adicionar Storybook no `ui-package`

---

## Anexo: exemplo de “ports” (contratos) em pseudo-código

```ts
export interface StoryEngineClient {
  listStories(): Promise<Story[]>;
  getStory(id: string): Promise<Story>;
  updateStory(id: string, patch: Partial<Story>): Promise<Story>;
}

export interface Storage {
  get<T>(key: string): Promise<T | null>;
  set<T>(key: string, value: T): Promise<void>;
}

export interface Notifier {
  success(message: string): void;
  error(message: string): void;
  info(message: string): void;
}

export type Services = {
  storyEngine: StoryEngineClient;
  storage: Storage;
  notify: Notifier;
};
```

---

Se você quiser, o próximo documento pode ser: **"Mapa de pastas e aliases TS + scripts pnpm"** (pra garantir import limpo entre apps irmãs) e **"Estratégia de build do plugin (Vite/esbuild)"**.

