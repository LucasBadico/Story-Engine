package mappers

import (
	"github.com/story-engine/main-service/internal/core/rpg"
	artifactrpgstatspb "github.com/story-engine/main-service/proto/artifact_rpg_stats"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ArtifactRPGStatsToProto converts an artifact RPG stats domain entity to a protobuf message
func ArtifactRPGStatsToProto(a *rpg.ArtifactRPGStats) *artifactrpgstatspb.ArtifactRPGStats {
	if a == nil {
		return nil
	}

	var eventID *string
	if a.EventID != nil {
		id := a.EventID.String()
		eventID = &id
	}

	stats := string(a.Stats)

	var reason *string
	if a.Reason != nil {
		r := *a.Reason
		reason = &r
	}

	var timeline *string
	if a.Timeline != nil {
		t := *a.Timeline
		timeline = &t
	}

	return &artifactrpgstatspb.ArtifactRPGStats{
		Id:         a.ID.String(),
		ArtifactId: a.ArtifactID.String(),
		EventId:    eventID,
		Stats:      stats,
		IsActive:   a.IsActive,
		Version:    int32(a.Version),
		Reason:     reason,
		Timeline:   timeline,
		CreatedAt:  timestamppb.New(a.CreatedAt),
	}
}

