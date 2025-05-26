package gone

import "reflect"

// FunctionProvider is a function, which first parameter is tagConf, and second parameter is a struct that can be injected.
// And the function must return a T type value and error.
type FunctionProvider[P, T any] func(tagConf string, param P) (T, error)

// XProvider is a Goner Provider was created by WrapFunctionProvider.
type XProvider[T any] struct {
	Flag
	injector FuncInjector `gone:"*"`
	create   func(tagConf string) (T, error)
}

func (p *XProvider[T]) Provide(tagConf string) (T, error) {
	obj, err := p.create(tagConf)
	if err != nil {
		return obj, ToError(err)
	}
	return obj, nil
}

// WrapFunctionProvider can wrap a FunctionProvider to a Provider.
func WrapFunctionProvider[P, T any](fn FunctionProvider[P, T]) *XProvider[T] {
	p := XProvider[T]{}

	p.create = func(tagConf string) (T, error) {
		f, err := p.injector.InjectWrapFunc(fn, func(pt reflect.Type, i int, injected bool) any {
			if i == 0 {
				return tagConf
			}
			return nil
		}, nil)

		if err != nil {
			return *new(T), err
		}
		results := f()

		var t T
		if results[0] == nil {
			t = *new(T)
		} else {
			t = results[0].(T)
		}
		if results[1] == nil {
			err = nil
		} else {
			err = results[1].(error)
		}
		return t, err
	}
	return &p
}

// WarpThirdComponent can wrap a third component to a Goner Provider which can make third component to inject Goners.
func WarpThirdComponent[T any](t T) Goner {
	provider := WrapFunctionProvider(func(tagConf string, param struct{}) (T, error) {
		return t, nil
	})
	return provider
}
