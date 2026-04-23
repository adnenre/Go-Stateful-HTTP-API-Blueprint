package repository

import (
	"context"
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Ping(ctx context.Context) error {
	if r.db == nil {
		return sql.ErrConnDone
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return sqlDB.PingContext(pingCtx)
}
