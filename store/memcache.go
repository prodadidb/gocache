package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"golang.org/x/sync/errgroup"
)

//go:generate mockgen -destination=./mock_store_memcache_interface_test.go -package=store_test -source=memcache.go

// MemcacheClientInterface represents a bradfitz/gomemcache client
type MemcacheClientInterface interface {
	Get(key string) (item *memcache.Item, err error)
	Set(item *memcache.Item) error
	Delete(item string) error
	FlushAll() error
	CompareAndSwap(item *memcache.Item) error
	Add(item *memcache.Item) error
}

const (
	// MemcacheType represents the storage type as a string value
	MemcacheType = "memcache"
	// MemcacheTagPattern represents the tag pattern to be used as a key in specified storage
	MemcacheTagPattern = "gocache_tag_%s"

	TagKeyExpiry = 720 * time.Hour
)

// MemcacheStore is a store for Memcache
type MemcacheStore struct {
	Client  MemcacheClientInterface
	Options *Options
}

// NewMemcache creates a new store to Memcache instance(s)
func NewMemcache(client MemcacheClientInterface, options ...Option) *MemcacheStore {
	return &MemcacheStore{
		Client:  client,
		Options: applyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *MemcacheStore) Get(_ context.Context, key any) (any, error) {
	item, err := s.Client.Get(key.(string))
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, NotFoundWithCause(errors.New("unable to retrieve data from memcache"))
	}

	return item.Value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *MemcacheStore) GetWithTTL(_ context.Context, key any) (any, time.Duration, error) {
	item, err := s.Client.Get(key.(string))
	if err != nil {
		return nil, 0, err
	}
	if item == nil {
		return nil, 0, NotFoundWithCause(errors.New("unable to retrieve data from memcache"))
	}

	return item.Value, time.Duration(item.Expiration) * time.Second, err
}

// Set defines data in Memcache for given key identifier
func (s *MemcacheStore) Set(ctx context.Context, key any, value any, options ...Option) error {
	opts := ApplyOptionsWithDefault(s.Options, options...)

	item := &memcache.Item{
		Key:        key.(string),
		Value:      value.([]byte),
		Expiration: int32(opts.Expiration.Seconds()),
	}

	err := s.Client.Set(item)
	if err != nil {
		return err
	}

	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *MemcacheStore) setTags(ctx context.Context, key any, tags []string) {
	group, _ := errgroup.WithContext(ctx)
	for _, tag := range tags {
		currentTag := tag
		group.Go(func() error {
			tagKey := fmt.Sprintf(MemcacheTagPattern, currentTag)

			var err error
			for i := 0; i < 3; i++ {
				if err = s.addKeyToTagValue(tagKey, key); err == nil {
					return nil
				}
				// loop to retry any failure (including race conditions)
			}

			return err
		})
	}

	_ = group.Wait()
}

func (s *MemcacheStore) addKeyToTagValue(tagKey string, key any) error {
	var (
		cacheKeys = []string{}
		result    *memcache.Item
		err       error
	)

	result, err = s.Client.Get(tagKey)
	if err == nil {
		cacheKeys = strings.Split(string(result.Value), ",")
	} else if !errors.Is(err, memcache.ErrCacheMiss) {
		return err
	}

	for _, cacheKey := range cacheKeys {
		// if key already exists, nothing to do
		if cacheKey == key.(string) {
			return nil
		}
	}

	cacheKeys = append(cacheKeys, key.(string))

	newVal := []byte(strings.Join(cacheKeys, ","))

	if result == nil {
		// if key didnt exist, use Add to create only if still not there
		return s.Client.Add(&memcache.Item{
			Key:        tagKey,
			Value:      newVal,
			Expiration: int32(TagKeyExpiry.Seconds()),
		})
	}

	// update existing value
	// using CompareAndSwap to ensure not to run over writes between Get and here
	result.Value = newVal
	result.Expiration = int32(TagKeyExpiry.Seconds())
	return s.Client.CompareAndSwap(result)
}

// Delete removes data from Memcache for given key identifier
func (s *MemcacheStore) Delete(_ context.Context, key any) error {
	return s.Client.Delete(key.(string))
}

// Invalidate invalidates some cache data in Memcache for given options
func (s *MemcacheStore) Invalidate(ctx context.Context, options ...InvalidateOption) error {
	opts := applyInvalidateOptions(options...)

	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(MemcacheTagPattern, tag)
			result, err := s.Get(ctx, tagKey)
			if err != nil {
				return nil
			}

			cacheKeys := []string{}
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}

			for _, cacheKey := range cacheKeys {
				_ = s.Delete(ctx, cacheKey)
			}
		}
	}

	return nil
}

// Clear resets all data in the store
func (s *MemcacheStore) Clear(_ context.Context) error {
	return s.Client.FlushAll()
}

// GetType returns the store type
func (s *MemcacheStore) GetType() string {
	return MemcacheType
}
