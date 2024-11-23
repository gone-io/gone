package sqlserver

import (
	"github.com/gone-io/gone"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type dial struct {
	gone.Flag
	gorm.Dialector

	DriverName        string `gone:"config,gorm.sqlserver.driver-name"`
	DSN               string `gone:"config,gorm.sqlserver.dsn"`
	DefaultStringSize int    `gone:"config,gorm.sqlserver.default-string-size"`
}

func (d *dial) Apply(*gorm.Config) error {
	if d.Dialector == nil {
		d.Dialector = sqlserver.New(sqlserver.Config{
			DriverName:        d.DriverName,
			DSN:               d.DSN,
			DefaultStringSize: d.DefaultStringSize,
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
