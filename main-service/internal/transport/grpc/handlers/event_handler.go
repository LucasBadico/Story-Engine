package handlers

import (
	"context"

	"github.com/google/uuid"
	eventapp "github.com/story-engine/main-service/internal/application/world/event"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	eventpb "github.com/story-engine/main-service/proto/event"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EventHandler implements the EventService gRPC service
type EventHandler struct {
	eventpb.UnimplementedEventServiceServer
	createEventUseCase      *eventapp.CreateEventUseCase
	getEventUseCase         *eventapp.GetEventUseCase
	listEventsUseCase       *eventapp.ListEventsUseCase
	updateEventUseCase      *eventapp.UpdateEventUseCase
	deleteEventUseCase      *eventapp.DeleteEventUseCase
	addReferenceUseCase     *eventapp.AddReferenceUseCase
	removeReferenceUseCase   *eventapp.RemoveReferenceUseCase
	getReferencesUseCase    *eventapp.GetReferencesUseCase
	updateReferenceUseCase  *eventapp.UpdateReferenceUseCase
	getChildrenUseCase      *eventapp.GetChildrenUseCase
	getAncestorsUseCase     *eventapp.GetAncestorsUseCase
	getDescendantsUseCase   *eventapp.GetDescendantsUseCase
	moveEventUseCase        *eventapp.MoveEventUseCase
	setEpochUseCase         *eventapp.SetEpochUseCase
	getEpochUseCase         *eventapp.GetEpochUseCase
	getTimelineUseCase      *eventapp.GetTimelineUseCase
	logger                  logger.Logger
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(
	createEventUseCase *eventapp.CreateEventUseCase,
	getEventUseCase *eventapp.GetEventUseCase,
	listEventsUseCase *eventapp.ListEventsUseCase,
	updateEventUseCase *eventapp.UpdateEventUseCase,
	deleteEventUseCase *eventapp.DeleteEventUseCase,
	addReferenceUseCase *eventapp.AddReferenceUseCase,
	removeReferenceUseCase *eventapp.RemoveReferenceUseCase,
	getReferencesUseCase *eventapp.GetReferencesUseCase,
	updateReferenceUseCase *eventapp.UpdateReferenceUseCase,
	getChildrenUseCase *eventapp.GetChildrenUseCase,
	getAncestorsUseCase *eventapp.GetAncestorsUseCase,
	getDescendantsUseCase *eventapp.GetDescendantsUseCase,
	moveEventUseCase *eventapp.MoveEventUseCase,
	setEpochUseCase *eventapp.SetEpochUseCase,
	getEpochUseCase *eventapp.GetEpochUseCase,
	getTimelineUseCase *eventapp.GetTimelineUseCase,
	logger logger.Logger,
) *EventHandler {
	return &EventHandler{
		createEventUseCase:     createEventUseCase,
		getEventUseCase:        getEventUseCase,
		listEventsUseCase:      listEventsUseCase,
		updateEventUseCase:     updateEventUseCase,
		deleteEventUseCase:     deleteEventUseCase,
		addReferenceUseCase:    addReferenceUseCase,
		removeReferenceUseCase: removeReferenceUseCase,
		getReferencesUseCase:   getReferencesUseCase,
		updateReferenceUseCase: updateReferenceUseCase,
		getChildrenUseCase:     getChildrenUseCase,
		getAncestorsUseCase:    getAncestorsUseCase,
		getDescendantsUseCase:  getDescendantsUseCase,
		moveEventUseCase:       moveEventUseCase,
		setEpochUseCase:        setEpochUseCase,
		getEpochUseCase:        getEpochUseCase,
		getTimelineUseCase:     getTimelineUseCase,
		logger:                 logger,
	}
}

// CreateEvent creates a new event
func (h *EventHandler) CreateEvent(ctx context.Context, req *eventpb.CreateEventRequest) (*eventpb.CreateEventResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

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

	var parentID *uuid.UUID
	if req.ParentId != nil && *req.ParentId != "" {
		parsedParentID, err := uuid.Parse(*req.ParentId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid parent_id: %v", err)
		}
		parentID = &parsedParentID
	}

	output, err := h.createEventUseCase.Execute(ctx, eventapp.CreateEventInput{
		TenantID:        tenantUUID,
		WorldID:         worldID,
		Name:            req.Name,
		Type:            eventType,
		Description:     description,
		Timeline:        timeline,
		Importance:      int(req.Importance),
		ParentID:        parentID,
		TimelinePosition: req.TimelinePosition,
		IsEpoch:         req.IsEpoch,
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
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	output, err := h.getEventUseCase.Execute(ctx, eventapp.GetEventInput{
		TenantID: tenantUUID,
		ID:       id,
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
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	worldID, err := uuid.Parse(req.WorldId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid world_id: %v", err)
	}

	output, err := h.listEventsUseCase.Execute(ctx, eventapp.ListEventsInput{
		TenantID: tenantUUID,
		WorldID:  worldID,
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

	var timelinePosition *float64
	if req.TimelinePosition != nil {
		timelinePosition = req.TimelinePosition
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	output, err := h.updateEventUseCase.Execute(ctx, eventapp.UpdateEventInput{
		TenantID:        tenantUUID,
		ID:              id,
		Name:            name,
		Type:            eventType,
		Description:     description,
		Timeline:        timeline,
		Importance:      importance,
		TimelinePosition: timelinePosition,
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
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	err = h.deleteEventUseCase.Execute(ctx, eventapp.DeleteEventInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.DeleteEventResponse{}, nil
}

// AddEventReference adds a reference to an event
func (h *EventHandler) AddEventReference(ctx context.Context, req *eventpb.AddEventReferenceRequest) (*eventpb.AddEventReferenceResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	var relationshipType *string
	if req.RelationshipType != nil {
		relationshipType = req.RelationshipType
	}

	err = h.addReferenceUseCase.Execute(ctx, eventapp.AddReferenceInput{
		TenantID:         tenantUUID,
		EventID:          eventID,
		EntityType:       req.EntityType,
		EntityID:         entityID,
		RelationshipType: relationshipType,
		Notes:            req.Notes,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the created relationship to return it
	output, err := h.getReferencesUseCase.Execute(ctx, eventapp.GetReferencesInput{
		TenantID: tenantUUID,
		EventID:  eventID,
	})
	if err != nil {
		return nil, err
	}

	// Find the relationship we just created
	for _, ref := range output.References {
		if ref.EntityType == req.EntityType && ref.EntityID == entityID {
			return &eventpb.AddEventReferenceResponse{
				Reference: mappers.EventReferenceToProto(ref),
			}, nil
		}
	}

	return nil, status.Errorf(codes.Internal, "failed to retrieve created relationship")
}

// RemoveEventReference removes a reference from an event
func (h *EventHandler) RemoveEventReference(ctx context.Context, req *eventpb.RemoveEventReferenceRequest) (*eventpb.RemoveEventReferenceResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	entityID, err := uuid.Parse(req.EntityId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid entity_id: %v", err)
	}

	err = h.removeReferenceUseCase.Execute(ctx, eventapp.RemoveReferenceInput{
		TenantID:   tenantUUID,
		EventID:    eventID,
		EntityType: req.EntityType,
		EntityID:   entityID,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.RemoveEventReferenceResponse{}, nil
}

// GetEventReferences lists references for an event
func (h *EventHandler) GetEventReferences(ctx context.Context, req *eventpb.GetEventReferencesRequest) (*eventpb.GetEventReferencesResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	output, err := h.getReferencesUseCase.Execute(ctx, eventapp.GetReferencesInput{
		TenantID: tenantUUID,
		EventID:  eventID,
	})
	if err != nil {
		return nil, err
	}

	references := make([]*eventpb.EventReference, len(output.References))
	for i, ref := range output.References {
		references[i] = mappers.EventReferenceToProto(ref)
	}

	return &eventpb.GetEventReferencesResponse{
		References:  references,
		TotalCount:  int32(len(references)),
	}, nil
}

// UpdateEventReference updates an event reference
func (h *EventHandler) UpdateEventReference(ctx context.Context, req *eventpb.UpdateEventReferenceRequest) (*eventpb.UpdateEventReferenceResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid id: %v", err)
	}

	var relationshipType *string
	if req.RelationshipType != nil {
		relationshipType = req.RelationshipType
	}

	var notes *string
	if req.Notes != nil {
		notes = req.Notes
	}

	err = h.updateReferenceUseCase.Execute(ctx, eventapp.UpdateReferenceInput{
		TenantID:         tenantUUID,
		ID:               id,
		RelationshipType: relationshipType,
		Notes:            notes,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the updated reference to return it
	// Note: We'd need a GetByID use case for this, but for now we'll just return success
	return &eventpb.UpdateEventReferenceResponse{}, nil
}

// GetEventChildren retrieves direct children of an event
func (h *EventHandler) GetEventChildren(ctx context.Context, req *eventpb.GetEventChildrenRequest) (*eventpb.GetEventChildrenResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	parentID, err := uuid.Parse(req.ParentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid parent_id: %v", err)
	}

	output, err := h.getChildrenUseCase.Execute(ctx, eventapp.GetChildrenInput{
		TenantID: tenantUUID,
		ParentID: parentID,
	})
	if err != nil {
		return nil, err
	}

	events := make([]*eventpb.Event, len(output.Events))
	for i, e := range output.Events {
		events[i] = mappers.EventToProto(e)
	}

	return &eventpb.GetEventChildrenResponse{
		Events: events,
	}, nil
}

// GetEventAncestors retrieves ancestors of an event
func (h *EventHandler) GetEventAncestors(ctx context.Context, req *eventpb.GetEventAncestorsRequest) (*eventpb.GetEventAncestorsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	output, err := h.getAncestorsUseCase.Execute(ctx, eventapp.GetAncestorsInput{
		TenantID: tenantUUID,
		EventID:  eventID,
	})
	if err != nil {
		return nil, err
	}

	events := make([]*eventpb.Event, len(output.Events))
	for i, e := range output.Events {
		events[i] = mappers.EventToProto(e)
	}

	return &eventpb.GetEventAncestorsResponse{
		Events: events,
	}, nil
}

// GetEventDescendants retrieves descendants of an event
func (h *EventHandler) GetEventDescendants(ctx context.Context, req *eventpb.GetEventDescendantsRequest) (*eventpb.GetEventDescendantsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	output, err := h.getDescendantsUseCase.Execute(ctx, eventapp.GetDescendantsInput{
		TenantID: tenantUUID,
		EventID:  eventID,
	})
	if err != nil {
		return nil, err
	}

	events := make([]*eventpb.Event, len(output.Events))
	for i, e := range output.Events {
		events[i] = mappers.EventToProto(e)
	}

	return &eventpb.GetEventDescendantsResponse{
		Events: events,
	}, nil
}

// MoveEvent moves an event to another parent
func (h *EventHandler) MoveEvent(ctx context.Context, req *eventpb.MoveEventRequest) (*eventpb.MoveEventResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	var parentID *uuid.UUID
	if req.ParentId != nil && *req.ParentId != "" {
		parsedParentID, err := uuid.Parse(*req.ParentId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid parent_id: %v", err)
		}
		parentID = &parsedParentID
	}

	err = h.moveEventUseCase.Execute(ctx, eventapp.MoveEventInput{
		TenantID: tenantUUID,
		EventID:  eventID,
		ParentID: parentID,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the updated event to return it
	getOutput, err := h.getEventUseCase.Execute(ctx, eventapp.GetEventInput{
		TenantID: tenantUUID,
		ID:       eventID,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.MoveEventResponse{
		Event: mappers.EventToProto(getOutput.Event),
	}, nil
}

// SetEventEpoch sets an event as epoch
func (h *EventHandler) SetEventEpoch(ctx context.Context, req *eventpb.SetEventEpochRequest) (*eventpb.SetEventEpochResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	eventID, err := uuid.Parse(req.EventId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid event_id: %v", err)
	}

	err = h.setEpochUseCase.Execute(ctx, eventapp.SetEpochInput{
		TenantID: tenantUUID,
		EventID:  eventID,
	})
	if err != nil {
		return nil, err
	}

	// Fetch the updated event to return it
	getOutput, err := h.getEventUseCase.Execute(ctx, eventapp.GetEventInput{
		TenantID: tenantUUID,
		ID:       eventID,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.SetEventEpochResponse{
		Event: mappers.EventToProto(getOutput.Event),
	}, nil
}

// GetEventEpoch retrieves the epoch event for a world
func (h *EventHandler) GetEventEpoch(ctx context.Context, req *eventpb.GetEventEpochRequest) (*eventpb.GetEventEpochResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	worldID, err := uuid.Parse(req.WorldId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid world_id: %v", err)
	}

	output, err := h.getEpochUseCase.Execute(ctx, eventapp.GetEpochInput{
		TenantID: tenantUUID,
		WorldID:  worldID,
	})
	if err != nil {
		return nil, err
	}

	return &eventpb.GetEventEpochResponse{
		Event: mappers.EventToProto(output.Event),
	}, nil
}

// GetTimeline retrieves events ordered by timeline position
func (h *EventHandler) GetTimeline(ctx context.Context, req *eventpb.GetTimelineRequest) (*eventpb.GetTimelineResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	worldID, err := uuid.Parse(req.WorldId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid world_id: %v", err)
	}

	var fromPos *float64
	if req.FromPos != nil {
		fromPos = req.FromPos
	}

	var toPos *float64
	if req.ToPos != nil {
		toPos = req.ToPos
	}

	output, err := h.getTimelineUseCase.Execute(ctx, eventapp.GetTimelineInput{
		TenantID: tenantUUID,
		WorldID:  worldID,
		FromPos:  fromPos,
		ToPos:    toPos,
	})
	if err != nil {
		return nil, err
	}

	events := make([]*eventpb.Event, len(output.Events))
	for i, e := range output.Events {
		events[i] = mappers.EventToProto(e)
	}

	return &eventpb.GetTimelineResponse{
		Events: events,
	}, nil
}

