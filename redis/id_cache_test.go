package redis_test

import (
	"context"
	"testing"

	"github.com/narslan/pipeline/container"
	"github.com/narslan/pipeline/redis"
)

func TestIDCacheService_Set(t *testing.T) {
	//Ensure id can be set.
	// Start containers for test.
	ctx := context.Background()
	rdbc, redisConnectionString := container.MustDeployRedis(ctx)
	defer container.MustCleanRedisContainer(ctx, rdbc)

	t.Run("OK", func(t *testing.T) {
		db := MustOpenCache(t, redisConnectionString)
		defer MustCloseCache(t, db)

		s := redis.NewIDCacheService(db)

		// set an id.
		want := uint32(42)
		err := s.Set(context.Background(), want)
		if err != nil {
			t.Fatal(err)
		}

		// Fetch product from database and compare.
		ok, err := s.Exists(context.Background(), want)
		if err != nil {
			t.Fatal(err)
		} else if !ok {
			t.Fatal("expected true, got false")
		}
	})

}

func TestIDCacheService_Exists(t *testing.T) {
	//Ensure id can be set.
	// Start containers for test.
	ctx := context.Background()
	rdbc, redisConnectionString := container.MustDeployRedis(ctx)
	defer container.MustCleanRedisContainer(ctx, rdbc)

	// Ensure false is returned if id does not exist.
	t.Run("ErrNotFound", func(t *testing.T) {
		db := MustOpenCache(t, redisConnectionString)
		defer MustCloseCache(t, db)

		s := redis.NewIDCacheService(db)

		// Fetch id from cache and compare.
		ok, err := s.Exists(context.Background(), 42)
		if err != nil {
			t.Fatal(err)
		} else if ok {
			t.Fatal("expected false, got true")
		}
	})
}
