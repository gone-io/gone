package gorm

import (
	"github.com/gone-io/gone"
)

// Priest gorm的priest
func Priest(cemetery gone.Cemetery) error {
	cemetery.Bury(NewLogger())
	return ProviderPriest(cemetery)
}
