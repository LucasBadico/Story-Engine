# Plano: Relacoes e Inferencias no Extract

## Contexto
O pipeline atual do llm-gateway-service extrai entidades (phase2) e faz match com entidades existentes (phase3). O objetivo e adicionar fases para descobrir relacoes entre entidades encontradas e tambem inferir relacoes com entidades existentes nao citadas explicitamente, mantendo compatibilidade com o main-service.

## Principio arquitetural
Toda relacao e uma aresta entre duas entidades, independentemente de origem. O gateway nunca define schema; ele apenas propoe relacoes normalizadas que convergem para o modelo unico do main-service.

## Objetivos
- Descobrir relacoes entre entidades detectadas e entidades existentes inferidas.
- Normalizar relation_type para um mapa oficial com mirrors (quando possivel).
- Persistir relacoes no main-service usando create_mirror quando aplicavel.
- Emitir progresso em extract/stream para as novas fases.
- Retornar relacoes como uma nova lista no retorno da API, alem das entidades.
- Manter evidence/confidence/polarity apenas no payload transiente (gateway/UI), nao persistir no main-service.

## Escopo do trabalho
### LLM Gateway
0) **Phase4: result_entities** (atual PhaseTemp)
   - Continua formatando e retornando entidades (sem mudanca de comportamento).
   - Emite `result_entities` no stream e popula o retorno nao-stream.

1) **Phase5: relation_discovery** (antiga Phase4)
   - Entrada: texto, findings do phase2, resultados do phase3 (matches confirmados), contexto.
   - Saida: lista de relacoes candidatas com evidencias e confidence.
   - Permite relacoes entre entidades detectadas e entidades existentes inferidas.

2) **Phase6: relation_normalize** (antiga Phase5)
   - Normaliza relation_type para o mapa oficial (com mirrors).
   - Valida direcao e evita duplicidade obvia.
   - Quando type nao existe no mapa, tenta aproximar para o mais proximo e marca como "custom" se necessario.
   - Para custom:*, gerar duas relacoes (direta e espelho) porque o main-service nao infere mirror.

3) **Phase7: relation_match** (novo)
   - Busca por relacoes ja existentes no main-service para dedup/conflitos.
   - Pode ser opcional inicialmente (decidir endpoint e payload).

4) **Phase8: result_relations** (novo)
   - Retorna as relacoes finalizadas (normalizadas + dedup status) no stream e no endpoint nao-stream.

5) **Eventos (extract/stream)**
   - phase.start / phase.done com phase: relation_discovery, relation_normalize, relation_match.
   - relation.candidate para cada relacao sugerida.
   - relation.normalized para relacao normalizada/pronta para persistencia.
   - Se o SSE ja entrega entidades, manter o fluxo e iniciar relacoes depois das entidades.
   - Adicionar result_entities e result_relations (manter result legado enquanto necessario).

### Main service
1) **Mapa oficial de relation_type + mirror**
   - Centralizar o mapa (ex: package relation ou configuracao) para uso interno e, opcionalmente, exportacao.
   - Manter compatibilidade com os tipos existentes.
   - Permitir extensao para novos tipos.
   - relation.types.json e *.relation.map.json versionados no repo; carregados em memoria na inicializacao.
   - Considerar endpoint estatico (ex: /static/relations.types, /static/character.relations) para o gateway/UI.

2) **Validacao/normalizacao opcional**
   - Se relation_type nao existir no mapa, aceitar mas registrar advertencia.
   - Expor lista de relacoes suportadas (endpoint opcional) para o gateway consumir.

3) **Persistencia**
    - Usar create_mirror quando type tiver inverso definido.

## Input modes (Phase5)
Phase5 suporta dois modos de entrada, escolhidos pelo gateway:
- **Modo A (full_text):** texto completo, quando curto o suficiente.
- **Modo B (spans + global_summary):** para textos longos. Cada relacao deve referenciar um span_id como evidencia.
Selecao sugerida:
- textos curtos (cenas/trechos pequenos): Modo A
- textos longos (capitulos/cenas extensas): Modo B com spans proximos a mentions

### Regras do modo B
- Cada relacao precisa ter evidence com span_id e quote curta.
- Spans devem ser estaveis e reutilizaveis pela UI.

