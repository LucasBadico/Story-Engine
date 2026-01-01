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
	storyRepo := postgres.NewStoryRepository(pgDB)
	chapterRepo := postgres.NewChapterRepository(pgDB)
	sceneRepo := postgres.NewSceneRepository(pgDB)
	beatRepo := postgres.NewBeatRepository(pgDB)
	proseBlockRepo := postgres.NewProseBlockRepository(pgDB)
	auditLogRepo := postgres.NewAuditLogRepository(pgDB)
	transactionRepo := postgres.NewTransactionRepository(pgDB)

	// Initialize use cases
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, auditLogRepo, log)
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
	storyHandler := httphandlers.NewStoryHandler(createStoryUseCase, cloneStoryUseCase, storyRepo, log)
	chapterHandler := httphandlers.NewChapterHandler(chapterRepo, storyRepo, log)
	sceneHandler := httphandlers.NewSceneHandler(sceneRepo, chapterRepo, storyRepo, log)
	beatHandler := httphandlers.NewBeatHandler(beatRepo, sceneRepo, storyRepo, log)

	// Create router
	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("POST /api/v1/tenants", tenantHandler.Create)
	mux.HandleFunc("GET /api/v1/tenants/{id}", tenantHandler.Get)

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

