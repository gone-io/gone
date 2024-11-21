package xorm

import (
	"fmt"
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"reflect"
	"testing"
	"xorm.io/xorm"
)

func Test_provider_Suck(t *testing.T) {
	controller := gomock.NewController(t)
	engineInterface := NewMockEngineInterface(controller)
	var defaultDb gone.XormEngine

	enginesMap := make(map[string]*xorm.Engine)
	for i := 0; i < 10; i++ {
		enginesMap[fmt.Sprintf("db%d", i)] = &xorm.Engine{}
	}

	gone.RunTest(func(i struct {
		p *provider `gone:"*"`
	}) {
		t.Run("get default gone.XormEngine", func(t *testing.T) {
			var X gone.XormEngine
			err := i.p.Suck("", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)
			assert.Equal(t, X, defaultDb)
		})

		t.Run("get default master gone.XormEngine", func(t *testing.T) {
			var X gone.XormEngine
			err := i.p.Suck("master", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)
			assert.Equal(t, X.(*wrappedEngine).EngineInterface, enginesMap["db0"])
		})

		t.Run("get default slave gone.XormEngine", func(t *testing.T) {
			var X gone.XormEngine
			err := i.p.Suck("slave", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)
			assert.Equal(t, X.(*wrappedEngine).EngineInterface, enginesMap["db1"])

			err = i.p.Suck("slave=1", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)
			assert.Equal(t, X.(*wrappedEngine).EngineInterface, enginesMap["db2"])

			err = i.p.Suck("slave=0", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)
			assert.Equal(t, X.(*wrappedEngine).EngineInterface, enginesMap["db1"])
		})

		t.Run("get default slave gone.XormEngine with error index", func(t *testing.T) {
			var X gone.XormEngine
			err := i.p.Suck("slave=-1", reflect.ValueOf(&X).Elem())
			assert.Error(t, err)

			err = i.p.Suck("slave=2", reflect.ValueOf(&X).Elem())
			assert.Error(t, err)
		})

		t.Run("get default slaves with []gone.XormEngine", func(t *testing.T) {
			var X []gone.XormEngine
			err := i.p.Suck("", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)
			assert.Equal(t, len(X), 2)
			assert.Equal(t, X[0].(*wrappedEngine).EngineInterface, enginesMap["db1"])
			assert.Equal(t, X[1].(*wrappedEngine).EngineInterface, enginesMap["db2"])
		})

		t.Run("get user.database gone.XormEngine", func(t *testing.T) {
			var X gone.XormEngine
			err := i.p.Suck("db=user.database", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)

			err = i.p.Suck("db=user.database,master", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)
			assert.Equal(t, X.(*wrappedEngine).EngineInterface, enginesMap["db3"])
		})

		t.Run("get user.database slave gone.XormEngine", func(t *testing.T) {
			var X gone.XormEngine
			err := i.p.Suck("db=user.database,slave=1", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)
			assert.Equal(t, X.(*wrappedEngine).EngineInterface, enginesMap["db5"])

			err = i.p.Suck("db=user.database,slave=2", reflect.ValueOf(&X).Elem())
			assert.Error(t, err)
		})

		t.Run("get user.database slave with []gone.XormEngine", func(t *testing.T) {
			var X []gone.XormEngine
			err := i.p.Suck("db=user.database", reflect.ValueOf(&X).Elem())
			assert.Nil(t, err)
			assert.Equal(t, len(X), 2)
			assert.Equal(t, X[0].(*wrappedEngine).EngineInterface, enginesMap["db4"])
			assert.Equal(t, X[1].(*wrappedEngine).EngineInterface, enginesMap["db5"])
		})

	}, func(cemetery gone.Cemetery) error {
		newFunc := func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
			return enginesMap[dataSourceName], nil
		}

		e := wrappedEngine{
			newFunc: newFunc,
			newSession: func(engineInterface xorm.EngineInterface) XInterface {
				session := NewMockXInterface(controller)
				return session
			},
			conf: Conf{
				EnableCluster: true,
			},
			masterConf: &ClusterNodeConf{
				DriverName: "mysql",
				DSN:        "db0",
			},
			slavesConf: []*ClusterNodeConf{
				{
					DriverName: "mysql",
					DSN:        "db1",
				},
				{
					DriverName: "mysql",
					DSN:        "db2",
				},
			},
		}

		err := e.create()
		if err != nil {
			return err
		}

		e.EngineInterface = engineInterface
		e.unitTest = true
		defaultDb = &e

		cemetery.Bury(NewProvider(&e))
		_ = config.Priest(cemetery)
		return nil
	})
}
