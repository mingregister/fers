package storage

type Client interface {
	// List all object keys (relative paths) under given prefix (empty => list all)
	List(prefix string) ([]string, error)
	// Upload object with given key and content
	Upload(key string, data []byte) error
	// Download object by key
	Download(key string) ([]byte, error)
	// Delete removes the value for a key.
	// Returns nil if successful or key doesn't exist.
	Delete(key string) error
}
