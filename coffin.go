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
	nameProvider        NamedProvider
	structFieldInjector StructFieldInjector
}

func newCoffin(goner any) *coffin {
	_, needInitBeforeUse := goner.(Initiator)
	if !needInitBeforeUse {
		_, needInitBeforeUse = goner.(InitiatorNoError)
	}
	provider := tryWrapGonerToProvider(goner)
	if provider != nil {
		needInitBeforeUse = true
	}

	var nameProvider NamedProvider
	if !needInitBeforeUse {
		nameProvider, needInitBeforeUse = goner.(NamedProvider)
	}

	if !needInitBeforeUse {
		_, needInitBeforeUse = goner.(StructFieldInjector)
	}

	var name string
	if namedGoner, ok := goner.(NamedGoner); ok {
		name = namedGoner.GonerName()
	}

	return &coffin{
		goner:             goner,
		name:              name,
		defaultTypeMap:    make(map[reflect.Type]bool),
		needInitBeforeUse: needInitBeforeUse,
		provider:          provider,
		nameProvider:      nameProvider,
	}
}

func (c *coffin) Name() string {
	if c.name != "" {
		return fmt.Sprintf("Goner(name=%s)", c.name)
	}
	return fmt.Sprintf("%T", c.goner)
}

func (c *coffin) CoundProvide(t reflect.Type) error {
	if IsCompatible(t, c.goner) {
		return nil
	}

	if c.nameProvider != nil {
		return nil
	}

	if c.provider != nil {
		if r := c.provider.Type(); r == t || r.Implements(t) {
			return nil
		}
	}
	return NewInnerErrorWithParams(GonerTypeNotMatch, "gone: %s cannot provide %s value", c.Name(), GetTypeName(t))
}

func (c *coffin) AddToDefault(t reflect.Type) error {
	if err := c.CoundProvide(t); err != nil {
		return err
	}
	c.defaultTypeMap[t] = true
	return nil
}

func (c *coffin) Provide(tagConf string, t reflect.Type) (any, error) {
	if IsCompatible(t, c.goner) {
		return c.goner, nil
	}

	if c.isDefault(t) {
		if c.provider != nil {
			if r := c.provider.Type(); r == t || r.Implements(t) {
				return c.provider.Provide(tagConf)
			}
		}
		if c.nameProvider != nil {
			return c.nameProvider.Provide(tagConf, t)
		}
	}
	return NewInnerErrorWithParams(NotSupport, "gone: %s cannot provide %s value", c.Name(), GetTypeName(t)), nil
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
