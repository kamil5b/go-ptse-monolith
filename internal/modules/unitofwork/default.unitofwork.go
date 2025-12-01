package unitofwork

import (
	"context"

	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/mongo"
)

type DefaultUnitOfWork struct {
	db       *sqlx.DB
	mongo    *mongo.Client
	uowSQL   *SQLUnitOfWork
	uowMongo *MongoUnitOfWork
}

func NewDefaultUnitOfWork(db *sqlx.DB, mongo *mongo.Client) *DefaultUnitOfWork {
	result := &DefaultUnitOfWork{db: db, mongo: mongo}
	if db != nil {
		result.uowSQL = NewSQLUnitOfWork(db)
	}
	if mongo != nil {
		result.uowMongo = NewMongoUnitOfWork(mongo)
	}
	return result
}

func (r *DefaultUnitOfWork) StartContext(ctx context.Context) context.Context {
	if r.uowSQL != nil {
		ctx = r.uowSQL.StartContext(ctx)
	}
	if r.uowMongo != nil {
		ctx = r.uowMongo.StartContext(ctx)
	}
	return ctx
}

func (r *DefaultUnitOfWork) DeferErrorContext(ctx context.Context, err error) error {
	if r.uowMongo != nil {
		err = r.uowMongo.DeferErrorContext(ctx, err)
	}
	if r.uowSQL != nil {
		err = r.uowSQL.DeferErrorContext(ctx, err)
	}
	return err
}
