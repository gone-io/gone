package gin

import (
	"github.com/gone-io/gone"
)

var load = gone.OnceLoad(func(loader gone.Loader) error {
	if err := loader.Load(&proxy{}); err != nil {
		return gone.ToError(err)
	}

	if err := loader.Load(
		&router{},
		gone.IsDefault(
			new(gone.RouteGroup),
			new(gone.IRouter),
		),
	); err != nil {
		return gone.ToError(err)
	}

	if err := loader.Load(&SysMiddleware{}); err != nil {
		return gone.ToError(err)
	}
	if err := loader.Load(NewGinResponser()); err != nil {
		return gone.ToError(err)
	}
	if err := loader.Load(NewGinServer()); err != nil {
		return gone.ToError(err)
	}

	if err := loader.Load(&httpInjector{}); err != nil {
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
