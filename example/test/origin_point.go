package test

import "github.com/gone-io/gone"

type originPoint struct {
	gone.Flag
}

//go:gone
func NewOriginPoint() gone.Goner {
	return &originPoint{}
}

func (o *originPoint) GetX() int {
	return 100
}
func (o *originPoint) GetY() int {
	return 200
}
