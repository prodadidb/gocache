package store_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/prodadidb/gocache/store"
)

func BenchmarkMemcacheSet(b *testing.B) {
	ctx := context.Background()

	s := store.NewMemcache(
		memcache.New("127.0.0.1:11211"),
		store.WithExpiration(100*time.Second),
	)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				key := fmt.Sprintf("test-%d", n)
				value := []byte(fmt.Sprintf("value-%d", n))

				_ = s.Set(ctx, key, value, store.WithTags([]string{fmt.Sprintf("tag-%d", n)}))
			}
		})
	}
}

func BenchmarkMemcacheGet(b *testing.B) {
	ctx := context.Background()

	s := store.NewMemcache(
		memcache.New("127.0.0.1:11211"),
		store.WithExpiration(100*time.Second),
	)

	key := "test"
	value := []byte("value")

	_ = s.Set(ctx, key, value, nil)

	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))
		b.Run(fmt.Sprintf("%d", n), func(b *testing.B) {
			for i := 0; i < b.N*n; i++ {
				_, _ = s.Get(ctx, key)
			}
		})
	}
}
