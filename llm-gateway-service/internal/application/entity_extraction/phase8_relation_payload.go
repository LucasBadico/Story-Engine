package entity_extraction

// Phase8RelationResult represents the final relation payload with matches.
type Phase8RelationResult struct {
	Source       Phase6NormalizedNode           `json:"source"`
	Target       Phase6NormalizedNode           `json:"target"`
	RelationType string                         `json:"relation_type"`
	Direction    string                         `json:"direction"`
	Summary      string                         `json:"summary,omitempty"`
	Confidence   float64                        `json:"confidence,omitempty"`
	Polarity     string                         `json:"polarity,omitempty"`
	Implicit     bool                           `json:"implicit,omitempty"`
	Evidence     Phase5Evidence                 `json:"evidence,omitempty"`
	Status       string                         `json:"status"`
	Dedup        Phase6Dedup                    `json:"dedup,omitempty"`
	CreateMirror bool                           `json:"create_mirror,omitempty"`
	MirrorOf     *string                        `json:"mirror_of,omitempty"`
	Matches      []Phase7RelationMatchCandidate `json:"matches,omitempty"`
}
