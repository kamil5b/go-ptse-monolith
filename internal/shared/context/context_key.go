package context

// Context keys for storing values in context
type ContextKey string

const (
	UserIDKey       ContextKey = "user_id"
	RequestIDKey    ContextKey = "request_id"
	SessionKey      ContextKey = "session"
	PostgresTxKey   ContextKey = "postgres_tx"
	MongoSessionKey ContextKey = "mongo_session"
)
