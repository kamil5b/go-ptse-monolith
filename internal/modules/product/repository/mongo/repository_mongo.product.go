package product

import (
	"context"
	"errors"
	"go-modular-monolith/internal/domain/product"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	col *mongo.Collection
}

func (r *MongoRepository) StartContext(ctx context.Context) context.Context {
	return ctx
}
func (r *MongoRepository) DeferErrorContext(ctx context.Context, err error) {
	// No-op for MongoDB as it doesn't support transactions in this example
}

func NewMongoRepository(client *mongo.Client, dbName string) *MongoRepository {
	col := client.Database(dbName).Collection("products")
	return &MongoRepository{col: col}
}

func (r *MongoRepository) NewTxContext(ctx context.Context, tx *sqlx.DB) context.Context {
	return ctx
}

func (r *MongoRepository) Create(ctx context.Context, p *product.Product) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	p.CreatedAt = time.Now().UTC()
	_, err := r.col.InsertOne(ctx, p)
	return err
}

func (r *MongoRepository) GetByID(ctx context.Context, id string) (*product.Product, error) {
	var p product.Product
	if err := r.col.FindOne(ctx, bson.M{"id": id}).Decode(&p); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		return nil, err
	}
	return &p, nil
}

func (r *MongoRepository) List(ctx context.Context) ([]product.Product, error) {
	cur, err := r.col.Find(ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var res []product.Product
	for cur.Next(ctx) {
		var p product.Product
		if err := cur.Decode(&p); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

func (r *MongoRepository) Update(ctx context.Context, p *product.Product) error {
	now := time.Now().UTC()
	p.UpdatedAt = &now
	upd := bson.M{"$set": bson.M{"name": p.Name, "description": p.Description, "updated_at": p.UpdatedAt, "updated_by": p.UpdatedBy}}
	_, err := r.col.UpdateOne(ctx, bson.M{"id": p.ID}, upd, options.Update().SetUpsert(false))
	return err
}

func (r *MongoRepository) SoftDelete(ctx context.Context, id, deletedBy string) error {
	now := time.Now().UTC()
	upd := bson.M{"$set": bson.M{"deleted_at": now, "deleted_by": deletedBy}}
	_, err := r.col.UpdateOne(ctx, bson.M{"id": id}, upd)
	return err
}
