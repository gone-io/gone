package xorm

import (
	"fmt"
	"github.com/gone-io/gone"
	"reflect"
	"strconv"
	"strings"
	"xorm.io/xorm"
)

const clusterKey = "db"
const defaultCluster = "database"

func NewProvider(engine *wrappedEngine) (gone.Vampire, gone.GonerOption) {
	var engineMap = make(map[string]*wrappedEngine)
	engineMap[""] = engine
	engineMap[defaultCluster] = engine

	return &provider{
		engineMap: engineMap,
		newFunc:   engine.newFunc,
		unitTest:  engine.unitTest,
	}, gone.GonerId("xorm")
}

type provider struct {
	gone.Flag
	engineMap map[string]*wrappedEngine

	heaven    gone.Heaven    `gone:"*"`
	cemetery  gone.Cemetery  `gone:"*"`
	configure gone.Configure `gone:"*"`
	log       gone.Logger    `gone:"*"`

	newFunc  func(driverName string, dataSourceName string) (xorm.EngineInterface, error)
	unitTest bool
}

var xormInterface = gone.GetInterfaceType(new(gone.XormEngine))
var xormInterfaceSlice = gone.GetInterfaceType(new([]gone.XormEngine))

func confMap(conf string) map[string]string {
	conf = strings.TrimSpace(conf)
	specs := strings.Split(conf, ",")
	m := make(map[string]string)
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		pairs := strings.Split(spec, "=")
		if len(pairs) == 1 {
			m[pairs[0]] = ""
		} else if len(pairs) > 1 {
			m[pairs[0]] = pairs[1]
		}
	}
	return m
}

func (p *provider) Suck(conf string, v reflect.Value) gone.SuckError {
	m := confMap(conf)
	clusterName := m[clusterKey]
	if clusterName == "" {
		clusterName = defaultCluster
	}

	db := p.engineMap[clusterName]
	if db == nil {
		var config Conf
		err := p.configure.Get(clusterName, &config, "")
		if err != nil {
			return gone.NewInnerError("failed to get config for cluster: "+clusterName, gone.InjectError)
		}

		var masterConf ClusterNodeConf
		err = p.configure.Get(clusterName+".cluster.master", &masterConf, "")
		if err != nil {
			return gone.NewInnerError("failed to get master config for cluster: "+clusterName, gone.InjectError)
		}

		var slavesConf []*ClusterNodeConf
		err = p.configure.Get(clusterName+".cluster.slaves", &slavesConf, "")
		if err != nil {
			return gone.NewInnerError("failed to get slaves config for cluster: "+clusterName, gone.InjectError)
		}

		db = newWrappedEngine()
		db.conf = config
		db.masterConf = &masterConf
		db.slavesConf = slavesConf

		//for test
		db.newFunc = p.newFunc
		db.unitTest = p.unitTest

		err = db.Start(p.cemetery)
		if err != nil {
			return gone.NewInnerError("failed to start xorm engine for cluster: "+clusterName, gone.InjectError)
		}

		p.heaven.BeforeStop(func(engine *wrappedEngine) func(cemetery gone.Cemetery) error {
			return func(cemetery gone.Cemetery) error {
				return engine.Stop(cemetery)
			}
		}(db))

		p.engineMap[clusterName] = db
	}

	if v.Type() == xormInterfaceSlice {
		if !db.conf.EnableCluster {
			return gone.NewInnerError(fmt.Sprintf("database(name=%s) is not enable cluster, cannot inject []gone.XormEngine", clusterName), gone.InjectError)
		}

		engines := db.group.Slaves()
		xormEngines := make([]gone.XormEngine, 0, len(engines))
		for _, eng := range engines {
			xormEngines = append(xormEngines, &wrappedEngine{
				EngineInterface: eng,
			})
		}
		v.Set(reflect.ValueOf(xormEngines))
		return nil
	}

	if v.Type() == xormInterface {
		if _, ok := m["master"]; ok {
			if !db.conf.EnableCluster {
				return gone.NewInnerError(fmt.Sprintf("database(name=%s) is not enable cluster, cannot inject master into gone.XormEngine", clusterName), gone.InjectError)
			}

			v.Set(reflect.ValueOf(&wrappedEngine{
				EngineInterface: db.group.Master(),
			}))
			return nil
		}

		if slaveIndex, ok := m["slave"]; ok {
			if !db.conf.EnableCluster {
				return gone.NewInnerError(fmt.Sprintf("database(name=%s) is not enable cluster, cannot inject slave into gone.XormEngine", clusterName), gone.InjectError)
			}

			slaves := db.group.Slaves()
			var index int64
			var err error
			if slaveIndex != "" {
				index, err = strconv.ParseInt(slaveIndex, 10, 64)
				if err != nil || index < 0 || index >= int64(len(slaves)) {
					return gone.NewInnerError(fmt.Sprintf("invalid slave index: %s, must be greater than or equal to 0 and less than %d ", slaveIndex, len(slaves)), gone.InjectError)
				}
			}

			v.Set(reflect.ValueOf(&wrappedEngine{
				EngineInterface: slaves[index],
			}))
			return nil
		}
		v.Set(reflect.ValueOf(db))
		return nil
	}
	return gone.CannotFoundGonerByTypeError(v.Type())
}

//database.cluster.enable=true
//database.cluster.master.driver-name=mysql
//database.cluster.master.dsn=${db.username}:${db.password}@tcp(${db.host}:${db.port})/${db.name}?charset=utf8mb4&loc=Local
//
//database.cluster.slaves[0].driver-name=mysql
//database.cluster.slaves[0].dsn=${db.username}:${db.password}@tcp(${db.host}:${db.port})/${db.name}?charset=utf8mb4&loc=Local
//
//database.cluster.slaves[1].driver-name=mysql
//database.cluster.slaves[1].dsn=${db.username}:${db.password}@tcp(${db.host}:${db.port})/${db.name}?charset=utf8mb4&loc=Local
//
//database.cluster.slaves[2].driver-name=mysql
//database.cluster.slaves[2].dsn=${db.username}:${db.password}@tcp(${db.host}:${db.port})/${db.name}?charset=utf8mb4&loc=Local

//func Test_iCommission_Tmp(t *testing.T) {
//	gone.RunTest(func(e struct {
//		group  gone.XormEngine   `gone:"*"`
//		master gone.XormEngine   `gone:"xorm,master"`
//		slave0 gone.XormEngine   `gone:"xorm,slave0"`
//		slave1 gone.XormEngine   `gone:"xorm,slave1"`
//		slave2 gone.XormEngine   `gone:"xorm,slave2"`
//		slaves []gone.XormEngine `gone:"xorm,xxx"`
//	}) {
//		assert.Equal(t, 3, len(e.slaves))
//		assert.Equal(t, e.slaves[0], e.slave0)
//		assert.Equal(t, e.slaves[1], e.slave1)
//		assert.Equal(t, e.slaves[2], e.slave2)
//		assert.NotNil(t, e.master)
//
//		err := e.master.Ping()
//		assert.Nil(t, err)
//		err = e.slave0.Ping()
//		assert.Nil(t, err)
//		err = e.slave1.Ping()
//		assert.Nil(t, err)
//		err = e.slave2.Ping()
//		assert.Nil(t, err)
//
//		err = e.group.Ping()
//		assert.Nil(t, err)
//	}, goner.XormPriest)
//}