## Payload Phase5 (relation_discovery)
### Estrutura (resumo)
- `request_id`: correlacao SSE/logs.
- `context`: tipo/id e refs opcionais (pov_ref, location_ref).
- `text`: modo full_text ou spans.
- `entity_findings`: refs estaveis (finding:<type>:<index>), nome, resumo, mentions[].
- `confirmed_matches`: mapeia finding_ref -> match canonical.
- `suggested_relations_by_source_type`: mapas por tipo (pair_candidates, descricao, sinais).
- `relation_type_semantics` (opcional): definicoes globais para grounding.

### Regras de validacao (gateway)
- full_text => text.text obrigatorio.
- spans => spans nao vazio, global_summary 3-8 bullets.
- mentions devem existir nos spans.
- relation_type deve preferir os mapas sugeridos (ou custom:* quando necessario).

### Phase5Input (Go + JSON)
Estrutura detalhada (derivada do v2):
```
type Phase5Input struct {
  RequestID string `json:"request_id"`
  Context Phase5Context `json:"context"`
  Text Phase5TextSpec `json:"text"`
  EntityFindings []EntityFinding `json:"entity_findings"`
  ConfirmedMatches []ConfirmedMatch `json:"confirmed_matches,omitempty"`
  SuggestedRelationsBySourceType map[string]PerEntityRelationMap `json:"suggested_relations_by_source_type"`
  RelationTypeSemantics map[string]string `json:"relation_type_semantics,omitempty"`
}

type Phase5Context struct {
  Type string `json:"type"` // story|world|scene|beat
  ID string `json:"id"`
  POVRef *string `json:"pov_ref,omitempty"`
  LocationRef *string `json:"location_ref,omitempty"`
}

type Phase5TextSpec struct {
  Mode string `json:"mode"` // full_text|spans
  Text *string `json:"text,omitempty"`
  GlobalSummary []string `json:"global_summary,omitempty"`
  Spans []Span `json:"spans,omitempty"`
}

type Span struct {
  SpanID string `json:"span_id"` // span:<n>
  Start int `json:"start"`
  End int `json:"end"`
  Text string `json:"text"`
}

type EntityFinding struct {
  Ref string `json:"ref"` // finding:<type>:<index>
  Type string `json:"type"`
  Name string `json:"name"`
  Summary string `json:"summary"`
  Mentions []string `json:"mentions,omitempty"` // span_ids
}

type ConfirmedMatch struct {
  FindingRef string `json:"finding_ref"`
  Match Match `json:"match"`
}

type Match struct {
  Ref string `json:"ref"` // match:<type>:<uuid>
  Type string `json:"type"`
  ID string `json:"id"`
  CanonicalName string `json:"canonical_name"`
  Similarity float64 `json:"similarity"`
}

type PerEntityRelationMap struct {
  EntityType string `json:"entity_type"`
  Version int `json:"version"`
  Relations map[string]RelationConstraintSpec `json:"relations"`
}

type RelationConstraintSpec struct {
  PairCandidates []string `json:"pair_candidates"`
  Description string `json:"description"`
  Contexts []string `json:"contexts,omitempty"`
  Signals []string `json:"signals,omitempty"`
  AntiSignals []string `json:"anti_signals,omitempty"`
  Constraints *RelationConstraints `json:"constraints,omitempty"`
}

type RelationConstraints struct {
  MinConfidence float64 `json:"min_confidence"`
  AllowImplicit bool `json:"allow_implicit"`
  RequiresEvidence bool `json:"requires_evidence"`
}
```

### JSON exemplo (modo A)
```
{
  "request_id": "req-123",
  "context": { "type": "scene", "id": "scene-uuid" },
  "text": { "mode": "full_text", "text": "Ari entered the Obsidian Tower..." },
  "entity_findings": [
    { "ref": "finding:character:0", "type": "character", "name": "Ari", "summary": "Young mage apprentice" }
  ],
  "confirmed_matches": [],
"suggested_relations_by_source_type": {
    "character": {
      "entity_type": "character",
      "version": 1,
      "relations": {
        "member_of": {
          "pair_candidates": ["faction"],
          "description": "Character belongs to a group/organization."
        }
      }
    }
  }
}
```

