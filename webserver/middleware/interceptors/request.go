package interceptors

import (
	"context"
	"strings"

	"github.com/NusaCrew/atlas-go/log"

	"google.golang.org/grpc"
)

const (
	LogCorrelationKey = "x-correlation-id"
)

func RequestLogger(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if strings.Contains(info.FullMethod, "Ping") || strings.Contains(info.FullMethod, "Health") {
		return handler(ctx, req)
	}

	log.Debug("%s: request: %+v", info.FullMethod, req)
	return handler(ctx, req)
}
