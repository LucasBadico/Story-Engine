Plano: Relações e Inferências no Extract

**Status:** proposta revisada com ajustes arquiteturais para aproveitar a unificação de relações em uma única tabela de _world entities_.

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Contexto

O pipeline atual do **llm-gateway-service** executa:

-   **Phase2**: extração de entidades
-   **Phase3**: matching com entidades existentes (world/story)

O objetivo agora é **inferir e normalizar relações**, tanto:

-   entre entidades detectadas no texto
-   quanto entre entidades detectadas e **entidades existentes não explicitamente citadas**

Tudo isso mantendo compatibilidade com o **main-service**, que recentemente passou por um **refactor para centralizar todas as relações/citações de world entities em uma única tabela**.

Este plano ajusta o fluxo para **usar essa unificação como vantagem estrutural**, evitando duplicações e inconsistências.

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Objetivos

-   Descobrir relações explícitas e implícitas entre entidades.
-   Normalizar relation_type usando um **mapa oficial com mirrors**.
-   Persistir relações no **main-service** respeitando:

-   direcionalidade
-   simetria
-   mirrors automáticos

-   Emitir progresso incremental via extract/stream.
-   Retornar relações como **primeiro-cidadão** na API (não apenas side-effect).

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Princípio Arquitetural Importante (Novo)

**Toda relação é uma aresta entre duas** ``**, independentemente de origem.**

Isso implica:

-   Phase4 e Phase5 **não criam estruturas paralelas**
-   Tudo converge para o mesmo modelo de dados do main-service
-   O gateway **nunca decide schema**, apenas **propõe relações normalizadas**

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Escopo do Trabalho

LLM Gateway

Phase4 — relation_discovery

**Responsibility:** Identify **candidate relationships** based on text evidence and known context.

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Input Modes (Important)

Phase4 supports **two input modes**, chosen dynamically by the gateway based on text size.

_Mode A — Full Text_

Used when the input text is short enough to fit safely in the prompt.

**Inputs:**

-   text (full raw text)
-   entity findings (Phase2)
-   confirmed matches (Phase3)
-   context metadata

**When to use:**

-   short scenes
-   beats
-   single paragraphs

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

_Mode B — Spans + Global Summary (Recommended for long texts)_

**Inputs:**

-   global_summary: short bullet summary of the full text (3–8 bullets)
-   spans[]: relevant text excerpts with stable IDs

-   each span includes span_id, start, end, text
-   spans are selected based on proximity to entity mentions

**Rules:**

-   Every proposed relation **must reference a **`` as evidence
-   Spans are preferred over raw offsets for stability and UI traceability

**When to use:**

-   chapters
-   long scenes
-   multi-paragraph excerpts

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase4 Inputs (Common to Both Modes)

-   Entity findings from Phase2

-   stable refs: finding:<entity_type>:<index>
-   minimal fields: ref, type, name, summary
-   optional: mentions[] (list of span_ids)

-   Confirmed matches from Phase3

-   preserve original ref
-   include canonical id when resolved

-   Context

-   context_type: story | world | scene | beat
-   context_id
-   optional: pov_ref, location_ref

-   Allowed relation constraints

-   derived from per-entity relation maps (*.relation.map.json)
-   inject **only the map matching the source entity type**

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase4Input (Exact Go + JSON Payload) — (New)

**Goal:** one request payload that supports both input modes (full text OR spans+summary), and carries everything Phase4 needs.

_Go structs (recommended)_

**package** relationdiscovery  
  
_// Phase4Input is the request payload for Phase4: RELATION_DISCOVERY._  
_// It supports two modes:_  
_// - FullText: provide Text_  
_// - Spans: provide GlobalSummary + Spans_  
_// Exactly one mode should be used per request._  
_//_  
_// Refs:_  
_// - Findings:  finding:<entity_type>:<index>_  
_// - Matches: match:<entity_type>:<uuid>_  
_// - Spans: span:<n>_  
_//_  
_// Note: Phase4 returns candidates only (non-persistable). Human confirmation happens later._  
  
**type** Phase4Input **struct**  {  
RequestID string  `json:"request_id"`  _// used to correlate SSE + logs_  
Context Phase4Context `json:"context"`  
Text  Phase4TextSpec `json:"text"`  
  
EntityFindings []EntityFinding `json:"entity_findings"`  
ConfirmedMatches []ConfirmedMatch `json:"confirmed_matches,omitempty"`  
  
_// AllowedRelationsBySourceType is derived from per-entity maps (e.g. character.relation.map.json)._  
_// The gateway SHOULD inject only the relevant source_type maps to keep prompts small._  
AllowedRelationsBySourceType **map**[string]PerEntityRelationMap `json:"allowed_relations_by_source_type"`  
  
_// Optional global semantics dictionary (subset) from relation.types.json._  
_// Useful for LLM grounding but not required._  
RelationTypeSemantics **map**[string]string  `json:"relation_type_semantics,omitempty"`  
}  
  
**type** Phase4Context **struct**  {  
Type string  `json:"type"`  _// story|world|scene|beat_  
ID string  `json:"id"`  
  
_// Optional helpers (if known at this stage)_  
POVRef *string  `json:"pov_ref,omitempty"`  _// e.g. finding:character:0 or match:character:<uuid>_  
LocationRef *string  `json:"location_ref,omitempty"`  _// e.g. finding:location:1 or match:location:<uuid>_  
}  
  
_// Phase4TextSpec supports Mode A (full text) or Mode B (spans + global summary)._  
_// Exactly one of the two must be populated._  
  
**type** Phase4TextSpec **struct**  {  
Mode string  `json:"mode"`  _// full_text|spans_  
  
_// Mode A_  
Text *string  `json:"text,omitempty"`  
  
_// Mode B_  
GlobalSummary []string  `json:"global_summary,omitempty"`  _// 3–8 bullets_  
Spans []Span `json:"spans,omitempty"`  
}  
  
**type** Span **struct**  {  
SpanID string  `json:"span_id"`  _// span:<n>_  
Start int  `json:"start"`  _// byte offset in original text (optional but recommended)_  
End int  `json:"end"`  _// byte offset in original text (optional but recommended)_  
Text string  `json:"text"`  
}  
  
**type** EntityFinding **struct**  {  
Ref string  `json:"ref"`  _// finding:<type>:<index>_  
Type string  `json:"type"`  _// character|faction|location|artifact|event|..._  
Name string  `json:"name"`  
Summary string  `json:"summary"`  
Mentions []string  `json:"mentions,omitempty"`  _// list of span_ids (Mode B)_  
}  
  
_// ConfirmedMatch maps a finding ref to a canonical entity ID from main-service._  
_// Keep the finding ref so UI/ref_map remains stable._  
  
**type** ConfirmedMatch **struct**  {  
FindingRef string  `json:"finding_ref"`  _// finding:<type>:<index>_  
Match  Match `json:"match"`  
}  
  
**type** Match **struct**  {  
Ref string  `json:"ref"`  _// match:<type>:<uuid>_  
Type string  `json:"type"`  _// character|faction|..._  
ID string  `json:"id"`  _// canonical UUID_  
CanonicalName string  `json:"canonical_name"`  
Similarity float64  `json:"similarity"`  
}  
  
_// PerEntityRelationMap is the injected, per-source-type constraints._  
_// This is a compacted form derived from *.relation.map.json._  
  
**type** PerEntityRelationMap **struct**  {  
EntityType string  `json:"entity_type"`  
Version int  `json:"version"`  
Relations **map**[string]RelationConstraintSpec `json:"relations"`  _// relation_type -> constraints_  
}  
  
