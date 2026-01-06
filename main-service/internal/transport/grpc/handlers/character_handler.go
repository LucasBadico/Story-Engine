package handlers

import (
	"context"

	"github.com/google/uuid"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	characterrelationshipapp "github.com/story-engine/main-service/internal/application/world/character_relationship"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/grpcctx"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	characterpb "github.com/story-engine/main-service/proto/character"
	eventpb "github.com/story-engine/main-service/proto/event"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CharacterHandler implements the CharacterService gRPC service
type CharacterHandler struct {
	characterpb.UnimplementedCharacterServiceServer
	createCharacterUseCase      *characterapp.CreateCharacterUseCase
	getCharacterUseCase         *characterapp.GetCharacterUseCase
	listCharactersUseCase        *characterapp.ListCharactersUseCase
	updateCharacterUseCase       *characterapp.UpdateCharacterUseCase
	deleteCharacterUseCase       *characterapp.DeleteCharacterUseCase
	getCharacterTraitsUseCase    *characterapp.GetCharacterTraitsUseCase
	getCharacterEventsUseCase   *characterapp.GetCharacterEventsUseCase
	addTraitToCharacterUseCase   *characterapp.AddTraitToCharacterUseCase
	updateCharacterTraitUseCase  *characterapp.UpdateCharacterTraitUseCase
	removeTraitFromCharacterUseCase *characterapp.RemoveTraitFromCharacterUseCase
	createCharacterRelationshipUseCase *characterrelationshipapp.CreateCharacterRelationshipUseCase
	getCharacterRelationshipUseCase *characterrelationshipapp.GetCharacterRelationshipUseCase
	listCharacterRelationshipsUseCase *characterrelationshipapp.ListCharacterRelationshipsUseCase
	updateCharacterRelationshipUseCase *characterrelationshipapp.UpdateCharacterRelationshipUseCase
	deleteCharacterRelationshipUseCase *characterrelationshipapp.DeleteCharacterRelationshipUseCase
	logger                       logger.Logger
}

// NewCharacterHandler creates a new CharacterHandler
func NewCharacterHandler(
	createCharacterUseCase *characterapp.CreateCharacterUseCase,
	getCharacterUseCase *characterapp.GetCharacterUseCase,
	listCharactersUseCase *characterapp.ListCharactersUseCase,
	updateCharacterUseCase *characterapp.UpdateCharacterUseCase,
	deleteCharacterUseCase *characterapp.DeleteCharacterUseCase,
	getCharacterTraitsUseCase *characterapp.GetCharacterTraitsUseCase,
	getCharacterEventsUseCase *characterapp.GetCharacterEventsUseCase,
	addTraitToCharacterUseCase *characterapp.AddTraitToCharacterUseCase,
	updateCharacterTraitUseCase *characterapp.UpdateCharacterTraitUseCase,
	removeTraitFromCharacterUseCase *characterapp.RemoveTraitFromCharacterUseCase,
	createCharacterRelationshipUseCase *characterrelationshipapp.CreateCharacterRelationshipUseCase,
	getCharacterRelationshipUseCase *characterrelationshipapp.GetCharacterRelationshipUseCase,
	listCharacterRelationshipsUseCase *characterrelationshipapp.ListCharacterRelationshipsUseCase,
	updateCharacterRelationshipUseCase *characterrelationshipapp.UpdateCharacterRelationshipUseCase,
	deleteCharacterRelationshipUseCase *characterrelationshipapp.DeleteCharacterRelationshipUseCase,
	logger logger.Logger,
) *CharacterHandler {
	return &CharacterHandler{
		createCharacterUseCase:      createCharacterUseCase,
		getCharacterUseCase:         getCharacterUseCase,
		listCharactersUseCase:       listCharactersUseCase,
		updateCharacterUseCase:      updateCharacterUseCase,
		deleteCharacterUseCase:      deleteCharacterUseCase,
		getCharacterTraitsUseCase:   getCharacterTraitsUseCase,
		getCharacterEventsUseCase:   getCharacterEventsUseCase,
		addTraitToCharacterUseCase:  addTraitToCharacterUseCase,
		updateCharacterTraitUseCase: updateCharacterTraitUseCase,
		removeTraitFromCharacterUseCase: removeTraitFromCharacterUseCase,
		createCharacterRelationshipUseCase: createCharacterRelationshipUseCase,
		getCharacterRelationshipUseCase: getCharacterRelationshipUseCase,
		listCharacterRelationshipsUseCase: listCharacterRelationshipsUseCase,
		updateCharacterRelationshipUseCase: updateCharacterRelationshipUseCase,
		deleteCharacterRelationshipUseCase: deleteCharacterRelationshipUseCase,
		logger:                      logger,
	}
}

// CreateCharacter creates a new character
func (h *CharacterHandler) CreateCharacter(ctx context.Context, req *characterpb.CreateCharacterRequest) (*characterpb.CreateCharacterResponse, error) {
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

	var archetypeID *uuid.UUID
	if req.ArchetypeId != "" {
		aid, err := uuid.Parse(req.ArchetypeId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid archetype_id: %v", err)
		}
		archetypeID = &aid
	}

	input := characterapp.CreateCharacterInput{
		TenantID:    tenantUUID,
		WorldID:     worldID,
		ArchetypeID: archetypeID,
		Name:        req.Name,
		Description: req.Description,
	}

	output, err := h.createCharacterUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &characterpb.CreateCharacterResponse{
		Character: mappers.CharacterToProto(output.Character),
	}, nil
}

// GetCharacter retrieves a character by ID
func (h *CharacterHandler) GetCharacter(ctx context.Context, req *characterpb.GetCharacterRequest) (*characterpb.GetCharacterResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid character id: %v", err)
	}

	output, err := h.getCharacterUseCase.Execute(ctx, characterapp.GetCharacterInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.GetCharacterResponse{
		Character: mappers.CharacterToProto(output.Character),
	}, nil
}

// ListCharacters lists characters for a world
func (h *CharacterHandler) ListCharacters(ctx context.Context, req *characterpb.ListCharactersRequest) (*characterpb.ListCharactersResponse, error) {
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

	limit := 50
	offset := 0
	if req.Pagination != nil {
		if req.Pagination.Limit > 0 {
			limit = int(req.Pagination.Limit)
		}
		if req.Pagination.Offset > 0 {
			offset = int(req.Pagination.Offset)
		}
	}

	output, err := h.listCharactersUseCase.Execute(ctx, characterapp.ListCharactersInput{
		TenantID: tenantUUID,
		WorldID:  worldID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	characters := make([]*characterpb.Character, len(output.Characters))
	for i, c := range output.Characters {
		characters[i] = mappers.CharacterToProto(c)
	}

	return &characterpb.ListCharactersResponse{
		Characters: characters,
		TotalCount: int32(output.Total),
	}, nil
}

// UpdateCharacter updates an existing character
func (h *CharacterHandler) UpdateCharacter(ctx context.Context, req *characterpb.UpdateCharacterRequest) (*characterpb.UpdateCharacterResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character id: %v", err)
	}

	var name *string
	if req.Name != nil {
		name = req.Name
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var archetypeID *uuid.UUID
	if req.ArchetypeId != nil {
		if *req.ArchetypeId == "" {
			archetypeID = nil
		} else {
			aid, err := uuid.Parse(*req.ArchetypeId)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid archetype_id: %v", err)
			}
			archetypeID = &aid
		}
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	input := characterapp.UpdateCharacterInput{
		TenantID:    tenantUUID,
		ID:          id,
		Name:        name,
		Description: description,
		ArchetypeID: archetypeID,
	}

	output, err := h.updateCharacterUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &characterpb.UpdateCharacterResponse{
		Character: mappers.CharacterToProto(output.Character),
	}, nil
}

// DeleteCharacter deletes a character
func (h *CharacterHandler) DeleteCharacter(ctx context.Context, req *characterpb.DeleteCharacterRequest) (*characterpb.DeleteCharacterResponse, error) {
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
		return nil, status.Errorf(codes.InvalidArgument, "invalid character id: %v", err)
	}

	err = h.deleteCharacterUseCase.Execute(ctx, characterapp.DeleteCharacterInput{
		TenantID: tenantUUID,
		ID:       id,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.DeleteCharacterResponse{}, nil
}

// GetCharacterTraits retrieves all traits for a character
func (h *CharacterHandler) GetCharacterTraits(ctx context.Context, req *characterpb.GetCharacterTraitsRequest) (*characterpb.GetCharacterTraitsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	output, err := h.getCharacterTraitsUseCase.Execute(ctx, characterapp.GetCharacterTraitsInput{
		TenantID:    tenantUUID,
		CharacterID: characterID,
	})
	if err != nil {
		return nil, err
	}

	traits := make([]*characterpb.CharacterTrait, len(output.Traits))
	for i, ct := range output.Traits {
		traits[i] = mappers.CharacterTraitToProto(ct)
	}

	return &characterpb.GetCharacterTraitsResponse{
		Traits: traits,
	}, nil
}

// AddTraitToCharacter adds a trait to a character
func (h *CharacterHandler) AddTraitToCharacter(ctx context.Context, req *characterpb.AddTraitToCharacterRequest) (*characterpb.AddTraitToCharacterResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	traitID, err := uuid.Parse(req.TraitId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait_id: %v", err)
	}

	err = h.addTraitToCharacterUseCase.Execute(ctx, characterapp.AddTraitToCharacterInput{
		TenantID:    tenantUUID,
		CharacterID: characterID,
		TraitID:     traitID,
		Value:       req.Value,
		Notes:       req.Notes,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.AddTraitToCharacterResponse{}, nil
}

// UpdateCharacterTrait updates a character trait
func (h *CharacterHandler) UpdateCharacterTrait(ctx context.Context, req *characterpb.UpdateCharacterTraitRequest) (*characterpb.UpdateCharacterTraitResponse, error) {
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	traitID, err := uuid.Parse(req.TraitId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait_id: %v", err)
	}

	var value *string
	if req.Value != nil {
		value = req.Value
	}

	var notes *string
	if req.Notes != nil {
		notes = req.Notes
	}

	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	output, err := h.updateCharacterTraitUseCase.Execute(ctx, characterapp.UpdateCharacterTraitInput{
		TenantID:    tenantUUID,
		CharacterID: characterID,
		TraitID:     traitID,
		Value:       value,
		Notes:       notes,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.UpdateCharacterTraitResponse{
		CharacterTrait: mappers.CharacterTraitToProto(output.CharacterTrait),
	}, nil
}

// RemoveTraitFromCharacter removes a trait from a character
func (h *CharacterHandler) RemoveTraitFromCharacter(ctx context.Context, req *characterpb.RemoveTraitFromCharacterRequest) (*characterpb.RemoveTraitFromCharacterResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	traitID, err := uuid.Parse(req.TraitId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait_id: %v", err)
	}

	err = h.removeTraitFromCharacterUseCase.Execute(ctx, characterapp.RemoveTraitFromCharacterInput{
		TenantID:    tenantUUID,
		CharacterID: characterID,
		TraitID:     traitID,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.RemoveTraitFromCharacterResponse{}, nil
}

// GetCharacterEvents retrieves all events for a character
func (h *CharacterHandler) GetCharacterEvents(ctx context.Context, req *characterpb.GetCharacterEventsRequest) (*characterpb.GetCharacterEventsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	output, err := h.getCharacterEventsUseCase.Execute(ctx, characterapp.GetCharacterEventsInput{
		TenantID:    tenantUUID,
		CharacterID: characterID,
	})
	if err != nil {
		return nil, err
	}

	// TODO: Improve by fetching full Event data from repository
	// Currently returns Events with only ID populated from EventReference
	events := make([]*eventpb.Event, len(output.EventReferences))
	for i, ref := range output.EventReferences {
		events[i] = &eventpb.Event{
			Id: ref.EventID.String(),
		}
	}

	return &characterpb.GetCharacterEventsResponse{
		Events: events,
	}, nil
}

// CreateCharacterRelationship creates a new relationship between two characters
func (h *CharacterHandler) CreateCharacterRelationship(ctx context.Context, req *characterpb.CreateCharacterRelationshipRequest) (*characterpb.CreateCharacterRelationshipResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	character1ID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	character2ID, err := uuid.Parse(req.OtherCharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid other_character_id: %v", err)
	}

	output, err := h.createCharacterRelationshipUseCase.Execute(ctx, characterrelationshipapp.CreateCharacterRelationshipInput{
		TenantID:         tenantUUID,
		Character1ID:     character1ID,
		Character2ID:     character2ID,
		RelationshipType: req.RelationshipType,
		Description:      req.Description,
		Bidirectional:    req.Bidirectional,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.CreateCharacterRelationshipResponse{
		Relationship: mappers.CharacterRelationshipToProto(output.Relationship),
	}, nil
}

// GetCharacterRelationship retrieves a character relationship by ID
func (h *CharacterHandler) GetCharacterRelationship(ctx context.Context, req *characterpb.GetCharacterRelationshipRequest) (*characterpb.GetCharacterRelationshipResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	relationshipID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid relationship id: %v", err)
	}

	output, err := h.getCharacterRelationshipUseCase.Execute(ctx, characterrelationshipapp.GetCharacterRelationshipInput{
		TenantID:       tenantUUID,
		ID:             relationshipID,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.GetCharacterRelationshipResponse{
		Relationship: mappers.CharacterRelationshipToProto(output.Relationship),
	}, nil
}

// ListCharacterRelationships retrieves all relationships for a character
func (h *CharacterHandler) ListCharacterRelationships(ctx context.Context, req *characterpb.ListCharacterRelationshipsRequest) (*characterpb.ListCharacterRelationshipsResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	output, err := h.listCharacterRelationshipsUseCase.Execute(ctx, characterrelationshipapp.ListCharacterRelationshipsInput{
		TenantID:    tenantUUID,
		CharacterID: characterID,
	})
	if err != nil {
		return nil, err
	}

	relationships := make([]*characterpb.CharacterRelationship, len(output.Relationships))
	for i, r := range output.Relationships {
		relationships[i] = mappers.CharacterRelationshipToProto(r)
	}

	return &characterpb.ListCharacterRelationshipsResponse{
		Relationships: relationships,
	}, nil
}

// UpdateCharacterRelationship updates a character relationship
func (h *CharacterHandler) UpdateCharacterRelationship(ctx context.Context, req *characterpb.UpdateCharacterRelationshipRequest) (*characterpb.UpdateCharacterRelationshipResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	relationshipID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid relationship id: %v", err)
	}

	var relationshipType *string
	if req.RelationshipType != nil {
		relationshipType = req.RelationshipType
	}

	var description *string
	if req.Description != nil {
		description = req.Description
	}

	var bidirectional *bool
	if req.Bidirectional != nil {
		bidirectional = req.Bidirectional
	}

	output, err := h.updateCharacterRelationshipUseCase.Execute(ctx, characterrelationshipapp.UpdateCharacterRelationshipInput{
		TenantID:         tenantUUID,
		ID:               relationshipID,
		RelationshipType: relationshipType,
		Description:      description,
		Bidirectional:    bidirectional,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.UpdateCharacterRelationshipResponse{
		Relationship: mappers.CharacterRelationshipToProto(output.Relationship),
	}, nil
}

// DeleteCharacterRelationship deletes a character relationship
func (h *CharacterHandler) DeleteCharacterRelationship(ctx context.Context, req *characterpb.DeleteCharacterRelationshipRequest) (*characterpb.DeleteCharacterRelationshipResponse, error) {
	tenantID, ok := grpcctx.TenantIDFromContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	relationshipID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid relationship id: %v", err)
	}

	err = h.deleteCharacterRelationshipUseCase.Execute(ctx, characterrelationshipapp.DeleteCharacterRelationshipInput{
		TenantID: tenantUUID,
		ID:       relationshipID,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.DeleteCharacterRelationshipResponse{}, nil
}


