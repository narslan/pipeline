package dataflow

import "context"

// Fetch represents methods for data retrieval from remote resources.
type Fetch interface {
	Get(ctx context.Context, url string) ([]byte, error)
}
