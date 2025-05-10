package gone

import "fmt"

// Load loads a Goner into the Core with optional configuration options.
//
// A Goner can be registered by name if it implements NamedGoner interface.
// If a Goner with the same name already exists:
// - If forceReplace option is true, the old one will be replaced
// - Otherwise returns LoadedError
//
// If the Goner implements Provider interface (has Provide method), it will be registered as a provider.
// If a provider for the same type already exists:
// - If forceReplace option is true, the old one will be replaced
// - Otherwise returns LoadedError
//
// Parameters:
//   - goner: The Goner instance to load
//   - options: Optional configuration options for the Goner
//
// Available Options:
//   - Name(name string): Set custom name for the Goner
//   - IsDefault(): Mark this Goner as the default implementation
//   - OnlyForName(): Only register by name, not as provider
//   - ForceReplace(): Replace existing Goner with same name/type
//   - Order(order int): Set initialization order (lower runs first)
//   - FillWhenInit(): Fill dependencies during initialization
//
// Returns error if:
//   - Any option.Apply() fails
//   - A Goner with same name already exists (without forceReplace)
//   - A Provider for same type already exists (without forceReplace)
func (s *Core) Load(goner Goner, options ...Option) error {
	if goner == nil {
		return NewInnerError("goner cannot be nil - must provide a valid Goner instance", LoadedError)
	}
	co := newCoffin(goner)

	if namedGoner, ok := goner.(NamedGoner); ok {
		co.name = namedGoner.GonerName()
	}

	for _, option := range options {
		if err := option.Apply(co); err != nil {
			return ToError(err)
		}
	}

	if co.name != "" {
		if oldCo, ok := s.nameMap[co.name]; ok {
			if co.forceReplace {
				for i := range s.coffins {
					if s.coffins[i] == oldCo {
						s.coffins[i] = co
					}
				}
				s.nameMap[co.name] = co
			} else {
				return NewInnerErrorWithParams(LoadedError, "goner with name %q is already loaded - use ForceReplace() option to override", co.name)
			}
		} else {
			s.nameMap[co.name] = co
		}
	}

	var forceReplaceFind = false
	if co.forceReplace {
		for i := range s.coffins {
			if s.coffins[i] == co {
				s.coffins[i] = co
				forceReplaceFind = true
				break
			}
		}
	}

	if !forceReplaceFind {
		s.coffins = append(s.coffins, co)
	}

	if co.provider != nil {
		provider := co.provider

		if co.onlyForName {
			return nil
		}

		if oldCo, ok := s.typeProviderDepMap[provider.Type()]; ok {
			if oldCo.goner == goner {
				return NewInnerErrorWithParams(LoadedError, "provider for type %s is already registered with the same goner instance", GetTypeName(provider.Type()))
			}

			if co.forceReplace {
				for i := range s.coffins {
					if s.coffins[i] == oldCo {
						s.coffins[i] = co
					}
				}
				s.typeProviderDepMap[provider.Type()] = co
				s.typeProviderMap[provider.Type()] = provider
			} else {
				return NewInnerErrorWithParams(LoadedError, "provider for type %s is already registered - use ForceReplace() option to override", GetTypeName(provider.Type()))
			}
		} else {
			s.typeProviderMap[provider.Type()] = provider
			s.typeProviderDepMap[provider.Type()] = co
		}
	}
	if provider, ok := co.goner.(NamedProvider); ok {
		for t := range co.defaultTypeMap {
			if _, ok := s.typeProviderDepMap[t]; ok {
				return NewInnerErrorWithParams(LoadedError, "provider for type %s is already registered - cannot use IsDefault option when Loading named provider: %T(name=%s)", GetTypeName(t), provider, provider.GonerName())
			}
			s.typeProviderDepMap[t] = co
		}
	}
	return nil
}

// MustLoad is similar to Load but panics if an error occurs during loading.
// This provides a more convenient way to load components when you expect the operation to succeed.
//
// Parameters:
//   - goner: The Goner instance to load
//   - options: Optional configuration options for the Goner
//
// Returns:
//   - Loader: The Loader instance for method chaining
//
// Panics if:
//   - goner is nil
//   - Any option.Apply() fails
//   - A Goner with same name already exists (without forceReplace)
//   - A Provider for same type already exists (without forceReplace)
//
// Example usage:
//
//	loader.MustLoad(&MyComponent{}, gone.Name("myComponent"))
func (s *Core) MustLoad(goner Goner, options ...Option) Loader {
	if err := s.Load(goner, options...); err != nil {
		panic(err)
	}
	return s
}

// MustLoadX loads either a Goner or LoadFunc into the Gone container.
// It is similar to MustLoad but provides more flexibility by accepting different types.
//
// Parameters:
//   - x: Either a Goner instance or a LoadFunc function
//
// The function handles two cases:
// 1. If x is a Goner: Loads it directly using MustLoad
// 2. If x is a LoadFunc: Executes the function if not already loaded
//
// Returns:
//   - Loader: The Loader instance for method chaining
//
// Panics if:
//   - x is neither a Goner nor a LoadFunc
//   - Loading the Goner fails
//   - The LoadFunc returns an error
//
// Example usage:
//
//	loader.MustLoadX(&MyComponent{})  // Load a Goner
//	loader.MustLoadX(func(l Loader) error {  // Load using LoadFunc
//	    return l.Load(&DependencyA{})
//	})
func (s *Core) MustLoadX(x any) Loader {
	switch f := x.(type) {
	case Goner:
		s.MustLoad(f)
	case LoadFunc:
		if !s.Loaded(genLoaderKey(f)) {
			if err := f(s); err != nil {
				panic(err)
			}
		}
	default:
		panic(fmt.Sprintf("MustLoadX: unknown type: %T, only Goner or LoadFunc is allowed", x))
	}
	return s
}

// Loaded checks if a component identified by the given LoaderKey has already been loaded.
// This is used internally to prevent duplicate loading of components, especially when using LoadFunc.
//
// Parameters:
//   - key: The LoaderKey that uniquely identifies a component
//
// Returns:
//   - bool: true if the component has been loaded, false otherwise
//
// Note: If the component hasn't been loaded, it will be marked as loaded before returning false.
// This ensures that subsequent calls with the same key will return true.
func (s *Core) Loaded(key LoaderKey) bool {
	if _, ok := s.loaderMap[key]; ok {
		return true
	} else {
		s.loaderMap[key] = true
		return false
	}
}
