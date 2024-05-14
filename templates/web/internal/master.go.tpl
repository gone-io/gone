package internal

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
)

//go:generate gone priest -s . -p $GOPACKAGE -f Priest -o priest.go
func MasterPriest(cemetery gone.Cemetery) error {
	_ = goner.GinPriest(cemetery)

	_ = Priest(cemetery)
	return nil
}
