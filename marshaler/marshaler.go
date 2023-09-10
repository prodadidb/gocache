package marshaler

import (
	"context"

	"github.com/prodadidb/gocache/cache"
	"github.com/prodadidb/gocache/store"
	"github.com/vmihailenco/msgpack"
)

// Marshaler is the struct that marshal and unmarshal cache values
type Marshaler struct {
	Cache cache.CacheInterface[any]
}

// New creates a new marshaler that marshals/unmarshals cache values
func New(cache cache.CacheInterface[any]) *Marshaler {
	return &Marshaler{
		Cache: cache,
	}
}

// Get obtains a value from cache and unmarshal value with given object
func (c *Marshaler) Get(ctx context.Context, key any, returnObj any) (any, error) {
	result, err := c.Cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	switch v := result.(type) {
	case []byte:
		err = msgpack.Unmarshal(v, returnObj)
	case string:
		err = msgpack.Unmarshal([]byte(v), returnObj)
	}

	if err != nil {
		return nil, err
	}

	return returnObj, nil
}

// Set sets a value in cache by marshaling value
func (c *Marshaler) Set(ctx context.Context, key, object any, options ...store.Option) error {
	bytes, err := msgpack.Marshal(object)
	if err != nil {
		return err
	}

	return c.Cache.Set(ctx, key, bytes, options...)
}

// Delete removes a value from the cache
func (c *Marshaler) Delete(ctx context.Context, key any) error {
	return c.Cache.Delete(ctx, key)
}

// Invalidate invalidate cache values using given options
func (c *Marshaler) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return c.Cache.Invalidate(ctx, options...)
}

// Clear reset all cache data
func (c *Marshaler) Clear(ctx context.Context) error {
	return c.Cache.Clear(ctx)
}
