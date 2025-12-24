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
	EnableCORS                 bool
	CORSAllowedOrigins         []string // If empty and EnableCORS is true, allows all origins (*)
	CORSAllowedMethods         []string // If empty, defaults to common methods
	CORSAllowedHeaders         []string // If empty, defaults to common headers
}

func corsMiddleware(config HTTPWebServerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowedOrigin := "*"
			if len(config.CORSAllowedOrigins) > 0 {
				origin := r.Header.Get("Origin")
				for _, allowed := range config.CORSAllowedOrigins {
					if allowed == origin || allowed == "*" {
						allowedOrigin = allowed
						break
					}
				}
			}

			allowedMethods := "GET, POST, PUT, DELETE, OPTIONS, PATCH"
			if len(config.CORSAllowedMethods) > 0 {
				allowedMethods = ""
				for i, method := range config.CORSAllowedMethods {
					if i > 0 {
						allowedMethods += ", "
					}
					allowedMethods += method
				}
			}

			allowedHeaders := "Content-Type, Authorization, correlation-id, authorization"
			if len(config.CORSAllowedHeaders) > 0 {
				allowedHeaders = ""
				for i, header := range config.CORSAllowedHeaders {
					if i > 0 {
						allowedHeaders += ", "
					}
					allowedHeaders += header
				}
			}

			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowedHeaders)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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

	var handler http.Handler = sMux
	if config.EnableCORS {
		handler = corsMiddleware(config)(sMux)
	}

	mux.Handle("/", handler)

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
