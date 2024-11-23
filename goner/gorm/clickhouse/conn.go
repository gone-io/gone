package postgres

import (
	"github.com/gone-io/gone"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type dial struct {
	gone.Flag
	gorm.Dialector

	driverName                   string `gone:"config,gorm.clickhouse.driver-name"`
	dsn                          string `gone:"config,gorm.clickhouse.dsn"`
	disableDatetimePrecision     bool   `gone:"config,gorm.clickhouse.disable-datetime-precision,default=false"`
	dontSupportRenameColumn      bool   `gone:"config,gorm.clickhouse.dont-support-rename-column,default=false"`
	dontSupportColumnPrecision   bool   `gone:"config,gorm.clickhouse.dont-support-column-precision,default=false"`
	dontSupportEmptyDefaultValue bool   `gone:"config,gorm.clickhouse.dont-support-empty-default-value,default=false"`
	skipInitializeWithVersion    bool   `gone:"config,gorm.clickhouse.skip-initialize-with-version,default=false"`
	defaultGranularity           int    `gone:"config,gorm.clickhouse.default-granularity,default="`
	defaultCompression           string `gone:"config,gorm.clickhouse.default-compression,default="`
	defaultIndexType             string `gone:"config,gorm.clickhouse.default-indexType,default="`
	defaultTableEngineOpts       string `gone:"config,gorm.clickhouse.default-table-engine-opts,default="`
}

func (d *dial) Apply(*gorm.Config) error {
	if d.Dialector == nil {
		d.Dialector = clickhouse.New(clickhouse.Config{
			DriverName:                   d.driverName,
			DSN:                          d.dsn,
			DisableDatetimePrecision:     d.disableDatetimePrecision,
			DontSupportRenameColumn:      d.dontSupportRenameColumn,
			DontSupportColumnPrecision:   d.dontSupportColumnPrecision,
			DontSupportEmptyDefaultValue: d.dontSupportEmptyDefaultValue,
			SkipInitializeWithVersion:    d.skipInitializeWithVersion,
			DefaultGranularity:           d.defaultGranularity,
			DefaultCompression:           d.defaultCompression,
			DefaultIndexType:             d.defaultIndexType,
			DefaultTableEngineOpts:       d.defaultTableEngineOpts,
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
