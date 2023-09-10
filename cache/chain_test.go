package cache_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/prodadidb/gocache/cache"
	"github.com/prodadidb/gocache/store"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mock_codec_interface_test.go -package=cache_test -source=../codec/interface.go
//go:generate mockgen -destination=mock_store_interface_test.go -package=cache_test -source=../store/interface.go

func TestNewChain(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ch1 := NewMockSetterCacheInterface[any](ctrl)
	ch2 := NewMockSetterCacheInterface[any](ctrl)

	// When
	ch := cache.NewChain[any](ch1, ch2)

	// Then
	assert.IsType(t, new(cache.ChainCache[any]), ch)

	assert.Equal(t, []cache.SetterCacheInterface[any]{ch1, ch2}, ch.Caches)
}

func TestChainGetCaches(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ch1 := NewMockSetterCacheInterface[any](ctrl)
	ch2 := NewMockSetterCacheInterface[any](ctrl)

	ch := cache.NewChain[any](ch1, ch2)

	// When
	caches := ch.GetCaches()

	// Then
	assert.Equal(t, []cache.SetterCacheInterface[any]{ch1, ch2}, caches)

	assert.Equal(t, ch1, caches[0])
	assert.Equal(t, ch2, caches[1])
}

func TestChainGetWhenAvailableInFirstCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	// Cache 1
	store1 := NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().AnyTimes().Return("store1")

	codec1 := NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().AnyTimes().Return(store1)

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().GetCodec().AnyTimes().Return(codec1)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValue,
		0*time.Second, nil)

	// Cache 2
	cache2 := NewMockSetterCacheInterface[any](ctrl)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Wait for data to be processed
	time.Sleep(100 * time.Millisecond)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainGetWhenAvailableInSecondCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	// Cache 1
	store1 := NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().AnyTimes().Return("store1")

	codec1 := NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().AnyTimes().Return(store1)

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().GetCodec().AnyTimes().Return(codec1)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(nil, 0*time.Second,
		errors.New("unable to find in cache 1"))
	cache1.EXPECT().Set(ctx, "my-key", cacheValue, &store.OptionsMatcher{}).AnyTimes().Return(nil)

	// Cache 2
	store2 := NewMockStoreInterface(ctrl)
	store2.EXPECT().GetType().AnyTimes().Return("store2")

	codec2 := NewMockCodecInterface(ctrl)
	codec2.EXPECT().GetStore().AnyTimes().Return(store2)

	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().GetCodec().AnyTimes().Return(codec2)
	cache2.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValue,
		0*time.Second, nil)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Wait for data to be processed
	time.Sleep(100 * time.Millisecond)

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestChainGetWhenNotAvailableInAnyCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	store1 := NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().Return("store1")

	codec1 := NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().Return(store1)

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().GetCodec().Return(codec1)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(nil, 0*time.Second,
		errors.New("unable to find in cache 1"))

	// Cache 2
	store2 := NewMockStoreInterface(ctrl)
	store2.EXPECT().GetType().Return("store2")

	codec2 := NewMockCodecInterface(ctrl)
	codec2.EXPECT().GetStore().Return(store2)

	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().GetCodec().Return(codec2)
	cache2.EXPECT().GetWithTTL(ctx, "my-key").Return(nil, 0*time.Second,
		errors.New("unable to find in cache 2"))

	ch := cache.NewChain[any](cache1, cache2)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Wait for data to be processed
	time.Sleep(100 * time.Millisecond)

	// Then
	assert.Equal(t, errors.New("unable to find in cache 2"), err)
	assert.Equal(t, nil, value)
}

func TestChainSet(t *testing.T) {
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
	cache1.EXPECT().Set(ctx, "my-key", cacheValue).Return(nil)

	// Cache 2
	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().Set(ctx, "my-key", cacheValue).Return(nil)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	err := ch.Set(ctx, "my-key", cacheValue)

	// Then
	assert.Nil(t, err)
}

