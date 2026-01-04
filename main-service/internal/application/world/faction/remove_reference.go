package faction

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// RemoveReferenceUseCase handles removing a reference from a faction
type RemoveReferenceUseCase struct {
	factionReferenceRepo repositories.FactionReferenceRepository
	logger               logger.Logger
}

// NewRemoveReferenceUseCase creates a new RemoveReferenceUseCase
func NewRemoveReferenceUseCase(
	factionReferenceRepo repositories.FactionReferenceRepository,
	logger logger.Logger,
) *RemoveReferenceUseCase {
	return &RemoveReferenceUseCase{
		factionReferenceRepo: factionReferenceRepo,
		logger:               logger,
	}
}

// RemoveReferenceInput represents the input for removing a reference
type RemoveReferenceInput struct {
	TenantID   uuid.UUID
	FactionID  uuid.UUID
	EntityType string
	EntityID   uuid.UUID
}

// Execute removes a reference from a faction
func (uc *RemoveReferenceUseCase) Execute(ctx context.Context, input RemoveReferenceInput) error {
	err := uc.factionReferenceRepo.DeleteByFactionAndEntity(ctx, input.TenantID, input.FactionID, input.EntityType, input.EntityID)
	if err != nil {
		uc.logger.Error("failed to remove faction reference", "error", err, "faction_id", input.FactionID)
		return err
	}

	uc.logger.Info("faction reference removed", "faction_id", input.FactionID, "entity_type", input.EntityType, "entity_id", input.EntityID)
	return nil
}

