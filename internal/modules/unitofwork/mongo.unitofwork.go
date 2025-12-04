package unitofwork

import (
	"context"

	sharedCtx "github.com/kamil5b/go-ptse-monolith/internal/shared/context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MongoUnitOfWork struct {
	client *mongo.Client
}

func NewMongoUnitOfWork(client *mongo.Client) *MongoUnitOfWork {
	return &MongoUnitOfWork{client: client}
}

func (u *MongoUnitOfWork) StartContext(ctx context.Context) context.Context {
	session, err := u.client.StartSession()
	if err != nil {
		return ctx
	}
	return context.WithValue(ctx, sharedCtx.MongoSessionKey, session)
}

func (u *MongoUnitOfWork) DeferErrorContext(ctx context.Context, err error) error {
	session := *sharedCtx.GetObjectFromContext[mongo.Session](ctx, sharedCtx.MongoSessionKey)
	if session != nil {
		defer session.EndSession(ctx)
		if err != nil {
			// abort transaction
			return session.AbortTransaction(ctx)
		} else {
			return session.CommitTransaction(ctx)
		}
	}
	return nil
}
