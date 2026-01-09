package mappers

import (
	"time"

	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	"github.com/story-engine/main-service/internal/core/world"
	eventpb "github.com/story-engine/main-service/proto/event"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EventToProto converts an event domain entity to a protobuf message
func EventToProto(e *world.Event) *eventpb.Event {
	if e == nil {
		return nil
	}

	pb := &eventpb.Event{
		Id:               e.ID.String(),
		WorldId:          e.WorldID.String(),
		Name:             e.Name,
		Importance:       int32(e.Importance),
		HierarchyLevel:   int32(e.HierarchyLevel),
		TimelinePosition: e.TimelinePosition,
		IsEpoch:          e.IsEpoch,
		CreatedAt:        timestamppb.New(e.CreatedAt),
		UpdatedAt:        timestamppb.New(e.UpdatedAt),
	}

	if e.Type != nil {
		pb.Type = e.Type
	}
	if e.Description != nil {
		pb.Description = e.Description
	}
	if e.Timeline != nil {
		pb.Timeline = e.Timeline
	}
	if e.ParentID != nil {
		parentIDStr := e.ParentID.String()
		pb.ParentId = &parentIDStr
	}

	return pb
}

// EventReferenceDTOToProto converts an EventReferenceDTO (compatibility layer) to a protobuf message
func EventReferenceDTOToProto(dto *eventapp.EventReferenceDTO) *eventpb.EventReference {
	if dto == nil {
		return nil
	}

	pb := &eventpb.EventReference{
		Id:         dto.ID.String(),
		EventId:    dto.EventID.String(),
		EntityType: dto.EntityType,
		EntityId:   dto.EntityID.String(),
		Notes:      dto.Notes,
	}

	if dto.RelationshipType != nil {
		pb.RelationshipType = dto.RelationshipType
	}

	// Parse CreatedAt string to timestamp
	if dto.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, dto.CreatedAt); err == nil {
			pb.CreatedAt = timestamppb.New(t)
		}
	}

	return pb
}

