# Prompt Templates Econômicos para Uso com Cursor

Este documento define **templates de prompt econômicos** para uso no Cursor, com o objetivo de **reduzir consumo excessivo de contexto (Cache Read)**, melhorar performance e manter controle consciente sobre quando o sistema deve ou não carregar contexto global do repositório.

Esses templates foram pensados especificamente para projetos **grandes, fortemente conectados e com múltiplos domínios**, como:
- Story Engine
- Worldbuilding
- RPG systems
- UI (Obsidian Plugin / Web)
- API / Backend

---

## Princípio Fundamental

> **O problema não é o modelo. É o planner de contexto.**

O Cursor decide **quanto contexto carregar antes** do modelo responder. Portanto:
- Trocar de modelo **não reduz** significativamente o Cache Read
- Escopo explícito + proibições claras **reduzem drasticamente** o consumo

---

## TEMPLATE 0 — Header Obrigatório (use sempre)

Use este bloco como **prefixo padrão** em prompts econômicos.

```text
Context policy:
- Do NOT scan or recall the full repository
- Do NOT infer from previous conversations
- Do NOT load unrelated files or entities
- Use only the context explicitly mentioned below
```

---

## TEMPLATE 1 — Conceitual Puro (Mais Econômico)

**Use quando:**
- Pensar ideias
- UX / UI conceitual
- Modelagem narrativa
- Worldbuilding teórico

```text
Context policy:
- Conceptual answer only
- No repository inspection
- No system-wide consistency required

Question:
[pergunta aqui]
```

**Características:**
- Cache Read mínimo
- Alta clareza conceitual
- Zero coerência global forçada

---

## TEMPLATE 2 — Domínio Isolado (Story / World / RPG)

**Use quando:**
- Trabalhar coerência interna de um domínio
- Evoluir worldbuilding ou RPG stats

```text
Context boundary:
- Consider ONLY the following domain: Worldbuilding
- Relevant entities: World, Location, Character, Faction
- Ignore UI, API, persistence, plugins, and infra

Rules:
- Do not assume knowledge outside this boundary
- If something is missing, ask instead of inferring

Task:
[task aqui]
```

---

## TEMPLATE 3 — UI do Obsidian (Sem Backend)

**Use quando:**
- Trabalhar UX / UI
- Refinar fluxo visual
- Pensar interações

```text
Context boundary:
- Scope: Obsidian Plugin UI
- Ignore backend, API, database, and domain rules
- Focus only on UX, layout, flow, and affordances

Constraints:
- Avoid modals unless strictly necessary
- Prefer inline editable fields
- Assume React-based rendering

Task:
[task aqui]
```

---

## TEMPLATE 4 — Backend / API Local (Sem UI)

**Use quando:**
- Refinar contratos
- Ajustar regras locais
- Pensar handlers ou lógica interna

```text
Context boundary:
- Scope: Backend API logic only
- Ignore UI, Obsidian plugin, and worldbuilding
- Do not reason about UX or presentation

Assumptions:
- API is already implemented and stable
- This is a local refinement, not a redesign

Task:
[task aqui]
```

---

## TEMPLATE 5 — Global (Caro, Uso Consciente)

**Use apenas quando:**
- Validar coerência geral
- Alinhar Story ↔ World ↔ UI ↔ API

```text
Global reasoning enabled.

System scope:
- Story
- World
- UI
- API

Goal:
[goal aqui]

Note:
- Prefer minimal context expansion
- Flag inconsistencies instead of auto-fixing
```

⚠️ **Este template gera alto consumo de contexto. Use intencionalmente.**

---

## TEMPLATE 6 — Anti Context Explosion (Emergência)

Use quando perceber que o Cursor começou a extrapolar contexto.

```text
STOP.
Reduce context usage.

Answer using first principles only.
Do not recall any system-specific details.
If unsure, state uncertainty instead of inferring.
```

---

## Boas Práticas Gerais

- Sempre declarar **escopo explícito**
- Evitar frases vagas como:
  - "considerando o sistema"
  - "no contexto do projeto"
- Preferir **domínios isolados**
- Usar raciocínio global apenas quando necessário

---

## Regra de Ouro

> **Contexto não é memória.**  
> **Contexto é uma ferramenta cirúrgica.**

Quanto maior o projeto, mais importante é controlar explicitamente o que o LLM pode ou não acessar.

---

## Observação Final

Esses templates refletem boas práticas que também devem ser aplicadas em:
- RAG systems
- Retrieval pipelines
- Prompt orchestration
- Context fencing

Ou seja: este documento não é só para o Cursor — ele é um **guia mental** para qualquer sistema LLM-driven.

