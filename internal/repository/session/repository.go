package session

import (
	"file_share/internal/cache"
	"fmt"
)

const (
	cacheKeyPrefix = "file:"
)

type Repository struct {
	cache cache.RedisClient
}

func NewRepository(cache cache.RedisClient) *Repository {
	return &Repository{cache: cache}
}

func (r *Repository) getCacheKey(uuid string) string {
	return fmt.Sprintf("%s%s", cacheKeyPrefix, uuid)
}
