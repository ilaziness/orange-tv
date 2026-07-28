package cache

import (
	"context"
	"time"
)

// NopCache is a no-op cache implementation that does nothing.
type NopCache struct{}

// NewNopCache creates a new no-op cache.
func NewNopCache() *NopCache {
	return &NopCache{}
}

// Get implements Cache.
func (n *NopCache) Get(ctx context.Context, key string) (any, error) {
	return nil, ErrCacheMiss
}

// Set implements Cache.
func (n *NopCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return nil
}

// Delete implements Cache.
func (n *NopCache) Delete(ctx context.Context, key string) error {
	return nil
}

// Exists implements Cache.
func (n *NopCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

// Clear implements Cache.
func (n *NopCache) Clear(ctx context.Context) error {
	return nil
}

// Close implements Cache.
func (n *NopCache) Close() error {
	return nil
}
