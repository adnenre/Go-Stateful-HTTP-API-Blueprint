package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client is the Redis client instance.
var Client *redis.Client

// InitRedis establishes a connection to Redis.
// It pings the server to verify connectivity.
func InitRedis(addr string) error {
	Client = redis.NewClient(&redis.Options{
		Addr: addr,
		// Optional: add password, DB, pool settings if needed
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}
	slog.Info("redis connected", "addr", addr)
	return nil
}
