package gone

import (
	"fmt"
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

	defaultTypeMap      map[reflect.Type]bool
	lazyFill            bool
	needInitBeforeUse   bool
	isFill              bool
	isInit              bool
	provider            *wrapProvider
	namedProvider       NamedProvider
	structFieldInjector StructFieldInjector
}

func newCoffin(goner any) *coffin {
	co := &coffin{
		goner:          goner,
		defaultTypeMap: make(map[reflect.Type]bool),
	}

	if namedGoner, ok := goner.(NamedGoner); ok {
		co.name = namedGoner.GonerName()
	}

	if namedProvider, ok := goner.(NamedProvider); ok {
		co.needInitBeforeUse = true
		co.namedProvider = namedProvider
	} else if provider := tryWrapGonerToProvider(goner); provider != nil {
		co.needInitBeforeUse = true
		co.provider = provider
		//co.defaultTypeMap[provider.Type()] = true
	} else if _, ok := goner.(Initiator); ok {
		co.needInitBeforeUse = true
	} else if _, ok := goner.(InitiatorNoError); ok {
		co.needInitBeforeUse = true
	} else if _, ok := goner.(StructFieldInjector); ok {
		co.needInitBeforeUse = true
	}

	return co
}

func (c *coffin) Name() string {
	if c.name != "" {
		return fmt.Sprintf("Goner(name=%s)", c.name)
	}
	return fmt.Sprintf("%T", c.goner)
}

func (c *coffin) CoundProvide(t reflect.Type, byName bool) error {
	if IsCompatible(t, c.goner) {
		return nil
	}

	if c.provider != nil && c.provider.ProvideTypeCompatible(t) {
		return nil
	}

	if c.namedProvider != nil && (byName || c.isDefault(t)) {
		return nil
	}

	return NewInnerErrorWithParams(GonerTypeNotMatch, "%q cannot provide %q value", c.Name(), GetTypeName(t))
}

func (c *coffin) AddToDefault(t reflect.Type) error {
	if err := c.CoundProvide(t, true); err != nil {
		return err
	}
	c.defaultTypeMap[t] = true
	return nil
}

func (c *coffin) Provide(byName bool, tagConf string, t reflect.Type) (any, error) {
	if IsCompatible(t, c.goner) {
		return c.goner, nil
	}

	if c.provider != nil && c.provider.ProvideTypeCompatible(t) {
		return c.provider.Provide(tagConf)
	}

	if c.namedProvider != nil && (byName || c.isDefault(t)) {
		return c.namedProvider.Provide(tagConf, t)
	}

	return nil, NewInnerErrorWithParams(NotSupport, "gone: %s cannot provide %s value", c.Name(), GetTypeName(t))
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
