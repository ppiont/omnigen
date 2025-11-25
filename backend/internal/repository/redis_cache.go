package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
	ttl    time.Duration
}

func NewRedisCache(client *redis.Client, logger *zap.Logger, ttl time.Duration) *RedisCache {
	return &RedisCache{
		client: client,
		logger: logger,
		ttl:    ttl,
	}
}

func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		r.logger.Error("Redis get failed", zap.String("key", key), zap.Error(err))
		return false, err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		r.logger.Error("Failed to unmarshal cache data", zap.String("key", key), zap.Error(err))
		return false, err
	}

	// Update access count/stats asynchronously
	go r.updateStats(context.Background(), key)

	return true, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		r.logger.Error("Redis set failed", zap.String("key", key), zap.Error(err))
		return err
	}

	return nil
}

func (r *RedisCache) updateStats(ctx context.Context, key string) {
	// Increment hit count for analytics
	// Implementation depends on stats collection strategy
}
