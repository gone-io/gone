package test

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
)

const pointNameA = "example-test-point-a"
const pointNameB = "example-test-point-b"

func NewPoint() (gone.Goner, gone.GonerId) {
	return &Point{}, pointNameA
}

func NewPointB() (gone.Goner, gone.GonerId) {
	return &Point{X: -1, Y: -1}, pointNameB
}

type Point struct {
	gone.GonerFlag
	X float64 `gone:"config,example.test.point.a-x"`
	Y float64 `gone:"config,example.test.point.a-y,default=200"`
}

type Line struct {
	gone.GonerFlag
	A Point `gone:"example-test-point-a"`
	B Point `gone:"example-test-point-b"`
}

func (*Line) Say() string {
	return ""
}

func Digger(cemetery gone.Cemetery) error {
	cemetery.Bury(NewPoint())
	cemetery.Bury(NewPointB())
	cemetery.Bury(&Line{})
	return config.Digger(cemetery)
}