### JSON exemplo (modo B)
```
{
  "request_id": "req-456",
  "context": { "type": "chapter", "id": "chapter-uuid", "pov_ref": "finding:character:0" },
  "text": {
    "mode": "spans",
    "global_summary": [
      "Ari enters the Obsidian Tower.",
      "She swears loyalty to the Order of the Sun.",
      "A hidden artifact is revealed in the lower vaults."
    ],
    "spans": [
      { "span_id": "span:1", "start": 0, "end": 120, "text": "Ari entered the Obsidian Tower..." },
      { "span_id": "span:2", "start": 121, "end": 240, "text": "Ari swore loyalty to the Order of the Sun..." }
    ]
  },
  "entity_findings": [
    { "ref": "finding:character:0", "type": "character", "name": "Ari", "summary": "Young mage apprentice", "mentions": ["span:1", "span:2"] },
    { "ref": "finding:faction:1", "type": "faction", "name": "Order of the Sun", "summary": "A militant religious order", "mentions": ["span:2"] }
  ],
  "confirmed_matches": [
    { "finding_ref": "finding:faction:1", "match": { "ref": "match:faction:uuid-abc", "type": "faction", "id": "uuid-abc", "canonical_name": "Order of the Sun", "similarity": 0.91 } }
  ],
"suggested_relations_by_source_type": {
    "character": {
      "entity_type": "character",
      "version": 1,
      "relations": {
        "member_of": {
          "pair_candidates": ["faction"],
          "description": "Character belongs to a group/organization."
        }
      }
    }
  }
}
```

## Entregaveis
- Implementacao das fases 5 a 8 no llm-gateway-service (relation_discovery, normalize, match, result_relations).
- Mapa oficial de relation types no main-service (com mirrors).
- Ajustes de contrato/payload para persistencia e retorno de relacoes na API.
- Documentacao da pipeline e eventos.

## Retorno da API (entities + relations)
### Endpoint tradicional
```
{
  "entities": [
    {
      "type": "character",
      "name": "Ari",
      "summary": "Aprendiz de magia",
      "found": true,
      "match": {
        "source_type": "character",
        "source_id": "uuid",
        "entity_name": "Ari Valen",
        "similarity": 0.91,
        "reason": "Nome e descricao batem"
      },
      "candidates": []
    }
  ],
  "relations": [
    {
      "source": { "type": "character", "ref": "finding:0" },
      "target": { "type": "faction", "ref": "match:character:uuid" },
      "relation_type": "member_of",
      "evidence": "Ari jurou lealdade a Ordem do Sol",
      "confidence": 0.78,
      "context": { "type": "story", "id": "uuid" }
    }
  ]
}
```

### SSE (extract/stream)
- `result_entities`: `{ "entities": [...] }`
- `result_relations`: `{ "relations": [...] }`
- manter `result` legado com `payload` completo enquanto necessario.

## Prompts por fase
### Phase5: relation_discovery
- **Objetivo:** extrair relacoes entre entidades encontradas e entidades existentes inferidas.
- **Inputs no prompt:** texto, contexto, lista de entidades encontradas, lista de matches confirmados.
- **Saida esperada:** JSON com relacoes candidatas contendo source/target, relation_type, evidence, confidence.
- **Regras chave:** usar relation_type sugerido; evidencias obrigatorias; preferir explicitas; implicitas apenas quando forte; usar custom:* quando necessario.

#### Prompt (canonical)
You are Phase5: RELATION_DISCOVERY.

Goal:
Given the text (or spans), discovered entities and confirmed matches, propose candidate relationships between entities.
Do NOT create entities. Do NOT assume facts not supported by the text.

Hard rules:
- Output MUST be valid JSON and match the provided schema.
- Each relation MUST include evidence with a span_id and a short quote.
- Prefer explicit relations; implicit relations are allowed only when strongly implied.
- Only propose relation_type values suggested by the provided per-entity relation maps.
- Enforce pair_candidates: if a relation type does not allow the target entity type, do not output it.
- If support is weak or ambiguous, either omit the relation or set polarity="uncertain" with low confidence.

Notes:
- Use suggested_relations_by_source_type as the source of truth for preferred relation_type values.
- If strongly supported but not in suggestions, use custom:<type>.

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

