# World View Tabs e Character Details View - Plano de Implementação

## Status

| Fase | Descrição | Status |
|------|-----------|--------|
| Fase 0 | API Backend | ✅ COMPLETA |
| Fase 1 | Tipos e API Client | 🔲 Pendente |
| Fase 2 | World View Tabs | 🔲 Pendente |
| Fase 3 | Character Details View | 🔲 Pendente |
| Fase 4 | Botões de Relacionamento | 🔲 Pendente |

---

## Objetivos

1. Adicionar 3 novas tabs na World View: Archetype, Lore, Factions
2. Reordenar tabs na ordem: Characters, Traits, Archetype, Events, Lore, Locations, Factions, Artifacts
3. Criar Character Details View com Traits, Events, Scenes e Relationships
4. Implementar TimeConfig na Events tab com visualização
5. Implementar Timeline visual com D3.js
6. Criar auto-relacionamento Character-Character (teia de personagens)
7. Adicionar botões de ação para criar relacionamentos entre entidades

---

## Endpoints Disponíveis (API Completa)

### Archetype
- `GET /api/v1/archetypes` - Listar archetypes
- `GET /api/v1/archetypes/{id}` - Buscar archetype
- `POST /api/v1/archetypes` - Criar archetype
- `PUT /api/v1/archetypes/{id}` - Atualizar archetype
- `DELETE /api/v1/archetypes/{id}` - Deletar archetype
- `GET /api/v1/archetypes/{id}/traits` - Listar traits do archetype
- `POST /api/v1/archetypes/{id}/traits` - Adicionar trait
- `DELETE /api/v1/archetypes/{id}/traits/{trait_id}` - Remover trait

### Faction
- `GET /api/v1/worlds/{world_id}/factions` - Listar factions
- `GET /api/v1/factions/{id}` - Buscar faction
- `POST /api/v1/worlds/{world_id}/factions` - Criar faction
- `PUT /api/v1/factions/{id}` - Atualizar faction
- `DELETE /api/v1/factions/{id}` - Deletar faction
- `GET /api/v1/factions/{id}/children` - Listar sub-factions
- `GET /api/v1/factions/{id}/references` - Listar referências
- `POST /api/v1/factions/{id}/references` - Adicionar referência
- `PUT /api/v1/faction-references/{id}` - Atualizar referência
- `DELETE /api/v1/factions/{id}/references/{entity_type}/{entity_id}` - Remover referência

### Lore
- `GET /api/v1/worlds/{world_id}/lores` - Listar lores
- `GET /api/v1/lores/{id}` - Buscar lore
- `POST /api/v1/worlds/{world_id}/lores` - Criar lore
- `PUT /api/v1/lores/{id}` - Atualizar lore
- `DELETE /api/v1/lores/{id}` - Deletar lore
- `GET /api/v1/lores/{id}/children` - Listar sub-lores
- `GET /api/v1/lores/{id}/references` - Listar referências
- `POST /api/v1/lores/{id}/references` - Adicionar referência
- `PUT /api/v1/lore-references/{id}` - Atualizar referência
- `DELETE /api/v1/lores/{id}/references/{entity_type}/{entity_id}` - Remover referência

### Character Traits
- `GET /api/v1/characters/{id}/traits` - Listar traits
- `POST /api/v1/characters/{id}/traits` - Adicionar trait
- `PUT /api/v1/characters/{id}/traits/{trait_id}` - Atualizar trait
- `DELETE /api/v1/characters/{id}/traits/{trait_id}` - Remover trait

### Character Events
- `GET /api/v1/characters/{id}/events` - Listar eventos do character

### Character Relationships
- `GET /api/v1/characters/{id}/relationships` - Listar relacionamentos
- `POST /api/v1/characters/{id}/relationships` - Criar relacionamento
- `PUT /api/v1/character-relationships/{id}` - Atualizar relacionamento
- `DELETE /api/v1/character-relationships/{id}` - Deletar relacionamento

### Event Characters/References
- `GET /api/v1/events/{id}/characters` - Listar characters
- `POST /api/v1/events/{id}/characters` - Adicionar character
- `DELETE /api/v1/events/{id}/characters/{character_id}` - Remover character
- `GET /api/v1/events/{id}/references` - Listar referências
- `POST /api/v1/events/{id}/references` - Adicionar referência
- `PUT /api/v1/event-references/{id}` - Atualizar referência
- `DELETE /api/v1/events/{id}/references/{entity_type}/{entity_id}` - Remover referência

