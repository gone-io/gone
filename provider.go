package gone

import "reflect"

type wrapProvider struct {
	value        any
	hasParameter bool
	t            reflect.Type
}

func tryWrapGonerToProvider(goner any) *wrapProvider {
	if goner == nil {
		return nil
	}
	gonerType := reflect.TypeOf(goner)
	if gonerType.Kind() != reflect.Ptr {
		return nil
	}

	method, ok := gonerType.MethodByName("Provide")
	if !ok {
		return nil
	}

	ft := method.Type
	hasParameter := ft.NumIn() == 2 && ft.In(1).Kind() == reflect.String
	isValid := (ft.NumIn() == 1 || hasParameter) && ft.NumOut() == 2 && ft.Out(1).Implements(errType)

	if !isValid {
		return nil
	}

	return &wrapProvider{
		value:        goner,
		hasParameter: hasParameter,
		t:            ft.Out(0),
	}
}

func (p *wrapProvider) Provide(conf string) (any, error) {
	if p.hasParameter {
		results := reflect.ValueOf(p.value).MethodByName("Provide").Call([]reflect.Value{
			reflect.ValueOf(conf),
		})
		if results[1].IsNil() {
			return results[0].Interface(), nil
		}
		return nil, results[1].Interface().(error)
	}

	results := reflect.ValueOf(p.value).MethodByName("Provide").Call(nil)
	if results[1].IsNil() {
		return results[0].Interface(), nil
	}
	return nil, results[1].Interface().(error)
}

func (p *wrapProvider) Type() reflect.Type {
	return p.t
}

func (p *wrapProvider) ProvideTypeCompatible(t reflect.Type) bool {
	if p.t == t {
		return true
	}
	if p.t.Kind() == reflect.Interface && t.Implements(p.t) {
		return true
	}
	return false
}

var errType = reflect.TypeOf((*error)(nil)).Elem()
