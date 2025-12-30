//go:build integration

package testing

import (
	"context"
	"net"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/story"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/database"
	"github.com/story-engine/main-service/internal/platform/logger"
	grpcserver "github.com/story-engine/main-service/internal/transport/grpc"
	"github.com/story-engine/main-service/internal/transport/grpc/handlers"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SetupTestServer creates a test gRPC server with test database
// Returns a client connection and cleanup function
func SetupTestServer(t *testing.T) (*grpc.ClientConn, func()) {
	// Setup test database
	db, cleanupDB := postgres.SetupTestDB(t)

	// Clean up tables
	ctx := context.Background()
	if err := postgres.TruncateTables(ctx, db); err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
	}

	// Initialize repositories
	tenantRepo := postgres.NewTenantRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	beatRepo := postgres.NewBeatRepository(db)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)

	// Initialize use cases
	log := logger.New()
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

	// Create gRPC server
	cfg := config.Load()
	grpcServer := grpcserver.NewServer(cfg, log)
	grpcServer.RegisterTenantService(tenantHandler)
	grpcServer.RegisterStoryService(storyHandler)

	// Start server on random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go func() {
		if err := grpcServer.GetServer().Serve(listener); err != nil {
			t.Logf("server error: %v", err)
		}
	}()

	// Create client connection
	conn, err := grpc.Dial(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	cleanup := func() {
		conn.Close()
		grpcServer.Stop()
		cleanupDB()
	}

	return conn, cleanup
}

