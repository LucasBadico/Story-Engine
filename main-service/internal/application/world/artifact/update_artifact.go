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
	artifactRepo         repositories.ArtifactRepository
	artifactReferenceRepo repositories.ArtifactReferenceRepository
	characterRepo        repositories.CharacterRepository
	locationRepo         repositories.LocationRepository
	worldRepo            repositories.WorldRepository
	auditLogRepo         repositories.AuditLogRepository
	logger               logger.Logger
}

// NewUpdateArtifactUseCase creates a new UpdateArtifactUseCase
func NewUpdateArtifactUseCase(
	artifactRepo repositories.ArtifactRepository,
	artifactReferenceRepo repositories.ArtifactReferenceRepository,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateArtifactUseCase {
	return &UpdateArtifactUseCase{
		artifactRepo:         artifactRepo,
		artifactReferenceRepo: artifactReferenceRepo,
		characterRepo:        characterRepo,
		locationRepo:         locationRepo,
		worldRepo:            worldRepo,
		auditLogRepo:         auditLogRepo,
		logger:               logger,
	}
}

// UpdateArtifactInput represents the input for updating an artifact
type UpdateArtifactInput struct {
	TenantID     uuid.UUID
	ID           uuid.UUID
	Name         *string
	Description  *string
	Rarity       *string
	CharacterIDs *[]uuid.UUID // nil = no change, empty slice = remove all
	LocationIDs  *[]uuid.UUID // nil = no change, empty slice = remove all
}

// UpdateArtifactOutput represents the output of updating an artifact
type UpdateArtifactOutput struct {
	Artifact *world.Artifact
}

// Execute updates an artifact
func (uc *UpdateArtifactUseCase) Execute(ctx context.Context, input UpdateArtifactInput) (*UpdateArtifactOutput, error) {
	a, err := uc.artifactRepo.GetByID(ctx, input.TenantID, input.ID)
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

	// Update character references
	if input.CharacterIDs != nil {
		// Delete all existing character references
		existingRefs, err := uc.artifactReferenceRepo.ListByArtifact(ctx, input.TenantID, a.ID)
		if err == nil {
			for _, ref := range existingRefs {
				if ref.EntityType == world.ArtifactReferenceEntityTypeCharacter {
					if err := uc.artifactReferenceRepo.Delete(ctx, input.TenantID, ref.ID); err != nil {
						uc.logger.Warn("failed to delete character reference", "error", err)
					}
				}
			}
		}

		// Create new character references
		for _, characterID := range *input.CharacterIDs {
			c, err := uc.characterRepo.GetByID(ctx, input.TenantID, characterID)
			if err != nil {
				return nil, err
			}
			if c.WorldID != a.WorldID {
				return nil, &platformerrors.ValidationError{
					Field:   "character_ids",
					Message: "all characters must belong to the same world",
				}
			}
			ref, err := world.NewArtifactReference(a.ID, world.ArtifactReferenceEntityTypeCharacter, characterID)
			if err != nil {
				return nil, err
			}
			if err := uc.artifactReferenceRepo.Create(ctx, ref); err != nil {
				uc.logger.Error("failed to create character reference", "error", err)
				return nil, err
			}
		}
	}

	// Update location references
	if input.LocationIDs != nil {
		// Delete all existing location references
		existingRefs, err := uc.artifactReferenceRepo.ListByArtifact(ctx, input.TenantID, a.ID)
		if err == nil {
			for _, ref := range existingRefs {
				if ref.EntityType == world.ArtifactReferenceEntityTypeLocation {
					if err := uc.artifactReferenceRepo.Delete(ctx, input.TenantID, ref.ID); err != nil {
						uc.logger.Warn("failed to delete location reference", "error", err)
					}
				}
			}
		}

		// Create new location references
		for _, locationID := range *input.LocationIDs {
			l, err := uc.locationRepo.GetByID(ctx, input.TenantID, locationID)
			if err != nil {
				return nil, err
			}
			if l.WorldID != a.WorldID {
				return nil, &platformerrors.ValidationError{
					Field:   "location_ids",
					Message: "all locations must belong to the same world",
				}
			}
			ref, err := world.NewArtifactReference(a.ID, world.ArtifactReferenceEntityTypeLocation, locationID)
			if err != nil {
				return nil, err
			}
			if err := uc.artifactReferenceRepo.Create(ctx, ref); err != nil {
				uc.logger.Error("failed to create location reference", "error", err)
				return nil, err
			}
		}
	}

	w, _ := uc.worldRepo.GetByID(ctx, input.TenantID, a.WorldID)
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

