package gorm

import (
	"github.com/gone-io/gone"
	"gorm.io/gorm/logger"
)

var load = gone.OnceLoad(func(loader gone.Loader) error {
	if err := loader.Load(&iLogger{}, gone.IsDefault(new(logger.Interface))); err != nil {
		return gone.ToError(err)
	}

	if err := loader.Load(&dbProvider{}); err != nil {
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
