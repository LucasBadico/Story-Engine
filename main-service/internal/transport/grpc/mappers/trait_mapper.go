package mappers

import (
	"github.com/story-engine/main-service/internal/core/world"
	traitpb "github.com/story-engine/main-service/proto/trait"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TraitToProto converts a trait domain entity to a protobuf message
func TraitToProto(t *world.Trait) *traitpb.Trait {
	if t == nil {
		return nil
	}

	return &traitpb.Trait{
		Id:          t.ID.String(),
		TenantId:    t.TenantID.String(),
		Name:        t.Name,
		Category:    t.Category,
		Description: t.Description,
		CreatedAt:   timestamppb.New(t.CreatedAt),
		UpdatedAt:   timestamppb.New(t.UpdatedAt),
	}
}


