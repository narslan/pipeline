package redis

import (
	"context"
	"strconv"

	"github.com/narslan/pipeline"
	"github.com/redis/go-redis/v9"
)

// IDCacheService represent a service for managing ids in redis cache.
type IDCacheService struct {
	cache *Cache
}

// Ensure service implements interface.
var _ dataflow.Cache = (*IDCacheService)(nil)

// NewIDCacheService returns a new instance of IDCacheService.
func NewIDCacheService(cache *Cache) *IDCacheService {
	return &IDCacheService{cache: cache}
}

// Set stores a given id into redis cache.
func (s *IDCacheService) Set(ctx context.Context, id uint32) error {

	// Convert id to a string.
	// FormatUInt is the fastest way to convert a uint to a string.
	uid := strconv.FormatUint(uint64(id), 10)

	// Save to redis and return error.
	return s.cache.Set(ctx, uid, "", 0).Err()
}

// Set stores a given id into redis cache.
func (s *IDCacheService) Exists(ctx context.Context, id uint32) (bool, error) {

	// Convert id to a string.
	pid := strconv.FormatUint(uint64(id), 10)

	_, err := s.cache.Get(ctx, pid).Result()

	// If product id does not exist, handle this particularly.
	if err == redis.Nil {
		return false, nil // id does not exist.
	} else if err != nil {
		return false, err // Some other error we have.
	}

	return true, nil // When we reach here, we have the key.
}