### Scene References
- `GET /api/v1/scenes/{id}/references` - Listar referências
- `POST /api/v1/scenes/{id}/references` - Adicionar referência
- `DELETE /api/v1/scenes/{id}/references/{entity_type}/{entity_id}` - Remover referência

### Timeline
- `GET /api/v1/worlds/{world_id}/timeline?from=&to=` - Buscar timeline
- `PUT /api/v1/events/{id}/epoch` - Marcar evento como epoch
- `GET /api/v1/events/{id}/ancestors` - Ancestrais do evento
- `GET /api/v1/events/{id}/descendants` - Descendentes do evento

### World TimeConfig
- `PUT /api/v1/worlds/{id}` - Atualizar world (inclui time_config)

---

## Fase 1 - Tipos e API Client

### 1.1 Adicionar tipos TypeScript

**Arquivo**: `obsidian-plugin/src/types.ts`

```typescript
// Archetype - global por tenant (NAO tem world_id)
export interface Archetype {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

// ArchetypeTrait - relaciona archetype com trait (template)
export interface ArchetypeTrait {
  id: string;
  archetype_id: string;
  trait_id: string;
  default_value: string;
  created_at: string;
}

// Faction - hierarquica, pertence a um world
export interface Faction {
  id: string;
  tenant_id: string;
  world_id: string;
  parent_id?: string | null;
  name: string;
  type?: string | null;
  description: string;
  beliefs: string;
  structure: string;
  symbols: string;
  hierarchy_level: number;
  created_at: string;
  updated_at: string;
}

// Lore - hierarquica, pertence a um world
export interface Lore {
  id: string;
  tenant_id: string;
  world_id: string;
  parent_id?: string | null;
  name: string;
  category?: string | null;
  description: string;
  rules: string;
  limitations: string;
  requirements: string;
  hierarchy_level: number;
  created_at: string;
  updated_at: string;
}

// CharacterTrait - trait atribuido a character com value customizado
export interface CharacterTrait {
  id: string;
  character_id: string;
  trait_id: string;
  trait_name: string;
  trait_category: string;
  trait_description: string;
  value: string;
  notes: string;
  created_at: string;
  updated_at: string;
}

// CharacterRelationship - auto-relacionamento entre characters
export interface CharacterRelationship {
  id: string;
  tenant_id: string;
  character1_id: string;
  character2_id: string;
  relationship_type: string; // "ally", "enemy", "family", "lover", "rival", "mentor", "student"
  description: string;
  bidirectional: boolean;
  created_at: string;
  updated_at: string;
}

// EventCharacter - relaciona evento com character
export interface EventCharacter {
  id: string;
  event_id: string;
  character_id: string;
  role?: string | null;
  created_at: string;
}

// EventReference - relaciona evento com entidades
export interface EventReference {
  id: string;
  event_id: string;
  entity_type: "character" | "location" | "artifact" | "faction";
  entity_id: string;
  relationship_type?: string | null;
  notes: string;
  created_at: string;
}

// SceneReference - relaciona scene com entity
export interface SceneReference {
  id: string;
  scene_id: string;
  entity_type: "character" | "location" | "artifact";
  entity_id: string;
  created_at: string;
}

// FactionReference - relaciona faction com outras entidades
export interface FactionReference {
  id: string;
  faction_id: string;
  entity_type: string;
  entity_id: string;
  role?: string | null;
  notes: string;
  created_at: string;
}

// LoreReference - relaciona lore com outras entidades
export interface LoreReference {
  id: string;
  lore_id: string;
  entity_type: string;
  entity_id: string;
  relationship_type?: string | null;
  notes: string;
  created_at: string;
}

// TimeConfig - configuracao de calendario do world
export interface TimeConfig {
  base_unit: string;
  hours_per_day: number;
  days_per_week: number;
  days_per_year: number;
  months_per_year: number;
  month_lengths?: number[];
  month_names?: string[];
  day_names?: string[];
  era_name?: string;
  year_zero?: number;
}

// WorldDate - data no calendario do world
export interface WorldDate {
  year: number;
  month: number;
  day: number;
  hour: number;
  minute: number;
}
```

### 1.2 Adicionar métodos na API Client

**Arquivo**: `obsidian-plugin/src/api/client.ts`

Métodos a adicionar (~35 métodos):

