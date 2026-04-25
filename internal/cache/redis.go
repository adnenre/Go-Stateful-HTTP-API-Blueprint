package cache

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client is the Redis client instance.
var Client *redis.Client

// InitRedis establishes a connection to Redis.
// It pings the server to verify connectivity.
// Logs success or failure and exits on error.
func InitRedis(addr string) {
	Client = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := Client.Ping(ctx).Err(); err != nil {
		slog.Error("redis connection failed", "error", err, "addr", addr)
		os.Exit(1)
	}
	slog.Info("redis connected", "addr", addr)
}