#### Regras de execucao (gateway)
1) Selecionar modo de input (full_text ou spans).
2) Construir spans (se Modo B) com span_id estavel.
3) Montar payload com entity_findings, confirmed_matches, suggested_relations_by_source_type.
4) Executar LLM em JSON-mode (se suportado) com limites conservadores.
5) Validar JSON, relation_type, pair_candidates e evidence.
6) Deduplicar por (source_ref, target_ref, relation_type, context_id).
7) Emitir events: phase.start, relation.candidate, phase.done.

### Phase6: relation_normalize
- **Objetivo:** normalizar relation_type usando o mapa oficial e validar direcao/mirror.
- **Inputs no prompt:** lista de relacoes candidatas, mapa oficial (com mirrors), regras de direcionalidade.
- **Saida esperada:** JSON com relacoes normalizadas (relation_type final, direction confirmada, create_mirror boolean).
- **Obs:** pode ser deterministico. LLM opcional apenas para mapear relation_type desconhecido ou gerar summary.

#### Responsabilidades (deterministico)
1) Resolver IDs (finding/match -> canonical id quando possivel).
2) Normalizar relation_type (oficial, custom:*, ou mapear para o mais proximo).
3) Aplicar direction e mirror (preferred_direction + symmetric).
4) Validar constraints (pair_candidates, min_confidence, requires_evidence).
5) Marcar status: ready | pending_entities | invalid.
6) Gerar summary (opcional) usando semantics.

#### Prompt (opcional)
Use apenas se for necessario mapear relation_type desconhecido ou gerar summary.
Output deve ser JSON estrito e curto.

Output schema:
{
  "normalized": [
    {
      "source_ref": "...",
      "target_ref": "...",
      "relation_type": "...",
      "relation_type_mapped_from": "...",
      "summary": "..."
    }
  ]
}

## Proposta de mapa de relacoes (relation_type + mirror)
Baseado no main-service, com espaco para extensoes:
```
parent_of     <-> child_of
child_of      <-> parent_of
sibling_of    <-> sibling_of
spouse_of     <-> spouse_of
ally_of       <-> ally_of
enemy_of      <-> enemy_of
member_of     <-> has_member
has_member    <-> member_of
leader_of     <-> led_by
led_by        <-> leader_of
located_in    <-> contains
contains      <-> located_in
owns          <-> owned_by
owned_by      <-> owns
mentor_of     <-> mentored_by
mentored_by   <-> mentor_of
```
Observacoes:
- Tipos novos podem ser aceitos como custom, mas o normalizador tenta mapear para o mais proximo do mapa oficial.
- Para relacoes simetricas (sibling_of, spouse_of, ally_of, enemy_of), o mirror e o proprio tipo.
- Para subtipos de faccao (organization, group), tratar como faction no prompt/normalizacao.

