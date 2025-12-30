package interceptors

import (
	"context"
	"strings"

	grpcctx "github.com/story-engine/main-service/internal/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor extracts tenant_id and user_id from metadata and injects into context
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// Extract tenant_id from metadata
		tenantIDs := md.Get("tenant_id")
		if len(tenantIDs) > 0 && tenantIDs[0] != "" {
			ctx = grpcctx.WithTenantID(ctx, strings.TrimSpace(tenantIDs[0]))
		}

		// Extract user_id from metadata (optional)
		userIDs := md.Get("user_id")
		if len(userIDs) > 0 && userIDs[0] != "" {
			ctx = grpcctx.WithUserID(ctx, strings.TrimSpace(userIDs[0]))
		}

		// For story operations, tenant_id is required
		if strings.HasPrefix(info.FullMethod, "/story.StoryService/") {
			if _, ok := grpcctx.TenantIDFromContext(ctx); !ok {
				return nil, status.Errorf(codes.Unauthenticated, "tenant_id is required in metadata")
			}
		}

		return handler(ctx, req)
	}
}

