package mappers

import (
	"github.com/story-engine/main-service/internal/core/story"
	scenepb "github.com/story-engine/main-service/proto/scene"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SceneReferenceToProto converts a scene reference domain entity to a protobuf message
func SceneReferenceToProto(ref *story.SceneReference) *scenepb.SceneReference {
	if ref == nil {
		return nil
	}

	return &scenepb.SceneReference{
		Id:         ref.ID.String(),
		SceneId:    ref.SceneID.String(),
		EntityType: string(ref.EntityType),
		EntityId:   ref.EntityID.String(),
		CreatedAt:  timestamppb.New(ref.CreatedAt),
	}
}

