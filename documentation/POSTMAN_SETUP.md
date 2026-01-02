# Postman Collection Setup Guide

Este guia explica como importar e usar a coleção Postman para testar a API do Story Engine.

## Arquivos Disponíveis

1. **Story_Engine_API.postman_collection.json** - Coleção completa com todos os endpoints
2. **Story_Engine_Local.postman_environment.json** - Ambiente local com variáveis pré-configuradas

## Como Importar

### 1. Importar a Coleção

1. Abra o Postman
2. Clique em **Import** (canto superior esquerdo)
3. Arraste o arquivo `Story_Engine_API.postman_collection.json` ou clique em **Upload Files**
4. Clique em **Import**

### 2. Importar o Ambiente

1. No Postman, clique no ícone de **engrenagem** (⚙️) no canto superior direito
2. Clique em **Import**
3. Arraste o arquivo `Story_Engine_Local.postman_environment.json`
4. Clique em **Import**
5. Selecione o ambiente **"Story Engine - Local"** no dropdown no canto superior direito

## Como Usar

### Variáveis Automáticas

A coleção usa **test scripts** que automaticamente extraem IDs das respostas e salvam em variáveis:

#### Story Module
- `tenant_id` - ID do tenant criado
- `story_id` - ID da história criada
- `chapter_id` - ID do capítulo criado
- `scene_id` - ID da cena criada
- `beat_id` - ID do beat criado
- `prose_block_id` - ID do bloco de prosa criado
- `prose_block_reference_id` - ID da referência de bloco de prosa criada
- `image_block_id` - ID do bloco de imagem criado
- `scene_reference_id` - ID da referência de cena criada

#### World Building Module
- `world_id` - ID do mundo criado
- `trait_id` - ID do traço criado
- `archetype_id` - ID do arquétipo criado
- `location_id` - ID da localização criada
- `character_id` - ID do personagem criado
- `character_trait_id` - ID do traço do personagem
- `artifact_id` - ID do artefato criado
- `artifact_reference_id` - ID da referência de artefato criada
- `event_id` - ID do evento criado

#### RPG Module
- `rpg_system_id` - ID do sistema RPG criado
- `character_rpg_stats_id` - ID das estatísticas RPG do personagem criadas
- `artifact_rpg_stats_id` - ID das estatísticas RPG do artefato criadas
- `skill_id` - ID da habilidade criada
- `character_skill_id` - ID da habilidade do personagem criada
- `rpg_class_id` - ID da classe RPG criada
- `rpg_class_skill_id` - ID da habilidade da classe RPG criada
- `inventory_item_id` - ID do item de inventário criado
- `character_inventory_id` - ID do inventário do personagem criado

#### Workflow Variables
- `workflow_tenant_id` - ID do tenant do workflow completo
- `workflow_story_id` - ID da história do workflow completo
- `workflow_chapter_id` - ID do capítulo do workflow completo
- `workflow_scene_id` - ID da cena do workflow completo
- `workflow_beat_id` - ID do beat do workflow completo
- `workflow_cloned_story_id` - ID da história clonada do workflow completo

**Nota:** Todas essas variáveis são preenchidas automaticamente pelos test scripts quando você executa os requests de criação correspondentes. Você pode verificar os valores salvos no **Postman Console** ou editando as variáveis da coleção.

### Fluxo Recomendado

1. **Health Check** - Verifica se o servidor está rodando
2. **Create Tenant** - Cria um tenant (salva `tenant_id` automaticamente)
3. **Create Story** - Cria uma história (usa `tenant_id` do header, salva `story_id`)
4. **Create Chapter** - Cria um capítulo (usa `story_id`, salva `chapter_id`)
5. **Create Scene** - Cria uma cena (usa `chapter_id`, salva `scene_id`)
6. **Create Beat** - Cria um beat (usa `scene_id`, salva `beat_id`)

### Workflow Completo

A coleção inclui uma pasta **"Complete Workflow"** que executa todo o fluxo em sequência:

1. Cria Tenant
2. Cria Story
3. Cria Chapter
4. Cria Scene
5. Cria Beat
6. Clona Story

Você pode executar toda a pasta de uma vez usando **"Run Collection"**.

## Estrutura da Coleção

