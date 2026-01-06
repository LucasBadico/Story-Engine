package grpc

import (
	"testing"
	"time"

	"github.com/google/uuid"
	beatpb "github.com/story-engine/main-service/proto/beat"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	contentblockpb "github.com/story-engine/main-service/proto/content_block"
	scenepb "github.com/story-engine/main-service/proto/scene"
	storypb "github.com/story-engine/main-service/proto/story"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestProtoToStory(t *testing.T) {
	storyID := uuid.New()
	tenantID := uuid.New()
	now := time.Now()

	pbStory := &storypb.Story{
		Id:        storyID.String(),
		TenantId:  tenantID.String(),
		Title:     "Test Story",
		Status:    "draft",
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	result := protoToStory(pbStory)

	if result.ID != storyID {
		t.Errorf("Expected ID %s, got %s", storyID, result.ID)
	}
	if result.TenantID != tenantID {
		t.Errorf("Expected TenantID %s, got %s", tenantID, result.TenantID)
	}
	if result.Title != "Test Story" {
		t.Errorf("Expected Title 'Test Story', got %s", result.Title)
	}
	if result.Status != "draft" {
		t.Errorf("Expected Status 'draft', got %s", result.Status)
	}
}

func TestProtoToChapter(t *testing.T) {
	chapterID := uuid.New()
	storyID := uuid.New()
	now := time.Now()

	pbChapter := &chapterpb.Chapter{
		Id:        chapterID.String(),
		StoryId:   storyID.String(),
		Number:    1,
		Title:     "Chapter 1",
		Status:    "draft",
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	result := protoToChapter(pbChapter)

	if result.ID != chapterID {
		t.Errorf("Expected ID %s, got %s", chapterID, result.ID)
	}
	if result.StoryID != storyID {
		t.Errorf("Expected StoryID %s, got %s", storyID, result.StoryID)
	}
	if result.Number != 1 {
		t.Errorf("Expected Number 1, got %d", result.Number)
	}
	if result.Title != "Chapter 1" {
		t.Errorf("Expected Title 'Chapter 1', got %s", result.Title)
	}
}

func TestProtoToScene(t *testing.T) {
	sceneID := uuid.New()
	storyID := uuid.New()
	chapterID := uuid.New()
	povID := uuid.New()
	now := time.Now()

	pbScene := &scenepb.Scene{
		Id:            sceneID.String(),
		StoryId:       storyID.String(),
		ChapterId:     stringPtr(chapterID.String()),
		OrderNum:      1,
		Goal:          "Test goal",
		TimeRef:       "Morning",
		PovCharacterId: stringPtr(povID.String()),
		CreatedAt:     timestamppb.New(now),
		UpdatedAt:     timestamppb.New(now),
	}

	result := protoToScene(pbScene)

	if result.ID != sceneID {
		t.Errorf("Expected ID %s, got %s", sceneID, result.ID)
	}
	if result.StoryID != storyID {
		t.Errorf("Expected StoryID %s, got %s", storyID, result.StoryID)
	}
	if result.ChapterID == nil || *result.ChapterID != chapterID {
		t.Errorf("Expected ChapterID %s, got %v", chapterID, result.ChapterID)
	}
	if result.POVCharacterID == nil || *result.POVCharacterID != povID {
		t.Errorf("Expected POVCharacterID %s, got %v", povID, result.POVCharacterID)
	}
	if result.Goal != "Test goal" {
		t.Errorf("Expected Goal 'Test goal', got %s", result.Goal)
	}
}

func TestProtoToScene_OptionalFields(t *testing.T) {
	sceneID := uuid.New()
	storyID := uuid.New()
	now := time.Now()

	pbScene := &scenepb.Scene{
		Id:        sceneID.String(),
		StoryId:   storyID.String(),
		OrderNum:  1,
		Goal:      "Test goal",
		TimeRef:   "Morning",
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	result := protoToScene(pbScene)

	if result.ChapterID != nil {
		t.Error("Expected ChapterID to be nil")
	}
	if result.POVCharacterID != nil {
		t.Error("Expected POVCharacterID to be nil")
	}
	if result.LocationID != nil {
		t.Error("Expected LocationID to be nil")
	}
}

func TestProtoToBeat(t *testing.T) {
	beatID := uuid.New()
	sceneID := uuid.New()
	now := time.Now()

	pbBeat := &beatpb.Beat{
		Id:        beatID.String(),
		SceneId:   sceneID.String(),
		OrderNum:  1,
		Type:      "setup",
		Intent:    "Introduce character",
		Outcome:   "Character introduced",
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	result := protoToBeat(pbBeat)

	if result.ID != beatID {
		t.Errorf("Expected ID %s, got %s", beatID, result.ID)
	}
	if result.SceneID != sceneID {
		t.Errorf("Expected SceneID %s, got %s", sceneID, result.SceneID)
	}
	if result.Type != "setup" {
		t.Errorf("Expected Type 'setup', got %s", result.Type)
	}
	if result.Intent != "Introduce character" {
		t.Errorf("Expected Intent 'Introduce character', got %s", result.Intent)
	}
}

func TestProtoToContentBlock(t *testing.T) {
	contentBlockID := uuid.New()
	chapterID := uuid.New()
	orderNum := int32(1)
	now := time.Now()

	metadata, err := structpb.NewStruct(map[string]interface{}{
		"language": "en",
	})
	if err != nil {
		t.Fatalf("failed to build metadata: %v", err)
	}

	pbContentBlock := &contentblockpb.ContentBlock{
		Id:        contentBlockID.String(),
		ChapterId: stringPtr(chapterID.String()),
		OrderNum:  &orderNum,
		Type:      "text",
		Kind:      "final",
		Content:   "Test content",
		Metadata:  metadata,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}

	result := protoToContentBlock(pbContentBlock)

	if result.ID != contentBlockID {
		t.Errorf("Expected ID %s, got %s", contentBlockID, result.ID)
	}
	if result.ChapterID == nil || *result.ChapterID != chapterID {
		t.Errorf("Expected ChapterID %s, got %v", chapterID, result.ChapterID)
	}
	if result.OrderNum == nil || *result.OrderNum != int(orderNum) {
		t.Errorf("Expected OrderNum %d, got %v", orderNum, result.OrderNum)
	}
	if result.Type != "text" {
		t.Errorf("Expected Type 'text', got %s", result.Type)
	}
	if result.Kind != "final" {
		t.Errorf("Expected Kind 'final', got %s", result.Kind)
	}
	if result.Content != "Test content" {
		t.Errorf("Expected Content 'Test content', got %s", result.Content)
	}
	if result.Metadata["language"] != "en" {
		t.Errorf("Expected metadata language 'en', got %v", result.Metadata["language"])
	}
}

func TestProtoToContentBlockReference(t *testing.T) {
	refID := uuid.New()
	contentBlockID := uuid.New()
	entityID := uuid.New()
	now := time.Now()

	pbRef := &contentblockpb.ContentBlockReference{
		Id:             refID.String(),
		ContentBlockId: contentBlockID.String(),
		EntityType:     "character",
		EntityId:       entityID.String(),
		CreatedAt:      timestamppb.New(now),
	}

	result := protoToContentBlockReference(pbRef)

	if result.ID != refID {
		t.Errorf("Expected ID %s, got %s", refID, result.ID)
	}
	if result.ContentBlockID != contentBlockID {
		t.Errorf("Expected ContentBlockID %s, got %s", contentBlockID, result.ContentBlockID)
	}
	if result.EntityType != "character" {
		t.Errorf("Expected EntityType 'character', got %s", result.EntityType)
	}
	if result.EntityID != entityID {
		t.Errorf("Expected EntityID %s, got %s", entityID, result.EntityID)
	}
}

func TestNewMainServiceClient_InvalidAddress(t *testing.T) {
	// Note: gRPC client may not fail immediately on invalid address
	// It uses lazy connection, so we skip this test or test with a real invalid scenario
	// In practice, the error would occur on first RPC call, not on client creation
	t.Skip("gRPC client uses lazy connection - error occurs on first RPC call, not on creation")
}

func TestMainServiceClient_Close(t *testing.T) {
	// This test requires a real gRPC server or a mock
	// For now, we'll skip it or test with a mock server
	// In a real scenario, you'd use bufconn or a test server
	t.Skip("Requires gRPC server setup - testing Close() would need a real connection")
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
