package context

import (
	"context"
)

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// WithUserID adds user ID to context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) (string, bool) {
	reqID, ok := ctx.Value(RequestIDKey).(string)
	return reqID, ok
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}
func GetObjectFromContext[T any](ctx context.Context, key any) *T {
	val := ctx.Value(key)
	obj, ok := val.(*T)
	if !ok {
		return nil
	}
	return obj
}
