//go:build integration

package testing

import (
	"net"
	"testing"

	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/logger"
	grpcserver "github.com/story-engine/main-service/internal/transport/grpc"
	beatpb "github.com/story-engine/main-service/proto/beat"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	prosepb "github.com/story-engine/main-service/proto/prose"
	scenepb "github.com/story-engine/main-service/proto/scene"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestHandlers contains all handler interfaces for testing
type TestHandlers struct {
	TenantHandler              tenantpb.TenantServiceServer
	StoryHandler               storypb.StoryServiceServer
	ChapterHandler             chapterpb.ChapterServiceServer
	SceneHandler               scenepb.SceneServiceServer
	BeatHandler                beatpb.BeatServiceServer
	ProseBlockHandler          prosepb.ProseBlockServiceServer
	ProseBlockReferenceHandler prosepb.ProseBlockReferenceServiceServer
}

// SetupTestServer creates a test gRPC server with provided handlers
// Returns a client connection and cleanup function
// Handlers must be provided to avoid import cycles
func SetupTestServer(t *testing.T, tenantHandler tenantpb.TenantServiceServer, storyHandler storypb.StoryServiceServer) (*grpc.ClientConn, func()) {
	return SetupTestServerWithHandlers(t, TestHandlers{
		TenantHandler: tenantHandler,
		StoryHandler:  storyHandler,
	})
}

// SetupTestServerWithHandlers creates a test gRPC server with all handlers
func SetupTestServerWithHandlers(t *testing.T, handlers TestHandlers) (*grpc.ClientConn, func()) {
	// Create gRPC server
	cfg := config.Load()
	log := logger.New()
	grpcServer := grpcserver.NewServer(cfg, log)

	// Register handlers if provided
	if handlers.TenantHandler != nil {
		grpcServer.RegisterTenantService(handlers.TenantHandler)
	}
	if handlers.StoryHandler != nil {
		grpcServer.RegisterStoryService(handlers.StoryHandler)
	}
	if handlers.ChapterHandler != nil {
		grpcServer.RegisterChapterService(handlers.ChapterHandler)
	}
	if handlers.SceneHandler != nil {
		grpcServer.RegisterSceneService(handlers.SceneHandler)
	}
	if handlers.BeatHandler != nil {
		grpcServer.RegisterBeatService(handlers.BeatHandler)
	}
	if handlers.ProseBlockHandler != nil {
		grpcServer.RegisterProseBlockService(handlers.ProseBlockHandler)
	}
	if handlers.ProseBlockReferenceHandler != nil {
		grpcServer.RegisterProseBlockReferenceService(handlers.ProseBlockReferenceHandler)
	}

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

