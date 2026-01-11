package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/story-engine/llm-gateway-service/internal/application/entity_extraction"
	"github.com/story-engine/llm-gateway-service/internal/platform/logger"
)

func TestLoadRelationMaps(t *testing.T) {
	t.Helper()

	types := map[string]entity_extraction.Phase6RelationTypeDefinition{
		"member_of": {
			Mirror:             "has_member",
			Symmetric:          false,
			PreferredDirection: "source_to_target",
			Semantics:          "Source is a member of target.",
		},
	}
	relMap := entity_extraction.Phase5PerEntityRelationMap{
		EntityType: "character",
		Version:    1,
		Relations: map[string]entity_extraction.Phase5RelationConstraintSpec{
			"member_of": {
				PairCandidates: []string{"faction"},
				Description:    "Character belongs to a group.",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/static/relations":
			_ = json.NewEncoder(w).Encode(types)
			return
		case "/api/v1/static/relations/character",
			"/api/v1/static/relations/faction",
			"/api/v1/static/relations/location",
			"/api/v1/static/relations/event":
			_ = json.NewEncoder(w).Encode(relMap)
			return
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	loadedTypes, suggested := loadRelationMaps(logger.New(), server.URL)
	if len(loadedTypes) != 1 {
		t.Fatalf("expected 1 relation type, got %d", len(loadedTypes))
	}
	if len(suggested) != 4 {
		t.Fatalf("expected 4 suggested maps, got %d", len(suggested))
	}
	if _, ok := suggested["character"]; !ok {
		t.Fatalf("expected character map in suggested relations")
	}
}
