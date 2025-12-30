package handlers

import (
	"context"

	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/application/tenant"
	grpcctx "github.com/story-engine/main-service/internal/transport/grpc"
	"github.com/story-engine/main-service/internal/transport/grpc/mappers"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/ports/repositories"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TenantHandler implements the TenantService gRPC service
// Note: This handler depends on generated proto code.
// After running `make proto-gen`, update the method signatures to use actual proto types.
type TenantHandler struct {
	// pb.UnimplementedTenantServiceServer // Uncomment after proto generation
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
// TODO: Update signature after proto generation:
// func (h *TenantHandler) CreateTenant(ctx context.Context, req *tenantpb.CreateTenantRequest) (*tenantpb.CreateTenantResponse, error)
func (h *TenantHandler) CreateTenant(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: After proto generation, replace with:
	// if req.Name == "" {
	//     return nil, status.Errorf(codes.InvalidArgument, "name is required")
	// }
	
	// Temporary implementation - extract from map or use reflection
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}
	
	name, _ := reqMap["name"].(string)
	if name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "name is required")
	}
	
	var createdBy *uuid.UUID
	if createdByStr, ok := reqMap["created_by_user_id"].(string); ok && createdByStr != "" {
		if id, err := uuid.Parse(createdByStr); err == nil {
			createdBy = &id
		}
	}

	input := tenant.CreateTenantInput{
		Name:      name,
		CreatedBy: createdBy,
	}

	output, err := h.createTenantUseCase.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	// TODO: After proto generation, replace with:
	// return &tenantpb.CreateTenantResponse{
	//     Tenant: mappers.TenantToProto(output.Tenant).(*tenantpb.Tenant),
	// }, nil
	
	return map[string]interface{}{
		"tenant": mappers.TenantToProto(output.Tenant),
	}, nil
}

// GetTenant retrieves a tenant by ID
// TODO: Update signature after proto generation:
// func (h *TenantHandler) GetTenant(ctx context.Context, req *tenantpb.GetTenantRequest) (*tenantpb.GetTenantResponse, error)
func (h *TenantHandler) GetTenant(ctx context.Context, req interface{}) (interface{}, error) {
	// TODO: After proto generation, replace with:
	// id, err := uuid.Parse(req.Id)
	// if err != nil {
	//     return nil, status.Errorf(codes.InvalidArgument, "invalid tenant id: %v", err)
	// }
	
	reqMap, ok := req.(map[string]interface{})
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "invalid request")
	}
	
	idStr, _ := reqMap["id"].(string)
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant id: %v", err)
	}

	tenant, err := h.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// TODO: After proto generation, replace with:
	// return &tenantpb.GetTenantResponse{
	//     Tenant: mappers.TenantToProto(tenant).(*tenantpb.Tenant),
	// }, nil
	
	return map[string]interface{}{
		"tenant": mappers.TenantToProto(tenant),
	}, nil
}