**type** RelationConstraintSpec **struct**  {  
PairCandidates []string  `json:"pair_candidates"`  
Description string  `json:"description"`  
Contexts []string  `json:"contexts,omitempty"`  
Signals []string  `json:"signals,omitempty"`  
AntiSignals []string  `json:"anti_signals,omitempty"`  
Constraints *RelationConstraints `json:"constraints,omitempty"`  
}  
  
**type** RelationConstraints **struct**  {  
MinConfidence float64  `json:"min_confidence"`  
AllowImplicit bool  `json:"allow_implicit"`  
RequiresEvidence bool  `json:"requires_evidence"`  
}

_JSON example — Mode A (Full Text)_

{  
"request_id":  "req-123",  
"context":  {  "type":  "scene",  "id":  "scene-uuid"  },  
"text":  {  "mode":  "full_text",  "text":  "Ari entered the Obsidian Tower..."  },  
"entity_findings":  [  
{  "ref":  "finding:character:0",  "type":  "character",  "name":  "Ari",  "summary":  "Young mage apprentice"  }  
],  
"confirmed_matches":  [],  
"allowed_relations_by_source_type":  {  
"character":  {  "entity_type":  "character",  "version":  1,  "relations":  {  "member_of":  {  "pair_candidates":  ["faction"],  "description":  "Character belongs to a group/organization."  }  }  }  
}  
}

_JSON example — Mode B (Spans + Summary)_

{  
"request_id":  "req-456",  
"context":  {  "type":  "chapter",  "id":  "chapter-uuid",  "pov_ref":  "finding:character:0"  },  
"text":  {  
"mode":  "spans",  
"global_summary":  [  
"Ari enters the Obsidian Tower.",  
"She swears loyalty to the Order of the Sun.",  
"A hidden artifact is revealed in the lower vaults."  
],  
"spans":  [  
{  "span_id":  "span:1",  "start":  0,  "end":  120,  "text":  "Ari entered the Obsidian Tower..."  },  
{  "span_id":  "span:2",  "start":  121,  "end":  240,  "text":  "Ari swore loyalty to the Order of the Sun..."  }  
]  
},  
"entity_findings":  [  
{  "ref":  "finding:character:0",  "type":  "character",  "name":  "Ari",  "summary":  "Young mage apprentice",  "mentions":  ["span:1",  "span:2"]  },  
{  "ref":  "finding:faction:1",  "type":  "faction",  "name":  "Order of the Sun",  "summary":  "A militant religious order",  "mentions":  ["span:2"]  }  
],  
"confirmed_matches":  [  
{  "finding_ref":  "finding:faction:1",  "match":  {  "ref":  "match:faction:uuid-abc",  "type":  "faction",  "id":  "uuid-abc",  "canonical_name":  "Order of the Sun",  "similarity":  0.91  }  }  
],  
"allowed_relations_by_source_type":  {  
"character":  {  "entity_type":  "character",  "version":  1,  "relations":  {  "member_of":  {  "pair_candidates":  ["faction"],  "description":  "Character belongs to a group/organization."  }  }  }  
}  
}

_Validation rules (gateway-side)_

-   text.mode == "full_text" => text.text must be set, spans must be empty.
-   text.mode == "spans" => spans must be non-empty, global_summary should be 3–8 bullets.
-   Every EntityFinding.mentions[] must reference an existing span_id (Mode B only).
-   AllowedRelationsBySourceType should include only the relevant source types (to keep prompts small).

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase4 Output (Candidate Relations)

Phase4 outputs **non-persistable candidates**:

{  
"relations":  [  
{  
"source":  {  "ref":  "finding:character:0",  "type":  "character"  },  
"target":  {  "ref":  "match:faction:uuid",  "type":  "faction"  },  
"relation_type":  "member_of",  
"polarity":  "asserted",  
"implicit":  **false**,  
"confidence":  0.78,  
"evidence":  {  
"span_id":  "span:7",  
"quote":  "Ari swore loyalty to the Order of the Sun."  
}  
}  
]  
}

Notes:

-   target.ref may point to finding:* or match:*
-   polarity: asserted | denied | uncertain
-   Evidence quotes must be short and verbatim

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase4 Prompt (English — Canonical)

You are Phase4: RELATION_DISCOVERY.  
  
Goal:  
Given the text (or spans), discovered entities and confirmed matches, propose candidate relationships between entities.  
Do NOT create entities. Do NOT assume facts not supported by the text.  
  
Hard rules:  
- Output MUST be valid JSON and match the provided schema.  
- Each relation MUST include evidence with a span_id and a short quote.  
- Prefer explicit relations; implicit relations are allowed only when strongly implied.  
- Only propose relation_type values allowed by the provided per-entity relation maps.  
- Enforce pair_candidates: if a relation type does not allow the target entity type, do not output it.  
- If support is weak or ambiguous, either omit the relation or set polarity="uncertain" with low confidence.  
  
Input provided:  
- context (type/id)  
- text OR (global_summary + spans[])  
- entity_findings[] (with refs like finding:<type>:<index>)  
- confirmed_matches[]  
- allowed_relations_by_source_type  
  
Output schema:  
{  
"relations": [  
{  
"source": { "ref": "...", "type": "..." },  
"target": { "ref": "...", "type": "..." },  
"relation_type": "...",  
"polarity": "asserted|denied|uncertain",  
"implicit": true|false,  
"confidence": 0.0-1.0,  
"evidence": { "span_id": "...", "quote": "..." }  
}  
]  
}  
  
Now produce relations.

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase4 Flow (Gateway-side) — (New)

1.  **Select input mode**

-   If len(text) is below threshold => Mode A (full_text)
-   Else => Mode B (spans + global_summary)

2.  **If Mode B, build spans**

-   Prefer paragraph-based spans
-   Ensure each span has a stable span_id (span:<n>) and offsets (start/end)
-   If Phase2 already provides mention offsets, generate spans around each mention (windowed by paragraph boundaries)

3.  **Build **``

-   Include entity_findings (Phase2)
-   Include confirmed_matches (Phase3)
-   Inject **only** the relevant allowed_relations_by_source_type maps (keep prompt small)
-   Optionally inject relation_type_semantics subset (global) for grounding

4.  **LLM call**

-   Use JSON-mode / constrained decoding when supported
-   Set conservative max output tokens (relations list is small)

5.  **Post-parse validation (strict)**

-   JSON must parse
-   relation_type must exist in allowed per-entity maps OR be custom:*
-   (source_type, relation_type) must allow target_type via pair_candidates
-   evidence.span_id must exist (Mode B)
-   clamp confidence to [0, 1]

6.  **Dedup + emit**

-   Deduplicate by (source_ref, target_ref, relation_type, context_id)
-   Emit SSE events:

-   phase.start (relation_discovery)
-   relation.candidate (per relation)
-   phase.done (relation_discovery)

7.  **Return**

-   Persist nothing (HITL)
-   Return candidate relations in API response and SSE result_relations

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase5 — relation_normalize

**Responsibility:** Transform Phase4 candidates into **persistable, UI-friendly** normalized relations.

This phase is also responsible for **comparing candidates against existing relations** already stored in the unified relation table.

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase5 Inputs

-   Candidate relations from Phase4
-   Official relation.types.json (mirror/symmetric/preferred_direction + semantics)
-   Per-entity constraints (*.relation.map.json) for strict (source_type, relation_type, target_type) validation
-   Current known entity resolution (finding:* -> canonical IDs) via confirmed matches + ref_map
-   Existing relations snapshot (optional but recommended)

-   minimal edges for the same context and/or same (source_id,target_id) pairs

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase5 Outputs

Phase5 outputs **normalized edges** plus a status used by the UI:

