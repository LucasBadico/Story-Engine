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
	artifactRepo := sqlite.NewArtifactRepository(sqliteDBWrapper)
	artifactReferenceRepo := sqlite.NewArtifactReferenceRepository(sqliteDBWrapper)
	eventRepo := sqlite.NewEventRepository(sqliteDBWrapper)
	eventCharacterRepo := sqlite.NewEventCharacterRepository(sqliteDBWrapper)
	eventLocationRepo := sqlite.NewEventLocationRepository(sqliteDBWrapper)
	eventArtifactRepo := sqlite.NewEventArtifactRepository(sqliteDBWrapper)
	factionRepo := sqlite.NewFactionRepository(sqliteDBWrapper)
	factionReferenceRepo := sqlite.NewFactionReferenceRepository(sqliteDBWrapper)
	loreRepo := sqlite.NewLoreRepository(sqliteDBWrapper)
	loreReferenceRepo := sqlite.NewLoreReferenceRepository(sqliteDBWrapper)
	storyRepo := sqlite.NewStoryRepository(sqliteDBWrapper)
	chapterRepo := sqlite.NewChapterRepository(sqliteDBWrapper)
	sceneRepo := sqlite.NewSceneRepository(sqliteDBWrapper)
	sceneReferenceRepo := sqlite.NewSceneReferenceRepository(sqliteDBWrapper)
	beatRepo := sqlite.NewBeatRepository(sqliteDBWrapper)
	contentBlockRepo := sqlite.NewContentBlockRepository(sqliteDBWrapper)
	contentBlockReferenceRepo := sqlite.NewContentBlockReferenceRepository(sqliteDBWrapper)
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
	deleteLocationUseCase := locationapp.NewDeleteLocationUseCase(locationRepo, auditLogRepo, log)
	getChildrenUseCase := locationapp.NewGetChildrenUseCase(locationRepo, log)
	getAncestorsUseCase := locationapp.NewGetAncestorsUseCase(locationRepo, log)
	getDescendantsUseCase := locationapp.NewGetDescendantsUseCase(locationRepo, log)
	moveLocationUseCase := locationapp.NewMoveLocationUseCase(locationRepo, auditLogRepo, log)
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, worldRepo, auditLogRepo, log)
	addTraitToCharacterUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	removeTraitFromCharacterUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
	updateCharacterTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	getCharacterTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	createArtifactUseCase := artifactapp.NewCreateArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, characterRepo, locationRepo, auditLogRepo, log)
	getArtifactUseCase := artifactapp.NewGetArtifactUseCase(artifactRepo, log)
	listArtifactsUseCase := artifactapp.NewListArtifactsUseCase(artifactRepo, log)
	updateArtifactUseCase := artifactapp.NewUpdateArtifactUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, worldRepo, auditLogRepo, log)
	deleteArtifactUseCase := artifactapp.NewDeleteArtifactUseCase(artifactRepo, artifactReferenceRepo, worldRepo, auditLogRepo, log)
	getArtifactReferencesUseCase := artifactapp.NewGetArtifactReferencesUseCase(artifactReferenceRepo, log)
	addArtifactReferenceUseCase := artifactapp.NewAddArtifactReferenceUseCase(artifactRepo, artifactReferenceRepo, characterRepo, locationRepo, log)
	removeArtifactReferenceUseCase := artifactapp.NewRemoveArtifactReferenceUseCase(artifactReferenceRepo, log)
	createEventUseCase := eventapp.NewCreateEventUseCase(eventRepo, worldRepo, auditLogRepo, log)
	getEventUseCase := eventapp.NewGetEventUseCase(eventRepo, log)
	listEventsUseCase := eventapp.NewListEventsUseCase(eventRepo, log)
	updateEventUseCase := eventapp.NewUpdateEventUseCase(eventRepo, auditLogRepo, log)
	deleteEventUseCase := eventapp.NewDeleteEventUseCase(eventRepo, eventCharacterRepo, eventLocationRepo, eventArtifactRepo, auditLogRepo, log)
	addCharacterToEventUseCase := eventapp.NewAddCharacterToEventUseCase(eventRepo, characterRepo, eventCharacterRepo, log)
	removeCharacterFromEventUseCase := eventapp.NewRemoveCharacterFromEventUseCase(eventCharacterRepo, log)
	getEventCharactersUseCase := eventapp.NewGetEventCharactersUseCase(eventCharacterRepo, log)
	addLocationToEventUseCase := eventapp.NewAddLocationToEventUseCase(eventRepo, locationRepo, eventLocationRepo, log)
	removeLocationFromEventUseCase := eventapp.NewRemoveLocationFromEventUseCase(eventLocationRepo, log)
	getEventLocationsUseCase := eventapp.NewGetEventLocationsUseCase(eventLocationRepo, log)
	addArtifactToEventUseCase := eventapp.NewAddArtifactToEventUseCase(eventRepo, artifactRepo, eventArtifactRepo, log)
	removeArtifactFromEventUseCase := eventapp.NewRemoveArtifactFromEventUseCase(eventArtifactRepo, log)
	getEventArtifactsUseCase := eventapp.NewGetEventArtifactsUseCase(eventArtifactRepo, log)
	createFactionUseCase := factionapp.NewCreateFactionUseCase(factionRepo, worldRepo, auditLogRepo, log)
	getFactionUseCase := factionapp.NewGetFactionUseCase(factionRepo, log)
	listFactionsUseCase := factionapp.NewListFactionsUseCase(factionRepo, log)
	updateFactionUseCase := factionapp.NewUpdateFactionUseCase(factionRepo, auditLogRepo, log)
	deleteFactionUseCase := factionapp.NewDeleteFactionUseCase(factionRepo, factionReferenceRepo, auditLogRepo, log)
	getFactionChildrenUseCase := factionapp.NewGetChildrenUseCase(factionRepo, log)
	addFactionReferenceUseCase := factionapp.NewAddReferenceUseCase(factionRepo, factionReferenceRepo, characterRepo, locationRepo, artifactRepo, eventRepo, loreRepo, loreReferenceRepo, log)
	removeFactionReferenceUseCase := factionapp.NewRemoveReferenceUseCase(factionReferenceRepo, log)
	getFactionReferencesUseCase := factionapp.NewGetReferencesUseCase(factionReferenceRepo, log)
	updateFactionReferenceUseCase := factionapp.NewUpdateReferenceUseCase(factionReferenceRepo, log)
	createLoreUseCase := loreapp.NewCreateLoreUseCase(loreRepo, worldRepo, auditLogRepo, log)
	getLoreUseCase := loreapp.NewGetLoreUseCase(loreRepo, log)
	listLoresUseCase := loreapp.NewListLoresUseCase(loreRepo, log)
	updateLoreUseCase := loreapp.NewUpdateLoreUseCase(loreRepo, auditLogRepo, log)
	deleteLoreUseCase := loreapp.NewDeleteLoreUseCase(loreRepo, loreReferenceRepo, auditLogRepo, log)
	getLoreChildrenUseCase := loreapp.NewGetChildrenUseCase(loreRepo, log)
	addLoreReferenceUseCase := loreapp.NewAddReferenceUseCase(loreRepo, loreReferenceRepo, characterRepo, locationRepo, artifactRepo, eventRepo, factionRepo, factionReferenceRepo, log)
	removeLoreReferenceUseCase := loreapp.NewRemoveReferenceUseCase(loreReferenceRepo, log)
	getLoreReferencesUseCase := loreapp.NewGetReferencesUseCase(loreReferenceRepo, log)
	updateLoreReferenceUseCase := loreapp.NewUpdateReferenceUseCase(loreReferenceRepo, log)
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
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, log)
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
	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)
	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	addSceneReferenceUseCase := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeSceneReferenceUseCase := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getSceneReferencesUseCase := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)
	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)
	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)
	createContentBlockReferenceUseCase := contentblockapp.NewCreateContentBlockReferenceUseCase(contentBlockReferenceRepo, contentBlockRepo, log)
	listContentBlockReferencesByContentBlockUseCase := contentblockapp.NewListContentBlockReferencesByContentBlockUseCase(contentBlockReferenceRepo, contentBlockRepo, log)
	listContentBlocksByEntityUseCase := contentblockapp.NewListContentBlocksByEntityUseCase(contentBlockReferenceRepo, contentBlockRepo, log)
	deleteContentBlockReferenceUseCase := contentblockapp.NewDeleteContentBlockReferenceUseCase(contentBlockReferenceRepo, log)

	// Create handlers (only Story + World Building)
	worldHandler := httphandlers.NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	locationHandler := httphandlers.NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)
	characterHandler := httphandlers.NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitToCharacterUseCase, removeTraitFromCharacterUseCase, updateCharacterTraitUseCase, getCharacterTraitsUseCase, nil, nil, log) // No RPG class use cases for offline mode
	artifactHandler := httphandlers.NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)
	eventHandler := httphandlers.NewEventHandler(createEventUseCase, getEventUseCase, listEventsUseCase, updateEventUseCase, deleteEventUseCase, addCharacterToEventUseCase, removeCharacterFromEventUseCase, getEventCharactersUseCase, addLocationToEventUseCase, removeLocationFromEventUseCase, getEventLocationsUseCase, addArtifactToEventUseCase, removeArtifactFromEventUseCase, getEventArtifactsUseCase, nil, log) // No RPG stat changes for offline mode
	factionHandler := httphandlers.NewFactionHandler(createFactionUseCase, getFactionUseCase, listFactionsUseCase, updateFactionUseCase, deleteFactionUseCase, getFactionChildrenUseCase, addFactionReferenceUseCase, removeFactionReferenceUseCase, getFactionReferencesUseCase, updateFactionReferenceUseCase, log)
	loreHandler := httphandlers.NewLoreHandler(createLoreUseCase, getLoreUseCase, listLoresUseCase, updateLoreUseCase, deleteLoreUseCase, getLoreChildrenUseCase, addLoreReferenceUseCase, removeLoreReferenceUseCase, getLoreReferencesUseCase, updateLoreReferenceUseCase, log)
	traitHandler := httphandlers.NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	archetypeHandler := httphandlers.NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitToArchetypeUseCase, removeTraitFromArchetypeUseCase, log)
	storyHandler := httphandlers.NewStoryHandler(createStoryUseCase, getStoryUseCase, updateStoryUseCase, listStoriesUseCase, cloneStoryUseCase, log)
	chapterHandler := httphandlers.NewChapterHandler(createChapterUseCase, getChapterUseCase, updateChapterUseCase, deleteChapterUseCase, listChaptersUseCase, log)
	sceneHandler := httphandlers.NewSceneHandler(createSceneUseCase, getSceneUseCase, updateSceneUseCase, deleteSceneUseCase, listScenesUseCase, moveSceneUseCase, addSceneReferenceUseCase, removeSceneReferenceUseCase, getSceneReferencesUseCase, log)
	beatHandler := httphandlers.NewBeatHandler(createBeatUseCase, getBeatUseCase, updateBeatUseCase, deleteBeatUseCase, listBeatsUseCase, moveBeatUseCase, log)
	contentBlockHandler := httphandlers.NewContentBlockHandler(createContentBlockUseCase, getContentBlockUseCase, updateContentBlockUseCase, deleteContentBlockUseCase, listContentBlocksUseCase, log)
	contentBlockReferenceHandler := httphandlers.NewContentBlockReferenceHandler(createContentBlockReferenceUseCase, listContentBlockReferencesByContentBlockUseCase, listContentBlocksByEntityUseCase, deleteContentBlockReferenceUseCase, log)

	// Create router
	router := http.NewServeMux()

	// Register routes (only Story + World Building, no User/Membership/RPG)

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
	router.HandleFunc("POST /api/v1/events/{id}/characters", eventHandler.AddCharacter)
	router.HandleFunc("GET /api/v1/events/{id}/characters", eventHandler.GetCharacters)
	router.HandleFunc("DELETE /api/v1/events/{id}/characters/{character_id}", eventHandler.RemoveCharacter)
	router.HandleFunc("POST /api/v1/events/{id}/locations", eventHandler.AddLocation)
	router.HandleFunc("GET /api/v1/events/{id}/locations", eventHandler.GetLocations)
	router.HandleFunc("DELETE /api/v1/events/{id}/locations/{location_id}", eventHandler.RemoveLocation)
	router.HandleFunc("POST /api/v1/events/{id}/artifacts", eventHandler.AddArtifact)
	router.HandleFunc("GET /api/v1/events/{id}/artifacts", eventHandler.GetArtifacts)
	router.HandleFunc("DELETE /api/v1/events/{id}/artifacts/{artifact_id}", eventHandler.RemoveArtifact)

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

	router.HandleFunc("POST /api/v1/content-blocks/{id}/references", contentBlockReferenceHandler.Create)
	router.HandleFunc("GET /api/v1/content-blocks/{id}/references", contentBlockReferenceHandler.ListByContentBlock)
	router.HandleFunc("GET /api/v1/scenes/{id}/content-blocks", contentBlockReferenceHandler.ListByScene)
	router.HandleFunc("GET /api/v1/beats/{id}/content-blocks", contentBlockReferenceHandler.ListByBeat)
	router.HandleFunc("DELETE /api/v1/content-block-references/{id}", contentBlockReferenceHandler.Delete)

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
