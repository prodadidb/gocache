package store_test

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/prodadidb/gocache/store"
)

// should be configured to connect to real Redis Cluster
func BenchmarkRedisClusterSet(b *testing.B) {
	ctx := context.Background()

	addr := strings.Split("redis:6379", ",")
	s := store.NewRedisCluster(redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addr,
	}), nil)

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

func BenchmarkRedisClusterGet(b *testing.B) {
	ctx := context.Background()

	addr := strings.Split("redis:6379", ",")
	s := store.NewRedisCluster(redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: addr,
	}), nil)

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
