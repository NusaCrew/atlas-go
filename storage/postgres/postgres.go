package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/NusaCrew/atlas-go/config"
	"github.com/NusaCrew/atlas-go/log"

	"github.com/golang-migrate/migrate/v4"
	psqlMigrator "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Storage interface {
	DB() *sql.DB
	Ping(ctx context.Context) error
	Close() error
}

type client struct {
	client *sql.DB
	dbName string
}

func (c *client) DB() *sql.DB {
	return c.client
}

func (c *client) Ping(ctx context.Context) error {
	return c.client.PingContext(ctx)
}

func (c *client) Close() error {
	err := c.client.Close()
	if err != nil {
		return fmt.Errorf("failed to close postgres connection: %w", err)
	}
	log.Info("successfully disconnected from postgres")
	return nil
}

func InitializeDatabase(ctx context.Context, conf config.PostgresConfig) (Storage, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		conf.Host, conf.Port, conf.DBName, conf.Username, conf.Password, conf.SSLMode)

	if conf.SSLMode != "disable" {
		connStr += fmt.Sprintf(" sslrootcert=%s", conf.SSLRootCert)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(conf.MaxIdleConns)
	db.SetMaxOpenConns(conf.MaxOpenConns)
	db.SetConnMaxIdleTime(conf.MaxIdleTime)
	db.SetConnMaxLifetime(conf.MaxLifetime)

	ctx, cancel := context.WithTimeout(ctx, conf.PingTimeout)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	c := &client{
		client: db,
		dbName: conf.DBName,
	}

	if conf.PostgresMigrationConfig.RunMigrations {
		log.Info("running postgresql migrations")

		if err = c.migrateDatabase(conf.PostgresMigrationConfig.MigrationsPath); err != nil {
			return nil, err
		}
	}

	log.Info("successfully connected to postgresql database")

	return c, nil
}

func (c *client) migrateDatabase(path string) error {
	driver, err := psqlMigrator.WithInstance(c.client, &psqlMigrator.Config{})
	if err != nil {
		return fmt.Errorf("migration driver creation failed: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance("file://"+path, c.dbName, driver)
	if err != nil {
		return fmt.Errorf("migration instance creation failed: %w", err)
	}

	if err = migrator.Migrate(getLatestMigrationVersion(path)); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func getLatestMigrationVersion(path string) uint {
	files, err := os.ReadDir(path)
	if err != nil {
		return 0
	}

	var maxVersion uint
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		parts := strings.SplitN(f.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}

		version, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			continue
		}

		if uint(version) > maxVersion {
			maxVersion = uint(version)
		}
	}

	return maxVersion
}
