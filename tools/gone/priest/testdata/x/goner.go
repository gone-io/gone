package x

import "github.com/gone-io/gone"

//go:gone
func New() gone.Goner {
	return &goner{}
}

type goner struct {
	gone.Flag
}
