package sqlite

import (
	"github.com/gone-io/gone"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type dial struct {
	gone.Flag
	gorm.Dialector

	DriverName string `gone:"config,gorm.sqlite.driver-name"`
	DSN        string `gone:"config,gorm.sqlite.dsn"`
}

func (d *dial) Apply(*gorm.Config) error {
	if d.Dialector == nil {
		d.Dialector = sqlite.New(sqlite.Config{
			DriverName: d.DriverName,
			DSN:        d.DSN,
		})
	}
	return nil
}

var load = gone.OnceLoad(func(loader gone.Loader) error {
	return loader.Load(
		&dial{},
		gone.IsDefault(new(gorm.Dialector)),
	)
})

func Load(loader gone.Loader) error {
	return load(loader)
}

// Priest Deprecated, use Load instead
func Priest(loader gone.Loader) error {
	return Load(loader)
}
