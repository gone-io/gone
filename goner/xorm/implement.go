package xorm

import (
	"github.com/gone-io/gone"
	"time"
	"xorm.io/xorm"
)

func NewXormEngine() (gone.Angel, gone.GonerId, gone.GonerOption) {
	return &engine{
		newFunc: newEngine,
	}, gone.IdGoneXorm, gone.IsDefault(true)
}

func newEngine(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
	return xorm.NewEngine(driverName, dataSourceName)
}

//go:generate mockgen -package xorm  xorm.io/xorm EngineInterface > ./engine_mock_test.go
type engine struct {
	gone.Flag
	xorm.EngineInterface
	gone.Logger `gone:"gone-logger"`

	driverName   string        `gone:"config,database.driver-name"`
	dsn          string        `gone:"config,database.dsn"`
	maxIdleCount int           `gone:"config,database.max-idle-count"`
	maxOpen      int           `gone:"config,database.max-open"`
	maxLifetime  time.Duration `gone:"config,database.max-lifetime"`
	showSql      bool          `gone:"config,database.showSql,default=false"`

	newFunc func(driverName string, dataSourceName string) (xorm.EngineInterface, error)
}

func (e *engine) GetOriginEngine() xorm.EngineInterface {
	return e.EngineInterface
}

func (e *engine) Start(gone.Cemetery) error {
	err := e.create()
	if err != nil {
		return err
	}
	e.config()
	return e.Ping()
}
func (e *engine) create() error {
	if e.EngineInterface != nil {
		return gone.NewInnerError("duplicate call Start()", gone.StartError)
	}

	var err error
	e.EngineInterface, err = e.newFunc(e.driverName, e.dsn)
	if err != nil {
		return gone.NewInnerError(err.Error(), gone.StartError)
	}
	return nil
}

func (e *engine) config() {
	e.SetConnMaxLifetime(e.maxLifetime)
	e.SetMaxOpenConns(e.maxOpen)
	e.SetMaxIdleConns(e.maxIdleCount)
	e.SetLogger(&dbLogger{Logger: e.Logger, showSql: e.showSql})
}

func (e *engine) Stop(gone.Cemetery) error {
	return e.EngineInterface.(*xorm.Engine).Close()
}

func (e *engine) Sqlx(sql string, args ...any) *xorm.Session {
	sql, args = sqlDeal(sql, args...)
	return e.SQL(sql, args...)
}
