package util

import "context"

func GetObjectFromContext[T any](ctx context.Context, key any) *T {
	val := ctx.Value(key)
	obj, ok := val.(*T)
	if !ok {
		return nil
	}
	return obj
}