## Mapa oficial (relation.types.json) - base expandida
Exemplo de definicoes com mirror/symmetric/preferred_direction:
```
{
  "parent_of": { "mirror": "child_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is a parent of target." },
  "child_of": { "mirror": "parent_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is a child of target." },
  "ancestor_of": { "mirror": "descendant_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is an ancestor of target." },
  "descendant_of": { "mirror": "ancestor_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is a descendant of target." },

  "sibling_of": { "mirror": "sibling_of", "symmetric": true, "preferred_direction": "source_to_target", "semantics": "Source and target are siblings." },
  "spouse_of": { "mirror": "spouse_of", "symmetric": true, "preferred_direction": "source_to_target", "semantics": "Source and target are spouses/partners." },
  "lover_of": { "mirror": "lover_of", "symmetric": true, "preferred_direction": "source_to_target", "semantics": "Source and target are lovers." },
  "ex_lover_of": { "mirror": "ex_lover_of", "symmetric": true, "preferred_direction": "source_to_target", "semantics": "Source and target were lovers in the past." },

  "guardian_of": { "mirror": "ward_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is the guardian of target." },
  "ward_of": { "mirror": "guardian_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is under the guardianship of target." },

  "ally_of": { "mirror": "ally_of", "symmetric": true, "preferred_direction": "source_to_target", "semantics": "Source and target are allies." },
  "enemy_of": { "mirror": "enemy_of", "symmetric": true, "preferred_direction": "source_to_target", "semantics": "Source and target are enemies." },
  "rival_of": { "mirror": "rival_of", "symmetric": true, "preferred_direction": "source_to_target", "semantics": "Source and target are rivals." },

  "trusts": { "mirror": "trusted_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source trusts target." },
  "trusted_by": { "mirror": "trusts", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is trusted by target." },
  "betrayed": { "mirror": "betrayed_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source betrayed target." },
  "betrayed_by": { "mirror": "betrayed", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source was betrayed by target." },

  "mentor_of": { "mirror": "mentored_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source mentors or trains target." },
  "mentored_by": { "mirror": "mentor_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is mentored or trained by target." },

  "member_of": { "mirror": "has_member", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source belongs to a group or organization." },
  "has_member": { "mirror": "member_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source has target as a member." },

  "leader_of": { "mirror": "led_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source leads target." },
  "led_by": { "mirror": "leader_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is led by target." },

  "founder_of": { "mirror": "founded_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source founded target." },
  "founded_by": { "mirror": "founder_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source was founded by target." },

  "at_war_with": { "mirror": "at_war_with", "symmetric": true, "preferred_direction": "source_to_target", "semantics": "Source and target are at war." },
  "in_truce_with": { "mirror": "in_truce_with", "symmetric": true, "preferred_direction": "source_to_target", "semantics": "Source and target are in a truce." },

  "vassal_of": { "mirror": "overlord_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is a vassal of target." },
  "overlord_of": { "mirror": "vassal_of", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is an overlord of target." },

  "located_in": { "mirror": "contains", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is located in target." },
  "contains": { "mirror": "located_in", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source contains target." },

  "controls": { "mirror": "controlled_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source controls target." },
  "controlled_by": { "mirror": "controls", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is controlled by target." },

  "owns": { "mirror": "owned_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source owns target." },
  "owned_by": { "mirror": "owns", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is owned by target." },

  "wields": { "mirror": "wielded_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source wields or uses an artifact." },
  "wielded_by": { "mirror": "wields", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is wielded by target." },

  "created": { "mirror": "created_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source created target." },
  "created_by": { "mirror": "created", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source was created by target." },

  "bound_to": { "mirror": "binds", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source is bound to target." },
  "binds": { "mirror": "bound_to", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source binds target." },

  "caused": { "mirror": "caused_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source caused target." },
  "caused_by": { "mirror": "caused", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source was caused by target." },

  "participated_in": { "mirror": "has_participant", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source entity participated in an event." },
  "has_participant": { "mirror": "participated_in", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Event has source entity as a participant." },

  "triggered": { "mirror": "triggered_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source triggered target (event/entity)." },
  "triggered_by": { "mirror": "triggered", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source was triggered by target." },

  "resulted_in": { "mirror": "resulted_from", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source resulted in target (event/entity/state)." },
  "resulted_from": { "mirror": "resulted_in", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source resulted from target." },

  "prevented": { "mirror": "prevented_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source prevented target." },
  "prevented_by": { "mirror": "prevented", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source was prevented by target." },

  "revealed": { "mirror": "revealed_by", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source revealed target." },
  "revealed_by": { "mirror": "revealed", "symmetric": false, "preferred_direction": "source_to_target", "semantics": "Source was revealed by target." }
}
```

