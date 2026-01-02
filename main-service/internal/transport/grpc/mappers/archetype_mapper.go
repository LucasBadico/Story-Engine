package mappers

import (
	"github.com/story-engine/main-service/internal/core/world"
	archetypepb "github.com/story-engine/main-service/proto/archetype"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ArchetypeToProto converts an archetype domain entity to a protobuf message
func ArchetypeToProto(a *world.Archetype) *archetypepb.Archetype {
	if a == nil {
		return nil
	}

	return &archetypepb.Archetype{
		Id:          a.ID.String(),
		TenantId:    a.TenantID.String(),
		Name:        a.Name,
		Description: a.Description,
		CreatedAt:   timestamppb.New(a.CreatedAt),
		UpdatedAt:   timestamppb.New(a.UpdatedAt),
	}
}