#### Archetype
- `getArchetypes(): Promise<Archetype[]>`
- `getArchetype(id: string): Promise<Archetype>`
- `createArchetype(data: Partial<Archetype>): Promise<Archetype>`
- `updateArchetype(id: string, data: Partial<Archetype>): Promise<Archetype>`
- `deleteArchetype(id: string): Promise<void>`
- `getArchetypeTraits(archetypeId: string): Promise<ArchetypeTrait[]>`
- `addArchetypeTrait(archetypeId: string, traitId: string, defaultValue?: string): Promise<void>`
- `removeArchetypeTrait(archetypeId: string, traitId: string): Promise<void>`

#### Faction
- `getFactions(worldId: string): Promise<Faction[]>`
- `getFaction(id: string): Promise<Faction>`
- `createFaction(worldId: string, data: Partial<Faction>): Promise<Faction>`
- `updateFaction(id: string, data: Partial<Faction>): Promise<Faction>`
- `deleteFaction(id: string): Promise<void>`
- `getFactionChildren(id: string): Promise<Faction[]>`
- `getFactionReferences(factionId: string): Promise<FactionReference[]>`
- `addFactionReference(factionId: string, entityType: string, entityId: string, role?: string, notes?: string): Promise<FactionReference>`
- `updateFactionReference(id: string, role?: string, notes?: string): Promise<FactionReference>`
- `removeFactionReference(factionId: string, entityType: string, entityId: string): Promise<void>`

#### Lore
- `getLores(worldId: string): Promise<Lore[]>`
- `getLore(id: string): Promise<Lore>`
- `createLore(worldId: string, data: Partial<Lore>): Promise<Lore>`
- `updateLore(id: string, data: Partial<Lore>): Promise<Lore>`
- `deleteLore(id: string): Promise<void>`
- `getLoreChildren(id: string): Promise<Lore[]>`
- `getLoreReferences(loreId: string): Promise<LoreReference[]>`
- `addLoreReference(loreId: string, entityType: string, entityId: string, relationshipType?: string, notes?: string): Promise<LoreReference>`
- `updateLoreReference(id: string, relationshipType?: string, notes?: string): Promise<LoreReference>`
- `removeLoreReference(loreId: string, entityType: string, entityId: string): Promise<void>`

#### Character Traits
- `getCharacterTraits(characterId: string): Promise<CharacterTrait[]>`
- `addCharacterTrait(characterId: string, traitId: string, value?: string, notes?: string): Promise<CharacterTrait>`
- `updateCharacterTrait(characterId: string, traitId: string, value?: string, notes?: string): Promise<CharacterTrait>`
- `removeCharacterTrait(characterId: string, traitId: string): Promise<void>`

#### Character Events
- `getCharacterEvents(characterId: string): Promise<EventCharacter[]>`

#### Character Relationships
- `getCharacterRelationships(characterId: string): Promise<CharacterRelationship[]>`
- `createCharacterRelationship(characterId: string, data: {...}): Promise<CharacterRelationship>`
- `updateCharacterRelationship(id: string, data: Partial<CharacterRelationship>): Promise<CharacterRelationship>`
- `deleteCharacterRelationship(id: string): Promise<void>`

#### Event Characters/References
- `getEventCharacters(eventId: string): Promise<EventCharacter[]>`
- `addEventCharacter(eventId: string, characterId: string, role?: string): Promise<EventCharacter>`
- `removeEventCharacter(eventId: string, characterId: string): Promise<void>`
- `getEventReferences(eventId: string): Promise<EventReference[]>`
- `addEventReference(eventId: string, entityType: string, entityId: string, relationshipType?: string, notes?: string): Promise<EventReference>`
- `updateEventReference(id: string, relationshipType?: string, notes?: string): Promise<EventReference>`
- `removeEventReference(eventId: string, entityType: string, entityId: string): Promise<void>`

#### Scene References
- `getSceneReferences(sceneId: string): Promise<SceneReference[]>`
- `addSceneReference(sceneId: string, entityType: string, entityId: string): Promise<SceneReference>`
- `removeSceneReference(sceneId: string, entityType: string, entityId: string): Promise<void>`

#### Timeline
- `getTimeline(worldId: string, fromPos?: number, toPos?: number): Promise<WorldEvent[]>`

#### World TimeConfig
- `updateWorldTimeConfig(worldId: string, timeConfig: TimeConfig): Promise<World>`

---

## Fase 2 - World View Tabs

### 2.1 Alterar tipo de worldTab

**Arquivo**: `obsidian-plugin/src/views/StoryListView.ts`

```typescript
// Antes:
worldTab: "characters" | "locations" | "artifacts" | "events" | "traits" = "characters";

// Depois:
worldTab: "characters" | "traits" | "archetypes" | "events" | "lore" | "locations" | "factions" | "artifacts" = "characters";
```

