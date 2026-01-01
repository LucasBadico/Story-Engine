package handlers

import (
	"context"

	"github.com/google/uuid"
	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	eventpb "github.com/story-engine/main-service/proto/event"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EventHandler implements the EventService gRPC service
type EventHandler struct {
	eventpb.UnimplementedEventServiceServer
	createEventUseCase         *eventapp.CreateEventUseCase
	getEventUseCase            *eventapp.GetEventUseCase
	listEventsUseCase          *eventapp.ListEventsUseCase
	updateEventUseCase         *eventapp.UpdateEventUseCase
	deleteEventUseCase         *eventapp.DeleteEventUseCase
	addCharacterUseCase        *eventapp.AddCharacterToEventUseCase
	removeCharacterUseCase     *eventapp.RemoveCharacterFromEventUseCase
	getCharactersUseCase       *eventapp.GetEventCharactersUseCase
	addLocationUseCase         *eventapp.AddLocationToEventUseCase
	removeLocationUseCase      *eventapp.RemoveLocationFromEventUseCase
	getLocationsUseCase        *eventapp.GetEventLocationsUseCase
	addArtifactUseCase         *eventapp.AddArtifactToEventUseCase
	removeArtifactUseCase      *eventapp.RemoveArtifactFromEventUseCase
	getArtifactsUseCase        *eventapp.GetEventArtifactsUseCase
	logger                     logger.Logger
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(
	createEventUseCase *eventapp.CreateEventUseCase,
	getEventUseCase *eventapp.GetEventUseCase,
	listEventsUseCase *eventapp.ListEventsUseCase,
	updateEventUseCase *eventapp.UpdateEventUseCase,
	deleteEventUseCase *eventapp.DeleteEventUseCase,
	addCharacterUseCase *eventapp.AddCharacterToEventUseCase,
	removeCharacterUseCase *eventapp.RemoveCharacterFromEventUseCase,
	getCharactersUseCase *eventapp.GetEventCharactersUseCase,
	addLocationUseCase *eventapp.AddLocationToEventUseCase,
	removeLocationUseCase *eventapp.RemoveLocationFromEventUseCase,
	getLocationsUseCase *eventapp.GetEventLocationsUseCase,
	addArtifactUseCase *eventapp.AddArtifactToEventUseCase,
	removeArtifactUseCase *eventapp.RemoveArtifactFromEventUseCase,
	getArtifactsUseCase *eventapp.GetEventArtifactsUseCase,
	logger logger.Logger,
) *EventHandler {
	return &EventHandler{
		createEventUseCase:     createEventUseCase,
		getEventUseCase:        getEventUseCase,
		listEventsUseCase:      listEventsUseCase,
		updateEventUseCase:     updateEventUseCase,
		deleteEventUseCase:     deleteEventUseCase,
		addCharacterUseCase:    addCharacterUseCase,
		removeCharacterUseCase: removeCharacterUseCase,
		getCharactersUseCase:   getCharactersUseCase,
		addLocationUseCase:     addLocationUseCase,
		removeLocationUseCase:  removeLocationUseCase,
		getLocationsUseCase:    getLocationsUseCase,
		addArtifactUseCase:     addArtifactUseCase,
		removeArtifactUseCase:  removeArtifactUseCase,
		getArtifactsUseCase:    getArtifactsUseCase,
		logger:                 logger,
	}
}

// CreateEvent creates a new event
func (h *EventHandler) CreateEvent(ctx context.Context, req *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error) {
	worldID, err := uuid.Parse(req.WorldId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid world_id: %v", err)
	}

	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}

	var eventType *string
	if req.Type != nil {
		eventType = req.Type
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var timeline *string
	if req.Timeline != nil {
		timeline = req.Timeline
	}

	output, err := h.createEventUseCase.Execute(ctx, eventapp.CreateEventInput{
		WorldID:     worldID,
		Name:        req.Name,
		Type:        eventType,
		Description: description,
		Timeline:    timeline,
		Importance:  int(req.Importance),
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.CreateEventResponse{
		Event: mappers.EventToProto(output.Event),
	}, nil
}

// GetEvent retrieves an event by ID
func (h *EventHandler) GetEvent(ctx context.Context, req *eventpb.GetEventRequest) (*eventpb.GetEventResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	output, err := h.getEventUseCase.Execute(ctx, eventapp.GetEventInput{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.GetEventResponse{
		Event: mappers.EventToProto(output.Event),
	}, nil
}

// ListEvents lists events for a world
func (h *EventHandler) ListEvents(ctx context.Context, req *eventpb.ListEventsRequest) (*eventpb.ListEventsResponse, error) {
	worldID, err := uuid.Parse(req.WorldId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid world_id: %v", err)
	}

	output, err := h.listEventsUseCase.Execute(ctx, eventapp.ListEventsInput{
		WorldID: worldID,
	})
	if err != nil {
		return nil, err
	}

	protoEvents := make([]*eventpb.Event, len(output.Events))
	for i, e := range output.Events {
		protoEvents[i] = mappers.EventToProto(e)
	}

	return &eventpb.ListEventsResponse{
		Events:     protoEvents,
		TotalCount: int32(len(protoEvents)),
	}, nil
}

// UpdateEvent updates an existing event
func (h *EventHandler) UpdateEvent(ctx context.Context, req *eventpb.UpdateEventRequest) (*eventpb.UpdateEventResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var eventType *string
	if req.Type != nil {
		eventType = req.Type
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var timeline *string
	if req.Timeline != nil {
		timeline = req.Timeline
	}

	var importance *int
	if req.Importance != nil {
		imp := int(*req.Importance)
		importance = &imp
	}

	output, err := h.updateEventUseCase.Execute(ctx, eventapp.UpdateEventInput{
		ID:          id,
		Name:        name,
		Type:        eventType,
		Description: description,
		Timeline:    timeline,
		Importance:  importance,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.UpdateEventResponse{
		Event: mappers.EventToProto(output.Event),
	}, nil
}

// DeleteEvent deletes an event
func (h *EventHandler) DeleteEvent(ctx context.Context, req *eventpb.DeleteEventRequest) (*eventpb.DeleteEventResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	err = h.deleteEventUseCase.Execute(ctx, eventapp.DeleteEventInput{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.DeleteEventResponse{}, nil
}

// AddCharacterToEvent adds a character to an event
func (h *EventHandler) AddCharacterToEvent(ctx context.Context, req *eventpb.AddCharacterToEventRequest) (*eventpb.AddCharacterToEventResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	var role *string
	if req.Role != nil {
		role = req.Role
	}

	err = h.addCharacterUseCase.Execute(ctx, eventapp.AddCharacterToEventInput{
		EventID:     eventID,
		CharacterID: characterID,
		Role:        role,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the created relationship to return it
	output, err := h.getCharactersUseCase.Execute(ctx, eventapp.GetEventCharactersInput{
		EventID: eventID,
	})
	if err != nil {
		return nil, err
	}

	// Find the relationship we just created
	for _, ec := range output.Characters {
		if ec.CharacterID == characterID {
			return &eventpb.AddCharacterToEventResponse{
				EventCharacter: mappers.EventCharacterToProto(ec),
			}, nil
		}
	}

	return nil, status.Errorf(codes.Internal, "failed to retrieve created relationship")
}

// RemoveCharacterFromEvent removes a character from an event
func (h *EventHandler) RemoveCharacterFromEvent(ctx context.Context, req *eventpb.RemoveCharacterFromEventRequest) (*eventpb.RemoveCharacterFromEventResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	err = h.removeCharacterUseCase.Execute(ctx, eventapp.RemoveCharacterFromEventInput{
		EventID:     eventID,
		CharacterID: characterID,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.RemoveCharacterFromEventResponse{}, nil
}

// GetEventCharacters lists characters for an event
func (h *EventHandler) GetEventCharacters(ctx context.Context, req *eventpb.GetEventCharactersRequest) (*eventpb.GetEventCharactersResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	output, err := h.getCharactersUseCase.Execute(ctx, eventapp.GetEventCharactersInput{
		EventID: eventID,
	})
	if err != nil {
		return nil, err
	}

	characters := make([]*eventpb.EventCharacter, len(output.Characters))
	for i, ec := range output.Characters {
		characters[i] = mappers.EventCharacterToProto(ec)
	}

	return &eventpb.GetEventCharactersResponse{
		Characters:  characters,
		TotalCount:  int32(len(characters)),
	}, nil
}

// AddLocationToEvent adds a location to an event
func (h *EventHandler) AddLocationToEvent(ctx context.Context, req *eventpb.AddLocationToEventRequest) (*eventpb.AddLocationToEventResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	locationID, err := uuid.Parse(req.LocationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
	}

	var significance *string
	if req.Significance != nil {
		significance = req.Significance
	}

	err = h.addLocationUseCase.Execute(ctx, eventapp.AddLocationToEventInput{
		EventID:      eventID,
		LocationID:   locationID,
		Significance: significance,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the created relationship to return it
	output, err := h.getLocationsUseCase.Execute(ctx, eventapp.GetEventLocationsInput{
		EventID: eventID,
	})
	if err != nil {
		return nil, err
	}

	// Find the relationship we just created
	for _, el := range output.Locations {
		if el.LocationID == locationID {
			return &eventpb.AddLocationToEventResponse{
				EventLocation: mappers.EventLocationToProto(el),
			}, nil
		}
	}

	return nil, status.Errorf(codes.Internal, "failed to retrieve created relationship")
}

// RemoveLocationFromEvent removes a location from an event
func (h *EventHandler) RemoveLocationFromEvent(ctx context.Context, req *eventpb.RemoveLocationFromEventRequest) (*eventpb.RemoveLocationFromEventResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	locationID, err := uuid.Parse(req.LocationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid location_id: %v", err)
	}

	err = h.removeLocationUseCase.Execute(ctx, eventapp.RemoveLocationFromEventInput{
		EventID:    eventID,
		LocationID: locationID,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.RemoveLocationFromEventResponse{}, nil
}

// GetEventLocations lists locations for an event
func (h *EventHandler) GetEventLocations(ctx context.Context, req *eventpb.GetEventLocationsRequest) (*eventpb.GetEventLocationsResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	output, err := h.getLocationsUseCase.Execute(ctx, eventapp.GetEventLocationsInput{
		EventID: eventID,
	})
	if err != nil {
		return nil, err
	}

	locations := make([]*eventpb.EventLocation, len(output.Locations))
	for i, el := range output.Locations {
		locations[i] = mappers.EventLocationToProto(el)
	}

	return &eventpb.GetEventLocationsResponse{
		Locations:   locations,
		TotalCount:  int32(len(locations)),
	}, nil
}

// AddArtifactToEvent adds an artifact to an event
func (h *EventHandler) AddArtifactToEvent(ctx context.Context, req *eventpb.AddArtifactToEventRequest) (*eventpb.AddArtifactToEventResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	artifactID, err := uuid.Parse(req.ArtifactId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact_id: %v", err)
	}

	var role *string
	if req.Role != nil {
		role = req.Role
	}

	err = h.addArtifactUseCase.Execute(ctx, eventapp.AddArtifactToEventInput{
		EventID:    eventID,
		ArtifactID: artifactID,
		Role:       role,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the created relationship to return it
	output, err := h.getArtifactsUseCase.Execute(ctx, eventapp.GetEventArtifactsInput{
		EventID: eventID,
	})
	if err != nil {
		return nil, err
	}

	// Find the relationship we just created
	for _, ea := range output.Artifacts {
		if ea.ArtifactID == artifactID {
			return &eventpb.AddArtifactToEventResponse{
				EventArtifact: mappers.EventArtifactToProto(ea),
			}, nil
		}
	}

	return nil, status.Errorf(codes.Internal, "failed to retrieve created relationship")
}

// RemoveArtifactFromEvent removes an artifact from an event
func (h *EventHandler) RemoveArtifactFromEvent(ctx context.Context, req *eventpb.RemoveArtifactFromEventRequest) (*eventpb.RemoveArtifactFromEventResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	artifactID, err := uuid.Parse(req.ArtifactId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid artifact_id: %v", err)
	}

	err = h.removeArtifactUseCase.Execute(ctx, eventapp.RemoveArtifactFromEventInput{
		EventID:    eventID,
		ArtifactID: artifactID,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.RemoveArtifactFromEventResponse{}, nil
}

// GetEventArtifacts lists artifacts for an event
func (h *EventHandler) GetEventArtifacts(ctx context.Context, req *eventpb.GetEventArtifactsRequest) (*eventpb.GetEventArtifactsResponse, error) {
	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	output, err := h.getArtifactsUseCase.Execute(ctx, eventapp.GetEventArtifactsInput{
		EventID: eventID,
	})
	if err != nil {
		return nil, err
	}

	artifacts := make([]*eventpb.EventArtifact, len(output.Artifacts))
	for i, ea := range output.Artifacts {
		artifacts[i] = mappers.EventArtifactToProto(ea)
	}

	return &eventpb.GetEventArtifactsResponse{
		Artifacts:   artifacts,
		TotalCount:  int32(len(artifacts)),
	}, nil
}

