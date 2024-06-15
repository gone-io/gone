package main

import "github.com/gone-io/gone"

//go:gone
func NewAdder() gone.Goner {
	return &Adder{}
}

//go:gone
func NewComputer() gone.Goner {
	return &Computer{}
}

type Adder struct {
	gone.Flag
}

func (a *Adder) Add(a1, a2 int) int {
	return a1 + a2
}

type Computer struct {
	gone.Flag
	adder Adder `gone:"*"`
}

func (c *Computer) Compute() {
	println("I want to compute!")
	println("1000 add 2000 is", c.adder.Add(1000, 2000))
}

// AfterRevive will execute after revive
func (c *Computer) AfterRevive() gone.AfterReviveError {
	// boot
	c.Compute()

	return nil
}
