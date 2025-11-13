package mongo

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/NusaCrew/atlas-go/config"
	"github.com/NusaCrew/atlas-go/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Storage interface {
	DB() *mongo.Database
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
}

type mongoClient struct {
	client   *mongo.Client
	database *mongo.Database
}

func InitializeDatabase(ctx context.Context, conf config.MongoConfig) (Storage, error) {
	ctx, cancel := context.WithTimeout(ctx, conf.ConnectTimeout)
	defer cancel()

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=%s",
		conf.Username,
		conf.Password,
		conf.Host,
		conf.Port,
		conf.DatabaseName,
		conf.AuthSource,
	)

	opts := options.Client().ApplyURI(uri).
		SetMaxPoolSize(conf.MaxPoolSize).
		SetMinPoolSize(conf.MinPoolSize).
		SetMaxConnIdleTime(conf.MaxConnIdleTime).
		SetConnectTimeout(conf.ConnectTimeout)

	if conf.SSLMode != "disable" {
		tlsConfig, err := buildTLSConfig(conf.MongoSSLConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config for mongodb: %w", err)
		}
		opts.SetTLSConfig(tlsConfig)
	}

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongodb: %v", err)
	}

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %v", err)
	}

	log.Info("successfully connected to mongodb")

	db := client.Database(conf.DatabaseName)

	return &mongoClient{
		client:   client,
		database: db,
	}, nil
}

func (db *mongoClient) DB() *mongo.Database {
	return db.database
}

func (db *mongoClient) Ping(ctx context.Context) error {
	return db.client.Ping(ctx, readpref.Primary())
}

func (db *mongoClient) Close(ctx context.Context) error {
	if err := db.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from mongodb: %v", err)
	}
	log.Info("successfully disconnected from mongodb")
	return nil
}

func buildTLSConfig(sslConfig config.MongoSSLConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: sslConfig.SSLInsecure,
	}

	if sslConfig.SSLCAFile != "" {
		caCert, err := os.ReadFile(sslConfig.SSLCAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	if sslConfig.SSLCertFile != "" && sslConfig.SSLKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(sslConfig.SSLCertFile, sslConfig.SSLKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
