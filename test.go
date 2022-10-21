package gone

type TestHeaven[T Goner] interface {
	Heaven
	Run(func(T))
	RunAtId(id GonerId, fn func(T))
}

type GonerTestKit func(testGoner Goner)

// TestKit 新建TestHeaven
func TestKit[T Goner](goner T, priests ...Priest) TestHeaven[T] {
	h := New(priests...).(*heaven)
	testKit := &testHeaven[T]{
		heaven: h,
		goner:  goner,
	}

	//将自己安葬了，便于其他组件引用 和 感知自己在TestKit
	h.cemetery.bury(testKit, IdGoneTestKit)
	return testKit
}

type testHeaven[T Goner] struct {
	*heaven
	goner T
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

func (h *testHeaven[T]) Run(fn func(T)) {
	h.burial()
	h.run(h.cemetery.bury(h.goner), fn)
}

func (h *testHeaven[T]) RunAtId(id GonerId, fn func(T)) {
	h.burial()
	tomb := h.cemetery.GetTomById(id)
	if tomb == nil {
		panic(CannotFoundGonerByIdError(id))
	}
	h.run(tomb, fn)
}
