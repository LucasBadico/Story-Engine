package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/tenant"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TenantHandler implements the TenantService gRPC service
type TenantHandler struct {
	tenantpb.UnimplementedTenantServiceServer
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

// CreateTenant creates a new tenant
func (h *TenantHandler) CreateTenant(ctx context.Context, req *tenantpb.CreateTenantRequest) (*tenantpb.CreateTenantResponse, error) {
	if req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	
	var createdBy *uuid.UUID
	if req.CreatedByUserId != "" {
		if id, err := uuid.Parse(req.CreatedByUserId); err == nil {
			createdBy = &id
		} else {
			return nil, status.Errorf(codes.InvalidArgument, "invalid created_by_user_id: %v", err)
		}
	}

	input := tenant.CreateTenantInput{
		Name:      req.Name,
		CreatedBy: createdBy,
	}

	output, err := h.createTenantUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	return &tenantpb.CreateTenantResponse{
		Tenant: mappers.TenantToProto(output.Tenant),
	}, nil
}

// GetTenant retrieves a tenant by ID
func (h *TenantHandler) GetTenant(ctx context.Context, req *tenantpb.GetTenantRequest) (*tenantpb.GetTenantResponse, error) {
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant id: %v", err)
	}

	tenant, err := h.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &tenantpb.GetTenantResponse{
		Tenant: mappers.TenantToProto(tenant),
	}, nil
}

