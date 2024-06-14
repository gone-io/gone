package gone_grpc

import (
	"github.com/gone-io/gone"
)

func ServerPriest(cemetery gone.Cemetery) error {
	cemetery.Bury(NewServer())
	return nil
}

func ClientRegisterPriest(cemetery gone.Cemetery) error {
	cemetery.Bury(NewRegister())
	return nil
}
