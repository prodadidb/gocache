package cache

import (
	"context"

	"github.com/prodadidb/gocache/metrics"
	"github.com/prodadidb/gocache/store"
)

const (
	// MetricType represents the metric cache type as a string value
	MetricType = "metric"
)

// MetricCache is the struct that specifies metrics available for different caches
type MetricCache[T any] struct {
	Metrics metrics.MetricsInterface
	Cache   CacheInterface[T]
}

// NewMetric creates a new cache with metrics and a given cache storage
func NewMetric[T any](metrics metrics.MetricsInterface, cache CacheInterface[T]) *MetricCache[T] {
	return &MetricCache[T]{
		Metrics: metrics,
		Cache:   cache,
	}
}

// Get obtains a value from cache and also records metrics
func (c *MetricCache[T]) Get(ctx context.Context, key any) (T, error) {
	result, err := c.Cache.Get(ctx, key)

	c.updateMetrics(c.Cache)

	return result, err
}

// Set sets a value from the cache
func (c *MetricCache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	return c.Cache.Set(ctx, key, object, options...)
}

// Delete removes a value from the cache
func (c *MetricCache[T]) Delete(ctx context.Context, key any) error {
	return c.Cache.Delete(ctx, key)
}

// Invalidate invalidates cache item from given options
func (c *MetricCache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return c.Cache.Invalidate(ctx, options...)
}

// Clear resets all cache data
func (c *MetricCache[T]) Clear(ctx context.Context) error {
	return c.Cache.Clear(ctx)
}

// Get obtains a value from cache and also records metrics
func (c *MetricCache[T]) updateMetrics(cache CacheInterface[T]) {
	switch current := cache.(type) {
	case *ChainCache[T]:
		for _, cache := range current.GetCaches() {
			c.updateMetrics(cache)
		}

	case SetterCacheInterface[T]:
		c.Metrics.RecordFromCodec(current.GetCodec())
	}
}

// GetType returns the cache type
func (c *MetricCache[T]) GetType() string {
	return MetricType
}
