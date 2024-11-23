package gorm

import (
	"github.com/gone-io/gone"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"reflect"
	"time"
)

func NewGorm() (gone.Goner, gone.GonerOption, gone.GonerOption) {
	return &iGorm{}, gone.GonerId("gorm"), gone.Provide(&gorm.DB{})
}

type iGorm struct {
	gone.Flag
	db     *gorm.DB
	dial   gorm.Dialector   `gone:"*"`
	logger logger.Interface `gone:"*"`

	// GORM perform single create, update, delete operations in transactions by default to ensure database data integrity
	// You can disable it by setting `SkipDefaultTransaction` to true
	SkipDefaultTransaction bool `gone:"config,gorm.skip-default-transaction"`

	// FullSaveAssociations full save associations
	FullSaveAssociations bool `gone:"config,gorm.full-save-associations"`

	// DryRun generate sql without execute
	DryRun bool `gone:"config,dry-run"`

	// PrepareStmt executes the given query in cached statement
	PrepareStmt bool `gone:"config,gorm.prepare-stmt"`

	// DisableAutomaticPing
	DisableAutomaticPing bool `gone:"config,gorm.disable-automatic-ping"`

	// DisableForeignKeyConstraintWhenMigrating
	DisableForeignKeyConstraintWhenMigrating bool `gone:"config,gorm.disable-foreign-key-constraint-when-migrating"`

	// IgnoreRelationshipsWhenMigrating
	IgnoreRelationshipsWhenMigrating bool `gone:"config,gorm.ignore-relationships-when-migrating"`

	// DisableNestedTransaction disable nested transaction
	DisableNestedTransaction bool `gone:"config,gorm.disable-nested-transaction"`

	// AllowGlobalUpdate allow global update
	AllowGlobalUpdate bool `gone:"config,gorm.allow-global-update"`

	// QueryFields executes the SQL query with all fields of the table
	QueryFields bool `gone:"config,gorm.query-fields"`

	// CreateBatchSize default create batch size
	CreateBatchSize int `gone:"config,gorm.create-batch-size"`

	// TranslateError enabling error translation
	TranslateError bool `gone:"config,gorm.translate-error"`

	// PropagateUnscoped propagate Unscoped to every other nested statement
	PropagateUnscoped bool `gone:"config,gorm.propagate-unscoped"`

	MaxIdle         int            `gone:"config,gorm.pool.max-idle"`
	MaxOpen         int            `gone:"config,gorm.pool.max-open"`
	ConnMaxLifetime *time.Duration `gone:"config,gorm.pool.conn-max-lifetime"`
}

func (s *iGorm) Start(gone.Cemetery) (err error) {
	s.db, err = gorm.Open(s.dial, &gorm.Config{
		SkipDefaultTransaction:                   s.SkipDefaultTransaction,
		FullSaveAssociations:                     s.FullSaveAssociations,
		Logger:                                   s.logger,
		DryRun:                                   s.DryRun,
		PrepareStmt:                              s.PrepareStmt,
		DisableAutomaticPing:                     s.DisableAutomaticPing,
		DisableForeignKeyConstraintWhenMigrating: s.DisableForeignKeyConstraintWhenMigrating,
		IgnoreRelationshipsWhenMigrating:         s.IgnoreRelationshipsWhenMigrating,
		DisableNestedTransaction:                 s.DisableNestedTransaction,
		AllowGlobalUpdate:                        s.AllowGlobalUpdate,
		QueryFields:                              s.QueryFields,
		CreateBatchSize:                          s.CreateBatchSize,
		TranslateError:                           s.TranslateError,
		PropagateUnscoped:                        s.PropagateUnscoped,
	})
	if err != nil {
		return gone.ToError(err)
	}
	db, err := s.db.DB()
	if err != nil {
		return gone.ToError(err)
	}

	if s.MaxIdle > 0 {
		db.SetMaxIdleConns(s.MaxIdle)
	}
	if s.MaxOpen > 0 {
		db.SetMaxOpenConns(s.MaxOpen)
	}

	if s.ConnMaxLifetime != nil {
		db.SetConnMaxLifetime(*s.ConnMaxLifetime)
	}

	return gone.ToError(db.Ping())
}

func (s *iGorm) Stop(gone.Cemetery) error {
	db, err := s.db.DB()
	if err != nil {
		return gone.ToError(err)
	}
	return gone.ToError(db.Close())
}

var dbType = reflect.TypeOf(new(gorm.DB))

func (s *iGorm) Suck(conf string, v reflect.Value) gone.SuckError {
	if v.Type() == dbType {
		v.Set(reflect.ValueOf(s.db))
		return nil
	} else {
		return gone.NewInnerError("only support *gorm.DB", gone.InjectError)
	}
}
