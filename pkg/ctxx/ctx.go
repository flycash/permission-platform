package ctxx

import (
	"context"
	"fmt"
)

type Key string

const (
	BizIDKey Key = "biz-id"
	UIDKey   Key = "uid"
)

func WithBizID(ctx context.Context, bizID int64) context.Context {
	return context.WithValue(ctx, BizIDKey, bizID)
}

func WithUID(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, UIDKey, uid)
}

func GetBizID(ctx context.Context) (int64, error) {
	value := ctx.Value(BizIDKey)
	if value == nil {
		return 0, fmt.Errorf("biz-id not found in context")
	}
	bizID, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid biz-id type, expected int64 got %T", value)
	}

	return bizID, nil
}

func GetUID(ctx context.Context) (int64, error) {
	value := ctx.Value(UIDKey)
	if value == nil {
		return 0, fmt.Errorf("uid not found in context")
	}

	uid, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("invalid uid type, expected int64 got %T", value)
	}

	return uid, nil
}
