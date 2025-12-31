package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/tenant"
	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
)

// TenantHandler handles HTTP requests for tenants
type TenantHandler struct {
	createTenantUseCase *tenant.CreateTenantUseCase
	tenantRepo          repositories.TenantRepository
	logger              logger.Logger
}

// NewTenantHandler creates a new TenantHandler
func NewTenantHandler(
	createTenantUseCase *tenant.CreateTenantUseCase,
	tenantRepo repositories.TenantRepository,
	logger logger.Logger,
) *TenantHandler {
	return &TenantHandler{
		createTenantUseCase: createTenantUseCase,
		tenantRepo:          tenantRepo,
		logger:              logger,
	}
}

// Create handles POST /api/v1/tenants
func (h *TenantHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "body",
			Message: "invalid JSON",
		}, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "name",
			Message: "name is required",
		}, http.StatusBadRequest)
		return
	}

	output, err := h.createTenantUseCase.Execute(r.Context(), tenant.CreateTenantInput{
		Name: req.Name,
	})
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant": output.Tenant,
	})
}

// Get handles GET /api/v1/tenants/{id}
func (h *TenantHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	tenantID, err := uuid.Parse(id)
	if err != nil {
		WriteError(w, &platformerrors.ValidationError{
			Field:   "id",
			Message: "invalid UUID format",
		}, http.StatusBadRequest)
		return
	}

	tenant, err := h.tenantRepo.GetByID(r.Context(), tenantID)
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tenant": tenant,
	})
}

