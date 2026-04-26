package redisstore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const resetTokenTTL = 1 * time.Hour

// GenerateResetToken creates a cryptographically secure random token.
func GenerateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// StoreResetToken associates a reset token with a user ID.
func StoreResetToken(ctx context.Context, rdb *redis.Client, token, userID string) error {
	key := fmt.Sprintf("reset:%s", token)
	return rdb.Set(ctx, key, userID, resetTokenTTL).Err()
}

// GetUserIDFromResetToken retrieves the user ID for a given token.
func GetUserIDFromResetToken(ctx context.Context, rdb *redis.Client, token string) (string, error) {
	key := fmt.Sprintf("reset:%s", token)
	userID, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // token not found or expired
	}
	return userID, err
}

// DeleteResetToken removes the token from Redis (after use).
func DeleteResetToken(ctx context.Context, rdb *redis.Client, token string) error {
	key := fmt.Sprintf("reset:%s", token)
	return rdb.Del(ctx, key).Err()
}
