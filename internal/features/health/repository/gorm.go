package repository

import (
	"context"
)

type gormRepository struct{}

func NewRepository() Repository {
	return &gormRepository{}
}

func (r *gormRepository) Ping(ctx context.Context) error {
	// In a real database, you would ping the connection.
	// For now, simulate a working database.
	return nil
}
