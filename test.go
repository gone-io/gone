package gone

import (
	"reflect"
)

// Test Use for writing test cases, refer to [example](https://github.com/gone-io/gone/blob/main/example/test/goner_test.go)
func Test[T Goner](fn func(goner T), priests ...Priest) {
	Prepare(priests...).Run(func(in struct {
		cemetery Cemetery `gone:"*"`
	}) {
		ft := reflect.TypeOf(fn)
		t := ft.In(0).Elem()
		theTombs := in.cemetery.GetTomByType(t)
		if len(theTombs) == 0 {
			panic(CannotFoundGonerByTypeError(t))
		}
		fn(theTombs[0].GetGoner().(T))
	})
}

// TestAt Use for writing test cases, test a specific ID of Goner
func TestAt[T Goner](id GonerId, fn func(goner T), priests ...Priest) {
	Prepare(priests...).Run(func(in struct {
		cemetery Cemetery `gone:"*"`
	}) {
		theTomb := in.cemetery.GetTomById(id)
		if theTomb == nil {
			panic(CannotFoundGonerByIdError(id))
		}
		g, ok := theTomb.GetGoner().(T)
		if !ok {
			panic(NotCompatibleError(reflect.TypeOf(g).Elem(), reflect.TypeOf(theTomb.GetGoner()).Elem()))
		}
		fn(g)
	})
}

type BuryMockCemetery struct {
	Cemetery
	m map[GonerId]Goner
}

func (c *BuryMockCemetery) Bury(g Goner, options ...GonerOption) Cemetery {
	for _, option := range options {
		if id, ok := option.(GonerId); ok {
			c.m[id] = g
			return c
		}
	}
	id := GetGoneDefaultId(g)
	c.m[id] = g
	return c
}

func (c *BuryMockCemetery) GetTomById(id GonerId) Tomb {
	goner := c.m[id]
	if goner == nil {
		return nil
	}
	return NewTomb(goner)
}

func (c *BuryMockCemetery) GetTomByType(t reflect.Type) (list []Tomb) {
	for _, g := range c.m {
		if reflect.TypeOf(g).Elem() == t {
			list = append(list, NewTomb(g))
		}
	}
	return list
}

func (c *BuryMockCemetery) BuryOnce(goner Goner, options ...GonerOption) Cemetery {
	var id GonerId

	for _, option := range options {
		switch option.(type) {
		case GonerId:
			id = option.(GonerId)
		}
	}
	if id == "" {
		panic(NewInnerError("GonerId is empty, must have gonerId option", MustHaveGonerId))
	}

	if nil == c.GetTomById(id) {
		c.Bury(goner, options...)
	}
	return c
}

func NewBuryMockCemeteryForTest() Cemetery {
	c := BuryMockCemetery{}
	c.m = make(map[GonerId]Goner)
	return &c
}

func (p *Preparer) Test(fn any) {
	p.AfterStart(fn).Run()
}
