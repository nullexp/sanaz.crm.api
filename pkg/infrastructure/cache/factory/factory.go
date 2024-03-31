package factory

import (
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/cache/protocol"
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/cache/redis"
)

type CacheType string

const (
	Redis CacheType = "Redis"

	Test CacheType = "Test"
)

func NewCacheStore(name CacheType, params ...any) protocol.RawCacheStore {
	switch name {

	case Test:

		panic("not implemented")

	case Redis:

		return redis.NewRedisClient(params[0].(string), params[1].(string), params[2].(string), params[3].(string))

	}

	panic("not implemented")
}
