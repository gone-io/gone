package xorm

import (
	"github.com/gone-io/gone"
	"io"
	"time"
	"xorm.io/xorm"
)

func NewXormEngine() (gone.Angel, gone.GonerId, gone.GonerOption, gone.GonerOption) {
	return newWrappedEngine(), gone.IdGoneXorm, gone.IsDefault(new(gone.XormEngine)), gone.Order3
}

func newWrappedEngine() *wrappedEngine {
	return &wrappedEngine{
		newFunc:    newEngine,
		newSession: newSession,
	}
}

func newEngine(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
	return xorm.NewEngine(driverName, dataSourceName)
}

func newSession(eng xorm.EngineInterface) XInterface {
	return eng.NewSession()
}

type ClusterNodeConf struct {
	DriverName string `properties:"driver-name" mapstructure:"driver-name"`
	DSN        string `properties:"dsn" mapstructure:"dsn"`
}

type Conf struct {
	DriverName    string             `properties:"driver-name" mapstructure:"driver-name"`
	Dsn           string             `properties:"dsn" mapstructure:"dsn"`
	MaxIdleCount  int                `properties:"max-idle-count" mapstructure:"max-idle-count"`
	MaxOpen       int                `properties:"max-open" mapstructure:"max-open"`
	MaxLifetime   time.Duration      `properties:"max-lifetime" mapstructure:"max-lifetime"`
	ShowSql       bool               `properties:"show-sql" mapstructure:"show-sql"`
	EnableCluster bool               `properties:"cluster.enable" mapstructure:"cluster.enable"`
	MasterConf    *ClusterNodeConf   `properties:"cluster.master" mapstructure:"cluster.master"`
	SlavesConf    []*ClusterNodeConf `properties:"cluster.slaves" mapstructure:"cluster.slaves"`
}

//go:generate mockgen -package xorm -destination=./engine_mock_test.go xorm.io/xorm EngineInterface
type wrappedEngine struct {
	gone.Flag
	xorm.EngineInterface

	newFunc    func(driverName string, dataSourceName string) (xorm.EngineInterface, error)
	newSession func(xorm.EngineInterface) XInterface

	log  gone.Logger `gone:"gone-logger"`
	conf *Conf       `gone:"config,database"`
}

func (e *wrappedEngine) GetOriginEngine() xorm.EngineInterface {
	return e.EngineInterface
}

func (e *wrappedEngine) Start(gone.Cemetery) error {
	err := e.create()
	if err != nil {
		return err
	}
	e.config()
	return e.Ping()
}
func (e *wrappedEngine) create() error {
	if e.EngineInterface != nil {
		return gone.NewInnerError("duplicate call Start()", gone.StartError)
	}

	if e.conf.EnableCluster {
		if e.conf.MasterConf == nil {
			return gone.NewInnerError("master config(database.cluster.master) is nil", gone.StartError)
		}

		if len(e.conf.SlavesConf) == 0 {
			return gone.NewInnerError("slaves config(database.cluster.slaves) is nil", gone.StartError)
		}
		master, err := e.newFunc(e.conf.MasterConf.DriverName, e.conf.MasterConf.DSN)
		if err != nil {
			return gone.NewInnerError(err.Error(), gone.StartError)
		}

		slaves := make([]*xorm.Engine, 0, len(e.conf.SlavesConf))
		for _, slave := range e.conf.SlavesConf {
			slaveEngine, err := e.newFunc(slave.DriverName, slave.DSN)
			if err != nil {
				return gone.NewInnerError(err.Error(), gone.StartError)
			}
			slaves = append(slaves, slaveEngine.(*xorm.Engine))
		}

		e.EngineInterface, err = xorm.NewEngineGroup(master, slaves)
		if err != nil {
			return gone.NewInnerError(err.Error(), gone.StartError)
		}
	} else {
		var err error
		e.EngineInterface, err = e.newFunc(e.conf.DriverName, e.conf.Dsn)
		if err != nil {
			return gone.NewInnerError(err.Error(), gone.StartError)
		}
	}
	return nil
}

func (e *wrappedEngine) config() {
	e.SetConnMaxLifetime(e.conf.MaxLifetime)
	e.SetMaxOpenConns(e.conf.MaxOpen)
	e.SetMaxIdleConns(e.conf.MaxIdleCount)
	e.SetLogger(&dbLogger{Logger: e.log, showSql: e.conf.ShowSql})
}

func (e *wrappedEngine) Stop(gone.Cemetery) error {
	return e.EngineInterface.(io.Closer).Close()
}

func (e *wrappedEngine) Sqlx(sql string, args ...any) *xorm.Session {
	sql, args = sqlDeal(sql, args...)
	return e.SQL(sql, args...)
}
