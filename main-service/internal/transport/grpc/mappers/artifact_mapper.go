package mappers

import (
	"time"

	artifactapp "github.com/story-engine/main-service/internal/application/world/artifact"
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

// ArtifactReferenceDTOToProto converts an ArtifactReferenceDTO (compatibility layer) to a protobuf message
func ArtifactReferenceDTOToProto(dto *artifactapp.ArtifactReferenceDTO) *artifactpb.ArtifactReference {
	if dto == nil {
		return nil
	}

	pb := &artifactpb.ArtifactReference{
		Id:         dto.ID.String(),
		ArtifactId: dto.ArtifactID.String(),
		EntityType: dto.EntityType,
		EntityId:   dto.EntityID.String(),
	}

	// Parse CreatedAt string to timestamp
	if dto.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, dto.CreatedAt); err == nil {
			pb.CreatedAt = timestamppb.New(t)
		}
	}

	return pb
}

