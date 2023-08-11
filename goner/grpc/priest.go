package gone_grpc

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
	"github.com/gone-io/gone/goner/tracer"
)

func ServerPriest(cemetery gone.Cemetery) error {
	_ = config.Priest(cemetery)
	_ = logrus.Priest(cemetery)
	_ = tracer.Priest(cemetery)
	cemetery.Bury(NewServer())
	return nil
}

func ClientRegisterPriest(cemetery gone.Cemetery) error {
	_ = config.Priest(cemetery)
	_ = logrus.Priest(cemetery)
	_ = tracer.Priest(cemetery)
	cemetery.Bury(NewRegister())
	return nil
}
