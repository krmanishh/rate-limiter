package store

import (
	"context"
	"time"
)

type Store interface {
	Get(
		ctx context.Context,
		key string,
	) (string, error)

	Set(
		ctx context.Context,
		key string,
		value interface{},
		expiration time.Duration,
	) error

	Incr(
		ctx context.Context,
		key string,
	) (int64, error)

	Expire(
		ctx context.Context,
		key string,
		expiration time.Duration,
	) error

	Delete(
		ctx context.Context,
		key string,
	) error

	Eval(
		ctx context.Context,
		script string,
		keys []string,
		args ...interface{},
	) (interface{}, error)
}
