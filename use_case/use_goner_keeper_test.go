package use_case

import (
	"github.com/gone-io/gone"
	"testing"
)

type useKeeper struct {
	gone.Flag
	keeper gone.GonerKeeper `gone:"*"`
	core   *gone.Core       `gone:"*"`
}

func (u *useKeeper) Test(t *testing.T) {
	goner := u.keeper.GetGonerByName("*")
	if goner != u.core {
		t.Fatal("keeper get core error")
	}
}

func TestGonerKeeper(t *testing.T) {
	gone.
		Prepare().
		Load(&useKeeper{}).
		Run(func(k *useKeeper) {
			k.Test(t)
		})

}
