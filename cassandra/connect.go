package cassandra

import (
	"time"

	"github.com/gocql/gocql"
)

// DB represents the database connection.
type DB struct {
	session *gocql.Session
}

// NewDB returns a new instance of DB associated with the given connection parameters.
func NewDB(host, keyspace, user, pass string) (*DB, error) {

	// Instantiate a cluster object.
	cluster := gocql.NewCluster(host)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: user,
		Password: pass,
	}

	// Cluster options.
	cluster.Keyspace = keyspace

	cluster.Consistency = gocql.One
	cluster.ProtoVersion = 4
	cluster.Timeout = 2 * time.Second
	cluster.ConnectTimeout = 5 * time.Second
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &DB{session: session}, nil
}

func (db *DB) Close() {
	db.session.Close()
}
