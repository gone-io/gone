package goner

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPriest(t *testing.T) {
	var fnList = []func(cemetery gone.Cemetery) error{
		BasePriest,
		ConfigPriest,
		LogrusLoggerPriest,
		ZapLoggerPriest,
		UrllibPriest,
		GrpcClientPriest,
		CMuxPriest,
	}

	for _, fn := range fnList {
		t.Run(gone.GetFuncName(fn), func(t *testing.T) {
			gone.Prepare(fn).Run()
		})
	}

	var servers = []func(cemetery gone.Cemetery) error{
		XormPriest,
		RedisPriest,
		SchedulePriest,
		GrpcServerPriest,
		GinPriest,
	}
	for _, fn := range servers {
		t.Run(gone.GetFuncName(fn), func(t *testing.T) {
			gone.Prepare().Test(func(cemetery gone.Cemetery) {
				err := fn(cemetery)
				assert.Nil(t, err)
			})
		})
	}
}
