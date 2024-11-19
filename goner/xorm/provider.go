package xorm

import (
	"fmt"
	"github.com/gone-io/gone"
	"reflect"
	"strconv"
	"strings"
)

func NewProvider(engine *engine) (gone.Vampire, gone.GonerOption) {
	return &provider{
		engine: engine,
	}, gone.GonerId("xorm")
}

type provider struct {
	gone.Flag
	engine      *engine
	gone.Logger `gone:"*"`
}

var xormInterface = gone.GetInterfaceType(new(gone.XormEngine))
var xormInterfaceSlice = gone.GetInterfaceType(new([]gone.XormEngine))

func (e *provider) Suck(conf string, v reflect.Value) gone.SuckError {
	if !e.engine.enableCluster {
		return gone.NewInnerError("cluster is not enabled, xorm only support cluster", gone.InjectError)
	}

	switch v.Type() {
	case xormInterface:
		if conf == "master" {
			v.Set(reflect.ValueOf(&engine{
				EngineInterface: e.engine.group.Master(),
			}))
			return nil
		} else {
			if strings.HasPrefix(conf, "slave") {
				slaveIndex := strings.TrimPrefix(conf, "slave")
				i, err := strconv.ParseInt(slaveIndex, 10, 64)
				if err != nil {
					return gone.NewInnerError("invalid slave index: "+slaveIndex, gone.InjectError)
				}
				slaves := e.engine.group.Slaves()
				if int(i) < len(slaves) {
					v.Set(reflect.ValueOf(
						&engine{
							EngineInterface: slaves[i],
						},
					))
					return nil
				}
			}
		}
		return gone.NewInnerError(fmt.Sprintf("invalid xorm interface conf: %s, only support master、salve{Index}", conf), gone.InjectError)

	case xormInterfaceSlice:
		if conf != "" {
			e.Warnf("ignore xorm interface slice conf: %s", conf)
		}

		engines := e.engine.group.Slaves()
		xormEngines := make([]gone.XormEngine, 0, len(engines))
		for _, eng := range engines {
			xormEngines = append(xormEngines, &engine{
				EngineInterface: eng,
			})
		}
		v.Set(reflect.ValueOf(xormEngines))
		return nil
	default:
		return gone.CannotFoundGonerByTypeError(v.Type())
	}
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
