package mappers

import (
	"github.com/story-engine/main-service/internal/core/rpg"
	rpgclasspb "github.com/story-engine/main-service/proto/rpg_class"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// RPGClassToProto converts an RPG class domain entity to a protobuf message
func RPGClassToProto(c *rpg.RPGClass) *rpgclasspb.RPGClass {
	if c == nil {
		return nil
	}

	var parentClassID *string
	if c.ParentClassID != nil {
		id := c.ParentClassID.String()
		parentClassID = &id
	}

	var description *string
	if c.Description != nil {
		desc := *c.Description
		description = &desc
	}

	var requirements *string
	if c.Requirements != nil {
		req := string(*c.Requirements)
		requirements = &req
	}

	var statBonuses *string
	if c.StatBonuses != nil {
		bonuses := string(*c.StatBonuses)
		statBonuses = &bonuses
	}

	return &rpgclasspb.RPGClass{
		Id:            c.ID.String(),
		RpgSystemId:   c.RPGSystemID.String(),
		ParentClassId: parentClassID,
		Name:          c.Name,
		Tier:          int32(c.Tier),
		Description:   description,
		Requirements:  requirements,
		StatBonuses:   statBonuses,
		CreatedAt:     timestamppb.New(c.CreatedAt),
		UpdatedAt:     timestamppb.New(c.UpdatedAt),
	}
}

