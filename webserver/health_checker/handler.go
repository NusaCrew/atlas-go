package health_checker

import (
	"context"
	"fmt"

	api_v1 "github.com/NusaCrew/atlas-go/example/protos/api/v1"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type healthHandler struct {
	api_v1.UnimplementedHealthServiceServer
	healthServer grpc_health_v1.HealthServer
	serviceName  string
}

func NewHealthHandler(healthServer grpc_health_v1.HealthServer, serviceName string) api_v1.HealthServiceServer {
	return &healthHandler{
		healthServer: healthServer,
		serviceName:  serviceName,
	}
}

func (h *healthHandler) Health(ctx context.Context, req *empty.Empty) (*empty.Empty, error) {
	statusResp, err := h.healthServer.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: h.serviceName,
	})
	if err != nil {
		return nil, err
	}

	if statusResp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return nil, fmt.Errorf("health failed on service %s, current status %s", h.serviceName, statusResp.GetStatus().String())
	}

	return &empty.Empty{}, nil
}
