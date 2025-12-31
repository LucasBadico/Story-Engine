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

- `tenant_id` - ID do tenant criado
- `story_id` - ID da história criada
- `chapter_id` - ID do capítulo criado
- `scene_id` - ID da cena criada
- `beat_id` - ID do beat criado
- `cloned_story_id` - ID da história clonada
- `story_version` - Número da versão após clone

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
│   └── Delete Scene
├── Beats
│   ├── Create Beat (salva beat_id)
│   ├── Get Beat
│   ├── List Beats
│   ├── Update Beat
│   └── Delete Beat
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

