package redis

import "github.com/gomodule/redigo/redis"

type Client struct {
	pool *redis.Pool
}
