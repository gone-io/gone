package gone

import (
	"fmt"
	"reflect"
)

type actionType int8

const (
	fillAction          actionType = 1
	initAction          actionType = 2
	goneTag                        = "gone"
	defaultProviderName            = "*"
)

// Flag is a marker struct used to identify components that can be managed by the gone framework.
// Embedding this struct in another struct indicates that it can be used with gone's dependency injection.
type Flag struct{}

func (g *Flag) goneFlag() {}

// NewCore creates and initializes a new Core instance.
// It initializes the internal maps for tracking dependencies by name and type,
// and loads itself as a Goner to enable self-injection.
// Returns a pointer to the initialized Core.
func NewCore() *Core {
	loader := Core{
		nameMap:            make(map[string]*coffin),
		typeProviderMap:    make(map[reflect.Type]*wrapProvider),
		typeProviderDepMap: make(map[reflect.Type]*coffin),
		loaderMap:          make(map[LoaderKey]bool),
		log:                GetDefaultLogger(),
	}

	_ = loader.Load(&loader, IsDefault())
	_ = loader.Load(&ConfigProvider{})
	_ = loader.Load(&EnvConfigure{}, Name("configure"), IsDefault(new(Configure)), OnlyForName())
	_ = loader.Load(defaultLog, IsDefault(new(Logger)))
	return &loader
}

func (t actionType) String() string {
	switch t {
	case fillAction:
		return "fill fields"
	case initAction:
		return "initialize"
	default:
		return "unknown"
	}
}

type dependency struct {
	coffin *coffin
	action actionType
}

func (d dependency) String() string {
	var name string
	if d.coffin.name != "" {
		name = fmt.Sprintf("%q", d.coffin.name)
	} else {
		name = fmt.Sprintf("%q", GetTypeName(reflect.TypeOf(d.coffin.goner)))
	}
	return fmt.Sprintf("<%s of %s>", d.action.String(), name)
}

type Core struct {
	Flag
	coffins []*coffin

	nameMap            map[string]*coffin
	typeProviderMap    map[reflect.Type]*wrapProvider
	typeProviderDepMap map[reflect.Type]*coffin
	loaderMap          map[LoaderKey]bool
	log                Logger `gone:"*"`
}

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
		return NewInnerError("goner cannot be nil", LoadedError)
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
				return NewInnerErrorWithParams(LoadedError, "goner(name=%s) is already loaded", co.name)
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

	if co.onlyForName {
		return nil
	}

	provider := tryWrapGonerToProvider(goner)
	if provider != nil {
		co.needInitBeforeUse = true

		if oldCo, ok := s.typeProviderDepMap[provider.Type()]; ok {
			if oldCo.goner == goner {
				return NewInnerErrorWithParams(LoadedError, "Provider provided %s already registered", GetTypeName(provider.Type()))
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
				return NewInnerErrorWithParams(LoadedError, "Provider provided %s already registered", GetTypeName(provider.Type()))
			}
		} else {
			s.typeProviderMap[provider.Type()] = provider
			s.typeProviderDepMap[provider.Type()] = co
		}
	}
	return nil
}

// Check performs dependency validation and determines initialization order:
// 1. Collects all dependencies between components
// 2. Validates there are no circular dependencies
// 3. Determines optimal initialization order based on dependencies
// 4. Returns ordered list of components to initialize and any validation errors
func (s *Core) Check() ([]dependency, error) {
	depsMap, err := s.collectDeps()
	if err != nil {
		return nil, ToError(err)
	}

	deps, orders := checkCircularDepsAndGetBestInitOrder(depsMap)
	if len(deps) > 0 {
		return nil, circularDepsError(deps)
	}

	if s.log.GetLevel() <= DebugLevel {
		for i, dep := range orders {
			s.log.Debugf("Order[%d]: %s\n", i, dep)
		}
	}
	for _, co := range s.coffins {
		orders = append(orders, dependency{co, fillAction})
	}
	return RemoveRepeat(orders), nil
}

func (s *Core) Install() error {
	orders, err := s.Check()
	if err != nil {
		return ToError(err)
	}

	for i, dep := range orders {
		if dep.action == fillAction {
			if err := s.safeFillOne(dep.coffin); err != nil {
				s.log.Debugf("Failed to %s at order[%d]: %s", dep, i, err)
				return ToError(err)
			}
		}
		if dep.action == initAction {
			if err = s.safeInitOne(dep.coffin); err != nil {
				s.log.Debugf("Failed to %s at order[%d]: %s", dep, i, err)
				return ToError(err)
			}
		}
	}
	return nil
}

func (s *Core) safeFillOne(coffin *coffin) (err error) {
	return SafeExecute(func() error {
		return s.fillOne(coffin)
	})
}

func (s *Core) safeInitOne(coffin *coffin) error {
	return SafeExecute(func() error {
		return s.initOne(coffin)
	})
}

