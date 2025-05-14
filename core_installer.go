package gone

import (
	"errors"
	"fmt"
	"reflect"
)

func newInstaller(iDependenceAnalyzer iDependenceAnalyzer, logger Logger) *installer {
	return &installer{
		iDependenceAnalyzer: iDependenceAnalyzer,
		logger:              logger,
	}
}

type installer struct {
	Flag
	iDependenceAnalyzer
	logger Logger `gone:"*"`
}

func (s *installer) injectField(
	asSlice, byName bool, extend string, depCoffins []*coffin,
	field reflect.StructField, v reflect.Value, coName string,
) error {
	if !field.IsExported() {
		v = BlackMagic(v)
	}

	if asSlice {
		return s.injectFieldAsSlice(extend, depCoffins, field, v, coName)
	} else {
		return s.injectFieldAsNotSlice(byName, extend, depCoffins[0], field, v, coName)
	}
}

func (s *installer) injectFieldAsSlice(extend string, depCoffins []*coffin, field reflect.StructField, v reflect.Value, coName string) error {
	elType := field.Type.Elem()
	slice := reflect.MakeSlice(field.Type, 0, len(depCoffins))
	for _, depCo := range depCoffins {
		if value, err := depCo.Provide(false, extend, elType); err != nil {
			return ToErrorWithMsg(err, fmt.Sprintf("%q failed to provide value for filed %q element of %q",
				depCo.Name(), field.Name, coName),
			)
		} else {
			slice = reflect.Append(slice, reflect.ValueOf(value))
		}
	}
	v.Set(slice)
	return nil
}

func (s *installer) injectFieldAsNotSlice(byName bool, extend string, depCo *coffin, field reflect.StructField, v reflect.Value, coName string) error {
	if value, err := depCo.Provide(byName, extend, field.Type); err != nil {
		var e Error
		if errors.As(err, &e) && e.Code() == NotSupport {
			if injector, ok := depCo.goner.(StructFieldInjector); ok {
				if err := injector.Inject(extend, field, v); err != nil {
					return ToErrorWithMsg(err,
						fmt.Sprintf("%q failed to inject for field %q in %q", depCo.Name(), field.Name, coName),
					)
				}
				return nil
			}
		}
		return ToErrorWithMsg(err,
			fmt.Sprintf("%q failed to provide value for field %q of %q", depCo.Name(), field.Name, coName),
		)
	} else {
		v.Set(reflect.ValueOf(value))
		return nil
	}
}

func (s *installer) fillOne(co *coffin) error {
	if err := s.doBeforeInit(co.goner); err != nil {
		return err
	}
	elem := reflect.TypeOf(co.goner).Elem()
	if elem.Kind() != reflect.Struct {
		return NewInnerErrorWithParams(GonerTypeNotMatch,
			"cannot inject: expected a pointer to struct, but got %q",
			co.Name(),
		)
	}

	elemV := reflect.ValueOf(co.goner).Elem()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)

		injectProcess := func(asSlice, byName bool, extend string, depCoffins ...*coffin) error {
			return s.injectField(asSlice, byName, extend, depCoffins, field, elemV.Field(i), co.Name())
		}

		if err := s.iDependenceAnalyzer.analyzerFieldDependencies(field, co.Name(), injectProcess); err != nil {
			return err
		}
	}

	co.isFill = true
	return nil
}

func (s *installer) doBeforeInit(goner any) error {
	if initiator, ok := goner.(BeforeInitiatorNoError); ok {
		initiator.BeforeInit()
	}

	if initiator, ok := goner.(BeforeInitiator); ok {
		err := initiator.BeforeInit()
		if err != nil {
			return ToError(err)
		}
	}
	return nil
}

func (s *installer) safeFillOne(c *coffin) error {
	return SafeExecute(func() error {
		return s.fillOne(c)
	})
}

func (s *installer) safeInitOne(c *coffin) error {
	return SafeExecute(func() error {
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
	})
}
