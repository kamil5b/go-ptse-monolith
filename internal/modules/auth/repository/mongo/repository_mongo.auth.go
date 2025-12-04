package mongo

import (
	"context"
	"time"

	"github.com/kamil5b/go-ptse-monolith/internal/modules/auth/domain"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	credentialsCollection = "auth_credentials"
	sessionsCollection    = "auth_sessions"
)

type MongoRepository struct {
	client *mongo.Client
	dbName string
}

func NewMongoRepository(client *mongo.Client, dbName string) *MongoRepository {
	return &MongoRepository{
		client: client,
		dbName: dbName,
	}
}

func (r *MongoRepository) getCredentialsCollection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(credentialsCollection)
}

func (r *MongoRepository) getSessionsCollection() *mongo.Collection {
	return r.client.Database(r.dbName).Collection(sessionsCollection)
}

func (r *MongoRepository) StartContext(ctx context.Context) context.Context {
	return ctx
}

func (r *MongoRepository) DeferErrorContext(ctx context.Context, err error) {
}

// Credential operations

func (r *MongoRepository) CreateCredential(ctx context.Context, cred *domain.Credential) error {
	if cred.ID == "" {
		cred.ID = uuid.NewString()
	}
	cred.CreatedAt = time.Now().UTC()
	cred.IsActive = true

	_, err := r.getCredentialsCollection().InsertOne(ctx, cred)
	return err
}

func (r *MongoRepository) GetCredentialByUsername(ctx context.Context, username string) (*domain.Credential, error) {
	var cred domain.Credential
	filter := bson.M{
		"username":   username,
		"deleted_at": bson.M{"$eq": nil},
	}

	err := r.getCredentialsCollection().FindOne(ctx, filter).Decode(&cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *MongoRepository) GetCredentialByEmail(ctx context.Context, email string) (*domain.Credential, error) {
	var cred domain.Credential
	filter := bson.M{
		"email":      email,
		"deleted_at": bson.M{"$eq": nil},
	}

	err := r.getCredentialsCollection().FindOne(ctx, filter).Decode(&cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *MongoRepository) GetCredentialByUserID(ctx context.Context, userID string) (*domain.Credential, error) {
	var cred domain.Credential
	filter := bson.M{
		"user_id":    userID,
		"deleted_at": bson.M{"$eq": nil},
	}

	err := r.getCredentialsCollection().FindOne(ctx, filter).Decode(&cred)
	if err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *MongoRepository) UpdateCredential(ctx context.Context, cred *domain.Credential) error {
	now := time.Now().UTC()
	cred.UpdatedAt = &now

	filter := bson.M{"id": cred.ID, "deleted_at": bson.M{"$eq": nil}}
	update := bson.M{
		"$set": bson.M{
			"username":   cred.Username,
			"email":      cred.Email,
			"is_active":  cred.IsActive,
			"updated_at": cred.UpdatedAt,
		},
	}

	_, err := r.getCredentialsCollection().UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	now := time.Now().UTC()
	filter := bson.M{"user_id": userID, "deleted_at": bson.M{"$eq": nil}}
	update := bson.M{
		"$set": bson.M{
			"password_hash": passwordHash,
			"updated_at":    now,
		},
	}

	_, err := r.getCredentialsCollection().UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	filter := bson.M{"user_id": userID, "deleted_at": bson.M{"$eq": nil}}
	update := bson.M{
		"$set": bson.M{
			"last_login_at": now,
			"updated_at":    now,
		},
	}

	_, err := r.getCredentialsCollection().UpdateOne(ctx, filter, update)
	return err
}

// Session operations

func (r *MongoRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	if session.ID == "" {
		session.ID = uuid.NewString()
	}
	session.CreatedAt = time.Now().UTC()

	_, err := r.getSessionsCollection().InsertOne(ctx, session)
	return err
}

func (r *MongoRepository) GetSessionByToken(ctx context.Context, token string) (*domain.Session, error) {
	var session domain.Session
	filter := bson.M{
		"token":      token,
		"revoked_at": bson.M{"$eq": nil},
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}

	err := r.getSessionsCollection().FindOne(ctx, filter).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *MongoRepository) GetSessionByID(ctx context.Context, id string) (*domain.Session, error) {
	var session domain.Session
	filter := bson.M{"id": id}

	err := r.getSessionsCollection().FindOne(ctx, filter).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *MongoRepository) GetSessionsByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	filter := bson.M{
		"user_id":    userID,
		"revoked_at": bson.M{"$eq": nil},
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.getSessionsCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []domain.Session
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *MongoRepository) RevokeSession(ctx context.Context, sessionID string) error {
	now := time.Now().UTC()
	filter := bson.M{"id": sessionID}
	update := bson.M{
		"$set": bson.M{
			"revoked_at": now,
			"updated_at": now,
		},
	}

	_, err := r.getSessionsCollection().UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoRepository) RevokeAllUserSessions(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	filter := bson.M{
		"user_id":    userID,
		"revoked_at": bson.M{"$eq": nil},
	}
	update := bson.M{
		"$set": bson.M{
			"revoked_at": now,
			"updated_at": now,
		},
	}

	_, err := r.getSessionsCollection().UpdateMany(ctx, filter, update)
	return err
}

func (r *MongoRepository) DeleteExpiredSessions(ctx context.Context) error {
	filter := bson.M{
		"expires_at": bson.M{"$lt": time.Now().UTC()},
	}

	_, err := r.getSessionsCollection().DeleteMany(ctx, filter)
	return err
}
