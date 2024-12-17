package gorm

import (
	"context"
	"github.com/gone-io/gone"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm/logger"
	"testing"
	"time"
)

//go:generate mockgen -package gorm -destination=./logger_mock_test.go github.com/gone-io/gone Logger

func Test_iLogger_Info(t *testing.T) {
	controller := gomock.NewController(t)
	mockLogger := NewMockLogger(controller)
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Trace(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes()

	gone.RunTest(func(in struct {
		logger logger.Interface `gone:"*"`
	}) {
		in.logger.LogMode(logger.Info)
		in.logger.Info(context.Background(), "info")
		in.logger.Warn(context.Background(), "warn")
		in.logger.Error(context.Background(), "error")
		begin := time.Now()
		in.logger.Trace(context.Background(), begin, func() (sql string, rowsAffected int64) {
			return "sql", 1
		}, nil)

		in.logger.Trace(context.Background(), begin.Add(-time.Second), func() (sql string, rowsAffected int64) {
			return "sql", 1
		}, nil)

	}, func(cemetery gone.Cemetery) error {
		cemetery.Bury(NewLogger())
		_ = gone.NewProviderPriest(func(tagConf string, param struct{}) (gone.Logger, error) {
			return mockLogger, nil
		})(cemetery)

		return config.Priest(cemetery)
	})
}