### 2.2 Reordenar tabs em renderWorldTabs()

```typescript
const tabs: { key: WorldTabKey; label: string }[] = [
  { key: "characters", label: "Characters" },
  { key: "traits", label: "Traits" },
  { key: "archetypes", label: "Archetypes" },
  { key: "events", label: "Events" },
  { key: "lore", label: "Lore" },
  { key: "locations", label: "Locations" },
  { key: "factions", label: "Factions" },
  { key: "artifacts", label: "Artifacts" },
];
```

### 2.3 Adicionar propriedades de estado

```typescript
archetypes: Archetype[] = [];
factions: Faction[] = [];
lores: Lore[] = [];
```

### 2.4 Atualizar loadWorldData()

```typescript
this.archetypes = await this.plugin.apiClient.getArchetypes(); // global por tenant
this.factions = await this.plugin.apiClient.getFactions(this.currentWorld.id);
this.lores = await this.plugin.apiClient.getLores(this.currentWorld.id);
```

### 2.5 Criar renderArchetypesTab()

- Lista de archetypes com nome e descrição
- Botões Edit/Delete por item
- Botão "Create Archetype"
- Ao clicar, pode abrir modal com traits do archetype

### 2.6 Criar renderLoreTab()

- Renderização hierárquica (parent_id, hierarchy_level)
- Campos: name, category (pill), description
- Modal de detalhes: rules, limitations, requirements
- Botão "Create Lore" / "Create Sub-Lore"

### 2.7 Criar renderFactionsTab()

- Renderização hierárquica (parent_id, hierarchy_level)
- Campos: name, type (pill), description
- Modal de detalhes: beliefs, structure, symbols
- Botão "Create Faction" / "Create Sub-Faction"

### 2.8 Atualizar renderWorldTabContent()

```typescript
switch (this.worldTab) {
  case "characters": this.renderCharactersTab(contentContainer); break;
  case "traits": this.renderTraitsTab(contentContainer); break;
  case "archetypes": this.renderArchetypesTab(contentContainer); break;
  case "events": this.renderEventsTab(contentContainer); break;
  case "lore": this.renderLoreTab(contentContainer); break;
  case "locations": this.renderLocationsTab(contentContainer); break;
  case "factions": this.renderFactionsTab(contentContainer); break;
  case "artifacts": this.renderArtifactsTab(contentContainer); break;
}
```

### 2.9 Events Tab - TimeConfig

Adicionar seções no renderEventsTab():
1. **Seção TimeConfig**: mostrar configuração atual ou botão para adicionar
2. **Botão "View Timeline"**: abre modal com D3.js
3. **Lista de eventos**: como antes, com botões de relacionamento

### 2.10 Timeline Modal com D3.js

**Novo arquivo**: `obsidian-plugin/src/views/TimelineModal.ts`

- Instalar: `npm install d3 @types/d3`
- Gráfico com eixo X (timeline_position) e Y (importance)
- Círculos para eventos, cor por tipo
- Linhas conectando pai-filho
- Labels para eventos importance >= 8
- Click em evento mostra detalhes

---

## Fase 3 - Character Details View

### 3.1 Estado e navegação

```typescript
currentCharacter: Character | null = null;
characterTab: "overview" | "traits" | "events" | "scenes" | "relationships" = "overview";
characterTraits: CharacterTrait[] = [];
characterEvents: { event: WorldEvent; role?: string }[] = [];
characterScenes: { scene: Scene; story: Story; type: "pov" | "coadjuvante" }[] = [];
characterRelationships: CharacterRelationship[] = [];

viewMode: "list" | "details" | "world-details" | "character-details" = "list";
```

### 3.2 Navegação

```typescript
async showCharacterDetails(character: Character) {
  this.currentCharacter = character;
  this.viewMode = "character-details";
  this.characterTab = "overview";
  await this.loadCharacterData();
  this.renderCharacterDetails();
}
```

Modificar `renderCharactersTab()` para click chamar `showCharacterDetails()`.

### 3.3 Header

- Botão "← Back to World"
- Título: character.name
- Archetype pill (se tiver)
- UUID com ícone de cópia
- Menu: "Edit Character"

### 3.4 Tab Overview

- Description (textarea editável)
- Archetype dropdown: lista archetypes + "None"

### 3.5 Tab Traits

- Lista de CharacterTraits
- Cada item: trait_name, trait_category (pill), value, notes
- Botões: Save, Delete
- Botão "Add Trait" com modal

