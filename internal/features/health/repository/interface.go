package repository

import "context"

// Repository defines the data access methods for health checks.
type Repository interface {
	PingDB(ctx context.Context) error
	PingRedis(ctx context.Context) error
}
