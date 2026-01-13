package extract

import (
	"github.com/story-engine/llm-gateway-service/internal/application/extract/entities"
	"github.com/story-engine/llm-gateway-service/internal/application/extract/relations"
)

type ExtractPayload struct {
	Entities  []entities.Phase4Entity          `json:"entities"`
	Relations []relations.Phase8RelationResult `json:"relations,omitempty"`
}
