package gone

import (
	"reflect"
	"sort"
)

// coffin represents a component container in the gone framework
type coffin struct {
	name  string
	goner any

	order        int
	onlyForName  bool
	forceReplace bool

	defaultTypeMap    map[reflect.Type]bool
	lazyFill          bool
	needInitBeforeUse bool
	isFill            bool
	isInit            bool
}

func newCoffin(goner any) *coffin {
	_, needInitBeforeUse := goner.(Initiator)
	if !needInitBeforeUse {
		_, needInitBeforeUse = goner.(InitiatorNoError)
	}
	if !needInitBeforeUse {
		_, needInitBeforeUse = goner.(NamedProvider)
	}

	return &coffin{
		goner:             goner,
		defaultTypeMap:    make(map[reflect.Type]bool),
		needInitBeforeUse: needInitBeforeUse,
	}
}

func (c *coffin) isDefault(t reflect.Type) bool {
	return c.defaultTypeMap[t]
}

// coffinList is a slice of coffin pointers that implements sort.Interface
type coffinList []*coffin

func (c coffinList) Len() int {
	return len(c)
}

func (c coffinList) Less(i, j int) bool {
	return c[i].order < c[j].order
}

func (c coffinList) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

// SortCoffins sorts a slice of coffins by their order
func SortCoffins(coffins []*coffin) {
	sort.Sort(coffinList(coffins))
}

func isInitiator(co *coffin) bool {
	if _, ok := co.goner.(Initiator); ok {
		return true
	}
	if _, ok := co.goner.(InitiatorNoError); ok {
		return true
	}
	return false
}
