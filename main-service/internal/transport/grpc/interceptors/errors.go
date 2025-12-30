package interceptors

import (
	"context"

	platformerrors "github.com/story-engine/main-service/internal/platform/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorInterceptor maps domain errors to gRPC status codes
func ErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		// Map domain errors to gRPC status codes
		if platformerrors.IsNotFound(err) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}

		if platformerrors.IsAlreadyExists(err) {
			return nil, status.Errorf(codes.AlreadyExists, err.Error())
		}

		if platformerrors.IsValidation(err) {
			return nil, status.Errorf(codes.InvalidArgument, err.Error())
		}

		// Check if it's already a gRPC status error
		if s, ok := status.FromError(err); ok {
			return nil, s.Err()
		}

		// Default to Internal error
		return nil, status.Errorf(codes.Internal, "internal server error: %v", err)
	}
}

