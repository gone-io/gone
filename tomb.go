package gone

import (
	"reflect"
)

func NewTomb(goner Goner) Tomb {
	return &tomb{goner: goner, defaultTypes: make(map[reflect.Type]void)}
}

type void struct{}

var voidValue void

type tomb struct {
	id           GonerId
	goner        Goner
	reviveFlag   bool
	defaultTypes map[reflect.Type]void

	order Order
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

func (tombs Tombs) Len() int {
	return len(tombs)
}

func (tombs Tombs) Less(i, j int) bool {
	return tombs[i].GetOrder() < tombs[j].GetOrder()
}

func (tombs Tombs) Swap(i, j int) {
	tombs[i], tombs[j] = tombs[j], tombs[i]
}

func (tombs Tombs) GetTomByType(t reflect.Type) (filterTombs []Tomb) {
	for _, tomb := range tombs {
		if IsCompatible(t, tomb.GetGoner()) {
			filterTombs = append(filterTombs, tomb)
		}
	}
	return
}

func (t *tomb) IsDefault(T reflect.Type) bool {
	_, existed := t.defaultTypes[T]
	return existed
}

func (t *tomb) SetDefault(T reflect.Type) Tomb {
	t.defaultTypes[T] = voidValue
	return t
}

func (t *tomb) GetOrder() Order {
	return t.order
}
func (t *tomb) SetOrder(order Order) Tomb {
	t.order = order
	return t
}
