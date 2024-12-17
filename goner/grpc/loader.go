package gone_grpc

import (
	"github.com/gone-io/gone"
)

var serverLoad = gone.OnceLoad(func(loader gone.Loader) error {
	if err := loader.Load(&server{
		createListener: createListener,
	}); err != nil {
		return gone.ToError(err)
	}
	return nil
})

func ServerLoad(loader gone.Loader) error {
	return serverLoad(loader)
}

func ServerPriest(loader gone.Loader) error {
	return ServerLoad(loader)
}

var clientRegisterLoad = gone.OnceLoad(func(loader gone.Loader) error {
	if err := loader.Load(NewRegister()); err != nil {
		return gone.ToError(err)
	}
	return nil
})

func ClientRegisterLoad(loader gone.Loader) error {
	return clientRegisterLoad(loader)
}

func ClientRegisterPriest(loader gone.Loader) error {
	return ClientRegisterLoad(loader)
}
