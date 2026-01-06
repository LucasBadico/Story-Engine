package grpc

import (
	"testing"
	"time"

	"github.com/google/uuid"
	eventpb "github.com/story-engine/main-service/proto/event"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestMapEventCharacters(t *testing.T) {
	eventID := uuid.New()
	refID := uuid.New()
	charID := uuid.New()
	now := time.Now()
	role := "witness"

	refs := []*eventpb.EventReference{
		{
			Id:               refID.String(),
			EventId:          eventID.String(),
			EntityType:       "character",
			EntityId:         charID.String(),
			RelationshipType: &role,
			CreatedAt:        timestamppb.New(now),
		},
		{
			Id:         uuid.New().String(),
			EventId:    eventID.String(),
			EntityType: "location",
			EntityId:   uuid.New().String(),
		},
	}

	result := mapEventCharacters(eventID, refs)
	if len(result) != 1 {
		t.Fatalf("expected 1 character, got %d", len(result))
	}
	if result[0].ID != refID {
		t.Errorf("expected ref ID %s, got %s", refID, result[0].ID)
	}
	if result[0].CharacterID != charID {
		t.Errorf("expected character ID %s, got %s", charID, result[0].CharacterID)
	}
	if result[0].Role == nil || *result[0].Role != role {
		t.Errorf("expected role %q, got %v", role, result[0].Role)
	}
	if result[0].CreatedAt != now.Unix() {
		t.Errorf("expected created_at %d, got %d", now.Unix(), result[0].CreatedAt)
	}
}

func TestMapEventLocations(t *testing.T) {
	eventID := uuid.New()
	refID := uuid.New()
	locationID := uuid.New()
	significance := "birthplace"

	refs := []*eventpb.EventReference{
		{
			Id:               refID.String(),
			EventId:          eventID.String(),
			EntityType:       "location",
			EntityId:         locationID.String(),
			RelationshipType: &significance,
		},
	}

	result := mapEventLocations(eventID, refs)
	if len(result) != 1 {
		t.Fatalf("expected 1 location, got %d", len(result))
	}
	if result[0].LocationID != locationID {
		t.Errorf("expected location ID %s, got %s", locationID, result[0].LocationID)
	}
	if result[0].Significance == nil || *result[0].Significance != significance {
		t.Errorf("expected significance %q, got %v", significance, result[0].Significance)
	}
}

func TestMapEventArtifacts_InvalidReference(t *testing.T) {
	eventID := uuid.New()

	refs := []*eventpb.EventReference{
		{
			Id:         "not-a-uuid",
			EventId:    eventID.String(),
			EntityType: "artifact",
			EntityId:   "bad-id",
		},
	}

	result := mapEventArtifacts(eventID, refs)
	if len(result) != 0 {
		t.Fatalf("expected 0 artifacts, got %d", len(result))
	}
}
