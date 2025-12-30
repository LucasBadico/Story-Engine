package mappers

import (
	"github.com/google/uuid"
	"github.com/story-engine/main-service/internal/core/tenant"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Note: This file depends on generated proto code.
// Run `make proto-gen` to generate proto files before building.

// TenantToProto converts a domain Tenant to a proto Tenant message
// This will work once proto files are generated
func TenantToProto(t *tenant.Tenant) interface{} {
	// TODO: Replace with actual proto type after generation
	// pb := &tenantpb.Tenant{
	// 	Id:        t.ID.String(),
	// 	Name:      t.Name,
	// 	Status:    string(t.Status),
	// 	CreatedAt: timestamppb.New(t.CreatedAt),
	// 	UpdatedAt: timestamppb.New(t.UpdatedAt),
	// }
	// return pb
	
	// Placeholder implementation - will be updated after proto generation
	return map[string]interface{}{
		"id":        t.ID.String(),
		"name":      t.Name,
		"status":    string(t.Status),
		"created_at": timestamppb.New(t.CreatedAt),
		"updated_at": timestamppb.New(t.UpdatedAt),
	}
}

// TenantIDFromProto converts a proto UUID string to domain UUID
func TenantIDFromProto(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

