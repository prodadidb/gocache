package cache_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prodadidb/gocache/cache"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

//go:generate mockgen -destination=mock_codec_interface_test.go -package=cache_test -source=../codec/interface.go
//go:generate mockgen -destination=mock_metrics_interface_test.go -package=cache_test -source=../metrics/interface.go
//go:generate mockgen -destination=mock_store_interface_test.go -package=cache_test -source=../store/interface.go

func TestNewMetric(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	metrics := NewMockMetricsInterface(ctrl)

	// When
	ch := cache.NewMetric[any](metrics, cache1)

	// Then
	assert.IsType(t, new(cache.MetricCache[any]), ch)

	assert.Equal(t, cache1, ch.Cache)
	assert.Equal(t, metrics, ch.Metrics)
}

func TestMetricGet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	codec1 := NewMockCodecInterface(ctrl)
	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Get(ctx, "my-key").Return(cacheValue, nil)
	cache1.EXPECT().GetCodec().Return(codec1)

	metrics := NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).AnyTimes()

	ch := cache.NewMetric[any](metrics, cache1)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMetricGetWhenChainCache(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cacheValue := &struct {
		Hello string
	}{
		Hello: "world",
	}

	store1 := NewMockStoreInterface(ctrl)
	store1.EXPECT().GetType().AnyTimes().Return("store1")

	codec1 := NewMockCodecInterface(ctrl)
	codec1.EXPECT().GetStore().AnyTimes().Return(store1)

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().GetWithTTL(ctx, "my-key").Return(cacheValue,
		0*time.Second, nil)
	cache1.EXPECT().GetCodec().AnyTimes().Return(codec1)

	chainCache := cache.NewChain[any](cache1)

	metrics := NewMockMetricsInterface(ctrl)
	metrics.EXPECT().RecordFromCodec(codec1).AnyTimes()

	ch := cache.NewMetric[any](metrics, chainCache)

	// When
	value, err := ch.Get(ctx, "my-key")

	// Then
	assert.Nil(t, err)
	assert.Equal(t, cacheValue, value)
}

func TestMetricSet(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	value := &struct {
		Hello string
	}{
		Hello: "world",
	}

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Set(ctx, "my-key", value).Return(nil)

	metrics := NewMockMetricsInterface(ctrl)

	ch := cache.NewMetric[any](metrics, cache1)

	// When
	err := ch.Set(ctx, "my-key", value)

	// Then
	assert.Nil(t, err)
}

func TestMetricDelete(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(nil)

	metrics := NewMockMetricsInterface(ctrl)

	ch := cache.NewMetric[any](metrics, cache1)

	// When
	err := ch.Delete(ctx, "my-key")

	// Then
	assert.Nil(t, err)
}

func TestMetricDeleteWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unable to delete key")

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Delete(ctx, "my-key").Return(expectedErr)

	metrics := NewMockMetricsInterface(ctrl)

	ch := cache.NewMetric[any](metrics, cache1)

	// When
	err := ch.Delete(ctx, "my-key")

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricInvalidate(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(nil)

	metrics := NewMockMetricsInterface(ctrl)

	ch := cache.NewMetric[any](metrics, cache1)

	// When
	err := ch.Invalidate(ctx)

	// Then
	assert.Nil(t, err)
}

func TestMetricInvalidateWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error while invalidating data")

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Invalidate(ctx).Return(expectedErr)

	metrics := NewMockMetricsInterface(ctrl)

	ch := cache.NewMetric[any](metrics, cache1)

	// When
	err := ch.Invalidate(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricClear(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(nil)

	metrics := NewMockMetricsInterface(ctrl)

	ch := cache.NewMetric[any](metrics, cache1)

	// When
	err := ch.Clear(ctx)

	// Then
	assert.Nil(t, err)
}

func TestMetricClearWhenError(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	ctx := context.Background()

	expectedErr := errors.New("unexpected error while clearing cache")

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	cache1.EXPECT().Clear(ctx).Return(expectedErr)

	metrics := NewMockMetricsInterface(ctrl)

	ch := cache.NewMetric[any](metrics, cache1)

	// When
	err := ch.Clear(ctx)

	// Then
	assert.Equal(t, expectedErr, err)
}

func TestMetricGetType(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	cache1 := NewMockSetterCacheInterface[any](ctrl)
	metrics := NewMockMetricsInterface(ctrl)

	ch := cache.NewMetric[any](metrics, cache1)

	// When - Then
	assert.Equal(t, cache.MetricType, ch.GetType())
}
