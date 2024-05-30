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

func NewBuryMockCemeteryForTest() Cemetery {
	return newCemetery()
}

// Test Use for writing test cases
// example:
//
//	gone.Prepare(priests...).Test(func(in struct{
//	    cemetery Cemetery `gone:"*"`
//	}) {
//
//	  // test code
//	})
func (p *Preparer) Test(fn any) {
	p.AfterStart(fn).Run()
}
