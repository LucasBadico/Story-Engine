package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

type contextKey string

const TenantIDKey contextKey = "tenant_id"

// TenantMiddleware extracts and validates the X-Tenant-ID header
// and injects it into the request context
// Routes that don't need tenant isolation are skipped (e.g., /api/v1/tenants, /health)
// OPTIONS requests (CORS preflight) are also skipped
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip tenant validation for OPTIONS requests (CORS preflight)
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip tenant validation for routes that don't need it
		path := r.URL.Path
		if path == "/health" || strings.HasPrefix(path, "/api/v1/tenants") {
			next.ServeHTTP(w, r)
			return
		}

		tenantIDStr := r.Header.Get("X-Tenant-ID")
		if tenantIDStr == "" {
			writeError(w, &platformerrors.ValidationError{
				Field:   "X-Tenant-ID",
				Message: "header is required",
			}, http.StatusUnauthorized)
			return
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			writeError(w, &platformerrors.ValidationError{
				Field:   "X-Tenant-ID",
				Message: "invalid UUID format",
			}, http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantID extracts the tenant ID from the context
func GetTenantID(ctx context.Context) uuid.UUID {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return tenantID
}

// writeError writes a standardized error response
func writeError(w http.ResponseWriter, err error, defaultStatus int) {
	var status int
	var response struct {
		Error   string            `json:"error"`
		Message string            `json:"message"`
		Code    string            `json:"code"`
		Details map[string]string `json:"details,omitempty"`
	}

	// Map domain errors to HTTP status codes
	switch e := err.(type) {
	case *platformerrors.ValidationError:
		status = http.StatusBadRequest
		details := make(map[string]string)
		if e.Field != "" {
			details["field"] = e.Field
		}
		response = struct {
			Error   string            `json:"error"`
			Message string            `json:"message"`
			Code    string            `json:"code"`
			Details map[string]string `json:"details,omitempty"`
		}{
			Error:   "validation_error",
			Message: e.Error(),
			Code:    "VALIDATION_ERROR",
			Details: details,
		}
	default:
		status = defaultStatus
		response = struct {
			Error   string            `json:"error"`
			Message string            `json:"message"`
			Code    string            `json:"code"`
			Details map[string]string `json:"details,omitempty"`
		}{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    "INTERNAL_ERROR",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

