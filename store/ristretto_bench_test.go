package store_test

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/dgraph-io/ristretto"
	"github.com/prodadidb/gocache/store"
)

func BenchmarkRistrettoSet(b *testing.B) {
	ctx := context.Background()

	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	s := store.NewRistretto(client, nil)

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

func BenchmarkRistrettoGet(b *testing.B) {
	ctx := context.Background()

	client, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
	s := store.NewRistretto(client, nil)

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
