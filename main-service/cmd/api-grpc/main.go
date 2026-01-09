package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	redisadapter "github.com/story-engine/main-service/internal/adapters/redis"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	artifactstatsapp "github.com/story-engine/main-service/internal/application/rpg/artifact_stats"
	characterinventoryapp "github.com/story-engine/main-service/internal/application/rpg/character_inventory"
	characterskillapp "github.com/story-engine/main-service/internal/application/rpg/character_skill"
	characterstatsapp "github.com/story-engine/main-service/internal/application/rpg/character_stats"
	rpgclassapp "github.com/story-engine/main-service/internal/application/rpg/rpg_class"
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	skillapp "github.com/story-engine/main-service/internal/application/rpg/skill"
	"github.com/story-engine/main-service/internal/application/story"
	beatapp "github.com/story-engine/main-service/internal/application/story/beat"
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	contentblockapp "github.com/story-engine/main-service/internal/application/story/content_block"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/application/tenant"
	worldapp "github.com/story-engine/main-service/internal/application/world"
	archetypeapp "github.com/story-engine/main-service/internal/application/world/archetype"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	factionapp "github.com/story-engine/main-service/internal/application/world/faction"
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	loreapp "github.com/story-engine/main-service/internal/application/world/lore"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/database"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	grpcserver "github.com/story-engine/main-service/internal/transport/grpc"
	"github.com/story-engine/main-service/internal/transport/grpc/handlers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	log := logger.New()

	log.Info("Starting gRPC server", "port", cfg.GRPC.Port)
	log.Info("LLM gateway notifications", "enabled", cfg.LLM.Enabled)

	var ingestionQueue queue.IngestionQueue
	if cfg.LLM.Enabled {
		redisClient := redis.NewClient(&redis.Options{
			Addr:     cfg.LLM.Redis.Addr,
			Password: cfg.LLM.Redis.Password,
			DB:       cfg.LLM.Redis.DB,
		})
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			log.Error("failed to connect to LLM Redis", "error", err)
		} else {
			ingestionQueue = redisadapter.NewIngestionQueue(redisClient)
		}
	}

	// Connect to database
	db, err := database.New(cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	pgDB := postgres.NewDB(db)

	// Initialize repositories
	tenantRepo := postgres.NewTenantRepository(pgDB)
	worldRepo := postgres.NewWorldRepository(pgDB)
	locationRepo := postgres.NewLocationRepository(pgDB)
	traitRepo := postgres.NewTraitRepository(pgDB)
	archetypeRepo := postgres.NewArchetypeRepository(pgDB)
	archetypeTraitRepo := postgres.NewArchetypeTraitRepository(pgDB)
	characterRepo := postgres.NewCharacterRepository(pgDB)
	characterTraitRepo := postgres.NewCharacterTraitRepository(pgDB)
	artifactRepo := postgres.NewArtifactRepository(pgDB)
	eventRepo := postgres.NewEventRepository(pgDB)
	factionRepo := postgres.NewFactionRepository(pgDB)
	loreRepo := postgres.NewLoreRepository(pgDB)
	storyRepo := postgres.NewStoryRepository(pgDB)
	chapterRepo := postgres.NewChapterRepository(pgDB)
	sceneRepo := postgres.NewSceneRepository(pgDB)
	beatRepo := postgres.NewBeatRepository(pgDB)
	contentBlockRepo := postgres.NewContentBlockRepository(pgDB)
	contentAnchorRepo := postgres.NewContentAnchorRepository(pgDB)
	entityRelationRepo := postgres.NewEntityRelationRepository(pgDB)
	auditLogRepo := postgres.NewAuditLogRepository(pgDB)
	transactionRepo := postgres.NewTransactionRepository(pgDB)
	rpgSystemRepo := postgres.NewRPGSystemRepository(pgDB)
	skillRepo := postgres.NewSkillRepository(pgDB)
	rpgClassRepo := postgres.NewRPGClassRepository(pgDB)
	rpgClassSkillRepo := postgres.NewRPGClassSkillRepository(pgDB)
	characterSkillRepo := postgres.NewCharacterSkillRepository(pgDB)
	characterStatsRepo := postgres.NewCharacterRPGStatsRepository(pgDB)
	artifactStatsRepo := postgres.NewArtifactRPGStatsRepository(pgDB)
	inventoryRepo := postgres.NewCharacterInventoryRepository(pgDB)
	inventoryItemRepo := postgres.NewInventoryItemRepository(pgDB)

	// Entity relation use cases (created early because they're used by other use cases)
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, ingestionQueue, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	listRelationsByWorldUseCase := relationapp.NewListRelationsByWorldUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, ingestionQueue, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)

	// Initialize use cases
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := worldapp.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := worldapp.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := worldapp.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := worldapp.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := worldapp.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
	createLocationUseCase := locationapp.NewCreateLocationUseCase(locationRepo, worldRepo, auditLogRepo, log)
	getLocationUseCase := locationapp.NewGetLocationUseCase(locationRepo, log)
	listLocationsUseCase := locationapp.NewListLocationsUseCase(locationRepo, log)
	updateLocationUseCase := locationapp.NewUpdateLocationUseCase(locationRepo, auditLogRepo, log)
	deleteLocationUseCase := locationapp.NewDeleteLocationUseCase(locationRepo, entityRelationRepo, auditLogRepo, log)
	getChildrenUseCase := locationapp.NewGetChildrenUseCase(locationRepo, log)
	getAncestorsUseCase := locationapp.NewGetAncestorsUseCase(locationRepo, log)
	getDescendantsUseCase := locationapp.NewGetDescendantsUseCase(locationRepo, log)
	moveLocationUseCase := locationapp.NewMoveLocationUseCase(locationRepo, auditLogRepo, log)
	createTraitUseCase := traitapp.NewCreateTraitUseCase(traitRepo, tenantRepo, auditLogRepo, log)
	getTraitUseCase := traitapp.NewGetTraitUseCase(traitRepo, log)
	listTraitsUseCase := traitapp.NewListTraitsUseCase(traitRepo, log)
	updateTraitUseCase := traitapp.NewUpdateTraitUseCase(traitRepo, auditLogRepo, log)
	deleteTraitUseCase := traitapp.NewDeleteTraitUseCase(traitRepo, auditLogRepo, log)
	createArchetypeUseCase := archetypeapp.NewCreateArchetypeUseCase(archetypeRepo, tenantRepo, auditLogRepo, log)
	getArchetypeUseCase := archetypeapp.NewGetArchetypeUseCase(archetypeRepo, log)
	listArchetypesUseCase := archetypeapp.NewListArchetypesUseCase(archetypeRepo, log)
	updateArchetypeUseCase := archetypeapp.NewUpdateArchetypeUseCase(archetypeRepo, auditLogRepo, log)
	deleteArchetypeUseCase := archetypeapp.NewDeleteArchetypeUseCase(archetypeRepo, archetypeTraitRepo, auditLogRepo, log)
	addTraitToArchetypeUseCase := archetypeapp.NewAddTraitToArchetypeUseCase(archetypeRepo, traitRepo, archetypeTraitRepo, log)
	removeTraitFromArchetypeUseCase := archetypeapp.NewRemoveTraitFromArchetypeUseCase(archetypeTraitRepo, log)
	getArchetypeTraitsUseCase := archetypeapp.NewGetArchetypeTraitsUseCase(archetypeTraitRepo, log)
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	getCharacterTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	getCharacterEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	addTraitToCharacterUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	updateCharacterTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	removeTraitFromCharacterUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	createArtifactUseCase := artifactapp.NewCreateArtifactUseCase(artifactRepo, createRelationUseCase, worldRepo, characterRepo, locationRepo, auditLogRepo, log)
	getArtifactUseCase := artifactapp.NewGetArtifactUseCase(artifactRepo, log)
	listArtifactsUseCase := artifactapp.NewListArtifactsUseCase(artifactRepo, log)
	updateArtifactUseCase := artifactapp.NewUpdateArtifactUseCase(artifactRepo, createRelationUseCase, listRelationsBySourceUseCase, deleteRelationUseCase, characterRepo, locationRepo, worldRepo, auditLogRepo, log)
	deleteArtifactUseCase := artifactapp.NewDeleteArtifactUseCase(artifactRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	getArtifactReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(listRelationsBySourceUseCase, log)
	addArtifactReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, entityRelationRepo, createRelationUseCase, characterRepo, locationRepo, log)
	removeArtifactReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	createEventUseCase := eventapp.NewCreateEventUseCase(eventRepo, worldRepo, auditLogRepo, log)
	getEventUseCase := eventapp.NewGetEventUseCase(eventRepo, log)
	listEventsUseCase := eventapp.NewListEventsUseCase(eventRepo, log)
	updateEventUseCase := eventapp.NewUpdateEventUseCase(eventRepo, auditLogRepo, log)
	deleteEventUseCase := eventapp.NewDeleteEventUseCase(eventRepo, entityRelationRepo, auditLogRepo, log)
	addEventReferenceUseCase := eventapp.NewAddReferenceUseCase(eventRepo, entityRelationRepo, createRelationUseCase, characterRepo, locationRepo, artifactRepo, factionRepo, loreRepo, log)
	removeEventReferenceUseCase := eventapp.NewRemoveReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getEventReferencesUseCase := eventapp.NewGetReferencesUseCase(listRelationsBySourceUseCase, log)
	updateEventReferenceUseCase := eventapp.NewUpdateReferenceUseCase(getRelationUseCase, updateRelationUseCase, log)
	getEventChildrenUseCase := eventapp.NewGetChildrenUseCase(eventRepo, log)
	getEventAncestorsUseCase := eventapp.NewGetAncestorsUseCase(eventRepo, log)
	getEventDescendantsUseCase := eventapp.NewGetDescendantsUseCase(eventRepo, log)
	moveEventUseCase := eventapp.NewMoveEventUseCase(eventRepo, log)
	setEventEpochUseCase := eventapp.NewSetEpochUseCase(eventRepo, log)
	getEventEpochUseCase := eventapp.NewGetEpochUseCase(eventRepo, log)
	getTimelineUseCase := eventapp.NewGetTimelineUseCase(eventRepo, log)
	createFactionUseCase := factionapp.NewCreateFactionUseCase(factionRepo, worldRepo, auditLogRepo, log)
	getFactionUseCase := factionapp.NewGetFactionUseCase(factionRepo, log)
	listFactionsUseCase := factionapp.NewListFactionsUseCase(factionRepo, log)
	updateFactionUseCase := factionapp.NewUpdateFactionUseCase(factionRepo, auditLogRepo, log)
	deleteFactionUseCase := factionapp.NewDeleteFactionUseCase(factionRepo, entityRelationRepo, auditLogRepo, log)
	getFactionChildrenUseCase := factionapp.NewGetChildrenUseCase(factionRepo, log)
	addFactionReferenceUseCase := factionapp.NewAddReferenceUseCase(factionRepo, entityRelationRepo, createRelationUseCase, characterRepo, locationRepo, artifactRepo, eventRepo, loreRepo, log)
	removeFactionReferenceUseCase := factionapp.NewRemoveReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getFactionReferencesUseCase := factionapp.NewGetReferencesUseCase(listRelationsBySourceUseCase, log)
	updateFactionReferenceUseCase := factionapp.NewUpdateReferenceUseCase(getRelationUseCase, updateRelationUseCase, log)
	createLoreUseCase := loreapp.NewCreateLoreUseCase(loreRepo, worldRepo, auditLogRepo, log)
	getLoreUseCase := loreapp.NewGetLoreUseCase(loreRepo, log)
	listLoresUseCase := loreapp.NewListLoresUseCase(loreRepo, log)
	updateLoreUseCase := loreapp.NewUpdateLoreUseCase(loreRepo, auditLogRepo, log)
	deleteLoreUseCase := loreapp.NewDeleteLoreUseCase(loreRepo, entityRelationRepo, auditLogRepo, log)
	getLoreChildrenUseCase := loreapp.NewGetChildrenUseCase(loreRepo, log)
	addLoreReferenceUseCase := loreapp.NewAddReferenceUseCase(loreRepo, entityRelationRepo, createRelationUseCase, characterRepo, locationRepo, artifactRepo, eventRepo, factionRepo, log)
	removeLoreReferenceUseCase := loreapp.NewRemoveReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getLoreReferencesUseCase := loreapp.NewGetReferencesUseCase(listRelationsBySourceUseCase, log)
	updateLoreReferenceUseCase := loreapp.NewUpdateReferenceUseCase(getRelationUseCase, updateRelationUseCase, log)
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, ingestionQueue, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, ingestionQueue, log)
	listStoriesUseCase := story.NewListStoriesUseCase(storyRepo, log)
	cloneStoryUseCase := story.NewCloneStoryUseCase(
		storyRepo,
		chapterRepo,
		sceneRepo,
		beatRepo,
		contentBlockRepo,
		auditLogRepo,
		transactionRepo,
		log,
	)
	versionGraphUseCase := story.NewGetStoryVersionGraphUseCase(storyRepo, log)
	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, ingestionQueue, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, ingestionQueue, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, ingestionQueue, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, ingestionQueue, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, entityRelationRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addSceneReferenceUseCase := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, storyRepo, createRelationUseCase, listRelationsBySourceUseCase, characterRepo, locationRepo, artifactRepo, log)
	removeSceneReferenceUseCase := sceneapp.NewRemoveSceneReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getSceneReferencesUseCase := sceneapp.NewGetSceneReferencesUseCase(listRelationsBySourceUseCase, log)
	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, ingestionQueue, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, ingestionQueue, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)
	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, ingestionQueue, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, ingestionQueue, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	createContentAnchorUseCase := contentblockapp.NewCreateContentAnchorUseCase(contentAnchorRepo, contentBlockRepo, log)
	listContentAnchorsByContentBlockUseCase := contentblockapp.NewListContentAnchorsByContentBlockUseCase(contentAnchorRepo, contentBlockRepo, log)
	listContentBlocksByEntityUseCase := contentblockapp.NewListContentBlocksByEntityUseCase(contentAnchorRepo, contentBlockRepo, log)
	deleteContentAnchorUseCase := contentblockapp.NewDeleteContentAnchorUseCase(contentAnchorRepo, log)

	createRPGSystemUseCase := rpgsystemapp.NewCreateRPGSystemUseCase(rpgSystemRepo, tenantRepo, log)
	getRPGSystemUseCase := rpgsystemapp.NewGetRPGSystemUseCase(rpgSystemRepo, log)
	listRPGSystemsUseCase := rpgsystemapp.NewListRPGSystemsUseCase(rpgSystemRepo, log)
	updateRPGSystemUseCase := rpgsystemapp.NewUpdateRPGSystemUseCase(rpgSystemRepo, log)
	deleteRPGSystemUseCase := rpgsystemapp.NewDeleteRPGSystemUseCase(rpgSystemRepo, log)
	createSkillUseCase := skillapp.NewCreateSkillUseCase(skillRepo, rpgSystemRepo, log)
	getSkillUseCase := skillapp.NewGetSkillUseCase(skillRepo, log)
	listSkillsUseCase := skillapp.NewListSkillsUseCase(skillRepo, log)
	updateSkillUseCase := skillapp.NewUpdateSkillUseCase(skillRepo, log)
	deleteSkillUseCase := skillapp.NewDeleteSkillUseCase(skillRepo, log)
	createRPGClassUseCase := rpgclassapp.NewCreateRPGClassUseCase(rpgClassRepo, rpgSystemRepo, log)
	getRPGClassUseCase := rpgclassapp.NewGetRPGClassUseCase(rpgClassRepo, log)
	listRPGClassesUseCase := rpgclassapp.NewListRPGClassesUseCase(rpgClassRepo, log)
	updateRPGClassUseCase := rpgclassapp.NewUpdateRPGClassUseCase(rpgClassRepo, log)
	deleteRPGClassUseCase := rpgclassapp.NewDeleteRPGClassUseCase(rpgClassRepo, log)
	addSkillToClassUseCase := rpgclassapp.NewAddSkillToClassUseCase(rpgClassSkillRepo, rpgClassRepo, skillRepo, log)
	listClassSkillsUseCase := rpgclassapp.NewListClassSkillsUseCase(rpgClassSkillRepo, log)
	removeSkillFromClassUseCase := rpgclassapp.NewRemoveSkillFromClassUseCase(rpgClassSkillRepo, rpgClassRepo, log)
	learnSkillUseCase := characterskillapp.NewLearnSkillUseCase(characterSkillRepo, characterRepo, skillRepo, log)
	updateCharacterSkillUseCase := characterskillapp.NewUpdateCharacterSkillUseCase(characterSkillRepo, skillRepo, log)
	deleteCharacterSkillUseCase := characterskillapp.NewDeleteCharacterSkillUseCase(characterSkillRepo, log)
	listCharacterSkillsUseCase := characterskillapp.NewListCharacterSkillsUseCase(characterSkillRepo, log)
	createCharacterStatsUseCase := characterstatsapp.NewCreateCharacterStatsUseCase(characterStatsRepo, characterRepo, eventRepo, log)
	getActiveCharacterStatsUseCase := characterstatsapp.NewGetActiveCharacterStatsUseCase(characterStatsRepo, log)
	listCharacterStatsHistoryUseCase := characterstatsapp.NewListCharacterStatsHistoryUseCase(characterStatsRepo, log)
	activateCharacterStatsVersionUseCase := characterstatsapp.NewActivateCharacterStatsVersionUseCase(characterStatsRepo, log)
	deleteAllCharacterStatsUseCase := characterstatsapp.NewDeleteAllCharacterStatsUseCase(characterStatsRepo, log)
	createArtifactStatsUseCase := artifactstatsapp.NewCreateArtifactStatsUseCase(artifactStatsRepo, artifactRepo, eventRepo, log)
	getActiveArtifactStatsUseCase := artifactstatsapp.NewGetActiveArtifactStatsUseCase(artifactStatsRepo, log)
	listArtifactStatsHistoryUseCase := artifactstatsapp.NewListArtifactStatsHistoryUseCase(artifactStatsRepo, log)
	activateArtifactStatsVersionUseCase := artifactstatsapp.NewActivateArtifactStatsVersionUseCase(artifactStatsRepo, log)
	addItemToInventoryUseCase := characterinventoryapp.NewAddItemToInventoryUseCase(inventoryRepo, characterRepo, inventoryItemRepo, log)
	updateInventoryItemUseCase := characterinventoryapp.NewUpdateCharacterInventoryUseCase(inventoryRepo, log)
	deleteInventoryItemUseCase := characterinventoryapp.NewDeleteCharacterInventoryUseCase(inventoryRepo, log)
	listInventoryUseCase := characterinventoryapp.NewListCharacterInventoryUseCase(inventoryRepo, log)

	createWorldUseCase.SetIngestionQueue(ingestionQueue)
	updateWorldUseCase.SetIngestionQueue(ingestionQueue)
	createLocationUseCase.SetIngestionQueue(ingestionQueue)
	updateLocationUseCase.SetIngestionQueue(ingestionQueue)
	createCharacterUseCase.SetIngestionQueue(ingestionQueue)
	updateCharacterUseCase.SetIngestionQueue(ingestionQueue)
	createArtifactUseCase.SetIngestionQueue(ingestionQueue)
	updateArtifactUseCase.SetIngestionQueue(ingestionQueue)
	createEventUseCase.SetIngestionQueue(ingestionQueue)
	updateEventUseCase.SetIngestionQueue(ingestionQueue)
	createFactionUseCase.SetIngestionQueue(ingestionQueue)
	updateFactionUseCase.SetIngestionQueue(ingestionQueue)
	createLoreUseCase.SetIngestionQueue(ingestionQueue)
	updateLoreUseCase.SetIngestionQueue(ingestionQueue)

	// Create handlers
	tenantHandler := handlers.NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := handlers.NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	locationHandler := handlers.NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)
	// TODO: Update CharacterHandler to use entity_relations instead of character_relationship use cases
	// For now, character relationship methods will return "not implemented" errors
	characterHandler := handlers.NewCharacterHandler(
		createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase,
		getCharacterTraitsUseCase, getCharacterEventsUseCase,
		addTraitToCharacterUseCase, updateCharacterTraitUseCase, removeTraitFromCharacterUseCase,
		nil, nil, nil, nil, nil, // Character relationship use cases - TODO: implement with entity_relations
		log)
	artifactHandler := handlers.NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)
	eventHandler := handlers.NewEventHandler(createEventUseCase, getEventUseCase, listEventsUseCase, updateEventUseCase, deleteEventUseCase, addEventReferenceUseCase, removeEventReferenceUseCase, getEventReferencesUseCase, updateEventReferenceUseCase, getEventChildrenUseCase, getEventAncestorsUseCase, getEventDescendantsUseCase, moveEventUseCase, setEventEpochUseCase, getEventEpochUseCase, getTimelineUseCase, log)
	factionHandler := handlers.NewFactionHandler(createFactionUseCase, getFactionUseCase, listFactionsUseCase, updateFactionUseCase, deleteFactionUseCase, getFactionChildrenUseCase, addFactionReferenceUseCase, removeFactionReferenceUseCase, getFactionReferencesUseCase, updateFactionReferenceUseCase, log)
	loreHandler := handlers.NewLoreHandler(createLoreUseCase, getLoreUseCase, listLoresUseCase, updateLoreUseCase, deleteLoreUseCase, getLoreChildrenUseCase, addLoreReferenceUseCase, removeLoreReferenceUseCase, getLoreReferencesUseCase, updateLoreReferenceUseCase, log)
	traitHandler := handlers.NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	archetypeHandler := handlers.NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitToArchetypeUseCase, removeTraitFromArchetypeUseCase, getArchetypeTraitsUseCase, log)
	storyHandler := handlers.NewStoryHandler(
		createStoryUseCase,
		getStoryUseCase,
		updateStoryUseCase,
		listStoriesUseCase,
		cloneStoryUseCase,
		versionGraphUseCase,
		log,
	)
	chapterHandler := handlers.NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)
	sceneHandler := handlers.NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addSceneReferenceUseCase, removeSceneReferenceUseCase, getSceneReferencesUseCase, log)
	beatHandler := handlers.NewBeatHandler(createBeatUseCase, getBeatUseCase, updateBeatUseCase, deleteBeatUseCase, listBeatsUseCase, moveBeatUseCase, log)
	contentBlockHandler := handlers.NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)
	contentAnchorHandler := handlers.NewContentAnchorHandler(createContentAnchorUseCase, listContentAnchorsByContentBlockUseCase, listContentBlocksByEntityUseCase, deleteContentAnchorUseCase, log)
	entityRelationHandler := handlers.NewEntityRelationHandler(createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, listRelationsByWorldUseCase, updateRelationUseCase, deleteRelationUseCase, log)
	rpgSystemHandler := handlers.NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)
	skillHandler := handlers.NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)
	rpgClassHandler := handlers.NewRPGClassHandler(createRPGClassUseCase, getRPGClassUseCase, listRPGClassesUseCase, updateRPGClassUseCase, deleteRPGClassUseCase, addSkillToClassUseCase, listClassSkillsUseCase, removeSkillFromClassUseCase, log)
	characterSkillHandler := handlers.NewCharacterSkillHandler(learnSkillUseCase, updateCharacterSkillUseCase, deleteCharacterSkillUseCase, listCharacterSkillsUseCase, log)
	characterRPGStatsHandler := handlers.NewCharacterRPGStatsHandler(createCharacterStatsUseCase, getActiveCharacterStatsUseCase, listCharacterStatsHistoryUseCase, activateCharacterStatsVersionUseCase, deleteAllCharacterStatsUseCase, log)
	artifactRPGStatsHandler := handlers.NewArtifactRPGStatsHandler(createArtifactStatsUseCase, getActiveArtifactStatsUseCase, listArtifactStatsHistoryUseCase, activateArtifactStatsVersionUseCase, log)
	inventoryHandler := handlers.NewInventoryHandler(addItemToInventoryUseCase, updateInventoryItemUseCase, deleteInventoryItemUseCase, listInventoryUseCase, log)

	// Create and configure gRPC server
	grpcServer := grpcserver.NewServer(cfg, log)
	grpcServer.RegisterTenantService(tenantHandler)
	grpcServer.RegisterWorldService(worldHandler)
	grpcServer.RegisterLocationService(locationHandler)
	grpcServer.RegisterCharacterService(characterHandler)
	grpcServer.RegisterArtifactService(artifactHandler)
	grpcServer.RegisterEventService(eventHandler)
	grpcServer.RegisterFactionService(factionHandler)
	grpcServer.RegisterLoreService(loreHandler)
	grpcServer.RegisterTraitService(traitHandler)
	grpcServer.RegisterArchetypeService(archetypeHandler)
	grpcServer.RegisterStoryService(storyHandler)
	grpcServer.RegisterChapterService(chapterHandler)
	grpcServer.RegisterSceneService(sceneHandler)
	grpcServer.RegisterBeatService(beatHandler)
	grpcServer.RegisterContentBlockService(contentBlockHandler)
	grpcServer.RegisterContentAnchorService(contentAnchorHandler)
	grpcServer.RegisterEntityRelationService(entityRelationHandler)
	grpcServer.RegisterRPGSystemService(rpgSystemHandler)
	grpcServer.RegisterSkillService(skillHandler)
	grpcServer.RegisterRPGClassService(rpgClassHandler)
	grpcServer.RegisterCharacterSkillService(characterSkillHandler)
	grpcServer.RegisterCharacterRPGStatsService(characterRPGStatsHandler)
	grpcServer.RegisterArtifactRPGStatsService(artifactRPGStatsHandler)
	grpcServer.RegisterInventoryService(inventoryHandler)

	// Setup graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := grpcServer.Start(cfg.GRPC.Port); err != nil {
			errChan <- err
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Error("server error", "error", err)
		os.Exit(1)
	case sig := <-sigChan:
		log.Info("received signal", "signal", sig)
		cancel()
		grpcServer.Stop()
		log.Info("gRPC server stopped")
	}
}
