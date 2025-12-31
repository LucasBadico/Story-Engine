package mappers

import (
	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/story"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Note: This file depends on generated proto code.
// Run `make proto-gen` to generate proto files before building.

// StoryToProto converts a domain Story to a proto Story message
// This will work once proto files are generated
func StoryToProto(s *story.Story) interface{} {
	// TODO: Replace with actual proto type after generation
	// pb := &storypb.Story{
	// 	Id:              s.ID.String(),
	// 	TenantId:        s.TenantID.String(),
	// 	Title:           s.Title,
	// 	Status:          string(s.Status),
	// 	VersionNumber:   int32(s.VersionNumber),
	// 	RootStoryId:     s.RootStoryID.String(),
	// 	PreviousStoryId: getStringPtr(s.PreviousStoryID),
	// 	CreatedAt:       timestamppb.New(s.CreatedAt),
	// 	UpdatedAt:       timestamppb.New(s.UpdatedAt),
	// }
	// return pb
	
	// Placeholder implementation - will be updated after proto generation
	result := map[string]interface{}{
		"id":            s.ID.String(),
		"tenant_id":     s.TenantID.String(),
		"title":         s.Title,
		"status":        string(s.Status),
		"version_number": int32(s.VersionNumber),
		"root_story_id":  s.RootStoryID.String(),
		"created_at":     timestamppb.New(s.CreatedAt),
		"updated_at":     timestamppb.New(s.UpdatedAt),
	}
	if s.PreviousStoryID != nil {
		result["previous_story_id"] = s.PreviousStoryID.String()
	}
	return result
}

// ChapterToProto converts a domain Chapter to a proto Chapter message
func ChapterToProto(c *story.Chapter) interface{} {
	// Placeholder implementation
	return map[string]interface{}{
		"id":         c.ID.String(),
		"story_id":   c.StoryID.String(),
		"number":     int32(c.Number),
		"title":      c.Title,
		"status":     string(c.Status),
		"created_at": timestamppb.New(c.CreatedAt),
		"updated_at": timestamppb.New(c.UpdatedAt),
	}
}

// SceneToProto converts a domain Scene to a proto Scene message
func SceneToProto(s *story.Scene) interface{} {
	// Placeholder implementation
	result := map[string]interface{}{
		"id":         s.ID.String(),
		"story_id":   s.StoryID.String(),
		"chapter_id": s.ChapterID.String(),
		"order_num":  int32(s.OrderNum),
		"goal":       s.Goal,
		"time_ref":   s.TimeRef,
		"created_at": timestamppb.New(s.CreatedAt),
		"updated_at": timestamppb.New(s.UpdatedAt),
	}
	if s.POVCharacterID != nil {
		result["pov_character_id"] = s.POVCharacterID.String()
	}
	if s.LocationID != nil {
		result["location_id"] = s.LocationID.String()
	}
	return result
}

// StoryIDFromProto converts a proto UUID string to domain UUID
func StoryIDFromProto(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// Helper function to convert UUID pointer to string pointer
func getStringPtr(uuidPtr *uuid.UUID) string {
	if uuidPtr == nil {
		return ""
	}
	return uuidPtr.String()
}

