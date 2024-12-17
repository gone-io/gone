package redis

import (
	"github.com/gone-io/gone"
	"reflect"
)

type provider struct {
	gone.Flag
	inner     *inner         `gone:"gone-redis-inner"`
	tracer    gone.Tracer    `gone:"gone-tracer"`
	configure gone.Configure `gone:"*"`
}

func (s *provider) GonerName() string {
	return IdGoneRedis
}

var cacheType = gone.GetInterfaceType(new(Cache))
var hashType = gone.GetInterfaceType(new(Hash))
var keyType = gone.GetInterfaceType(new(Key))
var lockerType = gone.GetInterfaceType(new(Locker))

func (s *provider) Provide(tagConf string, t reflect.Type) (any, error) {
	m, keys := gone.TagStringParse(tagConf)
	configKey := m["config"]

	var conf string
	if configKey != "" {
		if err := s.configure.Get(configKey, &conf, ""); err != nil {
			return nil, gone.ToError(err)
		}
	} else {
		conf = keys[0]
	}

	if conf == "" {
		return nil, gone.NewInnerError(
			"redis provider need a key tag, like `gone:\"redis,{key}\"` "+
				"or `gone:\"redis,config={configKey}\"`", gone.ProviderError)
	}

	switch t {
	case cacheType, keyType:
		return &cache{inner: &inner{
			Logger:      s.inner.Logger,
			pool:        s.inner.pool,
			cachePrefix: s.inner.buildKey(conf),
		}}, nil

	case lockerType:
		return &locker{
			tracer: s.tracer,
			inner: &inner{
				Logger:      s.inner.Logger,
				pool:        s.inner.pool,
				cachePrefix: s.inner.buildKey(conf),
			},
		}, nil
	case hashType:
		return &hash{
			key:   conf,
			inner: s.inner,
		}, nil
	default:
		return nil, gone.NewInnerErrorWithParams(
			gone.GonerTypeNotMatch,
			"Cannot find matched value for %q",
			gone.GetTypeName(t),
		)
	}
}

func (s *provider) ProvideHashForKey(key string) (Hash, error) {
	provide, err := s.Provide(key, hashType)
	if err != nil {
		return nil, gone.ToError(err)
	}
	return provide.(Hash), nil
}
