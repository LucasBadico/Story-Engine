package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	scenepb "github.com/story-engine/main-service/proto/scene"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SceneHandler implements the SceneService gRPC service
type SceneHandler struct {
	scenepb.UnimplementedSceneServiceServer
	sceneRepo   repositories.SceneRepository
	chapterRepo repositories.ChapterRepository
	storyRepo   repositories.StoryRepository
	logger      logger.Logger
}

// NewSceneHandler creates a new SceneHandler
func NewSceneHandler(
	sceneRepo repositories.SceneRepository,
	chapterRepo repositories.ChapterRepository,
	storyRepo repositories.StoryRepository,
	logger logger.Logger,
) *SceneHandler {
	return &SceneHandler{
		sceneRepo:   sceneRepo,
		chapterRepo: chapterRepo,
		storyRepo:   storyRepo,
		logger:      logger,
	}
}

// CreateScene creates a new scene
func (h *SceneHandler) CreateScene(ctx context.Context, req *scenepb.CreateSceneRequest) (*scenepb.CreateSceneResponse, error) {
	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	// Validate story exists
	_, err = h.storyRepo.GetByID(ctx, storyID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "story not found: %v", err)
	}

	var chapterID *uuid.UUID
	if req.ChapterId != nil && *req.ChapterId != "" {
		parsedChapterID, err := uuid.Parse(*req.ChapterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
		}
		// Validate chapter exists
		_, err = h.chapterRepo.GetByID(ctx, parsedChapterID)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "chapter not found: %v", err)
		}
		chapterID = &parsedChapterID
	}

	if req.OrderNum < 1 {
		return nil, status.Errorf(codes.InvalidArgument, "order_num must be greater than 0")
	}

	scene, err := story.NewScene(storyID, chapterID, int(req.OrderNum))
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid scene: %v", err)
	}

	if req.Goal != "" {
		scene.UpdateGoal(req.Goal)
	}
	if req.TimeRef != "" {
		scene.TimeRef = req.TimeRef
	}

	if req.PovCharacterId != nil && *req.PovCharacterId != "" {
		charID, err := uuid.Parse(*req.PovCharacterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid pov_character_id: %v", err)
		}
		scene.UpdatePOV(&charID)
	}

	if req.LocationId != nil && *req.LocationId != "" {
		locID, err := uuid.Parse(*req.LocationId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
		}
		scene.UpdateLocation(&locID)
	}

	if err := h.sceneRepo.Create(ctx, scene); err != nil {
		h.logger.Error("failed to create scene", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create scene: %v", err)
	}

	return &scenepb.CreateSceneResponse{
		Scene: sceneToProto(scene),
	}, nil
}

// GetScene retrieves a scene by ID
func (h *SceneHandler) GetScene(ctx context.Context, req *scenepb.GetSceneRequest) (*scenepb.GetSceneResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	scene, err := h.sceneRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "scene not found: %v", err)
	}

	return &scenepb.GetSceneResponse{
		Scene: sceneToProto(scene),
	}, nil
}

// UpdateScene updates an existing scene
func (h *SceneHandler) UpdateScene(ctx context.Context, req *scenepb.UpdateSceneRequest) (*scenepb.UpdateSceneResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	scene, err := h.sceneRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "scene not found: %v", err)
	}

	// Update fields if provided
	if req.OrderNum != nil {
		if *req.OrderNum < 1 {
			return nil, status.Errorf(codes.InvalidArgument, "order_num must be greater than 0")
		}
		scene.OrderNum = int(*req.OrderNum)
	}

	if req.Goal != nil {
		scene.UpdateGoal(*req.Goal)
	}

	if req.TimeRef != nil {
		scene.TimeRef = *req.TimeRef
	}

	if req.PovCharacterId != nil {
		if *req.PovCharacterId == "" {
			scene.UpdatePOV(nil)
		} else {
			charID, err := uuid.Parse(*req.PovCharacterId)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid pov_character_id: %v", err)
			}
			scene.UpdatePOV(&charID)
		}
	}

	if req.LocationId != nil {
		if *req.LocationId == "" {
			scene.UpdateLocation(nil)
		} else {
			locID, err := uuid.Parse(*req.LocationId)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
			}
			scene.UpdateLocation(&locID)
		}
	}

	if err := h.sceneRepo.Update(ctx, scene); err != nil {
		h.logger.Error("failed to update scene", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update scene: %v", err)
	}

	return &scenepb.UpdateSceneResponse{
		Scene: sceneToProto(scene),
	}, nil
}

// DeleteScene deletes a scene
func (h *SceneHandler) DeleteScene(ctx context.Context, req *scenepb.DeleteSceneRequest) (*scenepb.DeleteSceneResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	// Check if scene exists
	_, err = h.sceneRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "scene not found: %v", err)
	}

	if err := h.sceneRepo.Delete(ctx, id); err != nil {
		h.logger.Error("failed to delete scene", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to delete scene: %v", err)
	}

	return &scenepb.DeleteSceneResponse{
		Success: true,
	}, nil
}

// MoveScene moves a scene to a different chapter
func (h *SceneHandler) MoveScene(ctx context.Context, req *scenepb.MoveSceneRequest) (*scenepb.MoveSceneResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	scene, err := h.sceneRepo.GetByID(ctx, id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "scene not found: %v", err)
	}

	var newChapterID *uuid.UUID
	if req.ChapterId != nil && *req.ChapterId != "" {
		parsedChapterID, err := uuid.Parse(*req.ChapterId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
		}
		// Validate chapter exists
		_, err = h.chapterRepo.GetByID(ctx, parsedChapterID)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "chapter not found: %v", err)
		}
		newChapterID = &parsedChapterID
	}

	scene.UpdateChapter(newChapterID)

	if err := h.sceneRepo.Update(ctx, scene); err != nil {
		h.logger.Error("failed to move scene", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to move scene: %v", err)
	}

	return &scenepb.MoveSceneResponse{
		Scene: sceneToProto(scene),
	}, nil
}

// ListScenesByChapter lists scenes for a specific chapter
func (h *SceneHandler) ListScenesByChapter(ctx context.Context, req *scenepb.ListScenesByChapterRequest) (*scenepb.ListScenesByChapterResponse, error) {
	chapterID, err := uuid.Parse(req.ChapterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid chapter_id: %v", err)
	}

	scenes, err := h.sceneRepo.ListByChapter(ctx, chapterID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list scenes: %v", err)
	}

	protoScenes := make([]*scenepb.Scene, len(scenes))
	for i, s := range scenes {
		protoScenes[i] = sceneToProto(s)
	}

	return &scenepb.ListScenesByChapterResponse{
		Scenes:     protoScenes,
		TotalCount: int32(len(scenes)),
	}, nil
}

// ListScenesByStory lists all scenes for a story
func (h *SceneHandler) ListScenesByStory(ctx context.Context, req *scenepb.ListScenesByStoryRequest) (*scenepb.ListScenesByStoryResponse, error) {
	storyID, err := uuid.Parse(req.StoryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid story_id: %v", err)
	}

	scenes, err := h.sceneRepo.ListByStory(ctx, storyID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list scenes: %v", err)
	}

	protoScenes := make([]*scenepb.Scene, len(scenes))
	for i, s := range scenes {
		protoScenes[i] = sceneToProto(s)
	}

	return &scenepb.ListScenesByStoryResponse{
		Scenes:     protoScenes,
		TotalCount: int32(len(scenes)),
	}, nil
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
	if s.LocationID != nil {
		locStr := s.LocationID.String()
		pb.LocationId = &locStr
	}

	return pb
}

