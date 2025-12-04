package context

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserID(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "user123")
	userID, ok := GetUserID(ctx)

	assert.True(t, ok)
	assert.Equal(t, "user123", userID)
}

func TestGetUserIDNotSet(t *testing.T) {
	ctx := context.Background()
	userID, ok := GetUserID(ctx)

	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestGetUserIDWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, 123)
	userID, ok := GetUserID(ctx)

	assert.False(t, ok)
	assert.Empty(t, userID)
}

func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user456")

	userID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "user456", userID)
}

func TestWithUserIDPreservesParentContext(t *testing.T) {
	parentCtx := context.WithValue(context.Background(), RequestIDKey, "req123")
	ctx := WithUserID(parentCtx, "user789")

	// Should be able to get both values
	userID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "user789", userID)

	reqID, ok := GetRequestID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "req123", reqID)
}

func TestGetRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, "req123")
	reqID, ok := GetRequestID(ctx)

	assert.True(t, ok)
	assert.Equal(t, "req123", reqID)
}

func TestGetRequestIDNotSet(t *testing.T) {
	ctx := context.Background()
	reqID, ok := GetRequestID(ctx)

	assert.False(t, ok)
	assert.Empty(t, reqID)
}

func TestGetRequestIDWrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), RequestIDKey, 456)
	reqID, ok := GetRequestID(ctx)

	assert.False(t, ok)
	assert.Empty(t, reqID)
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req999")

	reqID, ok := GetRequestID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "req999", reqID)
}

func TestWithRequestIDPreservesParentContext(t *testing.T) {
	parentCtx := context.WithValue(context.Background(), UserIDKey, "user111")
	ctx := WithRequestID(parentCtx, "req111")

	// Should be able to get both values
	reqID, ok := GetRequestID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "req111", reqID)

	userID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "user111", userID)
}

func TestGetObjectFromContext(t *testing.T) {
	type TestObject struct {
		ID   string
		Name string
	}

	obj := &TestObject{ID: "123", Name: "test"}
	ctx := context.WithValue(context.Background(), "test_key", obj)

	result := GetObjectFromContext[TestObject](ctx, "test_key")
	require.NotNil(t, result)
	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "test", result.Name)
}

func TestGetObjectFromContextNotSet(t *testing.T) {
	type TestObject struct {
		ID string
	}

	ctx := context.Background()
	result := GetObjectFromContext[TestObject](ctx, "nonexistent")
	assert.Nil(t, result)
}

func TestGetObjectFromContextWrongType(t *testing.T) {
	type TestObject struct {
		ID string
	}

	ctx := context.WithValue(context.Background(), "test_key", "not_a_pointer")
	result := GetObjectFromContext[TestObject](ctx, "test_key")
	assert.Nil(t, result)
}

func TestGetObjectFromContextNilValue(t *testing.T) {
	type TestObject struct {
		ID string
	}

	ctx := context.WithValue(context.Background(), "test_key", nil)
	result := GetObjectFromContext[TestObject](ctx, "test_key")
	assert.Nil(t, result)
}

func TestContextKeyValues(t *testing.T) {
	tests := []struct {
		name string
		key  ContextKey
		val  string
	}{
		{"UserIDKey", UserIDKey, "user_id"},
		{"RequestIDKey", RequestIDKey, "request_id"},
		{"SessionKey", SessionKey, "session"},
		{"PostgresTxKey", PostgresTxKey, "postgres_tx"},
		{"MongoSessionKey", MongoSessionKey, "mongo_session"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, ContextKey(tt.val), tt.key)
		})
	}
}

func TestMultipleContextValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user1")
	ctx = WithRequestID(ctx, "req1")

	userID, _ := GetUserID(ctx)
	reqID, _ := GetRequestID(ctx)

	assert.Equal(t, "user1", userID)
	assert.Equal(t, "req1", reqID)
}

func TestContextChaining(t *testing.T) {
	ctx := WithRequestID(WithUserID(context.Background(), "user2"), "req2")

	userID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "user2", userID)

	reqID, ok := GetRequestID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "req2", reqID)
}
