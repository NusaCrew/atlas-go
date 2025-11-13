package webserver

import (
	"context"
	"fmt"
	"net"
	"time"

	api_v1 "github.com/NusaCrew/atlas-go/example/protos/api/v1"
	"github.com/NusaCrew/atlas-go/log"
	"github.com/NusaCrew/atlas-go/webserver/health_checker"
	"github.com/NusaCrew/atlas-go/webserver/middleware/interceptors"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type GRPCWebServerConfig struct {
	ServiceName                string
	Port                       int
	PingService                health_checker.ServicePinger
	GRPCServiceServerRegistrar func(grpcServer *grpc.Server)
	GRPCInterceptors           []grpc.UnaryServerInterceptor
}

type grpcServer struct {
	grpc         *grpc.Server
	grpcAddr     string
	grpcListener net.Listener
}

func NewGRPCWebServer(ctx context.Context, config GRPCWebServerConfig) (WebServer, error) {
	if config.GRPCServiceServerRegistrar == nil {
		return nil, fmt.Errorf("cannot run webserver without initialization server")
	}

	if config.PingService == nil {
		return nil, fmt.Errorf("cannot run webserver without ping service for health check")
	}

	interceptor := &interceptors.GRPCInterceptorWithTracer{
		ServiceName: config.ServiceName,
	}

	baseInterceptors := []grpc.UnaryServerInterceptor{
		grpc_prometheus.UnaryServerInterceptor,
		interceptor.Intercept,
		interceptors.ValidateRequest,
		interceptors.RequestLogger,
	}

	allInterceptors := append(baseInterceptors, config.GRPCInterceptors...)

	grpcSvc := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.ChainUnaryInterceptor(allInterceptors...),
	)
	grpc_prometheus.Register(grpcSvc)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSvc, healthServer)
	healthHandler := health_checker.NewHealthHandler(healthServer, config.ServiceName)
	api_v1.RegisterHealthServiceServer(grpcSvc, healthHandler)

	health_checker.PingServiceHealth(ctx, healthServer, config.ServiceName, 5*time.Second, config.PingService)

	config.GRPCServiceServerRegistrar(grpcSvc)

	grpcAddr := fmt.Sprintf(":%d", config.Port)
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return nil, err
	}

	return &grpcServer{
		grpc:         grpcSvc,
		grpcAddr:     grpcAddr,
		grpcListener: listener,
	}, nil
}

func (s *grpcServer) Run(ctx context.Context, errorChannel chan error) {
	log.Info("starting gRPC server on %s", s.grpcAddr)
	go func() { errorChannel <- s.grpc.Serve(s.grpcListener) }()
}

func (s *grpcServer) GetName() string {
	return "gRPC Server"
}

func (s *grpcServer) Stop() {
	s.grpc.GracefulStop()
	log.Info("stopping gRPC server on %s", s.grpcAddr)
}
