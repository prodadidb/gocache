package marshaler_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prodadidb/gocache/marshaler"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
	"github.com/vmihailenco/msgpack"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mock_cache_interface_test.go -package=marshaler_test -source=../cache/interface.go

type testCacheValue struct {
	Hello string
}

func TestNew(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache := NewMockCacheInterface[any](ctrl)

	// When
	m := marshaler.New(cache)

	// Then
	assert.IsType(t, new(marshaler.Marshaler), m)
	assert.Equal(t, cache, m.Cache)
}

func TestGetWhenStoreReturnsSliceOfBytes(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &testCacheValue{
		Hello: "world",
	}

	cacheValueBytes, err := msgpack.Marshal(cacheValue)
	if err != nil {
		assert.Error(t, err)
	}

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return(cacheValueBytes, nil)

	m := marshaler.New(cache)

	// When
	value, err := m.Get(ctx, "my-key", new(testCacheValue))

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestGetWhenStoreReturnsString(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &testCacheValue{
		Hello: "world",
	}

	cacheValueBytes, err := msgpack.Marshal(cacheValue)
	if err != nil {
		assert.Error(t, err)
	}

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return(string(cacheValueBytes), nil)

	m := marshaler.New(cache)

	// When
	value, err := m.Get(ctx, "my-key", new(testCacheValue))

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestGetWhenUnmarshalingError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return("unknown-string", nil)

	m := marshaler.New(cache)

	// When
	value, err := m.Get(ctx, "my-key", new(testCacheValue))

	// Then
	assert.NotNil(t, err)
	assert.Nil(t, value)
}

func TestGetWhenNotFoundInStore(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to find item in store")

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Get(ctx, "my-key").Return(nil, expectedErr)

	m := marshaler.New(cache)

	// When
	value, err := m.Get(ctx, "my-key", new(testCacheValue))

	// Then
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, value)
}

func TestSetWhenStruct(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &testCacheValue{
		Hello: "world",
	}

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Set(
		ctx,
		"my-key",
		[]byte{0x81, 0xa5, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0xa5, 0x77, 0x6f, 0x72, 0x6c, 0x64},
		store.OptionsMatcher{
			Expiration: 5 * time.Second,
		},
	).Return(nil)

	m := marshaler.New(cache)

	// When
	err := m.Set(ctx, "my-key", cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestSetWhenString(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := "test"

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Set(
		ctx,
		"my-key",
		[]byte{0xa4, 0x74, 0x65, 0x73, 0x74},
		store.OptionsMatcher{
			Expiration: 5 * time.Second,
		},
	).Return(nil)

	m := marshaler.New(cache)

	// When
	err := m.Set(ctx, "my-key", cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Nil(t, err)
}

func TestSetWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := "test"

	expectedErr := errors.New("an unexpected error occurred")

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Set(
		ctx,
		"my-key",
		[]byte{0xa4, 0x74, 0x65, 0x73, 0x74},
		store.OptionsMatcher{Expiration: 5 * time.Second},
	).Return(expectedErr)

	m := marshaler.New(cache)

	// When
	err := m.Set(ctx, "my-key", cacheValue, store.WithExpiration(5*time.Second))

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Delete(ctx, "my-key").Return(nil)

	m := marshaler.New(cache)

	// When
	err := m.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	m := marshaler.New(cache)

	// When
	err := m.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Invalidate(ctx, store.InvalidateOptionsMatcher{
		Tags: []string{"tag1"},
	}).Return(nil)

	m := marshaler.New(cache)

	// When
	err := m.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Nil(t, err)
}

func TestInvalidatingWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when invalidating data")

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Invalidate(ctx, store.InvalidateOptionsMatcher{Tags: []string{"tag1"}}).Return(expectedErr)

	m := marshaler.New(cache)

	// When
	err := m.Invalidate(ctx, store.WithInvalidateTags([]string{"tag1"}))

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Clear(ctx).Return(nil)

	m := marshaler.New(cache)

	// When
	err := m.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("an unexpected error occurred")

	cache := NewMockCacheInterface[any](ctrl)
	cache.EXPECT().Clear(ctx).Return(expectedErr)

	m := marshaler.New(cache)

	// When
	err := m.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}
