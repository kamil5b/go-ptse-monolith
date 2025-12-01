package unitofwork

import (
	"context"
	"go-modular-monolith/pkg/constant"
	"go-modular-monolith/pkg/util"

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
	return context.WithValue(ctx, constant.ContextKeyMongoSession, session)
}

func (u *MongoUnitOfWork) DeferErrorContext(ctx context.Context, err error) error {
	session := *util.GetObjectFromContext[mongo.Session](ctx, constant.ContextKeyMongoSession)
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
