package session

import (
	"context"
	"file_share/internal/entity"
	"file_share/internal/repository/converter"
	"time"

	"github.com/gomodule/redigo/redis"
)

var sessionTable = "session"

type sessionRow struct {
	Token     string    `db:"token"`
	Login     string    `db:"login"`
	Role      string    `db:"role"`
	ExpiresAt time.Time `db:"expires_at"`
}

func (r *Repository) CreateSession(ctx context.Context, key string, session entity.Session) error {
	cacheKey := r.getCacheKey(key)

	redisView := converter.SessionToRedisView(session)

	err := r.cache.HashSet(ctx, cacheKey, redisView)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetSession(ctx context.Context, token string) (entity.Session, error) {
	cacheKey := r.getCacheKey(token)

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

	return converter.SessionFromRedisView(sessionView), nil
}

func (r *Repository) DeleteSession(ctx context.Context, token string) error {
	cacheKey := r.getCacheKey(token)

	return r.cache.Del(ctx, cacheKey)
}
