package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Code    string            `json:"code"`
	Details map[string]string `json:"details,omitempty"`
}

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, err error, defaultStatus int) {
	var status int
	var response ErrorResponse

	// Map domain errors to HTTP status codes
	switch e := err.(type) {
	case *platformerrors.NotFoundError:
		status = http.StatusNotFound
		response = ErrorResponse{
			Error:   "not_found",
			Message: e.Error(),
			Code:    "NOT_FOUND",
		}
	case *platformerrors.AlreadyExistsError:
		status = http.StatusConflict
		response = ErrorResponse{
			Error:   "already_exists",
			Message: e.Error(),
			Code:    "ALREADY_EXISTS",
		}
	case *platformerrors.ValidationError:
		status = http.StatusBadRequest
		details := make(map[string]string)
		if e.Field != "" {
			details["field"] = e.Field
		}
		response = ErrorResponse{
			Error:   "validation_error",
			Message: e.Error(),
			Code:    "VALIDATION_ERROR",
			Details: details,
		}
	default:
		status = defaultStatus
		response = ErrorResponse{
			Error:   "internal_error",
			Message: err.Error(),
			Code:    "INTERNAL_ERROR",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// extractTenantID extracts tenant_id from the X-Tenant-ID header
func extractTenantID(r *http.Request) (uuid.UUID, error) {
	tenantIDStr := r.Header.Get("X-Tenant-ID")
	if tenantIDStr == "" {
		return uuid.Nil, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "header is required",
		}
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return uuid.Nil, &platformerrors.ValidationError{
			Field:   "X-Tenant-ID",
			Message: "invalid UUID format",
		}
	}

	return tenantID, nil
}
