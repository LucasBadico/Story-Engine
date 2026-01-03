package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	imageblockapp "github.com/story-engine/main-service/internal/application/story/image_block"
	"github.com/story-engine/main-service/internal/core/story"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/http/middleware"
)

// ImageBlockHandler handles HTTP requests for image blocks
type ImageBlockHandler struct {
	createImageBlockUseCase      *imageblockapp.CreateImageBlockUseCase
	getImageBlockUseCase         *imageblockapp.GetImageBlockUseCase
	listImageBlocksUseCase       *imageblockapp.ListImageBlocksUseCase
	updateImageBlockUseCase      *imageblockapp.UpdateImageBlockUseCase
	deleteImageBlockUseCase      *imageblockapp.DeleteImageBlockUseCase
	addReferenceUseCase          *imageblockapp.AddImageBlockReferenceUseCase
	removeReferenceUseCase       *imageblockapp.RemoveImageBlockReferenceUseCase
	getReferencesUseCase         *imageblockapp.GetImageBlockReferencesUseCase
	logger                       logger.Logger
}

// NewImageBlockHandler creates a new ImageBlockHandler
func NewImageBlockHandler(
	createImageBlockUseCase *imageblockapp.CreateImageBlockUseCase,
	getImageBlockUseCase *imageblockapp.GetImageBlockUseCase,
	listImageBlocksUseCase *imageblockapp.ListImageBlocksUseCase,
	updateImageBlockUseCase *imageblockapp.UpdateImageBlockUseCase,
	deleteImageBlockUseCase *imageblockapp.DeleteImageBlockUseCase,
	addReferenceUseCase *imageblockapp.AddImageBlockReferenceUseCase,
	removeReferenceUseCase *imageblockapp.RemoveImageBlockReferenceUseCase,
	getReferencesUseCase *imageblockapp.GetImageBlockReferencesUseCase,
	logger logger.Logger,
) *ImageBlockHandler {
	return &ImageBlockHandler{
		createImageBlockUseCase: createImageBlockUseCase,
		getImageBlockUseCase:    getImageBlockUseCase,
		listImageBlocksUseCase:   listImageBlocksUseCase,
		updateImageBlockUseCase:  updateImageBlockUseCase,
		deleteImageBlockUseCase: deleteImageBlockUseCase,
		addReferenceUseCase:     addReferenceUseCase,
		removeReferenceUseCase:  removeReferenceUseCase,
		getReferencesUseCase:    getReferencesUseCase,
		logger:                  logger,
	}
}

