package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewMemcache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockMemcacheClientInterface(ctrl)

	// When
	s := store.NewMemcache(client, store.WithExpiration(3*time.Second))

	// Then
	assert.IsType(t, new(store.MemcacheStore), s)
	assert.Equal(t, client, s.Client)
	assert.Equal(t, &store.Options{Expiration: 3 * time.Second}, s.Options)
}

func TestMemcacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(&memcache.Item{
		Value: cacheValue,
	}, nil)

	s := store.NewMemcache(client, store.WithExpiration(3*time.Second))

	// When
	value, err := s.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMemcacheGetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	expectedErr := errors.New("an unexpected error occurred")

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, expectedErr)

	s := store.NewMemcache(client, store.WithExpiration(3*time.Second))

	// When
	value, err := s.Get(ctx, cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
}

func TestMemcacheGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(&memcache.Item{
		Value:      cacheValue,
		Expiration: int32(5),
	}, nil)

	s := store.NewMemcache(client, store.WithExpiration(3*time.Second))

	// When
	value, ttl, err := s.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, 5*time.Second, ttl)
}

func TestMemcacheGetWithTTLWhenMissingItem(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, nil)

	s := store.NewMemcache(client, store.WithExpiration(3*time.Second))

	// When
	value, ttl, err := s.GetWithTTL(ctx, cacheKey)

	// Then
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestMemcacheGetWithTTLWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	expectedErr := errors.New("an unexpected error occurred")

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, expectedErr)

	s := store.NewMemcache(client, store.WithExpiration(3*time.Second))

	// When
	value, ttl, err := s.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestMemcacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Set(&memcache.Item{
		Key:        cacheKey,
		Value:      cacheValue,
		Expiration: int32(5),
	}).Return(nil)

	s := store.NewMemcache(client, store.WithExpiration(3*time.Second))

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestMemcacheSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Set(&memcache.Item{
		Key:        cacheKey,
		Value:      cacheValue,
		Expiration: int32(3),
	}).Return(nil)

	s := store.NewMemcache(client, store.WithExpiration(3*time.Second))

	// When
	err := s.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestMemcacheSetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	expectedErr := errors.New("an unexpected error occurred")

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Set(&memcache.Item{
		Key:        cacheKey,
		Value:      cacheValue,
		Expiration: int32(3),
	}).Return(expectedErr)

	s := store.NewMemcache(client, store.WithExpiration(3*time.Second))

	// When
	err := s.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMemcacheSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	tagKey := "gocache_tag_tag1"

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Set(gomock.Any()).AnyTimes().Return(nil)
	client.EXPECT().Get(tagKey).Return(nil, memcache.ErrCacheMiss)
	client.EXPECT().Add(&memcache.Item{
		Key:        tagKey,
		Value:      []byte(cacheKey),
		Expiration: int32(store.TagKeyExpiry.Seconds()),
	}).Return(nil)

	s := store.NewMemcache(client)

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestMemcacheSetWithTagsWhenAlreadyInserted(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Set(gomock.Any()).AnyTimes().Return(nil)
	client.EXPECT().Get("gocache_tag_tag1").Return(&memcache.Item{
		Value: []byte("my-key,a-second-key"),
	}, nil)

	s := store.NewMemcache(client)

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestMemcacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Delete(cacheKey).Return(nil)

	s := store.NewMemcache(client)

	// When
	err := s.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestMemcacheDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	cacheKey := "my-key"

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Delete(cacheKey).Return(expectedErr)

	s := store.NewMemcache(client)

	// When
	err := s.Delete(ctx, cacheKey)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMemcacheInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := &memcache.Item{
		Value: []byte("a23fdf987h2svc23,jHG2372x38hf74"),
	}

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, nil)
	client.EXPECT().Delete("a23fdf987h2svc23").Return(nil)
	client.EXPECT().Delete("jHG2372x38hf74").Return(nil)

	s := store.NewMemcache(client)

	// When
	err := s.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestMemcacheInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := &memcache.Item{
		Value: []byte("a23fdf987h2svc23,jHG2372x38hf74"),
	}

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, nil)
	client.EXPECT().Delete("a23fdf987h2svc23").Return(errors.New("unexpected error"))
	client.EXPECT().Delete("jHG2372x38hf74").Return(nil)

	s := store.NewMemcache(client)

	// When
	err := s.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestMemcacheClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().FlushAll().Return(nil)

	s := store.NewMemcache(client)

	// When
	err := s.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestMemcacheClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("an unexpected error occurred")

	client := NewMockMemcacheClientInterface(ctrl)
	client.EXPECT().FlushAll().Return(expectedErr)

	s := store.NewMemcache(client)

	// When
	err := s.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMemcacheGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockMemcacheClientInterface(ctrl)

	s := store.NewMemcache(client)

	// When - Then
	assert.Equal(t, store.MemcacheType, s.GetType())
}
