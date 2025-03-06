package pipeline_test

import (
	"context"
	"testing"

	"github.com/narslan/pipeline/cassandra"
	"github.com/narslan/pipeline/redis"
)

// MustOpenDB returns a new DB. Fatal on error.
func MustOpenDB(tb testing.TB, connectionHost string) *cassandra.DB {
	tb.Helper()

	db, err := cassandra.NewDB(connectionHost, "case_pipeline_test", "cassandra", "")
	if err != nil {
		tb.Fatal(err)
	}
	return db
}

// MustCloseDB closes the DB.
func MustCloseDB(tb testing.TB, db *cassandra.DB) {
	tb.Helper()
	db.Close()
}

// MustOpenCache returns a new Cache. Fatal on error.
func MustOpenCache(tb testing.TB, redisConnection string) *redis.Cache {
	tb.Helper()

	db, err := redis.NewCache(redisConnection, "", 0)
	if err != nil {
		tb.Fatal(err)
	}

	//Ensure that we always have a fresh database .
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
