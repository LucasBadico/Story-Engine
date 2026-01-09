package artifact

import (
	"context"

	"github.com/google/uuid"
	relationapp "github.com/story-engine/main-service/internal/application/relation"
	"github.com/story-engine/main-service/internal/core/audit"
	"github.com/story-engine/main-service/internal/core/world"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/queue"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// UpdateArtifactUseCase handles artifact updates
type UpdateArtifactUseCase struct {
	artifactRepo          repositories.ArtifactRepository
	createRelationUseCase *relationapp.CreateRelationUseCase
	listRelationsUseCase  *relationapp.ListRelationsBySourceUseCase
	deleteRelationUseCase *relationapp.DeleteRelationUseCase
	characterRepo         repositories.CharacterRepository
	locationRepo          repositories.LocationRepository
	worldRepo             repositories.WorldRepository
	auditLogRepo          repositories.AuditLogRepository
	ingestionQueue        queue.IngestionQueue
	logger                logger.Logger
}

// NewUpdateArtifactUseCase creates a new UpdateArtifactUseCase
func NewUpdateArtifactUseCase(
	artifactRepo repositories.ArtifactRepository,
	createRelationUseCase *relationapp.CreateRelationUseCase,
	listRelationsUseCase *relationapp.ListRelationsBySourceUseCase,
	deleteRelationUseCase *relationapp.DeleteRelationUseCase,
	characterRepo repositories.CharacterRepository,
	locationRepo repositories.LocationRepository,
	worldRepo repositories.WorldRepository,
	auditLogRepo repositories.AuditLogRepository,
	logger logger.Logger,
) *UpdateArtifactUseCase {
	return &UpdateArtifactUseCase{
		artifactRepo:          artifactRepo,
		createRelationUseCase: createRelationUseCase,
		listRelationsUseCase:  listRelationsUseCase,
		deleteRelationUseCase: deleteRelationUseCase,
		characterRepo:         characterRepo,
		locationRepo:          locationRepo,
		worldRepo:             worldRepo,
		auditLogRepo:          auditLogRepo,
		ingestionQueue:        nil,
		logger:                logger,
	}
}

func (uc *UpdateArtifactUseCase) SetIngestionQueue(queue queue.IngestionQueue) {
	uc.ingestionQueue = queue
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

	// Update character references using entity_relations
	if input.CharacterIDs != nil {
		// Delete all existing character references
		output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
			TenantID:   input.TenantID,
			SourceType: "artifact",
			SourceID:   a.ID,
			Options: repositories.ListOptions{
				Limit: 100,
			},
		})
		if err == nil {
			for _, rel := range output.Relations.Items {
				if rel.TargetType == "character" {
					if err := uc.deleteRelationUseCase.Execute(ctx, relationapp.DeleteRelationInput{
						TenantID: input.TenantID,
						ID:       rel.ID,
					}); err != nil {
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
			_, err = uc.createRelationUseCase.Execute(ctx, relationapp.CreateRelationInput{
				TenantID:     input.TenantID,
				WorldID:      a.WorldID,
				SourceType:   "artifact",
				SourceID:     a.ID,
				TargetType:   "character",
				TargetID:     characterID,
				RelationType: "mentions",
				Attributes:   make(map[string]interface{}),
				CreateMirror: false,
			})
			if err != nil {
				uc.logger.Error("failed to create character reference", "error", err)
				return nil, err
			}
		}
	}

	// Update location references using entity_relations
	if input.LocationIDs != nil {
		// Delete all existing location references
		output, err := uc.listRelationsUseCase.Execute(ctx, relationapp.ListRelationsBySourceInput{
			TenantID:   input.TenantID,
			SourceType: "artifact",
			SourceID:   a.ID,
			Options: repositories.ListOptions{
				Limit: 100,
			},
		})
		if err == nil {
			for _, rel := range output.Relations.Items {
				if rel.TargetType == "location" {
					if err := uc.deleteRelationUseCase.Execute(ctx, relationapp.DeleteRelationInput{
						TenantID: input.TenantID,
						ID:       rel.ID,
					}); err != nil {
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
			_, err = uc.createRelationUseCase.Execute(ctx, relationapp.CreateRelationInput{
				TenantID:     input.TenantID,
				WorldID:      a.WorldID,
				SourceType:   "artifact",
				SourceID:     a.ID,
				TargetType:   "location",
				TargetID:     locationID,
				RelationType: "mentions",
				Attributes:   make(map[string]interface{}),
				CreateMirror: false,
			})
			if err != nil {
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
	uc.enqueueIngestion(ctx, input.TenantID, a.ID)

	return &UpdateArtifactOutput{
		Artifact: a,
	}, nil
}

func (uc *UpdateArtifactUseCase) enqueueIngestion(ctx context.Context, tenantID uuid.UUID, artifactID uuid.UUID) {
	if uc.ingestionQueue == nil {
		return
	}
	if err := uc.ingestionQueue.Push(ctx, tenantID, "artifact", artifactID); err != nil {
		uc.logger.Error("failed to enqueue artifact ingestion", "error", err, "artifact_id", artifactID, "tenant_id", tenantID)
	}
}
