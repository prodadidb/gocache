package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewRedis(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockRedisClientInterface(ctrl)

	// When
	s := store.NewRedis(client, store.WithExpiration(6*time.Second))

	// Then
	assert.IsType(t, new(store.RedisStore), s)
	assert.Equal(t, client, s.Client)
	assert.Equal(t, &store.Options{Expiration: 6 * time.Second}, s.Options)
}

func TestRedisGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().Get(ctx, "my-key").Return(&redis.StringCmd{})

	s := store.NewRedis(client)

	// When
	value, err := s.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.NotNil(t, value)
}

func TestRedisSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set(ctx, "my-key", cacheValue, 5*time.Second).Return(&redis.StatusCmd{})

	s := store.NewRedis(client, store.WithExpiration(6*time.Second))

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestRedisSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set(ctx, "my-key", cacheValue, 6*time.Second).Return(&redis.StatusCmd{})

	s := store.NewRedis(client, store.WithExpiration(6*time.Second))

	// When
	err := s.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestRedisSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().Set(ctx, cacheKey, cacheValue, time.Duration(0)).Return(&redis.StatusCmd{})
	client.EXPECT().SAdd(ctx, "gocache_tag_tag1", "my-key").Return(&redis.IntCmd{})
	client.EXPECT().Expire(ctx, "gocache_tag_tag1", 720*time.Hour).Return(&redis.BoolCmd{})

	s := store.NewRedis(client)

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRedisDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().Del(ctx, "my-key").Return(&redis.IntCmd{})

	s := store.NewRedis(client)

	// When
	err := s.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestRedisInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := &redis.StringSliceCmd{}

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().SMembers(ctx, "gocache_tag_tag1").Return(cacheKeys)
	client.EXPECT().Del(ctx, "gocache_tag_tag1").Return(&redis.IntCmd{})

	s := store.NewRedis(client)

	// When
	err := s.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRedisClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockRedisClientInterface(ctrl)
	client.EXPECT().FlushAll(ctx).Return(&redis.StatusCmd{})

	s := store.NewRedis(client)

	// When
	err := s.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestRedisGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockRedisClientInterface(ctrl)

	s := store.NewRedis(client)

	// When - Then
	assert.Equal(t, store.RedisType, s.GetType())
}
