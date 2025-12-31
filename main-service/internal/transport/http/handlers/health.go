package handlers

import (
	"encoding/json"
	"net/http"
)

// HealthCheck handles health check requests
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

