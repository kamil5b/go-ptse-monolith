package constant

type ContextKey string

const (
	ContextKeyRequestID    ContextKey = "request_id"
	ContextKeyPostgresTx   ContextKey = "postgres_tx"
	ContextKeyMongoSession ContextKey = "mongo_session"
)
