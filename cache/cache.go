package cache

import (
	"context"
	"crypto"
	"fmt"
	"reflect"
	"time"

	"github.com/prodadidb/gocache/codec"
	"github.com/prodadidb/gocache/store"
)

const (
	// CacheType represents the cache type as a string value
	CacheType = "cache"
)

// Cache represents the configuration needed by a cache
type Cache[T any] struct {
	Codec codec.CodecInterface
}

// New instantiates a new cache entry
func New[T any](store store.StoreInterface) *Cache[T] {
	return &Cache[T]{
		Codec: codec.New(store),
	}
}

// Get returns the object stored in cache if it exists
func (c *Cache[T]) Get(ctx context.Context, key any) (T, error) {
	cacheKey := c.GetCacheKey(key)

	value, err := c.Codec.Get(ctx, cacheKey)
	if err != nil {
		return *new(T), err
	}

	if v, ok := value.(T); ok {
		return v, nil
	}

	return *new(T), nil
}

// GetWithTTL returns the object stored in cache and its corresponding TTL
func (c *Cache[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
	cacheKey := c.GetCacheKey(key)

	value, duration, err := c.Codec.GetWithTTL(ctx, cacheKey)
	if err != nil {
		return *new(T), duration, err
	}

	if v, ok := value.(T); ok {
		return v, duration, nil
	}

	return *new(T), duration, nil
}

// Set populates the cache item using the given key
func (c *Cache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	cacheKey := c.GetCacheKey(key)
	return c.Codec.Set(ctx, cacheKey, object, options...)
}

// Delete removes the cache item using the given key
func (c *Cache[T]) Delete(ctx context.Context, key any) error {
	cacheKey := c.GetCacheKey(key)
	return c.Codec.Delete(ctx, cacheKey)
}

// Invalidate invalidates cache item from given options
func (c *Cache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return c.Codec.Invalidate(ctx, options...)
}

// Clear resets all cache data
func (c *Cache[T]) Clear(ctx context.Context) error {
	return c.Codec.Clear(ctx)
}

// GetCodec returns the current codec
func (c *Cache[T]) GetCodec() codec.CodecInterface {
	return c.Codec
}

// GetType returns the cache type
func (c *Cache[T]) GetType() string {
	return CacheType
}

// GetCacheKey returns the cache key for the given key object by returning
// the key if type is string or by computing a Checksum of key structure
// if its type is other than string
func (c *Cache[T]) GetCacheKey(key any) string {
	switch v := key.(type) {
	case string:
		return v
	case CacheKeyGenerator:
		return v.GetCacheKey()
	default:
		return Checksum(key)
	}
}

// Checksum hashes a given object into a string
func Checksum(object any) string {
	digester := crypto.MD5.New()
	fmt.Fprint(digester, reflect.TypeOf(object))
	fmt.Fprint(digester, object)
	hash := digester.Sum(nil)

	return fmt.Sprintf("%x", hash)
}
