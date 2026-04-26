package redisstore

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

const otpTTL = 10 * time.Minute

// GenerateOTP creates a 6‑digit numeric OTP.
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// StoreOTP stores the OTP for the given email with a TTL.
func StoreOTP(ctx context.Context, rdb *redis.Client, email, otp string) error {
	key := fmt.Sprintf("otp:%s", email)
	return rdb.Set(ctx, key, otp, otpTTL).Err()
}

// VerifyOTP checks if the provided OTP matches the stored one.
// It deletes the OTP if verification succeeds.
func VerifyOTP(ctx context.Context, rdb *redis.Client, email, otp string) (bool, error) {
	key := fmt.Sprintf("otp:%s", email)
	stored, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // not found or expired
	}
	if err != nil {
		return false, err
	}
	if stored == otp {
		// Delete after successful verification
		_ = rdb.Del(ctx, key).Err()
		return true, nil
	}
	return false, nil
}
