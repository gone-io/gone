package redis

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/tracer"
)

var load = gone.OnceLoad(func(loader gone.Loader) error {
	if err := tracer.Load(loader); err != nil {
		return gone.ToError(err)
	}

	if err := loader.Load(&pool{}, gone.IsDefault(new(Pool))); err != nil {
		return gone.ToError(err)
	}
	if err := loader.Load(&inner{}); err != nil {
		return gone.ToError(err)
	}
	if err := loader.Load(
		&cache{},
		gone.IsDefault(new(Cache), new(Key)),
	); err != nil {
		return gone.ToError(err)
	}
	if err := loader.Load(&locker{}, gone.IsDefault(new(Locker))); err != nil {
		return gone.ToError(err)
	}
	if err := loader.Load(&provider{}, gone.IsDefault(new(HashProvider))); err != nil {
		return gone.ToError(err)
	}
	return nil
})

func Load(loader gone.Loader) error {
	return load(loader)
}

// Priest Deprecated, use Load instead
func Priest(loader gone.Loader) error {
	return Load(loader)
}
