package interceptors

import (
	"context"

	goProtoValidators "github.com/mwitkow/go-proto-validators"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ValidateRequest(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if validator, ok := req.(goProtoValidators.Validator); ok {
		if err := validator.Validate(); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "%s request is invalid, %s", info.FullMethod, err.Error())
		}
	}
	return handler(ctx, req)
}
