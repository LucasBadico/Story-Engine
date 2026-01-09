//go:build integration

package handlers

import (
	"testing"

	"github.com/story-engine/main-service/internal/adapters/db/postgres"
	"github.com/story-engine/main-service/internal/application/story"
	chapterapp "github.com/story-engine/main-service/internal/application/story/chapter"
	beatapp "github.com/story-engine/main-service/internal/application/story/beat"
	contentblockapp "github.com/story-engine/main-service/internal/application/story/content_block"
	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
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
	contentBlockRepo := postgres.NewContentBlockRepository(db)
	contentBlockRefRepo := postgres.NewContentBlockReferenceRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)

	// Initialize use cases
	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	createStoryUseCase := story.NewCreateStoryUseCase(storyRepo, tenantRepo, worldRepo, createWorldUseCase, auditLogRepo, nil, log)
	getStoryUseCase := story.NewGetStoryUseCase(storyRepo, log)
	updateStoryUseCase := story.NewUpdateStoryUseCase(storyRepo, nil, log)
	listStoriesUseCase := story.NewListStoriesUseCase(storyRepo, log)
	cloneStoryUseCase := story.NewCloneStoryUseCase(
		storyRepo,
		chapterRepo,
		sceneRepo,
		beatRepo,
		contentBlockRepo,
		auditLogRepo,
		transactionRepo,
		log,
	)
	versionGraphUseCase := story.NewGetStoryVersionGraphUseCase(storyRepo, log)

	createChapterUseCase := chapterapp.NewCreateChapterUseCase(chapterRepo, storyRepo, nil, log)
	getChapterUseCase := chapterapp.NewGetChapterUseCase(chapterRepo, log)
	updateChapterUseCase := chapterapp.NewUpdateChapterUseCase(chapterRepo, nil, log)
	deleteChapterUseCase := chapterapp.NewDeleteChapterUseCase(chapterRepo, log)
	listChaptersUseCase := chapterapp.NewListChaptersUseCase(chapterRepo, log)

	createSceneUseCase := sceneapp.NewCreateSceneUseCase(sceneRepo, chapterRepo, storyRepo, nil, log)
	getSceneUseCase := sceneapp.NewGetSceneUseCase(sceneRepo, log)
	updateSceneUseCase := sceneapp.NewUpdateSceneUseCase(sceneRepo, nil, log)
	deleteSceneUseCase := sceneapp.NewDeleteSceneUseCase(sceneRepo, entityRelationRepo, log)
	listScenesUseCase := sceneapp.NewListScenesUseCase(sceneRepo, log)
	moveSceneUseCase := sceneapp.NewMoveSceneUseCase(sceneRepo, chapterRepo, log)
	// Entity relations use cases for scene references
	summaryGenerator := relationapp.NewSummaryGenerator()
	createRelationUseCase := relationapp.NewCreateRelationUseCase(entityRelationRepo, summaryGenerator, nil, log)
	listRelationsBySourceUseCase := relationapp.NewListRelationsBySourceUseCase(entityRelationRepo, log)
	deleteRelationUseCase := relationapp.NewDeleteRelationUseCase(entityRelationRepo, log)
	addSceneReferenceUC := sceneapp.NewAddSceneReferenceUseCase(sceneRepo, storyRepo, createRelationUseCase, listRelationsBySourceUseCase, characterRepo, locationRepo, artifactRepo, log)
	removeSceneReferenceUC := sceneapp.NewRemoveSceneReferenceUseCase(listRelationsBySourceUseCase, deleteRelationUseCase, log)
	getSceneReferencesUC := sceneapp.NewGetSceneReferencesUseCase(listRelationsBySourceUseCase, log)

	createBeatUseCase := beatapp.NewCreateBeatUseCase(beatRepo, sceneRepo, nil, log)
	getBeatUseCase := beatapp.NewGetBeatUseCase(beatRepo, log)
	updateBeatUseCase := beatapp.NewUpdateBeatUseCase(beatRepo, nil, log)
	deleteBeatUseCase := beatapp.NewDeleteBeatUseCase(beatRepo, log)
	listBeatsUseCase := beatapp.NewListBeatsUseCase(beatRepo, log)
	moveBeatUseCase := beatapp.NewMoveBeatUseCase(beatRepo, sceneRepo, log)

	createContentBlockUseCase := contentblockapp.NewCreateContentBlockUseCase(contentBlockRepo, chapterRepo, nil, log)
	getContentBlockUseCase := contentblockapp.NewGetContentBlockUseCase(contentBlockRepo, log)
	updateContentBlockUseCase := contentblockapp.NewUpdateContentBlockUseCase(contentBlockRepo, nil, log)
	deleteContentBlockUseCase := contentblockapp.NewDeleteContentBlockUseCase(contentBlockRepo, log)
	listContentBlocksUseCase := contentblockapp.NewListContentBlocksUseCase(contentBlockRepo, log)

	createContentBlockReferenceUseCase := contentblockapp.NewCreateContentBlockReferenceUseCase(contentBlockRefRepo, contentBlockRepo, log)
	listContentBlockReferencesByContentBlockUseCase := contentblockapp.NewListContentBlockReferencesByContentBlockUseCase(contentBlockRefRepo, contentBlockRepo, log)
	listContentBlocksByEntityUseCase := contentblockapp.NewListContentBlocksByEntityUseCase(contentBlockRefRepo, contentBlockRepo, log)
	deleteContentBlockReferenceUseCase := contentblockapp.NewDeleteContentBlockReferenceUseCase(contentBlockRefRepo, log)

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
	contentBlockHandler := NewContentBlockHandler(
		createContentBlockUseCase,
		getContentBlockUseCase,
		updateContentBlockUseCase,
		deleteContentBlockUseCase,
		listContentBlocksUseCase,
		log,
	)
	contentBlockRefHandler := NewContentBlockReferenceHandler(
		createContentBlockReferenceUseCase,
		listContentBlockReferencesByContentBlockUseCase,
		listContentBlocksByEntityUseCase,
		deleteContentBlockReferenceUseCase,
		log,
	)

	// Use the testing package's SetupTestServerWithHandlers with all handlers
	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:                tenantHandler,
		StoryHandler:                 storyHandler,
		ChapterHandler:               chapterHandler,
		SceneHandler:                 sceneHandler,
		BeatHandler:                  beatHandler,
		ContentBlockHandler:          contentBlockHandler,
		ContentBlockReferenceHandler: contentBlockRefHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}
