package redis

import (
	"context"
	"time"

	"github.com/nullexp/sanaz.crm.api/pkg/infrastructure/cache/protocol"
	cache "github.com/patrickmn/go-cache"
)

type MemoryClient struct {
	client *cache.Cache

	defaultExpiration, cleanupInterval time.Duration
}

func NewRedisClient(defaultExpiration, cleanupInterval time.Duration) protocol.RawCacheStore {
	return &MemoryClient{defaultExpiration: defaultExpiration, cleanupInterval: cleanupInterval}
}

func (rc *MemoryClient) Connect() error {
	rc.client = cache.New(rc.defaultExpiration, rc.cleanupInterval)

	return nil
}

func (rc *MemoryClient) Disconnect() error {
	rc.client.Flush()

	return nil
}

func (rc *MemoryClient) Set(ctx context.Context, key string, value []byte) error {
	rc.client.Set(key, value, 0)

	return nil
}

func (rc *MemoryClient) Fetch(ctx context.Context, key string) ([]byte, error) {
	value, exist := rc.client.Get(key)

	if !exist {
		return nil, protocol.ErrCacheMissed
	}

	return value.([]byte), nil
}

func (rc *MemoryClient) Delete(ctx context.Context, key string) error {
	rc.client.Delete(key)

	return nil
}
