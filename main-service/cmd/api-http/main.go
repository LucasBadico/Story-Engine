package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	worldapp "github.com/story-engine/main-service/internal/application/world"
	archetypeapp "github.com/story-engine/main-service/internal/application/world/archetype"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	rpgsystemapp "github.com/story-engine/main-service/internal/application/rpg/rpg_system"
	characterstatsapp "github.com/story-engine/main-service/internal/application/rpg/character_stats"
	artifactstatsapp "github.com/story-engine/main-service/internal/application/rpg/artifact_stats"
	rpgeventapp "github.com/story-engine/main-service/internal/application/rpg/event"
	skillapp "github.com/story-engine/main-service/internal/application/rpg/skill"
	characterskillapp "github.com/story-engine/main-service/internal/application/rpg/character_skill"
	rpgclassapp "github.com/story-engine/main-service/internal/application/rpg/rpg_class"
	rpgcharacterapp "github.com/story-engine/main-service/internal/application/rpg/character"
	inventoryslotapp "github.com/story-engine/main-service/internal/application/rpg/inventory_slot"
	inventoryitemapp "github.com/story-engine/main-service/internal/application/rpg/inventory_item"
	characterinventoryapp "github.com/story-engine/main-service/internal/application/rpg/character_inventory"
	"github.com/story-engine/main-service/internal/application/story"
	imageblockapp "github.com/story-engine/main-service/internal/application/story/image_block"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/database"
	"github.com/story-engine/main-service/internal/platform/logger"
	httphandlers "github.com/story-engine/main-service/internal/transport/http/handlers"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	log := logger.New()

	log.Info("Starting HTTP server", "port", cfg.HTTP.Port)

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
	artifactReferenceRepo := postgres.NewArtifactReferenceRepository(pgDB)
	eventRepo := postgres.NewEventRepository(pgDB)
	eventCharacterRepo := postgres.NewEventCharacterRepository(pgDB)
	eventLocationRepo := postgres.NewEventLocationRepository(pgDB)
	eventArtifactRepo := postgres.NewEventArtifactRepository(pgDB)
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
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(pgDB)
	beatRepo := postgres.NewBeatRepository(pgDB)
	proseBlockRepo := postgres.NewProseBlockRepository(pgDB)
	proseBlockReferenceRepo := postgres.NewProseBlockReferenceRepository(pgDB)
	imageBlockRepo := postgres.NewImageBlockRepository(pgDB)
	imageBlockReferenceRepo := postgres.NewImageBlockReferenceRepository(pgDB)
	auditLogRepo := postgres.NewAuditLogRepository(pgDB)
	transactionRepo := postgres.NewTransactionRepository(pgDB)

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
	cloneStoryUseCase := story.NewCloneStoryUseCase(
		storyRepo,
		chapterRepo,
		sceneRepo,
		beatRepo,
		proseBlockRepo,
		auditLogRepo,
		transactionRepo,
		log,
	)
	addSceneReferenceUseCase := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeSceneReferenceUseCase := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getSceneReferencesUseCase := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)
	createImageBlockUseCase := imageblockapp.NewCreateImageBlockUseCase(imageBlockRepo, chapterRepo, log)
	getImageBlockUseCase := imageblockapp.NewGetImageBlockUseCase(imageBlockRepo, log)
	listImageBlocksUseCase := imageblockapp.NewListImageBlocksUseCase(imageBlockRepo, log)
	updateImageBlockUseCase := imageblockapp.NewUpdateImageBlockUseCase(imageBlockRepo, log)
	deleteImageBlockUseCase := imageblockapp.NewDeleteImageBlockUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	addImageBlockReferenceUseCase := imageblockapp.NewAddImageBlockReferenceUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	removeImageBlockReferenceUseCase := imageblockapp.NewRemoveImageBlockReferenceUseCase(imageBlockReferenceRepo, log)
	getImageBlockReferencesUseCase := imageblockapp.NewGetImageBlockReferencesUseCase(imageBlockReferenceRepo, log)
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

	// Create handlers
	tenantHandler := httphandlers.NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := httphandlers.NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	locationHandler := httphandlers.NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)
	characterHandler := httphandlers.NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitToCharacterUseCase, removeTraitFromCharacterUseCase, updateCharacterTraitUseCase, getCharacterTraitsUseCase, changeCharacterClassUseCase, getAvailableClassesUseCase, log)
	rpgClassHandler := httphandlers.NewRPGClassHandler(createRPGClassUseCase, getRPGClassUseCase, listRPGClassesUseCase, updateRPGClassUseCase, deleteRPGClassUseCase, addSkillToClassUseCase, listClassSkillsUseCase, log)
	inventoryHandler := httphandlers.NewInventoryHandler(createInventorySlotUseCase, listInventorySlotsUseCase, createInventoryItemUseCase, getInventoryItemUseCase, listInventoryItemsUseCase, updateInventoryItemUseCase, deleteInventoryItemUseCase, addItemToInventoryUseCase, listCharacterInventoryUseCase, updateCharacterInventoryUseCase, equipItemUseCase, unequipItemUseCase, deleteCharacterInventoryUseCase, transferItemUseCase, log)
	artifactHandler := httphandlers.NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)
	eventHandler := httphandlers.NewEventHandler(createEventUseCase, getEventUseCase, listEventsUseCase, updateEventUseCase, deleteEventUseCase, addCharacterToEventUseCase, removeCharacterFromEventUseCase, getEventCharactersUseCase, addLocationToEventUseCase, removeLocationFromEventUseCase, getEventLocationsUseCase, addArtifactToEventUseCase, removeArtifactFromEventUseCase, getEventArtifactsUseCase, getEventStatChangesUseCase, log)
	rpgSystemHandler := httphandlers.NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)
	characterRPGStatsHandler := httphandlers.NewCharacterRPGStatsHandler(createCharacterStatsUseCase, getActiveCharacterStatsUseCase, listCharacterStatsHistoryUseCase, activateCharacterStatsVersionUseCase, deleteAllCharacterStatsUseCase, log)
	artifactRPGStatsHandler := httphandlers.NewArtifactRPGStatsHandler(createArtifactStatsUseCase, getActiveArtifactStatsUseCase, listArtifactStatsHistoryUseCase, activateArtifactStatsVersionUseCase, log)
	skillHandler := httphandlers.NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)
	characterSkillHandler := httphandlers.NewCharacterSkillHandler(learnSkillUseCase, listCharacterSkillsUseCase, updateCharacterSkillUseCase, deleteCharacterSkillUseCase, log)
	traitHandler := httphandlers.NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	archetypeHandler := httphandlers.NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitToArchetypeUseCase, removeTraitFromArchetypeUseCase, log)
	storyHandler := httphandlers.NewStoryHandler(createStoryUseCase, cloneStoryUseCase, storyRepo, log)
	chapterHandler := httphandlers.NewChapterHandler(chapterRepo, storyRepo, log)
	sceneHandler := httphandlers.NewSceneHandler(sceneRepo, chapterRepo, storyRepo, addSceneReferenceUseCase, removeSceneReferenceUseCase, getSceneReferencesUseCase, log)
	beatHandler := httphandlers.NewBeatHandler(beatRepo, sceneRepo, storyRepo, log)
	proseBlockHandler := httphandlers.NewProseBlockHandler(proseBlockRepo, chapterRepo, log)
	proseBlockReferenceHandler := httphandlers.NewProseBlockReferenceHandler(proseBlockReferenceRepo, proseBlockRepo, log)
	imageBlockHandler := httphandlers.NewImageBlockHandler(createImageBlockUseCase, getImageBlockUseCase, listImageBlocksUseCase, updateImageBlockUseCase, deleteImageBlockUseCase, addImageBlockReferenceUseCase, removeImageBlockReferenceUseCase, getImageBlockReferencesUseCase, log)

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
	mux.HandleFunc("POST /api/v1/events/{id}/characters", eventHandler.AddCharacter)
	mux.HandleFunc("GET /api/v1/events/{id}/characters", eventHandler.GetCharacters)
	mux.HandleFunc("DELETE /api/v1/events/{id}/characters/{character_id}", eventHandler.RemoveCharacter)
	mux.HandleFunc("POST /api/v1/events/{id}/locations", eventHandler.AddLocation)
	mux.HandleFunc("GET /api/v1/events/{id}/locations", eventHandler.GetLocations)
	mux.HandleFunc("DELETE /api/v1/events/{id}/locations/{location_id}", eventHandler.RemoveLocation)
	mux.HandleFunc("POST /api/v1/events/{id}/artifacts", eventHandler.AddArtifact)
	mux.HandleFunc("GET /api/v1/events/{id}/artifacts", eventHandler.GetArtifacts)
	mux.HandleFunc("DELETE /api/v1/events/{id}/artifacts/{artifact_id}", eventHandler.RemoveArtifact)
	mux.HandleFunc("GET /api/v1/events/{id}/stat-changes", eventHandler.GetStatChanges)

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

	mux.HandleFunc("GET /api/v1/chapters/{id}/prose-blocks", proseBlockHandler.ListByChapter)
	mux.HandleFunc("POST /api/v1/chapters/{id}/prose-blocks", proseBlockHandler.Create)
	mux.HandleFunc("GET /api/v1/prose-blocks/{id}", proseBlockHandler.Get)
	mux.HandleFunc("PUT /api/v1/prose-blocks/{id}", proseBlockHandler.Update)
	mux.HandleFunc("DELETE /api/v1/prose-blocks/{id}", proseBlockHandler.Delete)

	mux.HandleFunc("POST /api/v1/prose-blocks/{id}/references", proseBlockReferenceHandler.Create)
	mux.HandleFunc("GET /api/v1/prose-blocks/{id}/references", proseBlockReferenceHandler.ListByProseBlock)
	mux.HandleFunc("GET /api/v1/scenes/{id}/prose-blocks", proseBlockReferenceHandler.ListByScene)
	mux.HandleFunc("GET /api/v1/beats/{id}/prose-blocks", proseBlockReferenceHandler.ListByBeat)
	mux.HandleFunc("DELETE /api/v1/prose-block-references/{id}", proseBlockReferenceHandler.Delete)

	mux.HandleFunc("GET /api/v1/chapters/{id}/image-blocks", imageBlockHandler.List)
	mux.HandleFunc("POST /api/v1/chapters/{id}/image-blocks", imageBlockHandler.Create)
	mux.HandleFunc("GET /api/v1/image-blocks/{id}", imageBlockHandler.Get)
	mux.HandleFunc("PUT /api/v1/image-blocks/{id}", imageBlockHandler.Update)
	mux.HandleFunc("DELETE /api/v1/image-blocks/{id}", imageBlockHandler.Delete)
	mux.HandleFunc("GET /api/v1/image-blocks/{id}/references", imageBlockHandler.GetReferences)
	mux.HandleFunc("POST /api/v1/image-blocks/{id}/references", imageBlockHandler.AddReference)
	mux.HandleFunc("DELETE /api/v1/image-blocks/{id}/references/{entity_type}/{entity_id}", imageBlockHandler.RemoveReference)

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

	mux.HandleFunc("GET /health", httphandlers.HealthCheck)

	// Wrap with middleware
	handler := middleware.Chain(
		mux,
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

