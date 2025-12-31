//go:build integration

package handlers

import (
	"context"
	"testing"

	"github.com/google/uuid"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTenantHandler_CreateTenant(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := tenantpb.NewTenantServiceClient(conn)

	t.Run("successful creation", func(t *testing.T) {
		req := &tenantpb.CreateTenantRequest{
			Name: "Test Tenant",
		}
		resp, err := client.CreateTenant(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.Tenant.Name != "Test Tenant" {
			t.Errorf("expected name to be 'Test Tenant', got '%s'", resp.Tenant.Name)
		}
		if resp.Tenant.Status != "active" {
			t.Errorf("expected status to be 'active', got '%s'", resp.Tenant.Status)
		}
		if resp.Tenant.Id == "" {
			t.Error("expected tenant ID to be set")
		}
	})

	t.Run("duplicate name", func(t *testing.T) {
		req := &tenantpb.CreateTenantRequest{
			Name: "Duplicate Tenant",
		}
		_, err := client.CreateTenant(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error on first creation: %v", err)
		}

		// Try to create duplicate
		_, err = client.CreateTenant(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for duplicate name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.AlreadyExists {
			t.Errorf("expected AlreadyExists, got %v", s.Code())
		}
	})

	t.Run("empty name", func(t *testing.T) {
		req := &tenantpb.CreateTenantRequest{
			Name: "",
		}
		_, err := client.CreateTenant(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for empty name")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}

func TestTenantHandler_GetTenant(t *testing.T) {
	conn, cleanup := setupTestServer(t)
	defer cleanup()

	client := tenantpb.NewTenantServiceClient(conn)

	t.Run("existing tenant", func(t *testing.T) {
		// First create a tenant
		createReq := &tenantpb.CreateTenantRequest{
			Name: "Get Test Tenant",
		}
		createResp, err := client.CreateTenant(context.Background(), createReq)
		if err != nil {
			t.Fatalf("failed to create tenant: %v", err)
		}

		// Now get it
		getReq := &tenantpb.GetTenantRequest{
			Id: createResp.Tenant.Id,
		}
		getResp, err := client.GetTenant(context.Background(), getReq)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if getResp.Tenant.Id != createResp.Tenant.Id {
			t.Errorf("expected ID %s, got %s", createResp.Tenant.Id, getResp.Tenant.Id)
		}
		if getResp.Tenant.Name != "Get Test Tenant" {
			t.Errorf("expected name 'Get Test Tenant', got '%s'", getResp.Tenant.Name)
		}
	})

	t.Run("non-existing tenant", func(t *testing.T) {
		req := &tenantpb.GetTenantRequest{
			Id: uuid.New().String(),
		}
		_, err := client.GetTenant(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for non-existing tenant")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.NotFound {
			t.Errorf("expected NotFound, got %v", s.Code())
		}
	})

	t.Run("invalid tenant ID", func(t *testing.T) {
		req := &tenantpb.GetTenantRequest{
			Id: "not-a-uuid",
		}
		_, err := client.GetTenant(context.Background(), req)
		if err == nil {
			t.Fatal("expected error for invalid ID")
		}

		s, _ := status.FromError(err)
		if s.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument, got %v", s.Code())
		}
	})
}
