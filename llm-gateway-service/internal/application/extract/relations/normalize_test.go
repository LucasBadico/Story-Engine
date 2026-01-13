package relations

import (
	"strings"
	"testing"
)

func TestPhase6RelationNormalizeUseCase(t *testing.T) {
	useCase := NewPhase6RelationNormalizeUseCase(nil)

	input := Phase6RelationNormalizeInput{
		RequestID: "req-6",
		Context: Phase5Context{
			Type: "scene",
			ID:   "scene-1",
		},
		Relations: []Phase5RelationCandidate{
			{
				Source:       Phase5RelationNode{Ref: "finding:character:0", Type: "character"},
				Target:       Phase5RelationNode{Ref: "match:faction:1", Type: "faction"},
				RelationType: "member_of",
				Polarity:     "asserted",
				Implicit:     false,
				Confidence:   0.8,
				Evidence:     Phase5Evidence{SpanID: "span:1", Quote: "Ari swore loyalty."},
			},
		},
		RefMap: map[string]Phase6ResolvedRef{
			"finding:character:0": {ID: "uuid-char", Type: "character", Name: "Ari"},
			"match:faction:1":     {ID: "uuid-faction", Type: "faction", Name: "Order of the Sun"},
		},
		SuggestedRelationsBySourceType: map[string]Phase5PerEntityRelationMap{
			"character": {
				EntityType: "character",
				Version:    1,
				Relations: map[string]Phase5RelationConstraintSpec{
					"member_of": {
						PairCandidates: []string{"faction"},
					},
				},
			},
		},
		RelationTypes: map[string]Phase6RelationTypeDefinition{
			"member_of": {Mirror: "has_member", PreferredDirection: "source_to_target"},
		},
	}

	output, err := useCase.Execute(t.Context(), input)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if len(output.Relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(output.Relations))
	}
	rel := output.Relations[0]
	if !rel.CreateMirror {
		t.Fatalf("expected create_mirror true")
	}
	if rel.Status != "ready" {
		t.Fatalf("expected status ready, got %s", rel.Status)
	}
	if rel.Source.ID == "" || rel.Target.ID == "" {
		t.Fatalf("expected resolved ids")
	}
	if rel.Summary == "" {
		t.Fatalf("expected summary")
	}
}

func TestPhase6RelationNormalizeUseCase_CustomMirrorSuggestion(t *testing.T) {
	useCase := NewPhase6RelationNormalizeUseCase(nil)

	input := Phase6RelationNormalizeInput{
		RequestID: "req-7",
		Context: Phase5Context{
			Type: "scene",
			ID:   "scene-1",
		},
		Relations: []Phase5RelationCandidate{
			{
				Source:       Phase5RelationNode{Ref: "finding:character:0", Type: "character"},
				Target:       Phase5RelationNode{Ref: "finding:character:1", Type: "character"},
				RelationType: "custom:bonded_to",
				Polarity:     "asserted",
				Implicit:     false,
				Confidence:   0.6,
				Evidence:     Phase5Evidence{SpanID: "span:1", Quote: "They were bound."},
			},
		},
		RefMap: map[string]Phase6ResolvedRef{
			"finding:character:0": {ID: "uuid-a", Type: "character"},
			"finding:character:1": {ID: "uuid-b", Type: "character"},
		},
		SuggestedRelationsBySourceType: map[string]Phase5PerEntityRelationMap{
			"character": {
				EntityType: "character",
				Version:    1,
				Relations:  map[string]Phase5RelationConstraintSpec{},
			},
		},
		RelationTypes: map[string]Phase6RelationTypeDefinition{},
	}

	output, err := useCase.Execute(t.Context(), input)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if len(output.Relations) != 2 {
		t.Fatalf("expected 2 relations for custom mirror suggestion, got %d", len(output.Relations))
	}
	if output.Relations[0].CreateMirror {
		t.Fatalf("expected create_mirror false for custom")
	}
	if output.Relations[1].MirrorOf == nil {
		t.Fatalf("expected mirror_of on mirrored custom relation")
	}
	if output.Relations[0].Summary == "" {
		t.Fatalf("expected summary fallback for custom")
	}
}

func TestPhase6RelationNormalizeUseCase_DowngradeCustomRelationType(t *testing.T) {
	useCase := NewPhase6RelationNormalizeUseCase(nil)

	input := Phase6RelationNormalizeInput{
		RequestID: "req-8",
		Context: Phase5Context{
			Type: "scene",
			ID:   "scene-1",
		},
		Relations: []Phase5RelationCandidate{
			{
				Source:       Phase5RelationNode{Ref: "finding:character:0", Type: "character"},
				Target:       Phase5RelationNode{Ref: "finding:location:0", Type: "location"},
				RelationType: "custom:contains",
				Polarity:     "asserted",
				Implicit:     false,
				Confidence:   0.6,
				Evidence:     Phase5Evidence{SpanID: "span:1", Quote: "The camp contains the shrine."},
			},
		},
		RefMap: map[string]Phase6ResolvedRef{
			"finding:character:0": {ID: "uuid-a", Type: "character", Name: "Ari"},
			"finding:location:0":  {ID: "uuid-b", Type: "location", Name: "Camp"},
		},
		SuggestedRelationsBySourceType: map[string]Phase5PerEntityRelationMap{
			"character": {
				EntityType: "character",
				Version:    1,
				Relations: map[string]Phase5RelationConstraintSpec{
					"contains": {
						PairCandidates: []string{"location"},
					},
				},
			},
		},
		RelationTypes: map[string]Phase6RelationTypeDefinition{
			"contains": {Mirror: "contained_by", PreferredDirection: "source_to_target"},
		},
	}

	output, err := useCase.Execute(t.Context(), input)
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if len(output.Relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(output.Relations))
	}
	rel := output.Relations[0]
	if rel.RelationType != "contains" {
		t.Fatalf("expected relation_type contains, got %s", rel.RelationType)
	}
	if strings.HasPrefix(rel.RelationType, "custom:") {
		t.Fatalf("expected custom prefix removed")
	}
}