```
Story Engine API
├── Health Check
│   └── Health Check
├── Tenants
│   ├── Create Tenant (salva tenant_id)
│   └── Get Tenant
├── Stories
│   ├── Create Story (salva story_id)
│   ├── Get Story
│   ├── List Stories
│   ├── Update Story
│   └── Clone Story (salva cloned_story_id)
├── Chapters
│   ├── Create Chapter (salva chapter_id)
│   ├── Get Chapter
│   ├── List Chapters
│   ├── Update Chapter
│   └── Delete Chapter
├── Scenes
│   ├── Create Scene (salva scene_id)
│   ├── Get Scene
│   ├── List Scenes
│   ├── Update Scene
│   ├── Delete Scene
│   ├── Get Scene References
│   ├── Add Scene Reference (salva scene_reference_id)
│   └── Remove Scene Reference
├── Beats
│   ├── Create Beat (salva beat_id)
│   ├── Get Beat
│   ├── List Beats
│   ├── Update Beat
│   └── Delete Beat
├── Prose Blocks
│   ├── Create Prose Block (salva prose_block_id)
│   ├── Get Prose Block
│   ├── Update Prose Block
│   ├── Delete Prose Block
│   ├── Create Prose Block Reference (salva prose_block_reference_id)
│   └── Delete Prose Block Reference
├── Image Blocks
│   ├── Create Image Block (salva image_block_id)
│   ├── Get Image Block
│   ├── Update Image Block
│   ├── List Image Blocks
│   └── Delete Image Block
├── World Building
│   ├── Worlds
│   │   ├── Create World (salva world_id)
│   │   ├── List Worlds
│   │   ├── Get World
│   │   ├── Update World
│   │   └── Delete World
│   ├── Traits
│   │   ├── Create Trait (salva trait_id)
│   │   ├── List Traits
│   │   ├── Get Trait
│   │   ├── Update Trait
│   │   └── Delete Trait
│   ├── Archetypes
│   │   ├── Create Archetype (salva archetype_id)
│   │   ├── List Archetypes
│   │   ├── Get Archetype
│   │   ├── Update Archetype
│   │   ├── Delete Archetype
│   │   ├── Add Trait to Archetype
│   │   └── Remove Trait from Archetype
│   ├── Locations
│   │   ├── Create Location (salva location_id)
│   │   ├── List Locations
│   │   ├── Get Location
│   │   ├── Update Location
│   │   ├── Delete Location
│   │   ├── Get Children
│   │   ├── Get Ancestors
│   │   ├── Get Descendants
│   │   └── Move Location
│   ├── Characters
│   │   ├── Create Character (salva character_id)
│   │   ├── List Characters
│   │   ├── Get Character
│   │   ├── Update Character
│   │   ├── Delete Character
│   │   ├── Get Character Traits
│   │   ├── Add Trait from Template
│   │   ├── Update Character Trait
│   │   ├── Remove Trait from Character
│   │   └── Apply Archetype
│   ├── Artifacts
│   │   ├── Create Artifact (salva artifact_id)
│   │   ├── List Artifacts
│   │   ├── Get Artifact
│   │   ├── Update Artifact
│   │   ├── Delete Artifact
│   │   ├── Get Artifact References
│   │   ├── Add Artifact Reference (salva artifact_reference_id)
│   │   └── Remove Artifact Reference
│   └── Events
│       ├── Create Event (salva event_id)
│       ├── List Events
│       ├── Get Event
│       ├── Update Event
│       ├── Delete Event
│       ├── Add Character to Event
│       ├── Remove Character from Event
│       ├── Get Event Characters
│       ├── Add Location to Event
│       ├── Remove Location from Event
│       ├── Get Event Locations
│       ├── Add Artifact to Event
│       ├── Remove Artifact from Event
│       ├── Get Event Artifacts
│       └── Get Event Stat Changes
├── RPG Module
│   ├── RPG Systems
│   │   ├── List RPG Systems
│   │   ├── Get RPG System
│   │   ├── Create Custom System (salva rpg_system_id)
│   │   ├── Update Custom System
│   │   └── Delete Custom System
│   ├── Character RPG Stats
│   │   ├── Create Initial Stats (salva character_rpg_stats_id)
│   │   ├── Get Current Stats
│   │   ├── Get Stats History
│   │   ├── Activate Stats Version
│   │   └── Delete Character Stats
│   ├── Artifact RPG Stats
│   │   ├── Create Initial Stats (salva artifact_rpg_stats_id)
│   │   ├── Get Current Stats
│   │   ├── Get Stats History
│   │   └── Activate Stats Version
│   ├── Skills
│   │   ├── List Skills
│   │   ├── Create Skill (salva skill_id)
│   │   ├── Get Skill
│   │   ├── Update Skill
│   │   └── Delete Skill
│   ├── Character Skills
│   │   ├── Learn Skill (salva character_skill_id)
│   │   ├── List Character Skills
│   │   ├── Update Character Skill
│   │   └── Remove Character Skill
│   ├── Classes
│   │   ├── List Classes
│   │   ├── Create Class (salva rpg_class_id)
│   │   ├── Get RPG Class
│   │   ├── Update RPG Class
│   │   ├── Delete RPG Class
│   │   ├── Add Skill to Class (salva rpg_class_skill_id)
│   │   ├── Remove Skill from Class
│   │   ├── List Class Skills
│   │   └── Change Character Class
│   └── Inventory
│       ├── List Inventory Items
│       ├── Create Inventory Item (salva inventory_item_id)
│       ├── List Character Inventory
│       ├── Add Item to Inventory (salva character_inventory_id)
│       ├── Update Inventory Item
│       └── Remove Item from Inventory
└── Complete Workflow
    ├── 1. Create Tenant
    ├── 2. Create Story
    ├── 3. Create Chapter
    ├── 4. Create Scene
    ├── 5. Create Beat
    └── 6. Clone Story
```

