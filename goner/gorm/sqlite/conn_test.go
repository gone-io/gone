package sqlite

import (
	"github.com/gone-io/gone"
	goneGorm "github.com/gone-io/gone/goner/gorm"
	gone_viper "github.com/gone-io/gone/goner/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestPriest(t *testing.T) {
	gone.RunTest(func(in struct {
		dial gorm.Dialector `gone:"*"`
	}) {
		assert.NotNil(t, in.dial)
		err := in.dial.(goneGorm.Applier).Apply(nil)
		assert.Nil(t, err)

	}, Load, gone_viper.Load)
}
