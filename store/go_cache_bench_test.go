package store_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/prodadidb/gocache/store"
)

func BenchmarkGoCacheSet(b *testing.B) {
	ctx := context.Background()

	client := cache.New(10*time.Second, 30*time.Second)

	s := store.NewGoCache(client, nil)

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

func BenchmarkGoCacheGet(b *testing.B) {
	ctx := context.Background()

	client := cache.New(10*time.Second, 30*time.Second)

	s := store.NewGoCache(client, nil)

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
