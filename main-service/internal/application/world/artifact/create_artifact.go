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
	artifactRepo         repositories.ArtifactRepository
	artifactReferenceRepo repositories.ArtifactReferenceRepository
	worldRepo            repositories.WorldRepository
	characterRepo        repositories.CharacterRepository
	locationRepo         repositories.LocationRepository
	auditLogRepo         repositories.AuditLogRepository
	logger               logger.Logger
}

// NewCreateArtifactUseCase creates a new CreateArtifactUseCase
func NewCreateArtifactUseCase(
	artifactRepo repositories.ArtifactRepository,
	artifactReferenceRepo repositories.ArtifactReferenceRepository,
	worldRepo repositories.WorldRepository,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *CreateArtifactUseCase {
	return &CreateArtifactUseCase{
		artifactRepo:         artifactRepo,
		artifactReferenceRepo: artifactReferenceRepo,
		worldRepo:            worldRepo,
		characterRepo:        characterRepo,
		locationRepo:         locationRepo,
		auditLogRepo:         auditLogRepo,
		logger:               logger,
	}
}

// CreateArtifactInput represents the input for creating an artifact
type CreateArtifactInput struct {
	TenantID     uuid.UUID
	WorldID      uuid.UUID
	CharacterIDs []uuid.UUID
	LocationIDs  []uuid.UUID
	Name         string
	Description  string
	Rarity       string
}

// CreateArtifactOutput represents the output of creating an artifact
type CreateArtifactOutput struct {
	Artifact *world.Artifact
}

// Execute creates a new artifact
func (uc *CreateArtifactUseCase) Execute(ctx context.Context, input CreateArtifactInput) (*CreateArtifactOutput, error) {
	// Validate world exists
	w, err := uc.worldRepo.GetByID(ctx, input.TenantID, input.WorldID)
	if err != nil {
		return nil, err
	}

	// Validate characters exist if provided
	for _, characterID := range input.CharacterIDs {
		c, err := uc.characterRepo.GetByID(ctx, input.TenantID, characterID)
		if err != nil {
			return nil, err
		}
		if c.WorldID != input.WorldID {
			return nil, &platformerrors.ValidationError{
				Field:   "character_ids",
				Message: "all characters must belong to the same world",
			}
		}
	}

	// Validate locations exist if provided
	for _, locationID := range input.LocationIDs {
		l, err := uc.locationRepo.GetByID(ctx, input.TenantID, locationID)
		if err != nil {
			return nil, err
		}
		if l.WorldID != input.WorldID {
			return nil, &platformerrors.ValidationError{
				Field:   "location_ids",
				Message: "all locations must belong to the same world",
			}
		}
	}

	newArtifact, err := world.NewArtifact(input.TenantID, input.WorldID, input.Name)
	if err != nil {
		return nil, &platformerrors.ValidationError{
			Field:   "name",
			Message: err.Error(),
		}
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

	// Create character references
	for _, characterID := range input.CharacterIDs {
		ref, err := world.NewArtifactReference(newArtifact.ID, world.ArtifactReferenceEntityTypeCharacter, characterID)
		if err != nil {
			return nil, err
		}
		if err := uc.artifactReferenceRepo.Create(ctx, ref); err != nil {
			uc.logger.Error("failed to create character reference", "error", err)
			return nil, err
		}
	}

	// Create location references
	for _, locationID := range input.LocationIDs {
		ref, err := world.NewArtifactReference(newArtifact.ID, world.ArtifactReferenceEntityTypeLocation, locationID)
		if err != nil {
			return nil, err
		}
		if err := uc.artifactReferenceRepo.Create(ctx, ref); err != nil {
			uc.logger.Error("failed to create location reference", "error", err)
			return nil, err
		}
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

