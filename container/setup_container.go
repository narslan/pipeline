package container

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"log"

	"github.com/testcontainers/testcontainers-go"
	cassandracp "github.com/testcontainers/testcontainers-go/modules/cassandra" // Rename for avoiding name clash.
	rediscp "github.com/testcontainers/testcontainers-go/modules/redis"         // Rename for avoiding name clash.
	"github.com/testcontainers/testcontainers-go/wait"
)

// Deploy cassandra and redis containers for test.
func MustDeployCassandra(ctx context.Context) (*cassandracp.CassandraContainer, string) {

	// Setup cassandra image.
	cassandraContainer, err := cassandracp.RunContainer(ctx,
		testcontainers.WithImage("cassandra:4.1.3"),

		//This requires a testdata folder and an init.cql under it.
		cassandracp.WithInitScripts(filepath.Join("testdata", "init.cql")),

		// We wait utill database can respond.
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("9042/tcp"),
			wait.ForExec([]string{"cqlsh", "-e", "SELECT release_version FROM system.local"}).
				WithResponseMatcher(func(body io.Reader) bool {
					data, _ := io.ReadAll(body)
					return strings.Contains(string(data), "4.1.3")
				}),
		),
	)

	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	// Obtain connection port from container.
	cassandraConnectionHost, err := cassandraContainer.ConnectionHost(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return cassandraContainer, cassandraConnectionHost

}

// Deploy cassandra and redis containers for test.
func MustDeployRedis(ctx context.Context) (*rediscp.RedisContainer, string) {

	redisContainer, err := rediscp.RunContainer(ctx,
		testcontainers.WithImage("redis:alpine"),
		rediscp.WithLogLevel(rediscp.LogLevelVerbose),
	)

	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	redisConnectionString, err := redisContainer.ConnectionString(ctx)

	if err != nil {
		log.Fatal(err)
	}

	redisConnection := strings.ReplaceAll(redisConnectionString, "redis://", "")
	return redisContainer, redisConnection

}

// Clean up containers.
func MustCleanCassandraContainer(ctx context.Context, cdbC *cassandracp.CassandraContainer) {

	err := cdbC.Terminate(ctx)
	if err != nil {
		log.Fatalf("failed to terminate container: %s", err)
	}

}

func MustCleanRedisContainer(ctx context.Context, rC *rediscp.RedisContainer) {

	err := rC.Terminate(ctx)
	if err != nil {
		log.Fatalf("failed to terminate container: %s", err)
	}

}
