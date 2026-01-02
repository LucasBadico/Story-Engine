package mappers

import (
	"github.com/story-engine/main-service/internal/core/rpg"
	characterrpgstatspb "github.com/story-engine/main-service/proto/character_rpg_stats"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CharacterRPGStatsToProto converts a character RPG stats domain entity to a protobuf message
func CharacterRPGStatsToProto(c *rpg.CharacterRPGStats) *characterrpgstatspb.CharacterRPGStats {
	if c == nil {
		return nil
	}

	var eventID *string
	if c.EventID != nil {
		id := c.EventID.String()
		eventID = &id
	}

	baseStats := string(c.BaseStats)

	var derivedStats *string
	if c.DerivedStats != nil {
		stats := string(*c.DerivedStats)
		derivedStats = &stats
	}

	var progression *string
	if c.Progression != nil {
		prog := string(*c.Progression)
		progression = &prog
	}

	var reason *string
	if c.Reason != nil {
		r := *c.Reason
		reason = &r
	}

	var timeline *string
	if c.Timeline != nil {
		t := *c.Timeline
		timeline = &t
	}

	return &characterrpgstatspb.CharacterRPGStats{
		Id:            c.ID.String(),
		CharacterId:   c.CharacterID.String(),
		EventId:       eventID,
		BaseStats:     baseStats,
		DerivedStats:  derivedStats,
		Progression:   progression,
		IsActive:      c.IsActive,
		Version:       int32(c.Version),
		Reason:        reason,
		Timeline:      timeline,
		CreatedAt:     timestamppb.New(c.CreatedAt),
	}
}

