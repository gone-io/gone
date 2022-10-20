package goner

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/gin"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
	"github.com/gone-io/gone/goner/xorm"
)

func BasePriest(cemetery gone.Cemetery) error {
	_ = tracer.Priest(cemetery)
	_ = logrus.Priest(cemetery)
	_ = config.Priest(cemetery)
	return nil
}

func GinPriest(cemetery gone.Cemetery) error {
	_ = gin.Priest(cemetery)
	return nil
}

func XormPriest(cemetery gone.Cemetery) error {
	_ = xorm.Priest(cemetery)
	return nil
}
