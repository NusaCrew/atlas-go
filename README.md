# Atlas - Go
A Go library that provides reusable components and utilities for building microservices. This library aims to standardize common patterns across services and reduce boilerplate code.

## 📦 Components

### Config
Configuration management with support for environment variables and secret managers.

```go
import "github.com/NusaCrew/atlas-go/config"

type MyConfig struct {
    DBHost     string `env:"DB_HOST"`
    DBPassword string `env:"DB_PASSWORD" secret:"prod/db/password"`
}

cfg := &MyConfig{}
err := config.LoadConfig(ctx, cfg)
```

**Features:**
- Load from environment variables via `godotenv`
- Auto-fetch secrets from secret managers (AWS Secrets Manager, etc.)
    - set `ENABLE_LOADING_SECRET: "true"` and register `SECRET_MANAGER_NAME` and `SECRET_MANAGER_REFERENCE_ID` to get data from service provider
- Use struct tag `secret:"name"` for automatic secret injection

---

### Log
Structured logging with support for different log levels and field injection.

```go
import "github.com/NusaCrew/atlas-go/log"

log.Initialize("Auth Service")

log.Info("server started on port %d", 8080)
log.Error("failed to connect: %s", err.Error())
log.WithFields(map[string]any{
    "username": "John",
    "age":3
}).Info("user logged in")
```

**Features:**
- Log levels: Debug, Info, Warn, Error
- Structured logging with fields
- Distributed tracing support

---

### Secret
Secret management abstraction with support for multiple providers.

```go
import (
    "github.com/NusaCrew/atlas-go/secret"
    "github.com/NusaCrew/atlas-go/secret/factory"
)

// Create secret manager
manager, err := factory.NewSecretManager(
    ctx,
    secret.ProviderAWS,
    "us-east-1",
)

// Get secret
secretValue, err := manager.GetSecret(ctx, "venmo_api_key")
```

**Features:**
- Get secret value by key from Secret Provider

**Currently Supported Providers:**
- AWS Secrets Manager

---

### Storage

#### PostgreSQL
PostgreSQL connection management for dependency injection with automatic migrations. Config automatically parsed via environment variables.

```go
import (
    "github.com/NusaCrew/atlas-go/storage/postgres"
    "github.com/NusaCrew/atlas-go/config"
)

type MyConfig struct {
    config.PostgresConfig
}

cfg := &MyConfig{}
config.LoadConfig(ctx, cfg)

db, err := postgres.InitializeDatabase(ctx, cfg.PostgresConfig)

// Use the database
db.DB().QueryContext(ctx, "SELECT * FROM users")
```

**Features:**
- Connection pooling configuration
- Auto-run migrations on startup
- Health check via `Ping()`

#### MongoDB
MongoDB connection management with health checks.

```go
import (
    "github.com/NusaCrew/atlas-go/storage/mongo"
    "github.com/NusaCrew/atlas-go/config"
)

type MyConfig struct {
    config.MongoConfig
}

cfg := &MyConfig{}
config.LoadConfig(ctx, cfg)

client, err := mongo.InitializeDatabase(ctx, cfg.MongoConfig)

// Use the client
collection := client.DB().Collection("users").
```

#### Redis
Redis client with connection pooling.

```go
import (
    "github.com/NusaCrew/atlas-go/storage/redis"
    "github.com/NusaCrew/atlas-go/config"
)

type MyConfig struct {
    config.MongoConfig
}

cfg := &MyConfig{}
config.LoadConfig(ctx, cfg)

client, err := redis.NewRedisClient(ctx, cfg.RedisConfig)

client.Set(ctx, "key", "value", 0)
```

---

### Pagination
Pagination utilities from protos parameters to apply pagination to database.

```go
import "github.com/NusaCrew/atlas-go/pagination"

// Apply to SQL query (PostgreSQL)
query := "SELECT * FROM users"
paginatedQuery := pagination.ApplyPaginationToQuery(query, page)

// Build response
response := pagination.ConstructPaginationResponse(currentPage, pageSize, totalCount)
```

**Features:**
- Page-based and cursor-based pagination
- SQL query builder integration
- Standardized response format

---

### Webserver
HTTP and gRPC server initializer with health checks.

```go
import (
    "github.com/NusaCrew/atlas-go/webserver"
    "github.com/NusaCrew/atlas-go/config"
)

type MyConfig struct {
    config.MongoConfig
}

var cfg MyConfig

func main() {
    ctx := context.Background()
    config.LoadConfig(ctx, &cfg)

    userService := //init service somewhere as health check pinger
    servers := initServers(ctx, userService)

    rootCmd := &cobra.Command{
		Use:   "/main [sub]",
		Short: "AuthD Service",
	}

	rootCmd.AddCommand(webserver.RunServersCommand(ctx, servers...))

	if err := rootCmd.Execute(); err != nil {
		log.Error("app ended with error: %s", err.Error())
		return
	}

}

func initServers(ctx context.Context, userService ports.UserService) []webserver.WebServer {
	authHandler := handler.NewAuthHandler(userService) // init service grpc handler definition

	grpcServer, err := webserver.NewGRPCWebServer(ctx, webserver.GRPCWebServerConfig{
		ServiceName: cfg.ServiceName,
		Port:        cfg.GRPCPort,
		PingService: userService,
		RegisterGRPCServiceServer: func(grpcServer *grpc.Server) {
			api_v1.RegisterAuthServiceServer(grpcServer, authHandler)
		},
        GRPCInterceptors: nil, // adjust local interceptors as needed
	})
	if err != nil {
		log.Error("failed to initialize grpc server: %s", err.Error())
		return nil
	}

	httpServer, err := webserver.NewHTTPWebServer(ctx, webserver.HTTPWebServerConfig{
		GRPCHost: cfg.GRPCHost,
		GRPCPort: cfg.GRPCPort,
		HTTPPort: cfg.HTTPPort,
		RegisterHTTPServiceServer: func(ctx context.Context, sMux *runtime.ServeMux, addr string, dialOpts []grpc.DialOption) error {
			return api_v1.RegisterAuthServiceHandlerFromEndpoint(ctx, sMux, addr, dialOpts)
		},
	})
	if err != nil {
		log.Error("failed to initialize http server: %s", err.Error())
		return nil
	}

	return []webserver.WebServer{grpcServer, httpServer}
}


```
## 🚀 Installation

```bash
GOPROXY=direct GOSUMDB=off go get "github.com/NusaCrew/atlas-go@latest"
```

or for specific tag
```bash
GOPROXY=direct GOSUMDB=off go get "github.com/NusaCrew/atlas-go@v1.0.2"
```

## 📄 License
Nusacrew @ 2025