func (s *Core) fillOne(coffin *coffin) error {
	goner := coffin.goner

	if initiator, ok := goner.(BeforeInitiatorNoError); ok {
		initiator.BeforeInit()
	}

	if initiator, ok := goner.(BeforeInitiator); ok {
		err := initiator.BeforeInit()
		if err != nil {
			return ToError(err)
		}
	}

	elem := reflect.TypeOf(goner).Elem()
	if elem.Kind() != reflect.Struct {
		return NewInnerErrorWithParams(GonerTypeNotMatch, "goner must be a struct")
	}

	elemV := reflect.ValueOf(goner).Elem()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		v := elemV.Field(i)
		if !field.IsExported() {
			v = BlackMagic(v)
		}

		if tag, ok := field.Tag.Lookup(goneTag); ok {
			goneName, extend := ParseGoneTag(tag)
			if goneName == "" {
				goneName = defaultProviderName
			}

			co, err := s.getDepByName(goneName)
			if err != nil {
				return ToErrorWithMsg(err, fmt.Sprintf("Cannot find matched value for field %q of %q", field.Name, GetTypeName(elem)))
			}

			if IsCompatible(field.Type, co.goner) {
				v.Set(reflect.ValueOf(co.goner))
				continue
			}

			if provider, ok := co.goner.(NamedProvider); ok {
				provide, err := provider.Provide(extend, field.Type)
				if err != nil {
					return ToErrorWithMsg(err, fmt.Sprintf("Cannot find matched value for field %q of %q", field.Name, GetTypeName(elem)))
				}

				if provide == nil {
					return NewInnerErrorWithParams(GonerTypeNotMatch, "The value provided by provider(%T) cannot match to field %q of %q", provider, field.Name, GetTypeName(elem))
				}

				if IsCompatible(field.Type, provide) {
					v.Set(reflect.ValueOf(provide))
					continue
				}
				return NewInnerErrorWithParams(GonerTypeNotMatch, "The value provided by provider(%T) cannot match to field %q of %q", provider, field.Name, GetTypeName(elem))
			}
			if injector, ok := co.goner.(StructFieldInjector); ok {
				err = injector.Inject(extend, field, v)
				if err != nil {
					return ToErrorWithMsg(err, fmt.Sprintf("Cannot find matched value for field %q of %q", field.Name, GetTypeName(elem)))
				}
				continue
			}

			return NewInnerErrorWithParams(GonerTypeNotMatch, "Cannot find matched value for field %q of %q", field.Name, GetTypeName(elem))
		}
	}

	coffin.isFill = true
	return nil
}

func (s *Core) initOne(c *coffin) error {
	goner := c.goner
	if initiator, ok := goner.(InitiatorNoError); ok {
		initiator.Init()
	}
	if initiator, ok := goner.(Initiator); ok {
		if err := initiator.Init(); err != nil {
			return ToError(err)
		}
	}
	c.isInit = true
	return nil
}

func (s *Core) InjectStruct(goner any) error {
	of := reflect.TypeOf(goner)
	if of.Kind() != reflect.Ptr {
		return NewInnerError("goner should be a pointer to a struct", InjectError)
	}
	if of.Elem().Kind() != reflect.Struct {
		return NewInnerError("goner should be a pointer to a struct", InjectError)
	}
	co := &coffin{
		goner: goner,
	}
	err := s.safeFillOne(co)
	if err != nil {
		return ToError(err)
	}
	return s.safeFillOne(co)
}

func (s *Core) GetGonerByName(name string) any {
	co := s.nameMap[name]
	if co != nil {
		return co.goner
	}
	return nil
}

func (s *Core) GetGonerByType(t reflect.Type) any {
	T := s.getDefaultCoffinByType(t)
	if T != nil {
		return T.goner
	}
	return nil
}

func (s *Core) getCoffinsByType(t reflect.Type) (coffins []*coffin) {
	for _, tomb := range s.coffins {
		if tomb.onlyForName {
			continue
		}
		if IsCompatible(t, tomb.goner) {
			coffins = append(coffins, tomb)
		}
	}
	return
}

func (s *Core) getDefaultCoffinByType(t reflect.Type) *coffin {
	s.log.Debugf("Get Default Goner By Type: %s", GetTypeName(t))

	coffins := s.getCoffinsByType(t)
	if len(coffins) > 0 {
		for _, c := range coffins {
			if c.isDefault(t) {
				return c
			}
		}
		if len(coffins) > 1 {
			s.log.Warnf("Found multiple Goner for type %s; should set default one when loading", GetTypeName(t))
		}
		return coffins[0]
	}
	return nil
}

func (s *Core) GonerName() string {
	return defaultProviderName
}

