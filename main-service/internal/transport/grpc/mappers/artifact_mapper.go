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

	return &artifactpb.Artifact{
		Id:          a.ID.String(),
		WorldId:     a.WorldID.String(),
		Name:        a.Name,
		Description: a.Description,
		Rarity:      a.Rarity,
		CreatedAt:   timestamppb.New(a.CreatedAt),
		UpdatedAt:   timestamppb.New(a.UpdatedAt),
	}
}

// ArtifactReferenceToProto converts an artifact reference domain entity to a protobuf message
func ArtifactReferenceToProto(ref *world.ArtifactReference) *artifactpb.ArtifactReference {
	if ref == nil {
		return nil
	}

	return &artifactpb.ArtifactReference{
		Id:         ref.ID.String(),
		ArtifactId: ref.ArtifactID.String(),
		EntityType: string(ref.EntityType),
		EntityId:   ref.EntityID.String(),
		CreatedAt:  timestamppb.New(ref.CreatedAt),
	}
}

