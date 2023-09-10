package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/prodadidb/gocache/cache"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewLoadable(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := NewMockSetterCacheInterface[any](ctrl)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "test data loaded", nil
	}

	// When
	ch := cache.NewLoadable[any](loadFunc, cache1)

	// Then
	assert.IsType(t, new(cache.LoadableCache[any]), ch)

	assert.IsType(t, new(cache.LoadFunction[any]), &ch.LoadFunc)
	assert.Equal(t, cache1, ch.Cache)
}

func TestLoadableGetWhenAlreadyInCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return nil, errors.New("should not be called")
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestLoadableGetWhenNotAvailableInLoadFunc(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(nil, errors.New("unable to find in cache 1"))

	loadFunc := func(_ context.Context, key any) (any, error) {
		return nil, errors.New("an error has occurred while loading data from custom source")
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Then
	assert.Nil(t, value)
	assert.Equal(t, errors.New("an error has occurred while loading data from custom source"), err)
}

func TestLoadableGetWhenAvailableInLoadFunc(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	// Cache 1
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(nil, errors.New("unable to find in cache 1"))
	cache1.EXPECT().Set(ctx, "my-key", cacheValue).AnyTimes().Return(nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return cacheValue, nil
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Wait for data to be processed
	for len(ch.SetChannel) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestLoadableDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When
	err := ch.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestLoadableDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When
	err := ch.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When
	err := ch.Invalidate(ctx)

	// Then
	assert.Nil(t, err)
}

func TestLoadableInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when invalidating data")

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(expectedErr)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When
	err := ch.Invalidate(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(nil)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When
	err := ch.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestLoadableClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error when invalidating data")

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(expectedErr)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "a value", nil
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When
	err := ch.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestLoadableGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := NewMockSetterCacheInterface[any](ctrl)

	loadFunc := func(_ context.Context, key any) (any, error) {
		return "test data loaded", nil
	}

	ch := cache.NewLoadable[any](loadFunc, cache1)

	// When - Then
	assert.Equal(t, cache.LoadableType, ch.GetType())
}

func TestLoadableGocache(t *testing.T) {
	gocacheClient := gocache.New(5*time.Second, 5*time.Second)
	gocacheStore := store.NewGoCache(gocacheClient, store.WithExpiration(5*time.Second))

	cacheValue := "my-value"
	loadFunc := func(ctx context.Context, accountID any) (string, error) {
		return cacheValue, nil
	}

	ch := cache.NewLoadable[string](loadFunc, cache.New[string](gocacheStore))

	// When
	value, err := ch.Get(context.Background(), "my-key")

	// Wait for data to be processed
	for len(ch.SetChannel) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}