func (s *Core) Provide(tagConf string, t reflect.Type) (any, error) {
	c := s.getDefaultCoffinByType(t)
	if c != nil {
		return c.goner, nil
	}

	if provider, ok := s.typeProviderMap[t]; ok {
		if provider != nil {
			return provider.Provide(tagConf)
		}
	}

	if t.Kind() == reflect.Slice {
		elem := t.Elem()
		coffins := s.getCoffinsByType(elem)

		pv := reflect.New(t)
		v := pv.Elem()

		if len(coffins) > 0 {
			for _, co := range coffins {
				v.Set(reflect.Append(v, reflect.ValueOf(co.goner)))
			}
		}
		if provider, ok := s.typeProviderMap[elem]; ok && provider != nil {
			provide, err := provider.Provide(tagConf)
			if err != nil {
				return nil, err
			}
			v.Set(reflect.Append(v, reflect.ValueOf(provide)))
		}
		return v.Interface(), nil
	}
	return nil, NewInnerError("Cannot find matched Goner for type %s", NotSupport)
}

// FuncInjectHook is a function type used for customizing parameter injection in functions.
// Parameters:
//   - pt: The type of parameter being injected
//   - i: The index of the parameter in the function signature
//   - injected: Whether the parameter has already been injected
//
// Returns any value that should be used as the injected parameter, or nil to continue with default injection
type FuncInjectHook func(pt reflect.Type, i int, injected bool) any

// InjectFuncParameters injects parameters into a function by:
// 1. Using injectBefore hook if provided
// 2. Using Core's Provide method to get dependencies
// 3. Creating and filling struct parameters if needed
// 4. Using injectAfter hook if provided
// Returns the injected parameter values or error if injection fails
func (s *Core) InjectFuncParameters(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (args []reflect.Value, err error) {
	ft := reflect.TypeOf(fn)

	if ft.Kind() != reflect.Func {
		return nil, NewInnerError("InjectFuncParameters, fn must be a function", NotSupport)
	}

	in := ft.NumIn()

	for i := 0; i < in; i++ {
		pt := ft.In(i)

		injected := false

		if injectBefore != nil {
			v := injectBefore(pt, i, false)
			if v != nil {
				args = append(args, reflect.ValueOf(v))
				injected = true
			}
		}

		if !injected {
			if v, err := s.Provide("", pt); err != nil && !IsError(err, NotSupport) {
				return nil, ToErrorWithMsg(err, fmt.Sprintf("Inject %dth parameter of %s error", i+1, GetFuncName(fn)))
			} else if v != nil {
				args = append(args, reflect.ValueOf(v))
				injected = true
			}
		}

		if !injected {
			if pt.Kind() == reflect.Struct {
				parameter := reflect.New(pt)
				if err = s.InjectStruct(parameter.Interface()); err != nil {
					return nil, ToErrorWithMsg(err, fmt.Sprintf("Inject %dth parameter of %s error", i+1, GetFuncName(fn)))
				}
				args = append(args, parameter.Elem())
				injected = true
			}

			if pt.Kind() == reflect.Ptr && pt.Elem().Kind() == reflect.Struct {
				parameter := reflect.New(pt.Elem())
				if err = s.InjectStruct(parameter.Interface()); err != nil {
					return nil, ToErrorWithMsg(err, fmt.Sprintf("Inject %dth parameter of %s error", i+1, GetFuncName(fn)))
				}
				args = append(args, parameter)
				injected = true
			}
		}

		if injectAfter != nil {
			v := injectAfter(pt, i, injected)
			if v != nil {
				args = append(args, reflect.ValueOf(v))
				injected = true
			}
		}

		if !injected {
			return nil, NewInnerError(fmt.Sprintf("Inject %dth parameter of %s Failed", i+1, GetFuncName(fn)), NotSupport)
		}
	}
	return
}

// InjectWrapFunc wraps a function with dependency injection.
// It injects dependencies into the function parameters and returns a wrapper function that:
// 1. Calls the original function with injected parameters
// 2. Converts return values to []any, handling nil interface values appropriately
// Parameters:
//   - fn: The function to wrap
//   - injectBefore: Optional hook called before standard injection
//   - injectAfter: Optional hook called after standard injection
//
// Returns:
//   - Wrapper function that returns results as []any
//   - Error if injection fails
func (s *Core) InjectWrapFunc(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (func() []any, error) {
	args, err := s.InjectFuncParameters(fn, injectBefore, injectAfter)
	if err != nil {
		return nil, err
	}

	return func() (results []any) {
		values := reflect.ValueOf(fn).Call(args)
		for _, arg := range values {
			if arg.Kind() == reflect.Interface {
				elem := arg.Elem()
				switch elem.Kind() {
				case reflect.Chan,
					reflect.Func,
					reflect.Interface,
					reflect.Map,
					reflect.Ptr,
					reflect.Slice,
					reflect.UnsafePointer:
					if elem.IsNil() {
						results = append(results, nil)
						continue
					}
				default:
				}
			}
			results = append(results, arg.Interface())
		}
		return results
	}, nil
}

func (s *Core) Loaded(key LoaderKey) bool {
	if _, ok := s.loaderMap[key]; ok {
		return true
	} else {
		s.loaderMap[key] = true
		return false
	}
}
