package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	redisadapter "github.com/story-engine/main-service/internal/adapters/redis"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	artifactstatsapp "github.com/story-engine/main-service/internal/application/rpg/artifact_stats"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	characterinventoryapp "github.com/story-engine/main-service/internal/application/rpg/character_inventory"
	characterskillapp "github.com/story-engine/main-service/internal/application/rpg/character_skill"
	characterstatsapp "github.com/story-engine/main-service/internal/application/rpg/character_stats"
	rpgeventapp "github.com/story-engine/main-service/internal/application/rpg/event"
	inventoryitemapp "github.com/story-engine/main-service/internal/application/rpg/inventory_item"
	inventoryslotapp "github.com/story-engine/main-service/internal/application/rpg/inventory_slot"
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
	httphandlers "github.com/story-engine/main-service/internal/transport/http/handlers"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	log := logger.New()

	log.Info("Starting HTTP server", "port", cfg.HTTP.Port)
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
	rpgSystemRepo := postgres.NewRPGSystemRepository(pgDB)
	characterRPGStatsRepo := postgres.NewCharacterRPGStatsRepository(pgDB)
	artifactRPGStatsRepo := postgres.NewArtifactRPGStatsRepository(pgDB)
	skillRepo := postgres.NewSkillRepository(pgDB)
	characterSkillRepo := postgres.NewCharacterSkillRepository(pgDB)
	rpgClassRepo := postgres.NewRPGClassRepository(pgDB)
	rpgClassSkillRepo := postgres.NewRPGClassSkillRepository(pgDB)
	inventorySlotRepo := postgres.NewInventorySlotRepository(pgDB)
	inventoryItemRepo := postgres.NewInventoryItemRepository(pgDB)
	characterInventoryRepo := postgres.NewCharacterInventoryRepository(pgDB)
	storyRepo := postgres.NewStoryRepository(pgDB)
	chapterRepo := postgres.NewChapterRepository(pgDB)
	sceneRepo := postgres.NewSceneRepository(pgDB)
	beatRepo := postgres.NewBeatRepository(pgDB)
	contentBlockRepo := postgres.NewContentBlockRepository(pgDB)
	contentAnchorRepo := postgres.NewContentAnchorRepository(pgDB)
	entityRelationRepo := postgres.NewEntityRelationRepository(pgDB)
	auditLogRepo := postgres.NewAuditLogRepository(pgDB)
	transactionRepo := postgres.NewTransactionRepository(pgDB)

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
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	addTraitToCharacterUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitFromCharacterUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateCharacterTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getCharacterTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	getCharacterEventsUseCase := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
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
	createCharacterStatsUseCase := characterstatsapp.NewCreateCharacterStatsUseCase(characterRPGStatsRepo, characterRepo, eventRepo, log)
	getActiveCharacterStatsUseCase := characterstatsapp.NewGetActiveCharacterStatsUseCase(characterRPGStatsRepo, log)
	listCharacterStatsHistoryUseCase := characterstatsapp.NewListCharacterStatsHistoryUseCase(characterRPGStatsRepo, log)
	activateCharacterStatsVersionUseCase := characterstatsapp.NewActivateCharacterStatsVersionUseCase(characterRPGStatsRepo, log)
	deleteAllCharacterStatsUseCase := characterstatsapp.NewDeleteAllCharacterStatsUseCase(characterRPGStatsRepo, log)
	createArtifactStatsUseCase := artifactstatsapp.NewCreateArtifactStatsUseCase(artifactRPGStatsRepo, artifactRepo, eventRepo, log)
	getActiveArtifactStatsUseCase := artifactstatsapp.NewGetActiveArtifactStatsUseCase(artifactRPGStatsRepo, log)
	listArtifactStatsHistoryUseCase := artifactstatsapp.NewListArtifactStatsHistoryUseCase(artifactRPGStatsRepo, log)
	activateArtifactStatsVersionUseCase := artifactstatsapp.NewActivateArtifactStatsVersionUseCase(artifactRPGStatsRepo, log)
	getEventStatChangesUseCase := rpgeventapp.NewGetEventStatChangesUseCase(characterRPGStatsRepo, artifactRPGStatsRepo, log)
	createSkillUseCase := skillapp.NewCreateSkillUseCase(skillRepo, rpgSystemRepo, log)
	getSkillUseCase := skillapp.NewGetSkillUseCase(skillRepo, log)
	listSkillsUseCase := skillapp.NewListSkillsUseCase(skillRepo, log)
	updateSkillUseCase := skillapp.NewUpdateSkillUseCase(skillRepo, log)
	deleteSkillUseCase := skillapp.NewDeleteSkillUseCase(skillRepo, log)
	learnSkillUseCase := characterskillapp.NewLearnSkillUseCase(characterSkillRepo, characterRepo, skillRepo, log)
	listCharacterSkillsUseCase := characterskillapp.NewListCharacterSkillsUseCase(characterSkillRepo, log)
	updateCharacterSkillUseCase := characterskillapp.NewUpdateCharacterSkillUseCase(characterSkillRepo, skillRepo, log)
	deleteCharacterSkillUseCase := characterskillapp.NewDeleteCharacterSkillUseCase(characterSkillRepo, log)
	createRPGClassUseCase := rpgclassapp.NewCreateRPGClassUseCase(rpgClassRepo, rpgSystemRepo, log)
	getRPGClassUseCase := rpgclassapp.NewGetRPGClassUseCase(rpgClassRepo, log)
	listRPGClassesUseCase := rpgclassapp.NewListRPGClassesUseCase(rpgClassRepo, log)
	updateRPGClassUseCase := rpgclassapp.NewUpdateRPGClassUseCase(rpgClassRepo, log)
	deleteRPGClassUseCase := rpgclassapp.NewDeleteRPGClassUseCase(rpgClassRepo, log)
	addSkillToClassUseCase := rpgclassapp.NewAddSkillToClassUseCase(rpgClassSkillRepo, rpgClassRepo, skillRepo, log)
	listClassSkillsUseCase := rpgclassapp.NewListClassSkillsUseCase(rpgClassSkillRepo, log)
	changeCharacterClassUseCase := rpgcharacterapp.NewChangeCharacterClassUseCase(characterRepo, rpgClassRepo, log)
	getAvailableClassesUseCase := rpgcharacterapp.NewGetAvailableClassesUseCase(characterRepo, rpgClassRepo, rpgSystemRepo, log)
	createInventorySlotUseCase := inventoryslotapp.NewCreateInventorySlotUseCase(inventorySlotRepo, rpgSystemRepo, log)
	listInventorySlotsUseCase := inventoryslotapp.NewListInventorySlotsUseCase(inventorySlotRepo, log)
	createInventoryItemUseCase := inventoryitemapp.NewCreateInventoryItemUseCase(inventoryItemRepo, rpgSystemRepo, artifactRepo, log)
	getInventoryItemUseCase := inventoryitemapp.NewGetInventoryItemUseCase(inventoryItemRepo, log)
	listInventoryItemsUseCase := inventoryitemapp.NewListInventoryItemsUseCase(inventoryItemRepo, log)
	updateInventoryItemUseCase := inventoryitemapp.NewUpdateInventoryItemUseCase(inventoryItemRepo, log)
	deleteInventoryItemUseCase := inventoryitemapp.NewDeleteInventoryItemUseCase(inventoryItemRepo, log)
	addItemToInventoryUseCase := characterinventoryapp.NewAddItemToInventoryUseCase(characterInventoryRepo, characterRepo, inventoryItemRepo, log)
	listCharacterInventoryUseCase := characterinventoryapp.NewListCharacterInventoryUseCase(characterInventoryRepo, log)
	updateCharacterInventoryUseCase := characterinventoryapp.NewUpdateCharacterInventoryUseCase(characterInventoryRepo, log)
	equipItemUseCase := characterinventoryapp.NewEquipItemUseCase(characterInventoryRepo, log)
	unequipItemUseCase := characterinventoryapp.NewUnequipItemUseCase(characterInventoryRepo, log)
	deleteCharacterInventoryUseCase := characterinventoryapp.NewDeleteCharacterInventoryUseCase(characterInventoryRepo, log)
	transferItemUseCase := characterinventoryapp.NewTransferItemUseCase(characterInventoryRepo, characterRepo, log)

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
	tenantHandler := httphandlers.NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := httphandlers.NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	locationHandler := httphandlers.NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)
	characterHandler := httphandlers.NewCharacterHandler(
		createCharacterUseCase,
		getCharacterUseCase,
		listCharactersUseCase,
		updateCharacterUseCase,
		deleteCharacterUseCase,
		addTraitToCharacterUseCase,
		removeTraitFromCharacterUseCase,
		updateCharacterTraitUseCase,
		getCharacterTraitsUseCase,
		getCharacterEventsUseCase,
		createRelationUseCase,
		getRelationUseCase,
		listRelationsBySourceUseCase,
		listRelationsByTargetUseCase,
		updateRelationUseCase,
		deleteRelationUseCase,
		changeCharacterClassUseCase,
		getAvailableClassesUseCase,
		log,
	)
	rpgClassHandler := httphandlers.NewRPGClassHandler(createRPGClassUseCase, getRPGClassUseCase, listRPGClassesUseCase, updateRPGClassUseCase, deleteRPGClassUseCase, addSkillToClassUseCase, listClassSkillsUseCase, log)
	inventoryHandler := httphandlers.NewInventoryHandler(createInventorySlotUseCase, listInventorySlotsUseCase, createInventoryItemUseCase, getInventoryItemUseCase, listInventoryItemsUseCase, updateInventoryItemUseCase, deleteInventoryItemUseCase, addItemToInventoryUseCase, listCharacterInventoryUseCase, updateCharacterInventoryUseCase, equipItemUseCase, unequipItemUseCase, deleteCharacterInventoryUseCase, transferItemUseCase, log)
	artifactHandler := httphandlers.NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)
	eventHandler := httphandlers.NewEventHandler(createEventUseCase, getEventUseCase, listEventsUseCase, updateEventUseCase, deleteEventUseCase, addEventReferenceUseCase, removeEventReferenceUseCase, getEventReferencesUseCase, updateEventReferenceUseCase, getEventChildrenUseCase, getEventAncestorsUseCase, getEventDescendantsUseCase, moveEventUseCase, setEventEpochUseCase, getEventEpochUseCase, getTimelineUseCase, getEventStatChangesUseCase, log)
	factionHandler := httphandlers.NewFactionHandler(createFactionUseCase, getFactionUseCase, listFactionsUseCase, updateFactionUseCase, deleteFactionUseCase, getFactionChildrenUseCase, addFactionReferenceUseCase, removeFactionReferenceUseCase, getFactionReferencesUseCase, updateFactionReferenceUseCase, log)
	loreHandler := httphandlers.NewLoreHandler(createLoreUseCase, getLoreUseCase, listLoresUseCase, updateLoreUseCase, deleteLoreUseCase, getLoreChildrenUseCase, addLoreReferenceUseCase, removeLoreReferenceUseCase, getLoreReferencesUseCase, updateLoreReferenceUseCase, log)
	rpgSystemHandler := httphandlers.NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)
	characterRPGStatsHandler := httphandlers.NewCharacterRPGStatsHandler(createCharacterStatsUseCase, getActiveCharacterStatsUseCase, listCharacterStatsHistoryUseCase, activateCharacterStatsVersionUseCase, deleteAllCharacterStatsUseCase, log)
	artifactRPGStatsHandler := httphandlers.NewArtifactRPGStatsHandler(createArtifactStatsUseCase, getActiveArtifactStatsUseCase, listArtifactStatsHistoryUseCase, activateArtifactStatsVersionUseCase, log)
	skillHandler := httphandlers.NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)
	characterSkillHandler := httphandlers.NewCharacterSkillHandler(learnSkillUseCase, listCharacterSkillsUseCase, updateCharacterSkillUseCase, deleteCharacterSkillUseCase, log)
	traitHandler := httphandlers.NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	archetypeHandler := httphandlers.NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitToArchetypeUseCase, removeTraitFromArchetypeUseCase, getArchetypeTraitsUseCase, log)
	storyHandler := httphandlers.NewStoryHandler(createStoryUseCase, getStoryUseCase, updateStoryUseCase, listStoriesUseCase, cloneStoryUseCase, log)
	chapterHandler := httphandlers.NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)
	sceneHandler := httphandlers.NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addSceneReferenceUseCase, removeSceneReferenceUseCase, getSceneReferencesUseCase, log)
	beatHandler := httphandlers.NewBeatHandler(createBeatUseCase, getBeatUseCase, updateBeatUseCase, deleteBeatUseCase, listBeatsUseCase, moveBeatUseCase, log)
	contentBlockHandler := httphandlers.NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)
	contentAnchorHandler := httphandlers.NewContentAnchorHandler(createContentAnchorUseCase, listContentAnchorsByContentBlockUseCase, listContentBlocksByEntityUseCase, deleteContentAnchorUseCase, log)
	entityRelationHandler := httphandlers.NewEntityRelationHandler(createRelationUseCase, getRelationUseCase, listRelationsBySourceUseCase, listRelationsByTargetUseCase, listRelationsByWorldUseCase, updateRelationUseCase, deleteRelationUseCase, log)
	relationMapHandler := httphandlers.NewRelationMapHandler()

	// Create router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("POST /api/v1/tenants", tenantHandler.Create)
	mux.HandleFunc("GET /api/v1/tenants/{id}", tenantHandler.Get)

	mux.HandleFunc("POST /api/v1/worlds", worldHandler.Create)
	mux.HandleFunc("GET /api/v1/worlds", worldHandler.List)
	mux.HandleFunc("GET /api/v1/worlds/{id}", worldHandler.Get)
	mux.HandleFunc("PUT /api/v1/worlds/{id}", worldHandler.Update)
	mux.HandleFunc("DELETE /api/v1/worlds/{id}", worldHandler.Delete)

	mux.HandleFunc("POST /api/v1/worlds/{world_id}/locations", locationHandler.Create)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/locations", locationHandler.List)
	mux.HandleFunc("GET /api/v1/locations/{id}", locationHandler.Get)
	mux.HandleFunc("PUT /api/v1/locations/{id}", locationHandler.Update)
	mux.HandleFunc("DELETE /api/v1/locations/{id}", locationHandler.Delete)
	mux.HandleFunc("GET /api/v1/locations/{id}/children", locationHandler.GetChildren)
	mux.HandleFunc("GET /api/v1/locations/{id}/ancestors", locationHandler.GetAncestors)
	mux.HandleFunc("GET /api/v1/locations/{id}/descendants", locationHandler.GetDescendants)
	mux.HandleFunc("PUT /api/v1/locations/{id}/move", locationHandler.Move)

	mux.HandleFunc("POST /api/v1/worlds/{world_id}/characters", characterHandler.Create)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/characters", characterHandler.List)
	mux.HandleFunc("GET /api/v1/characters/{id}", characterHandler.Get)
	mux.HandleFunc("PUT /api/v1/characters/{id}", characterHandler.Update)
	mux.HandleFunc("DELETE /api/v1/characters/{id}", characterHandler.Delete)
	mux.HandleFunc("GET /api/v1/characters/{id}/traits", characterHandler.GetTraits)
	mux.HandleFunc("POST /api/v1/characters/{id}/traits", characterHandler.AddTrait)
	mux.HandleFunc("PUT /api/v1/characters/{id}/traits/{trait_id}", characterHandler.UpdateTrait)
	mux.HandleFunc("DELETE /api/v1/characters/{id}/traits/{trait_id}", characterHandler.RemoveTrait)
	mux.HandleFunc("GET /api/v1/characters/{id}/events", characterHandler.GetEvents)
	mux.HandleFunc("GET /api/v1/characters/{id}/relationships", characterHandler.ListRelationships)
	mux.HandleFunc("POST /api/v1/characters/{id}/relationships", characterHandler.CreateRelationship)
	mux.HandleFunc("PUT /api/v1/character-relationships/{id}", characterHandler.UpdateRelationship)
	mux.HandleFunc("DELETE /api/v1/character-relationships/{id}", characterHandler.DeleteRelationship)

	mux.HandleFunc("POST /api/v1/worlds/{world_id}/artifacts", artifactHandler.Create)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/artifacts", artifactHandler.List)
	mux.HandleFunc("GET /api/v1/artifacts/{id}", artifactHandler.Get)
	mux.HandleFunc("PUT /api/v1/artifacts/{id}", artifactHandler.Update)
	mux.HandleFunc("DELETE /api/v1/artifacts/{id}", artifactHandler.Delete)
	mux.HandleFunc("GET /api/v1/artifacts/{id}/references", artifactHandler.GetReferences)
	mux.HandleFunc("POST /api/v1/artifacts/{id}/references", artifactHandler.AddReference)
	mux.HandleFunc("DELETE /api/v1/artifacts/{id}/references/{entity_type}/{entity_id}", artifactHandler.RemoveReference)

	mux.HandleFunc("POST /api/v1/worlds/{world_id}/events", eventHandler.Create)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/events", eventHandler.List)
	mux.HandleFunc("GET /api/v1/events/{id}", eventHandler.Get)
	mux.HandleFunc("PUT /api/v1/events/{id}", eventHandler.Update)
	mux.HandleFunc("DELETE /api/v1/events/{id}", eventHandler.Delete)
	mux.HandleFunc("POST /api/v1/events/{id}/references", eventHandler.AddReference)
	mux.HandleFunc("GET /api/v1/events/{id}/references", eventHandler.GetReferences)
	mux.HandleFunc("PUT /api/v1/event-references/{id}", eventHandler.UpdateReference)
	mux.HandleFunc("DELETE /api/v1/events/{id}/references/{entity_type}/{entity_id}", eventHandler.RemoveReference)
	mux.HandleFunc("GET /api/v1/events/{id}/children", eventHandler.GetChildren)
	mux.HandleFunc("GET /api/v1/events/{id}/ancestors", eventHandler.GetAncestors)
	mux.HandleFunc("GET /api/v1/events/{id}/descendants", eventHandler.GetDescendants)
	mux.HandleFunc("PUT /api/v1/events/{id}/move", eventHandler.MoveEvent)
	mux.HandleFunc("PUT /api/v1/events/{id}/epoch", eventHandler.SetEpoch)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/epoch", eventHandler.GetEpoch)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/timeline", eventHandler.GetTimeline)
	mux.HandleFunc("GET /api/v1/events/{id}/stat-changes", eventHandler.GetStatChanges)

	mux.HandleFunc("POST /api/v1/worlds/{world_id}/factions", factionHandler.Create)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/factions", factionHandler.List)
	mux.HandleFunc("GET /api/v1/factions/{id}", factionHandler.Get)
	mux.HandleFunc("PUT /api/v1/factions/{id}", factionHandler.Update)
	mux.HandleFunc("DELETE /api/v1/factions/{id}", factionHandler.Delete)
	mux.HandleFunc("GET /api/v1/factions/{id}/children", factionHandler.GetChildren)
	mux.HandleFunc("POST /api/v1/factions/{id}/references", factionHandler.AddReference)
	mux.HandleFunc("GET /api/v1/factions/{id}/references", factionHandler.GetReferences)
	mux.HandleFunc("PUT /api/v1/faction-references/{id}", factionHandler.UpdateReference)
	mux.HandleFunc("DELETE /api/v1/factions/{id}/references/{entity_type}/{entity_id}", factionHandler.RemoveReference)

	mux.HandleFunc("POST /api/v1/worlds/{world_id}/lores", loreHandler.Create)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/lores", loreHandler.List)
	mux.HandleFunc("GET /api/v1/lores/{id}", loreHandler.Get)
	mux.HandleFunc("PUT /api/v1/lores/{id}", loreHandler.Update)
	mux.HandleFunc("DELETE /api/v1/lores/{id}", loreHandler.Delete)
	mux.HandleFunc("GET /api/v1/lores/{id}/children", loreHandler.GetChildren)
	mux.HandleFunc("POST /api/v1/lores/{id}/references", loreHandler.AddReference)
	mux.HandleFunc("GET /api/v1/lores/{id}/references", loreHandler.GetReferences)
	mux.HandleFunc("PUT /api/v1/lore-references/{id}", loreHandler.UpdateReference)
	mux.HandleFunc("DELETE /api/v1/lores/{id}/references/{entity_type}/{entity_id}", loreHandler.RemoveReference)

	mux.HandleFunc("POST /api/v1/traits", traitHandler.Create)
	mux.HandleFunc("GET /api/v1/traits", traitHandler.List)
	mux.HandleFunc("GET /api/v1/traits/{id}", traitHandler.Get)
	mux.HandleFunc("PUT /api/v1/traits/{id}", traitHandler.Update)
	mux.HandleFunc("DELETE /api/v1/traits/{id}", traitHandler.Delete)

	mux.HandleFunc("POST /api/v1/archetypes", archetypeHandler.Create)
	mux.HandleFunc("GET /api/v1/archetypes", archetypeHandler.List)
	mux.HandleFunc("GET /api/v1/archetypes/{id}", archetypeHandler.Get)
	mux.HandleFunc("PUT /api/v1/archetypes/{id}", archetypeHandler.Update)
	mux.HandleFunc("DELETE /api/v1/archetypes/{id}", archetypeHandler.Delete)
	mux.HandleFunc("GET /api/v1/archetypes/{id}/traits", archetypeHandler.GetTraits)
	mux.HandleFunc("POST /api/v1/archetypes/{id}/traits", archetypeHandler.AddTrait)
	mux.HandleFunc("DELETE /api/v1/archetypes/{id}/traits/{trait_id}", archetypeHandler.RemoveTrait)

	mux.HandleFunc("POST /api/v1/stories", storyHandler.Create)
	mux.HandleFunc("GET /api/v1/stories/{id}", storyHandler.Get)
	mux.HandleFunc("PUT /api/v1/stories/{id}", storyHandler.Update)
	mux.HandleFunc("GET /api/v1/stories", storyHandler.List)
	mux.HandleFunc("POST /api/v1/stories/{id}/clone", storyHandler.Clone)

	mux.HandleFunc("POST /api/v1/chapters", chapterHandler.Create)
	mux.HandleFunc("GET /api/v1/chapters/{id}", chapterHandler.Get)
	mux.HandleFunc("PUT /api/v1/chapters/{id}", chapterHandler.Update)
	mux.HandleFunc("GET /api/v1/stories/{id}/chapters", chapterHandler.List)
	mux.HandleFunc("DELETE /api/v1/chapters/{id}", chapterHandler.Delete)

	mux.HandleFunc("POST /api/v1/scenes", sceneHandler.Create)
	mux.HandleFunc("GET /api/v1/scenes/{id}", sceneHandler.Get)
	mux.HandleFunc("PUT /api/v1/scenes/{id}", sceneHandler.Update)
	mux.HandleFunc("PUT /api/v1/scenes/{id}/move", sceneHandler.Move)
	mux.HandleFunc("GET /api/v1/stories/{id}/scenes", sceneHandler.ListByStory)
	mux.HandleFunc("GET /api/v1/chapters/{id}/scenes", sceneHandler.List)
	mux.HandleFunc("DELETE /api/v1/scenes/{id}", sceneHandler.Delete)
	mux.HandleFunc("GET /api/v1/scenes/{id}/references", sceneHandler.GetReferences)
	mux.HandleFunc("POST /api/v1/scenes/{id}/references", sceneHandler.AddReference)
	mux.HandleFunc("DELETE /api/v1/scenes/{id}/references/{entity_type}/{entity_id}", sceneHandler.RemoveReference)

	mux.HandleFunc("POST /api/v1/beats", beatHandler.Create)
	mux.HandleFunc("GET /api/v1/beats/{id}", beatHandler.Get)
	mux.HandleFunc("PUT /api/v1/beats/{id}", beatHandler.Update)
	mux.HandleFunc("PUT /api/v1/beats/{id}/move", beatHandler.Move)
	mux.HandleFunc("GET /api/v1/stories/{id}/beats", beatHandler.ListByStory)
	mux.HandleFunc("GET /api/v1/scenes/{id}/beats", beatHandler.List)
	mux.HandleFunc("DELETE /api/v1/beats/{id}", beatHandler.Delete)

	mux.HandleFunc("GET /api/v1/chapters/{id}/content-blocks", contentBlockHandler.ListByChapter)
	mux.HandleFunc("POST /api/v1/chapters/{id}/content-blocks", contentBlockHandler.Create)
	mux.HandleFunc("GET /api/v1/content-blocks/{id}", contentBlockHandler.Get)
	mux.HandleFunc("PUT /api/v1/content-blocks/{id}", contentBlockHandler.Update)
	mux.HandleFunc("DELETE /api/v1/content-blocks/{id}", contentBlockHandler.Delete)

	mux.HandleFunc("POST /api/v1/content-blocks/{id}/anchors", contentAnchorHandler.Create)
	mux.HandleFunc("GET /api/v1/content-blocks/{id}/anchors", contentAnchorHandler.ListByContentBlock)
	mux.HandleFunc("POST /api/v1/content-blocks/{id}/references", contentAnchorHandler.Create)            // deprecated alias
	mux.HandleFunc("GET /api/v1/content-blocks/{id}/references", contentAnchorHandler.ListByContentBlock) // deprecated alias
	mux.HandleFunc("GET /api/v1/scenes/{id}/content-blocks", contentAnchorHandler.ListByScene)
	mux.HandleFunc("GET /api/v1/beats/{id}/content-blocks", contentAnchorHandler.ListByBeat)
	mux.HandleFunc("DELETE /api/v1/content-anchors/{id}", contentAnchorHandler.Delete)
	mux.HandleFunc("DELETE /api/v1/content-block-references/{id}", contentAnchorHandler.Delete) // deprecated alias

	mux.HandleFunc("GET /api/v1/rpg-systems", rpgSystemHandler.List)
	mux.HandleFunc("POST /api/v1/rpg-systems", rpgSystemHandler.Create)
	mux.HandleFunc("GET /api/v1/rpg-systems/{id}", rpgSystemHandler.Get)
	mux.HandleFunc("PUT /api/v1/rpg-systems/{id}", rpgSystemHandler.Update)
	mux.HandleFunc("DELETE /api/v1/rpg-systems/{id}", rpgSystemHandler.Delete)

	mux.HandleFunc("GET /api/v1/characters/{id}/rpg-stats", characterRPGStatsHandler.GetActive)
	mux.HandleFunc("POST /api/v1/characters/{id}/rpg-stats", characterRPGStatsHandler.Create)
	mux.HandleFunc("GET /api/v1/characters/{id}/rpg-stats/history", characterRPGStatsHandler.ListHistory)
	mux.HandleFunc("PUT /api/v1/characters/{id}/rpg-stats/{stats_id}/activate", characterRPGStatsHandler.ActivateVersion)
	mux.HandleFunc("DELETE /api/v1/characters/{id}/rpg-stats", characterRPGStatsHandler.DeleteAll)

	mux.HandleFunc("GET /api/v1/artifacts/{id}/rpg-stats", artifactRPGStatsHandler.GetActive)
	mux.HandleFunc("POST /api/v1/artifacts/{id}/rpg-stats", artifactRPGStatsHandler.Create)
	mux.HandleFunc("GET /api/v1/artifacts/{id}/rpg-stats/history", artifactRPGStatsHandler.ListHistory)
	mux.HandleFunc("PUT /api/v1/artifacts/{id}/rpg-stats/{stats_id}/activate", artifactRPGStatsHandler.ActivateVersion)

	mux.HandleFunc("GET /api/v1/rpg-systems/{id}/skills", skillHandler.List)
	mux.HandleFunc("POST /api/v1/rpg-systems/{id}/skills", skillHandler.Create)
	mux.HandleFunc("GET /api/v1/rpg-skills/{id}", skillHandler.Get)
	mux.HandleFunc("PUT /api/v1/rpg-skills/{id}", skillHandler.Update)
	mux.HandleFunc("DELETE /api/v1/rpg-skills/{id}", skillHandler.Delete)

	mux.HandleFunc("GET /api/v1/characters/{id}/skills", characterSkillHandler.List)
	mux.HandleFunc("POST /api/v1/characters/{id}/skills", characterSkillHandler.Learn)
	mux.HandleFunc("PUT /api/v1/characters/{id}/skills/{skill_id}", characterSkillHandler.Update)
	mux.HandleFunc("DELETE /api/v1/characters/{id}/skills/{skill_id}", characterSkillHandler.Delete)
	mux.HandleFunc("PUT /api/v1/character-skills/{id}", characterSkillHandler.UpdateByID)
	mux.HandleFunc("DELETE /api/v1/character-skills/{id}", characterSkillHandler.DeleteByID)

	mux.HandleFunc("GET /api/v1/rpg-systems/{id}/classes", rpgClassHandler.List)
	mux.HandleFunc("POST /api/v1/rpg-systems/{id}/classes", rpgClassHandler.Create)
	mux.HandleFunc("GET /api/v1/rpg-classes/{id}", rpgClassHandler.Get)
	mux.HandleFunc("PUT /api/v1/rpg-classes/{id}", rpgClassHandler.Update)
	mux.HandleFunc("DELETE /api/v1/rpg-classes/{id}", rpgClassHandler.Delete)
	mux.HandleFunc("GET /api/v1/rpg-classes/{id}/skills", rpgClassHandler.ListSkills)
	mux.HandleFunc("POST /api/v1/rpg-classes/{id}/skills", rpgClassHandler.AddSkill)

	mux.HandleFunc("PUT /api/v1/characters/{id}/class", characterHandler.ChangeClass)
	mux.HandleFunc("GET /api/v1/characters/{id}/available-classes", characterHandler.GetAvailableClasses)

	mux.HandleFunc("GET /api/v1/rpg-systems/{id}/inventory-slots", inventoryHandler.ListSlots)
	mux.HandleFunc("POST /api/v1/rpg-systems/{id}/inventory-slots", inventoryHandler.CreateSlot)
	mux.HandleFunc("GET /api/v1/rpg-systems/{id}/inventory-items", inventoryHandler.ListItems)
	mux.HandleFunc("POST /api/v1/rpg-systems/{id}/inventory-items", inventoryHandler.CreateItem)
	mux.HandleFunc("GET /api/v1/inventory-items/{id}", inventoryHandler.GetItem)
	mux.HandleFunc("PUT /api/v1/inventory-items/{id}", inventoryHandler.UpdateItem)
	mux.HandleFunc("DELETE /api/v1/inventory-items/{id}", inventoryHandler.DeleteItem)
	mux.HandleFunc("GET /api/v1/characters/{id}/inventory", inventoryHandler.ListInventory)
	mux.HandleFunc("POST /api/v1/characters/{id}/inventory", inventoryHandler.AddItem)
	mux.HandleFunc("PUT /api/v1/character-inventory/{id}", inventoryHandler.UpdateInventory)
	mux.HandleFunc("PUT /api/v1/character-inventory/{id}/equip", inventoryHandler.EquipItem)
	mux.HandleFunc("PUT /api/v1/character-inventory/{id}/unequip", inventoryHandler.UnequipItem)
	mux.HandleFunc("DELETE /api/v1/character-inventory/{id}", inventoryHandler.DeleteInventory)
	mux.HandleFunc("POST /api/v1/character-inventory/{id}/transfer", inventoryHandler.TransferItem)

	// Entity relations routes
	mux.HandleFunc("POST /api/v1/relations", entityRelationHandler.Create)
	mux.HandleFunc("GET /api/v1/relations/{id}", entityRelationHandler.Get)
	mux.HandleFunc("PUT /api/v1/relations/{id}", entityRelationHandler.Update)
	mux.HandleFunc("DELETE /api/v1/relations/{id}", entityRelationHandler.Delete)
	mux.HandleFunc("GET /api/v1/worlds/{world_id}/relations", entityRelationHandler.ListByWorld)
	mux.HandleFunc("GET /api/v1/relations/source", entityRelationHandler.ListBySource)
	mux.HandleFunc("GET /api/v1/relations/target", entityRelationHandler.ListByTarget)

	mux.HandleFunc("GET /api/v1/static/relations", relationMapHandler.Types)
	mux.HandleFunc("GET /api/v1/static/relations/{entity_type}", relationMapHandler.Map)

	mux.HandleFunc("GET /health", httphandlers.HealthCheck)

	// Wrap with middleware
	// TenantMiddleware validates X-Tenant-ID header and injects tenantID into context
	// Routes like /api/v1/tenants and /health are skipped by the middleware
	handler := middleware.Chain(
		mux,
		middleware.TenantMiddleware,
		middleware.Logging(log),
		middleware.Recovery(log),
		middleware.CORS(),
	)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		log.Info("HTTP server listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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

		// Graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Error("server shutdown error", "error", err)
			os.Exit(1)
		}

		log.Info("HTTP server stopped")
	}
}