## Configuração

### Base URL

Por padrão, a coleção usa `http://localhost:8080`. Para mudar:

1. Selecione o ambiente **"Story Engine - Local"**
2. Edite a variável `base_url`
3. Ou edite diretamente na coleção (variável de coleção)

### Headers Automáticos

A coleção automaticamente adiciona:
- `Content-Type: application/json` em requisições POST/PUT
- `X-Tenant-ID: {{tenant_id}}` em requisições que precisam do tenant

## Testes Automáticos

Cada request que cria um recurso tem um **test script** que:

1. Verifica se a resposta é 201 (Created) ou 200 (OK)
2. Extrai o ID do recurso criado
3. Salva em uma variável de coleção
4. Loga no console do Postman

Você pode ver os logs no **Postman Console** (View → Show Postman Console).

## Executar Toda a Coleção

1. Clique com botão direito na coleção
2. Selecione **Run collection**
3. Configure:
   - **Iterations**: Quantas vezes executar
   - **Delay**: Delay entre requests (recomendado: 100-500ms)
4. Clique em **Run Story Engine API**

## Troubleshooting

### Variáveis não estão sendo salvas

- Verifique se o ambiente está selecionado
- Verifique os logs no Postman Console
- Certifique-se de que a resposta está retornando o formato esperado

### Erro 400 Bad Request

- Verifique se o `X-Tenant-ID` header está presente
- Verifique se os UUIDs nas variáveis são válidos
- Execute os requests na ordem correta (criar tenant antes de story)

### Erro 404 Not Found

- Certifique-se de executar os requests na ordem correta
- Verifique se as variáveis foram preenchidas corretamente
- Use o **Complete Workflow** para garantir a ordem correta

## Exemplo de Uso Manual

1. Execute **"Create Tenant"**
   - Variável `tenant_id` é preenchida automaticamente

2. Execute **"Create Story"**
   - Usa `{{tenant_id}}` do header automaticamente
   - Variável `story_id` é preenchida automaticamente

3. Execute **"Create Chapter"**
   - Usa `{{story_id}}` do body automaticamente
   - Variável `chapter_id` é preenchida automaticamente

4. Continue com Scene e Beat seguindo o mesmo padrão

## Notas Importantes

- **X-Tenant-ID Header**: Todas as requisições de Stories precisam do header `X-Tenant-ID`
- **Ordem Importante**: Crie recursos na ordem: Tenant → Story → Chapter → Scene → Beat
- **Variáveis de Workflow**: A pasta "Complete Workflow" usa variáveis separadas (`workflow_*`) para não interferir com testes manuais

