package goner

import (
	"github.com/gone-io/gone"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasePriest(t *testing.T) {
	gone.Prepare().Run(func(cemetery gone.Cemetery) {
		err := BasePriest(cemetery)
		assert.Nil(t, err)
	})
}

func TestGinPriest(t *testing.T) {
	gone.Prepare().Run(func(cemetery gone.Cemetery) {
		err := GinPriest(cemetery)
		assert.Nil(t, err)
	})
}

func TestXormPriest(t *testing.T) {
	gone.Prepare().Run(func(cemetery gone.Cemetery) {
		err := XormPriest(cemetery)
		assert.Nil(t, err)
	})
}

func TestRedisPriest(t *testing.T) {
	cemetery := gone.NewBuryMockCemeteryForTest()
	err := RedisPriest(cemetery)
	assert.Nil(t, err)
}

func TestSchedulePriest(t *testing.T) {
	cemetery := gone.NewBuryMockCemeteryForTest()
	err := SchedulePriest(cemetery)
	assert.Nil(t, err)
}

func TestUrllibPriest(t *testing.T) {
	cemetery := gone.NewBuryMockCemeteryForTest()
	err := UrllibPriest(cemetery)
	assert.Nil(t, err)
}

func TestGrpcServerPriest(t *testing.T) {
	cemetery := gone.NewBuryMockCemeteryForTest()
	err := GrpcServerPriest(cemetery)
	assert.Nil(t, err)
}

func TestGrpcClientPriest(t *testing.T) {
	cemetery := gone.NewBuryMockCemeteryForTest()
	err := GrpcClientPriest(cemetery)
	assert.Nil(t, err)
}

func TestConfigPriest(t *testing.T) {
	cemetery := gone.NewBuryMockCemeteryForTest()
	err := ConfigPriest(cemetery)
	assert.Nil(t, err)
}

func TestLogrusLoggerPriest(t *testing.T) {
	gone.Prepare().Test(func(cemetery gone.Cemetery) {
		err := LogrusLoggerPriest(cemetery)
		assert.Nil(t, err)
	})
}

func TestZapLoggerPriest(t *testing.T) {
	gone.Prepare().Test(func(cemetery gone.Cemetery) {
		err := ZapLoggerPriest(cemetery)
		assert.Nil(t, err)
	})
}
