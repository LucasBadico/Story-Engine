# LangGraph como Orquestrador do Entity Extractor

## Contexto

O Story Engine já possui:

- entidades ingeridas (world knowledge)
- busca semântica funcionando (embeddings via Ollama, com possibilidade de troca futura)
- integração com Obsidian (seleção de texto → API)

O próximo passo é construir um **Entity Extractor** que:

1. Recebe um trecho selecionado
2. Sugere entidades relevantes do *World* (characters, traits, archetypes, locations, relations, etc.)
3. Decide entre `create`, `update`, `link_only` ou `ignore`
4. Retorna evidências e patches aplicáveis
5. Permite aprovação humana via UI

Este documento descreve como usar **LangGraph** para orquestrar esse fluxo de forma escalável, inspecionável e extensível.

---

## Por que LangGraph?

LangGraph é adequado porque o problema **não é linear**:

- Nem sempre os mesmos tipos de entidades aparecem
- Algumas decisões dependem de buscas intermediárias
- Ambiguidades exigem *human-in-the-loop*
- O estado precisa ser compartilhado entre etapas

LangGraph oferece:

- Grafo explícito (nós + edges condicionais)
- Estado tipado e compartilhado
- Branching e loops
- Checkpoints e interrupções
- Separação clara entre orquestração e lógica de domínio

Importante: **LangGraph não precisa dominar o domínio**. Ele apenas coordena funções puras do seu sistema.

---

## Visão Geral do Pipeline

```
[Input]
   ↓
[Router LLM]
   ↓ (tipos prováveis)
[Extract Candidates]
   ↓
[Retrieve Matches]
   ↓
[Decide Actions]
   ↓
[Validate / Normalize]
   ↓
[Human Approval (opcional)]
   ↓
[Output Suggestions]
```

---

## Conceito Central: World Entities

Neste estágio **não lidamos com Story structure** (story/chapter/scene/beat), pois isso já está implícito no arquivo do Obsidian.

O foco é exclusivamente **World Knowledge**:

- Character
- Trait
- Archetype
- Location
- Relation
- Faction
- &#x20;Artefact
- Event

Essas entidades podem hoje estar tecnicamente associadas a `story_id`, mas conceitualmente pertencem ao *World layer*.

---

## Entity Registry (Base do Sistema)

Antes de qualquer LLM, existe um **Entity Registry** — um catálogo declarativo dos tipos possíveis.

Cada tipo descreve:

- `entity_type`
- campos principais
- campos obrigatórios
- heurísticas de identificação
- exemplos mínimos

Esse registry:

- é gerado a partir do seu modelo de domínio
- é passado (resumido) para os prompts
- garante consistência e evita alucinações

A LLM **não aprende o banco**, ela respeita o contrato.

---

## Nó 1 — Router LLM

### Objetivo

Inferir **quais tipos de entidades podem existir** no trecho selecionado.

### Pipeline do Passo Zero (Pré-Processamento por Parágrafo/Chunk)

Antes do Router, o texto é segmentado para garantir robustez em textos longos e
para manter rastreabilidade por trecho nas fases seguintes.

**Segmentação**
- **Primeiro nível**: split por parágrafos (linhas vazias).
- **Se o parágrafo for grande**: quebrar por sentenças até um limite.
- **Fallback**: janela fixa com pequeno overlap (10–15%) para trechos ainda longos.

**Metadados por trecho**
- `paragraph_id`
- `chunk_id`
- `start_offset` / `end_offset`
- `text`

**Execução**
- Rodar o Router em **todos os chunks**.
- Guardar `candidates` por chunk e **agregar por parágrafo**:
  - `confidence[type] = max(confidence)`
  - `occurrences[type] = count_chunks_with_type`

**Regra de parada**
- **Não parar entre parágrafos**.
- **Pode parar dentro de um parágrafo** somente se ele foi quebrado em múltiplos chunks
  e já encontrou todos os tipos possíveis ou atingiu um limite de chunks.

**Saída do Passo Zero**
- `paragraph_results[]` com:
  - `chunks[]` (texto + offsets)
  - `chunk_candidates[]`
  - `paragraph_candidates[]`

