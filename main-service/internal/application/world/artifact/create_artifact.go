package artifact

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// CreateArtifactUseCase handles artifact creation
type CreateArtifactUseCase struct {
	artifactRepo repositories.ArtifactRepository
	worldRepo    repositories.WorldRepository
	characterRepo repositories.CharacterRepository
	locationRepo repositories.LocationRepository
	auditLogRepo repositories.AuditLogRepository
	logger       logger.Logger
}

// NewCreateArtifactUseCase creates a new CreateArtifactUseCase
func NewCreateArtifactUseCase(
	artifactRepo repositories.ArtifactRepository,
	worldRepo repositories.WorldRepository,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateArtifactUseCase {
	return &CreateArtifactUseCase{
		artifactRepo: artifactRepo,
		worldRepo:    worldRepo,
		characterRepo: characterRepo,
		locationRepo: locationRepo,
		auditLogRepo: auditLogRepo,
		logger:       logger,
	}
}

// CreateArtifactInput represents the input for creating an artifact
type CreateArtifactInput struct {
	WorldID     uuid.UUID
	CharacterID *uuid.UUID
	LocationID  *uuid.UUID
	Name        string
	Description string
	Rarity      string
}

// CreateArtifactOutput represents the output of creating an artifact
type CreateArtifactOutput struct {
	Artifact *world.Artifact
}

// Execute creates a new artifact
func (uc *CreateArtifactUseCase) Execute(ctx context.Context, input CreateArtifactInput) (*CreateArtifactOutput, error) {
	// Validate world exists
	w, err := uc.worldRepo.GetByID(ctx, input.WorldID)
	if err != nil {
		return nil, err
	}

	// Validate character exists if provided
	if input.CharacterID != nil {
		c, err := uc.characterRepo.GetByID(ctx, *input.CharacterID)
		if err != nil {
			return nil, err
		}
		if c.WorldID != input.WorldID {
			return nil, &platformerrors.ValidationError{
				Field:   "character_id",
				Message: "character must belong to the same world",
			}
		}
	}

	// Validate location exists if provided
	if input.LocationID != nil {
		l, err := uc.locationRepo.GetByID(ctx, *input.LocationID)
		if err != nil {
			return nil, err
		}
		if l.WorldID != input.WorldID {
			return nil, &platformerrors.ValidationError{
				Field:   "location_id",
				Message: "location must belong to the same world",
			}
		}
	}

	newArtifact, err := world.NewArtifact(input.WorldID, input.Name)
	if err != nil {
		return nil, err
	}

	if input.CharacterID != nil {
		newArtifact.SetCharacter(input.CharacterID)
	}
	if input.LocationID != nil {
		newArtifact.SetLocation(input.LocationID)
	}
	if input.Description != "" {
		newArtifact.UpdateDescription(input.Description)
	}
	if input.Rarity != "" {
		newArtifact.UpdateRarity(input.Rarity)
	}

	if err := newArtifact.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "artifact",
			Message: err.Error(),
		}
	}

	if err := uc.artifactRepo.Create(ctx, newArtifact); err != nil {
		uc.logger.Error("failed to create artifact", "error", err, "name", input.Name)
		return nil, err
	}

	auditLog := audit.NewAuditLog(
		w.TenantID,
		nil,
		audit.ActionCreate,
		audit.EntityTypeArtifact,
		newArtifact.ID,
		map[string]interface{}{
			"name":     newArtifact.Name,
			"world_id": newArtifact.WorldID.String(),
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("artifact created", "artifact_id", newArtifact.ID, "name", newArtifact.Name)

	return &CreateArtifactOutput{
		Artifact: newArtifact,
	}, nil
}

