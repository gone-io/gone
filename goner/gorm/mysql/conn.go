package mysql

import (
	"github.com/gone-io/gone"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type dial struct {
	gone.Flag
	gorm.Dialector

	DriverName                    string `gone:"config,gorm.mysql.driver-name"`
	DSN                           string `gone:"config,gorm.mysql.dsn"`
	ServerVersion                 string `gone:"config,gorm.mysql.server-version"`
	SkipInitializeWithVersion     bool   `gone:"config,gorm.mysql.skip-initialize-with-version"`
	DefaultStringSize             uint   `gone:"config,gorm.mysql.default-string-size"`
	DefaultDatetimePrecision      *int   `gone:"config,gorm.mysql.default-datetime-precision"`
	DisableWithReturning          bool   `gone:"config,gorm.mysql.disable-with-returning"`
	DisableDatetimePrecision      bool   `gone:"config,gorm.mysql.disable-datetime-precision"`
	DontSupportRenameIndex        bool   `gone:"config,gorm.mysql.dont-support-rename-index"`
	DontSupportRenameColumn       bool   `gone:"config,gorm.mysql.dont-support-rename-column"`
	DontSupportForShareClause     bool   `gone:"config,gorm.mysql.dont-support-for-share-clause"`
	DontSupportNullAsDefaultValue bool   `gone:"config,gorm.mysql.dont-support-null-as-default-value"`
	DontSupportRenameColumnUnique bool   `gone:"config,gorm.mysql.dont-support-rename-column-unique"`
	// As of MySQL 8.0.19, ALTER TABLE permits more general (and SQL standard) syntax
	// for dropping and altering existing constraints of any type.
	// see https://dev.mysql.com/doc/refman/8.0/en/alter-table.html
	DontSupportDropConstraint bool `gone:"config,gorm.mysql.dont-support-drop-constraint"`
}

func (d *dial) Init() error {
	if d.Dialector == nil {
		d.Dialector = mysql.New(mysql.Config{
			DriverName:                    d.DriverName,
			ServerVersion:                 d.ServerVersion,
			DSN:                           d.DSN,
			SkipInitializeWithVersion:     d.SkipInitializeWithVersion,
			DefaultStringSize:             d.DefaultStringSize,
			DefaultDatetimePrecision:      d.DefaultDatetimePrecision,
			DisableWithReturning:          d.DisableWithReturning,
			DisableDatetimePrecision:      d.DisableDatetimePrecision,
			DontSupportRenameIndex:        d.DontSupportRenameIndex,
			DontSupportRenameColumn:       d.DontSupportRenameColumn,
			DontSupportForShareClause:     d.DontSupportForShareClause,
			DontSupportNullAsDefaultValue: d.DontSupportNullAsDefaultValue,
			DontSupportRenameColumnUnique: d.DontSupportRenameColumnUnique,
			DontSupportDropConstraint:     d.DontSupportDropConstraint,
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
