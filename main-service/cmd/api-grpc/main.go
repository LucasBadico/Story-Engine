package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
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
	imageblockapp "github.com/story-engine/main-service/internal/application/story/image_block"
	proseblockapp "github.com/story-engine/main-service/internal/application/story/prose_block"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/application/tenant"
	worldapp "github.com/story-engine/main-service/internal/application/world"
	archetypeapp "github.com/story-engine/main-service/internal/application/world/archetype"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/database"
	"github.com/story-engine/main-service/internal/platform/logger"
	grpcserver "github.com/story-engine/main-service/internal/transport/grpc"
	"github.com/story-engine/main-service/internal/transport/grpc/handlers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup logger
	log := logger.New()

	log.Info("Starting gRPC server", "port", cfg.GRPC.Port)

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
	storyRepo := postgres.NewStoryRepository(pgDB)
	chapterRepo := postgres.NewChapterRepository(pgDB)
	sceneRepo := postgres.NewSceneRepository(pgDB)
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(pgDB)
	beatRepo := postgres.NewBeatRepository(pgDB)
	proseBlockRepo := postgres.NewProseBlockRepository(pgDB)
	auditLogRepo := postgres.NewAuditLogRepository(pgDB)
	transactionRepo := postgres.NewTransactionRepository(pgDB)
	rpgSystemRepo := postgres.NewRPGSystemRepository(pgDB)
	skillRepo := postgres.NewSkillRepository(pgDB)
	rpgClassRepo := postgres.NewRPGClassRepository(pgDB)
	rpgClassSkillRepo := postgres.NewRPGClassSkillRepository(pgDB)
	characterSkillRepo := postgres.NewCharacterSkillRepository(pgDB)
	characterStatsRepo := postgres.NewCharacterRPGStatsRepository(pgDB)
	artifactStatsRepo := postgres.NewArtifactRPGStatsRepository(pgDB)
	imageBlockRepo := postgres.NewImageBlockRepository(pgDB)
	inventoryRepo := postgres.NewCharacterInventoryRepository(pgDB)
	inventoryItemRepo := postgres.NewInventoryItemRepository(pgDB)

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
	createCharacterUseCase := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUseCase := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUseCase := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUseCase := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUseCase := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, worldRepo, auditLogRepo, log)
	getCharacterTraitsUseCase := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	addTraitToCharacterUseCase := characterapp.NewAddTraitToCharacterUseCase(characterRepo, traitRepo, characterTraitRepo, log)
	updateCharacterTraitUseCase := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, traitRepo, log)
	removeTraitFromCharacterUseCase := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)
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
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, log)
	listStoriesUseCase := story.NewListStoriesUseCase(storyRepo, log)
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
	versionGraphUseCase := story.NewGetStoryVersionGraphUseCase(storyRepo, log)
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
	proseBlockReferenceRepo := postgres.NewProseBlockReferenceRepository(pgDB)
	createProseBlockUseCase := proseblockapp.NewCreateProseBlockUseCase(proseBlockRepo, chapterRepo, log)
	getProseBlockUseCase := proseblockapp.NewGetProseBlockUseCase(proseBlockRepo, log)
	updateProseBlockUseCase := proseblockapp.NewUpdateProseBlockUseCase(proseBlockRepo, log)
	deleteProseBlockUseCase := proseblockapp.NewDeleteProseBlockUseCase(proseBlockRepo, log)
	listProseBlocksUseCase := proseblockapp.NewListProseBlocksUseCase(proseBlockRepo, log)
	createProseBlockReferenceUseCase := proseblockapp.NewCreateProseBlockReferenceUseCase(proseBlockReferenceRepo, proseBlockRepo, log)
	listProseBlockReferencesByProseBlockUseCase := proseblockapp.NewListProseBlockReferencesByProseBlockUseCase(proseBlockReferenceRepo, proseBlockRepo, log)
	listProseBlocksByEntityUseCase := proseblockapp.NewListProseBlocksByEntityUseCase(proseBlockReferenceRepo, proseBlockRepo, log)
	deleteProseBlockReferenceUseCase := proseblockapp.NewDeleteProseBlockReferenceUseCase(proseBlockReferenceRepo, log)
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
	createImageBlockUseCase := imageblockapp.NewCreateImageBlockUseCase(imageBlockRepo, chapterRepo, log)
	getImageBlockUseCase := imageblockapp.NewGetImageBlockUseCase(imageBlockRepo, log)
	listImageBlocksUseCase := imageblockapp.NewListImageBlocksUseCase(imageBlockRepo, log)
	updateImageBlockUseCase := imageblockapp.NewUpdateImageBlockUseCase(imageBlockRepo, log)
	imageBlockReferenceRepo := postgres.NewImageBlockReferenceRepository(pgDB)
	deleteImageBlockUseCase := imageblockapp.NewDeleteImageBlockUseCase(imageBlockRepo, imageBlockReferenceRepo, log)
	addItemToInventoryUseCase := characterinventoryapp.NewAddItemToInventoryUseCase(inventoryRepo, characterRepo, inventoryItemRepo, log)
	updateInventoryItemUseCase := characterinventoryapp.NewUpdateCharacterInventoryUseCase(inventoryRepo, log)
	deleteInventoryItemUseCase := characterinventoryapp.NewDeleteCharacterInventoryUseCase(inventoryRepo, log)
	listInventoryUseCase := characterinventoryapp.NewListCharacterInventoryUseCase(inventoryRepo, log)

	// Create handlers
	tenantHandler := handlers.NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := handlers.NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	locationHandler := handlers.NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)
	characterHandler := handlers.NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, getCharacterTraitsUseCase, addTraitToCharacterUseCase, updateCharacterTraitUseCase, removeTraitFromCharacterUseCase, log)
	artifactHandler := handlers.NewArtifactHandler(createArtifactUseCase, getArtifactUseCase, listArtifactsUseCase, updateArtifactUseCase, deleteArtifactUseCase, getArtifactReferencesUseCase, addArtifactReferenceUseCase, removeArtifactReferenceUseCase, log)
	eventHandler := handlers.NewEventHandler(createEventUseCase, getEventUseCase, listEventsUseCase, updateEventUseCase, deleteEventUseCase, addCharacterToEventUseCase, removeCharacterFromEventUseCase, getEventCharactersUseCase, addLocationToEventUseCase, removeLocationFromEventUseCase, getEventLocationsUseCase, addArtifactToEventUseCase, removeArtifactFromEventUseCase, getEventArtifactsUseCase, log)
	traitHandler := handlers.NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	archetypeHandler := handlers.NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitToArchetypeUseCase, removeTraitFromArchetypeUseCase, log)
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
	proseBlockHandler := handlers.NewProseBlockHandler(createProseBlockUseCase, getProseBlockUseCase, updateProseBlockUseCase, deleteProseBlockUseCase, listProseBlocksUseCase, log)
	proseBlockReferenceHandler := handlers.NewProseBlockReferenceHandler(createProseBlockReferenceUseCase, listProseBlockReferencesByProseBlockUseCase, listProseBlocksByEntityUseCase, deleteProseBlockReferenceUseCase, log)
	rpgSystemHandler := handlers.NewRPGSystemHandler(createRPGSystemUseCase, getRPGSystemUseCase, listRPGSystemsUseCase, updateRPGSystemUseCase, deleteRPGSystemUseCase, log)
	skillHandler := handlers.NewSkillHandler(createSkillUseCase, getSkillUseCase, listSkillsUseCase, updateSkillUseCase, deleteSkillUseCase, log)
	rpgClassHandler := handlers.NewRPGClassHandler(createRPGClassUseCase, getRPGClassUseCase, listRPGClassesUseCase, updateRPGClassUseCase, deleteRPGClassUseCase, addSkillToClassUseCase, listClassSkillsUseCase, removeSkillFromClassUseCase, log)
	characterSkillHandler := handlers.NewCharacterSkillHandler(learnSkillUseCase, updateCharacterSkillUseCase, deleteCharacterSkillUseCase, listCharacterSkillsUseCase, log)
	characterRPGStatsHandler := handlers.NewCharacterRPGStatsHandler(createCharacterStatsUseCase, getActiveCharacterStatsUseCase, listCharacterStatsHistoryUseCase, activateCharacterStatsVersionUseCase, deleteAllCharacterStatsUseCase, log)
	artifactRPGStatsHandler := handlers.NewArtifactRPGStatsHandler(createArtifactStatsUseCase, getActiveArtifactStatsUseCase, listArtifactStatsHistoryUseCase, activateArtifactStatsVersionUseCase, log)
	imageBlockHandler := handlers.NewImageBlockHandler(createImageBlockUseCase, getImageBlockUseCase, listImageBlocksUseCase, updateImageBlockUseCase, deleteImageBlockUseCase, log)
	inventoryHandler := handlers.NewInventoryHandler(addItemToInventoryUseCase, updateInventoryItemUseCase, deleteInventoryItemUseCase, listInventoryUseCase, log)

	// Create and configure gRPC server
	grpcServer := grpcserver.NewServer(cfg, log)
	grpcServer.RegisterTenantService(tenantHandler)
	grpcServer.RegisterWorldService(worldHandler)
	grpcServer.RegisterLocationService(locationHandler)
	grpcServer.RegisterCharacterService(characterHandler)
	grpcServer.RegisterArtifactService(artifactHandler)
	grpcServer.RegisterEventService(eventHandler)
	grpcServer.RegisterTraitService(traitHandler)
	grpcServer.RegisterArchetypeService(archetypeHandler)
	grpcServer.RegisterStoryService(storyHandler)
	grpcServer.RegisterChapterService(chapterHandler)
	grpcServer.RegisterSceneService(sceneHandler)
	grpcServer.RegisterBeatService(beatHandler)
	grpcServer.RegisterProseBlockService(proseBlockHandler)
	grpcServer.RegisterProseBlockReferenceService(proseBlockReferenceHandler)
	grpcServer.RegisterRPGSystemService(rpgSystemHandler)
	grpcServer.RegisterSkillService(skillHandler)
	grpcServer.RegisterRPGClassService(rpgClassHandler)
	grpcServer.RegisterCharacterSkillService(characterSkillHandler)
	grpcServer.RegisterCharacterRPGStatsService(characterRPGStatsHandler)
	grpcServer.RegisterArtifactRPGStatsService(artifactRPGStatsHandler)
	grpcServer.RegisterImageBlockService(imageBlockHandler)
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
