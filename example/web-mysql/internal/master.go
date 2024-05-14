package internal

import (
	_ "github.com/go-sql-driver/mysql" //导入mysql驱动
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner"
)

//go:generate gone priest -s . -p $GOPACKAGE -f Priest -o priest.go
func MasterPriest(cemetery gone.Cemetery) error {
	_ = goner.XormPriest(cemetery)
	_ = goner.GinPriest(cemetery)

	_ = Priest(cemetery)
	return nil
}
