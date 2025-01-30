package redis_test

import (
	"context"
	"testing"

	"github.com/narslan/dataflow/container"
	"github.com/narslan/dataflow/redis"
)

// Ensure the test database can open & close.
func TestCache(t *testing.T) {

	// Start containers for test.
	ctx := context.Background()
	rdbc, redisConnectionString := container.MustDeployRedis(ctx)
	defer container.MustCleanRedisContainer(ctx, rdbc)

	// Setup redis connection.
	db := MustOpenCache(t, redisConnectionString)
	MustCloseCache(t, db)
}

// MustOpenCache returns a new Cache. Fatal on error.
func MustOpenCache(tb testing.TB, connectionString string) *redis.Cache {
	tb.Helper()
	db, err := redis.NewCache(connectionString, "", 0)
	if err != nil {
		tb.Fatal(err)
	}

	//Ensure that we always have a fresh database. This empties the database.
	db.FlushDB(context.Background())
	return db
}

// MustCloseCache closes the Cache. Fatal on error.
func MustCloseCache(tb testing.TB, db *redis.Cache) {
	tb.Helper()
	err := db.ShutDown()
	if err != nil {
		tb.Fatal(err)
	}
}
