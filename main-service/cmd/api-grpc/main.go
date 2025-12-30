package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/story"
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
	versionGraphUseCase := story.NewGetStoryVersionGraphUseCase(storyRepo, log)

	// Create handlers
	tenantHandler := handlers.NewTenantHandler(createTenantUseCase, tenantRepo, log)
	storyHandler := handlers.NewStoryHandler(
		createStoryUseCase,
		cloneStoryUseCase,
		versionGraphUseCase,
		storyRepo,
		log,
	)

	// Create and configure gRPC server
	grpcServer := grpcserver.NewServer(cfg, log)
	grpcServer.RegisterTenantService(tenantHandler)
	grpcServer.RegisterStoryService(storyHandler)

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
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

