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
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	locationapp "github.com/story-engine/main-service/internal/application/world/location"
	traitapp "github.com/story-engine/main-service/internal/application/world/trait"
	"github.com/story-engine/main-service/internal/application/story"
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
	storyRepo := postgres.NewStoryRepository(pgDB)
	chapterRepo := postgres.NewChapterRepository(pgDB)
	sceneRepo := postgres.NewSceneRepository(pgDB)
	beatRepo := postgres.NewBeatRepository(pgDB)
	proseBlockRepo := postgres.NewProseBlockRepository(pgDB)
	proseBlockReferenceRepo := postgres.NewProseBlockReferenceRepository(pgDB)
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

	// Create handlers
	tenantHandler := httphandlers.NewTenantHandler(createTenantUseCase, tenantRepo, log)
	worldHandler := httphandlers.NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	locationHandler := httphandlers.NewLocationHandler(createLocationUseCase, getLocationUseCase, listLocationsUseCase, updateLocationUseCase, deleteLocationUseCase, getChildrenUseCase, getAncestorsUseCase, getDescendantsUseCase, moveLocationUseCase, log)
	characterHandler := httphandlers.NewCharacterHandler(createCharacterUseCase, getCharacterUseCase, listCharactersUseCase, updateCharacterUseCase, deleteCharacterUseCase, addTraitToCharacterUseCase, removeTraitFromCharacterUseCase, updateCharacterTraitUseCase, getCharacterTraitsUseCase, log)
	traitHandler := httphandlers.NewTraitHandler(createTraitUseCase, getTraitUseCase, listTraitsUseCase, updateTraitUseCase, deleteTraitUseCase, log)
	archetypeHandler := httphandlers.NewArchetypeHandler(createArchetypeUseCase, getArchetypeUseCase, listArchetypesUseCase, updateArchetypeUseCase, deleteArchetypeUseCase, addTraitToArchetypeUseCase, removeTraitFromArchetypeUseCase, log)
	storyHandler := httphandlers.NewStoryHandler(createStoryUseCase, cloneStoryUseCase, storyRepo, log)
	chapterHandler := httphandlers.NewChapterHandler(chapterRepo, storyRepo, log)
	sceneHandler := httphandlers.NewSceneHandler(sceneRepo, chapterRepo, storyRepo, log)
	beatHandler := httphandlers.NewBeatHandler(beatRepo, sceneRepo, storyRepo, log)
	proseBlockHandler := httphandlers.NewProseBlockHandler(proseBlockRepo, chapterRepo, log)
	proseBlockReferenceHandler := httphandlers.NewProseBlockReferenceHandler(proseBlockReferenceRepo, proseBlockRepo, log)

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

