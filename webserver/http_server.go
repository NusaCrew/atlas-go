package webserver

import (
	"context"
	"fmt"
	"net/http"

	api_v1 "github.com/NusaCrew/atlas-go/example/protos/api/v1"
	"github.com/NusaCrew/atlas-go/log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

type httpServer struct {
	httpAddr string
	mux      *http.ServeMux
}

func AllowCorrelationID(key string) (string, bool) {
	if key == "correlation-id" {
		return key, true
	}
	return runtime.DefaultHeaderMatcher(key)
}

type HTTPWebServerConfig struct {
	GRPCHost                   string
	GRPCPort                   int
	HTTPPort                   int
	HTTPServiceServerRegistrar func(ctx context.Context, sMux *runtime.ServeMux, addr string, dialOpts []grpc.DialOption) error
}

func NewHTTPWebServer(ctx context.Context, config HTTPWebServerConfig) (WebServer, error) {
	sMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(AllowCorrelationID),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
		}),
	)

	addr := fmt.Sprintf("%s:%d", config.GRPCHost, config.GRPCPort)

	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := api_v1.RegisterHealthServiceHandlerFromEndpoint(ctx, sMux, addr, dialOpts)
	if err != nil {
		return nil, err
	}

	err = config.HTTPServiceServerRegistrar(ctx, sMux, addr, dialOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to register HTTP service server with error: %s", err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("/", sMux)

	return &httpServer{
		httpAddr: fmt.Sprintf(":%d", config.HTTPPort),
		mux:      mux,
	}, nil
}

func (s *httpServer) Run(ctx context.Context, errorChannel chan error) {
	log.Info("starting HTTP server on %s", s.httpAddr)
	go func() { errorChannel <- http.ListenAndServe(s.httpAddr, s.mux) }()
}

func (s *httpServer) GetName() string {
	return "HTTP Server"
}

func (s *httpServer) Stop() {
	log.Info("stopping HTTP server on %s", s.httpAddr)
}
