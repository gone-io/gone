package gone

import (
	"reflect"
)

type TestHeaven[T Goner] interface {
	Heaven
	WithId(id GonerId) TestHeaven[T]
	WithPriest(priests ...Priest) TestHeaven[T]
	Run()
}

type testHeaven[T Goner] struct {
	*heaven
	testFn      func(T)
	testGonerId GonerId
}

func (h *testHeaven[T]) WithId(id GonerId) TestHeaven[T] {
	h.testGonerId = id
	return h
}

func (h *testHeaven[T]) WithPriest(priests ...Priest) TestHeaven[T] {
	h.heaven = New(priests...).(*heaven)
	return h
}

func (h *testHeaven[T]) installAngelHook(deps []Tomb) {
	angleTombs := Tombs(deps).GetTomByType(getAngelType())
	for _, tomb := range angleTombs {
		angel := tomb.GetGoner().(Angel)
		h.BeforeStart(angel.Start)
		h.BeforeStop(angel.Stop)
	}
}

func (h *testHeaven[T]) run(tomb Tomb, fn func(T)) {
	deps, err := h.cemetery.reviveDependence(tomb)
	if err != nil {
		panic(err)
	}
	deps = append(deps, tomb)
	h.installAngelHook(deps)
	h.startFlow()
	fn(tomb.GetGoner().(T))
	h.stopFlow()
}

func (h *testHeaven[T]) getTestGonerType() reflect.Type {
	t := new(T)
	return reflect.TypeOf(t).Elem()
}

func (h *testHeaven[T]) Run() {
	//将自己安葬了，便于其他组件引用 和 感知自己在TestKit
	h.cemetery.Bury(h, IdGoneTestKit)

	h.burial()

	paramType := h.getTestGonerType()
	var tomb Tomb = nil
	if h.testGonerId != "" {
		tomb = h.cemetery.GetTomById(h.testGonerId)
		if tomb == nil {
			panic(CannotFoundGonerByIdError(h.testGonerId))
		}
		if tomb != nil && !isCompatible(paramType, tomb.GetGoner()) {
			panic(NotCompatibleError(paramType, reflect.TypeOf(tomb.GetGoner()).Elem()))
		}
	} else {
		list := h.cemetery.GetTomByType(paramType)
		if len(list) > 0 {
			if len(list) > 1 {
				h.Warnf("more than one Goner found by type")
			}
			tomb = list[0]
		}

		if tomb == nil {
			panic(CannotFoundGonerByTypeError(paramType))
		}
	}
	h.run(tomb, h.testFn)
	return
}

func TestKit[T Goner](fn func(T)) TestHeaven[T] {
	return &testHeaven[T]{testFn: fn}
}

// Test 用于编写测试用例，参考[示例](https://github.com/gone-io/gone/blob/main/example/test/goner_test.go)
func Test[T Goner](fn func(T), priests ...Priest) {
	TestKit(fn).WithPriest(priests...).Run()
}

// TestAt 用于编写测试用例，测试某个特定ID的Goner
func TestAt[T Goner](id GonerId, fn func(T), priests ...Priest) {
	TestKit(fn).WithId(id).WithPriest(priests...).Run()
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
