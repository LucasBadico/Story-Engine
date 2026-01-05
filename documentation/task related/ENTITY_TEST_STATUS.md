# Entity Test Status

## Como rodar testes por entidade

```bash
cd /Users/badico/Repositories/story-engine/main-service
go test -v -tags integration ./internal/transport/grpc/handlers/... -run "TestNOME"
```

Exemplo:
```bash
go test -v -tags integration ./internal/transport/grpc/handlers/... -run "TestTenantHandler"
```

---

## Lista de Entidades (gRPC Handlers)

| # | Entidade | Pattern para -run | Status |
|---|----------|-------------------|--------|
| 1 | **Tenant** | `TestTenantHandler` | ✅ ok|
| 2 | **World** | `TestWorldHandler` |  ❌ Falhando |
| 3 | **Story** | `TestStoryHandler` | ❌ Falhando |
| 4 | **Chapter** | `TestChapterHandler` | ❌ Falhando |
| 5 | **Scene** | `TestSceneHandler` | ⏳ Pendente |
| 6 | **Beat** | `TestBeatHandler` | ⏳ Pendente |
| 7 | **ProseBlock** | `TestProseBlockHandler` | ⏳ Pendente |
| 8 | **ImageBlock** | `TestImageBlockHandler` | ⏳ Pendente |
| 9 | **Character** | `TestCharacterHandler` | ⏳ Pendente |
| 10 | **CharacterSkill** | `TestCharacterSkillHandler` | ⏳ Pendente |
| 11 | **CharacterRPGStats** | `TestCharacterRPGStatsHandler` | ⏳ Pendente |
| 12 | **Inventory** | `TestInventoryHandler` | ⏳ Pendente |
| 13 | **Archetype** | `TestArchetypeHandler` | ⏳ Pendente |
| 14 | **Trait** | `TestTraitHandler` | ⏳ Pendente |
| 15 | **Artifact** | `TestArtifactHandler` | ⏳ Pendente |
| 16 | **ArtifactRPGStats** | `TestArtifactRPGStatsHandler` | ⏳ Pendente |
| 17 | **Location** | `TestLocationHandler` | ⏳ Pendente |
| 18 | **Event** | `TestEventHandler` | ⏳ Pendente |
| 19 | **RPGSystem** | `TestRPGSystemHandler` | ⏳ Pendente |
| 20 | **RPGClass** | `TestRPGClassHandler` | ⏳ Pendente |
| 21 | **Skill** | `TestSkillHandler` | ⏳ Pendente |

---

## Hierarquia de Dependências

```
Tenant (raiz - sem dependências)
├── World (depende de Tenant)
│   ├── Character (depende de World)
│   │   ├── CharacterSkill (depende de Character + Skill)
│   │   ├── CharacterRPGStats (depende de Character + RPGSystem)
│   │   └── Inventory (depende de Character)
│   ├── Archetype (depende de World)
│   │   └── Trait (depende de World, usado por Archetype)
│   ├── Artifact (depende de World)
│   │   └── ArtifactRPGStats (depende de Artifact + RPGSystem)
│   ├── Location (depende de World)
│   └── Event (depende de World)
├── Story (depende de Tenant, pode ter World)
│   ├── Chapter (depende de Story)
│   │   ├── Scene (depende de Chapter)
│   │   │   └── Beat (depende de Scene)
│   │   ├── ProseBlock (depende de Chapter)
│   │   └── ImageBlock (depende de Chapter)
└── RPGSystem (depende de Tenant)
    ├── RPGClass (depende de RPGSystem)
    └── Skill (depende de RPGSystem)
```

---

## Ordem Sugerida de Teste

Baseado nas dependências, testar nesta ordem:

1. `TestTenantHandler` - Base de tudo
2. `TestWorldHandler` - Segundo nível
3. `TestRPGSystemHandler` - Independente de World
4. `TestSkillHandler` - Depende de RPGSystem
5. `TestRPGClassHandler` - Depende de RPGSystem + Skill
6. `TestStoryHandler` - Depende de Tenant
7. `TestChapterHandler` - Depende de Story
8. `TestSceneHandler` - Depende de Chapter
9. `TestBeatHandler` - Depende de Scene
10. `TestProseBlockHandler` - Depende de Chapter
11. `TestImageBlockHandler` - Depende de Chapter
12. `TestTraitHandler` - Depende de World
13. `TestArchetypeHandler` - Depende de World + Trait
14. `TestCharacterHandler` - Depende de World
15. `TestCharacterSkillHandler` - Depende de Character + Skill
16. `TestCharacterRPGStatsHandler` - Depende de Character + RPGSystem
17. `TestInventoryHandler` - Depende de Character
18. `TestArtifactHandler` - Depende de World
19. `TestArtifactRPGStatsHandler` - Depende de Artifact + RPGSystem
20. `TestLocationHandler` - Depende de World
21. `TestEventHandler` - Depende de World

---

## Log de Progresso

| Data | Entidade | Resultado | Notas |
|------|----------|-----------|-------|
| | | | |