## Mapas por entidade (exemplos)
### character.relation.map.json (exemplo)
```
{
  "entity_type": "character",
  "version": 1,
  "relations": {
    "member_of": {
      "pair_candidates": ["faction", "organization", "group"],
      "description": "Character belongs to a group/organization.",
      "contexts": ["when a character joins a faction or guild", "when a character is enlisted in a guard/unit", "when a character is part of a cult or order"],
      "signals": ["member of", "joined", "part of", "enlisted", "swore loyalty to"],
      "anti_signals": ["visited", "met", "heard about"],
      "constraints": { "min_confidence": 0.55, "allow_implicit": true, "requires_evidence": true }
    },
    "leader_of": {
      "pair_candidates": ["faction", "organization", "group"],
      "description": "Character leads a group/organization.",
      "contexts": ["when a character commands a unit", "when a character rules a faction", "when a character is appointed as captain/leader"],
      "signals": ["leads", "commands", "captain of", "chief of", "rules"],
      "anti_signals": ["temporary", "acting", "substitute"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": true, "requires_evidence": true }
    },
    "mentor_of": {
      "pair_candidates": ["character"],
      "description": "Character mentors another character.",
      "contexts": ["when a character trains another", "when a master/apprentice relationship exists"],
      "signals": ["trained", "mentored", "taught", "apprentice"],
      "anti_signals": ["met once", "brief advice"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": true, "requires_evidence": true }
    },
    "ally_of": {
      "pair_candidates": ["character", "faction", "organization", "group"],
      "description": "Character is allied with another entity.",
      "contexts": ["when they cooperate toward a goal", "when alliances are declared"],
      "signals": ["allied with", "teamed up", "formed an alliance", "joined forces"],
      "anti_signals": ["neutral", "temporary truce"],
      "constraints": { "min_confidence": 0.55, "allow_implicit": true, "requires_evidence": true }
    },
    "enemy_of": {
      "pair_candidates": ["character", "faction", "organization", "group"],
      "description": "Character is enemies with another entity.",
      "contexts": ["when they are in conflict", "when rivalry/feud is explicit"],
      "signals": ["enemy", "hunted", "swore revenge", "at war with"],
      "anti_signals": ["dislikes", "annoyed"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": true, "requires_evidence": true }
    },
    "owns": {
      "pair_candidates": ["artifact", "location"],
      "description": "Character owns an item or property.",
      "contexts": ["when an item is possessed", "when property is legally owned"],
      "signals": ["owns", "belongs to", "property", "title deed"],
      "anti_signals": ["borrowed", "stole", "found"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": false, "requires_evidence": true }
    },
    "located_in": {
      "pair_candidates": ["location"],
      "description": "Character is physically located in a place.",
      "contexts": ["when the narrative places a character at a location"],
      "signals": ["in", "at", "inside", "entered", "arrived"],
      "anti_signals": ["dreamed", "remembered", "heard about"],
      "constraints": { "min_confidence": 0.5, "allow_implicit": true, "requires_evidence": true }
    }
  }
}
```

### faction.relation.map.json (exemplo)
```
{
  "entity_type": "faction",
  "version": 1,
  "relations": {
    "has_member": {
      "pair_candidates": ["character"],
      "description": "Faction has a character as a member.",
      "contexts": ["when a roster, initiation, or membership is stated"],
      "signals": ["members include", "recruited", "initiation", "joined"],
      "anti_signals": ["met", "visited"],
      "constraints": { "min_confidence": 0.55, "allow_implicit": true, "requires_evidence": true }
    },
    "led_by": {
      "pair_candidates": ["character"],
      "description": "Faction is led by a character.",
      "contexts": ["when a leader is named", "when command structure is explicit"],
      "signals": ["led by", "commanded by", "under the rule of"],
      "anti_signals": ["influenced by"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": true, "requires_evidence": true }
    },
    "ally_of": {
      "pair_candidates": ["faction", "organization", "group"],
      "description": "Faction is allied with another group.",
      "contexts": ["when alliances/treaties exist"],
      "signals": ["alliance", "treaty", "joined forces"],
      "anti_signals": ["trade", "neutral"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": true, "requires_evidence": true }
    },
    "enemy_of": {
      "pair_candidates": ["faction", "organization", "group"],
      "description": "Faction is hostile to another group.",
      "contexts": ["when factions are at war", "when hostility is declared"],
      "signals": ["at war", "enemy", "hostile", "raided"],
      "anti_signals": ["competition"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": true, "requires_evidence": true }
    },
    "contains": {
      "pair_candidates": ["location"],
      "description": "Faction controls/contains territory (locations).",
      "contexts": ["when a faction controls a city/fort/region"],
      "signals": ["controls", "holds", "occupies", "territory"],
      "anti_signals": ["visited"],
      "constraints": { "min_confidence": 0.55, "allow_implicit": true, "requires_evidence": true }
    }
  }
}
```

### location.relation.map.json (exemplo)
```
{
  "entity_type": "location",
  "version": 1,
  "relations": {
    "contains": {
      "pair_candidates": ["location", "character", "artifact", "faction", "organization", "group"],
      "description": "Location contains another entity.",
      "contexts": ["when an entity is inside a place", "when a place includes sub-locations"],
      "signals": ["inside", "within", "contained", "housed"],
      "anti_signals": ["near", "adjacent"],
      "constraints": { "min_confidence": 0.5, "allow_implicit": true, "requires_evidence": true }
    },
    "located_in": {
      "pair_candidates": ["location"],
      "description": "Location is located within another location.",
      "contexts": ["when a place is part of a larger region"],
      "signals": ["in", "within", "part of", "region of"],
      "anti_signals": ["near"],
      "constraints": { "min_confidence": 0.55, "allow_implicit": true, "requires_evidence": true }
    },
    "owned_by": {
      "pair_candidates": ["character", "faction", "organization", "group"],
      "description": "Location/property is owned by an entity.",
      "contexts": ["when a property belongs to someone"],
      "signals": ["owned by", "property of", "estate"],
      "anti_signals": ["occupied"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": false, "requires_evidence": true }
    }
  }
}
```

