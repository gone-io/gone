package gone

import (
	"reflect"
)

// Option is an interface for configuring components loaded into the gone framework.
type Option interface {
	Apply(c *coffin) error
}

type option struct {
	apply func(c *coffin) error
}

func (o option) Apply(c *coffin) error {
	if o.apply == nil {
		return nil
	}
	return o.apply(c)
}

// IsDefault returns an Option that marks a component as the default implementation for its type.
// When multiple components of the same type exist, the default one will be used for injection
// if no specific name is requested.
//
// Example usage:
//
//	gone.Load(&EnvConfigure{}, gone.IsDefault())
//
// This marks EnvConfigure as the default implementation to use when injecting its interface type.
func IsDefault(objPointers ...any) Option {
	for i, p := range objPointers {
		if p == nil {
			panic(NewInnerErrorWithParams(LoadedError, "gone: IsDefault() requires a non-nil pointer, %dth parameter got nil", i+1))
		}
		of := reflect.TypeOf(p)
		if of.Kind() != reflect.Ptr {
			panic(NewInnerErrorWithParams(LoadedError, "gone: IsDefault() requires a pointer, %dth parameter got <%T> ", i+1, p))
		}
	}

	return option{
		apply: func(c *coffin) error {
			typeMap := make(map[reflect.Type]bool)
			for i, p := range objPointers {
				of := reflect.TypeOf(p)
				if of.Kind() != reflect.Ptr {
					return NewInnerErrorWithParams(LoadedError, "gone: IsDefault() requires a pointer, %dth parameter got <%T> ", i+1, p)
				}
				if _, ok := typeMap[of.Elem()]; !ok {
					typeMap[of.Elem()] = true
				}
			}

			for t := range typeMap {
				c.defaultTypeMap[t] = true
			}
			if len(typeMap) == 0 {
				c.defaultTypeMap[reflect.TypeOf(c.goner)] = true
			}
			return nil
		},
	}
}

// Order returns an Option that sets the start order for a component.
// Components with lower order values will be started before those with higher values.
// This can be used to control started sequence when specific ordering is required.
//
// Example usage:
//
//	gone.Load(&Database{}, gone.Order(1))  // started first
//	gone.Load(&Service{}, gone.Order(2))   // started second
//
// Parameters:
//   - order: Integer value indicating relative started order
func Order(order int) Option {
	return option{
		apply: func(c *coffin) error {
			c.order = order
			return nil
		},
	}
}

func HighStartPriority() Option {
	return Order(-100)
}

func MediumStartPriority() Option {
	return Order(0)
}

func LowStartPriority() Option {
	return Order(100)
}

// Name returns an Option that sets a custom name for a component.
// Components can be looked up by this name when injecting dependencies.
//
// Example usage:
//
//	gone.Load(&EnvConfigure{}, gone.GonerName("configure"))
//
// Parameters:
//   - name: String identifier to use for this component
func Name(name string) Option {
	return option{
		apply: func(c *coffin) error {
			c.name = name
			return nil
		},
	}
}

// OnlyForName returns an Option that marks a component as only available for name-based injection.
// When this option is used, the component will not be registered as a type provider,
// meaning it can only be injected by explicitly referencing its name.
//
// Example usage:
//
//	gone.Load(&EnvConfigure{}, gone.GonerName("configure"), gone.OnlyForName())
//	// Now EnvConfigure can only be injected using `gone:"configure"` tag
//	// And not through interface type matching
func OnlyForName() Option {
	return option{
		apply: func(c *coffin) error {
			c.onlyForName = true
			return nil
		},
	}
}

// ForceReplace returns an Option that allows replacing existing components with the same name or type.
// When loading a component with this option:
// - If a component with the same name already exists, it will be replaced
// - If a provider for the same type already exists, it will be replaced
//
// Example usage:
//
//	gone.Load(&MyService{}, gone.GonerName("service"), gone.ForceReplace())
//	// This will replace any existing component named "service"
func ForceReplace() Option {
	return option{
		apply: func(c *coffin) error {
			c.forceReplace = true
			return nil
		},
	}
}

func LazyFill() Option {
	return option{
		apply: func(c *coffin) error {
			c.lazyFill = true
			return nil
		},
	}
}
