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

// UpdateArtifactUseCase handles artifact updates
type UpdateArtifactUseCase struct {
	artifactRepo repositories.ArtifactRepository
	characterRepo repositories.CharacterRepository
	locationRepo repositories.LocationRepository
	worldRepo repositories.WorldRepository
	auditLogRepo repositories.AuditLogRepository
	logger        logger.Logger
}

// NewUpdateArtifactUseCase creates a new UpdateArtifactUseCase
func NewUpdateArtifactUseCase(
	artifactRepo repositories.ArtifactRepository,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateArtifactUseCase {
	return &UpdateArtifactUseCase{
		artifactRepo: artifactRepo,
		characterRepo: characterRepo,
		locationRepo: locationRepo,
		worldRepo: worldRepo,
		auditLogRepo: auditLogRepo,
		logger:        logger,
	}
}

// UpdateArtifactInput represents the input for updating an artifact
type UpdateArtifactInput struct {
	ID          uuid.UUID
	Name        *string
	Description *string
	Rarity      *string
	CharacterID *uuid.UUID
	LocationID  *uuid.UUID
}

// UpdateArtifactOutput represents the output of updating an artifact
type UpdateArtifactOutput struct {
	Artifact *world.Artifact
}

// Execute updates an artifact
func (uc *UpdateArtifactUseCase) Execute(ctx context.Context, input UpdateArtifactInput) (*UpdateArtifactOutput, error) {
	a, err := uc.artifactRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		if err := a.UpdateName(*input.Name); err != nil {
			return nil, &platformerrors.ValidationError{
				Field:   "name",
				Message: err.Error(),
			}
		}
	}
	if input.Description != nil {
		a.UpdateDescription(*input.Description)
	}
	if input.Rarity != nil {
		a.UpdateRarity(*input.Rarity)
	}
	if input.CharacterID != nil {
		if *input.CharacterID != uuid.Nil {
			c, err := uc.characterRepo.GetByID(ctx, *input.CharacterID)
			if err != nil {
				return nil, err
			}
			if c.WorldID != a.WorldID {
				return nil, &platformerrors.ValidationError{
					Field:   "character_id",
					Message: "character must belong to the same world",
				}
			}
		}
		a.SetCharacter(input.CharacterID)
	}
	if input.LocationID != nil {
		if *input.LocationID != uuid.Nil {
			l, err := uc.locationRepo.GetByID(ctx, *input.LocationID)
			if err != nil {
				return nil, err
			}
			if l.WorldID != a.WorldID {
				return nil, &platformerrors.ValidationError{
					Field:   "location_id",
					Message: "location must belong to the same world",
				}
			}
		}
		a.SetLocation(input.LocationID)
	}

	if err := a.Validate(); err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "artifact",
			Message: err.Error(),
		}
	}

	if err := uc.artifactRepo.Update(ctx, a); err != nil {
		uc.logger.Error("failed to update artifact", "error", err, "artifact_id", input.ID)
		return nil, err
	}

	w, _ := uc.worldRepo.GetByID(ctx, a.WorldID)
	auditLog := audit.NewAuditLog(
		w.TenantID,
		nil,
		audit.ActionUpdate,
		audit.EntityTypeArtifact,
		a.ID,
		map[string]interface{}{
			"name": a.Name,
		},
	)
	if err := uc.auditLogRepo.Create(ctx, auditLog); err != nil {
		uc.logger.Warn("failed to create audit log", "error", err)
	}

	uc.logger.Info("artifact updated", "artifact_id", a.ID, "name", a.Name)

	return &UpdateArtifactOutput{
		Artifact: a,
	}, nil
}

