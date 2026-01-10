package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/story-engine/main-service/internal/adapters/db/sqlite"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/application/story"
	beatapp "github.com/story-engine/main-service/internal/application/story/beat"
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	contentblockapp "github.com/story-engine/main-service/internal/application/story/content_block"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
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
	platformtenant "github.com/story-engine/main-service/internal/platform/tenant"
	httphandlers "github.com/story-engine/main-service/internal/transport/http/handlers"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	log := logger.New()

	log.Info("Starting offline HTTP server", "port", cfg.HTTP.Port)
	log.Info("LLM gateway notifications disabled in offline mode")
	if cfg.LLM.Enabled {
		cfg.LLM.Enabled = false
	}

	// Connect to SQLite database
	sqliteDB, err := database.NewSQLite(cfg.Database)
	if err != nil {
		log.Error("failed to connect to SQLite database", "error", err)
		os.Exit(1)
	}
	defer sqliteDB.Close()

	// Apply migrations
	log.Info("Applying database migrations...")
	if err := sqlite.ApplyMigrations(sqliteDB.DB()); err != nil {
		log.Error("failed to apply migrations", "error", err)
		os.Exit(1)
	}
	log.Info("Database migrations applied successfully")

	sqliteDBWrapper := sqlite.NewDB(sqliteDB)

	// Initialize repositories (only Story + World Building, no User/Membership/RPG)
	tenantRepo := sqlite.NewTenantRepository(sqliteDBWrapper)
	worldRepo := sqlite.NewWorldRepository(sqliteDBWrapper)
	locationRepo := sqlite.NewLocationRepository(sqliteDBWrapper)
	traitRepo := sqlite.NewTraitRepository(sqliteDBWrapper)
	archetypeRepo := sqlite.NewArchetypeRepository(sqliteDBWrapper)
	archetypeTraitRepo := sqlite.NewArchetypeTraitRepository(sqliteDBWrapper)
	characterRepo := sqlite.NewCharacterRepository(sqliteDBWrapper)
	characterTraitRepo := sqlite.NewCharacterTraitRepository(sqliteDBWrapper)
	entityRelationRepo := sqlite.NewEntityRelationRepository(sqliteDBWrapper)
	artifactRepo := sqlite.NewArtifactRepository(sqliteDBWrapper)
	eventRepo := sqlite.NewEventRepository(sqliteDBWrapper)
	factionRepo := sqlite.NewFactionRepository(sqliteDBWrapper)
	loreRepo := sqlite.NewLoreRepository(sqliteDBWrapper)
	storyRepo := sqlite.NewStoryRepository(sqliteDBWrapper)
	chapterRepo := sqlite.NewChapterRepository(sqliteDBWrapper)
	sceneRepo := sqlite.NewSceneRepository(sqliteDBWrapper)
	beatRepo := sqlite.NewBeatRepository(sqliteDBWrapper)
	contentBlockRepo := sqlite.NewContentBlockRepository(sqliteDBWrapper)
	contentAnchorRepo := sqlite.NewContentAnchorRepository(sqliteDBWrapper)
	auditLogRepo := sqlite.NewNoopAuditLogRepository() // No-op for offline mode
	transactionRepo := sqlite.NewTransactionRepository(sqliteDBWrapper)

	// Setup default tenant for offline mode
	ctx := context.Background()
	defaultTenantID, err := platformtenant.SetupDefaultTenant(ctx, tenantRepo, log)
	if err != nil {
		log.Error("failed to setup default tenant", "error", err)
		os.Exit(1)
	}

	// Initialize use cases (only Story + World Building)
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
	// Entity relation use cases
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	getRelationUseCase := relationapp.NewGetRelationUseCase(entityRelationRepo, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	updateRelationUseCase := relationapp.NewUpdateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
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
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, nil, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, nil, log)
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
	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, nil, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, nil, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, nil, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, nil, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, entityRelationRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addSceneReferenceUseCase := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, storyRepo, createRelationUseCase, listRelationsBySourceUseCase, characterRepo, locationRepo, artifactRepo, log)
	removeSceneReferenceUseCase := sceneapp.NewRemoveSceneReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getSceneReferencesUseCase := sceneapp.NewGetSceneReferencesUseCase(listRelationsBySourceUseCase, log)
	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, nil, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, nil, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)
	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, nil, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, nil, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	createContentAnchorUseCase := contentblockapp.NewCreateContentAnchorUseCase(contentAnchorRepo, contentBlockRepo, log)
	listContentAnchorsByContentBlockUseCase := contentblockapp.NewListContentAnchorsByContentBlockUseCase(contentAnchorRepo, contentBlockRepo, log)
	listContentBlocksByEntityUseCase := contentblockapp.NewListContentBlocksByEntityUseCase(contentAnchorRepo, contentBlockRepo, log)
	deleteContentAnchorUseCase := contentblockapp.NewDeleteContentAnchorUseCase(contentAnchorRepo, log)

	// Create handlers (only Story + World Building)
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
		nil, // changeClassUseCase - not available in offline mode
		nil, // getAvailableClassesUseCase - not available in offline mode
		log,
	)
	artifactHandler := httphandlers.NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)
	eventHandler := httphandlers.NewEventHandler(createEventUseCase, getEventUseCase, listEventsUseCase, updateEventUseCase, deleteEventUseCase, addEventReferenceUseCase, removeEventReferenceUseCase, getEventReferencesUseCase, updateEventReferenceUseCase, getEventChildrenUseCase, getEventAncestorsUseCase, getEventDescendantsUseCase, moveEventUseCase, setEventEpochUseCase, getEventEpochUseCase, getTimelineUseCase, nil, log) // No RPG stat changes for offline mode
	factionHandler := httphandlers.NewFactionHandler(createFactionUseCase, getFactionUseCase, listFactionsUseCase, updateFactionUseCase, deleteFactionUseCase, getFactionChildrenUseCase, addFactionReferenceUseCase, removeFactionReferenceUseCase, getFactionReferencesUseCase, updateFactionReferenceUseCase, log)
	loreHandler := httphandlers.NewLoreHandler(createLoreUseCase, getLoreUseCase, listLoresUseCase, updateLoreUseCase, deleteLoreUseCase, getLoreChildrenUseCase, addLoreReferenceUseCase, removeLoreReferenceUseCase, getLoreReferencesUseCase, updateLoreReferenceUseCase, log)
	traitHandler := httphandlers.NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	archetypeHandler := httphandlers.NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitToArchetypeUseCase, removeTraitFromArchetypeUseCase, getArchetypeTraitsUseCase, log)
	storyHandler := httphandlers.NewStoryHandler(createStoryUseCase, getStoryUseCase, updateStoryUseCase, listStoriesUseCase, cloneStoryUseCase, log)
	chapterHandler := httphandlers.NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)
	sceneHandler := httphandlers.NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addSceneReferenceUseCase, removeSceneReferenceUseCase, getSceneReferencesUseCase, log)
	beatHandler := httphandlers.NewBeatHandler(createBeatUseCase, getBeatUseCase, updateBeatUseCase, deleteBeatUseCase, listBeatsUseCase, moveBeatUseCase, log)
	contentBlockHandler := httphandlers.NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)
	contentAnchorHandler := httphandlers.NewContentAnchorHandler(createContentAnchorUseCase, listContentAnchorsByContentBlockUseCase, listContentBlocksByEntityUseCase, deleteContentAnchorUseCase, log)
	relationMapHandler := httphandlers.NewRelationMapHandler()

	// Create router
	router := http.NewServeMux()

	// Register routes (only Story + World Building, no User/Membership/RPG)

	// Relation map routes
	router.HandleFunc("GET /api/v1/static/relations", relationMapHandler.Types)
	router.HandleFunc("GET /api/v1/static/relations/{entity_type}", relationMapHandler.Map)

	// World Building routes
	router.HandleFunc("POST /api/v1/worlds", worldHandler.Create)
	router.HandleFunc("GET /api/v1/worlds", worldHandler.List)
	router.HandleFunc("GET /api/v1/worlds/{id}", worldHandler.Get)
	router.HandleFunc("PUT /api/v1/worlds/{id}", worldHandler.Update)
	router.HandleFunc("DELETE /api/v1/worlds/{id}", worldHandler.Delete)

	router.HandleFunc("POST /api/v1/worlds/{world_id}/locations", locationHandler.Create)
	router.HandleFunc("GET /api/v1/worlds/{world_id}/locations", locationHandler.List)
	router.HandleFunc("GET /api/v1/locations/{id}", locationHandler.Get)
	router.HandleFunc("PUT /api/v1/locations/{id}", locationHandler.Update)
	router.HandleFunc("DELETE /api/v1/locations/{id}", locationHandler.Delete)
	router.HandleFunc("GET /api/v1/locations/{id}/children", locationHandler.GetChildren)
	router.HandleFunc("GET /api/v1/locations/{id}/ancestors", locationHandler.GetAncestors)
	router.HandleFunc("GET /api/v1/locations/{id}/descendants", locationHandler.GetDescendants)
	router.HandleFunc("PUT /api/v1/locations/{id}/move", locationHandler.Move)

	router.HandleFunc("POST /api/v1/worlds/{world_id}/characters", characterHandler.Create)
	router.HandleFunc("GET /api/v1/worlds/{world_id}/characters", characterHandler.List)
	router.HandleFunc("GET /api/v1/characters/{id}", characterHandler.Get)
	router.HandleFunc("PUT /api/v1/characters/{id}", characterHandler.Update)
	router.HandleFunc("DELETE /api/v1/characters/{id}", characterHandler.Delete)
	router.HandleFunc("GET /api/v1/characters/{id}/traits", characterHandler.GetTraits)
	router.HandleFunc("POST /api/v1/characters/{id}/traits", characterHandler.AddTrait)
	router.HandleFunc("PUT /api/v1/characters/{id}/traits/{trait_id}", characterHandler.UpdateTrait)
	router.HandleFunc("DELETE /api/v1/characters/{id}/traits/{trait_id}", characterHandler.RemoveTrait)
	router.HandleFunc("GET /api/v1/characters/{id}/events", characterHandler.GetEvents)
	router.HandleFunc("GET /api/v1/characters/{id}/relationships", characterHandler.ListRelationships)
	router.HandleFunc("POST /api/v1/characters/{id}/relationships", characterHandler.CreateRelationship)
	router.HandleFunc("PUT /api/v1/character-relationships/{id}", characterHandler.UpdateRelationship)
	router.HandleFunc("DELETE /api/v1/character-relationships/{id}", characterHandler.DeleteRelationship)

	router.HandleFunc("POST /api/v1/worlds/{world_id}/artifacts", artifactHandler.Create)
	router.HandleFunc("GET /api/v1/worlds/{world_id}/artifacts", artifactHandler.List)
	router.HandleFunc("GET /api/v1/artifacts/{id}", artifactHandler.Get)
	router.HandleFunc("PUT /api/v1/artifacts/{id}", artifactHandler.Update)
	router.HandleFunc("DELETE /api/v1/artifacts/{id}", artifactHandler.Delete)
	router.HandleFunc("GET /api/v1/artifacts/{id}/references", artifactHandler.GetReferences)
	router.HandleFunc("POST /api/v1/artifacts/{id}/references", artifactHandler.AddReference)
	router.HandleFunc("DELETE /api/v1/artifacts/{id}/references/{entity_type}/{entity_id}", artifactHandler.RemoveReference)

	router.HandleFunc("POST /api/v1/worlds/{world_id}/events", eventHandler.Create)
	router.HandleFunc("GET /api/v1/worlds/{world_id}/events", eventHandler.List)
	router.HandleFunc("GET /api/v1/events/{id}", eventHandler.Get)
	router.HandleFunc("PUT /api/v1/events/{id}", eventHandler.Update)
	router.HandleFunc("DELETE /api/v1/events/{id}", eventHandler.Delete)
	router.HandleFunc("POST /api/v1/events/{id}/references", eventHandler.AddReference)
	router.HandleFunc("GET /api/v1/events/{id}/references", eventHandler.GetReferences)
	router.HandleFunc("PUT /api/v1/event-references/{id}", eventHandler.UpdateReference)
	router.HandleFunc("DELETE /api/v1/events/{id}/references/{entity_type}/{entity_id}", eventHandler.RemoveReference)
	router.HandleFunc("GET /api/v1/events/{id}/children", eventHandler.GetChildren)
	router.HandleFunc("GET /api/v1/events/{id}/ancestors", eventHandler.GetAncestors)
	router.HandleFunc("GET /api/v1/events/{id}/descendants", eventHandler.GetDescendants)
	router.HandleFunc("PUT /api/v1/events/{id}/move", eventHandler.MoveEvent)
	router.HandleFunc("PUT /api/v1/events/{id}/epoch", eventHandler.SetEpoch)
	router.HandleFunc("GET /api/v1/worlds/{world_id}/epoch", eventHandler.GetEpoch)
	router.HandleFunc("GET /api/v1/worlds/{world_id}/timeline", eventHandler.GetTimeline)

	router.HandleFunc("POST /api/v1/worlds/{world_id}/factions", factionHandler.Create)
	router.HandleFunc("GET /api/v1/worlds/{world_id}/factions", factionHandler.List)
	router.HandleFunc("GET /api/v1/factions/{id}", factionHandler.Get)
	router.HandleFunc("PUT /api/v1/factions/{id}", factionHandler.Update)
	router.HandleFunc("DELETE /api/v1/factions/{id}", factionHandler.Delete)
	router.HandleFunc("GET /api/v1/factions/{id}/children", factionHandler.GetChildren)
	router.HandleFunc("POST /api/v1/factions/{id}/references", factionHandler.AddReference)
	router.HandleFunc("GET /api/v1/factions/{id}/references", factionHandler.GetReferences)
	router.HandleFunc("PUT /api/v1/faction-references/{id}", factionHandler.UpdateReference)
	router.HandleFunc("DELETE /api/v1/factions/{id}/references/{entity_type}/{entity_id}", factionHandler.RemoveReference)

	router.HandleFunc("POST /api/v1/worlds/{world_id}/lores", loreHandler.Create)
	router.HandleFunc("GET /api/v1/worlds/{world_id}/lores", loreHandler.List)
	router.HandleFunc("GET /api/v1/lores/{id}", loreHandler.Get)
	router.HandleFunc("PUT /api/v1/lores/{id}", loreHandler.Update)
	router.HandleFunc("DELETE /api/v1/lores/{id}", loreHandler.Delete)
	router.HandleFunc("GET /api/v1/lores/{id}/children", loreHandler.GetChildren)
	router.HandleFunc("POST /api/v1/lores/{id}/references", loreHandler.AddReference)
	router.HandleFunc("GET /api/v1/lores/{id}/references", loreHandler.GetReferences)
	router.HandleFunc("PUT /api/v1/lore-references/{id}", loreHandler.UpdateReference)
	router.HandleFunc("DELETE /api/v1/lores/{id}/references/{entity_type}/{entity_id}", loreHandler.RemoveReference)

	router.HandleFunc("POST /api/v1/traits", traitHandler.Create)
	router.HandleFunc("GET /api/v1/traits", traitHandler.List)
	router.HandleFunc("GET /api/v1/traits/{id}", traitHandler.Get)
	router.HandleFunc("PUT /api/v1/traits/{id}", traitHandler.Update)
	router.HandleFunc("DELETE /api/v1/traits/{id}", traitHandler.Delete)

	router.HandleFunc("POST /api/v1/archetypes", archetypeHandler.Create)
	router.HandleFunc("GET /api/v1/archetypes", archetypeHandler.List)
	router.HandleFunc("GET /api/v1/archetypes/{id}", archetypeHandler.Get)
	router.HandleFunc("PUT /api/v1/archetypes/{id}", archetypeHandler.Update)
	router.HandleFunc("DELETE /api/v1/archetypes/{id}", archetypeHandler.Delete)
	router.HandleFunc("GET /api/v1/archetypes/{id}/traits", archetypeHandler.GetTraits)
	router.HandleFunc("POST /api/v1/archetypes/{id}/traits", archetypeHandler.AddTrait)
	router.HandleFunc("DELETE /api/v1/archetypes/{id}/traits/{trait_id}", archetypeHandler.RemoveTrait)

	// Story routes
	router.HandleFunc("POST /api/v1/stories", storyHandler.Create)
	router.HandleFunc("GET /api/v1/stories/{id}", storyHandler.Get)
	router.HandleFunc("PUT /api/v1/stories/{id}", storyHandler.Update)
	router.HandleFunc("GET /api/v1/stories", storyHandler.List)
	router.HandleFunc("POST /api/v1/stories/{id}/clone", storyHandler.Clone)

	router.HandleFunc("POST /api/v1/chapters", chapterHandler.Create)
	router.HandleFunc("GET /api/v1/chapters/{id}", chapterHandler.Get)
	router.HandleFunc("PUT /api/v1/chapters/{id}", chapterHandler.Update)
	router.HandleFunc("GET /api/v1/stories/{id}/chapters", chapterHandler.List)
	router.HandleFunc("DELETE /api/v1/chapters/{id}", chapterHandler.Delete)

	router.HandleFunc("POST /api/v1/scenes", sceneHandler.Create)
	router.HandleFunc("GET /api/v1/scenes/{id}", sceneHandler.Get)
	router.HandleFunc("PUT /api/v1/scenes/{id}", sceneHandler.Update)
	router.HandleFunc("PUT /api/v1/scenes/{id}/move", sceneHandler.Move)
	router.HandleFunc("GET /api/v1/stories/{id}/scenes", sceneHandler.ListByStory)
	router.HandleFunc("GET /api/v1/chapters/{id}/scenes", sceneHandler.List)
	router.HandleFunc("DELETE /api/v1/scenes/{id}", sceneHandler.Delete)
	router.HandleFunc("GET /api/v1/scenes/{id}/references", sceneHandler.GetReferences)
	router.HandleFunc("POST /api/v1/scenes/{id}/references", sceneHandler.AddReference)
	router.HandleFunc("DELETE /api/v1/scenes/{id}/references/{entity_type}/{entity_id}", sceneHandler.RemoveReference)

	router.HandleFunc("POST /api/v1/beats", beatHandler.Create)
	router.HandleFunc("GET /api/v1/beats/{id}", beatHandler.Get)
	router.HandleFunc("PUT /api/v1/beats/{id}", beatHandler.Update)
	router.HandleFunc("PUT /api/v1/beats/{id}/move", beatHandler.Move)
	router.HandleFunc("GET /api/v1/stories/{id}/beats", beatHandler.ListByStory)
	router.HandleFunc("GET /api/v1/scenes/{id}/beats", beatHandler.List)
	router.HandleFunc("DELETE /api/v1/beats/{id}", beatHandler.Delete)

	router.HandleFunc("GET /api/v1/chapters/{id}/content-blocks", contentBlockHandler.ListByChapter)
	router.HandleFunc("POST /api/v1/chapters/{id}/content-blocks", contentBlockHandler.Create)
	router.HandleFunc("GET /api/v1/content-blocks/{id}", contentBlockHandler.Get)
	router.HandleFunc("PUT /api/v1/content-blocks/{id}", contentBlockHandler.Update)
	router.HandleFunc("DELETE /api/v1/content-blocks/{id}", contentBlockHandler.Delete)

	router.HandleFunc("POST /api/v1/content-blocks/{id}/anchors", contentAnchorHandler.Create)
	router.HandleFunc("GET /api/v1/content-blocks/{id}/anchors", contentAnchorHandler.ListByContentBlock)
	router.HandleFunc("POST /api/v1/content-blocks/{id}/references", contentAnchorHandler.Create)            // deprecated alias
	router.HandleFunc("GET /api/v1/content-blocks/{id}/references", contentAnchorHandler.ListByContentBlock) // deprecated alias
	router.HandleFunc("GET /api/v1/scenes/{id}/content-blocks", contentAnchorHandler.ListByScene)
	router.HandleFunc("GET /api/v1/beats/{id}/content-blocks", contentAnchorHandler.ListByBeat)
	router.HandleFunc("DELETE /api/v1/content-anchors/{id}", contentAnchorHandler.Delete)
	router.HandleFunc("DELETE /api/v1/content-block-references/{id}", contentAnchorHandler.Delete) // deprecated alias

	router.HandleFunc("GET /health", httphandlers.HealthCheck)

	// Wrap with middleware
	// OfflineTenantMiddleware injects the default tenant ID into the request context
	// No need for X-Tenant-ID header in offline mode
	handler := middleware.Chain(
		router,
		middleware.OfflineTenantMiddleware(defaultTenantID),
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
		log.Info("Offline HTTP server listening", "address", server.Addr)
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
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error("server shutdown error", "error", err)
			os.Exit(1)
		}

		log.Info("Offline HTTP server stopped")
	}
}
