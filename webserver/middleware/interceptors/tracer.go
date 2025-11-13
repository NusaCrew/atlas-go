package interceptors

import (
	"context"
	"strings"

	"github.com/NusaCrew/atlas-go/log"

	"google.golang.org/grpc"
)

type GRPCInterceptorWithTracer struct {
	ServiceName string
}

func (i *GRPCInterceptorWithTracer) Intercept(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	methods := strings.Split(info.FullMethod, "/")
	methodName := methods[len(methods)-1]

	tracer := log.NewTracer(ctx, methodName, i.ServiceName)
	resp, err := handler(ctx, req)

	tracer.TraceResponse(err)
	return resp, err

}
