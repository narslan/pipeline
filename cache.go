package dataflow

import "context"

// A cache is used as an optimization tool. It is used for easier data access.
// If the cache has a value that we are looking up,
// we don't need ask other services to give that.
// We save IDs of the Products in a cache, while storing the Products in a database.
// When we want to update the database at another time,
// we ask the cache first about which IDs it has. The Products of IDs in the cache
// will not be updated. Otherwise we go and update the Product and cache with fresh data.

// Cache represents the cache methods for ProductIDs.
// A Cache interface might be implemented concretely using Redis, memcached or in-memory.

type Cache interface {
	Set(ctx context.Context, id uint32) error
	Exists(ctx context.Context, id uint32) (bool, error)
}