{  
"source":  {  "ref":  "finding:character:0",  "id":  "uuid-src"  },  
"target":  {  "ref":  "match:faction:uuid",  "id":  "uuid-tgt"  },  
"relation_type":  "member_of",  
"direction":  "source_to_target",  
"create_mirror":  **true**,  
"confidence":  0.78,  
"polarity":  "asserted",  
"implicit":  **false**,  
"evidence":  {  "span_id":  "span:2",  "quote":  "Ari swore loyalty to the Order of the Sun."  },  
"status":  "ready",  
"dedup":  {  "is_duplicate":  **false**,  "reason":  ""  },  
"summary":  "Ari is a member of Order of the Sun."  
}

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase5 Responsibilities (Explicit)

1.  **Resolve IDs / pending entities**

-   Convert finding:* and match:* refs into canonical IDs when possible.
-   If source_id or target_id cannot be resolved => status = "pending_entities".

2.  **Normalize relation type**

-   If relation_type is official => keep it.
-   If relation_type is custom:* => keep it and mark for telemetry.
-   If relation_type is unknown (not custom) => map to closest official or convert to custom:*.

3.  **Enforce direction + mirror rules**

-   Use preferred_direction to standardize edge orientation.
-   If mirror is defined => set create_mirror=true.
-   If symmetric=true => ensure mirror equals the same type.

4.  **Validate against per-entity constraints**

-   Enforce pair_candidates for (source_type, relation_type) => must allow target_type.
-   Enforce min confidence / evidence rules.
-   If invalid => status = "invalid".

5.  **Compare with existing relations (dedup / conflict detection)**

-   Use unified relation table as source of truth.
-   Check duplicates by canonical key:

-   (source_id, target_id, relation_type, context_type, context_id)
-   also check mirror orientation if applicable

-   If already exists => dedup.is_duplicate=true and status="ready" (UI may hide or mark as existing).
-   If conflicts exist (e.g. ally_of and enemy_of simultaneously in same context) => keep candidate but mark warning:

-   status="ready" + dedup.reason="conflict_with_existing" (UI decides).

6.  **Generate relation summary (explicit requirement)**

The official relation map should be usable to generate **stable, human-readable summaries** for ingestion and UI.

-   Use relation.types.json.semantics + relation_type + (source_name,target_name) to generate:

-   summary (short, single sentence)
-   optional ui_label

-   These summaries can be persisted (optional) or computed at read-time.

Examples:

-   member_of: “{source} is a member of {target}.”
-   led_by: “{source} is led by {target}.”
-   participated_in: “{source} participated in {target}.”

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Phase5 Prompt (Optional)

Phase5 can be implemented fully deterministically. If you keep an LLM step, it MUST be constrained and used only for:

-   mapping unknown relation_type -> closest official
-   generating natural-language summaries

If LLM is used here, prefer a small model and strict JSON mode.

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Eventos (extract/stream)

Novos eventos

-   phase.start / phase.done

-   relation_discovery
-   relation_normalize

-   relation.candidate

-   Emitido por Phase4

-   relation.normalized

-   Emitido por Phase5

Compatibilidade

-   result_entities permanece inalterado
-   result_relations adicionado após entidades
-   result legado continua enquanto necessário

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Feedback loop (Human-in-the-loop) — (Novo)

**Assim como entidades, relações NÃO são “criadas automaticamente”.** O gateway **só sugere**; o humano confirma/cria via UI.

Princípios

-   O gateway retorna:

-   entidades sugeridas (com matches/candidates)
-   **relações sugeridas** (com evidence/confidence)

-   O front-end oferece:

-   **Aceitar / Rejeitar / Ajustar** (tipo, direção, notas)

-   Persistência no main-service ocorre **somente após ação humana**.

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Instruções para o Front-end (Contrato + Regras) — (Novo)

1) Todo nó de relação tem ref e id opcional

-   ref **sempre** vem.
-   id vem **quando existir** (match confirmado ou entidade já criada).

Exemplo:

{  
"source":  {  "ref":  "finding:character:0",  "id":  **null**  },  
"target":  {  "ref":  "match:faction:2f6c...",  "id":  "2f6c..."  }  
}

2) status vem calculado pelo gateway

Para reduzir lógica no front, cada relação deve vir com:

-   ready → pode persistir agora
-   pending_entities  → falta id em source e/ou target
-   invalid → não deve persistir (ex.: normalização falhou, duplicada óbvia, confidence abaixo do threshold)

3) Relações com entidades ainda não criadas

Quando source.id == null ou target.id == null:

-   A relação é exibida como **pendente** (status=pending_entities).
-   A UI **desabilita** o botão “Criar relação” até resolver ambos os IDs.

Exemplo:

{  
"source":  {  "ref":  "finding:character:0",  "id":  **null**  },  
"target":  {  "ref":  "finding:faction:2",  "id":  **null**  },  
"relation_type":  "member_of",  
"evidence":  "Ari jurou lealdade à Ordem do Sol",  
"confidence":  0.78,  
"status":  "pending_entities"  
}

4) Como “amarrar” o ID temporário

Recomendação padrão:

-   finding:<entity_type>:<index> (estável dentro do request)

Opcional (futuro):

-   temp_uuid gerado no gateway por request se precisar de estabilidade cross-request.

5) ref_map (front mantém; gateway pode ecoar)

O front deve manter um dicionário:

{  
"ref_map":  {  
"finding:character:0":  "uuid-real-character",  
"finding:faction:2":  "uuid-real-faction"  
}  
}

-   Ao **criar** ou **selecionar** um match para uma entidade sugerida, o front atualiza ref_map.
-   Quando ref_map resolver ambos os lados da relação, a UI habilita “Criar relação”.

6) Regra única de resolução (utilitário recomendado)

Implementar uma função:

-   resolveId(node) = node.id ?? ref_map[node.ref] ?? null

7) Habilitar persistência da relação

O botão “Criar relação” fica habilitado se:

-   resolveId(source) != null
-   resolveId(target) != null
-   relation.status != "invalid"

8) Chamada de persistência

Quando habilitado e confirmado pelo humano:

-   substituir source.ref/target.ref por source_id/target_id resolvidos
-   chamar CreateRelation(source_id, target_id, relation_type, context, evidence, confidence)

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Events as First-Class World Entities — (New)

**Events are first-class** ``** and participate in the same unified relation table.**

This allows:

-   causal chains
-   timeline reasoning
-   relationship evolution
-   explanation of _why_ something is true in the world

Events follow the same rules as any other entity:

-   have IDs
-   can be source or target
-   are never auto-created (HITL applies)

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Event Relation Types (Global)

The following relations are added to relation.types.json and are **event-centric**, but usable by any entity.

