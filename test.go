package gone

type TestHeaven[T Goner] interface {
	Heaven
	Run(func(T))
	RunAtId(id GonerId, fn func(T))
}

type GonerTestKit func(testGoner Goner)

// TestKit 新建TestHeaven
func TestKit[T Goner](goner T, digGraves ...Digger) TestHeaven[T] {
	h := New(digGraves...)
	return &testHeaven[T]{
		heaven: h.(*heaven),
		goner:  goner,
	}
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
	deps, err := h.cemetery.reviveOneDep(tomb)
	h.installAngelHook(deps)

	if err != nil {
		panic(err)
	}
	h.startFlow()
	fn(h.goner)
	h.stopFlow()
}

func (h *testHeaven[T]) Run(fn func(T)) {
	h.dig()
	h.run(h.cemetery.bury(h.goner), fn)
}

func (h *testHeaven[T]) RunAtId(id GonerId, fn func(T)) {
	h.dig()
	tomb := h.cemetery.GetTomById(id)
	if tomb == nil {
		panic(CannotFoundGonerByIdError(id))
	}
	h.run(tomb, fn)
}
