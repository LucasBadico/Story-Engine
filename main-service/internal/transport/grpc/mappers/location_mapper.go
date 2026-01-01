package mappers

import (
	"github.com/story-engine/main-service/internal/core/world"
	locationpb "github.com/story-engine/main-service/proto/location"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// LocationToProto converts a location domain entity to a protobuf message
func LocationToProto(l *world.Location) *locationpb.Location {
	if l == nil {
		return nil
	}

	pb := &locationpb.Location{
		Id:             l.ID.String(),
		WorldId:        l.WorldID.String(),
		Name:           l.Name,
		Type:           l.Type,
		Description:    l.Description,
		HierarchyLevel: int32(l.HierarchyLevel),
		CreatedAt:      timestamppb.New(l.CreatedAt),
		UpdatedAt:      timestamppb.New(l.UpdatedAt),
	}

	if l.ParentID != nil {
		pb.ParentId = l.ParentID.String()
	}

	return pb
}

