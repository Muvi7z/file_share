package redis

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

func (c *Client) SAdd(ctx context.Context, key, value string) error {
	return c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("SADD", key, value)
		return err
	})
}

func (c *Client) SRem(ctx context.Context, key, value string) error {
	return c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		_, err := conn.Do("SREM", key, value)
		return err
	})
}

func (c *Client) SIsMember(ctx context.Context, key, value string) (bool, error) {
	var isMember bool
	err := c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		result, err := redis.Int(conn.Do("SISMEMBER", redis.Args{key}.Add(value)...))
		if err != nil {
			return err
		}

		isMember = result > 0
		return nil
	})

	if err != nil {
		return false, err
	}

	return isMember, nil
}

func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	var members []string

	err := c.withConn(ctx, func(ctx context.Context, conn redis.Conn) error {
		result, err := redis.Strings(conn.Do("SMEMBERS", key))
		if err != nil {
			return err
		}

		members = result
		return nil
	})
	if err != nil {
		return nil, err
	}

	return members, nil
}
