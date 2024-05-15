package test

import (
	"github.com/gone-io/gone"
	"math"
)

//go:gone
func NewDistanceCalculator() gone.Goner {
	return &distanceCalculator{}
}

type distanceCalculator struct {
	gone.Flag

	originPoint IPoint `gone:"*"`
}

func (d *distanceCalculator) CalculateDistanceFromOrigin(x, y int) float64 {
	originX, originY := d.originPoint.GetX(), d.originPoint.GetY()
	return math.Sqrt(math.Pow(float64(x-originX), 2) + math.Pow(float64(y-originY), 2))
}
