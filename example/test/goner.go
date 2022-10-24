package test

import (
	"github.com/gone-io/gone"
	"github.com/gone-io/gone/goner/config"
)

const pointNameA = "example-test-point-a"
const pointNameB = "example-test-point-b"

func NewPointA() (gone.Goner, gone.GonerId) {
	return &Point{}, pointNameA
}

func NewPointB() (gone.Goner, gone.GonerId) {
	return &Point{X: -1, Y: -1}, pointNameB
}

type Point struct {
	gone.Flag
	X int `gone:"config,example.test.point.a-x"`
	Y int `gone:"config,example.test.point.a-y,default=200"`
}

type Line struct {
	gone.Flag
	A *Point `gone:"example-test-point-a"`
	B *Point `gone:"example-test-point-b"`
}

func (*Line) Say() string {
	return ""
}

func NewLine() *Line {
	return &Line{}
}

func Priest(cemetery gone.Cemetery) error {
	cemetery.Bury(NewPointA())
	cemetery.Bury(NewPointB())
	cemetery.Bury(NewLine())
	return config.Priest(cemetery)
}
