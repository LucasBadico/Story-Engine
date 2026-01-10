package handlers

import (
	"net/http"

	"github.com/story-engine/main-service/internal/platform/relationmaps"
)

type RelationMapHandler struct{}

func NewRelationMapHandler() *RelationMapHandler {
	return &RelationMapHandler{}
}

// Types handles GET /api/v1/static/relations
func (h *RelationMapHandler) Types(w http.ResponseWriter, r *http.Request) {
	payload, err := relationmaps.TypesJSON()
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}

// Map handles GET /api/v1/static/relations/{entity_type}
func (h *RelationMapHandler) Map(w http.ResponseWriter, r *http.Request) {
	entityType := r.PathValue("entity_type")
	payload, err := relationmaps.MapJSON(entityType)
	if err != nil {
		WriteError(w, err, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(payload)
}
