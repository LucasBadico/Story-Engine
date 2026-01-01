//go:build integration

package handlers

import (
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
	// Setup test database (each test gets a fresh cloned database)
	db, cleanupDB := postgres.SetupTestDB(t)

	// Initialize repositories
	tenantRepo := postgres.NewTenantRepository(db)
	storyRepo := postgres.NewStoryRepository(db)
	chapterRepo := postgres.NewChapterRepository(db)
	sceneRepo := postgres.NewSceneRepository(db)
	beatRepo := postgres.NewBeatRepository(db)
	proseBlockRepo := postgres.NewProseBlockRepository(db)
	proseBlockRefRepo := postgres.NewProseBlockReferenceRepository(db)
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
	chapterHandler := NewChapterHandler(chapterRepo, storyRepo, log)
	sceneHandler := NewSceneHandler(sceneRepo, chapterRepo, storyRepo, log)
	beatHandler := NewBeatHandler(beatRepo, sceneRepo, storyRepo, log)
	proseBlockHandler := NewProseBlockHandler(proseBlockRepo, chapterRepo, log)
	proseBlockRefHandler := NewProseBlockReferenceHandler(proseBlockRefRepo, proseBlockRepo, log)

	// Use the testing package's SetupTestServerWithHandlers with all handlers
	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:              tenantHandler,
		StoryHandler:               storyHandler,
		ChapterHandler:             chapterHandler,
		SceneHandler:               sceneHandler,
		BeatHandler:                beatHandler,
		ProseBlockHandler:          proseBlockHandler,
		ProseBlockReferenceHandler: proseBlockRefHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}
