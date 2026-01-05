package mappers

import (
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

// EventReferenceToProto converts an event-reference relationship to a protobuf message
func EventReferenceToProto(er *world.EventReference) *eventpb.EventReference {
	if er == nil {
		return nil
	}

	pb := &eventpb.EventReference{
		Id:         er.ID.String(),
		EventId:    er.EventID.String(),
		EntityType: er.EntityType,
		EntityId:   er.EntityID.String(),
		Notes:      er.Notes,
		CreatedAt:  timestamppb.New(er.CreatedAt),
	}

	if er.RelationshipType != nil {
		pb.RelationshipType = er.RelationshipType
	}

	return pb
}


