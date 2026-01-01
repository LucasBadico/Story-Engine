package mappers

import (
	"github.com/story-engine/main-service/internal/core/world"
	worldpb "github.com/story-engine/main-service/proto/world"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// WorldToProto converts a world domain entity to a protobuf message
func WorldToProto(w *world.World) *worldpb.World {
	if w == nil {
		return nil
	}

	return &worldpb.World{
		Id:          w.ID.String(),
		TenantId:    w.TenantID.String(),
		Name:        w.Name,
		Description: w.Description,
		Genre:       w.Genre,
		IsImplicit:  w.IsImplicit,
		CreatedAt:   timestamppb.New(w.CreatedAt),
		UpdatedAt:   timestamppb.New(w.UpdatedAt),
	}
}

