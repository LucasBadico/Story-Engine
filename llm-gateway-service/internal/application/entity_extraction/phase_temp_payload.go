package entity_extraction

import "github.com/google/uuid"

type PhaseTempPayloadUseCase struct{}

func NewPhaseTempPayloadUseCase() *PhaseTempPayloadUseCase {
	return &PhaseTempPayloadUseCase{}
}

type PhaseTempPayload struct {
	Entities []PhaseTempEntity `json:"entities"`
}

type PhaseTempEntity struct {
	Type       string           `json:"type"`
	Name       string           `json:"name"`
	Summary    string           `json:"summary,omitempty"`
	Match      *PhaseTempMatch  `json:"match,omitempty"`
	Candidates []PhaseTempMatch `json:"candidates,omitempty"`
}

type PhaseTempMatch struct {
	SourceType string    `json:"source_type"`
	SourceID   uuid.UUID `json:"source_id"`
	EntityName string    `json:"entity_name,omitempty"`
	Similarity float64   `json:"similarity"`
	Reason     string    `json:"reason,omitempty"`
}

func (u *PhaseTempPayloadUseCase) Execute(input Phase3MatchOutput) PhaseTempPayload {
	entities := make([]PhaseTempEntity, 0, len(input.Results))
	for _, result := range input.Results {
		entity := PhaseTempEntity{
			Type:    result.EntityType,
			Name:    result.Name,
			Summary: result.Summary,
		}

		if result.Match != nil {
			entity.Match = &PhaseTempMatch{
				SourceType: string(result.Match.Candidate.SourceType),
				SourceID:   result.Match.Candidate.SourceID,
				EntityName: result.Match.Candidate.EntityName,
				Similarity: result.Match.Candidate.Similarity,
				Reason:     result.Match.Reason,
			}
		}

		if len(result.Candidates) > 0 {
			candidates := make([]PhaseTempMatch, 0, len(result.Candidates))
			for _, candidate := range result.Candidates {
				candidates = append(candidates, PhaseTempMatch{
					SourceType: string(candidate.SourceType),
					SourceID:   candidate.SourceID,
					EntityName: candidate.EntityName,
					Similarity: candidate.Similarity,
				})
			}
			entity.Candidates = candidates
		}

		entities = append(entities, entity)
	}

	return PhaseTempPayload{Entities: entities}
}
