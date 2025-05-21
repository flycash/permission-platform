package plan2

import (
	"context"
	"errors"
	"gitee.com/flycash/permission-platform/internal/repository/cache"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisClient struct {
	client redis.Cmdable
}

func (r *RedisClient) Set(ctx context.Context, key string, val any, expiration time.Duration) error {
	return r.client.Set(ctx, key, val, expiration).Err()
}

func (r *RedisClient) SetNX(ctx context.Context, key string, val any, expiration time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, val, expiration).Result()
}

func (r *RedisClient) Get(ctx context.Context, key string) cache.Value {
	r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Delete(ctx context.Context, key ...string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (r *RedisClient) Ping(ctx context.Context) error {
	res, err := r.client.Ping(ctx).Result()
	if err != nil {
		return err
	}
	if res != "PONG" {
		return errors.New("ping不通")
	}
	return nil
}
