package xorm

import (
	"github.com/gone-io/gone"
	"go.uber.org/mock/gomock"
	"testing"
	"xorm.io/xorm"
)

func Test_provider_Suck(t *testing.T) {
	controller := gomock.NewController(t)

	gone.RunTest(func() {
		println("ok")

	}, func(cemetery gone.Cemetery) error {
		e := wrappedEngine{
			//Logger: in.logger,
			newFunc: func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
				engineInterface := NewMockEngineInterface(controller)
				engineInterface.EXPECT().Ping().Return(nil).AnyTimes()

				return engineInterface, nil
			},
			newSession: func(engineInterface xorm.EngineInterface) XInterface {
				session := NewMockXInterface(controller)
				//session.EXPECT().Begin().Return(nil)
				//session.EXPECT().Close().Return(nil)
				//session.EXPECT().Rollback().Return(errors.New("error"))

				return session
			},
			conf: &Conf{
				EnableCluster: true,
				MasterConf:    &ClusterNodeConf{},
				SlavesConf: []*ClusterNodeConf{
					{
						DriverName: "mysql",
						DSN:        "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
					},
					{
						DriverName: "mysql",
						DSN:        "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local",
					},
				},
			},
		}
		err := e.Start(cemetery)
		if err != nil {
			return err
		}

		cemetery.Bury(NewProvider(&e))
		return nil
	})
}
