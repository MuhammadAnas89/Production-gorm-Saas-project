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

func (s *CacheService) Set(key string, value interface{}, ttl time.Duration) error {

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return config.RedisClient.Set(config.Ctx, key, jsonBytes, ttl).Err()
}

func (s *CacheService) Get(key string, dest interface{}) error {
	val, err := config.RedisClient.Get(config.Ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (s *CacheService) Delete(key string) error {
	return config.RedisClient.Del(config.Ctx, key).Err()
}

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
