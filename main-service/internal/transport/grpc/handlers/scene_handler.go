package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/story/scene"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	scenepb "github.com/story-engine/main-service/proto/scene"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SceneHandler implements the SceneService gRPC service
type SceneHandler struct {
	scenepb.UnimplementedSceneServiceServer
	createSceneUseCase    *scene.CreateSceneUseCase
	getSceneUseCase       *scene.GetSceneUseCase
	updateSceneUseCase    *scene.UpdateSceneUseCase
	deleteSceneUseCase    *scene.DeleteSceneUseCase
	listScenesUseCase     *scene.ListScenesUseCase
	moveSceneUseCase      *scene.MoveSceneUseCase
	addReferenceUC        *scene.AddSceneReferenceUseCase
	removeReferenceUC     *scene.RemoveSceneReferenceUseCase
	getReferencesUC       *scene.GetSceneReferencesUseCase
	logger                 logger.Logger
}

// NewSceneHandler creates a new SceneHandler
func NewSceneHandler(
	createSceneUseCase *scene.CreateSceneUseCase,
	getSceneUseCase *scene.GetSceneUseCase,
	updateSceneUseCase *scene.UpdateSceneUseCase,
	deleteSceneUseCase *scene.DeleteSceneUseCase,
	listScenesUseCase *scene.ListScenesUseCase,
	moveSceneUseCase *scene.MoveSceneUseCase,
	addReferenceUC *scene.AddSceneReferenceUseCase,
	removeReferenceUC *scene.RemoveSceneReferenceUseCase,
	getReferencesUC *scene.GetSceneReferencesUseCase,
	logger logger.Logger,
) *SceneHandler {
	return &SceneHandler{
		createSceneUseCase: createSceneUseCase,
		getSceneUseCase:    getSceneUseCase,
		updateSceneUseCase: updateSceneUseCase,
		deleteSceneUseCase: deleteSceneUseCase,
		listScenesUseCase:  listScenesUseCase,
		moveSceneUseCase:   moveSceneUseCase,
		addReferenceUC:     addReferenceUC,
		removeReferenceUC:  removeReferenceUC,
		getReferencesUC:     getReferencesUC,
		logger:             logger,
	}
}

// CreateScene creates a new scene
func (h *SceneHandler) CreateScene(ctx context.Context, req *scenepb.CreateSceneRequest) (*scenepb.CreateSceneResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	var chapterID *uuid.UUID
	if req.ChapterId != nil && *req.ChapterId != "" {
		parsedChapterID, err := uuid.Parse(*req.ChapterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
		}
		chapterID = &parsedChapterID
	}

	var povCharacterID *uuid.UUID
	if req.PovCharacterId != nil && *req.PovCharacterId != "" {
		charID, err := uuid.Parse(*req.PovCharacterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid pov_character_id: %v", err)
		}
		povCharacterID = &charID
	}

	output, err := h.createSceneUseCase.Execute(ctx, scene.CreateSceneInput{
		TenantID:       tenantUUID,
		StoryID:        storyID,
		ChapterID:      chapterID,
		OrderNum:       int(req.OrderNum),
		POVCharacterID: povCharacterID,
		TimeRef:        req.TimeRef,
		Goal:           req.Goal,
	})
	if err != nil {
		return nil, err
	}

	return &scenepb.CreateSceneResponse{
		Scene: sceneToProto(output.Scene),
	}, nil
}

// GetScene retrieves a scene by ID
func (h *SceneHandler) GetScene(ctx context.Context, req *scenepb.GetSceneRequest) (*scenepb.GetSceneResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	output, err := h.getSceneUseCase.Execute(ctx, scene.GetSceneInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &scenepb.GetSceneResponse{
		Scene: sceneToProto(output.Scene),
	}, nil
}

// UpdateScene updates an existing scene
func (h *SceneHandler) UpdateScene(ctx context.Context, req *scenepb.UpdateSceneRequest) (*scenepb.UpdateSceneResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	var orderNum *int
	if req.OrderNum != nil {
		n := int(*req.OrderNum)
		orderNum = &n
	}

	var povCharacterID *uuid.UUID
	if req.PovCharacterId != nil {
		if *req.PovCharacterId == "" {
			povCharacterID = nil
		} else {
			charID, err := uuid.Parse(*req.PovCharacterId)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid pov_character_id: %v", err)
			}
			povCharacterID = &charID
		}
	}

	output, err := h.updateSceneUseCase.Execute(ctx, scene.UpdateSceneInput{
		TenantID:       tenantUUID,
		ID:             id,
		OrderNum:       orderNum,
		POVCharacterID: povCharacterID,
		TimeRef:        req.TimeRef,
		Goal:           req.Goal,
	})
	if err != nil {
		return nil, err
	}

	return &scenepb.UpdateSceneResponse{
		Scene: sceneToProto(output.Scene),
	}, nil
}

// DeleteScene deletes a scene
func (h *SceneHandler) DeleteScene(ctx context.Context, req *scenepb.DeleteSceneRequest) (*scenepb.DeleteSceneResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	if err := h.deleteSceneUseCase.Execute(ctx, scene.DeleteSceneInput{
		TenantID: tenantUUID,
		ID:       id,
	}); err != nil {
		return nil, err
	}

	return &scenepb.DeleteSceneResponse{
		Success: true,
	}, nil
}

