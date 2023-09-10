package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewRistretto(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockRistrettoClientInterface(ctrl)

	// When
	s := store.NewRistretto(client, store.WithCost(8))

	// Then
	assert.IsType(t, new(store.RistrettoStore), s)
	assert.Equal(t, client, s.Client)
	assert.Equal(t, &store.Options{Cost: 8}, s.Options)
}

func TestRistrettoGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(cacheValue, true)

	s := store.NewRistretto(client)

	// When
	value, err := s.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestRistrettoGetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, false)

	s := store.NewRistretto(client)

	// When
	value, err := s.Get(ctx, cacheKey)

	// Then
	assert.Nil(t, value)
	assert.IsType(t, &store.NotFound{}, err)
}

func TestRistrettoGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(cacheValue, true)

	s := store.NewRistretto(client)

	// When
	value, ttl, err := s.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestRistrettoGetWithTTLWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().Get(cacheKey).Return(nil, false)

	s := store.NewRistretto(client)

	// When
	value, ttl, err := s.GetWithTTL(ctx, cacheKey)

	// Then
	assert.Nil(t, value)
	assert.IsType(t, &store.NotFound{}, err)
	assert.Equal(t, 0*time.Second, ttl)
}

func TestRistrettoSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(4), 0*time.Second).Return(true)

	s := store.NewRistretto(client, store.WithCost(7))

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithCost(4))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoSetWhenNoOptionsGiven(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(7), 0*time.Second).Return(true)

	s := store.NewRistretto(client, store.WithCost(7))

	// When
	err := s.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestRistrettoSetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := "my-cache-value"

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(7), 0*time.Second).Return(false)

	s := store.NewRistretto(client, store.WithCost(7))

	// When
	err := s.Set(ctx, cacheKey, cacheValue)

	// Then
	assert.Equal(t, fmt.Errorf("An error has occurred while setting value '%v' on key '%v'", cacheValue, cacheKey), err)
}

func TestRistrettoSetWithTags(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(0), 0*time.Second).Return(true)
	client.EXPECT().Get("gocache_tag_tag1").Return(nil, true)
	client.EXPECT().SetWithTTL("gocache_tag_tag1", []byte("my-key"), int64(0), 720*time.Hour).Return(true)

	s := store.NewRistretto(client)

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoSetWithTagsWhenAlreadyInserted(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"
	cacheValue := []byte("my-cache-value")

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().SetWithTTL(cacheKey, cacheValue, int64(0), 0*time.Second).Return(true)
	client.EXPECT().Get("gocache_tag_tag1").Return([]byte("my-key,a-second-key"), true)
	client.EXPECT().SetWithTTL("gocache_tag_tag1", []byte("my-key,a-second-key"), int64(0), 720*time.Hour).Return(true)

	s := store.NewRistretto(client)

	// When
	err := s.Set(ctx, cacheKey, cacheValue, store.WithTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKey := "my-key"

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().Del(cacheKey)

	s := store.NewRistretto(client)

	// When
	err := s.Delete(ctx, cacheKey)

	// Then
	assert.Nil(t, err)
}

func TestRistrettoInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, true)
	client.EXPECT().Del("a23fdf987h2svc23")
	client.EXPECT().Del("jHG2372x38hf74")

	s := store.NewRistretto(client)

	// When
	err := s.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheKeys := []byte("a23fdf987h2svc23,jHG2372x38hf74")

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().Get("gocache_tag_tag1").Return(cacheKeys, false)

	s := store.NewRistretto(client)

	// When
	err := s.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestRistrettoClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	client := NewMockRistrettoClientInterface(ctrl)
	client.EXPECT().Clear()

	s := store.NewRistretto(client)

	// When
	err := s.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestRistrettoGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	client := NewMockRistrettoClientInterface(ctrl)

	s := store.NewRistretto(client)

	// When - Then
	assert.Equal(t, store.RistrettoType, s.GetType())
}
