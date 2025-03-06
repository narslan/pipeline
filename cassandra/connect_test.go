package cassandra_test

import (
	"context"
	"testing"

	"github.com/narslan/pipeline/cassandra"
	"github.com/narslan/pipeline/container"
)

// Ensure the test database can open & close.
func TestDB(t *testing.T) {
	// Start containers for test.
	ctx := context.Background()
	cdbc, cassandraConnectionHost := container.MustDeployCassandra(ctx)
	defer container.MustCleanCassandraContainer(ctx, cdbc)

	db := MustOpenDB(t, cassandraConnectionHost)
	defer MustCloseDB(t, db)
}

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
