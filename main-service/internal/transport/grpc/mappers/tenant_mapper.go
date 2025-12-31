package mappers

import (
	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/tenant"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TenantToProto converts a domain Tenant to a proto Tenant message
func TenantToProto(t *tenant.Tenant) *tenantpb.Tenant {
	return &tenantpb.Tenant{
		Id:        t.ID.String(),
		Name:      t.Name,
		Status:    string(t.Status),
		CreatedAt: timestamppb.New(t.CreatedAt),
		UpdatedAt: timestamppb.New(t.UpdatedAt),
	}
}

// TenantIDFromProto converts a proto UUID string to domain UUID
func TenantIDFromProto(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

