package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CommonRepositoryInterface interface {
	Ping(ctx context.Context) error
	GetCollection(name string) *mongo.Collection
	RunInTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error
}

type CommonRepository struct {
	Storage
}

func (c *CommonRepository) Ping(ctx context.Context) error {
	return c.Storage.Ping(ctx)
}

func (c *CommonRepository) GetCollection(name string) *mongo.Collection {
	return c.DB().Collection(name)
}

func (c *CommonRepository) RunInTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
	session, err := c.DB().Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	txnOpts := options.Transaction()

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		return nil, fn(sessCtx)
	}, txnOpts)

	return err
}
