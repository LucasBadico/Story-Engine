package ingest

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
	"github.com/story-engine/llm-gateway-service/internal/ports/embeddings"
	grpcclient "github.com/story-engine/llm-gateway-service/internal/ports/grpc"
	"github.com/story-engine/llm-gateway-service/internal/ports/repositories"
)

func TestIngestProseBlockUseCase_Execute_Success(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	proseBlockID := uuid.New()
	chapterID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	proseBlock := &grpcclient.ProseBlock{
		ID:        proseBlockID,
		ChapterID: chapterID,
		OrderNum:  1,
		Kind:      "final",
		Content:   "This is prose content.",
		WordCount: 4,
	}
	mockClient.AddProseBlock(proseBlock)

	useCase := NewIngestProseBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestProseBlockInput{
		TenantID:     tenantID,
		ProseBlockID: proseBlockID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID == uuid.Nil {
		t.Error("Expected DocumentID to be set")
	}
	if output.ChunkCount != 1 {
		t.Errorf("Expected ChunkCount 1, got %d", output.ChunkCount)
	}
}

func TestIngestProseBlockUseCase_Execute_ProseBlockNotFound(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	proseBlockID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestProseBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestProseBlockInput{
		TenantID:     tenantID,
		ProseBlockID: proseBlockID,
	}

	_, err := useCase.Execute(ctx, input)
	if err == nil {
		t.Error("Expected error when prose block not found")
	}
}

func TestIngestProseBlockUseCase_Execute_WithBeatReference(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	proseBlockID := uuid.New()
	chapterID := uuid.New()
	sceneID := uuid.New()
	beatID := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	proseBlock := &grpcclient.ProseBlock{
		ID:        proseBlockID,
		ChapterID: chapterID,
		OrderNum:  1,
		Kind:      "final",
		Content:   "This is prose content.",
		WordCount: 4,
	}
	mockClient.AddProseBlock(proseBlock)

	scene := &grpcclient.Scene{
		ID:        sceneID,
		StoryID:   uuid.New(),
		ChapterID: &chapterID,
		OrderNum:  1,
		Goal:      "Test goal",
		TimeRef:   "Morning",
	}
	mockClient.AddScene(scene)

	beat := &grpcclient.Beat{
		ID:       beatID,
		SceneID:  sceneID,
		OrderNum: 1,
		Type:     "setup",
		Intent:   "Introduce character",
		Outcome:  "Character introduced",
	}
	mockClient.AddBeat(beat)

	ref := &grpcclient.ProseBlockReference{
		ID:          uuid.New(),
		ProseBlockID: proseBlockID,
		EntityType:  "beat",
		EntityID:    beatID,
	}
	mockClient.AddProseBlockReference(ref)

	useCase := NewIngestProseBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestProseBlockInput{
		TenantID:     tenantID,
		ProseBlockID: proseBlockID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if output.DocumentID == uuid.Nil {
		t.Error("Expected DocumentID to be set")
	}

	// Verify chunk was created with metadata
	chunks, _ := mockChunkRepo.ListByDocument(ctx, output.DocumentID)
	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	chunk := chunks[0]
	if chunk.BeatID == nil || *chunk.BeatID != beatID {
		t.Error("Expected BeatID to be set")
	}
	if chunk.BeatType == nil || *chunk.BeatType != "setup" {
		t.Error("Expected BeatType to be set")
	}
	if chunk.SceneID == nil || *chunk.SceneID != sceneID {
		t.Error("Expected SceneID to be set")
	}
}

func TestIngestProseBlockUseCase_Execute_WithCharacterReferences(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	proseBlockID := uuid.New()
	chapterID := uuid.New()
	characterID1 := uuid.New()
	characterID2 := uuid.New()

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	proseBlock := &grpcclient.ProseBlock{
		ID:        proseBlockID,
		ChapterID: chapterID,
		OrderNum:  1,
		Kind:      "final",
		Content:   "John and Mary walked together.",
		WordCount: 5,
	}
	mockClient.AddProseBlock(proseBlock)

	ref1 := &grpcclient.ProseBlockReference{
		ID:          uuid.New(),
		ProseBlockID: proseBlockID,
		EntityType:  "character",
		EntityID:    characterID1,
	}
	mockClient.AddProseBlockReference(ref1)

	ref2 := &grpcclient.ProseBlockReference{
		ID:          uuid.New(),
		ProseBlockID: proseBlockID,
		EntityType:  "character",
		EntityID:    characterID2,
	}
	mockClient.AddProseBlockReference(ref2)

	useCase := NewIngestProseBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	input := IngestProseBlockInput{
		TenantID:     tenantID,
		ProseBlockID: proseBlockID,
	}

	output, err := useCase.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	chunks, _ := mockChunkRepo.ListByDocument(ctx, output.DocumentID)
	if len(chunks) != 1 {
		t.Fatalf("Expected 1 chunk, got %d", len(chunks))
	}

	chunk := chunks[0]
	if len(chunk.Characters) != 2 {
		t.Errorf("Expected 2 characters, got %d", len(chunk.Characters))
	}
}

func TestIngestProseBlockUseCase_buildEnrichedContent(t *testing.T) {
	proseBlock := &grpcclient.ProseBlock{
		Content: "This is prose content.",
		Kind:    "final",
	}

	metadata := &ProseBlockMetadata{
		BeatType:     stringPtr("setup"),
		BeatIntent:   stringPtr("Introduce character"),
		Characters:   []string{"John", "Mary"},
		LocationName: stringPtr("Library"),
		Timeline:     stringPtr("Morning"),
	}

	mockClient := grpcclient.NewMockMainServiceClient()
	mockDocRepo := repositories.NewMockDocumentRepository()
	mockChunkRepo := repositories.NewMockChunkRepository()
	mockEmbedder := embeddings.NewMockEmbedder(768)
	log := logger.New()

	useCase := NewIngestProseBlockUseCase(mockClient, mockDocRepo, mockChunkRepo, mockEmbedder, log)

	content := useCase.buildEnrichedContent(proseBlock, metadata)

	if content == "" {
		t.Error("Expected non-empty content")
	}
	if !strings.Contains(content, "This is prose content") {
		t.Error("Expected content to contain prose block content")
	}
	if !strings.Contains(content, "setup") {
		t.Error("Expected content to contain beat type")
	}
	if !strings.Contains(content, "John") || !strings.Contains(content, "Mary") {
		t.Error("Expected content to contain characters")
	}
	if !strings.Contains(content, "Library") {
		t.Error("Expected content to contain location")
	}
	if !strings.Contains(content, "Morning") {
		t.Error("Expected content to contain timeline")
	}
}

func stringPtr(s string) *string {
	return &s
}

