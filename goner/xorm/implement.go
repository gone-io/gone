package xorm

import (
	"github.com/gone-io/gone"
	"time"
	"xorm.io/xorm"
)

func NewXormEngine() (gone.Angel, gone.GonerId, gone.GonerOption, gone.GonerOption) {
	return &engine{
		newFunc: newEngine,
	}, gone.IdGoneXorm, gone.IsDefault(new(gone.XormEngine)), gone.Order3
}

func newEngine(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
	return xorm.NewEngine(driverName, dataSourceName)
}

type ClusterNodeConf struct {
	DriverName string `properties:"driver-name" mapstructure:"driver-name"`
	DSN        string `properties:"dsn" mapstructure:"dsn"`
}

//go:generate mockgen -package xorm -destination=./engine_mock_test.go xorm.io/xorm EngineInterface
type engine struct {
	gone.Flag
	xorm.EngineInterface
	gone.Logger `gone:"gone-logger"`

	driverName    string             `gone:"config,database.driver-name"`
	dsn           string             `gone:"config,database.dsn"`
	maxIdleCount  int                `gone:"config,database.max-idle-count"`
	maxOpen       int                `gone:"config,database.max-open"`
	maxLifetime   time.Duration      `gone:"config,database.max-lifetime"`
	showSql       bool               `gone:"config,database.showSql,default=false"`
	enableCluster bool               `gone:"config,database.cluster.enable,default=false"`
	masterConf    *ClusterNodeConf   `gone:"config,database.cluster.master"`
	slavesConf    []*ClusterNodeConf `gone:"config,database.cluster.slaves"`

	group *xorm.EngineGroup

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

	if e.enableCluster {
		if e.masterConf == nil {
			return gone.NewInnerError("master config(database.cluster.master) is nil", gone.StartError)
		}

		if len(e.slavesConf) == 0 {
			return gone.NewInnerError("slaves config(database.cluster.slaves) is nil", gone.StartError)
		}
		master, err := e.newFunc(e.masterConf.DriverName, e.masterConf.DSN)
		if err != nil {
			return gone.NewInnerError(err.Error(), gone.StartError)
		}

		slaves := make([]*xorm.Engine, 0, len(e.slavesConf))
		for _, slave := range e.slavesConf {
			slaveEngine, err := e.newFunc(slave.DriverName, slave.DSN)
			if err != nil {
				return gone.NewInnerError(err.Error(), gone.StartError)
			}
			slaves = append(slaves, slaveEngine.(*xorm.Engine))
		}

		e.group, err = xorm.NewEngineGroup(master, slaves)
		if err != nil {
			return gone.NewInnerError(err.Error(), gone.StartError)
		}
		e.EngineInterface = e.group
	} else {
		var err error
		e.EngineInterface, err = e.newFunc(e.driverName, e.dsn)
		if err != nil {
			return gone.NewInnerError(err.Error(), gone.StartError)
		}
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
	if e.group != nil {
		return e.group.Close()
	} else {
		return e.EngineInterface.(*xorm.Engine).Close()
	}
}

func (e *engine) Sqlx(sql string, args ...any) *xorm.Session {
	sql, args = sqlDeal(sql, args...)
	return e.SQL(sql, args...)
}
