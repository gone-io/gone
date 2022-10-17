package gone

import "os"

type TestHeaven interface {
	Heaven
	TestRun(testGoner Goner, kit GonerTestKit) error
}

type GonerTestKit func(testGoner Goner)

// New 新建Heaven
func NewForTest(digGraves ...Digger) TestHeaven {
	cemetery := NewCemetery()
	return &heaven{
		cemetery:  cemetery.Bury(cemetery, IdGoneCemetery),
		digGraves: digGraves,
		signal:    make(chan os.Signal),
	}
}

func (h *heaven) TestRun(testGoner Goner, kit GonerTestKit) error {
	h.dig()
	h.cemetery.Bury(testGoner)
	err := h.cemetery.reviveOne(NewTomb(testGoner))
	if err != nil {
		return err
	}
	kit(testGoner)
	return nil
}
