//go:build integration

package testing

import (
	"net"
	"testing"

	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/logger"
	grpcserver "github.com/story-engine/main-service/internal/transport/grpc"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SetupTestServer creates a test gRPC server with provided handlers
// Returns a client connection and cleanup function
// Handlers must be provided to avoid import cycles
func SetupTestServer(t *testing.T, tenantHandler tenantpb.TenantServiceServer, storyHandler storypb.StoryServiceServer) (*grpc.ClientConn, func()) {
	// Create gRPC server
	cfg := config.Load()
	log := logger.New()
	grpcServer := grpcserver.NewServer(cfg, log)
	grpcServer.RegisterTenantService(tenantHandler)
	grpcServer.RegisterStoryService(storyHandler)

	// Start server on random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go func() {
		if err := grpcServer.GetServer().Serve(listener); err != nil {
			t.Logf("server error: %v", err)
		}
	}()

	// Create client connection
	conn, err := grpc.Dial(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}

	cleanup := func() {
		conn.Close()
		grpcServer.Stop()
	}

	return conn, cleanup
}