### event.relation.map.json (exemplo)
```
{
  "entity_type": "event",
  "version": 1,
  "relations": {
    "has_participant": {
      "pair_candidates": ["character", "faction", "organization", "group", "artifact", "location"],
      "description": "Event involves an entity as a participant.",
      "contexts": ["battles", "rituals", "meetings", "ceremonies", "discoveries"],
      "signals": ["took part in", "participated", "was present", "attended"],
      "constraints": { "min_confidence": 0.55, "allow_implicit": true, "requires_evidence": true }
    },
    "triggered": {
      "pair_candidates": ["event"],
      "description": "Event triggered another event.",
      "contexts": ["chain reactions", "escalations", "aftermath"],
      "signals": ["triggered", "led to", "sparked"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": true, "requires_evidence": true }
    },
    "resulted_in": {
      "pair_candidates": ["event", "artifact", "location", "faction", "character"],
      "description": "Event resulted in a change or new entity/state.",
      "contexts": ["aftermath", "consequences", "outcomes"],
      "signals": ["resulted in", "ended with", "gave rise to"],
      "constraints": { "min_confidence": 0.6, "allow_implicit": true, "requires_evidence": true }
    },
    "revealed": {
      "pair_candidates": ["artifact", "character", "location", "faction"],
      "description": "Event revealed hidden information or entities.",
      "contexts": ["discoveries", "revelations"],
      "signals": ["revealed", "discovered", "uncovered"],
      "constraints": { "min_confidence": 0.55, "allow_implicit": true, "requires_evidence": true }
    }
  }
}
```

## Phase5 output (candidatos)
Cada relacao candidata inclui:
- source/target com ref (finding:* ou match:*).
- relation_type, confidence, implicit, polarity.
- evidence com span_id + quote curta.
Notas:
- polarity: asserted | denied | uncertain
- quote curta e verbatim

### Exemplo Phase5 output
```
{
  "relations": [
    {
      "source": { "ref": "finding:character:0", "type": "character" },
      "target": { "ref": "match:faction:uuid", "type": "faction" },
      "relation_type": "member_of",
      "polarity": "asserted",
      "implicit": false,
      "confidence": 0.78,
      "evidence": { "span_id": "span:2", "quote": "Ari swore loyalty to the Order of the Sun." }
    }
  ]
}
```

## Phase6 output (normalizadas)
Cada relacao normalizada inclui:
- source/target com ref + id (quando resolvido).
- relation_type normalizado, direction, create_mirror.
- status: ready | pending_entities | invalid.
- dedup: is_duplicate + reason.
- summary gerado (opcional).

### Exemplo Phase6 output
```
{
  "source": { "ref": "finding:character:0", "id": "uuid-src" },
  "target": { "ref": "match:faction:uuid", "id": "uuid-tgt" },
  "relation_type": "member_of",
  "direction": "source_to_target",
  "create_mirror": true,
  "confidence": 0.78,
  "polarity": "asserted",
  "implicit": false,
  "evidence": { "span_id": "span:2", "quote": "Ari swore loyalty to the Order of the Sun." },
  "status": "ready",
  "dedup": { "is_duplicate": false, "reason": "" },
  "summary": "Ari is a member of Order of the Sun."
}
```

## Phase7 (relation_match) - dedup/lookup
- Buscar relacoes existentes no main-service para dedup/conflito.
- Precisamos decidir o endpoint/consulta e payload minimo.

#### Regras sugeridas
- Dedup por (source_id, target_id, relation_type, context_type, context_id).
- Se mirror aplicavel, checar ambas orientacoes.
- Marcar dedup.is_duplicate=true sem bloquear a UI (status pode continuar ready).