func TestChainSetWhenErrorOnSetting(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	expectedErr := errors.New("an unexpected error occurred while setting data")

	// Cache 1
	store1 := NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().Return("store1")

	codec1 := NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().Return(store1)

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().GetCodec().Return(codec1)
	cache1.EXPECT().Set(ctx, "my-key", cacheValue).Return(expectedErr)

	// Cache 2
	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().Set(ctx, "my-key", cacheValue)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	err := ch.Set(ctx, "my-key", cacheValue)

	// Then
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("error 1 of 1: Unable to set item into cache with store 'store1': %s", expectedErr.Error()), err.Error())
}

func TestChainDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(nil)

	// Cache 2
	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().Delete(ctx, "my-key").Return(nil)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	err := ch.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestChainDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(errors.New("an error has occurred while deleting key"))

	// Cache 2
	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().Delete(ctx, "my-key").Return(nil)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	err := ch.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestChainInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(nil)

	// Cache 2
	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().Invalidate(ctx).Return(nil)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	err := ch.Invalidate(ctx)

	// Then
	assert.Nil(t, err)
}

func TestChainInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(errors.New("an unexpected error has occurred while invalidation data"))

	// Cache 2
	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().Invalidate(ctx).Return(nil)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	err := ch.Invalidate(ctx)

	// Then
	assert.Nil(t, err)
}

func TestChainClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(nil)

	// Cache 2
	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().Clear(ctx).Return(nil)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	err := ch.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestChainClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	// Cache 1
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(errors.New("an unexpected error has occurred while invalidation data"))

	// Cache 2
	cache2 := NewMockSetterCacheInterface[any](ctrl)
	cache2.EXPECT().Clear(ctx).Return(nil)

	ch := cache.NewChain[any](cache1, cache2)

	// When
	err := ch.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestChainGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := NewMockSetterCacheInterface[any](ctrl)

	ch := cache.NewChain[any](cache1)

	// When - Then
	assert.Equal(t, cache.ChainType, ch.GetType())
}

func TestCacheChecksum(t *testing.T) {
	testCases := []struct {
		value        any
		expectedHash string
	}{
		{value: 273273623, expectedHash: "a187c153af38575778244cb3796536da"},
		{value: "hello-world", expectedHash: "f31215be6928a6f6e0c7c1cf2c68054e"},
		{value: []byte(`hello-world`), expectedHash: "f097ebac995e666eb074e019cd39d99b"},
		{value: struct{ Label string }{}, expectedHash: "2938da2beee350d6ea988e404109f428"},
		{value: struct{ Label string }{Label: "hello-world"}, expectedHash: "4119a1c8530a0420859f1c6ecf2dc0b7"},
		{value: struct{ Label string }{Label: "hello-everyone"}, expectedHash: "1d7e7ed4acd56d2635f7cb33aa702bdd"},
	}

	for _, tc := range testCases {
		value := cache.Checksum(tc.value)

		assert.Equal(t, tc.expectedHash, value)
	}
}

func TestChainSetWhenErrorInChain(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)
	store1 := NewMockSetterCacheInterface[any](ctrl)

	store1.EXPECT().GetType().AnyTimes().Return("store1")
	codec1 := NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().AnyTimes().Return(store1)
	store1.EXPECT().GetCodec().AnyTimes().Return(codec1)

	ctx := context.Background()
	key := "test-key"
	value := "test-value"
	interError := errors.New("an issue occurred with the cache")
	store1.EXPECT().Set(ctx, key, value, nil).DoAndReturn(func(_, _, _, _ interface{}) error {
		return interError
	})

	store2 := NewMockSetterCacheInterface[any](ctrl)

	ch := cache.NewChain[any](store1, store2)

	// assert store2 set is called
	store2.EXPECT().Set(ctx, key, value, nil).Return(nil)

	// When - Then
	err := ch.Set(ctx, key, value, nil)

	expErr := errors.New("error 1 of 1: Unable to set item into cache with store 'store1': an issue occurred with the cache")
	// Then
	assert.Equal(t, expErr, err)
}
