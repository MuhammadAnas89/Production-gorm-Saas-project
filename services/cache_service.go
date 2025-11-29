package services

import (
	"encoding/json"
	"go-multi-tenant/config"
	"time"
)

type CacheService struct{}

func NewCacheService() *CacheService {
	return &CacheService{}
}

// 1. Data Save Karna (Set)
func (s *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
	// Go Struct ko JSON mein convert karo
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}
	// Config se RedisClient use karo
	return config.RedisClient.Set(config.Ctx, key, jsonBytes, ttl).Err()
}

// 2. Data Fetch Karna (Get)
func (s *CacheService) Get(key string, dest interface{}) error {
	val, err := config.RedisClient.Get(config.Ctx, key).Result()
	if err != nil {
		return err
	}
	// JSON ko wapis Struct mein convert karo
	return json.Unmarshal([]byte(val), dest)
}

// 3. Data Delete Karna
func (s *CacheService) Delete(key string) error {
	return config.RedisClient.Del(config.Ctx, key).Err()
}

// 4. Pattern Clear Karna (Optional: Jaise "tenant:1:*")
func (s *CacheService) ClearPattern(pattern string) error {
	iter := config.RedisClient.Scan(config.Ctx, 0, pattern, 0).Iterator()
	for iter.Next(config.Ctx) {
		err := config.RedisClient.Del(config.Ctx, iter.Val()).Err()
		if err != nil {
			return err
		}
	}
	return iter.Err()
}
