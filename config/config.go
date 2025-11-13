package config

import (
	"errors"
	"time"
)

type AppConfig struct {
	Environment         string `env:"ENVIRONMENT,required"`
	ServiceName         string `env:"SERVICE_NAME" envDefault:""`
	EnableLoadingSecret bool   `env:"ENABLE_LOADING_SECRET" envDefault:"false"`

	HTTPPort int    `env:"HTTP_PORT" envDefault:"8080"`
	GRPCPort int    `env:"GRPC_PORT" envDefault:"8081"`
	GRPCHost string `env:"GRPC_HOST" envDefault:"localhost"`

	Project struct {
		ID     string `env:"PROJECT_ID" envDefault:""`
		Name   string `env:"PROJECT_NAME" envDefault:""`
		Region string `env:"PROJECT_REGION" envDefault:"asia-southeast-2"`
	}

	SecretManagerConfig
}

func (c *AppConfig) IsEnableLoadingSecret() bool {
	return c.EnableLoadingSecret
}

func (c *AppConfig) GetSecretManagerName() string {
	return c.SecretManagerName
}

func (c *AppConfig) GetSecretManagerReferenceID() string {
	return c.SecretManagerReferenceID
}

type SecretManagerConfig struct {
	SecretManagerName        string `env:"SECRET_MANAGER_NAME" envDefault:"aws"`
	SecretManagerReferenceID string `env:"SECRET_MANAGER_REFERENCE_ID" envDefault:"us-east-1"`
}

type MongoConnectionConfig struct {
	ConnectTimeout  time.Duration `env:"MONGO_CONNECT_TIMEOUT" envDefault:"10s"`
	MaxPoolSize     uint64        `env:"MONGO_MAX_POOL_SIZE" envDefault:"10"`
	MinPoolSize     uint64        `env:"MONGO_MIN_POOL_SIZE" envDefault:"0"`
	MaxConnIdleTime time.Duration `env:"MONGO_MAX_CONN_IDLE_TIME" envDefault:"5m"`
}

type MongoSSLConfig struct {
	SSLMode     string `env:"MONGO_SSL_MODE" envDefault:"disable"`
	SSLCAFile   string `env:"MONGO_SSL_CA_FILE" envDefault:""`
	SSLCertFile string `env:"MONGO_SSL_CERT_FILE" envDefault:""`
	SSLKeyFile  string `env:"MONGO_SSL_KEY_FILE" envDefault:""`
	SSLInsecure bool   `env:"MONGO_SSL_INSECURE" envDefault:"false"`
}

type MongoConfig struct {
	Host         string `env:"MONGO_HOST" envDefault:"localhost"`
	Port         int    `env:"MONGO_PORT" envDefault:"27017"`
	Username     string `env:"MONGO_USERNAME" envDefault:"mongo"`
	Password     string `env:"MONGO_PASSWORD" envDefault:"password"`
	DatabaseName string `env:"MONGO_DATABASE_NAME" envDefault:"mongo"`
	AuthSource   string `env:"MONGO_AUTH_SOURCE" envDefault:"admin"`
	MongoConnectionConfig
	MongoSSLConfig
}

type RedisConfig struct {
	Host string `env:"REDIS_HOST" envDefault:"localhost"`
	Port int    `env:"REDIS_PORT" envDefault:"6379"`
}

type PostgresMigrationConfig struct {
	RunMigrations  bool   `env:"POSTGRES_RUN_MIGRATIONS" envDefault:"false"`
	MigrationsPath string `env:"POSTGRES_MIGRATIONS_PATH" envDefault:""`
}

type PostgresSSLConfig struct {
	SSLMode     string `env:"POSTGRES_SSL_MODE" envDefault:"disable"`
	SSLRootCert string `env:"POSTGRES_SSL_ROOT_CERT" envDefault:""`
}

type PostgresConnectionConfig struct {
	PingTimeout  time.Duration `env:"POSTGRES_PING_TIMEOUT" envDefault:"5s"`
	MaxIdleTime  time.Duration `env:"POSTGRES_MAX_IDLE_TIME" envDefault:"5m"`
	MaxLifetime  time.Duration `env:"POSTGRES_MAX_LIFETIME" envDefault:"1h"`
	MaxOpenConns int           `env:"POSTGRES_MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns int           `env:"POSTGRES_MAX_IDLE_CONNS" envDefault:"5"`
}

type PostgresConfig struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	DBName   string `env:"POSTGRES_DBNAME" envDefault:"postgres"`
	Username string `env:"POSTGRES_USERNAME" envDefault:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"password"`
	PostgresConnectionConfig
	PostgresSSLConfig
	PostgresMigrationConfig
}

func (c PostgresConfig) Validate() error {
	validations := []struct {
		condition bool
		message   string
	}{
		{c.Host == "", "database host is required"},
		{c.Port == 0, "database port is required"},
		{c.DBName == "", "database name is required"},
		{c.Username == "", "database username is required"},
		{c.Password == "", "database password is required"},
		{c.PingTimeout == 0, "database ping timeout is missing"},
		{c.MaxIdleTime == 0, "database max idle time is missing"},
		{c.MaxLifetime == 0, "database max lifetime is missing"},
		{c.SSLMode == "", "ssl mode is required"},
		{c.RunMigrations && c.MigrationsPath == "", "migration config: migrations path is required"},
		{c.SSLMode != "disable" && c.SSLRootCert == "", "ssl root cert is required"},
	}

	for _, v := range validations {
		if v.condition {
			return errors.New(v.message)
		}
	}
	return nil
}
