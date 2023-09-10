package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

//go:generate mockgen -destination=./mock_store_bigcache_interface_test.go -package=store_test -source=bigcache.go

// BigcacheClientInterface represents a allegro/bigcache client
type BigcacheClientInterface interface {
	Get(key string) ([]byte, error)
	Set(key string, entry []byte) error
	Delete(key string) error
	Reset() error
}

const (
	// BigcacheType represents the storage type as a string value
	BigcacheType = "bigcache"
	// BigcacheTagPattern represents the tag pattern to be used as a key in specified storage
	BigcacheTagPattern = "gocache_tag_%s"
)

// BigcacheStore is a store for Bigcache
type BigcacheStore struct {
	Client  BigcacheClientInterface
	Options *Options
}

// NewBigcache creates a new store to Bigcache instance(s)
func NewBigcache(client BigcacheClientInterface, options ...Option) *BigcacheStore {
	return &BigcacheStore{
		Client:  client,
		Options: applyOptions(options...),
	}
}

// Get returns data stored from a given key
func (s *BigcacheStore) Get(_ context.Context, key any) (any, error) {
	item, err := s.Client.Get(key.(string))
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, NotFoundWithCause(errors.New("unable to retrieve data from bigcache"))
	}

	return item, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *BigcacheStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	item, err := s.Get(ctx, key)
	return item, 0, err
}

// Set defines data in Bigcache for given key identifier
func (s *BigcacheStore) Set(ctx context.Context, key any, value any, options ...Option) error {
	opts := ApplyOptionsWithDefault(s.Options, options...)

	var val []byte
	switch v := value.(type) {
	case string:
		val = []byte(v)
	case []byte:
		val = v
	default:
		return errors.New("value type not supported by Bigcache store")
	}

	err := s.Client.Set(key.(string), val)
	if err != nil {
		return err
	}

	if tags := opts.Tags; len(tags) > 0 {
		s.setTags(ctx, key, tags)
	}

	return nil
}

func (s *BigcacheStore) setTags(ctx context.Context, key any, tags []string) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf(BigcacheTagPattern, tag)
		cacheKeys := []string{}

		if result, err := s.Get(ctx, tagKey); err == nil {
			if bytes, ok := result.([]byte); ok {
				cacheKeys = strings.Split(string(bytes), ",")
			}
		}

		alreadyInserted := false
		for _, cacheKey := range cacheKeys {
			if cacheKey == key.(string) {
				alreadyInserted = true
				break
			}
		}

		if !alreadyInserted {
			cacheKeys = append(cacheKeys, key.(string))
		}

		_ = s.Set(ctx, tagKey, []byte(strings.Join(cacheKeys, ",")), WithExpiration(720*time.Hour))
	}
}

// Delete removes data from Bigcache for given key identifier
func (s *BigcacheStore) Delete(_ context.Context, key any) error {
	return s.Client.Delete(key.(string))
}

// Invalidate invalidates some cache data in Bigcache for given options
func (s *BigcacheStore) Invalidate(ctx context.Context, options ...InvalidateOption) error {
	opts := applyInvalidateOptions(options...)

	if tags := opts.Tags; len(tags) > 0 {
		for _, tag := range tags {
			tagKey := fmt.Sprintf(BigcacheTagPattern, tag)
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
func (s *BigcacheStore) Clear(_ context.Context) error {
	return s.Client.Reset()
}

// GetType returns the store type
func (s *BigcacheStore) GetType() string {
	return BigcacheType
}
