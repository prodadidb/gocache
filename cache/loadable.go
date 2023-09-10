package cache

import (
	"context"
	"sync"

	"github.com/prodadidb/gocache/store"
)

const (
	// LoadableType represents the loadable cache type as a string value
	LoadableType = "loadable"
)

type loadableKeyValue[T any] struct {
	key   any
	value T
}

type LoadFunction[T any] func(ctx context.Context, key any) (T, error)

// LoadableCache represents a cache that uses a function to load data
type LoadableCache[T any] struct {
	LoadFunc   LoadFunction[T]
	Cache      CacheInterface[T]
	SetChannel chan *loadableKeyValue[T]
	SetterWg   *sync.WaitGroup
}

// NewLoadable instanciates a new cache that uses a function to load data
func NewLoadable[T any](loadFunc LoadFunction[T], cache CacheInterface[T]) *LoadableCache[T] {
	loadable := &LoadableCache[T]{
		LoadFunc:   loadFunc,
		Cache:      cache,
		SetChannel: make(chan *loadableKeyValue[T], 10000),
		SetterWg:   &sync.WaitGroup{},
	}

	loadable.SetterWg.Add(1)
	go loadable.setter()

	return loadable
}

func (c *LoadableCache[T]) setter() {
	defer c.SetterWg.Done()

	for item := range c.SetChannel {
		_ = c.Set(context.Background(), item.key, item.value)
	}
}

// Get returns the object stored in cache if it exists
func (c *LoadableCache[T]) Get(ctx context.Context, key any) (T, error) {
	var err error

	object, err := c.Cache.Get(ctx, key)
	if err == nil {
		return object, err
	}

	// Unable to find in cache, try to load it from load function
	object, err = c.LoadFunc(ctx, key)
	if err != nil {
		return object, err
	}

	// Then, put it back in cache
	c.SetChannel <- &loadableKeyValue[T]{key, object}

	return object, err
}

// Set sets a value in available caches
func (c *LoadableCache[T]) Set(ctx context.Context, key any, object T, options ...store.Option) error {
	return c.Cache.Set(ctx, key, object, options...)
}

// Delete removes a value from cache
func (c *LoadableCache[T]) Delete(ctx context.Context, key any) error {
	return c.Cache.Delete(ctx, key)
}

// Invalidate invalidates cache item from given options
func (c *LoadableCache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return c.Cache.Invalidate(ctx, options...)
}

// Clear resets all cache data
func (c *LoadableCache[T]) Clear(ctx context.Context) error {
	return c.Cache.Clear(ctx)
}

// GetType returns the cache type
func (c *LoadableCache[T]) GetType() string {
	return LoadableType
}

func (c *LoadableCache[T]) Close() error {
	close(c.SetChannel)
	c.SetterWg.Wait()

	return nil
}
