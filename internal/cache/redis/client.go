package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

type Client struct {
	pool              *redis.Pool
	logger            Logger
	connectionTimeout time.Duration
}

type Logger interface {
	Info(ctx context.Context, message string, args ...any)
	Error(ctx context.Context, err error, args ...any)
}

type redisFn func(ctx context.Context, conn redis.Conn) error

func NewClient(pool *redis.Pool, logger Logger, connectionTimeout time.Duration) *Client {
	return &Client{
		pool:              pool,
		logger:            logger,
		connectionTimeout: connectionTimeout,
	}
}

func (c *Client) withConn(ctx context.Context, fn redisFn) error {
	conn, err := c.getConn(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := conn.Close(); cerr != nil {
			c.logger.Error(ctx, fmt.Errorf("failed to close redis connection"), zap.Error(cerr))
		}
	}()

	return fn(ctx, conn)
}

func (c *Client) getConn(ctx context.Context) (redis.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, c.connectionTimeout)
	defer cancel()

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		c.logger.Error(ctx, fmt.Errorf("failed to close redis connection"), zap.Error(err))
		return nil, err
	}

	return conn, nil
}

func (c *Client) Set(ctx context.Context, key string, value any) error {
	return c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("SET", key, value)
		return err
	})
}

func (c *Client) SetWithTTL(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("SET", key, value, "EX", int(ttl.Seconds()))
		return err
	})
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	var result []byte
	err := c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		val, err := redis.Bytes(conn.Do("GET", key))
		if err != nil {
			return err
		}

		result = val
		return nil
	})

	return result, err
}

func (c *Client) HashSet(ctx context.Context, key string, values any) error {
	return c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("HSET", redis.Args{key}.AddFlat(values)...)
		return err
	})
}

func (c *Client) HGetAll(ctx context.Context, key string) ([]any, error) {
	var values []any
	err := c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		result, err := redis.Values(conn.Do("HGETALL", key))
		if err != nil {
			return err
		}

		values = result
		return nil
	})

	return values, err
}

func (c *Client) Del(ctx context.Context, key string) error {
	return c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("DEL", key)
		return err
	})
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	var exists bool
	err := c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		val, err := redis.Bool(conn.Do("EXPIRE", key))
		if err != nil {
			return err
		}

		exists = val
		return nil
	})

	return exists, err
}
