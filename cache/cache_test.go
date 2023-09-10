package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prodadidb/gocache/cache"
	"github.com/prodadidb/gocache/codec"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mock_store_interface_test.go -package=cache_test -source=../store/interface.go

func TestNew(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	s := NewMockStoreInterface(ctrl)

	// When
	ch := cache.New[any](s)

	// Then
	assert.IsType(t, new(cache.Cache[any]), ch)
	assert.IsType(t, new(codec.Codec), ch.Codec)

	assert.Equal(t, s, ch.Codec.GetStore())
}

func TestCacheSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	mockedStore := NewMockStoreInterface(ctrl)
	mockedStore.EXPECT().Set(ctx, "my-key", value, store.OptionsMatcher{
		Expiration: 5 * time.Second,
	}).Return(nil)

	ch := cache.New[any](mockedStore)

	// When
	err := ch.Set(ctx, "my-key", value, store.WithExpiration(5*time.Second))
	assert.Nil(t, err)
}

func TestCacheSetWhenErrorOccurs(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	storeErr := errors.New("an error has occurred while inserting data into store")

	mockedStore := NewMockStoreInterface(ctrl)
	mockedStore.EXPECT().Set(ctx, "my-key", value, store.OptionsMatcher{
		Expiration: 5 * time.Second,
	}).Return(storeErr)

	ch := cache.New[any](mockedStore)

	// When
	err := ch.Set(ctx, "my-key", value, store.WithExpiration(5*time.Second))
	assert.Equal(t, storeErr, err)
}

func TestCacheGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)

	ch := cache.New[any](s)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestCacheGetWhenNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	returnedErr := errors.New("unable to find item in store")

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Get(ctx, "my-key").Return(nil, returnedErr)

	ch := cache.New[any](s)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, returnedErr, err)
}

func TestCacheGetWithTTL(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}
	expiration := 1 * time.Second

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().GetWithTTL(ctx, "my-key").
		Return(cacheValue, expiration, nil)

	ch := cache.New[any](s)

	// When
	value, ttl, err := ch.GetWithTTL(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
	assert.Equal(t, expiration, ttl)
}

func TestCacheGetWithTTLWhenNotFound(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	returnedErr := errors.New("unable to find item in store")
	expiration := 0 * time.Second

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().GetWithTTL(ctx, "my-key").
		Return(nil, expiration, returnedErr)

	ch := cache.New[any](s)

	// When
	value, ttl, err := ch.GetWithTTL(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, returnedErr, err)
	assert.Equal(t, expiration, ttl)
}

func TestCacheGetCodec(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	s := NewMockStoreInterface(ctrl)

	ch := cache.New[any](s)

	// When
	value := ch.GetCodec()

	// Then
	assert.IsType(t, new(codec.Codec), value)
	assert.Equal(t, s, value.GetStore())
}

func TestCacheGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	s := NewMockStoreInterface(ctrl)

	ch := cache.New[any](s)

	// When - Then
	assert.Equal(t, cache.CacheType, ch.GetType())
}

func TestCacheGetCacheKeyWhenKeyIsString(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	s := NewMockStoreInterface(ctrl)

	ch := cache.New[any](s)

	// When
	computedKey := ch.GetCacheKey("my-Key")

	// Then
	assert.Equal(t, "my-Key", computedKey)
}

func TestCacheGetCacheKeyWhenKeyIsStruct(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	s := NewMockStoreInterface(ctrl)

	ch := cache.New[any](s)

	// When
	key := &struct {
		Hello string
	}{
		Hello: "world",
	}

	computedKey := ch.GetCacheKey(key)

	// Then
	assert.Equal(t, "8144fe5310cf0e62ac83fd79c113aad2", computedKey)
}

type StructWithGenerator struct{}

func (_ *StructWithGenerator) GetCacheKey() string {
	return "my-generated-key"
}

func TestCacheGetCacheKeyWhenKeyImplementsGenerator(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	s := NewMockStoreInterface(ctrl)

	ch := cache.New[any](s)

	// When
	key := &StructWithGenerator{}

	generatedKey := ch.GetCacheKey(key)
	// Then
	assert.Equal(t, "my-generated-key", generatedKey)
}

func TestCacheDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Delete(ctx, "my-key").Return(nil)

	ch := cache.New[any](s)

	// When
	err := ch.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestCacheInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	mockedStore := NewMockStoreInterface(ctrl)
	mockedStore.EXPECT().Invalidate(ctx, store.InvalidateOptionsMatcher{
		Tags: []string{"tag1"},
	}).Return(nil)

	ch := cache.New[any](mockedStore)

	// When
	err := ch.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestCacheInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error during invalidation")

	mockedStore := NewMockStoreInterface(ctrl)
	mockedStore.EXPECT().Invalidate(ctx, store.InvalidateOptionsMatcher{
		Tags: []string{"tag1"},
	}).Return(expectedErr)

	ch := cache.New[any](mockedStore)

	// When
	err := ch.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestCacheClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Clear(ctx).Return(nil)

	ch := cache.New[any](s)

	// When
	err := ch.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestCacheClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error during invalidation")

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Clear(ctx).Return(expectedErr)

	ch := cache.New[any](s)

	// When
	err := ch.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestCacheDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	s := NewMockStoreInterface(ctrl)
	s.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	ch := cache.New[any](s)

	// When
	err := ch.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}
