package mappers

import (
	"time"

	sceneapp "github.com/story-engine/main-service/internal/application/story/scene"
	scenepb "github.com/story-engine/main-service/proto/scene"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SceneReferenceDTOToProto converts a SceneReferenceDTO (compatibility layer) to a protobuf message
func SceneReferenceDTOToProto(dto *sceneapp.SceneReferenceDTO) *scenepb.SceneReference {
	if dto == nil {
		return nil
	}

	pb := &scenepb.SceneReference{
		Id:         dto.ID.String(),
		SceneId:    dto.SceneID.String(),
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

