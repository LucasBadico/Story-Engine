//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/story"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/platform/logger"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
)

// setupTestServer creates a test gRPC server with all handlers initialized
// Returns a client connection and cleanup function
func setupTestServer(t *testing.T) (*grpc.ClientConn, func()) {
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
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	storyHandler := NewStoryHandler(
		createStoryUseCase,
		cloneStoryUseCase,
		versionGraphUseCase,
		storyRepo,
		log,
	)

	// Use the testing package's SetupTestServer with our handlers
	conn, cleanupServer := grpctesting.SetupTestServer(t, tenantHandler, storyHandler)

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}
