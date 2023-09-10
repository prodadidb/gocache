package store_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/prodadidb/gocache/store"
)

func BenchmarkBigcacheSet(b *testing.B) {
	ctx := context.Background()

	client, _ := bigcache.New(ctx, bigcache.DefaultConfig(5*time.Minute))
	s := store.NewBigcache(client, nil)

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

func BenchmarkBigcacheGet(b *testing.B) {
	ctx := context.Background()

	client, _ := bigcache.New(ctx, bigcache.DefaultConfig(5*time.Minute))
	s := store.NewBigcache(client, nil)

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
