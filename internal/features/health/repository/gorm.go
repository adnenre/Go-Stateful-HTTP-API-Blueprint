package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type gormRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewRepository(db *gorm.DB, rdb *redis.Client) Repository {
	return &gormRepository{db: db, rdb: rdb}
}

func (r *gormRepository) PingDB(ctx context.Context) error {
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

// PingRedis checks Redis connectivity.
func (r *gormRepository) PingRedis(ctx context.Context) error {
	if r.rdb == nil {
		return sql.ErrConnDone // reuse same error type
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.rdb.Ping(pingCtx).Err()
}
