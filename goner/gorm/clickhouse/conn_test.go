package postgres

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
	goneGorm "github.com/gone-io/gone/goner/gorm"
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

	}, func(cemetery gone.Cemetery) error {
		_ = config.Priest(cemetery)
		return Priest(cemetery)
	})
}
