package gone

import (
	"fmt"
	"reflect"
)

func newCore() *core {
	k := newKeeper()
	l := GetDefaultLogger()
	a := newDependenceAnalyzer(k, l)
	i := newInstaller(a, l)
	c := &core{
		iKeeper:             k,
		iDependenceAnalyzer: a,
		iInstaller:          i,
		logger:              l,
		loaderMap:           make(map[LoaderKey]struct{}),
	}

	_ = k.load(k)
	_ = k.load(a)
	_ = k.load(i)
	_ = k.load(&ConfigProvider{})
	_ = k.load(&EnvConfigure{}, Name("configure"), IsDefault(new(Configure)))
	_ = k.load(l.(Goner), IsDefault(new(Logger)))
	_ = k.load(c, Name(DefaultProviderName))

	return c
}

type core struct {
	Flag
	iKeeper             iKeeper
	iInstaller          iInstaller
	iDependenceAnalyzer iDependenceAnalyzer
	logger              Logger `gone:"*"`

	loaderMap map[LoaderKey]struct{}
}

// InjectFuncParameters injects parameters into a function by:
// 1. Using injectBefore hook if provided
// 2. Using Core's Provide method to get dependencies
// 3. Creating and filling struct parameters if needed
// 4. Using injectAfter hook if provided
// Returns the injected parameter values or error if injection fails
func (s *core) InjectFuncParameters(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (args []reflect.Value, err error) {
	ft := reflect.TypeOf(fn)

	if ft.Kind() != reflect.Func {
		return nil, NewInnerError(fmt.Sprintf("cannot inject parameters: expected a function, got %v", ft.Kind()), NotSupport)
	}

	in := ft.NumIn()

	for i := 0; i < in; i++ {
		pt := ft.In(i)
		paramName := fmt.Sprintf("parameter #%d (%s)", i+1, GetTypeName(pt))

		injected := false

		if injectBefore != nil {
			if v := injectBefore(pt, i, false); v != nil {
				args = append(args, reflect.ValueOf(v))
				injected = true
			}
		}

		if !injected {
			if v, err := s.ProvideNth(i+1, pt, GetFuncName(fn)); err != nil {
				return nil, err
			} else if !v.IsZero() {
				args = append(args, v)
				injected = true
			}
		}

		if !injected {
			if pt.Kind() == reflect.Struct {
				parameter := reflect.New(pt)
				if err = s.InjectStruct(parameter.Interface()); err != nil {
					return nil, ToErrorWithMsg(err, fmt.Sprintf("failed to inject struct fields for %s in %s", paramName, GetFuncName(fn)))
				}
				args = append(args, parameter.Elem())
				injected = true
			}

			if pt.Kind() == reflect.Ptr && pt.Elem().Kind() == reflect.Struct {
				parameter := reflect.New(pt.Elem())
				if err = s.InjectStruct(parameter.Interface()); err != nil {
					return nil, ToErrorWithMsg(err, fmt.Sprintf("failed to inject struct pointer fields for %s in %s", paramName, GetFuncName(fn)))
				}
				args = append(args, parameter)
				injected = true
			}
		}

		if injectAfter != nil {
			if v := injectAfter(pt, i, injected); v != nil {
				args = append(args, reflect.ValueOf(v))
				injected = true
			}
		}

		if !injected {
			return nil, NewInnerError(fmt.Sprintf("no suitable injector found for %s in %s", paramName, GetFuncName(fn)), NotSupport)
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
func (s *core) InjectWrapFunc(fn any, injectBefore FuncInjectHook, injectAfter FuncInjectHook) (func() []any, error) {
	args, err := s.InjectFuncParameters(fn, injectBefore, injectAfter)
	if err != nil {
		return nil, err
	}

	return func() (results []any) {
		values := reflect.ValueOf(fn).Call(args)
		for _, arg := range values {
			switch arg.Kind() {
			case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
				if arg.IsNil() {
					results = append(results, nil)
					continue
				}
				fallthrough
			default:
				results = append(results, arg.Interface())
			}
		}
		return results
	}, nil
}

func (s *core) InjectStruct(goner any) error {
	of := reflect.TypeOf(goner)
	if of.Kind() != reflect.Ptr {
		return NewInnerError("goner must be a pointer to a struct, got non-pointer type", InjectError)
	}
	if of.Elem().Kind() != reflect.Struct {
		return NewInnerError("goner must be a pointer to a struct, got pointer to non-struct type", InjectError)
	}

	return ToError(s.iInstaller.safeFillOne(newCoffin(goner)))
}

func (s *core) GetGonerByName(name string) any {
	co := s.iKeeper.getByName(name)
	if co != nil {
		return co.goner
	}
	return nil
}

func (s *core) GetGonerByType(t reflect.Type) any {
	if co := s.iKeeper.selectOneCoffin(t, "*", func() {
		s.logger.Warnf("found multiple value without a default when calling GetGonerByType(%s) - using first one. ", GetTypeName(t))
	}); co != nil {
		if v, err := co.Provide(false, "", t); err != nil {
			panic(err)
		} else {
			return v
		}
	}
	return nil
}

func (s *core) ProvideNth(n int, t reflect.Type, funcName string) (reflect.Value, error) {
	field := reflect.StructField{
		Name: fmt.Sprintf("The%dthParameter", n),
		Type: t,
		Tag:  `gone:"*" option:"allowNil"`,
	}
	v := reflect.New(t).Elem()

	if err := s.iInstaller.analyzerFieldDependencies(field, funcName, func(asSlice, byName bool, extend string, coffins ...*coffin) error {
		return s.iInstaller.injectField(asSlice, byName, extend, coffins, field, v, funcName)
	}); err != nil {
		return v, ToErrorWithMsg(err, fmt.Sprintf("can not provide nth parameter for %s", funcName))
	}
	return v, nil
}

// Check performs dependency validation and determines initialization order:
// 1. Collects all dependencies between components
// 2. Validates there are no circular dependencies
// 3. Determines optimal initialization order based on dependencies
// 4. Returns ordered list of components to initialize and any validation errors
func (s *core) Check() ([]dependency, error) {
	deps, orders, err := s.iDependenceAnalyzer.checkCircularDepsAndGetBestInitOrder()
	if err != nil {
		return nil, err
	}
	if len(deps) > 0 {
		return nil, circularDepsError(deps)
	}

	if s.logger.GetLevel() <= DebugLevel {
		for i, dep := range orders {
			s.logger.Debugf("Order[%d]: %s\n", i, dep)
		}
	}
	for _, co := range s.iKeeper.getAllCoffins() {
		orders = append(orders, dependency{co, fillAction})
	}
	return RemoveRepeat(orders), nil
}

func (s *core) Install() error {
	orders, err := s.Check()
	if err != nil {
		return ToError(err)
	}

	for i, dep := range orders {
		if dep.action == fillAction {
			if err := s.iInstaller.safeFillOne(dep.coffin); err != nil {
				s.logger.Debugf("failed to %s at order[%d]: %s", dep, i, err)
				return ToError(err)
			}
		}
		if dep.action == initAction {
			if err = s.iInstaller.safeInitOne(dep.coffin); err != nil {
				s.logger.Debugf("failed to %s at order[%d]: %s", dep, i, err)
				return ToError(err)
			}
		}
	}
	return nil
}

var _ GonerKeeper = (*core)(nil)
var _ Loader = (*core)(nil)
var _ StructInjector = (*core)(nil)
var _ FuncInjector = (*core)(nil)
