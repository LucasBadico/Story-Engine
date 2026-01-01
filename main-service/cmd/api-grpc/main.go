package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	worldapp "github.com/story-engine/main-service/internal/application/world"
	archetypeapp "github.com/story-engine/main-service/internal/application/world/archetype"
	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	"github.com/story-engine/main-service/internal/application/story"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/application/tenant"
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
	addSceneReferenceUseCase := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeSceneReferenceUseCase := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getSceneReferencesUseCase := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)

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
		cloneStoryUseCase,
		versionGraphUseCase,
		storyRepo,
		log,
	)
	sceneHandler := handlers.NewSceneHandler(sceneRepo, chapterRepo, storyRepo, addSceneReferenceUseCase, removeSceneReferenceUseCase, getSceneReferencesUseCase, log)

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
	grpcServer.RegisterSceneService(sceneHandler)

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