{  
"participated_in":  {  "mirror":  "has_participant",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source entity participated in an event."  },  
"has_participant":  {  "mirror":  "participated_in",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Event has source entity as a participant."  },  
  
"triggered":  {  "mirror":  "triggered_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source event triggered target event/entity."  },  
"triggered_by":  {  "mirror":  "triggered",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source event was triggered by target."  },  
  
"caused":  {  "mirror":  "caused_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source event/entity caused target."  },  
"caused_by":  {  "mirror":  "caused",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source event/entity was caused by target."  },  
  
"resulted_in":  {  "mirror":  "resulted_from",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source event resulted in target entity/state/event."  },  
"resulted_from":  {  "mirror":  "resulted_in",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source entity/state/event resulted from target."  },  
  
"prevented":  {  "mirror":  "prevented_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source entity/event prevented target event."  },  
"prevented_by":  {  "mirror":  "prevented",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source event was prevented by target."  },  
  
"revealed":  {  "mirror":  "revealed_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source event revealed target information/entity."  },  
"revealed_by":  {  "mirror":  "revealed",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source entity was revealed by target event."  }  
}

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

File: event.relation.map.json

{  
"entity_type":  "event",  
"version":  1,  
"relations":  {  
"has_participant":  {  
"pair_candidates":  ["character",  "faction",  "organization",  "group",  "artifact",  "location"],  
"description":  "Event involves an entity as a participant.",  
"contexts":  ["battles",  "rituals",  "meetings",  "ceremonies",  "discoveries"],  
"signals":  ["took part in",  "participated",  "was present",  "attended"],  
"constraints":  {  "min_confidence":  0.55,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"triggered":  {  
"pair_candidates":  ["event"],  
"description":  "Event triggered another event.",  
"contexts":  ["chain reactions",  "escalations",  "aftermath"],  
"signals":  ["triggered",  "led to",  "sparked"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"resulted_in":  {  
"pair_candidates":  ["event",  "artifact",  "location",  "faction",  "character"],  
"description":  "Event resulted in a change or new entity/state.",  
"contexts":  ["aftermath",  "consequences",  "outcomes"],  
"signals":  ["resulted in",  "ended with",  "gave rise to"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"revealed":  {  
"pair_candidates":  ["artifact",  "character",  "location",  "faction"],  
"description":  "Event revealed hidden information or entities.",  
"contexts":  ["discoveries",  "revelations"],  
"signals":  ["revealed",  "discovered",  "uncovered"],  
"constraints":  {  "min_confidence":  0.55,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
}  
}  
}

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Main Service

Official Relation Map (Types + Mirrors)

-   Single source of truth in the main-service.
-   Optionally exported (read-only) for the gateway and UI.

_File:_ _relation.types.json_

{  
"parent_of":  {  "mirror":  "child_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is a parent of target."  },  
"child_of":  {  "mirror":  "parent_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is a child of target."  },  
"ancestor_of":  {  "mirror":  "descendant_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is an ancestor of target."  },  
"descendant_of":  {  "mirror":  "ancestor_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is a descendant of target."  },  
  
"sibling_of":  {  "mirror":  "sibling_of",  "symmetric":  **true**,  "preferred_direction":  "source_to_target",  "semantics":  "Source and target are siblings."  },  
"spouse_of":  {  "mirror":  "spouse_of",  "symmetric":  **true**,  "preferred_direction":  "source_to_target",  "semantics":  "Source and target are spouses/partners."  },  
"lover_of":  {  "mirror":  "lover_of",  "symmetric":  **true**,  "preferred_direction":  "source_to_target",  "semantics":  "Source and target are lovers."  },  
"ex_lover_of":  {  "mirror":  "ex_lover_of",  "symmetric":  **true**,  "preferred_direction":  "source_to_target",  "semantics":  "Source and target were lovers in the past."  },  
  
"guardian_of":  {  "mirror":  "ward_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is the guardian of target."  },  
"ward_of":  {  "mirror":  "guardian_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is under the guardianship of target."  },  
  
"ally_of":  {  "mirror":  "ally_of",  "symmetric":  **true**,  "preferred_direction":  "source_to_target",  "semantics":  "Source and target are allies."  },  
"enemy_of":  {  "mirror":  "enemy_of",  "symmetric":  **true**,  "preferred_direction":  "source_to_target",  "semantics":  "Source and target are enemies."  },  
"rival_of":  {  "mirror":  "rival_of",  "symmetric":  **true**,  "preferred_direction":  "source_to_target",  "semantics":  "Source and target are rivals."  },  
  
"trusts":  {  "mirror":  "trusted_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source trusts target."  },  
"trusted_by":  {  "mirror":  "trusts",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is trusted by target."  },  
"betrayed":  {  "mirror":  "betrayed_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source betrayed target."  },  
"betrayed_by":  {  "mirror":  "betrayed",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source was betrayed by target."  },  
  
"mentor_of":  {  "mirror":  "mentored_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source mentors or trains target."  },  
"mentored_by":  {  "mirror":  "mentor_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is mentored or trained by target."  },  
  
"member_of":  {  "mirror":  "has_member",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source belongs to a group or organization."  },  
"has_member":  {  "mirror":  "member_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source has target as a member."  },  
  
"leader_of":  {  "mirror":  "led_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source leads target."  },  
"led_by":  {  "mirror":  "leader_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is led by target."  },  
  
"founder_of":  {  "mirror":  "founded_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source founded target."  },  
"founded_by":  {  "mirror":  "founder_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source was founded by target."  },  
  
"at_war_with":  {  "mirror":  "at_war_with",  "symmetric":  **true**,  "preferred_direction":  "source_to_target",  "semantics":  "Source and target are at war."  },  
"in_truce_with":  {  "mirror":  "in_truce_with",  "symmetric":  **true**,  "preferred_direction":  "source_to_target",  "semantics":  "Source and target are in a truce."  },  
  
"vassal_of":  {  "mirror":  "overlord_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is a vassal of target."  },  
"overlord_of":  {  "mirror":  "vassal_of",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is an overlord of target."  },  
  
"located_in":  {  "mirror":  "contains",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is located in target."  },  
"contains":  {  "mirror":  "located_in",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source contains target."  },  
  
"controls":  {  "mirror":  "controlled_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source controls target."  },  
"controlled_by":  {  "mirror":  "controls",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is controlled by target."  },  
  
"owns":  {  "mirror":  "owned_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source owns target."  },  
"owned_by":  {  "mirror":  "owns",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is owned by target."  },  
  
"wields":  {  "mirror":  "wielded_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source wields or uses an artifact."  },  
"wielded_by":  {  "mirror":  "wields",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is wielded by target."  },  
  
"created":  {  "mirror":  "created_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source created target."  },  
"created_by":  {  "mirror":  "created",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source was created by target."  },  
  
"bound_to":  {  "mirror":  "binds",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source is bound to target."  },  
"binds":  {  "mirror":  "bound_to",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source binds target."  },  
  
"caused":  {  "mirror":  "caused_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source caused target."  },  
"caused_by":  {  "mirror":  "caused",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source was caused by target."  },  
  
"participated_in":  {  "mirror":  "has_participant",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source entity participated in an event."  },  
"has_participant":  {  "mirror":  "participated_in",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Event has source entity as a participant."  },  
  
"triggered":  {  "mirror":  "triggered_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source triggered target (event/entity)."  },  
"triggered_by":  {  "mirror":  "triggered",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source was triggered by target."  },  
  
"resulted_in":  {  "mirror":  "resulted_from",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source resulted in target (event/entity/state)."  },  
"resulted_from":  {  "mirror":  "resulted_in",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source resulted from target."  },  
  
"prevented":  {  "mirror":  "prevented_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source prevented target."  },  
"prevented_by":  {  "mirror":  "prevented",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source was prevented by target."  },  
  
"revealed":  {  "mirror":  "revealed_by",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source revealed target."  },  
"revealed_by":  {  "mirror":  "revealed",  "symmetric":  **false**,  "preferred_direction":  "source_to_target",  "semantics":  "Source was revealed by target."  }  
}

Note: preferred_direction is used by Phase5 to standardize how we store/emit edges when the LLM output is ambiguous.

Note: preferred_direction is used by Phase5 to standardize how we store/emit edges when the LLM output is ambiguous.

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

Practical Package (Per-Entity Relation Constraints) — (New)

Goal: keep a **global** type definition (relation.types.json) and add **per-entity** constraints to help:

-   Phase4: guide the LLM toward plausible edges
-   Phase5: validate (source_type, relation_type, target_type) and mark invalid edges early
-   UI: show better labels/examples for confirmation

Design rules

-   relation.types.json is the only place that defines mirror/symmetric/preferred_direction.
-   Each *.relation.map.json defines **what makes sense when** ``** is the SOURCE**.
-   Per-entity maps provide:

-   pair_candidates (allowed target types)
-   description and contexts (prompt-friendly semantics)
-   signals / anti_signals (heuristics for discovery)
-   optional constraints (min_confidence, allow_implicit, requires_evidence)

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

File: character.relation.map.json

{  
"entity_type":  "character",  
"version":  1,  
"relations":  {  
"member_of":  {  
"pair_candidates":  ["faction",  "organization",  "group"],  
"description":  "Character belongs to a group/organization.",  
"contexts":  [  
"when a character joins a faction or guild",  
"when a character is enlisted in a guard/unit",  
"when a character is part of a cult or order"  
],  
"signals":  ["member of",  "joined",  "part of",  "enlisted",  "swore loyalty to"],  
"anti_signals":  ["visited",  "met",  "heard about"],  
"constraints":  {  "min_confidence":  0.55,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"leader_of":  {  
"pair_candidates":  ["faction",  "organization",  "group"],  
"description":  "Character leads a group/organization.",  
"contexts":  [  
"when a character commands a unit",  
"when a character rules a faction",  
"when a character is appointed as captain/leader"  
],  
"signals":  ["leads",  "commands",  "captain of",  "chief of",  "rules"],  
"anti_signals":  ["temporary",  "acting",  "substitute"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"mentor_of":  {  
"pair_candidates":  ["character"],  
"description":  "Character mentors another character.",  
"contexts":  ["when a character trains another",  "when a master/apprentice relationship exists"],  
"signals":  ["trained",  "mentored",  "taught",  "apprentice"],  
"anti_signals":  ["met once",  "brief advice"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"ally_of":  {  
"pair_candidates":  ["character",  "faction",  "organization",  "group"],  
"description":  "Character is allied with another entity.",  
"contexts":  ["when they cooperate toward a goal",  "when alliances are declared"],  
"signals":  ["allied with",  "teamed up",  "formed an alliance",  "joined forces"],  
"anti_signals":  ["neutral",  "temporary truce"],  
"constraints":  {  "min_confidence":  0.55,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"enemy_of":  {  
"pair_candidates":  ["character",  "faction",  "organization",  "group"],  
"description":  "Character is enemies with another entity.",  
"contexts":  ["when they are in conflict",  "when rivalry/feud is explicit"],  
"signals":  ["enemy",  "hunted",  "swore revenge",  "at war with"],  
"anti_signals":  ["dislikes",  "annoyed"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"owns":  {  
"pair_candidates":  ["artifact",  "location"],  
"description":  "Character owns an item or property.",  
"contexts":  ["when an item is possessed",  "when property is legally owned"],  
"signals":  ["owns",  "belongs to",  "property",  "title deed"],  
"anti_signals":  ["borrowed",  "stole",  "found"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **false**,  "requires_evidence":  **true**  }  
},  
"located_in":  {  
"pair_candidates":  ["location"],  
"description":  "Character is physically located in a place.",  
"contexts":  ["when the narrative places a character at a location"],  
"signals":  ["in",  "at",  "inside",  "entered",  "arrived"],  
"anti_signals":  ["dreamed",  "remembered",  "heard about"],  
"constraints":  {  "min_confidence":  0.5,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
}  
}  
}

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

File: faction.relation.map.json

{  
"entity_type":  "faction",  
"version":  1,  
"relations":  {  
"has_member":  {  
"pair_candidates":  ["character"],  
"description":  "Faction has a character as a member.",  
"contexts":  ["when a roster, initiation, or membership is stated"],  
"signals":  ["members include",  "recruited",  "initiation",  "joined"],  
"anti_signals":  ["met",  "visited"],  
"constraints":  {  "min_confidence":  0.55,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"led_by":  {  
"pair_candidates":  ["character"],  
"description":  "Faction is led by a character.",  
"contexts":  ["when a leader is named",  "when command structure is explicit"],  
"signals":  ["led by",  "commanded by",  "under the rule of"],  
"anti_signals":  ["influenced by"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"ally_of":  {  
"pair_candidates":  ["faction",  "organization",  "group"],  
"description":  "Faction is allied with another group.",  
"contexts":  ["when alliances/treaties exist"],  
"signals":  ["alliance",  "treaty",  "joined forces"],  
"anti_signals":  ["trade",  "neutral"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"enemy_of":  {  
"pair_candidates":  ["faction",  "organization",  "group"],  
"description":  "Faction is hostile to another group.",  
"contexts":  ["when factions are at war",  "when hostility is declared"],  
"signals":  ["at war",  "enemy",  "hostile",  "raided"],  
"anti_signals":  ["competition"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"contains":  {  
"pair_candidates":  ["location"],  
"description":  "Faction controls/contains territory (locations).  
",  
"contexts":  ["when a faction controls a city/fort/region"],  
"signals":  ["controls",  "holds",  "occupies",  "territory"],  
"anti_signals":  ["visited"],  
"constraints":  {  "min_confidence":  0.55,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
}  
}  
}

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

File: location.relation.map.json

{  
"entity_type":  "location",  
"version":  1,  
"relations":  {  
"contains":  {  
"pair_candidates":  ["location",  "character",  "artifact",  "faction",  "organization",  "group"],  
"description":  "Location contains another entity.",  
"contexts":  ["when an entity is inside a place",  "when a place includes sub-locations"],  
"signals":  ["inside",  "within",  "contained",  "housed"],  
"anti_signals":  ["near",  "adjacent"],  
"constraints":  {  "min_confidence":  0.5,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"located_in":  {  
"pair_candidates":  ["location"],  
"description":  "Location is located within another location.",  
"contexts":  ["when a place is part of a larger region"],  
"signals":  ["in",  "within",  "part of",  "region of"],  
"anti_signals":  ["near"],  
"constraints":  {  "min_confidence":  0.55,  "allow_implicit":  **true**,  "requires_evidence":  **true**  }  
},  
"owned_by":  {  
"pair_candidates":  ["character",  "faction",  "organization",  "group"],  
"description":  "Location/property is owned by an entity.",  
"contexts":  ["when a property belongs to someone"],  
"signals":  ["owned by",  "property of",  "estate"],  
"anti_signals":  ["occupied"],  
"constraints":  {  "min_confidence":  0.6,  "allow_implicit":  **false**,  "requires_evidence":  **true**  }  
}  
}  
}

![pastedGraphic.png](blob:https://stackedit.io/9d9f6aa8-fb5e-4180-ba8d-2351828805aa)

File: artifact.relation.map.json

{
  "entity_type": "artifact",
  "version": 1,
  "relations": {
    "owned_by": {
      "pair_candidates": [
        "character",
        "faction",
        "organization",
        "group"
      ],
      "description": "Artifact is owned or possessed by an entity.",
      "contexts": [
        "when an artifact belongs to a character",
        "when an artifact is held by a faction or organization",
        "when legal or symbolic ownership is established"
      ],
      "signals": [
        "owned by",
        "belongs to",
        "in the possession of",
        "carried by",
        "kept by"
      ],
      "anti_signals": [
        "borrowed",
        "temporarily held",
        "stolen",
        "found"
      ],
      "constraints": {
        "min_confidence": 0.6,
        "allow_implicit": true,
        "requires_evidence": true
      }
    },

    "wielded_by": {
      "pair_candidates": [
        "character"
      ],
      "description": "Artifact is actively wielded or used by a character.",
      "contexts": [
        "when a character uses a weapon",
        "when a magical artifact is actively controlled",
        "when an artifact is bound to a wielder"
      ],
      "signals": [
        "wielded by",
        "used by",
        "brandished by",
        "channeled through",
        "bound to"
      ],
      "anti_signals": [
        "owned by",
        "stored",
        "sealed",
        "lost"
      ],
      "constraints": {
        "min_confidence": 0.6,
        "allow_implicit": true,
        "requires_evidence": true
      }
    },

    "created_by": {
      "pair_candidates": [
        "character",
        "organization",
        "faction"
      ],
      "description": "Artifact was created or forged by an entity.",
      "contexts": [
        "when an artifact is forged by a smith",
        "when a magical item is created by a mage",
        "when an artifact has a known creator"
      ],
      "signals": [
        "forged by",
        "created by",
        "crafted by",
        "made by"
      ],
      "anti_signals": [
        "found",
        "discovered",
        "inherited"
      ],
      "constraints": {
        "min_confidence": 0.65,
        "allow_implicit": true,
        "requires_evidence": true
      }
    },

    "bound_to": {
      "pair_candidates": [
        "character",
        "location",
        "event"
      ],
      "description": "Artifact is magically or symbolically bound to an entity, place, or event.",
      "contexts": [
        "when an artifact is soulbound",
        "when an artifact is bound to a location",
        "when an artifact is tied to a ritual or prophecy"
      ],
      "signals": [
        "bound to",
        "soulbound",
        "linked to",
        "cannot be separated from"
      ],
      "anti_signals": [
        "stored",
        "kept",
        "owned"
      ],
      "constraints": {
        "min_confidence": 0.65,
        "allow_implicit": true,
        "requires_evidence": true
      }
    },

    "located_in": {
      "pair_candidates": [
        "location"
      ],
      "description": "Artifact is located or stored in a place.",
      "contexts": [
        "when an artifact is hidden somewhere",
        "when an artifact is stored or sealed in a location",
        "when the narrative places an artifact in a place"
      ],
      "signals": [
        "located in",
        "hidden in",
        "stored in",
        "sealed in",
        "kept at"
      ],
      "anti_signals": [
        "mentioned",
        "rumored",
        "dreamed of"
      ],
      "constraints": {
        "min_confidence": 0.55,
        "allow_implicit": true,
        "requires_evidence": true
      }
    },

    "revealed_by": {
      "pair_candidates": [
        "event"
      ],
      "description": "Artifact was revealed or discovered by an event.",
      "contexts": [
        "when an artifact is discovered during an event",
        "when a ritual or catastrophe reveals an artifact",
        "when an exploration uncovers an artifact"
      ],
      "signals": [
        "revealed by",
        "discovered during",
        "uncovered in",
        "found after"
      ],
      "anti_signals": [
        "always known",
        "public knowledge"
      ],
      "constraints": {
        "min_confidence": 0.55,
        "allow_implicit": true,
        "requires_evidence": true
      }
    }
  }
}



Você tem razão — no trecho que você colou **não aparece a seção final de “Next Steps / Próximas etapas”** (e também vale a pena fechar com “Deliverables” + “Risks” pra não ficar solto).

Abaixo está um bloco **pronto pra você colar no final do documento salvo**, **em inglês** e alinhado com tudo que já está aí (Phase4/5, HITL, eventos, mapas, dedup/summary etc.).

---

## Next Steps (Roadmap)

### MVP (Iteration 1)

1. **Finalize relation maps**

   * Ensure `relation.types.json` is the single source of truth for:

     * `mirror`, `symmetric`, `preferred_direction`, `semantics`
   * Complete per-entity maps:

     * `character.relation.map.json`
     * `faction.relation.map.json`
     * `location.relation.map.json`
     * `artifact.relation.map.json`
     * `event.relation.map.json`

2. **Implement Phase4 (relation_discovery)**

   * Support **two input modes**:

     * `full_text`
     * `spans + global_summary`
   * Emit SSE:

     * `phase.start (relation_discovery)`
     * `relation.candidate` (per relation)
     * `phase.done (relation_discovery)`
   * Strict gateway-side validation:

     * JSON parse
     * `pair_candidates` enforcement
     * `evidence.span_id` existence (Mode B)
     * confidence clamp `[0..1]`
   * Do **not** persist (HITL only).

3. **Implement Phase5 (relation_normalize)**

   * Resolve IDs:

     * `finding:*` / `match:*` => `id` when possible
     * set `status=pending_entities` when unresolved
   * Normalize type:

     * official -> keep
     * unknown -> map to closest OR `custom:*`
   * Enforce:

     * direction + mirror/symmetry rules
     * per-entity `pair_candidates` and constraints
   * Produce:

     * `status` (`ready|pending_entities|invalid`)
     * `dedup` metadata
     * **`summary`** using `relation.types.json.semantics`

4. **Existing relations snapshot for dedup/conflicts**

   * Add a lightweight main-service endpoint or query path to fetch:

     * existing relations by `context_type/context_id`
     * optionally filtered by involved IDs
   * Phase5 uses this snapshot to:

     * mark duplicates
     * flag conflicts (warnings)

5. **Wire the UI contract**

   * Ensure response includes:

     * node `{ ref, id? }`
     * `status`
     * `dedup`
     * `summary`
   * Frontend implements:

     * `ref_map` resolution
     * disable “Create relation” until both IDs resolve

6. **Persistence flow (human confirmed)**

   * Add/confirm main-service API call:

     * `CreateRelation(source_id, target_id, relation_type, context, evidence, confidence)`
   * If `mirror` exists and `create_mirror=true`:

     * persist both edges (or main-service auto-creates mirror)

7. **Testing**

   * Unit tests:

     * Phase4 input validation (mode rules, spans, mentions)
     * Phase4 output validation (pair_candidates, evidence)
     * Phase5 normalization/direction/mirror
     * Phase5 dedup + conflict detection
   * Golden fixtures with sample stories/scenes.

---

## Deliverables (What “Done” Means)

* **Gateway**

  * Phase4 implemented with both modes + canonical prompt
  * Phase5 implemented with strict normalization + summaries + dedup metadata
  * SSE events emitting candidates + normalized relations
  * API returns `{ entities, relations }`

* **Main-service**

  * `relation.types.json` stored and used as official map
  * Endpoint/query to fetch existing relations snapshot for a context (optional but recommended)
  * CreateRelation supports mirror creation or returns `create_mirror` guidance

* **Documentation**

  * Updated pipeline docs (Phase2..Phase5)
  * UI contract for `ref/id/status/ref_map`

---

## Risks & Mitigations (Final)

* **Hallucinated relations**

  * Require `evidence` always
  * Use `min_confidence` constraints
  * Prefer `asserted` polarity; allow `uncertain` but low confidence

* **Over-generation / noise**

  * Enforce per-entity `pair_candidates`
  * Dedup in Phase4 (by refs) and Phase5 (by IDs + context)

* **Ambiguous direction**

  * Phase5 uses `preferred_direction` + mirror rules

* **Schema drift**

  * `relation.types.json` remains the only source of truth for relation invariants
  * per-entity maps only constrain candidate pairs and heuristics

---
# Phase 5 — `relation_normalize` (Full Spec)

> **Purpose:** Convert Phase4 relation candidates into **UI-friendly, persistable-ready** normalized relations, by applying:
>
> * canonical ID resolution (finding/match refs → UUIDs when possible)
> * official relation type normalization (with mirror/symmetry)
> * per-entity constraint validation (pair candidates, confidence rules)
> * deduplication and conflict/warning detection against the **unified relation table**
> * stable summary generation for UI/ingestion
>
> **Important:** Phase5 does **not** auto-create entities or relations. It only normalizes + annotates. Persisting remains **Human-in-the-loop**.

---

## Inputs

Phase5 consumes:

1. **Candidate relations** (Phase4 output)
2. **Official relation types map**: `relation.types.json` (mirror/symmetric/preferred_direction + semantics)
3. **Per-entity constraint maps**: `*.relation.map.json` (strict `(source_type, relation_type, target_type)` constraints)
4. **Entity resolution sources**:

   * confirmed matches (`finding_ref -> canonical id`)
   * optional UI `ref_map` (if the UI already resolved some findings)
5. **Existing relations snapshot** from main-service (recommended)

---

## Why ingestion must be reviewed (explicit dependency)

Because Phase5 must compare candidates with what already exists in the **unified relation table**, the ingestion path (create/update/import) must guarantee:

* **Canonical key** is always derivable:

  * `source_entity_id`, `target_entity_id`, `relation_type`, `context_type`, `context_id`
* **Mirror behavior** is consistent:

  * either main-service auto-creates mirrors, or the client does — but the rule must be stable
* **Evidence & confidence** can be stored and read back (recommended)
* **Read-back parity:** existing relation snapshot must contain enough fields to dedup reliably

---

## Existing Relations Snapshot (Phase5 dependency)

### Recommended read-only endpoint shape

Fetch by context:

* `context_type` + `context_id`
* optional filters: `entity_ids[]` or `(source_id,target_id)` pairs for payload reduction

### Snapshot schema (minimal)

```json
{ "context": {"type":"scene","id":"scene-uuid"}, "relations": [ { "source_id":"uuid-src", "target_id":"uuid-tgt", "relation_type":"member_of", "created_at":"2026-01-10T12:00:00Z", "summary":"Ari is a member of Order of the Sun." } ] }
```

Optional fields:

* `evidence`, `confidence`, `implicit`, `polarity`
* `mirror_of` (if you choose to track linkage)
* `revision_id` (if story versioning impacts validity)

---

## Phase5 data contract

### Phase5Input (recommended)

```json
{ "request_id":"req-123", "context": {"type":"scene","id":"scene-uuid"}, "candidates":[ { "relation_ref":"rel:0", "source": {"ref":"finding:character:0","type":"character"}, "target": {"ref":"match:faction:uuid-abc","type":"faction"}, "relation_type":"pledged_loyalty", "polarity":"asserted", "implicit":false, "confidence":0.78, "evidence": {"span_id":"span:2","quote":"Ari swore loyalty to the Order of the Sun."} } ], "confirmed_matches":[ { "finding_ref":"finding:faction:1", "match": {"ref":"match:faction:uuid-abc","type":"faction","id":"uuid-abc","canonical_name":"Order of the Sun","similarity":0.91} } ], "ref_map": {"finding:character:0":"uuid-char"}, "relation_types": {"member_of": {"mirror":"has_member","symmetric":false,"preferred_direction":"source_to_target","semantics":"Source belongs to a group or organization."} }, "per_entity_maps": {"character": {"entity_type":"character","version":1,"relations": {"member_of": {"pair_candidates":["faction"],"description":"Character belongs to a group/organization.","constraints": {"min_confidence":0.55,"allow_implicit":true,"requires_evidence":true} } } } }, "existing_snapshot": {"context": {"type":"scene","id":"scene-uuid"}, "relations": [ {"source_id":"uuid-char","target_id":"uuid-abc","relation_type":"member_of"} ] } }
```

Notes:

* `relation_ref` must be stable and echoed by any Phase5 helper LLMs.
* JSON examples are intentionally single-line friendly.

---

## Phase5 output

### Phase5NormalizedRelation (UI-first)

```json
{ "relation_ref":"rel:0", "source": {"ref":"finding:character:0","id":"uuid-src","type":"character"}, "target": {"ref":"match:faction:uuid-abc","id":"uuid-tgt","type":"faction"}, "relation_type":"member_of", "direction":"source_to_target", "create_mirror":true, "confidence":0.78, "polarity":"asserted", "implicit":false, "evidence": {"span_id":"span:2","quote":"Ari swore loyalty to the Order of the Sun."}, "status":"ready", "dedup": {"is_duplicate":false,"reason":""}, "warnings": [], "summary":"Ari is a member of Order of the Sun." }
```

`status` values:

* `ready` — can be persisted (IDs resolved; valid)
* `pending_entities` — missing source_id and/or target_id
* `invalid` — violates constraints, insufficient evidence, too low confidence, unknown mapping failure

---

## Phase5 responsibilities (explicit)

### 1) Resolve canonical IDs

**Resolution rule (canonical):**

* `resolveId(node) = node.id ?? ref_map[node.ref] ?? confirmed_matches[finding_ref].match.id ?? null`

If either side is null:

* `status = "pending_entities"`
* keep normalized relation for UI, but disable persistence

### 2) Normalize relation types

Rules:

* If candidate type exists in `relation.types.json` ⇒ keep
* If candidate type starts with `custom:` ⇒ keep, mark telemetry
* Else ⇒ map to closest official or convert to `custom:<snake_case>`

### 3) Enforce direction + mirror/symmetry

* Use `preferred_direction` to standardize orientation
* If `mirror` exists ⇒ `create_mirror=true`
* If `symmetric=true` ⇒ mirror must equal itself, and (A,B) == (B,A)

### 4) Validate using per-entity constraints

For a relation to be valid:

* `(source_type, relation_type)` must exist in the injected per-entity map OR be `custom:*` (configurable)
* `target_type` must be in `pair_candidates`
* confidence >= `min_confidence`
* evidence required when `requires_evidence=true`

If invalid ⇒ `status="invalid"`

### 5) Compare against existing relations (dedup + warnings)

Dedup checks (canonical):

1. **Direct duplicate**

* `(source_id, target_id, relation_type, context_type, context_id)`

2. **Mirror duplicate**

* if mirror exists, check `(target_id, source_id, mirror_type, context_type, context_id)`

3. **Symmetric duplicate**

* if symmetric, treat swapped endpoints as duplicate

If already exists:

* set `dedup.is_duplicate=true`
* keep `status="ready"` (UI may hide or show “already exists”)

Conflict detection (soft warnings):

* domain-specific (e.g. ally_of vs enemy_of)
* Phase5 does not resolve; it flags warnings

### 6) Generate stable summaries (explicit requirement)

Phase5 MUST produce a short single-sentence `summary`.

Priority:

1. deterministic template using `relation.types.json` semantics and names
2. optional helper LLM for naturalness (strict JSON mode)

---

## Phase5 helper prompts (Optional LLM)

> Phase5 can be deterministic. If an LLM is used, it must be constrained and used only for:
>
> * mapping unknown relation types
> * generating summaries

### Prompt — RELATION_TYPE_NORMALIZER (single-line JSON)

```text
You are Phase5 helper: RELATION_TYPE_NORMALIZER.

Goal:
Map each input relation_type into one of:
- an official relation type from the provided list, OR
- a custom type in the form "custom:<snake_case>"

Hard rules:
- Output MUST be valid JSON only.
- The JSON output MUST be a single line (no newlines, no indentation, no extra spaces).
- You must NOT invent entities, IDs, or relations.
- Use ONLY the provided official relation types for mapping.
- If none fits, return custom:<normalized> (snake_case).
- Do NOT change source/target or direction.
- You MUST preserve and echo back relation_ref exactly as provided.

Inputs:
- official_relation_types[] (strings)
- candidates[] with { relation_ref, relation_type, evidence_quote }

Output schema (single-line JSON):
{ "normalized":[ { "relation_ref":"...", "input_relation_type":"...", "output_relation_type":"...", "reason":"short reason" } ] }
```

### Prompt — RELATION_SUMMARY_GENERATOR (single-line JSON)

```text
You are Phase5 helper: RELATION_SUMMARY_GENERATOR.

Goal:
Generate a short, stable, single-sentence summary for each relation.

Hard rules:
- Output MUST be valid JSON only.
- The JSON output MUST be a single line (no newlines, no indentation, no extra spaces).
- Use the provided semantics as the primary meaning.
- Keep summaries short and factual; do NOT add extra facts.
- You MUST preserve and echo back relation_ref exactly as provided.

Inputs:
- relations[] with { relation_ref, source_name, target_name, relation_type, semantics }

Output schema (single-line JSON):
{ "summaries":[ { "relation_ref":"...", "summary":"..." } ] }
```

---

## Phase5 flow (gateway-side)

1. Build Phase5Input

* candidates from Phase4
* relation.types.json
* per-entity maps
* ref_map + confirmed_matches
* existing relations snapshot (recommended)

2. Resolve IDs

* mark pending_entities where unresolved

3. Normalize relation_type

* deterministic first, LLM helper only for unknown types

4. Apply direction + mirror

5. Validate constraints

* mark invalid

6. Dedup + warnings

* compare with existing snapshot

7. Generate summary

* deterministic or LLM helper

8. Emit SSE

* phase.start (relation_normalize)
* relation.normalized (per relation)
* phase.done (relation_normalize)

9. Return

* return normalized relations in API + SSE `result_relations`
* do NOT persist automatically (HITL)

---

# Addendum — Relation Ingestion Review Plan (required for Phase6 matching)

## Goal

Revise relation ingestion in **main-service** so that the data produced by UI/import/automation can be queried by Phase6 using the same mental model as Phase4/5/6 (refs, evidence-like snippets, summaries, canonical dedup keys).

## Key idea

From each ingested relation edge in the unified relation table, generate a **Relation Ingestion Chunk** (RIC) that mirrors Phase4/Phase5 artifacts.

* Phase5/6/7 reason about:

  * `(source, target, relation_type, context)`
  * evidence (span/quote) + confidence
  * human-readable summary
* Ingestion should produce a compact index record with those same attributes so Phase6 can:

  * deduplicate reliably
  * detect conflicts
  * optionally do semantic matching/search

## Proposed split: Phase5 + Phase6

### Phase5 — relation_normalize (pure transform)

* Input: Phase4 candidates
* Output: normalized candidates (type normalized + direction/mirror applied + validated)
* No snapshot queries
* No dedup
* No summary generation (optional)

### Phase6 — relation_matcher (reads + compares)

* Input: Phase5 normalized relations + ref_map + confirmed matches
* Fetch: existing relations snapshot OR relation ingestion chunks (RIC)
* Output: UI-ready relations with dedup/conflict + summaries

This separation makes Phase5 deterministic and keeps all “compare with stored” concerns in Phase6.

---

## Detailed plan: revise relation ingestion

### Step 0 — Define canonical key + required read fields

**CanonicalKey** (must be derivable for every stored relation):

* `source_entity_id`
* `target_entity_id`
* `relation_type`
* `context_type`
* `context_id`

**Read-back minimum fields for Phase6 snapshot:**

* `relation_id`
* canonical key above
* `created_at` (optional)
* `evidence` (optional but recommended)
* `confidence` (optional but recommended)

### Step 1 — Decide mirror responsibility (single source of truth)

Pick ONE rule (recommended: main-service):

* **Option A (recommended): main-service creates mirrors**

  * Ingest request submits the “primary” edge.
  * main-service checks `relation.types.json` and creates mirror automatically when applicable.

* Option B: client creates mirrors

  * gateway/UI sends both edges (error-prone)

Phase6 must assume a stable rule, otherwise dedup will drift.

### Step 2 — Add a Relation Ingestion Chunk (RIC)

Create a derived record for each ingested relation that can be used for matching/search.

**Storage options:**

* A) Separate table `relation_chunks` (recommended if you want RAG/search)
* B) Denormalized columns on the relation row (`summary`, `evidence_quote`)
* C) Both (A for search, B for fast reads)

**RIC fields (recommended):**

* `relation_id` (FK)
* `context_type`, `context_id`
* `source_entity_id`, `target_entity_id`, `relation_type`
* `summary` (stable sentence)
* `evidence_quote` (short)
* `evidence_ref` (span_id if available, else `null`)
* `confidence` (0..1)
* `chunk_text` (string used for embedding; can be summary + evidence)
* `embedding` (optional)
* `updated_at`

**RIC chunk_text template (example):**

* `"{source_name} {relation_type} {target_name}. Context={context_type}:{context_id}. Evidence={evidence_quote}"`

### Step 3 — Define summary templates in official relation map

Extend `relation.types.json` with an optional stable summary template field (or keep it in code):

* `summary_template`: e.g. `"{source} is a member of {target}."`

Phase6 (and ingestion) can use the same template, guaranteeing consistent summaries.

### Step 4 — Add ingestion hook

Whenever a relation is created/updated:

1. Validate relation against official types (warn if unknown, allow custom)
2. Apply mirror creation rule (if main-service owned)
3. Generate/Update RIC for primary edge and mirror edge
4. (Optional) enqueue embedding job for `chunk_text`

### Step 5 — Add a snapshot endpoint tailored for Phase6

Provide a fast read API for matching:

* `GET /relations/snapshot?context_type=&context_id=&entity_ids=...`

Return minimal fields required for dedup/conflict:

* canonical key + relation_id + relation_type + summary (optional)

This avoids Phase6 needing heavy queries.

### Step 6 — Phase6 matching strategy

Given a Phase5 normalized candidate:

1. Resolve IDs using `confirmed_matches + ref_map`.
2. If unresolved => `pending_entities`.
3. Compute CanonicalKey.
4. Compare against snapshot (exact match):

   * if exists => `dedup.is_duplicate=true`.
5. Check conflicts (optional heuristic set):

   * e.g. `ally_of` vs `enemy_of` for same pair+context
   * `member_of` vs `leader_of` may coexist (not conflict)
6. Generate summary using template + names.

---

## Phase6 input/output contracts

### Phase6Input (Go/JSON)

* `request_id`
* `context`
* `normalized_relations[]` (output of Phase5)
* `ref_map` (finding->uuid)
* `confirmed_matches[]` (optional)
* `existing_snapshot` (recommended)

### Phase6Output

* `relations[]` with:

  * `source {ref,id}` / `target {ref,id}`
  * `relation_type`
  * `create_mirror`
  * `dedup {is_duplicate, reason}`
  * `warnings[]`
  * `summary`
  * `status ready|pending_entities|invalid`

---

## Phase5 helper prompt (Type normalizer) — revised

> Note: Phase5 helper is optional; use it only to map unknown types → official/custom.

Single-line JSON requirement is enforced.

**Prompt — RELATION_TYPE_NORMALIZER (single-line JSON)**

You are Phase5 helper: RELATION_TYPE_NORMALIZER. Goal: Map each input relation_type into one of: - an official relation type from the provided list, OR - a custom type in the form "custom:<snake_case>" Hard rules: - Output MUST be valid JSON only (single line, no newlines). - You must NOT invent entities, IDs, or relations. - Use ONLY the provided official relation types for mapping. - If none fits, return custom:<normalized> (snake_case). - Do NOT change source/target/direction. Inputs: - official_relation_types[] (strings) - candidates[] with { source_ref, target_ref, relation_type, evidence_quote } Output: { "normalized": [ { "source_ref": "...", "target_ref": "...", "input_relation_type": "...", "output_relation_type": "...", "reason": "short reason" } ] }

This includes entity refs so the caller can apply mappings per candidate deterministically.

---

## Phase6 helper prompts (optional)

### Prompt — RELATION_SUMMARY_GENERATOR (single-line JSON)

You are Phase6 helper: RELATION_SUMMARY_GENERATOR. Goal: Generate a short, stable summary sentence for each relation using the provided relation semantics and entity names. Hard rules: Output MUST be valid JSON only (single line, no newlines). Do NOT invent facts. Use only the provided names and relation_type. Inputs: relations[] with { source_name, target_name, relation_type, semantics } Output: { "summaries": [ { "source_name":"...","target_name":"...","relation_type":"...","summary":"..." } ] }

### Prompt — RELATION_CONFLICT_CLASSIFIER (single-line JSON)

You are Phase6 helper: RELATION_CONFLICT_CLASSIFIER. Goal: Given a candidate relation and a list of existing relations for the same pair/context, mark whether there is a potential conflict. Hard rules: Output MUST be valid JSON only (single line, no newlines). Do NOT invent facts. Inputs: candidate { source_id, target_id, relation_type } existing[] { relation_type } Output: { "conflicts": { "has_conflict": true|false, "reason": "..." } }
