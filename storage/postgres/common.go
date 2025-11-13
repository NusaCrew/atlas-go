package postgres

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type CommonRepositoryInterface interface {
	Ping(ctx context.Context) error
	Builder(tx *sql.Tx) sq.StatementBuilderType
	RunInSQLTransaction(ctx context.Context, isolationLevel sql.IsolationLevel, fn func(tx *sql.Tx) error) error
}

type CommonRepository struct {
	Storage
}

func (c *CommonRepository) Ping(ctx context.Context) error {
	return c.Storage.Ping(ctx)
}

func (c *CommonRepository) Builder(tx *sql.Tx) sq.StatementBuilderType {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(c.DB())
	if tx != nil {
		builder = builder.RunWith(tx)
	}
	return builder
}

func (c *CommonRepository) RunInSQLTransaction(ctx context.Context, isolationLevel sql.IsolationLevel, fn func(tx *sql.Tx) error) error {
	tx, err := c.DB().BeginTx(ctx, &sql.TxOptions{
		Isolation: isolationLevel,
	})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	return tx.Commit()
}
