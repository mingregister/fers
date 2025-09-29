package storage

import "context"

type Client interface {
	// List all object keys (relative paths) under given prefix (empty => list all)
	List(ctx context.Context, prefix string) ([]string, error)
	// Upload object with given key and content
	Upload(ctx context.Context, key string, data []byte) error
	// Download object by key
	Download(ctx context.Context, key string) ([]byte, error)
	// Delete removes the value for a key.
	// Returns nil if successful or key doesn't exist.
	Delete(ctx context.Context, key string) error
}
