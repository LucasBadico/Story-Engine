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
		Id:         e.ID.String(),
		WorldId:    e.WorldID.String(),
		Name:       e.Name,
		Importance: int32(e.Importance),
		CreatedAt:  timestamppb.New(e.CreatedAt),
		UpdatedAt:  timestamppb.New(e.UpdatedAt),
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

	return pb
}

// EventCharacterToProto converts an event-character relationship to a protobuf message
func EventCharacterToProto(ec *world.EventCharacter) *eventpb.EventCharacter {
	if ec == nil {
		return nil
	}

	pb := &eventpb.EventCharacter{
		Id:          ec.ID.String(),
		EventId:     ec.EventID.String(),
		CharacterId: ec.CharacterID.String(),
		CreatedAt:   timestamppb.New(ec.CreatedAt),
	}

	if ec.Role != nil {
		pb.Role = ec.Role
	}

	return pb
}

// EventLocationToProto converts an event-location relationship to a protobuf message
func EventLocationToProto(el *world.EventLocation) *eventpb.EventLocation {
	if el == nil {
		return nil
	}

	pb := &eventpb.EventLocation{
		Id:         el.ID.String(),
		EventId:    el.EventID.String(),
		LocationId: el.LocationID.String(),
		CreatedAt:  timestamppb.New(el.CreatedAt),
	}

	if el.Significance != nil {
		pb.Significance = el.Significance
	}

	return pb
}

// EventArtifactToProto converts an event-artifact relationship to a protobuf message
func EventArtifactToProto(ea *world.EventArtifact) *eventpb.EventArtifact {
	if ea == nil {
		return nil
	}

	pb := &eventpb.EventArtifact{
		Id:         ea.ID.String(),
		EventId:    ea.EventID.String(),
		ArtifactId: ea.ArtifactID.String(),
		CreatedAt:  timestamppb.New(ea.CreatedAt),
	}

	if ea.Role != nil {
		pb.Role = ea.Role
	}

	return pb
}

