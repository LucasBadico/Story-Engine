# Obsidian Plugin UI (React) — Proposta e Boas Práticas

> Contexto: estamos construindo um plugin do Obsidian para o *Story Engine*, com entidades **Worlds**, **Stories** e um módulo opcional de **RPG/Stats**. Queremos uma UX fluida, evitando modais, com campos aparecendo inline na própria view.

---

## 1) Proposta de modelo (nível 0)

### 1.1 Entidades de topo (Level 0)
- **World**: existe fora de qualquer Story. Um World pode ser usado por várias Stories.
- **Story**: pertence a um Workspace/Tenant (conceito do backend) e pode referenciar um World.
- **RPG System / Stats**: configuração e esquema de stats que vivem no **World** (ou no nível 0), e a Story apenas **ativa** e **seleciona** o que usa.

### 1.2 Relações
- Uma **Story** pode apontar para um **World** (opcional).
- **RPG/Stats** são definidos no **World** (fonte de verdade).
- A **Story** guarda apenas:
  - `worldId` (ou none)
  - `featuresEnabled`: flags (ex.: `rpgEnabled`, `memoryEnabled`, `timelinesEnabled`…)
  - `rpgProfileId` (qual conjunto de stats do World ela usa)
  - configurações específicas *da história* (ex.: intensidade, progressão, limites) **sem duplicar schema do World**.

### 1.3 UX no Story Details
No **Story Details**, a UI exibe opções relacionadas ao World/RPG **somente quando**:
- a story tem `worldId` definido
- e a feature `rpgEnabled` está ativa

Isso mantém a tela simples e “progressiva”: primeiro o usuário escolhe o World → depois ativa features → só então aparece o resto.

---

## 2) Princípios de UX: evitar modais

### 2.1 Por que evitar modais
- Modais interrompem fluxo, criam “context switch”.
- Em apps *document-like* (Obsidian), o usuário espera editar inline.

### 2.2 Substitutos melhores que modais
- **Inline edit**: campo vira input no lugar.
- **Expandable sections / accordions**: configurações avançadas ficam “dobráveis”.
- **Side panel** (painel lateral) dentro da view: para edições maiores (ex.: schema de stats).
- **Drawer** dentro da própria view (não global): útil em telas estreitas.
- **Popover leve**: para escolhas pequenas (select, quick actions).

### 2.3 Boas práticas de “inline edit”
- **Salvar automático com debounce** para campos simples.
- **Salvar explícito** (botão) para edições de alto impacto (ex.: schema).
- Mostrar estado do dado:
  - `Saved` / `Saving…` / `Unsaved changes` / `Error`.
- **Cancel / revert** fácil (Esc ou botão).

---

## 3) Boas práticas gerais para UI no Obsidian

### 3.1 Respeitar o “Obsidian-like”
- Use fontes, espaçamentos e tons que não briguem com o tema.
- Prefira CSS com variáveis do Obsidian (quando possível).

### 3.2 Performance e responsividade
- Evitar re-render agressivo.
- Usar **virtualization** se listas ficarem grandes.
- Evitar bibliotecas de UI pesadas (MUI/Ant/Bootstrap), que “dominam” o styling.

### 3.3 Arquitetura UI
Separar:
- **View** (Obsidian): ciclo de vida, mounting/unmounting.
- **UI** (React): componentes e estado.
- **Services**: calls gRPC/HTTP, caching, storage local.

---

## 4) Por que React faz sentido aqui

Você já domina React → então o custo mental cai, e a confiabilidade sobe.

React é útil se:
- UI vai ter estado complexo (feature flags, forms, painéis, listas e filtros).
- você quer evoluir isso para um app web depois.

Trade-offs (aceitáveis):
- bundle maior
- build mais chato

---

## 5) Como usar React corretamente no Obsidian

### 5.1 Regra de ouro
**Nunca monte React fora do container da view.**

### 5.2 Padrão de montagem
- Obsidian cria uma `View` com `containerEl`.
- Você cria um `rootEl` dentro dela.
- Monta `ReactDOM.createRoot(rootEl).render(<App/>)`.
- No `onClose`/`unload`: `root.unmount()`.

### 5.3 Build recomendado
- Use **Vite** ou **esbuild**.
- Gere um bundle único para o plugin.
- Evite dynamic imports e dependências que dependam de browser APIs “especiais”.

### 5.4 Estilos
- Preferir **CSS normal** do plugin (styles.css).
- Se usar Tailwind: só se você já tiver uma estratégia de build sólida e escopo bem controlado.
- Isolar estilos em um root (ex.: `.se-root * {}`) para não vazar.

### 5.5 Estado e dados
- Estado local com `useState/useReducer`.
- Estado compartilhado:
  - Context + reducers simples
  - ou Zustand (leve) se necessário
- Cache e I/O:
  - Services com funções puras
  - hooks tipo `useQuery` caseiro (ou TanStack Query se fizer sentido)

### 5.6 Ciclo de vida (muito importante)
- View abriu → montar app
- View fechou → cancelar requests, limpar listeners, unmount

Checklist:
- [ ] remove event listeners
- [ ] abort fetch / cancel streams
- [ ] unmount react root

---

## 6) Layout sugerido para o plugin

### 6.1 Story List (esquerda)
- Lista de stories
- filtro/search
- quick actions

### 6.2 Story Details (principal)
- Header: título, status de save, ações
- Seção: World
  - select World
  - link para World details
- Seção: Features
  - toggles (RPG, Memory, etc.)
- Seção condicional: RPG
  - aparece só se `worldId` e `rpgEnabled`
  - selecionar `rpgProfileId`
  - editar valores (se existirem)

### 6.3 World Details
- Definição do World
- RPG schema editor (painel lateral/section)

---

## 7) Convenções de UI para não virar caos

### 7.1 Componentes base
- Button
- Input
- Select
- Toggle
- Section (collapsible)
- InlineEditable
- Toast/Notice (usando Notice do Obsidian)

### 7.2 Estados padrão
- Empty state (sem world / sem story)
- Loading state
- Error state (com retry)
- Saving indicator

---

## 8) Decisões já alinhadas

- Worlds existem fora de stories.
- RPG/stats pertence ao World e a Story só referencia/ativa.
- UX deve evitar modais; preferir inline edit/sections/panels.
- React será usado por familiaridade e confiabilidade.

---

## 9) Próximos passos

1) Definir schemas mínimos:
- World
- Story
- RPGProfile (dentro do World)

2) Definir telas:
- Story List
- Story Details (com seções condicionais)
- World Details

3) Definir build:
- Vite + React
- pipeline de release do plugin

4) Implementar o “shell”:
- Obsidian View monta React
- Navigation interna (sem router pesado)

---

## 10) Anexos (templates)

### 10.1 Checklist: view mounting
- Criar `rootEl`
- `createRoot`
- render `<App/>`
- `onClose`: abort/cancel + `root.unmount()`

### 10.2 Checklist: sem modais
- Preferir sections/inline edit
- Apenas popover leve para escolhas pequenas
- Side panel dentro da view para edições grandes

