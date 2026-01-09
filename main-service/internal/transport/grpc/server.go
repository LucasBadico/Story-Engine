package grpc

import (
	"fmt"
	"net"

	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/interceptors"
	archetypepb "github.com/story-engine/main-service/proto/archetype"
	artifactpb "github.com/story-engine/main-service/proto/artifact"
	artifactrpgstatspb "github.com/story-engine/main-service/proto/artifact_rpg_stats"
	beatpb "github.com/story-engine/main-service/proto/beat"
	chapterpb "github.com/story-engine/main-service/proto/chapter"
	characterpb "github.com/story-engine/main-service/proto/character"
	characterskillpb "github.com/story-engine/main-service/proto/character_skill"
	characterrpgstatspb "github.com/story-engine/main-service/proto/character_rpg_stats"
	eventpb "github.com/story-engine/main-service/proto/event"
	factionpb "github.com/story-engine/main-service/proto/faction"
	inventorypb "github.com/story-engine/main-service/proto/inventory"
	lorepb "github.com/story-engine/main-service/proto/lore"
	locationpb "github.com/story-engine/main-service/proto/location"
	contentblockpb "github.com/story-engine/main-service/proto/content_block"
	rpgclasspb "github.com/story-engine/main-service/proto/rpg_class"
	rpgsystempb "github.com/story-engine/main-service/proto/rpg_system"
	scenepb "github.com/story-engine/main-service/proto/scene"
	skillpb "github.com/story-engine/main-service/proto/skill"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
	traitpb "github.com/story-engine/main-service/proto/trait"
	worldpb "github.com/story-engine/main-service/proto/world"
	entityrelationpb "github.com/story-engine/main-service/proto/entity_relation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps the gRPC server
type Server struct {
	grpcServer *grpc.Server
	config     *config.Config
	logger     logger.Logger
	listener   net.Listener
}

// NewServer creates a new gRPC server with interceptors
func NewServer(cfg *config.Config, log logger.Logger) *Server {
	// Create interceptor chain
	unaryInterceptors := []grpc.UnaryServerInterceptor{
		interceptors.RecoveryInterceptor(log),
		interceptors.LoggingInterceptor(log),
		interceptors.ErrorInterceptor(),
		interceptors.AuthInterceptor(),
	}

	// Create gRPC server options
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.MaxRecvMsgSize(cfg.GRPC.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(cfg.GRPC.MaxSendMsgSize),
	}

	grpcServer := grpc.NewServer(opts...)

	// Enable reflection if configured (useful for grpcurl and testing)
	if cfg.GRPC.EnableReflection {
		reflection.Register(grpcServer)
	}

	return &Server{
		grpcServer: grpcServer,
		config:     cfg,
		logger:     log,
	}
}

// RegisterTenantService registers the TenantService handler
func (s *Server) RegisterTenantService(handler tenantpb.TenantServiceServer) {
	tenantpb.RegisterTenantServiceServer(s.grpcServer, handler)
	s.logger.Info("TenantService registered")
}

// RegisterStoryService registers the StoryService handler
func (s *Server) RegisterStoryService(handler storypb.StoryServiceServer) {
	storypb.RegisterStoryServiceServer(s.grpcServer, handler)
	s.logger.Info("StoryService registered")
}

// RegisterWorldService registers the WorldService handler
func (s *Server) RegisterWorldService(handler worldpb.WorldServiceServer) {
	worldpb.RegisterWorldServiceServer(s.grpcServer, handler)
	s.logger.Info("WorldService registered")
}

// RegisterTraitService registers the TraitService handler
func (s *Server) RegisterTraitService(handler traitpb.TraitServiceServer) {
	traitpb.RegisterTraitServiceServer(s.grpcServer, handler)
	s.logger.Info("TraitService registered")
}

// RegisterArchetypeService registers the ArchetypeService handler
func (s *Server) RegisterArchetypeService(handler archetypepb.ArchetypeServiceServer) {
	archetypepb.RegisterArchetypeServiceServer(s.grpcServer, handler)
	s.logger.Info("ArchetypeService registered")
}

// RegisterLocationService registers the LocationService handler
func (s *Server) RegisterLocationService(handler locationpb.LocationServiceServer) {
	locationpb.RegisterLocationServiceServer(s.grpcServer, handler)
	s.logger.Info("LocationService registered")
}

// RegisterCharacterService registers the CharacterService handler
func (s *Server) RegisterCharacterService(handler characterpb.CharacterServiceServer) {
	characterpb.RegisterCharacterServiceServer(s.grpcServer, handler)
	s.logger.Info("CharacterService registered")
}

// RegisterArtifactService registers the ArtifactService handler
func (s *Server) RegisterArtifactService(handler artifactpb.ArtifactServiceServer) {
	artifactpb.RegisterArtifactServiceServer(s.grpcServer, handler)
	s.logger.Info("ArtifactService registered")
}

// RegisterEventService registers the EventService handler
func (s *Server) RegisterEventService(handler eventpb.EventServiceServer) {
	eventpb.RegisterEventServiceServer(s.grpcServer, handler)
	s.logger.Info("EventService registered")
}

// RegisterFactionService registers the FactionService handler
func (s *Server) RegisterFactionService(handler factionpb.FactionServiceServer) {
	factionpb.RegisterFactionServiceServer(s.grpcServer, handler)
	s.logger.Info("FactionService registered")
}

// RegisterLoreService registers the LoreService handler
func (s *Server) RegisterLoreService(handler lorepb.LoreServiceServer) {
	lorepb.RegisterLoreServiceServer(s.grpcServer, handler)
	s.logger.Info("LoreService registered")
}

// RegisterChapterService registers the ChapterService handler
func (s *Server) RegisterChapterService(handler chapterpb.ChapterServiceServer) {
	chapterpb.RegisterChapterServiceServer(s.grpcServer, handler)
	s.logger.Info("ChapterService registered")
}

// RegisterSceneService registers the SceneService handler
func (s *Server) RegisterSceneService(handler scenepb.SceneServiceServer) {
	scenepb.RegisterSceneServiceServer(s.grpcServer, handler)
	s.logger.Info("SceneService registered")
}

// RegisterBeatService registers the BeatService handler
func (s *Server) RegisterBeatService(handler beatpb.BeatServiceServer) {
	beatpb.RegisterBeatServiceServer(s.grpcServer, handler)
	s.logger.Info("BeatService registered")
}

// RegisterContentBlockService registers the ContentBlockService handler
func (s *Server) RegisterContentBlockService(handler contentblockpb.ContentBlockServiceServer) {
	contentblockpb.RegisterContentBlockServiceServer(s.grpcServer, handler)
	s.logger.Info("ContentBlockService registered")
}

// RegisterContentBlockReferenceService registers the ContentBlockReferenceService handler
func (s *Server) RegisterContentBlockReferenceService(handler contentblockpb.ContentBlockReferenceServiceServer) {
	contentblockpb.RegisterContentBlockReferenceServiceServer(s.grpcServer, handler)
	s.logger.Info("ContentBlockReferenceService registered")
}

// RegisterRPGSystemService registers the RPGSystemService handler
func (s *Server) RegisterRPGSystemService(handler rpgsystempb.RPGSystemServiceServer) {
	rpgsystempb.RegisterRPGSystemServiceServer(s.grpcServer, handler)
	s.logger.Info("RPGSystemService registered")
}

// RegisterSkillService registers the SkillService handler
func (s *Server) RegisterSkillService(handler skillpb.SkillServiceServer) {
	skillpb.RegisterSkillServiceServer(s.grpcServer, handler)
	s.logger.Info("SkillService registered")
}

// RegisterRPGClassService registers the RPGClassService handler
func (s *Server) RegisterRPGClassService(handler rpgclasspb.RPGClassServiceServer) {
	rpgclasspb.RegisterRPGClassServiceServer(s.grpcServer, handler)
	s.logger.Info("RPGClassService registered")
}

// RegisterCharacterSkillService registers the CharacterSkillService handler
func (s *Server) RegisterCharacterSkillService(handler characterskillpb.CharacterSkillServiceServer) {
	characterskillpb.RegisterCharacterSkillServiceServer(s.grpcServer, handler)
	s.logger.Info("CharacterSkillService registered")
}

// RegisterCharacterRPGStatsService registers the CharacterRPGStatsService handler
func (s *Server) RegisterCharacterRPGStatsService(handler characterrpgstatspb.CharacterRPGStatsServiceServer) {
	characterrpgstatspb.RegisterCharacterRPGStatsServiceServer(s.grpcServer, handler)
	s.logger.Info("CharacterRPGStatsService registered")
}

// RegisterArtifactRPGStatsService registers the ArtifactRPGStatsService handler
func (s *Server) RegisterArtifactRPGStatsService(handler artifactrpgstatspb.ArtifactRPGStatsServiceServer) {
	artifactrpgstatspb.RegisterArtifactRPGStatsServiceServer(s.grpcServer, handler)
	s.logger.Info("ArtifactRPGStatsService registered")
}

// RegisterInventoryService registers the InventoryService handler
func (s *Server) RegisterInventoryService(handler inventorypb.InventoryServiceServer) {
	inventorypb.RegisterInventoryServiceServer(s.grpcServer, handler)
	s.logger.Info("InventoryService registered")
}

// RegisterEntityRelationService registers the EntityRelationService handler
func (s *Server) RegisterEntityRelationService(handler entityrelationpb.EntityRelationServiceServer) {
	entityrelationpb.RegisterEntityRelationServiceServer(s.grpcServer, handler)
	s.logger.Info("EntityRelationService registered")
}

// Start starts the gRPC server
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.listener = listener
	s.logger.Info("gRPC server starting", "address", addr)

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.logger.Info("gRPC server stopping")
		s.grpcServer.GracefulStop()
	}
}

// GetServer returns the underlying gRPC server (for testing)
func (s *Server) GetServer() *grpc.Server {
	return s.grpcServer
}

