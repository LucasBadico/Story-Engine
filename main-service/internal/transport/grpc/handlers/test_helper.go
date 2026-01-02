//go:build integration

package handlers

import (
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/story"
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	beatapp "github.com/story-engine/main-service/internal/application/story/beat"
	proseblockapp "github.com/story-engine/main-service/internal/application/story/prose_block"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
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
	worldRepo := postgres.NewWorldRepository(db)
	sceneReferenceRepo := postgres.NewSceneReferenceRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)

	// Initialize use cases
	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
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
	addSceneReferenceUC := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, sceneReferenceRepo, characterRepo, locationRepo, artifactRepo, log)
	removeSceneReferenceUC := sceneapp.NewRemoveSceneReferenceUseCase(sceneReferenceRepo, log)
	getSceneReferencesUC := sceneapp.NewGetSceneReferencesUseCase(sceneReferenceRepo, log)

	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)

	createProseBlockUseCase := proseblockapp.NewCreateProseBlockUseCase(proseBlockRepo, chapterRepo, log)
	getProseBlockUseCase := proseblockapp.NewGetProseBlockUseCase(proseBlockRepo, log)
	updateProseBlockUseCase := proseblockapp.NewUpdateProseBlockUseCase(proseBlockRepo, log)
	deleteProseBlockUseCase := proseblockapp.NewDeleteProseBlockUseCase(proseBlockRepo, log)
	listProseBlocksUseCase := proseblockapp.NewListProseBlocksUseCase(proseBlockRepo, log)

	createProseBlockReferenceUseCase := proseblockapp.NewCreateProseBlockReferenceUseCase(proseBlockRefRepo, proseBlockRepo, log)
	listProseBlockReferencesByProseBlockUseCase := proseblockapp.NewListProseBlockReferencesByProseBlockUseCase(proseBlockRefRepo, proseBlockRepo, log)
	listProseBlocksByEntityUseCase := proseblockapp.NewListProseBlocksByEntityUseCase(proseBlockRefRepo, proseBlockRepo, log)
	deleteProseBlockReferenceUseCase := proseblockapp.NewDeleteProseBlockReferenceUseCase(proseBlockRefRepo, log)

	// Create handlers
	tenantHandler := NewTenantHandler(createTenantUseCase, tenantRepo, log)
	storyHandler := NewStoryHandler(
		createStoryUseCase,
		getStoryUseCase,
		updateStoryUseCase,
		listStoriesUseCase,
		cloneStoryUseCase,
		versionGraphUseCase,
		log,
	)
	chapterHandler := NewChapterHandler(
		createChapterUseCase,
		getChapterUseCase,
		updateChapterUseCase,
		deleteChapterUseCase,
		listChaptersUseCase,
		log,
	)
	sceneHandler := NewSceneHandler(
		createSceneUseCase,
		getSceneUseCase,
		updateSceneUseCase,
		deleteSceneUseCase,
		listScenesUseCase,
		moveSceneUseCase,
		addSceneReferenceUC,
		removeSceneReferenceUC,
		getSceneReferencesUC,
		log,
	)
	beatHandler := NewBeatHandler(
		createBeatUseCase,
		getBeatUseCase,
		updateBeatUseCase,
		deleteBeatUseCase,
		listBeatsUseCase,
		moveBeatUseCase,
		log,
	)
	proseBlockHandler := NewProseBlockHandler(
		createProseBlockUseCase,
		getProseBlockUseCase,
		updateProseBlockUseCase,
		deleteProseBlockUseCase,
		listProseBlocksUseCase,
		log,
	)
	proseBlockRefHandler := NewProseBlockReferenceHandler(
		createProseBlockReferenceUseCase,
		listProseBlockReferencesByProseBlockUseCase,
		listProseBlocksByEntityUseCase,
		deleteProseBlockReferenceUseCase,
		log,
	)

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
