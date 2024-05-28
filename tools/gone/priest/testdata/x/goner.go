package x

import "github.com/gone-io/gone"

//go:gone
func New() gone.Goner {
	return &goner{}
}

type goner struct {
	gone.Flag
}

//go:gone
func Priest(cemetery gone.Cemetery) error {
	//todo
	return nil
}

//test//test//test//test//test//test//test//test//test//test//test//test//test//test//test//test
