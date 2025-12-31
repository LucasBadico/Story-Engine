package mappers

import (
	"github.com/story-engine/main-service/internal/core/story"
	storypb "github.com/story-engine/main-service/proto/story"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// StoryToProto converts a domain Story to a proto Story message
func StoryToProto(s *story.Story) *storypb.Story {
	pb := &storypb.Story{
		Id:            s.ID.String(),
		TenantId:      s.TenantID.String(),
		Title:         s.Title,
		Status:        string(s.Status),
		VersionNumber: int32(s.VersionNumber),
		RootStoryId:   s.RootStoryID.String(),
		CreatedAt:     timestamppb.New(s.CreatedAt),
		UpdatedAt:     timestamppb.New(s.UpdatedAt),
	}
	
	if s.PreviousStoryID != nil {
		pb.PreviousStoryId = s.PreviousStoryID.String()
	}
	
	return pb
}

// ChapterToProto converts a domain Chapter to a proto Chapter message
func ChapterToProto(c *story.Chapter) *storypb.Chapter {
	return &storypb.Chapter{
		Id:        c.ID.String(),
		StoryId:   c.StoryID.String(),
		Number:    int32(c.Number),
		Title:     c.Title,
		Status:    string(c.Status),
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

// SceneToProto converts a domain Scene to a proto Scene message
func SceneToProto(s *story.Scene) *storypb.Scene {
	pb := &storypb.Scene{
		Id:        s.ID.String(),
		StoryId:   s.StoryID.String(),
		ChapterId: s.ChapterID.String(),
		OrderNum:  int32(s.OrderNum),
		Goal:      s.Goal,
		TimeRef:   s.TimeRef,
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
	}
	
	if s.POVCharacterID != nil {
		pb.PovCharacterId = s.POVCharacterID.String()
	}
	if s.LocationID != nil {
		pb.LocationId = s.LocationID.String()
	}
	
	return pb
}