// MoveScene moves a scene to a different chapter
func (h *SceneHandler) MoveScene(ctx context.Context, req *scenepb.MoveSceneRequest) (*scenepb.MoveSceneResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	var newChapterID *uuid.UUID
	if req.ChapterId != nil && *req.ChapterId != "" {
		parsedChapterID, err := uuid.Parse(*req.ChapterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
		}
		newChapterID = &parsedChapterID
	}

	output, err := h.moveSceneUseCase.Execute(ctx, scene.MoveSceneInput{
		TenantID:     tenantUUID,
		SceneID:      id,
		NewChapterID: newChapterID,
	})
	if err != nil {
		return nil, err
	}

	return &scenepb.MoveSceneResponse{
		Scene: sceneToProto(output.Scene),
	}, nil
}

// ListScenesByChapter lists scenes for a specific chapter
func (h *SceneHandler) ListScenesByChapter(ctx context.Context, req *scenepb.ListScenesByChapterRequest) (*scenepb.ListScenesByChapterResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	chapterID, err := uuid.Parse(req.ChapterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
	}

	output, err := h.listScenesUseCase.Execute(ctx, scene.ListScenesInput{
		TenantID:  tenantUUID,
		ChapterID: &chapterID,
	})
	if err != nil {
		return nil, err
	}

	protoScenes := make([]*scenepb.Scene, len(output.Scenes))
	for i, s := range output.Scenes {
		protoScenes[i] = sceneToProto(s)
	}

	return &scenepb.ListScenesByChapterResponse{
		Scenes:     protoScenes,
		TotalCount: int32(output.Total),
	}, nil
}

// ListScenesByStory lists all scenes for a story
func (h *SceneHandler) ListScenesByStory(ctx context.Context, req *scenepb.ListScenesByStoryRequest) (*scenepb.ListScenesByStoryResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	output, err := h.listScenesUseCase.Execute(ctx, scene.ListScenesInput{
		TenantID: tenantUUID,
		StoryID:  storyID,
	})
	if err != nil {
		return nil, err
	}

	protoScenes := make([]*scenepb.Scene, len(output.Scenes))
	for i, s := range output.Scenes {
		protoScenes[i] = sceneToProto(s)
	}

	return &scenepb.ListScenesByStoryResponse{
		Scenes:     protoScenes,
		TotalCount: int32(output.Total),
	}, nil
}

// GetSceneReferences retrieves all references for a scene
func (h *SceneHandler) GetSceneReferences(ctx context.Context, req *scenepb.GetSceneReferencesRequest) (*scenepb.GetSceneReferencesResponse, error) {
	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene_id: %v", err)
	}

	output, err := h.getReferencesUC.Execute(ctx, scene.GetSceneReferencesInput{
		TenantID: tenantUUID,
		SceneID:  sceneID,
	})
	if err != nil {
		return nil, err
	}

	references := make([]*scenepb.SceneReference, len(output.References))
	for i, ref := range output.References {
		references[i] = mappers.SceneReferenceDTOToProto(ref)
	}

	return &scenepb.GetSceneReferencesResponse{
		References: references,
		TotalCount: int32(len(references)),
	}, nil
}

// AddSceneReference adds a reference to a scene
func (h *SceneHandler) AddSceneReference(ctx context.Context, req *scenepb.AddSceneReferenceRequest) (*scenepb.AddSceneReferenceResponse, error) {
	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	entityType := req.EntityType
	if entityType != "character" && entityType != "location" && entityType != "artifact" {
		return nil, status.Errorf(codes.InvalidArgument, "entity_type must be 'character', 'location', or 'artifact'")
	}

	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	err = h.addReferenceUC.Execute(ctx, scene.AddSceneReferenceInput{
		TenantID:   tenantUUID,
		SceneID:    sceneID,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the created reference to return it
	output, err := h.getReferencesUC.Execute(ctx, scene.GetSceneReferencesInput{
		TenantID: tenantUUID,
		SceneID:  sceneID,
	})
	if err != nil {
		return nil, err
	}

	// Find the reference we just created
	var createdRef *scene.SceneReferenceDTO
	for _, ref := range output.References {
		if ref.EntityType == entityType && ref.EntityID == entityID {
			createdRef = ref
			break
		}
	}

	if createdRef == nil {
		return nil, status.Errorf(codes.Internal, "failed to retrieve created reference")
	}

	return &scenepb.AddSceneReferenceResponse{
		Reference: mappers.SceneReferenceDTOToProto(createdRef),
	}, nil
}

// RemoveSceneReference removes a reference from a scene
func (h *SceneHandler) RemoveSceneReference(ctx context.Context, req *scenepb.RemoveSceneReferenceRequest) (*scenepb.RemoveSceneReferenceResponse, error) {
	sceneID, err := uuid.Parse(req.SceneId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	entityType := req.EntityType
	if entityType != "character" && entityType != "location" && entityType != "artifact" {
		return nil, status.Errorf(codes.InvalidArgument, "entity_type must be 'character', 'location', or 'artifact'")
	}

	// Extract tenant_id from context (set by auth interceptor)
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	err = h.removeReferenceUC.Execute(ctx, scene.RemoveSceneReferenceInput{
		TenantID:   tenantUUID,
		SceneID:    sceneID,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	return &scenepb.RemoveSceneReferenceResponse{}, nil
}

// sceneToProto converts a domain Scene to a proto Scene message
func sceneToProto(s *story.Scene) *scenepb.Scene {
	pb := &scenepb.Scene{
		Id:        s.ID.String(),
		StoryId:   s.StoryID.String(),
		OrderNum:  int32(s.OrderNum),
		Goal:      s.Goal,
		TimeRef:   s.TimeRef,
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
	}

	if s.ChapterID != nil {
		chapterIDStr := s.ChapterID.String()
		pb.ChapterId = &chapterIDStr
	}
	if s.POVCharacterID != nil {
		povStr := s.POVCharacterID.String()
		pb.PovCharacterId = &povStr
	}

	return pb
}

