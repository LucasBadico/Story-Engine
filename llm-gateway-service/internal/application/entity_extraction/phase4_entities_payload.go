package entity_extraction

import "github.com/google/uuid"

type Phase4EntitiesPayloadUseCase struct{}

func NewPhase4EntitiesPayloadUseCase() *Phase4EntitiesPayloadUseCase {
	return &Phase4EntitiesPayloadUseCase{}
}

type Phase4EntitiesPayload struct {
	Entities  []Phase4Entity   `json:"entities"`
	Relations []Phase4Relation `json:"relations,omitempty"`
}

type Phase4Entity struct {
	Type       string        `json:"type"`
	Name       string        `json:"name"`
	Summary    string        `json:"summary,omitempty"`
	Found      bool          `json:"found"`
	Match      *Phase4Match  `json:"match,omitempty"`
	Candidates []Phase4Match `json:"candidates,omitempty"`
}

type Phase4Match struct {
	SourceType string            `json:"source_type"`
	SourceID   uuid.UUID         `json:"source_id"`
	EntityName string            `json:"entity_name,omitempty"`
	Similarity float64           `json:"similarity"`
	Reason     string            `json:"reason,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

type Phase4Relation struct {
	// Placeholder for Phase5+ output.
}

func (u *Phase4EntitiesPayloadUseCase) Execute(input Phase3MatchOutput) Phase4EntitiesPayload {
	entities := make([]Phase4Entity, 0, len(input.Results))
	for _, result := range input.Results {
		entity := Phase4Entity{
			Type:    result.EntityType,
			Name:    result.Name,
			Summary: result.Summary,
			Found:   result.Match != nil,
		}

		if result.Match != nil {
			entity.Match = &Phase4Match{
				SourceType: string(result.Match.Candidate.SourceType),
				SourceID:   result.Match.Candidate.SourceID,
				EntityName: result.Match.Candidate.EntityName,
				Similarity: result.Match.Candidate.Similarity,
				Reason:     result.Match.Reason,
				Metadata:   result.Match.Candidate.Metadata,
			}
		}

		if len(result.Candidates) > 0 {
			candidates := make([]Phase4Match, 0, len(result.Candidates))
			for _, candidate := range result.Candidates {
				candidates = append(candidates, Phase4Match{
					SourceType: string(candidate.SourceType),
					SourceID:   candidate.SourceID,
					EntityName: candidate.EntityName,
					Similarity: candidate.Similarity,
					Metadata:   candidate.Metadata,
				})
			}
			entity.Candidates = candidates
		}

		entities = append(entities, entity)
	}

	return Phase4EntitiesPayload{Entities: entities, Relations: []Phase4Relation{}}
}
