package mysql

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

// Priest is the entry point of the gorm mysql module
func Priest(cemetery gone.Cemetery) error {
	cemetery.Bury(
		&dial{},
		gone.IsDefault(new(gorm.Dialector)),
	)
	return nil
}
