package use_case

import (
	"github.com/gone-io/gone/v2"
	"testing"
)

type useKeeper struct {
	gone.Flag
	keeper gone.GonerKeeper `gone:"*"`
	core   gone.Loader      `gone:"*"`
}

func (u *useKeeper) Test(t *testing.T) {
	goner := u.keeper.GetGonerByName(gone.DefaultProviderName)
	if goner != u.core {
		t.Fatal("keeper get core error")
	}
}

func TestGonerKeeper(t *testing.T) {
	gone.
		NewApp().
		Load(&useKeeper{}).
		Run(func(k *useKeeper) {
			k.Test(t)
		})

}
