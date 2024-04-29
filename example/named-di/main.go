package main

import "github.com/gone-io/gone"

type AGoner struct {
	gone.Flag //tell the framework that this struct is a Goner
	Name      string
}

func (g *AGoner) Say() {
	println("I am the AGoner, My name is", g.Name)
}

type BGoner struct {
	gone.Flag         //tell the framework that this struct is a Goner
	a         *AGoner `gone:"*"`  //匿名注入一个AGoner
	a1        *AGoner `gone:"A1"` //具名注入A1
	a2        *AGoner `gone:"A2"` //具名注入A2
}

// AfterRevive executed After the Goner is revived; After `gone.Run`, gone framework detects the AfterRevive function on goners and runs it.
func (g *BGoner) AfterRevive() gone.AfterReviveError {
	g.a.Say()

	return nil
}

// NewA1 构造A1 AGoner
func NewA1() (gone.Goner, gone.GonerId) {
	return &AGoner{Name: "Injected Goner1"}, "A1"
}

// NewA2 构造A2 AGoner
func NewA2() (gone.Goner, gone.GonerId) {
	return &AGoner{Name: "Injected Goner2"}, "A2"
}

// Priest Responsible for putting Goners that need to be used into the framework
func Priest(cemetery gone.Cemetery) error {
	cemetery.
		Bury(NewA1()).
		Bury(NewA2()).
		Bury(&BGoner{})
	return nil
}

func main() {

	// start gone framework
	gone.Run(Priest)
}
