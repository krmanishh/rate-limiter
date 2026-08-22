package redisstore

import (
	"context"
	"time"

	"github.com/krmanishh/rate-limiter/internal/store"
	"github.com/redis/go-redis/v9"
)

var _ store.Store = (*Client)(nil)

type Client struct {
	client *redis.Client
}

func New(addr string) *Client {
	return NewWithPassword(addr, "")
}

// NewWithPassword connects with Redis AUTH. An empty password behaves
// exactly like New (go-redis skips AUTH when Password is "").
func NewWithPassword(addr, password string) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})

	return &Client{
		client: client,
	}
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *Client) Get(
	ctx context.Context,
	key string,
) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Set(
	ctx context.Context,
	key string,
	value interface{},
	expiration time.Duration,
) error {
	return c.client.Set(
		ctx,
		key,
		value,
		expiration,
	).Err()
}

func (c *Client) Incr(
	ctx context.Context,
	key string,
) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

func (c *Client) Expire(
	ctx context.Context,
	key string,
	expiration time.Duration,
) error {
	return c.client.Expire(
		ctx,
		key,
		expiration,
	).Err()
}

func (c *Client) Delete(
	ctx context.Context,
	key string,
) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Eval(
	ctx context.Context,
	script string,
	keys []string,
	args ...interface{},
) (interface{}, error) {
	return c.client.Eval(
		ctx,
		script,
		keys,
		args...,
	).Result()
}
