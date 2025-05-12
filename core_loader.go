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
func (s *core) Load(goner Goner, options ...Option) error {
	return s.iKeeper.load(goner, options...)
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
func (s *core) MustLoad(goner Goner, options ...Option) Loader {
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
func (s *core) MustLoadX(x any) Loader {
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
		panic(ToError(fmt.Sprintf("MustLoadX: unknown type: %T, only Goner or LoadFunc is allowed", x)))
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
func (s *core) Loaded(key LoaderKey) bool {
	if _, ok := s.loaderMap[key]; ok {
		return true
	} else {
		s.loaderMap[key] = struct{}{}
		return false
	}
}
