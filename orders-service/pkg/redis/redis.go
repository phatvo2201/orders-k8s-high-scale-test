package redis_helper

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

func init() {
	redisAddr := os.Getenv("REDIS_ADDR")

	RedisClient = redis.NewClient(&redis.Options{Addr: redisAddr})
	if _, err := RedisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}
}

func AcquireLock(key string, ttl time.Duration) (bool, error) {
	return RedisClient.SetNX(context.Background(), key, "locked", ttl).Result()
}

func ReleaseLock(key string) error {
	return RedisClient.Del(context.Background(), key).Err()
}
