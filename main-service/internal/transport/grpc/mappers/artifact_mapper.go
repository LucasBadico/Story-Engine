package mappers

import (
	"github.com/story-engine/main-service/internal/core/world"
	artifactpb "github.com/story-engine/main-service/proto/artifact"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ArtifactToProto converts an artifact domain entity to a protobuf message
func ArtifactToProto(a *world.Artifact) *artifactpb.Artifact {
	if a == nil {
		return nil
	}

	pb := &artifactpb.Artifact{
		Id:          a.ID.String(),
		WorldId:     a.WorldID.String(),
		Name:        a.Name,
		Description: a.Description,
		Rarity:      a.Rarity,
		CreatedAt:   timestamppb.New(a.CreatedAt),
		UpdatedAt:   timestamppb.New(a.UpdatedAt),
	}

	if a.CharacterID != nil {
		pb.CharacterId = a.CharacterID.String()
	}

	if a.LocationID != nil {
		pb.LocationId = a.LocationID.String()
	}

	return pb
}

