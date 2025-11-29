package product

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository struct {
	col *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, dbName string) *MongoRepository {
	col := client.Database(dbName).Collection("products")
	return &MongoRepository{col: col}
}

func (r *MongoRepository) Create(p *Product) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	p.CreatedAt = time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.col.InsertOne(ctx, p)
	return err
}

func (r *MongoRepository) GetByID(id string) (*Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var p Product
	if err := r.col.FindOne(ctx, bson.M{"id": id}).Decode(&p); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, err
		}
		return nil, err
	}
	return &p, nil
}

func (r *MongoRepository) List() ([]Product, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cur, err := r.col.Find(ctx, bson.M{"deleted_at": bson.M{"$exists": false}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var res []Product
	for cur.Next(ctx) {
		var p Product
		if err := cur.Decode(&p); err != nil {
			return nil, err
		}
		res = append(res, p)
	}
	return res, nil
}

func (r *MongoRepository) Update(p *Product) error {
	now := time.Now().UTC()
	p.UpdatedAt = &now
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	upd := bson.M{"$set": bson.M{"name": p.Name, "description": p.Description, "updated_at": p.UpdatedAt, "updated_by": p.UpdatedBy}}
	_, err := r.col.UpdateOne(ctx, bson.M{"id": p.ID}, upd, options.Update().SetUpsert(false))
	return err
}

func (r *MongoRepository) SoftDelete(id, deletedBy string) error {
	now := time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	upd := bson.M{"$set": bson.M{"deleted_at": now, "deleted_by": deletedBy}}
	_, err := r.col.UpdateOne(ctx, bson.M{"id": id}, upd)
	return err
}