## Relacoes e HITL
- O gateway so sugere relacoes; persistencia ocorre apos confirmacao humana.
- A UI deve permitir aceitar/rejeitar/ajustar tipo/direcao.

## Feedback loop (HITL)
- Entidades e relacoes nao sao criadas automaticamente.
- O front-end expõe: Aceitar / Rejeitar / Ajustar.
- Persistencia acontece somente apos acao humana.

## Contrato front-end (resumo)
- Cada no de relacao tem `ref` e `id` opcional.
- status calculado no gateway:
  - ready: pode persistir
  - pending_entities: falta id
  - invalid: nao persistir
- Regra de resolucao: `resolveId(node) = node.id ?? ref_map[node.ref] ?? null`.
- Persistencia habilita quando ambos IDs resolvidos e status != invalid.

### Instrucoes detalhadas para o front-end
1) Cada no tem `ref` sempre; `id` quando resolvido.
2) Manter `ref_map` local:
```
{ "ref_map": { "finding:character:0": "uuid-real-character" } }
```
3) `resolveId(node) = node.id ?? ref_map[node.ref] ?? null`.
4) Habilitar persistencia se ambos IDs resolvidos e status != invalid.
5) Ao persistir:
   - Substituir refs por IDs resolvidos.
   - Enviar create_mirror quando aplicavel.
6) Para custom:*, gateway retorna dois edges; UI cria ambos se confirmado.

## Mapas por entidade (restricoes)
- `relation.types.json` define mirror/symmetric/preferred_direction.
- `*.relation.map.json` define pair_candidates, sinais e constraints por source_type.
- O gateway injeta apenas mapas relevantes para manter prompts pequenos.

## Relacoes para eventos (extensoes)
- Eventos sao entidades de primeira classe e participam da tabela unificada.
- Tipos sugeridos: participated_in/has_participant, triggered/triggered_by, caused/caused_by, resulted_in/resulted_from, prevented/prevented_by, revealed/revealed_by.

### Tipos de relacao para eventos (sugeridos)
- participated_in <-> has_participant
- triggered <-> triggered_by
- caused <-> caused_by
- resulted_in <-> resulted_from
- prevented <-> prevented_by
- revealed <-> revealed_by

## Etapas de acao (consolidado)
1) **Definir mapas oficiais**
   - Criar/atualizar `relation.types.json` e `*.relation.map.json` no main-service.
   - Carregar mapas em memoria no bootstrap e preparar exportacao estatica (opcional).

2) **Atualizar pipeline do llm-gateway**
   - Phase4 (atual) continua como formatador de entities -> `result_entities`.
   - Adicionar Phase5 relation_discovery (prompt + usecase + eventos).
   - Adicionar Phase6 relation_normalize (mapa oficial + custom + create_mirror).
   - Adicionar Phase7 relation_match (dedup/lookup no main-service; definir endpoint/consulta).
   - Phase8 result_relations (retorno no stream e no endpoint nao-stream).

3) **Integracao com ingestao de relations**
   - Ajustar contrato de criacao: UI chama main-service com `create_mirror` quando permitido.
   - Para custom:*, gateway retorna dois edges (direto + espelho) para o UI decidir criar ambos.

4) **Main-service**
   - Usar mapa oficial para validacao opcional de relation_type.
   - Expor endpoint estatico dos mapas para o gateway/UI (se adotado).
   - Preparar endpoint de consulta de relacoes existentes para Phase7 (se adotado).

5) **Documentacao e contratos**
   - Atualizar contrato do endpoint nao-stream (entities + relations).
   - Documentar SSE (`result_entities`, `result_relations`, `result` legado).

## Dependencias
- Mapa oficial precisa estar definido e alinhado com o gateway.
- Reuso dos modelos e prompts de extracao existentes.

## Riscos e Mitigacoes
- **Relacoes inventadas**: aplicar thresholds de confidence e exigir evidencias.
- **Ambiguidade de direcao**: normalizador escolhe direcao preferida e usa mirror.
- **Tipos novos**: fallback para custom e telemetria para revisao.

## Proximos Passos
- Definir e consolidar o mapa oficial de relation_type + mirror.
- Implementar Phase5 a Phase8 no llm-gateway-service.
- Atualizar main-service para usar/validar o mapa.
- Ajustar testes e adicionar coverage para as novas fases.
