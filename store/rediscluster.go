package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

//go:generate mockgen -destination=./mock_store_rediscluster_interface_test.go -package=store_test -source=rediscluster.go

// RedisClusterClientInterface represents a go-redis/redis clusclient
type RedisClusterClientInterface interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	TTL(ctx context.Context, key string) *redis.DurationCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
	Set(ctx context.Context, key string, values any, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	FlushAll(ctx context.Context) *redis.StatusCmd
	SAdd(ctx context.Context, key string, members ...any) *redis.IntCmd
	SMembers(ctx context.Context, key string) *redis.StringSliceCmd
}

const (
	// RedisClusterType represents the storage type as a string value
	RedisClusterType = "rediscluster"
	// RedisClusterTagPattern represents the tag pattern to be used as a key in specified storage
	RedisClusterTagPattern = "gocache_tag_%s"
)

// RedisClusterStore is a store for Redis
type RedisClusterStore struct {
	Clusclient RedisClusterClientInterface
	Options    *Options
}

// NewRedisCluster creates a new store to Redis instance(s)
func NewRedisCluster(client RedisClusterClientInterface, options ...Option) *RedisClusterStore {
	return &RedisClusterStore{
		Clusclient: client,
		Options:    applyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *RedisClusterStore) Get(ctx context.Context, key any) (any, error) {
	object, err := s.Clusclient.Get(ctx, key.(string)).Result()
	if err == redis.Nil {
		return nil, NotFoundWithCause(err)
	}
	return object, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *RedisClusterStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	object, err := s.Clusclient.Get(ctx, key.(string)).Result()
	if err == redis.Nil {
		return nil, 0, NotFoundWithCause(err)
	}
	if err != nil {
		return nil, 0, err
	}

	ttl, err := s.Clusclient.TTL(ctx, key.(string)).Result()
	if err != nil {
		return nil, 0, err
	}

	return object, ttl, err
}

// Set defines data in Redis for given key identifier
func (s *RedisClusterStore) Set(ctx context.Context, key any, value any, options ...Option) error {
	opts := ApplyOptionsWithDefault(s.Options, options...)

	err := s.Clusclient.Set(ctx, key.(string), value, opts.Expiration).Err()
	if err != nil {
		return err
	}

	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *RedisClusterStore) setTags(ctx context.Context, key any, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(RedisTagPattern, tag)
		s.Clusclient.SAdd(ctx, tagKey, key.(string))
		s.Clusclient.Expire(ctx, tagKey, 720*time.Hour)
	}
}

// Delete removes data from Redis for given key identifier
func (s *RedisClusterStore) Delete(ctx context.Context, key any) error {
	_, err := s.Clusclient.Del(ctx, key.(string)).Result()
	return err
}

// Invalidate invalidates some cache data in Redis for given options
func (s *RedisClusterStore) Invalidate(ctx context.Context, options ...InvalidateOption) error {
	opts := applyInvalidateOptions(options...)

	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(RedisTagPattern, tag)
			cacheKeys, err := s.Clusclient.SMembers(ctx, tagKey).Result()
			if err != nil {
				continue
			}

			for _, cacheKey := range cacheKeys {
				_ = s.Delete(ctx, cacheKey)
			}

			_ = s.Delete(ctx, tagKey)
		}
	}

	return nil
}

// Clear resets all data in the store
func (s *RedisClusterStore) Clear(ctx context.Context) error {
	if err := s.Clusclient.FlushAll(ctx).Err(); err != nil {
		return err
	}

	return nil
}

// GetType returns the store type
func (s *RedisClusterStore) GetType() string {
	return RedisClusterType
}
