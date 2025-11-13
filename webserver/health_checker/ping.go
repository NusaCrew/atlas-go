package health_checker

import (
	"context"
	"time"

	"github.com/NusaCrew/atlas-go/log"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type ServicePinger interface {
	Ping(ctx context.Context) error
}

func PingServiceHealth(ctx context.Context, heathServer *health.Server, serviceName string, duration time.Duration, pinger ServicePinger) {
	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		currentStatus := grpc_health_v1.HealthCheckResponse_NOT_SERVING

		checkHealth := func() grpc_health_v1.HealthCheckResponse_ServingStatus {
			if err := pinger.Ping(ctx); err != nil {
				return grpc_health_v1.HealthCheckResponse_NOT_SERVING
			}
			return grpc_health_v1.HealthCheckResponse_SERVING
		}

		updateStatus := func(newStatus grpc_health_v1.HealthCheckResponse_ServingStatus) {
			if currentStatus == newStatus {
				return
			}

			heathServer.SetServingStatus(serviceName, newStatus)
			if newStatus == grpc_health_v1.HealthCheckResponse_NOT_SERVING {
				log.Error("service %s status changed from %s to %s", serviceName, currentStatus.String(), newStatus.String())
			}
			currentStatus = newStatus
		}

		updateStatus(checkHealth())

		for {
			select {
			case <-ticker.C:
				updateStatus(checkHealth())
			case <-ctx.Done():
				updateStatus(grpc_health_v1.HealthCheckResponse_NOT_SERVING)
				log.Info("context cancelled, service %s status changed to %s", serviceName, grpc_health_v1.HealthCheckResponse_NOT_SERVING.String())
				return
			}
		}
	}()
}
