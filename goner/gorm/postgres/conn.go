package postgres

import (
	"github.com/gone-io/gone"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type dial struct {
	gone.Flag
	gorm.Dialector

	driverName           string `gone:"config,gorm.postgres.driver-name"`
	dsn                  string `gone:"config,gorm.postgres.dsn"`
	withoutQuotingCheck  bool   `gone:"config,gorm.postgres.without-quoting-check,default=false"`
	preferSimpleProtocol bool   `gone:"config,gorm.postgres.prefer-simple-protocol,default=false"`
	withoutReturning     bool   `gone:"config,gorm.postgres.without-returning=default"`
}

func (d *dial) Apply(*gorm.Config) error {
	if d.Dialector == nil {
		d.Dialector = postgres.New(postgres.Config{
			DriverName:           d.driverName,
			DSN:                  d.dsn,
			WithoutReturning:     d.withoutReturning,
			PreferSimpleProtocol: d.preferSimpleProtocol,
			WithoutQuotingCheck:  d.withoutQuotingCheck,
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