// Create handles POST /api/v1/chapters/{id}/image-blocks
func (h *ImageBlockHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	chapterIDStr := r.PathValue("id")

	var chapterID *uuid.UUID
	if chapterIDStr != "" {
		parsedChapterID, err := uuid.Parse(chapterIDStr)
		if err != nil {
			WriteError(w, &platformerrors.ValidationError{
				Field:   "id",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}
		chapterID = &parsedChapterID
	}

	var req struct {
		OrderNum *int    `json:"order_num,omitempty"`
		Kind     string  `json:"kind"`
		ImageURL string  `json:"image_url"`
		AltText  *string `json:"alt_text,omitempty"`
		Caption  *string `json:"caption,omitempty"`
		Width    *int    `json:"width,omitempty"`
		Height   *int    `json:"height,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	kind := story.ImageKind(req.Kind)
	output, err := h.createImageBlockUseCase.Execute(r.Context(), imageblockapp.CreateImageBlockInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
		OrderNum:  req.OrderNum,
		Kind:      kind,
		ImageURL:  req.ImageURL,
		AltText:   req.AltText,
		Caption:   req.Caption,
		Width:     req.Width,
		Height:    req.Height,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"image_block": output.ImageBlock,
	})
}

// Get handles GET /api/v1/image-blocks/{id}
func (h *ImageBlockHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	imageBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getImageBlockUseCase.Execute(r.Context(), imageblockapp.GetImageBlockInput{
		TenantID: tenantID,
		ID:       imageBlockID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"image_block": output.ImageBlock,
	})
}

// List handles GET /api/v1/chapters/{id}/image-blocks
func (h *ImageBlockHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	chapterIDStr := r.PathValue("id")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.listImageBlocksUseCase.Execute(r.Context(), imageblockapp.ListImageBlocksInput{
		TenantID:  tenantID,
		ChapterID: chapterID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"image_blocks": output.ImageBlocks,
		"total":        len(output.ImageBlocks),
	})
}

// Update handles PUT /api/v1/image-blocks/{id}
func (h *ImageBlockHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	imageBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		ChapterID *uuid.UUID `json:"chapter_id,omitempty"`
		OrderNum *int       `json:"order_num,omitempty"`
		Kind     *string    `json:"kind,omitempty"`
		ImageURL *string    `json:"image_url,omitempty"`
		AltText  *string    `json:"alt_text,omitempty"`
		Caption  *string    `json:"caption,omitempty"`
		Width    *int       `json:"width,omitempty"`
		Height   *int       `json:"height,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	var kind *story.ImageKind
	if req.Kind != nil {
		k := story.ImageKind(*req.Kind)
		kind = &k
	}

	output, err := h.updateImageBlockUseCase.Execute(r.Context(), imageblockapp.UpdateImageBlockInput{
		TenantID:  tenantID,
		ID:        imageBlockID,
		ChapterID: req.ChapterID,
		OrderNum:  req.OrderNum,
		Kind:      kind,
		ImageURL:  req.ImageURL,
		AltText:   req.AltText,
		Caption:   req.Caption,
		Width:     req.Width,
		Height:    req.Height,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"image_block": output.ImageBlock,
	})
}

// Delete handles DELETE /api/v1/image-blocks/{id}
func (h *ImageBlockHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	imageBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.deleteImageBlockUseCase.Execute(r.Context(), imageblockapp.DeleteImageBlockInput{
		TenantID: tenantID,
		ID:       imageBlockID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetReferences handles GET /api/v1/image-blocks/{id}/references
func (h *ImageBlockHandler) GetReferences(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	imageBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.getReferencesUseCase.Execute(r.Context(), imageblockapp.GetImageBlockReferencesInput{
		TenantID:     tenantID,
		ImageBlockID: imageBlockID,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"references": output.References,
		"total":      len(output.References),
	})
}

// AddReference handles POST /api/v1/image-blocks/{id}/references
func (h *ImageBlockHandler) AddReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	imageBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	var req struct {
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	entityType := story.ImageBlockReferenceEntityType(req.EntityType)
	if !isValidImageBlockReferenceEntityType(entityType) {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type, must be one of: scene, beat, chapter, character, location, artifact, event, world",
		}, http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(req.EntityID)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.addReferenceUseCase.Execute(r.Context(), imageblockapp.AddImageBlockReferenceInput{
		TenantID:     tenantID,
		ImageBlockID: imageBlockID,
		EntityType:   entityType,
		EntityID:     entityID,
	}); err != nil {
		WriteError(w, err, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// RemoveReference handles DELETE /api/v1/image-blocks/{id}/references/{entity_type}/{entity_id}
func (h *ImageBlockHandler) RemoveReference(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	id := r.PathValue("id")
	entityTypeStr := r.PathValue("entity_type")
	entityIDStr := r.PathValue("entity_id")

	imageBlockID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	entityType := story.ImageBlockReferenceEntityType(entityTypeStr)
	if !isValidImageBlockReferenceEntityType(entityType) {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_type",
			Message: "invalid entity type, must be one of: scene, beat, chapter, character, location, artifact, event, world",
		}, http.StatusBadRequest)
		return
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "entity_id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	if err := h.removeReferenceUseCase.Execute(r.Context(), imageblockapp.RemoveImageBlockReferenceInput{
		TenantID:     tenantID,
		ImageBlockID: imageBlockID,
		EntityType:   entityType,
		EntityID:     entityID,
	}); err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func isValidImageBlockReferenceEntityType(entityType story.ImageBlockReferenceEntityType) bool {
	return entityType == story.ImageBlockReferenceEntityTypeScene ||
		entityType == story.ImageBlockReferenceEntityTypeBeat ||
		entityType == story.ImageBlockReferenceEntityTypeChapter ||
		entityType == story.ImageBlockReferenceEntityTypeCharacter ||
		entityType == story.ImageBlockReferenceEntityTypeLocation ||
		entityType == story.ImageBlockReferenceEntityTypeArtifact ||
		entityType == story.ImageBlockReferenceEntityTypeEvent ||
		entityType == story.ImageBlockReferenceEntityTypeWorld
}


