package gone

import "reflect"

func NewTomb(goner Goner) Tomb {
	return &tomb{goner: goner}
}

type tomb struct {
	id         GonerId
	goner      Goner
	reviveFlag bool
	isDefault  bool
}

func (t *tomb) SetId(id GonerId) Tomb {
	t.id = id
	return t
}

func (t *tomb) GetId() GonerId {
	return t.id
}

func (t *tomb) GetGoner() Goner {
	return t.goner
}

func (t *tomb) GonerIsRevive(flags ...bool) bool {
	if len(flags) > 0 {
		t.reviveFlag = flags[0]
	}
	return t.reviveFlag
}

type Tombs []Tomb

func (tombs Tombs) GetTomByType(t reflect.Type) (filterTombs []Tomb) {
	for _, tomb := range tombs {
		if IsCompatible(t, tomb.GetGoner()) {
			filterTombs = append(filterTombs, tomb)
		}
	}
	return
}

func (t *tomb) IsDefault() bool {
	return t.isDefault
}

func (t *tomb) SetDefault(isDefault bool) Tomb {
	t.isDefault = isDefault
	return t
}
