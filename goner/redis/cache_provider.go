package redis

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
	"reflect"
)

func NewCacheProvider() (gone.Vampire, gone.GonerId) {
	return &cacheProvider{}, gone.IdGoneRedisProvider
}

type cacheProvider struct {
	gone.Flag
	inner  *inner        `gone:"gone-redis-inner"`
	tracer tracer.Tracer `gone:"gone-tracer"`
}

var cacheType = gone.GetInterfaceType(new(Cache))
var hashType = gone.GetInterfaceType(new(Hash))
var keyType = gone.GetInterfaceType(new(Key))
var lockerType = gone.GetInterfaceType(new(Locker))

func (p *cacheProvider) Suck(conf string, v reflect.Value) (err gone.SuckError) {
	if conf == "" {
		return CacheProviderNeedKeyError()
	}

	var value interface{}
	switch v.Type() {
	case cacheType, keyType:
		value = &cache{inner: &inner{
			Logger:      p.inner.Logger,
			pool:        p.inner.pool,
			cachePrefix: p.inner.buildKey(conf),
		}}

	case lockerType:
		value = &locker{
			tracer: p.tracer,
			inner: &inner{
				Logger:      p.inner.Logger,
				pool:        p.inner.pool,
				cachePrefix: p.inner.buildKey(conf),
			},
		}
	case hashType:
		value = &hash{
			key:   conf,
			inner: p.inner,
		}
	default:
		return gone.CannotFoundGonerByTypeError(v.Type())
	}

	v.Set(reflect.ValueOf(value))
	return
}
