package repository

import "context"

type Repository interface {
	// No methods needed for health, but defined for future extensibility.
	Ping(ctx context.Context) error
}
