package cache

import "context"

type RedisClient interface {
	Set(ctx context.Context, key string, value any) error
}
