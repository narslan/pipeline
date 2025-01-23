package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Cache represents a connection to the Redis database.
type Cache struct {
	*redis.Client
}

// NewCache returns a new instance of Cache associated with the given connection params.
// NewCache("localhost:6379", "", 0)

func NewCache(addr, pass string, db int) (*Cache, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       db,
	})

	err := rdb.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return &Cache{Client: rdb}, nil

}

// ShutDown closes db connection. Use ShutDown instead of Close
// because embedded struct redis.Client already have Close method.
// We have to choose another name due to conflict.
func (db *Cache) ShutDown() error {
	return db.Close()
}