Esses resultados alimentam diretamente as fases de extração e deduplicação.

### Input

- Texto selecionado
- Contexto leve (ex.: título do world/story)
- Lista de tipos possíveis (do registry)

### Output

```json
{
  "candidates": [
    {"type": "character", "confidence": 0.86, "why": "Nomes próprios com ações"},
    {"type": "relation", "confidence": 0.64, "why": "Interação entre duas pessoas"}
  ]
}
```

### Observações

- Nenhuma entidade é criada aqui
- Confiança baixa em todos → pipeline encerra cedo
- Atua como *dynamic type gating*

---

## Nó 2 — Extract Candidates

### Objetivo

Extrair **candidatos brutos** para cada tipo selecionado.

### Input

- Texto
- Tipos habilitados pelo router
- Registry

### Output

```json
{
  "character": [
    {"name": "Efrain", "evidence": "Efrain saiu batendo a porta"}
  ],
  "trait": [
    {"label": "impulsivo", "evidence": "saiu batendo a porta"}
  ]
}
```

### Observações

- Não decide update/create
- Extrai apenas o que o texto *afirma*
- Sempre retorna evidências

---

## Nó 3 — Retrieve Matches (Tool Node)

### Objetivo

Buscar entidades **já existentes** no seu sistema.

### Implementação

- Tool que chama o backend Go
- Combina:
  - match exato (chave natural)
  - busca semântica (embeddings)
  - opcional: FTS

### Output

```json
{
  "Efrain": [
    {"id": "char_123", "name": "Efrain", "score": 0.91},
    {"id": "char_991", "name": "Efraim", "score": 0.72}
  ]
}
```

---

## Nó 4 — Decide Actions LLM

### Objetivo

Decidir o que fazer com cada candidato:

- `create`
- `update`
- `link_only`
- `ignore`

### Input

- Candidatos extraídos
- Matches existentes
- Regras de decisão

### Output

```json
{
  "suggestions": [
    {
      "entity_type": "character",
      "action": "update",
      "target_id": "char_123",
      "patch": [{"op":"add","path":"/traits/impulsivo","value":true}],
      "confidence": 0.78
    }
  ]
}
```

---

## Nó 5 — Validate / Normalize (Opcional)

Regras duras e normalização:

- não criar entidade sem campo obrigatório
- normalizar traits/archetypes (vocabulário controlado)
- marcar `needs_review` se ambíguo

Esse nó **não usa LLM**.

---

## Nó 6 — Human-in-the-Loop (Opcional)

LangGraph permite **interrupção** do grafo:

- pipeline pausa
- sugestões são enviadas ao Obsidian
- usuário aprova, edita ou rejeita
- execução pode continuar ou finalizar

Esse padrão encaixa perfeitamente no fluxo do plugin.

---

## State do LangGraph (Resumo)

Campos principais:

- `selection_text`
- `context`
- `route`
- `candidates_by_type`
- `matches_by_candidate`
- `suggestions`
- `errors`
- `telemetry`

O state é incremental e transparente.

---

## Boas Práticas

- Nós devem ser **funções puras**
- Lógica de domínio não vive no LangGraph
- LangGraph apenas orquestra
- Tools chamam seu backend existente
- Registry é a fonte da verdade

---

## MVP Recomendado

1. Tipos: `character`, `location`, `trait`, `relation`
2. Grafo com 4 nós (router → extract → retrieve → decide)
3. Sem checkpoint no início
4. UI inline no Obsidian (sem modais)

---

## Evoluções Futuras

- Persistência de checkpoints
- Reprocessamento incremental
- Relações como entidades derivadas
- Versionamento de world knowledge
- Métricas de precisão por tipo

---

## Conclusão

LangGraph é uma **boa camada de orquestração** para o Entity Extractor:

- mantém o pipeline explícito
- reduz acoplamento
- habilita human-in-the-loop
- escala com complexidade narrativa

Ele complementa (não substitui) sua arquitetura atual baseada em serviços, embeddings e modelos de domínio.
