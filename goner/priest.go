package goner

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/gin"
	gone_grpc "github.com/gone-io/gone/goner/grpc"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/redis"
	"github.com/gone-io/gone/goner/schedule"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/gone-io/gone/goner/urllib"
	"github.com/gone-io/gone/goner/xorm"
	gone_zap "github.com/gone-io/gone/goner/zap"
)

func BasePriest(cemetery gone.Cemetery) error {
	_ = tracer.Priest(cemetery)
	_ = logrus.Priest(cemetery)
	_ = config.Priest(cemetery)
	return nil
}

func ConfigPriest(cemetery gone.Cemetery) error {
	return config.Priest(cemetery)
}

func LogrusLoggerPriest(cemetery gone.Cemetery) error {
	return logrus.Priest(cemetery)
}

func ZapLoggerPriest(cemetery gone.Cemetery) error {
	return gone_zap.Priest(cemetery)
}

func GinPriest(cemetery gone.Cemetery) error {
	_ = gin.Priest(cemetery)
	return nil
}

func XormPriest(cemetery gone.Cemetery) error {
	_ = xorm.Priest(cemetery)
	return nil
}

func RedisPriest(cemetery gone.Cemetery) error {
	_ = redis.Priest(cemetery)
	return nil
}

func SchedulePriest(cemetery gone.Cemetery) error {
	_ = schedule.Priest(cemetery)
	return nil
}

func UrllibPriest(cemetery gone.Cemetery) error {
	return urllib.Priest(cemetery)
}

func GrpcServerPriest(cemetery gone.Cemetery) error {
	return gone_grpc.ServerPriest(cemetery)
}

func GrpcClientPriest(cemetery gone.Cemetery) error {
	return gone_grpc.ClientRegisterPriest(cemetery)
}
