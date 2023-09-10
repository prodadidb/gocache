package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewGoCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockGoCacheClientInterface(ctrl)
	// When
	s := store.NewGoCache(client, store.WithCost(8))

	// Then
	assert.IsType(t, new(store.GoCacheStore), s)
	assert.Equal(t, client, s.Client)
	assert.Equal(t, &store.Options{Cost: 8}, s.Options)
}

func TestGoCacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(cacheValue, true)

	s := store.NewGoCache(client)

	// When
	value, err := s.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestGoCacheGetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, false)

	s := store.NewGoCache(client)

	// When
	value, err := s.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, value)
	assert.ErrorIs(t, err, store.NotFound{})
}

func TestGoCacheGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().GetWithExpiration(cacheKey).Return(cacheValue, time.Now(), true)

	s := store.NewGoCache(client)

	// When
	value, ttl, err := s.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, int64(0), ttl.Milliseconds())
}

func TestGoCacheGetWithTTLWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().GetWithExpiration(cacheKey).Return(nil, time.Now(), false)

	s := store.NewGoCache(client)

	// When
	value, ttl, err := s.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, value)
	assert.ErrorIs(t, err, store.NotFound{})
	assert.Equal(t, 0*time.Second, ttl)
}

func TestGoCacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)

	s := store.NewGoCache(client)

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithCost(4))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)

	s := store.NewGoCache(client)

	// When
	err := s.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)
	client.EXPECT().Get("gocache_tag_tag1").Return(nil, true)
	cacheKeys := map[string]struct{}{"my-key": {}}
	client.EXPECT().Set("gocache_tag_tag1", cacheKeys, 720*time.Hour)

	s := store.NewGoCache(client)

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheSetWithTagsWhenAlreadyInserted(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Set(cacheKey, cacheValue, 0*time.Second)

	cacheKeys := map[string]struct{}{"my-key": {}, "a-second-key": {}}
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, true)

	s := store.NewGoCache(client)

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Delete(cacheKey)

	s := store.NewGoCache(client)

	// When
	err := s.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := map[string]struct{}{"a23fdf987h2svc23": {}, "jHG2372x38hf74": {}}

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, true)
	client.EXPECT().Delete("a23fdf987h2svc23")
	client.EXPECT().Delete("jHG2372x38hf74")

	s := store.NewGoCache(client)

	// When
	err := s.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, false)

	s := store.NewGoCache(client)

	// When
	err := s.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestGoCacheClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockGoCacheClientInterface(ctrl)
	client.EXPECT().Flush()

	s := store.NewGoCache(client)

	// When
	err := s.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestGoCacheGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockGoCacheClientInterface(ctrl)

	s := store.NewGoCache(client)

	// When - Then
	assert.Equal(t, store.GoCacheType, s.GetType())
}

func TestGoCacheSetTagsConcurrency(t *testing.T) {
	ctx := context.Background()

	client := cache.New(10*time.Second, 30*time.Second)
	s := store.NewGoCache(client)

	for i := 0; i < 200; i++ {
		go func(i int) {
			key := fmt.Sprintf("%d", i)

			err := s.Set(
				ctx,
				key,
				[]string{"one", "two"},
				store.WithTags([]string{"tag1", "tag2", "tag3"}),
			)
			assert.Nil(t, err, err)
		}(i)
	}
}

func TestGoCacheInvalidateConcurrency(t *testing.T) {
	ctx := context.Background()

	client := cache.New(10*time.Second, 30*time.Second)
	s := store.NewGoCache(client)

	var tags []string
	for i := 0; i < 200; i++ {
		tags = append(tags, fmt.Sprintf("tag%d", i))
	}

	for i := 0; i < 200; i++ {

		go func(i int) {
			key := fmt.Sprintf("%d", i)

			err := s.Set(ctx, key, []string{"one", "two"}, store.WithTags(tags))
			assert.Nil(t, err, err)
		}(i)

		go func(i int) {
			err := s.Invalidate(ctx, store.WithInvalidateTags([]string{fmt.Sprintf("tag%d", i)}))
			assert.Nil(t, err, err)
		}(i)

	}
}
