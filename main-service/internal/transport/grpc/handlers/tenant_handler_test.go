//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	grpctesting "github.com/story-engine/main-service/internal/transport/grpc/testing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Note: These tests depend on generated proto code.
// Run `make proto-gen` before running tests.

func TestTenantHandler_CreateTenant(t *testing.T) {
	conn, cleanup := grpctesting.SetupTestServer(t)
	defer cleanup()

	// TODO: After proto generation, create actual client:
	// client := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		// TODO: After proto generation:
		// req := &tenantpb.CreateTenantRequest{
		//     Name: "Test Tenant",
		// }
		// resp, err := client.CreateTenant(context.Background(), req)
		// if err != nil {
		//     t.Fatalf("unexpected error: %v", err)
		// }
		// if resp.Tenant.Name != "Test Tenant" {
		//     t.Errorf("expected name to be 'Test Tenant', got '%s'", resp.Tenant.Name)
		// }
		
		// Placeholder test
		_ = conn
		t.Skip("Skipping until proto files are generated")
	})

	t.Run("duplicate name", func(t *testing.T) {
		// TODO: After proto generation:
		// req := &tenantpb.CreateTenantRequest{
		//     Name: "Duplicate Tenant",
		// }
		// _, err := client.CreateTenant(context.Background(), req)
		// if err != nil {
		//     t.Fatalf("unexpected error on first creation: %v", err)
		// }
		// 
		// // Try to create duplicate
		// _, err = client.CreateTenant(context.Background(), req)
		// if err == nil {
		//     t.Fatal("expected error for duplicate name")
		// }
		// 
		// s, _ := status.FromError(err)
		// if s.Code() != codes.AlreadyExists {
		//     t.Errorf("expected AlreadyExists, got %v", s.Code())
		// }
		
		t.Skip("Skipping until proto files are generated")
	})
}

func TestTenantHandler_GetTenant(t *testing.T) {
	conn, cleanup := grpctesting.SetupTestServer(t)
	defer cleanup()

	// TODO: After proto generation:
	// client := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing tenant", func(t *testing.T) {
		// TODO: Create tenant first, then get it
		t.Skip("Skipping until proto files are generated")
	})

	t.Run("non-existing tenant", func(t *testing.T) {
		// TODO: After proto generation:
		// req := &tenantpb.GetTenantRequest{
		//     Id: uuid.New().String(),
		// }
		// _, err := client.GetTenant(context.Background(), req)
		// if err == nil {
		//     t.Fatal("expected error for non-existing tenant")
		// }
		// 
		// s, _ := status.FromError(err)
		// if s.Code() != codes.NotFound {
		//     t.Errorf("expected NotFound, got %v", s.Code())
		// }
		
		t.Skip("Skipping until proto files are generated")
	})
}