### 3.6 Tab Events

- Lista de eventos do character (via getCharacterEvents)
- Cada item: nome, role
- Botões: Edit Role, Remove
- Botão "Add Event" com modal

### 3.7 Tab Scenes

**Workaround** (sem endpoint otimizado):
1. Buscar stories do world
2. Para cada story, buscar scenes
3. Filtrar por pov_character_id (tipo: "pov")
4. Filtrar SceneReferences (tipo: "coadjuvante")

Lista: goal, story name, tipo pill ("POV" verde / "Coadjuvante" azul)

### 3.8 Tab Relationships

- Lista de CharacterRelationships
- Cada item: nome do outro character, tipo (pill), direção (↔/→), description
- Botões: Edit, Delete
- Botão "Add Relationship" com modal:
  - Dropdown characters (exceto atual)
  - Dropdown relationship_type
  - Campo description
  - Checkbox bidirectional

### 3.9 loadCharacterData()

```typescript
async loadCharacterData() {
  this.characterTraits = await this.plugin.apiClient.getCharacterTraits(this.currentCharacter.id);
  this.characterEvents = await this.plugin.apiClient.getCharacterEvents(this.currentCharacter.id);
  this.characterRelationships = await this.plugin.apiClient.getCharacterRelationships(this.currentCharacter.id);
  // Scenes: workaround iterando stories/scenes
}
```

---

## Fase 4 - Botões de Relacionamento

### 4.1 Character Item

```typescript
// Botão relacionamento
const relBtn = actionsDiv.createEl("button");
setIcon(relBtn, "users");
relBtn.title = "Add Relationship";
relBtn.onclick = () => this.showAddCharacterRelationshipModal(character);
```

### 4.2 Event Item

```typescript
// Link to Faction
const factionBtn = actionsDiv.createEl("button");
setIcon(factionBtn, "flag");
factionBtn.onclick = () => this.showAddEventReferenceModal(event, "faction");

// Set Parent Event
const eventLinkBtn = actionsDiv.createEl("button");
setIcon(eventLinkBtn, "git-branch");
eventLinkBtn.onclick = () => this.showSetEventParentModal(event);
```

### 4.3 Faction Item

```typescript
// Link to Entity
const linkBtn = actionsDiv.createEl("button");
setIcon(linkBtn, "link");
linkBtn.onclick = () => this.showAddFactionReferenceModal(faction);

// Create Sub-Faction
const subBtn = actionsDiv.createEl("button");
setIcon(subBtn, "folder-plus");
subBtn.onclick = () => this.showCreateFactionModal(faction.id);
```

### 4.4 Tabela de Relacionamentos

| Entidade | Pode relacionar com | Via |
|----------|---------------------|-----|
| Character | Character | CharacterRelationship |
| Character | Trait | CharacterTrait |
| Character | Event | EventCharacter |
| Event | Character | EventCharacter |
| Event | Faction | EventReference |
| Event | Event | Event.parent_id |
| Faction | Character/Location/Artifact/Event | FactionReference |
| Faction | Faction | Faction.parent_id |
| Lore | Qualquer entidade | LoreReference |

---

## Arquivos a Modificar

| Arquivo | Alterações |
|---------|-----------|
| `obsidian-plugin/src/types.ts` | ~12 tipos novos |
| `obsidian-plugin/src/api/client.ts` | ~35 métodos novos |
| `obsidian-plugin/src/views/StoryListView.ts` | Tabs, Character Details, botões |
| `obsidian-plugin/src/views/TimelineModal.ts` | CRIAR com D3.js |
| `obsidian-plugin/styles.css` | Estilos timeline, time-config, grids |
| `obsidian-plugin/package.json` | Adicionar d3, @types/d3 |

---

## Considerações

### Pontos Críticos

1. **Archetype**: Global por tenant (NÃO tem world_id)
2. **Hierarquia**: Factions, Lore, Events e Locations têm parent/child
3. **D3.js**: Verificar compatibilidade com Obsidian
4. **Scenes**: Workaround necessário até criar endpoint otimizado

### Estilos CSS Necessários

```css
.story-engine-timeline-modal { width: 90vw; max-width: 900px; }
.story-engine-timeline-chart { overflow-x: auto; }
.story-engine-timeline-chart svg { background: var(--background-secondary); border-radius: 8px; }
.story-engine-time-config-section { background: var(--background-secondary); padding: 16px; border-radius: 8px; margin-bottom: 16px; }
.story-engine-time-config-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 8px; margin-top: 8px; }
```

