package xorm

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/logrus"
	"reflect"
	"time"
	"xorm.io/xorm"
)

func NewXormEngine() (gone.Angel, gone.GonerId) {
	return &engine{}, gone.IdGoneXorm
}

type engine struct {
	gone.Flag
	*xorm.Engine
	logrus.Logger `gone:"gone-logger"`

	driverName   string        `gone:"config,database.driver-name"`
	dsn          string        `gone:"config,database.dsn"`
	maxIdleCount int           `gone:"config,database.max-idle-count"`
	maxOpen      int           `gone:"config,database.max-open"`
	maxLifetime  time.Duration `gone:"config,database.max-lifetime"`
	showSql      bool          `gone:"config,database.showSql,default=false"`
}

func (e *engine) GetOriginEngine() *xorm.Engine {
	return e.Engine
}

func (e *engine) Start(gone.Cemetery) error {
	if e.Engine != nil {
		panic("duplicate call Start()")
	}

	var err error
	e.Engine, err = xorm.NewEngine(e.driverName, e.dsn)
	if err != nil {
		panic(err)
	}

	e.SetConnMaxLifetime(e.maxLifetime)
	e.SetMaxOpenConns(e.maxOpen)
	e.SetMaxIdleConns(e.maxIdleCount)
	e.SetLogger(&dbLogger{Logger: e.Logger, showSql: e.showSql})
	return e.Ping()
}

func (e *engine) Stop(gone.Cemetery) error {
	return e.Close()
}

type NameMap map[string]any

var NameMapType = reflect.TypeOf(&NameMap{}).Elem()

func (e *engine) Sqlx(sql string, args ...any) *xorm.Session {
	if len(args) == 1 && reflect.TypeOf(args[0]) == NameMapType {
		sql, args = MustNamed(sql, args[0])
	} else {
		sql, args = MustIn(sql, args...)
	}
	return e.SQL(sql, args...)
}
