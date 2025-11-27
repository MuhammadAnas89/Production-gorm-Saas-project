package config

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// Global Client
var RedisClient *redis.Client
var ctx = context.Background()

func InitRedis(cfg *Config) error {
	// Docker local chala rahe ho to host "localhost" hi hoga
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:        redisAddr,
		Password:    redisPassword,
		DB:          0, // Default DB
		DialTimeout: 10 * time.Second,
	})

	// Ping karke check karo connect hua ya nahi
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("✅ Redis connected successfully!")
	return nil
}

// 1. Struct ko JSON bana kar Redis mein save karna
func SetCacheStruct(key string, value interface{}, ttl time.Duration) error {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, jsonBytes, ttl).Err()
}

// 2. Redis se JSON utha kar wapis Struct banana

func GetCacheStruct(key string, dest interface{}) error {
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// 3. Cache Delete karna (Invalidation ke liye)
func DeleteCache(key string) error {
	return RedisClient.Del(ctx, key).Err()
}
