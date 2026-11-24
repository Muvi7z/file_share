package cache

import (
	"context"
	"file_share/internal/entity"
	"file_share/internal/repository/converter"
	"time"

	"github.com/gomodule/redigo/redis"
)

func (r *Repository) SetSession(ctx context.Context, key string, session entity.Session, ttl time.Duration) error {
	cacheKey := r.getCacheKey(key)

	redisView := converter.SessionToRedisView(session)

	err := r.cache.HashSet(ctx, cacheKey, redisView)
	if err != nil {
		return err
	}

	return r.cache.Expire(ctx, cacheKey, ttl)
}

func (r *Repository) GetSession(ctx context.Context, id string) (entity.Session, error) {
	cacheKey := r.getCacheKey(id)

	values, err := r.cache.HGetAll(ctx, cacheKey)
	if err != nil {
		return entity.Session{}, err
	}

	if values == nil || len(values) == 0 {
		return entity.Session{}, entity.ErrSessionNotFound
	}

	var sessionView entity.SessionRedisView

	err = redis.ScanStruct(values, &sessionView)
	if err != nil {
		return entity.Session{}, err
	}

	return converter.SessionFromRedisView(sessionView)
}

func (r *Repository) DeleteSession(ctx context.Context, id string) error {
	cacheKey := r.getCacheKey(id)

	return r.cache.Del(ctx, cacheKey)
}
