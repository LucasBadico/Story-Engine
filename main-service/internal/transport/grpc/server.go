package grpc

import (
	"fmt"
	"net"

	"github.com/story-engine/main-service/internal/platform/config"
	"github.com/story-engine/main-service/internal/platform/logger"
	"github.com/story-engine/main-service/internal/transport/grpc/interceptors"
	storypb "github.com/story-engine/main-service/proto/story"
	tenantpb "github.com/story-engine/main-service/proto/tenant"
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

