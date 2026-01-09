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
	contentblockpb "github.com/story-engine/main-service/proto/content_block"
	scenepb "github.com/story-engine/main-service/proto/scene"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	traitpb "github.com/story-engine/main-service/proto/trait"
	worldpb "github.com/story-engine/main-service/proto/world"
	archetypepb "github.com/story-engine/main-service/proto/archetype"
	locationpb "github.com/story-engine/main-service/proto/location"
	characterpb "github.com/story-engine/main-service/proto/character"
	artifactpb "github.com/story-engine/main-service/proto/artifact"
	eventpb "github.com/story-engine/main-service/proto/event"
	rpgsystempb "github.com/story-engine/main-service/proto/rpg_system"
	skillpb "github.com/story-engine/main-service/proto/skill"
	rpgclasspb "github.com/story-engine/main-service/proto/rpg_class"
	characterskillpb "github.com/story-engine/main-service/proto/character_skill"
	characterrpgstatspb "github.com/story-engine/main-service/proto/character_rpg_stats"
	artifactrpgstatspb "github.com/story-engine/main-service/proto/artifact_rpg_stats"
	inventorypb "github.com/story-engine/main-service/proto/inventory"
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
	ContentBlockHandler          contentblockpb.ContentBlockServiceServer
	ContentAnchorHandler contentblockpb.ContentAnchorServiceServer
	WorldHandler               worldpb.WorldServiceServer
	TraitHandler               traitpb.TraitServiceServer
	ArchetypeHandler           archetypepb.ArchetypeServiceServer
	LocationHandler            locationpb.LocationServiceServer
	CharacterHandler           characterpb.CharacterServiceServer
	ArtifactHandler            artifactpb.ArtifactServiceServer
	EventHandler               eventpb.EventServiceServer
	RPGSystemHandler           rpgsystempb.RPGSystemServiceServer
	SkillHandler               skillpb.SkillServiceServer
	RPGClassHandler            rpgclasspb.RPGClassServiceServer
	CharacterSkillHandler      characterskillpb.CharacterSkillServiceServer
	CharacterRPGStatsHandler   characterrpgstatspb.CharacterRPGStatsServiceServer
	ArtifactRPGStatsHandler    artifactrpgstatspb.ArtifactRPGStatsServiceServer
	InventoryHandler           inventorypb.InventoryServiceServer
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
	if handlers.ContentBlockHandler != nil {
		grpcServer.RegisterContentBlockService(handlers.ContentBlockHandler)
	}
	if handlers.ContentAnchorHandler != nil {
		grpcServer.RegisterContentAnchorService(handlers.ContentAnchorHandler)
	}
	if handlers.WorldHandler != nil {
		grpcServer.RegisterWorldService(handlers.WorldHandler)
	}
	if handlers.TraitHandler != nil {
		grpcServer.RegisterTraitService(handlers.TraitHandler)
	}
	if handlers.ArchetypeHandler != nil {
		grpcServer.RegisterArchetypeService(handlers.ArchetypeHandler)
	}
	if handlers.LocationHandler != nil {
		grpcServer.RegisterLocationService(handlers.LocationHandler)
	}
	if handlers.CharacterHandler != nil {
		grpcServer.RegisterCharacterService(handlers.CharacterHandler)
	}
	if handlers.ArtifactHandler != nil {
		grpcServer.RegisterArtifactService(handlers.ArtifactHandler)
	}
	if handlers.EventHandler != nil {
		grpcServer.RegisterEventService(handlers.EventHandler)
	}
	if handlers.RPGSystemHandler != nil {
		grpcServer.RegisterRPGSystemService(handlers.RPGSystemHandler)
	}
	if handlers.SkillHandler != nil {
		grpcServer.RegisterSkillService(handlers.SkillHandler)
	}
	if handlers.RPGClassHandler != nil {
		grpcServer.RegisterRPGClassService(handlers.RPGClassHandler)
	}
	if handlers.CharacterSkillHandler != nil {
		grpcServer.RegisterCharacterSkillService(handlers.CharacterSkillHandler)
	}
	if handlers.CharacterRPGStatsHandler != nil {
		grpcServer.RegisterCharacterRPGStatsService(handlers.CharacterRPGStatsHandler)
	}
	if handlers.ArtifactRPGStatsHandler != nil {
		grpcServer.RegisterArtifactRPGStatsService(handlers.ArtifactRPGStatsHandler)
	}
	if handlers.InventoryHandler != nil {
		grpcServer.RegisterInventoryService(handlers.InventoryHandler)
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

