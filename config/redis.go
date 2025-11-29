package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Global Variables (Services inko use karengi)
var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis(cfg *Config) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:        cfg.RedisAddr,
		Password:    cfg.RedisPass,
		DB:          0, // Default DB
		DialTimeout: 10 * time.Second,
	})

	// Ping karke check karo connection zinda hai ya nahi
	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✅ Redis connected successfully!")
	return nil
}
