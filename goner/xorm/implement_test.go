package xorm

import (
	"errors"
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
	"xorm.io/xorm"
)

func Test_engine(t *testing.T) {
	gone.Prepare(func(cemetery gone.Cemetery) error {
		_ = config.Priest(cemetery)
		_ = logrus.Priest(cemetery)
		return nil
	}).AfterStart(func(in struct {
		logger   gone.Logger   `gone:"gone-logger"`
		cemetery gone.Cemetery `gone:"gone-cemetery"`
	}) {

		controller := gomock.NewController(t)
		defer controller.Finish()

		engineInterface := NewMockEngineInterface(controller)
		engineInterface.EXPECT().SetConnMaxLifetime(gomock.Any())
		engineInterface.EXPECT().SetMaxOpenConns(gomock.Any())
		engineInterface.EXPECT().SetMaxIdleConns(gomock.Any())
		engineInterface.EXPECT().SetLogger(gomock.Any())
		engineInterface.EXPECT().Ping()
		engineInterface.EXPECT().SQL(gomock.Any(), gomock.Any()).Return(nil)

		e := wrappedEngine{
			log: in.logger,
			newFunc: func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
				return nil, errors.New("test")
			},
		}

		err := e.Start(in.cemetery)
		assert.Error(t, err)

		e.newFunc = func(driverName string, dataSourceName string) (xorm.EngineInterface, error) {
			return engineInterface, nil
		}

		err = e.Start(in.cemetery)
		assert.NoError(t, err)

		originEngine := e.GetOriginEngine()
		assert.Equalf(t, engineInterface, originEngine, "origin wrappedEngine is not equal")

		_ = e.Sqlx("select * from user where id = ?", 1)

		err = e.Start(in.cemetery)
		assert.Error(t, err)

	}).Run()
}
