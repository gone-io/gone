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
	a         *AGoner `gone:"*"` //Gone Tag `gone` tell the framework that this field will be injected by the framework
}

// AfterRevive executed After the Goner is revived; After `gone.Run`, gone framework detects the AfterRevive function on goners and runs it.
func (g *BGoner) AfterRevive() gone.AfterReviveError {
	g.a.Say()

	return nil
}

// Priest Responsible for putting Goners that need to be used into the framework
func Priest(cemetery gone.Cemetery) error {
	cemetery.Bury(&AGoner{Name: "Injected Goner"})
	cemetery.Bury(&BGoner{})
	return nil
}

func main() {

	// start gone framework
	gone.Run(Priest)
}
