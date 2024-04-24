package store

import (
	"context"
	"time"
)

type Store interface {
	Get(ctx context.Context, key string) (int, error)
	IsBlocked(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string, expiration time.Duration) (int, error)
	Block(ctx context.Context, key string, blockDuration time.Duration) error
}
