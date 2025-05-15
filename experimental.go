package gone

import "reflect"

// FunctionProvider is an experimental type that may change or be removed in future releases.
type FunctionProvider[P, T any] func(tagConf string, param P) (T, error)

// XProvider is an experimental type that may change or be removed in future releases.
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

// WrapFunctionProvider is an experimental function that may change or be removed in future releases.
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
