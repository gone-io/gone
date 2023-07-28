package gone_grpc

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	"github.com/gone-io/gone/goner/logrus"
)

func ServerPriest(cemetery gone.Cemetery) error {
	_ = config.Priest(cemetery)
	_ = logrus.Priest(cemetery)
	cemetery.Bury(NewServer())
	return nil
}
