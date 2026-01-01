package handlers

import (
	"context"

	"github.com/google/uuid"
	characterapp "github.com/story-engine/main-service/internal/application/world/character"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	characterpb "github.com/story-engine/main-service/proto/character"
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
	addTraitToCharacterUseCase   *characterapp.AddTraitToCharacterUseCase
	updateCharacterTraitUseCase  *characterapp.UpdateCharacterTraitUseCase
	removeTraitFromCharacterUseCase *characterapp.RemoveTraitFromCharacterUseCase
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
	addTraitToCharacterUseCase *characterapp.AddTraitToCharacterUseCase,
	updateCharacterTraitUseCase *characterapp.UpdateCharacterTraitUseCase,
	removeTraitFromCharacterUseCase *characterapp.RemoveTraitFromCharacterUseCase,
	logger logger.Logger,
) *CharacterHandler {
	return &CharacterHandler{
		createCharacterUseCase:      createCharacterUseCase,
		getCharacterUseCase:         getCharacterUseCase,
		listCharactersUseCase:       listCharactersUseCase,
		updateCharacterUseCase:      updateCharacterUseCase,
		deleteCharacterUseCase:      deleteCharacterUseCase,
		getCharacterTraitsUseCase:   getCharacterTraitsUseCase,
		addTraitToCharacterUseCase:  addTraitToCharacterUseCase,
		updateCharacterTraitUseCase: updateCharacterTraitUseCase,
		removeTraitFromCharacterUseCase: removeTraitFromCharacterUseCase,
		logger:                      logger,
	}
}

// CreateCharacter creates a new character
func (h *CharacterHandler) CreateCharacter(ctx context.Context, req *characterpb.CreateCharacterRequest) (*characterpb.CreateCharacterResponse, error) {
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
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character id: %v", err)
	}

	output, err := h.getCharacterUseCase.Execute(ctx, characterapp.GetCharacterInput{
		ID: id,
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
		WorldID: worldID,
		Limit:   limit,
		Offset:  offset,
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

	input := characterapp.UpdateCharacterInput{
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
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character id: %v", err)
	}

	err = h.deleteCharacterUseCase.Execute(ctx, characterapp.DeleteCharacterInput{
		ID: id,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.DeleteCharacterResponse{}, nil
}

// GetCharacterTraits retrieves all traits for a character
func (h *CharacterHandler) GetCharacterTraits(ctx context.Context, req *characterpb.GetCharacterTraitsRequest) (*characterpb.GetCharacterTraitsResponse, error) {
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	output, err := h.getCharacterTraitsUseCase.Execute(ctx, characterapp.GetCharacterTraitsInput{
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
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	traitID, err := uuid.Parse(req.TraitId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait_id: %v", err)
	}

	err = h.addTraitToCharacterUseCase.Execute(ctx, characterapp.AddTraitToCharacterInput{
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

	output, err := h.updateCharacterTraitUseCase.Execute(ctx, characterapp.UpdateCharacterTraitInput{
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
	characterID, err := uuid.Parse(req.CharacterId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid character_id: %v", err)
	}

	traitID, err := uuid.Parse(req.TraitId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid trait_id: %v", err)
	}

	err = h.removeTraitFromCharacterUseCase.Execute(ctx, characterapp.RemoveTraitFromCharacterInput{
		CharacterID: characterID,
		TraitID:     traitID,
	})
	if err != nil {
		return nil, err
	}

	return &characterpb.RemoveTraitFromCharacterResponse{}, nil
}

