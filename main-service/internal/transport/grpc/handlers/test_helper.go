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
	plotapp "github.com/story-engine/main-service/internal/application/story/plot"
	timelineeventapp "github.com/story-engine/main-service/internal/application/story/timeline_event"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/application/world"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	characterstateapp "github.com/story-engine/main-service/internal/application/world/character_state"
	relationshipstateapp "github.com/story-engine/main-service/internal/application/world/relationship_state"
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
	contentAnchorRepo := postgres.NewContentAnchorRepository(db)
	auditLogRepo := postgres.NewAuditLogRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db)
	worldRepo := postgres.NewWorldRepository(db)
	entityRelationRepo := postgres.NewEntityRelationRepository(db)
	characterRepo := postgres.NewCharacterRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	artifactRepo := postgres.NewArtifactRepository(db)
	characterStateRepo := postgres.NewCharacterStateRepository(db)
	relationshipStateRepo := postgres.NewRelationshipStateRepository(db)
	timelineEventRepo := postgres.NewTimelineEventRepository(db)
	plotRepo := postgres.NewPlotRepository(db)
	plotDevelopmentRepo := postgres.NewPlotDevelopmentRepository(db)

	// Initialize use cases
	log := logger.New()
	createTenantUseCase := tenant.NewCreateTenantUseCase(tenantRepo, auditLogRepo, log)
	createWorldUseCase := world.NewCreateWorldUseCase(worldRepo, tenantRepo, auditLogRepo, log)
	getWorldUseCase := world.NewGetWorldUseCase(worldRepo, log)
	listWorldsUseCase := world.NewListWorldsUseCase(worldRepo, log)
	updateWorldUseCase := world.NewUpdateWorldUseCase(worldRepo, auditLogRepo, log)
	deleteWorldUseCase := world.NewDeleteWorldUseCase(worldRepo, auditLogRepo, log)
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

	// Character use cases (needed for character_state tests)
	archetypeRepo := postgres.NewArchetypeRepository(db)
	characterTraitRepo := postgres.NewCharacterTraitRepository(db)
	createCharacterUC := characterapp.NewCreateCharacterUseCase(characterRepo, worldRepo, archetypeRepo, auditLogRepo, log)
	getCharacterUC := characterapp.NewGetCharacterUseCase(characterRepo, log)
	listCharactersUC := characterapp.NewListCharactersUseCase(characterRepo, log)
	updateCharacterUC := characterapp.NewUpdateCharacterUseCase(characterRepo, archetypeRepo, worldRepo, auditLogRepo, log)
	deleteCharacterUC := characterapp.NewDeleteCharacterUseCase(characterRepo, characterTraitRepo, entityRelationRepo, worldRepo, auditLogRepo, log)
	getCharacterTraitsUC := characterapp.NewGetCharacterTraitsUseCase(characterTraitRepo, log)
	listRelationsByTargetUseCase := relationapp.NewListRelationsByTargetUseCase(entityRelationRepo, log)
	getCharacterEventsUC := characterapp.NewGetCharacterEventsUseCase(listRelationsByTargetUseCase, log)
	addTraitToCharacterUC := characterapp.NewAddTraitToCharacterUseCase(characterRepo, nil, characterTraitRepo, log)
	updateCharacterTraitUC := characterapp.NewUpdateCharacterTraitUseCase(characterTraitRepo, nil, log)
	removeTraitFromCharacterUC := characterapp.NewRemoveTraitFromCharacterUseCase(characterTraitRepo, log)

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

	createContentAnchorUseCase := contentblockapp.NewCreateContentAnchorUseCase(contentAnchorRepo, contentBlockRepo, log)
	listContentAnchorsByContentBlockUseCase := contentblockapp.NewListContentAnchorsByContentBlockUseCase(contentAnchorRepo, contentBlockRepo, log)
	listContentBlocksByEntityUseCase := contentblockapp.NewListContentBlocksByEntityUseCase(contentAnchorRepo, contentBlockRepo, log)
	deleteContentAnchorUseCase := contentblockapp.NewDeleteContentAnchorUseCase(contentAnchorRepo, log)

	// Character state use cases
	createCharacterStateUC := characterstateapp.NewCreateCharacterStateUseCase(characterStateRepo, characterRepo, auditLogRepo, log)
	getCharacterStateUC := characterstateapp.NewGetCharacterStateUseCase(characterStateRepo, log)
	getLatestCharacterStateUC := characterstateapp.NewGetLatestCharacterStateUseCase(characterStateRepo, log)
	listCharacterStatesUC := characterstateapp.NewListCharacterStatesUseCase(characterStateRepo, log)
	deleteCharacterStateUC := characterstateapp.NewDeleteCharacterStateUseCase(characterStateRepo, auditLogRepo, log)

	// Relationship state use cases
	createRelationshipStateUC := relationshipstateapp.NewCreateRelationshipStateUseCase(relationshipStateRepo, auditLogRepo, log)
	getRelationshipStateUC := relationshipstateapp.NewGetRelationshipStateUseCase(relationshipStateRepo, log)
	listRelationshipStatesUC := relationshipstateapp.NewListRelationshipStatesUseCase(relationshipStateRepo, log)
	deleteRelationshipStateUC := relationshipstateapp.NewDeleteRelationshipStateUseCase(relationshipStateRepo, auditLogRepo, log)

	// Timeline event use cases
	createTimelineEventUC := timelineeventapp.NewCreateTimelineEventUseCase(timelineEventRepo, storyRepo, auditLogRepo, log)
	getTimelineEventUC := timelineeventapp.NewGetTimelineEventUseCase(timelineEventRepo, log)
	listTimelineEventsByStoryUC := timelineeventapp.NewListTimelineEventsByStoryUseCase(timelineEventRepo, log)
	listTimelineEventsByChapterUC := timelineeventapp.NewListTimelineEventsByChapterUseCase(timelineEventRepo, log)
	deleteTimelineEventUC := timelineeventapp.NewDeleteTimelineEventUseCase(timelineEventRepo, auditLogRepo, log)

	// Plot use cases
	createPlotUC := plotapp.NewCreatePlotUseCase(plotRepo, storyRepo, auditLogRepo, log)
	getPlotUC := plotapp.NewGetPlotUseCase(plotRepo, log)
	listPlotsUC := plotapp.NewListPlotsUseCase(plotRepo, log)
	updatePlotUC := plotapp.NewUpdatePlotUseCase(plotRepo, auditLogRepo, log)
	deletePlotUC := plotapp.NewDeletePlotUseCase(plotRepo, auditLogRepo, log)
	addPlotEntityUC := plotapp.NewAddPlotEntityUseCase(plotRepo, log)
	createPlotDevelopmentUC := plotapp.NewCreatePlotDevelopmentUseCase(plotDevelopmentRepo, plotRepo, auditLogRepo, log)
	listPlotDevelopmentsUC := plotapp.NewListPlotDevelopmentsUseCase(plotDevelopmentRepo, log)

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
	contentAnchorHandler := NewContentAnchorHandler(
		createContentAnchorUseCase,
		listContentAnchorsByContentBlockUseCase,
		listContentBlocksByEntityUseCase,
		deleteContentAnchorUseCase,
		log,
	)
	worldHandler := NewWorldHandler(createWorldUseCase, getWorldUseCase, listWorldsUseCase, updateWorldUseCase, deleteWorldUseCase, log)
	characterHandler := NewCharacterHandler(
		createCharacterUC, getCharacterUC, listCharactersUC, updateCharacterUC, deleteCharacterUC,
		getCharacterTraitsUC, getCharacterEventsUC,
		addTraitToCharacterUC, updateCharacterTraitUC, removeTraitFromCharacterUC,
		nil, nil, nil, nil, nil, // Character relationship use cases - not needed for tests
		log,
	)
	characterStateHandler := NewCharacterStateHandler(createCharacterStateUC, getCharacterStateUC, getLatestCharacterStateUC, listCharacterStatesUC, deleteCharacterStateUC, log)
	relationshipStateHandler := NewRelationshipStateHandler(createRelationshipStateUC, getRelationshipStateUC, listRelationshipStatesUC, deleteRelationshipStateUC, log)
	timelineEventHandler := NewTimelineEventHandler(createTimelineEventUC, getTimelineEventUC, listTimelineEventsByStoryUC, listTimelineEventsByChapterUC, deleteTimelineEventUC, log)
	plotHandler := NewPlotHandler(createPlotUC, getPlotUC, listPlotsUC, updatePlotUC, deletePlotUC, addPlotEntityUC, createPlotDevelopmentUC, listPlotDevelopmentsUC, log)

	// Use the testing package's SetupTestServerWithHandlers with all handlers
	conn, cleanupServer := grpctesting.SetupTestServerWithHandlers(t, grpctesting.TestHandlers{
		TenantHandler:              tenantHandler,
		StoryHandler:               storyHandler,
		ChapterHandler:             chapterHandler,
		SceneHandler:               sceneHandler,
		BeatHandler:                beatHandler,
		ContentBlockHandler:        contentBlockHandler,
		ContentAnchorHandler:       contentAnchorHandler,
		WorldHandler:               worldHandler,
		CharacterHandler:           characterHandler,
		CharacterStateHandler:      characterStateHandler,
		RelationshipStateHandler:   relationshipStateHandler,
		TimelineEventHandler:       timelineEventHandler,
		PlotHandler:                plotHandler,
	})

	cleanup := func() {
		cleanupServer()
		cleanupDB()
	}

	return conn, cleanup
}
