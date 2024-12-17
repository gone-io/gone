package xorm

import (
	"github.com/gone-io/gone"
)

var load = gone.OnceLoad(func(loader gone.Loader) error {
	engine := newWrappedEngine()
	if err := loader.Load(
		engine,
		gone.IsDefault(new(gone.XormEngine)),
		gone.HighStartPriority(),
	); err != nil {
		return gone.ToError(err)
	}

	if err := loader.Load(newProvider(engine)); err != nil {
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
