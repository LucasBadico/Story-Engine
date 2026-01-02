package mappers

import (
	"github.com/story-engine/main-service/internal/core/rpg"
	skillpb "github.com/story-engine/main-service/proto/skill"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SkillToProto converts a skill domain entity to a protobuf message
func SkillToProto(s *rpg.Skill) *skillpb.Skill {
	if s == nil {
		return nil
	}

	var category *string
	if s.Category != nil {
		cat := string(*s.Category)
		category = &cat
	}

	var skillType *string
	if s.Type != nil {
		t := string(*s.Type)
		skillType = &t
	}

	var description *string
	if s.Description != nil {
		desc := *s.Description
		description = &desc
	}

	var prerequisites *string
	if s.Prerequisites != nil {
		prereq := string(*s.Prerequisites)
		prerequisites = &prereq
	}

	var effectsSchema *string
	if s.EffectsSchema != nil {
		effects := string(*s.EffectsSchema)
		effectsSchema = &effects
	}

	return &skillpb.Skill{
		Id:            s.ID.String(),
		RpgSystemId:   s.RPGSystemID.String(),
		Name:          s.Name,
		Category:      category,
		Type:          skillType,
		Description:   description,
		Prerequisites: prerequisites,
		MaxRank:       int32(s.MaxRank),
		EffectsSchema: effectsSchema,
		CreatedAt:     timestamppb.New(s.CreatedAt),
		UpdatedAt:     timestamppb.New(s.UpdatedAt),
	}
}

